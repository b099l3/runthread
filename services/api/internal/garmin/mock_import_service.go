package garmin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/providerimport"
	"github.com/runthread/runthread/services/api/internal/repository"
)

const (
	ProviderNameGarmin          = "garmin"
	mockImportEventTypeActivity = "mock_activity_import"
	mockActivityPayloadKind     = "mock_activity"
)

type MockImportService struct {
	providers          repository.ProviderStore
	importedActivities repository.ImportedActivityRepository
	now                func() time.Time
}

type MockImportServiceOption func(*MockImportService)

func NewMockImportService(providers repository.ProviderStore, importedActivities repository.ImportedActivityRepository, opts ...MockImportServiceOption) (*MockImportService, error) {
	if providers == nil {
		return nil, errors.New("provider store is required")
	}
	if importedActivities == nil {
		return nil, errors.New("imported activity repository is required")
	}

	service := &MockImportService{
		providers:          providers,
		importedActivities: importedActivities,
		now:                time.Now,
	}
	for _, opt := range opts {
		opt(service)
	}
	if service.now == nil {
		return nil, errors.New("clock is required")
	}
	return service, nil
}

func WithMockImportClock(now func() time.Time) MockImportServiceOption {
	return func(service *MockImportService) {
		service.now = now
	}
}

type MockImportRequest struct {
	Connection repository.ProviderConnection
	Payload    MockActivityPayload
	RawPayload []byte
	ReceivedAt time.Time
	DeliveryID string
}

type MockImportResult struct {
	ProviderActivity repository.ProviderActivity
	ImportedActivity domain.ImportedActivity
	ImportEvent      repository.ProviderImportEvent
	PayloadStored    bool
}

func (s *MockImportService) ImportActivity(ctx context.Context, req MockImportRequest) (MockImportResult, error) {
	if err := ctx.Err(); err != nil {
		return MockImportResult{}, err
	}
	if err := req.Connection.Validate(); err != nil {
		return MockImportResult{}, fmt.Errorf("invalid provider connection: %w", err)
	}
	if req.Connection.Provider != ProviderNameGarmin {
		return MockImportResult{}, fmt.Errorf("unsupported provider %q", req.Connection.Provider)
	}
	if req.Payload.AthleteID == "" {
		req.Payload.AthleteID = req.Connection.AthleteID
	}
	if req.Payload.AthleteID != req.Connection.AthleteID {
		return MockImportResult{}, fmt.Errorf("payload athlete %q does not match connection athlete %q", req.Payload.AthleteID, req.Connection.AthleteID)
	}

	importedActivity, err := NormalizeMockActivity(req.Payload)
	failureReason := ""
	if err != nil {
		failureReason = fmt.Sprintf("normalise mock Garmin activity: %v", err)
	} else {
		importedActivity.ID = providerimport.DeterministicID("runthread:imported-activity", req.Connection.ID, req.Payload.ActivityID)
	}

	// The mock request carries a connection directly, so persist it before calling
	// the provider-neutral orchestrator that loads connections by ID.
	if err := s.providers.SaveProviderConnection(ctx, req.Connection); err != nil {
		return MockImportResult{}, fmt.Errorf("save provider connection: %w", err)
	}
	orchestrator, err := providerimport.NewService(s.providers, s.importedActivities)
	if err != nil {
		return MockImportResult{}, err
	}
	orchestrator.Now = s.now

	var imported *domain.ImportedActivity
	if failureReason == "" {
		imported = &importedActivity
	}
	result, importErr := orchestrator.ImportActivity(ctx, providerimport.ImportRequest{
		AthleteID:            req.Connection.AthleteID,
		ProviderConnectionID: req.Connection.ID,
		ProviderActivity: providerimport.ProviderActivityInput{
			ProviderActivityID:   req.Payload.ActivityID,
			ProviderActivityType: req.Payload.GarminActivityType,
			StartedAt:            req.Payload.StartTime,
		},
		ImportedActivity: imported,
		RawPayload:       req.RawPayload,
		PayloadKind:      mockActivityPayloadKind,
		Delivery: providerimport.DeliveryMetadata{
			EventType:  mockImportEventTypeActivity,
			DeliveryID: req.DeliveryID,
			ReceivedAt: req.ReceivedAt,
		},
		FailureReason: failureReason,
	})
	mockResult := MockImportResult{
		ProviderActivity: result.ProviderActivity,
		ImportEvent:      result.ImportEvent,
		PayloadStored:    result.PayloadStored,
	}
	if result.ImportedActivity != nil {
		mockResult.ImportedActivity = *result.ImportedActivity
	}
	return mockResult, importErr
}
