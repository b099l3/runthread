package strava

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/providerimport"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestRunInitialBackfillImportsActivitiesThroughProviderPipeline(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := connectedStravaConnection("connection-1", "athlete-1")
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testBackfillService(t, store, &fakeActivityFetcher{
		summaries: []MockActivitySummary{{ActivityID: "strava-activity-1"}},
		details: map[string]MockActivityPayload{
			"strava-activity-1": backfillPayload("strava-activity-1", "Run"),
		},
	})

	result, err := service.RunInitialBackfill(ctx, RunBackfillRequest{ProviderConnectionID: connection.ID})
	if err != nil {
		t.Fatalf("RunInitialBackfill returned error: %v", err)
	}

	if result.Status != BackfillStatusCompleted {
		t.Fatalf("status = %q, want completed", result.Status)
	}
	if result.Imported != 1 {
		t.Fatalf("imported = %d, want 1", result.Imported)
	}
	if len(result.Imports) != 1 {
		t.Fatalf("imports = %d, want 1", len(result.Imports))
	}
	imported := result.Imports[0].ImportedActivity
	if imported == nil {
		t.Fatal("expected imported activity")
	}
	if imported.Type != domain.ActivityTypeRun {
		t.Fatalf("activity type = %q, want run", imported.Type)
	}

	stored, err := store.GetImportedActivity(ctx, imported.ID)
	if err != nil {
		t.Fatalf("GetImportedActivity returned error: %v", err)
	}
	if stored.AthleteID != connection.AthleteID {
		t.Fatalf("stored athlete = %q, want %q", stored.AthleteID, connection.AthleteID)
	}

	providerActivity := result.Imports[0].ProviderActivity
	if providerActivity.Status != repository.ProviderActivityStatusNormalised {
		t.Fatalf("provider activity status = %q, want normalised", providerActivity.Status)
	}
	payloads, err := store.ListProviderActivityPayloads(ctx, providerActivity.ID)
	if err != nil {
		t.Fatalf("ListProviderActivityPayloads returned error: %v", err)
	}
	if len(payloads) != 1 {
		t.Fatalf("payloads = %d, want 1", len(payloads))
	}
}

func TestRunInitialBackfillIsIdempotentForSameStravaActivity(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := connectedStravaConnection("connection-1", "athlete-1")
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testBackfillService(t, store, &fakeActivityFetcher{
		summaries: []MockActivitySummary{{ActivityID: "strava-activity-1"}},
		details: map[string]MockActivityPayload{
			"strava-activity-1": backfillPayload("strava-activity-1", "Run"),
		},
	})

	first, err := service.RunInitialBackfill(ctx, RunBackfillRequest{ProviderConnectionID: connection.ID})
	if err != nil {
		t.Fatalf("first RunInitialBackfill returned error: %v", err)
	}
	second, err := service.RunInitialBackfill(ctx, RunBackfillRequest{ProviderConnectionID: connection.ID})
	if err != nil {
		t.Fatalf("second RunInitialBackfill returned error: %v", err)
	}

	if first.Imports[0].ProviderActivity.ID != second.Imports[0].ProviderActivity.ID {
		t.Fatalf("provider activity id changed from %q to %q", first.Imports[0].ProviderActivity.ID, second.Imports[0].ProviderActivity.ID)
	}
	if first.Imports[0].ImportedActivity.ID != second.Imports[0].ImportedActivity.ID {
		t.Fatalf("imported activity id changed from %q to %q", first.Imports[0].ImportedActivity.ID, second.Imports[0].ImportedActivity.ID)
	}

	activities, err := store.ListProviderActivitiesByAthlete(ctx, connection.AthleteID)
	if err != nil {
		t.Fatalf("ListProviderActivitiesByAthlete returned error: %v", err)
	}
	if len(activities) != 1 {
		t.Fatalf("provider activities = %d, want 1", len(activities))
	}
}

func TestRunInitialBackfillIgnoresUnsupportedNonRunActivity(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := connectedStravaConnection("connection-1", "athlete-1")
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testBackfillService(t, store, &fakeActivityFetcher{
		summaries: []MockActivitySummary{{ActivityID: "strava-ride-1"}},
		details: map[string]MockActivityPayload{
			"strava-ride-1": backfillPayload("strava-ride-1", "Ride"),
		},
	})

	result, err := service.RunInitialBackfill(ctx, RunBackfillRequest{ProviderConnectionID: connection.ID})
	if err != nil {
		t.Fatalf("RunInitialBackfill returned error: %v", err)
	}

	if result.Ignored != 1 {
		t.Fatalf("ignored = %d, want 1", result.Ignored)
	}
	if result.Imported != 0 {
		t.Fatalf("imported = %d, want 0", result.Imported)
	}
	if result.Imports[0].ProviderActivity.Status != repository.ProviderActivityStatusIgnored {
		t.Fatalf("provider activity status = %q, want ignored", result.Imports[0].ProviderActivity.Status)
	}
	if result.Imports[0].ImportedActivity != nil {
		t.Fatal("expected no imported activity for ignored non-run")
	}
}

func TestRunInitialBackfillDefersWhenRateLimited(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := connectedStravaConnection("connection-1", "athlete-1")
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testBackfillService(t, store, &fakeActivityFetcher{listErr: ErrRateLimited})

	result, err := service.RunInitialBackfill(ctx, RunBackfillRequest{ProviderConnectionID: connection.ID})
	if err != nil {
		t.Fatalf("RunInitialBackfill returned error: %v", err)
	}

	if result.Status != BackfillStatusDeferred {
		t.Fatalf("status = %q, want deferred", result.Status)
	}
	if !result.Deferred {
		t.Fatal("expected deferred result")
	}
	saved, err := store.GetProviderConnection(ctx, connection.ID)
	if err != nil {
		t.Fatalf("GetProviderConnection returned error: %v", err)
	}
	if saved.Status != repository.ProviderConnectionStatusConnected {
		t.Fatalf("connection status = %q, want connected", saved.Status)
	}
	if !strings.Contains(saved.LastError, "rate limit") {
		t.Fatalf("last error = %q, want rate limit", saved.LastError)
	}
}

func TestRunInitialBackfillRequiresConnectedStravaConnection(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := connectedStravaConnection("connection-1", "athlete-1")
	connection.Status = repository.ProviderConnectionStatusPending
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testBackfillService(t, store, &fakeActivityFetcher{})

	_, err := service.RunInitialBackfill(ctx, RunBackfillRequest{ProviderConnectionID: connection.ID})

	assertBackfillError(t, err, "must be connected")
}

func testBackfillService(t *testing.T, store *repository.InMemoryStore, fetcher ActivityFetcher) BackfillService {
	t.Helper()

	importer, err := providerimport.NewService(store, store)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	importer.Now = backfillNow
	return BackfillService{
		Providers: store,
		Importer:  importer,
		Fetcher:   fetcher,
		Now:       backfillNow,
	}
}

func connectedStravaConnection(id string, athleteID string) repository.ProviderConnection {
	return repository.ProviderConnection{
		ID:             id,
		AthleteID:      athleteID,
		Provider:       ProviderName,
		ProviderUserID: "strava-athlete-1",
		Status:         repository.ProviderConnectionStatusConnected,
		TokenReference: "token-ref-1",
		ConnectedAt:    backfillNow(),
		CreatedAt:      backfillNow(),
		UpdatedAt:      backfillNow(),
	}
}

func backfillPayload(activityID string, sportType string) MockActivityPayload {
	return MockActivityPayload{
		ActivityID:       activityID,
		AthleteID:        "athlete-1",
		StravaSportType:  sportType,
		Name:             "Morning Run",
		StartDate:        time.Date(2026, time.June, 10, 7, 30, 0, 0, time.UTC),
		ElapsedTime:      2700,
		MovingTime:       2640,
		DistanceMeters:   8800,
		AverageHeartRate: 151,
	}
}

func backfillNow() time.Time {
	return time.Date(2026, time.June, 10, 9, 0, 0, 0, time.UTC)
}

type fakeActivityFetcher struct {
	summaries []MockActivitySummary
	details   map[string]MockActivityPayload
	listErr   error
	detailErr error
}

func (f *fakeActivityFetcher) ListBackfillActivities(ctx context.Context, req BackfillListRequest) ([]MockActivitySummary, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if f.listErr != nil {
		return nil, f.listErr
	}
	return append([]MockActivitySummary(nil), f.summaries...), nil
}

func (f *fakeActivityFetcher) FetchActivityDetail(ctx context.Context, req ActivityDetailRequest) (MockActivityPayload, error) {
	if err := ctx.Err(); err != nil {
		return MockActivityPayload{}, err
	}
	if f.detailErr != nil {
		return MockActivityPayload{}, f.detailErr
	}
	payload, ok := f.details[req.ActivityID]
	if !ok {
		return MockActivityPayload{}, errors.New("activity detail not found")
	}
	return payload, nil
}

func assertBackfillError(t *testing.T, err error, contains string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected error containing %q, got %q", contains, err.Error())
	}
}
