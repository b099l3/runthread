package strava

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/providerimport"
	"github.com/runthread/runthread/services/api/internal/repository"
)

var ErrRateLimited = errors.New("strava rate limit")
var ErrTemporaryFailure = errors.New("strava temporary failure")

type ActivityFetcher interface {
	ListBackfillActivities(ctx context.Context, req BackfillListRequest) ([]MockActivitySummary, error)
	FetchActivityDetail(ctx context.Context, req ActivityDetailRequest) (MockActivityPayload, error)
}

type BackfillListRequest struct {
	Connection repository.ProviderConnection
	Since      time.Time
	Until      time.Time
}

type ActivityDetailRequest struct {
	Connection repository.ProviderConnection
	ActivityID string
}

type MockActivitySummary struct {
	ActivityID string
}

type BackfillService struct {
	Providers repository.ProviderStore
	Importer  providerimport.Service
	Fetcher   ActivityFetcher
	Now       func() time.Time
}

type RunBackfillRequest struct {
	ProviderConnectionID string
	Since                time.Time
	Until                time.Time
}

type BackfillStatus string

const (
	BackfillStatusCompleted BackfillStatus = "completed"
	BackfillStatusDeferred  BackfillStatus = "deferred"
	BackfillStatusPartial   BackfillStatus = "partial"
	BackfillStatusFailed    BackfillStatus = "failed"
)

type BackfillResult struct {
	Status   BackfillStatus
	Imported int
	Ignored  int
	Failed   int
	Deferred bool
	Errors   []string
	Imports  []providerimport.ImportResult
}

func (s BackfillService) RunInitialBackfill(ctx context.Context, req RunBackfillRequest) (BackfillResult, error) {
	if err := s.validate(); err != nil {
		return BackfillResult{}, err
	}
	if req.ProviderConnectionID == "" {
		return BackfillResult{}, errors.New("provider connection id is required")
	}

	connection, err := s.Providers.GetProviderConnection(ctx, req.ProviderConnectionID)
	if err != nil {
		return BackfillResult{}, fmt.Errorf("get provider connection: %w", err)
	}
	if connection.Provider != ProviderName {
		return BackfillResult{}, fmt.Errorf("provider connection provider %q is not strava", connection.Provider)
	}
	if connection.Status != repository.ProviderConnectionStatusConnected {
		return BackfillResult{}, fmt.Errorf("strava provider connection must be connected")
	}

	started := connection
	started.Status = repository.ProviderConnectionStatusSyncing
	started.UpdatedAt = s.now()
	if err := s.Providers.SaveProviderConnection(ctx, started); err != nil {
		return BackfillResult{}, fmt.Errorf("mark provider connection syncing: %w", err)
	}

	summaries, err := s.Fetcher.ListBackfillActivities(ctx, BackfillListRequest{
		Connection: connection,
		Since:      req.Since,
		Until:      req.Until,
	})
	if err != nil {
		if errors.Is(err, ErrRateLimited) {
			return s.deferBackfill(ctx, started, err)
		}
		return s.failBackfill(ctx, started, err)
	}

	result := BackfillResult{Status: BackfillStatusCompleted}
	for _, summary := range summaries {
		if summary.ActivityID == "" {
			result.Failed++
			result.Errors = append(result.Errors, "strava activity summary id is required")
			continue
		}
		detail, err := s.Fetcher.FetchActivityDetail(ctx, ActivityDetailRequest{
			Connection: connection,
			ActivityID: summary.ActivityID,
		})
		if err != nil {
			if errors.Is(err, ErrRateLimited) {
				result.Deferred = true
				result.Errors = append(result.Errors, err.Error())
				break
			}
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("fetch Strava activity %s: %v", summary.ActivityID, err))
			continue
		}

		importResult, imported, ignored, err := s.importActivity(ctx, connection, detail)
		result.Imports = append(result.Imports, importResult)
		if imported {
			result.Imported++
		}
		if ignored {
			result.Ignored++
		}
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, err.Error())
		}
	}

	connection.Status = repository.ProviderConnectionStatusConnected
	connection.LastSyncAt = s.now()
	connection.UpdatedAt = s.now()
	if result.Deferred {
		connection.LastError = ErrRateLimited.Error()
		result.Status = BackfillStatusDeferred
	} else if result.Failed > 0 {
		connection.LastError = "one or more Strava backfill activities failed"
		result.Status = BackfillStatusPartial
	} else {
		connection.LastError = ""
	}
	if err := s.Providers.SaveProviderConnection(ctx, connection); err != nil {
		return result, fmt.Errorf("save provider connection after backfill: %w", err)
	}
	return result, nil
}

func (s BackfillService) importActivity(ctx context.Context, connection repository.ProviderConnection, payload MockActivityPayload) (providerimport.ImportResult, bool, bool, error) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return providerimport.ImportResult{}, false, false, fmt.Errorf("marshal mock Strava activity payload: %w", err)
	}

	importRequest := providerimport.ImportRequest{
		AthleteID:            connection.AthleteID,
		ProviderConnectionID: connection.ID,
		ProviderActivity: providerimport.ProviderActivityInput{
			ProviderActivityID:   payload.ActivityID,
			ProviderActivityType: payload.StravaSportType,
			StartedAt:            payload.StartDate,
		},
		RawPayload:  rawPayload,
		PayloadKind: "strava_activity_detail",
		Delivery: providerimport.DeliveryMetadata{
			EventType:  "strava_backfill",
			DeliveryID: providerimport.DeterministicID("runthread:strava-backfill-delivery", connection.ID, payload.ActivityID),
			ReceivedAt: s.now(),
		},
	}

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
	activity.ID = providerimport.DeterministicID("runthread:imported-activity", connection.ID, payload.ActivityID)
	importRequest.ImportedActivity = &activity

	result, err := s.Importer.ImportActivity(ctx, importRequest)
	return result, err == nil, false, err
}

func (s BackfillService) deferBackfill(ctx context.Context, connection repository.ProviderConnection, cause error) (BackfillResult, error) {
	connection.Status = repository.ProviderConnectionStatusConnected
	connection.LastError = cause.Error()
	connection.UpdatedAt = s.now()
	if err := s.Providers.SaveProviderConnection(ctx, connection); err != nil {
		return BackfillResult{}, fmt.Errorf("save deferred provider connection: %w", err)
	}
	return BackfillResult{
		Status:   BackfillStatusDeferred,
		Deferred: true,
		Errors:   []string{cause.Error()},
	}, nil
}

func (s BackfillService) failBackfill(ctx context.Context, connection repository.ProviderConnection, cause error) (BackfillResult, error) {
	connection.Status = repository.ProviderConnectionStatusError
	connection.LastError = cause.Error()
	connection.UpdatedAt = s.now()
	if err := s.Providers.SaveProviderConnection(ctx, connection); err != nil {
		return BackfillResult{}, fmt.Errorf("save failed provider connection: %w", err)
	}
	return BackfillResult{
		Status: BackfillStatusFailed,
		Failed: 1,
		Errors: []string{cause.Error()},
	}, cause
}

func (s BackfillService) validate() error {
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

func (s BackfillService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}
