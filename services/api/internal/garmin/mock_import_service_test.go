package garmin

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestMockImportServiceImportsAndLinksActivity(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	service := newTestMockImportService(t, store)

	result, err := service.ImportActivity(ctx, MockImportRequest{
		Connection: testProviderConnection(),
		Payload:    testMockPayload(),
		RawPayload: []byte(`{"activityId":"garmin-activity-1"}`),
		ReceivedAt: testTime(1),
		DeliveryID: "delivery-1",
	})
	if err != nil {
		t.Fatalf("ImportActivity returned error: %v", err)
	}

	if !result.PayloadStored {
		t.Fatal("expected raw payload to be stored")
	}
	if result.ProviderActivity.Status != repository.ProviderActivityStatusNormalised {
		t.Fatalf("provider activity status = %q, want normalised", result.ProviderActivity.Status)
	}
	if result.ProviderActivity.ImportedActivityID == "" {
		t.Fatal("expected provider activity to link imported activity")
	}
	if result.ImportedActivity.ID != result.ProviderActivity.ImportedActivityID {
		t.Fatalf("imported activity link = %q, want %q", result.ProviderActivity.ImportedActivityID, result.ImportedActivity.ID)
	}
	if result.ImportEvent.Status != repository.ProviderImportEventStatusProcessed {
		t.Fatalf("import event status = %q, want processed", result.ImportEvent.Status)
	}

	storedActivity, err := store.GetProviderActivity(ctx, result.ProviderActivity.ID)
	if err != nil {
		t.Fatalf("get provider activity: %v", err)
	}
	if storedActivity.ImportedActivityID != result.ImportedActivity.ID {
		t.Fatalf("stored imported activity id = %q, want %q", storedActivity.ImportedActivityID, result.ImportedActivity.ID)
	}

	payloads, err := store.ListProviderActivityPayloads(ctx, result.ProviderActivity.ID)
	if err != nil {
		t.Fatalf("list provider activity payloads: %v", err)
	}
	if len(payloads) != 1 {
		t.Fatalf("payload count = %d, want 1", len(payloads))
	}

	imported, err := store.GetImportedActivity(ctx, result.ImportedActivity.ID)
	if err != nil {
		t.Fatalf("get imported activity: %v", err)
	}
	if imported.ID != result.ImportedActivity.ID {
		t.Fatalf("stored imported activity id = %q, want %q", imported.ID, result.ImportedActivity.ID)
	}
}

func TestMockImportServiceSkipsPayloadWhenRawPayloadNotProvided(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	service := newTestMockImportService(t, store)

	result, err := service.ImportActivity(ctx, MockImportRequest{
		Connection: testProviderConnection(),
		Payload:    testMockPayload(),
		ReceivedAt: testTime(1),
	})
	if err != nil {
		t.Fatalf("ImportActivity returned error: %v", err)
	}
	if result.PayloadStored {
		t.Fatal("expected payload storage to be skipped")
	}

	payloads, err := store.ListProviderActivityPayloads(ctx, result.ProviderActivity.ID)
	if err != nil {
		t.Fatalf("list provider activity payloads: %v", err)
	}
	if len(payloads) != 0 {
		t.Fatalf("payload count = %d, want 0", len(payloads))
	}
}

func TestMockImportServiceRecordsFailedImport(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	service := newTestMockImportService(t, store)
	payload := testMockPayload()
	payload.DurationSeconds = 0

	result, err := service.ImportActivity(ctx, MockImportRequest{
		Connection: testProviderConnection(),
		Payload:    payload,
		ReceivedAt: testTime(1),
	})
	if err == nil {
		t.Fatal("expected import error")
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

	storedEvents, err := store.ListProviderImportEventsByStatus(ctx, repository.ProviderImportEventStatusFailed)
	if err != nil {
		t.Fatalf("list failed events: %v", err)
	}
	if len(storedEvents) != 1 {
		t.Fatalf("failed event count = %d, want 1", len(storedEvents))
	}
}

func TestMockImportServiceRejectsMismatchedAthlete(t *testing.T) {
	store := repository.NewInMemoryStore()
	service := newTestMockImportService(t, store)
	payload := testMockPayload()
	payload.AthleteID = "other-athlete"

	_, err := service.ImportActivity(context.Background(), MockImportRequest{
		Connection: testProviderConnection(),
		Payload:    payload,
	})
	if err == nil {
		t.Fatal("expected athlete mismatch error")
	}
	if !strings.Contains(err.Error(), "does not match connection athlete") {
		t.Fatalf("expected athlete mismatch error, got %q", err.Error())
	}
}

func newTestMockImportService(t *testing.T, store *repository.InMemoryStore) *MockImportService {
	t.Helper()

	service, err := NewMockImportService(store, store, WithMockImportClock(func() time.Time {
		return testTime(2)
	}))
	if err != nil {
		t.Fatalf("NewMockImportService returned error: %v", err)
	}
	return service
}

func testProviderConnection() repository.ProviderConnection {
	return repository.ProviderConnection{
		ID:             "11111111-1111-1111-1111-111111111111",
		AthleteID:      "athlete-1",
		Provider:       ProviderNameGarmin,
		ProviderUserID: "garmin-user-1",
		Status:         repository.ProviderConnectionStatusConnected,
		ConnectedAt:    testTime(0),
	}
}

func testMockPayload() MockActivityPayload {
	return MockActivityPayload{
		ActivityID:         "garmin-activity-1",
		AthleteID:          "athlete-1",
		GarminActivityType: "running",
		StartTime:          testTime(1),
		DurationSeconds:    2700,
		DistanceMeters:     9000,
		AverageHeartRate:   148,
	}
}

func testTime(day int) time.Time {
	return time.Date(2026, time.June, day+1, 9, 0, 0, 0, time.UTC)
}
