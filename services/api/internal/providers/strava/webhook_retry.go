package strava

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/runthread/runthread/services/api/internal/providerimport"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type RetryWebhookImportsRequest struct {
	Limit int
}

type RetryWebhookImportsResult struct {
	Attempted int
	Succeeded int
	Failed    int
	Skipped   int
	Imports   []providerimport.ImportResult
	Errors    []string
}

func (s WebhookService) RetryFailedWebhookImports(ctx context.Context, req RetryWebhookImportsRequest) (RetryWebhookImportsResult, error) {
	if err := s.validateRetry(); err != nil {
		return RetryWebhookImportsResult{}, err
	}
	events, err := s.Providers.ListProviderImportEventsByStatus(ctx, repository.ProviderImportEventStatusFailed)
	if err != nil {
		return RetryWebhookImportsResult{}, fmt.Errorf("list failed provider import events: %w", err)
	}

	var result RetryWebhookImportsResult
	for _, importEvent := range events {
		if req.Limit > 0 && result.Attempted >= req.Limit {
			break
		}
		if !isRetryableStravaWebhookImport(importEvent) {
			result.Skipped++
			continue
		}

		retryEvent, connection, err := s.retryEvent(ctx, importEvent)
		if err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, err.Error())
			continue
		}

		result.Attempted++
		importResult, err := s.retryWebhookImport(ctx, connection, retryEvent)
		result.Imports = append(result.Imports, importResult)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		result.Succeeded++
	}
	return result, nil
}

func (s WebhookService) retryWebhookImport(ctx context.Context, connection repository.ProviderConnection, event WebhookEvent) (providerimport.ImportResult, error) {
	detail, err := s.Fetcher.FetchActivityDetail(ctx, ActivityDetailRequest{
		Connection: connection,
		ActivityID: event.ObjectID,
	})
	if err != nil {
		result, recordErr := s.recordWebhookFailure(ctx, connection, event, err)
		if recordErr != nil {
			return providerimport.ImportResult{}, recordErr
		}
		if result.Import == nil {
			return providerimport.ImportResult{}, err
		}
		return *result.Import, err
	}
	importResult, _, _, err := s.importWebhookActivity(ctx, connection, event, detail)
	return importResult, err
}

func (s WebhookService) retryEvent(ctx context.Context, importEvent repository.ProviderImportEvent) (WebhookEvent, repository.ProviderConnection, error) {
	if importEvent.ProviderConnectionID == "" {
		return WebhookEvent{}, repository.ProviderConnection{}, errors.New("retryable Strava webhook import missing provider connection id")
	}
	if importEvent.ProviderActivityID == "" {
		return WebhookEvent{}, repository.ProviderConnection{}, errors.New("retryable Strava webhook import missing provider activity id")
	}
	connection, err := s.Providers.GetProviderConnection(ctx, importEvent.ProviderConnectionID)
	if err != nil {
		return WebhookEvent{}, repository.ProviderConnection{}, fmt.Errorf("get retry provider connection: %w", err)
	}
	activity, err := s.Providers.GetProviderActivity(ctx, importEvent.ProviderActivityID)
	if err != nil {
		return WebhookEvent{}, repository.ProviderConnection{}, fmt.Errorf("get retry provider activity: %w", err)
	}

	aspect := strings.TrimPrefix(importEvent.EventType, "strava_webhook_")
	eventID := importEvent.DeliveryID
	if eventID == "" {
		eventID = importEvent.ID
	}
	return WebhookEvent{
		EventID:        eventID,
		AspectType:     aspect,
		ObjectType:     "activity",
		ObjectID:       activity.ProviderActivityID,
		ProviderUserID: connection.ProviderUserID,
		ReceivedAtUnix: importEvent.ReceivedAt.Unix(),
	}, connection, nil
}

func isRetryableStravaWebhookImport(event repository.ProviderImportEvent) bool {
	if event.Provider != ProviderName {
		return false
	}
	if event.Status != repository.ProviderImportEventStatusFailed {
		return false
	}
	if event.EventType != "strava_webhook_create" && event.EventType != "strava_webhook_update" {
		return false
	}
	return strings.HasPrefix(event.Error, RetryableWebhookFailurePrefix)
}

func (s WebhookService) validateRetry() error {
	if s.Providers == nil {
		return errors.New("provider store is required")
	}
	if s.Importer.Providers == nil || s.Importer.ImportedActivities == nil {
		return errors.New("provider import service is required")
	}
	if s.Fetcher == nil {
		return errors.New("strava activity fetcher is required")
	}
	return nil
}
