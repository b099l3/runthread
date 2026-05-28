package strava

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/runthread/runthread/services/api/internal/providerimport"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type WebhookVerifier interface {
	VerifyWebhook(ctx context.Context, req VerifyWebhookRequest) error
}

type WebhookDeduper interface {
	SeenWebhookEvent(ctx context.Context, eventID string) (bool, error)
	MarkWebhookEventSeen(ctx context.Context, eventID string) error
}

type VerifyWebhookRequest struct {
	Body      []byte
	Signature string
}

type WebhookService struct {
	Providers repository.ProviderStore
	Importer  providerimport.Service
	Fetcher   ActivityFetcher
	Verifier  WebhookVerifier
	Deduper   WebhookDeduper
	Now       func() time.Time
}

type HandleWebhookRequest struct {
	Body      []byte
	Signature string
}

type WebhookEvent struct {
	EventID         string `json:"event_id"`
	AspectType      string `json:"aspect_type"`
	ObjectType      string `json:"object_type"`
	ObjectID        string `json:"object_id"`
	ProviderUserID  string `json:"owner_id"`
	ProviderUserID2 string `json:"provider_user_id"`
	SubscriptionID  string `json:"subscription_id"`
	EventTimeUnix   int64  `json:"event_time"`
	ReceivedAtUnix  int64  `json:"received_at"`
}

type WebhookAction string

const (
	WebhookActionImported  WebhookAction = "imported"
	WebhookActionIgnored   WebhookAction = "ignored"
	WebhookActionDeleted   WebhookAction = "deleted"
	WebhookActionDuplicate WebhookAction = "duplicate"
	WebhookActionFailed    WebhookAction = "failed"
)

const RetryableWebhookFailurePrefix = "retryable Strava webhook fetch failure"

type WebhookResult struct {
	Action WebhookAction
	Event  WebhookEvent
	Import *providerimport.ImportResult
}

func (s WebhookService) HandleWebhook(ctx context.Context, req HandleWebhookRequest) (WebhookResult, error) {
	if err := s.validate(); err != nil {
		return WebhookResult{}, err
	}
	if len(req.Body) == 0 {
		return WebhookResult{}, errors.New("webhook body is required")
	}
	if err := s.Verifier.VerifyWebhook(ctx, VerifyWebhookRequest{Body: req.Body, Signature: req.Signature}); err != nil {
		return WebhookResult{Action: WebhookActionFailed}, fmt.Errorf("verify Strava webhook: %w", err)
	}

	var event WebhookEvent
	if err := json.Unmarshal(req.Body, &event); err != nil {
		return WebhookResult{Action: WebhookActionFailed}, fmt.Errorf("decode Strava webhook: %w", err)
	}
	event.normalise()
	if err := event.validate(); err != nil {
		return WebhookResult{Action: WebhookActionFailed, Event: event}, err
	}

	seen, err := s.Deduper.SeenWebhookEvent(ctx, event.EventID)
	if err != nil {
		return WebhookResult{Action: WebhookActionFailed, Event: event}, fmt.Errorf("check Strava webhook duplicate: %w", err)
	}
	if seen {
		return WebhookResult{Action: WebhookActionDuplicate, Event: event}, nil
	}

	connection, err := s.connectedConnectionForEvent(ctx, event)
	if err != nil {
		return WebhookResult{Action: WebhookActionFailed, Event: event}, err
	}

	result, err := s.routeEvent(ctx, connection, event)
	if err != nil {
		return result, err
	}
	if err := s.Deduper.MarkWebhookEventSeen(ctx, event.EventID); err != nil {
		return WebhookResult{Action: WebhookActionFailed, Event: event}, fmt.Errorf("mark Strava webhook seen: %w", err)
	}
	return result, nil
}

func (s WebhookService) routeEvent(ctx context.Context, connection repository.ProviderConnection, event WebhookEvent) (WebhookResult, error) {
	if event.ObjectType == "athlete" {
		return s.recordAthleteEvent(ctx, connection, event)
	}

	switch event.AspectType {
	case "create", "update":
		detail, err := s.Fetcher.FetchActivityDetail(ctx, ActivityDetailRequest{
			Connection: connection,
			ActivityID: event.ObjectID,
		})
		if err != nil {
			result, importErr := s.recordWebhookFailure(ctx, connection, event, err)
			if importErr != nil {
				return result, importErr
			}
			return result, err
		}
		importResult, imported, ignored, err := s.importWebhookActivity(ctx, connection, event, detail)
		result := WebhookResult{Event: event, Import: &importResult}
		if imported {
			result.Action = WebhookActionImported
		} else if ignored {
			result.Action = WebhookActionIgnored
		} else if err != nil {
			result.Action = WebhookActionFailed
		}
		return result, err
	case "delete":
		importResult, err := s.recordWebhookDelete(ctx, connection, event)
		result := WebhookResult{Action: WebhookActionDeleted, Event: event, Import: &importResult}
		return result, err
	default:
		importResult, err := s.recordWebhookIgnored(ctx, connection, event, "unsupported Strava webhook aspect type")
		result := WebhookResult{Action: WebhookActionIgnored, Event: event, Import: &importResult}
		return result, err
	}
}

func (s WebhookService) recordAthleteEvent(ctx context.Context, connection repository.ProviderConnection, event WebhookEvent) (WebhookResult, error) {
	if event.AspectType != "delete" {
		return WebhookResult{Action: WebhookActionIgnored, Event: event}, nil
	}
	now := s.now()
	connection.Status = repository.ProviderConnectionStatusDisconnected
	connection.DisconnectedAt = now
	connection.UpdatedAt = now
	connection.LastError = "Strava access revoked"
	if err := s.Providers.SaveProviderConnection(ctx, connection); err != nil {
		return WebhookResult{Action: WebhookActionFailed, Event: event}, fmt.Errorf("record Strava athlete disconnect: %w", err)
	}
	return WebhookResult{Action: WebhookActionDeleted, Event: event}, nil
}

func (s WebhookService) importWebhookActivity(ctx context.Context, connection repository.ProviderConnection, event WebhookEvent, payload MockActivityPayload) (providerimport.ImportResult, bool, bool, error) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return providerimport.ImportResult{}, false, false, fmt.Errorf("marshal mock Strava webhook activity payload: %w", err)
	}
	importRequest := s.webhookImportRequest(connection, event)
	importRequest.ProviderActivity.ProviderActivityType = payload.StravaSportType
	importRequest.ProviderActivity.StartedAt = payload.StartDate
	importRequest.RawPayload = rawPayload
	importRequest.PayloadKind = "strava_webhook_activity_detail"

	activity, err := NormaliseMockActivity(payload)
	if err != nil {
		if errors.Is(err, ErrUnsupportedActivityType) {
			importRequest.IgnoreReason = err.Error()
			result, importErr := s.Importer.ImportActivity(ctx, importRequest)
			return result, false, true, importErr
		}
		importRequest.FailureReason = err.Error()
		result, importErr := s.Importer.ImportActivity(ctx, importRequest)
		if importErr != nil {
			return result, false, false, importErr
		}
		return result, false, false, err
	}
	activity.ID = stravaImportedActivityID(event.ObjectID)
	importRequest.ImportedActivity = &activity

	result, err := s.Importer.ImportActivity(ctx, importRequest)
	return result, err == nil, false, err
}

func (s WebhookService) recordWebhookDelete(ctx context.Context, connection repository.ProviderConnection, event WebhookEvent) (providerimport.ImportResult, error) {
	req := s.webhookImportRequest(connection, event)
	req.IgnoreReason = "Strava activity deleted; existing imported activity preserved for review"
	result, err := s.Importer.ImportActivity(ctx, req)
	if err != nil {
		return result, err
	}
	result.ProviderActivity.Status = repository.ProviderActivityStatusDeleted
	result.ProviderActivity.LastSyncedAt = s.now()
	result.ProviderActivity.LastError = req.IgnoreReason
	if err := s.Providers.SaveProviderActivity(ctx, result.ProviderActivity); err != nil {
		return result, fmt.Errorf("mark Strava provider activity deleted: %w", err)
	}
	return result, nil
}

func (s WebhookService) recordWebhookIgnored(ctx context.Context, connection repository.ProviderConnection, event WebhookEvent, reason string) (providerimport.ImportResult, error) {
	req := s.webhookImportRequest(connection, event)
	req.IgnoreReason = reason
	return s.Importer.ImportActivity(ctx, req)
}

func (s WebhookService) recordWebhookFailure(ctx context.Context, connection repository.ProviderConnection, event WebhookEvent, cause error) (WebhookResult, error) {
	req := s.webhookImportRequest(connection, event)
	req.FailureReason = cause.Error()
	if IsRetryableWebhookError(cause) {
		req.FailureReason = RetryableWebhookFailurePrefix + ": " + cause.Error()
	}
	result, err := s.Importer.ImportActivity(ctx, req)
	wrapped := fmt.Errorf("fetch Strava webhook activity detail: %w", cause)
	if err != nil {
		wrapped = fmt.Errorf("%w: %v", wrapped, err)
	}
	return WebhookResult{Action: WebhookActionFailed, Event: event, Import: &result}, wrapped
}

func IsRetryableWebhookError(err error) bool {
	return errors.Is(err, ErrRateLimited) || errors.Is(err, ErrTemporaryFailure)
}

func (s WebhookService) webhookImportRequest(connection repository.ProviderConnection, event WebhookEvent) providerimport.ImportRequest {
	return providerimport.ImportRequest{
		AthleteID:            connection.AthleteID,
		ProviderConnectionID: connection.ID,
		ProviderActivity: providerimport.ProviderActivityInput{
			ProviderActivityID: event.ObjectID,
		},
		Delivery: providerimport.DeliveryMetadata{
			EventType:  "strava_webhook_" + event.AspectType,
			DeliveryID: event.EventID,
			ReceivedAt: event.receivedAt(s.now()),
		},
	}
}

func (s WebhookService) connectedConnectionForEvent(ctx context.Context, event WebhookEvent) (repository.ProviderConnection, error) {
	providerUserID := event.providerUserID()
	if providerUserID == "" {
		return repository.ProviderConnection{}, errors.New("strava webhook provider user id is required")
	}
	connections, err := s.Providers.ListProviderConnectionsByStatus(ctx, repository.ProviderConnectionStatusConnected)
	if err != nil {
		return repository.ProviderConnection{}, fmt.Errorf("list connected provider connections: %w", err)
	}
	for _, connection := range connections {
		if connection.Provider == ProviderName && connection.ProviderUserID == providerUserID {
			return connection, nil
		}
	}
	return repository.ProviderConnection{}, fmt.Errorf("connected Strava provider connection not found")
}

func (s WebhookService) validate() error {
	if s.Providers == nil {
		return errors.New("provider store is required")
	}
	if s.Importer.Providers == nil || s.Importer.ImportedActivities == nil {
		return errors.New("provider import service is required")
	}
	if s.Fetcher == nil {
		return errors.New("strava activity fetcher is required")
	}
	if s.Verifier == nil {
		return errors.New("strava webhook verifier is required")
	}
	if s.Deduper == nil {
		return errors.New("strava webhook deduper is required")
	}
	return nil
}

func (s WebhookService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (e WebhookEvent) validate() error {
	if e.EventID == "" {
		return errors.New("strava webhook event id is required")
	}
	if e.ObjectType != "activity" && e.ObjectType != "athlete" {
		return fmt.Errorf("unsupported Strava webhook object type %q", e.ObjectType)
	}
	if e.ObjectID == "" {
		return errors.New("strava webhook object id is required")
	}
	if e.AspectType == "" {
		return errors.New("strava webhook aspect type is required")
	}
	return nil
}

func (e WebhookEvent) providerUserID() string {
	if e.ProviderUserID != "" {
		return e.ProviderUserID
	}
	if e.ProviderUserID2 != "" {
		return e.ProviderUserID2
	}
	if e.ObjectType == "athlete" {
		return e.ObjectID
	}
	return ""
}

func (e WebhookEvent) receivedAt(fallback time.Time) time.Time {
	if e.ReceivedAtUnix <= 0 {
		if e.EventTimeUnix > 0 {
			return time.Unix(e.EventTimeUnix, 0).UTC()
		}
		return fallback
	}
	return time.Unix(e.ReceivedAtUnix, 0).UTC()
}

func (e *WebhookEvent) normalise() {
	if e.EventID == "" && e.SubscriptionID != "" && e.ObjectID != "" && e.AspectType != "" {
		timestamp := e.EventTimeUnix
		if timestamp <= 0 {
			timestamp = e.ReceivedAtUnix
		}
		e.EventID = providerimport.DeterministicID("runthread:strava-webhook-event", e.SubscriptionID, e.ObjectType, e.ObjectID, e.AspectType, strconv.FormatInt(timestamp, 10))
	}
}

func (e *WebhookEvent) UnmarshalJSON(data []byte) error {
	type rawWebhookEvent struct {
		EventID         json.RawMessage `json:"event_id"`
		AspectType      string          `json:"aspect_type"`
		ObjectType      string          `json:"object_type"`
		ObjectID        json.RawMessage `json:"object_id"`
		ProviderUserID  json.RawMessage `json:"owner_id"`
		ProviderUserID2 json.RawMessage `json:"provider_user_id"`
		SubscriptionID  json.RawMessage `json:"subscription_id"`
		EventTimeUnix   int64           `json:"event_time"`
		ReceivedAtUnix  int64           `json:"received_at"`
	}
	var raw rawWebhookEvent
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	e.EventID = webhookString(raw.EventID)
	e.AspectType = raw.AspectType
	e.ObjectType = raw.ObjectType
	e.ObjectID = webhookString(raw.ObjectID)
	e.ProviderUserID = webhookString(raw.ProviderUserID)
	e.ProviderUserID2 = webhookString(raw.ProviderUserID2)
	e.SubscriptionID = webhookString(raw.SubscriptionID)
	e.EventTimeUnix = raw.EventTimeUnix
	e.ReceivedAtUnix = raw.ReceivedAtUnix
	return nil
}

func webhookString(value json.RawMessage) string {
	if len(value) == 0 || string(value) == "null" {
		return ""
	}
	var stringValue string
	if err := json.Unmarshal(value, &stringValue); err == nil {
		return stringValue
	}
	var number json.Number
	if err := json.Unmarshal(value, &number); err == nil {
		return number.String()
	}
	return string(value)
}
