package providerimport

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestServiceImportsNormalisedActivity(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := testConnection()
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("save provider connection: %v", err)
	}
	service := testService(t, store)

	result, err := service.ImportActivity(ctx, ImportRequest{
		AthleteID:            connection.AthleteID,
		ProviderConnectionID: connection.ID,
		ProviderActivity: ProviderActivityInput{
			ProviderActivityID:   "garmin-activity-1",
			ProviderActivityType: "running",
			StartedAt:            testTime(1),
		},
		ImportedActivity: testImportedActivity(connection.AthleteID),
		RawPayload:       []byte(`{"activityId":"garmin-activity-1"}`),
		PayloadKind:      "activity",
		Delivery: DeliveryMetadata{
			EventType:  "webhook_activity",
			DeliveryID: "delivery-1",
			ReceivedAt: testTime(2),
		},
	})
	if err != nil {
		t.Fatalf("ImportActivity returned error: %v", err)
	}

	if result.ProviderActivity.Status != repository.ProviderActivityStatusNormalised {
		t.Fatalf("provider activity status = %q, want normalised", result.ProviderActivity.Status)
	}
	if result.ImportedActivity == nil {
		t.Fatal("expected imported activity in result")
	}
	if result.ProviderActivity.ImportedActivityID != result.ImportedActivity.ID {
		t.Fatalf("provider activity imported id = %q, want %q", result.ProviderActivity.ImportedActivityID, result.ImportedActivity.ID)
	}
	if result.ImportEvent.Status != repository.ProviderImportEventStatusProcessed {
		t.Fatalf("import event status = %q, want processed", result.ImportEvent.Status)
	}
	if !result.PayloadStored {
		t.Fatal("expected payload to be stored")
	}

	assertStoredProcessedImport(t, ctx, store, result)
}

func TestServiceUpdatesExistingProviderActivity(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := testConnection()
	existingImportedActivityID := "existing-imported-activity"
	existing := repository.ProviderActivity{
		ID:                   "existing-provider-activity",
		ProviderConnectionID: connection.ID,
		AthleteID:            connection.AthleteID,
		ImportedActivityID:   existingImportedActivityID,
		Provider:             connection.Provider,
		ProviderActivityID:   "garmin-activity-1",
		ProviderActivityType: "old_type",
		StartedAt:            testTime(0),
		Status:               repository.ProviderActivityStatusFailed,
		FirstSeenAt:          testTime(0),
		LastError:            "previous error",
	}
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("save provider connection: %v", err)
	}
	if err := store.SaveProviderActivity(ctx, existing); err != nil {
		t.Fatalf("save existing provider activity: %v", err)
	}

	result, err := testService(t, store).ImportActivity(ctx, ImportRequest{
		ProviderConnectionID: connection.ID,
		ProviderActivity: ProviderActivityInput{
			ProviderActivityID:   existing.ProviderActivityID,
			ProviderActivityType: "running",
			StartedAt:            testTime(1),
		},
		ImportedActivity: testImportedActivity(connection.AthleteID),
	})
	if err != nil {
		t.Fatalf("ImportActivity returned error: %v", err)
	}
	if result.ProviderActivity.ID != existing.ID {
		t.Fatalf("provider activity id = %q, want existing %q", result.ProviderActivity.ID, existing.ID)
	}
	if result.ProviderActivity.Status != repository.ProviderActivityStatusNormalised {
		t.Fatalf("provider activity status = %q, want normalised", result.ProviderActivity.Status)
	}
	if result.ImportedActivity == nil {
		t.Fatal("expected imported activity")
	}
	if result.ImportedActivity.ID != existingImportedActivityID {
		t.Fatalf("imported activity id = %q, want existing %q", result.ImportedActivity.ID, existingImportedActivityID)
	}
}

func TestServiceRecordsIgnoredImport(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := testConnection()
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("save provider connection: %v", err)
	}

	result, err := testService(t, store).ImportActivity(ctx, ImportRequest{
		ProviderConnectionID: connection.ID,
		ProviderActivity: ProviderActivityInput{
			ProviderActivityID:   "garmin-walk-1",
			ProviderActivityType: "walking",
			StartedAt:            testTime(1),
		},
		IgnoreReason: "unsupported activity type for training import",
	})
	if err != nil {
		t.Fatalf("ImportActivity returned error: %v", err)
	}
	if result.ProviderActivity.Status != repository.ProviderActivityStatusIgnored {
		t.Fatalf("provider activity status = %q, want ignored", result.ProviderActivity.Status)
	}
	if result.ImportEvent.Status != repository.ProviderImportEventStatusIgnored {
		t.Fatalf("import event status = %q, want ignored", result.ImportEvent.Status)
	}
	if result.ImportedActivity != nil {
		t.Fatalf("expected no imported activity, got %#v", result.ImportedActivity)
	}
}

func TestServiceRecordsFailedImport(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := testConnection()
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("save provider connection: %v", err)
	}
	invalid := testImportedActivity(connection.AthleteID)
	invalid.Duration = 0

	result, err := testService(t, store).ImportActivity(ctx, ImportRequest{
		ProviderConnectionID: connection.ID,
		ProviderActivity: ProviderActivityInput{
			ProviderActivityID:   "garmin-activity-1",
			ProviderActivityType: "running",
			StartedAt:            testTime(1),
		},
		ImportedActivity: invalid,
	})
	if err == nil {
		t.Fatal("expected failed import error")
	}
	if !strings.Contains(err.Error(), "invalid imported activity") {
		t.Fatalf("expected invalid imported activity error, got %q", err.Error())
	}
	if result.ProviderActivity.Status != repository.ProviderActivityStatusFailed {
		t.Fatalf("provider activity status = %q, want failed", result.ProviderActivity.Status)
	}
	if result.ImportEvent.Status != repository.ProviderImportEventStatusFailed {
		t.Fatalf("import event status = %q, want failed", result.ImportEvent.Status)
	}
}

func TestServiceRecordsExplicitFailedImport(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := testConnection()
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("save provider connection: %v", err)
	}

	result, err := testService(t, store).ImportActivity(ctx, ImportRequest{
		ProviderConnectionID: connection.ID,
		ProviderActivity: ProviderActivityInput{
			ProviderActivityID:   "garmin-activity-1",
			ProviderActivityType: "running",
			StartedAt:            testTime(1),
		},
		FailureReason: "normalise mock Garmin activity: duration seconds must be positive",
	})
	if err == nil {
		t.Fatal("expected failed import error")
	}
	if !strings.Contains(err.Error(), "normalise mock Garmin activity") {
		t.Fatalf("expected normalisation error, got %q", err.Error())
	}
	if result.ProviderActivity.Status != repository.ProviderActivityStatusFailed {
		t.Fatalf("provider activity status = %q, want failed", result.ProviderActivity.Status)
	}
	if result.ImportEvent.Status != repository.ProviderImportEventStatusFailed {
		t.Fatalf("import event status = %q, want failed", result.ImportEvent.Status)
	}
}

func TestServiceRequiresExistingProviderConnection(t *testing.T) {
	_, err := testService(t, repository.NewInMemoryStore()).ImportActivity(context.Background(), ImportRequest{
		ProviderConnectionID: "missing",
		ProviderActivity: ProviderActivityInput{
			ProviderActivityID: "garmin-activity-1",
		},
		ImportedActivity: testImportedActivity("athlete-1"),
	})
	if err == nil {
		t.Fatal("expected missing connection error")
	}
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected repository.ErrNotFound, got %v", err)
	}
}

func assertStoredProcessedImport(t *testing.T, ctx context.Context, store *repository.InMemoryStore, result ImportResult) {
	t.Helper()

	if _, err := store.GetImportedActivity(ctx, result.ImportedActivity.ID); err != nil {
		t.Fatalf("expected saved imported activity: %v", err)
	}
	storedActivity, err := store.GetProviderActivity(ctx, result.ProviderActivity.ID)
	if err != nil {
		t.Fatalf("expected saved provider activity: %v", err)
	}
	if storedActivity.ImportedActivityID != result.ImportedActivity.ID {
		t.Fatalf("stored provider activity imported id = %q, want %q", storedActivity.ImportedActivityID, result.ImportedActivity.ID)
	}
	storedEvent, err := store.GetProviderImportEvent(ctx, result.ImportEvent.ID)
	if err != nil {
		t.Fatalf("expected saved provider import event: %v", err)
	}
	if storedEvent.Status != repository.ProviderImportEventStatusProcessed {
		t.Fatalf("stored event status = %q, want processed", storedEvent.Status)
	}
	payloads, err := store.ListProviderActivityPayloads(ctx, result.ProviderActivity.ID)
	if err != nil {
		t.Fatalf("list provider payloads: %v", err)
	}
	if len(payloads) != 1 {
		t.Fatalf("payload count = %d, want 1", len(payloads))
	}
}

func testService(t *testing.T, store *repository.InMemoryStore) Service {
	t.Helper()

	service, err := NewService(store, store)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	service.Now = func() time.Time {
		return testTime(3)
	}
	return service
}

func testConnection() repository.ProviderConnection {
	return repository.ProviderConnection{
		ID:             "connection-1",
		AthleteID:      "athlete-1",
		Provider:       "garmin",
		ProviderUserID: "garmin-user-1",
		Status:         repository.ProviderConnectionStatusConnected,
		ConnectedAt:    testTime(0),
	}
}

func testImportedActivity(athleteID string) *domain.ImportedActivity {
	return &domain.ImportedActivity{
		ID:              "imported-activity-1",
		AthleteID:       athleteID,
		Type:            domain.ActivityTypeRun,
		StartedAt:       testTime(1),
		Duration:        45 * time.Minute,
		Distance:        domain.Distance{Meters: 9000},
		AveragePace:     domain.Pace{SecondsPerKilometer: 300},
		AverageHeartBPM: 148,
	}
}

func testTime(day int) time.Time {
	return time.Date(2026, time.June, day+1, 9, 0, 0, 0, time.UTC)
}
