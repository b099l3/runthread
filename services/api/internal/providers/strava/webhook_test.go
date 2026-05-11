package strava

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/providerimport"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestHandleWebhookImportsCreateEvent(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := connectedStravaConnection("connection-1", "athlete-1")
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testWebhookService(t, store, &fakeActivityFetcher{
		details: map[string]MockActivityPayload{
			"strava-activity-1": backfillPayload("strava-activity-1", "Run"),
		},
	}, &fakeWebhookVerifier{}, newFakeWebhookDeduper())

	result, err := service.HandleWebhook(ctx, HandleWebhookRequest{
		Body:      webhookBody(t, webhookEvent("event-1", "create", "strava-activity-1")),
		Signature: "signature-1",
	})
	if err != nil {
		t.Fatalf("HandleWebhook returned error: %v", err)
	}

	if result.Action != WebhookActionImported {
		t.Fatalf("action = %q, want imported", result.Action)
	}
	if result.Import == nil || result.Import.ImportedActivity == nil {
		t.Fatal("expected imported activity")
	}
	if result.Import.ImportedActivity.Type != domain.ActivityTypeRun {
		t.Fatalf("activity type = %q, want run", result.Import.ImportedActivity.Type)
	}
	if result.Import.ImportEvent.EventType != "strava_webhook_create" {
		t.Fatalf("event type = %q, want strava_webhook_create", result.Import.ImportEvent.EventType)
	}
	stored, err := store.GetImportedActivity(ctx, result.Import.ImportedActivity.ID)
	if err != nil {
		t.Fatalf("GetImportedActivity returned error: %v", err)
	}
	if stored.AthleteID != "athlete-1" {
		t.Fatalf("stored athlete = %q, want athlete-1", stored.AthleteID)
	}
}

func TestHandleWebhookImportsUpdateEvent(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := connectedStravaConnection("connection-1", "athlete-1")
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testWebhookService(t, store, &fakeActivityFetcher{
		details: map[string]MockActivityPayload{
			"strava-activity-1": backfillPayload("strava-activity-1", "TrailRun"),
		},
	}, &fakeWebhookVerifier{}, newFakeWebhookDeduper())

	result, err := service.HandleWebhook(ctx, HandleWebhookRequest{
		Body: webhookBody(t, webhookEvent("event-1", "update", "strava-activity-1")),
	})
	if err != nil {
		t.Fatalf("HandleWebhook returned error: %v", err)
	}

	if result.Action != WebhookActionImported {
		t.Fatalf("action = %q, want imported", result.Action)
	}
	if result.Import.ImportedActivity.Type != domain.ActivityTypeTrailRun {
		t.Fatalf("activity type = %q, want trail_run", result.Import.ImportedActivity.Type)
	}
}

func TestHandleWebhookRecordsDeleteEventAsIgnoredProviderImport(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := connectedStravaConnection("connection-1", "athlete-1")
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testWebhookService(t, store, &fakeActivityFetcher{}, &fakeWebhookVerifier{}, newFakeWebhookDeduper())

	result, err := service.HandleWebhook(ctx, HandleWebhookRequest{
		Body: webhookBody(t, webhookEvent("event-1", "delete", "strava-activity-1")),
	})
	if err != nil {
		t.Fatalf("HandleWebhook returned error: %v", err)
	}

	if result.Action != WebhookActionDeleted {
		t.Fatalf("action = %q, want deleted", result.Action)
	}
	if result.Import.ProviderActivity.Status != repository.ProviderActivityStatusIgnored {
		t.Fatalf("provider activity status = %q, want ignored", result.Import.ProviderActivity.Status)
	}
	if result.Import.ImportEvent.Status != repository.ProviderImportEventStatusIgnored {
		t.Fatalf("import event status = %q, want ignored", result.Import.ImportEvent.Status)
	}
}

func TestHandleWebhookSkipsDuplicateEvent(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := connectedStravaConnection("connection-1", "athlete-1")
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	deduper := newFakeWebhookDeduper()
	deduper.seen["event-1"] = true
	service := testWebhookService(t, store, &fakeActivityFetcher{}, &fakeWebhookVerifier{}, deduper)

	result, err := service.HandleWebhook(ctx, HandleWebhookRequest{
		Body: webhookBody(t, webhookEvent("event-1", "create", "strava-activity-1")),
	})
	if err != nil {
		t.Fatalf("HandleWebhook returned error: %v", err)
	}

	if result.Action != WebhookActionDuplicate {
		t.Fatalf("action = %q, want duplicate", result.Action)
	}
	if deduper.marked != 0 {
		t.Fatalf("marked duplicates = %d, want 0", deduper.marked)
	}
}

func TestHandleWebhookRejectsFailedVerification(t *testing.T) {
	service := testWebhookService(t, repository.NewInMemoryStore(), &fakeActivityFetcher{}, &fakeWebhookVerifier{err: errors.New("bad signature")}, newFakeWebhookDeduper())

	_, err := service.HandleWebhook(context.Background(), HandleWebhookRequest{
		Body: webhookBody(t, webhookEvent("event-1", "create", "strava-activity-1")),
	})

	assertWebhookError(t, err, "verify Strava webhook")
}

func TestHandleWebhookRejectsInvalidPayload(t *testing.T) {
	service := testWebhookService(t, repository.NewInMemoryStore(), &fakeActivityFetcher{}, &fakeWebhookVerifier{}, newFakeWebhookDeduper())

	_, err := service.HandleWebhook(context.Background(), HandleWebhookRequest{
		Body: []byte(`{"event_id":"event-1","aspect_type":"create","object_type":"athlete"}`),
	})

	assertWebhookError(t, err, "unsupported Strava webhook object type")
}

func TestHandleWebhookRecordsFetchFailure(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := connectedStravaConnection("connection-1", "athlete-1")
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testWebhookService(t, store, &fakeActivityFetcher{detailErr: errors.New("temporary fetch failure")}, &fakeWebhookVerifier{}, newFakeWebhookDeduper())

	result, err := service.HandleWebhook(ctx, HandleWebhookRequest{
		Body: webhookBody(t, webhookEvent("event-1", "create", "strava-activity-1")),
	})

	assertWebhookError(t, err, "fetch Strava webhook activity detail")
	if result.Action != WebhookActionFailed {
		t.Fatalf("action = %q, want failed", result.Action)
	}
	if result.Import == nil {
		t.Fatal("expected failed import event")
	}
	if result.Import.ImportEvent.Status != repository.ProviderImportEventStatusFailed {
		t.Fatalf("import event status = %q, want failed", result.Import.ImportEvent.Status)
	}
}

func testWebhookService(t *testing.T, store *repository.InMemoryStore, fetcher ActivityFetcher, verifier WebhookVerifier, deduper WebhookDeduper) WebhookService {
	t.Helper()

	importer, err := providerimport.NewService(store, store)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	importer.Now = backfillNow
	return WebhookService{
		Providers: store,
		Importer:  importer,
		Fetcher:   fetcher,
		Verifier:  verifier,
		Deduper:   deduper,
		Now:       backfillNow,
	}
}

func webhookEvent(eventID string, aspect string, objectID string) WebhookEvent {
	return WebhookEvent{
		EventID:        eventID,
		AspectType:     aspect,
		ObjectType:     "activity",
		ObjectID:       objectID,
		ProviderUserID: "strava-athlete-1",
		ReceivedAtUnix: backfillNow().Unix(),
	}
}

func webhookBody(t *testing.T, event WebhookEvent) []byte {
	t.Helper()

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal webhook event: %v", err)
	}
	return body
}

type fakeWebhookVerifier struct {
	err error
}

func (v *fakeWebhookVerifier) VerifyWebhook(ctx context.Context, req VerifyWebhookRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return v.err
}

type fakeWebhookDeduper struct {
	seen   map[string]bool
	marked int
}

func newFakeWebhookDeduper() *fakeWebhookDeduper {
	return &fakeWebhookDeduper{seen: make(map[string]bool)}
}

func (d *fakeWebhookDeduper) SeenWebhookEvent(ctx context.Context, eventID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return d.seen[eventID], nil
}

func (d *fakeWebhookDeduper) MarkWebhookEventSeen(ctx context.Context, eventID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	d.seen[eventID] = true
	d.marked++
	return nil
}

func assertWebhookError(t *testing.T, err error, contains string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected error containing %q, got %q", contains, err.Error())
	}
}
