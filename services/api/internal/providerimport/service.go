package providerimport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/repository"
)

const (
	defaultEventType          = "activity_import"
	defaultPayloadKind        = "activity"
	providerActivityIDPrefix  = "runthread:provider-activity"
	providerPayloadIDPrefix   = "runthread:provider-activity-payload"
	providerImportEventPrefix = "runthread:provider-import-event"
)

type Service struct {
	Providers          repository.ProviderStore
	ImportedActivities repository.ImportedActivityRepository
	Now                func() time.Time
}

type ProviderActivityInput struct {
	ProviderActivityID   string
	ProviderActivityType string
	StartedAt            time.Time
}

type DeliveryMetadata struct {
	EventType  string
	DeliveryID string
	ReceivedAt time.Time
}

type ImportRequest struct {
	AthleteID            string
	ProviderConnectionID string
	ProviderActivity     ProviderActivityInput
	ImportedActivity     *domain.ImportedActivity
	RawPayload           []byte
	PayloadKind          string
	Delivery             DeliveryMetadata
	IgnoreReason         string
	FailureReason        string
}

type ImportResult struct {
	Connection       repository.ProviderConnection
	ProviderActivity repository.ProviderActivity
	ImportedActivity *domain.ImportedActivity
	ImportEvent      repository.ProviderImportEvent
	PayloadStored    bool
}

func NewService(providers repository.ProviderStore, importedActivities repository.ImportedActivityRepository) (Service, error) {
	if providers == nil {
		return Service{}, errors.New("provider store is required")
	}
	if importedActivities == nil {
		return Service{}, errors.New("imported activity repository is required")
	}
	return Service{
		Providers:          providers,
		ImportedActivities: importedActivities,
		Now:                time.Now,
	}, nil
}

func (s Service) ImportActivity(ctx context.Context, req ImportRequest) (ImportResult, error) {
	if err := ctx.Err(); err != nil {
		return ImportResult{}, err
	}
	if s.Providers == nil {
		return ImportResult{}, errors.New("provider store is required")
	}
	if s.ImportedActivities == nil {
		return ImportResult{}, errors.New("imported activity repository is required")
	}
	if req.ProviderConnectionID == "" {
		return ImportResult{}, errors.New("provider connection id is required")
	}
	if req.ProviderActivity.ProviderActivityID == "" {
		return ImportResult{}, errors.New("provider activity id is required")
	}

	connection, err := s.Providers.GetProviderConnection(ctx, req.ProviderConnectionID)
	if err != nil {
		return ImportResult{}, fmt.Errorf("get provider connection: %w", err)
	}
	if err := connection.Validate(); err != nil {
		return ImportResult{}, fmt.Errorf("invalid provider connection: %w", err)
	}
	if req.AthleteID != "" && req.AthleteID != connection.AthleteID {
		return ImportResult{}, fmt.Errorf("request athlete %q does not match provider connection athlete %q", req.AthleteID, connection.AthleteID)
	}

	receivedAt := req.Delivery.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = s.now()
	}

	activity, err := s.providerActivity(ctx, connection, req.ProviderActivity, receivedAt)
	if err != nil {
		return ImportResult{Connection: connection}, err
	}
	event := newImportEvent(connection, activity.ID, req.ProviderActivity.ProviderActivityID, req.Delivery, receivedAt)

	if err := s.Providers.SaveProviderActivity(ctx, activity); err != nil {
		return ImportResult{Connection: connection, ProviderActivity: activity, ImportEvent: event}, fmt.Errorf("save provider activity: %w", err)
	}
	if err := s.Providers.SaveProviderImportEvent(ctx, event); err != nil {
		return ImportResult{Connection: connection, ProviderActivity: activity, ImportEvent: event}, fmt.Errorf("save received provider import event: %w", err)
	}

	result := ImportResult{
		Connection:       connection,
		ProviderActivity: activity,
		ImportEvent:      event,
	}

	if len(req.RawPayload) > 0 {
		payloadKind := req.PayloadKind
		if payloadKind == "" {
			payloadKind = defaultPayloadKind
		}
		payload := repository.ProviderActivityPayload{
			ID:                 DeterministicID(providerPayloadIDPrefix, activity.ID, payloadKind),
			ProviderActivityID: activity.ID,
			Payload:            append([]byte(nil), req.RawPayload...),
			PayloadKind:        payloadKind,
			ReceivedAt:         receivedAt,
		}
		if err := s.Providers.SaveProviderActivityPayload(ctx, payload); err != nil {
			return s.fail(ctx, result, fmt.Errorf("save provider activity payload: %w", err))
		}
		result.PayloadStored = true
	}

	if req.ImportedActivity == nil {
		if req.FailureReason != "" {
			return s.fail(ctx, result, errors.New(req.FailureReason))
		}
		if req.IgnoreReason == "" {
			return s.fail(ctx, result, errors.New("imported activity or ignore reason is required"))
		}
		result.ProviderActivity.Status = repository.ProviderActivityStatusIgnored
		result.ProviderActivity.LastSyncedAt = s.now()
		result.ProviderActivity.LastError = req.IgnoreReason
		result.ImportEvent.Status = repository.ProviderImportEventStatusIgnored
		result.ImportEvent.ProcessedAt = s.now()
		result.ImportEvent.Error = req.IgnoreReason
		return s.saveTerminalState(ctx, result)
	}

	importedActivity := *req.ImportedActivity
	if importedActivity.AthleteID != connection.AthleteID {
		return s.fail(ctx, result, fmt.Errorf("imported activity athlete %q does not match provider connection athlete %q", importedActivity.AthleteID, connection.AthleteID))
	}
	if err := importedActivity.Validate(); err != nil {
		return s.fail(ctx, result, fmt.Errorf("invalid imported activity: %w", err))
	}
	if err := s.ImportedActivities.SaveImportedActivity(ctx, importedActivity); err != nil {
		return s.fail(ctx, result, fmt.Errorf("save imported activity: %w", err))
	}

	result.ImportedActivity = &importedActivity
	result.ProviderActivity.ImportedActivityID = importedActivity.ID
	result.ProviderActivity.Status = repository.ProviderActivityStatusNormalised
	result.ProviderActivity.LastSyncedAt = s.now()
	result.ProviderActivity.LastError = ""
	result.ImportEvent.Status = repository.ProviderImportEventStatusProcessed
	result.ImportEvent.ProcessedAt = s.now()

	return s.saveTerminalState(ctx, result)
}

func (s Service) providerActivity(ctx context.Context, connection repository.ProviderConnection, input ProviderActivityInput, firstSeenAt time.Time) (repository.ProviderActivity, error) {
	existing, err := s.Providers.GetProviderActivityByProviderID(ctx, connection.ID, input.ProviderActivityID)
	if err == nil {
		existing.ProviderActivityType = input.ProviderActivityType
		existing.StartedAt = input.StartedAt
		existing.Status = repository.ProviderActivityStatusReceived
		existing.LastSyncedAt = time.Time{}
		existing.LastError = ""
		existing.UpdatedAt = firstSeenAt
		return existing, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return repository.ProviderActivity{}, fmt.Errorf("get provider activity by provider id: %w", err)
	}

	return repository.ProviderActivity{
		ID:                   DeterministicID(providerActivityIDPrefix, connection.ID, input.ProviderActivityID),
		ProviderConnectionID: connection.ID,
		AthleteID:            connection.AthleteID,
		Provider:             connection.Provider,
		ProviderActivityID:   input.ProviderActivityID,
		ProviderActivityType: input.ProviderActivityType,
		StartedAt:            input.StartedAt,
		Status:               repository.ProviderActivityStatusReceived,
		FirstSeenAt:          firstSeenAt,
		CreatedAt:            firstSeenAt,
		UpdatedAt:            firstSeenAt,
	}, nil
}

func (s Service) saveTerminalState(ctx context.Context, result ImportResult) (ImportResult, error) {
	if err := s.Providers.SaveProviderActivity(ctx, result.ProviderActivity); err != nil {
		return result, fmt.Errorf("save terminal provider activity: %w", err)
	}
	if err := s.Providers.SaveProviderImportEvent(ctx, result.ImportEvent); err != nil {
		return result, fmt.Errorf("save terminal provider import event: %w", err)
	}
	return result, nil
}

func (s Service) fail(ctx context.Context, result ImportResult, cause error) (ImportResult, error) {
	result.ProviderActivity.Status = repository.ProviderActivityStatusFailed
	result.ProviderActivity.LastSyncedAt = s.now()
	result.ProviderActivity.LastError = cause.Error()
	result.ImportEvent.Status = repository.ProviderImportEventStatusFailed
	result.ImportEvent.ProcessedAt = s.now()
	result.ImportEvent.Error = cause.Error()
	_, _ = s.saveTerminalState(ctx, result)
	return result, cause
}

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func newImportEvent(connection repository.ProviderConnection, providerActivityID string, externalActivityID string, delivery DeliveryMetadata, receivedAt time.Time) repository.ProviderImportEvent {
	eventType := delivery.EventType
	if eventType == "" {
		eventType = defaultEventType
	}
	deliveryID := delivery.DeliveryID
	if deliveryID == "" {
		deliveryID = DeterministicID("runthread:provider-delivery", connection.ID, externalActivityID, eventType)
	}

	return repository.ProviderImportEvent{
		ID:                   DeterministicID(providerImportEventPrefix, connection.ID, externalActivityID, eventType, deliveryID),
		ProviderConnectionID: connection.ID,
		ProviderActivityID:   providerActivityID,
		Provider:             connection.Provider,
		EventType:            eventType,
		DeliveryID:           deliveryID,
		Status:               repository.ProviderImportEventStatusReceived,
		ReceivedAt:           receivedAt,
	}
}

func DeterministicID(parts ...string) string {
	encoded, _ := json.Marshal(parts)
	return uuid.NewSHA1(uuid.NameSpaceURL, encoded).String()
}
