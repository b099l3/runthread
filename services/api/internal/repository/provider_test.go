package repository

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestProviderPersistenceValidation(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "connection", err: ProviderConnection{}.Validate()},
		{name: "activity", err: ProviderActivity{}.Validate()},
		{name: "payload", err: ProviderActivityPayload{}.Validate()},
		{name: "event", err: ProviderImportEvent{}.Validate()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestInMemoryStoreSavesAndGetsProviderRecords(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	connection := providerConnection("connection-1", "athlete-1", ProviderConnectionStatusConnected)
	activity := providerActivity("provider-activity-1", connection.ID, connection.AthleteID, ProviderActivityStatusReceived)
	payload := providerActivityPayload("payload-1", activity.ID)
	event := providerImportEvent("event-1", connection.ID, activity.ID, ProviderImportEventStatusReceived)

	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("save provider connection: %v", err)
	}
	if err := store.SaveProviderActivity(ctx, activity); err != nil {
		t.Fatalf("save provider activity: %v", err)
	}
	if err := store.SaveProviderActivityPayload(ctx, payload); err != nil {
		t.Fatalf("save provider activity payload: %v", err)
	}
	if err := store.SaveProviderImportEvent(ctx, event); err != nil {
		t.Fatalf("save provider import event: %v", err)
	}

	if got := mustGetProviderConnection(t, store, connection.ID); got.ID != connection.ID {
		t.Fatalf("expected connection id %q, got %q", connection.ID, got.ID)
	}
	if got := mustGetProviderActivity(t, store, activity.ID); got.ID != activity.ID {
		t.Fatalf("expected activity id %q, got %q", activity.ID, got.ID)
	}
	if got := mustGetProviderActivityPayload(t, store, payload.ID); got.ID != payload.ID {
		t.Fatalf("expected payload id %q, got %q", payload.ID, got.ID)
	}
	if got := mustGetProviderImportEvent(t, store, event.ID); got.ID != event.ID {
		t.Fatalf("expected import event id %q, got %q", event.ID, got.ID)
	}
}

func TestInMemoryStoreListsProviderRecords(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	connection := providerConnection("connection-1", "athlete-1", ProviderConnectionStatusConnected)
	otherConnection := providerConnection("connection-2", "athlete-2", ProviderConnectionStatusPending)
	activity := providerActivity("provider-activity-1", connection.ID, connection.AthleteID, ProviderActivityStatusReceived)
	otherActivity := providerActivity("provider-activity-2", otherConnection.ID, otherConnection.AthleteID, ProviderActivityStatusIgnored)
	payload := providerActivityPayload("payload-1", activity.ID)
	event := providerImportEvent("event-1", connection.ID, activity.ID, ProviderImportEventStatusReceived)
	otherEvent := providerImportEvent("event-2", otherConnection.ID, otherActivity.ID, ProviderImportEventStatusProcessed)

	for _, record := range []ProviderConnection{connection, otherConnection} {
		if err := store.SaveProviderConnection(ctx, record); err != nil {
			t.Fatalf("save provider connection: %v", err)
		}
	}
	for _, record := range []ProviderActivity{activity, otherActivity} {
		if err := store.SaveProviderActivity(ctx, record); err != nil {
			t.Fatalf("save provider activity: %v", err)
		}
	}
	if err := store.SaveProviderActivityPayload(ctx, payload); err != nil {
		t.Fatalf("save provider activity payload: %v", err)
	}
	for _, record := range []ProviderImportEvent{event, otherEvent} {
		if err := store.SaveProviderImportEvent(ctx, record); err != nil {
			t.Fatalf("save provider import event: %v", err)
		}
	}

	assertProviderConnectionIDs(t, mustListProviderConnectionsByAthlete(t, store, "athlete-1"), []string{connection.ID})
	assertProviderConnectionIDs(t, mustListProviderConnectionsByStatus(t, store, ProviderConnectionStatusPending), []string{otherConnection.ID})
	assertProviderActivityIDs(t, mustListProviderActivitiesByAthlete(t, store, "athlete-1"), []string{activity.ID})
	assertProviderActivityIDs(t, mustListProviderActivitiesByStatus(t, store, ProviderActivityStatusIgnored), []string{otherActivity.ID})
	assertProviderPayloadIDs(t, mustListProviderActivityPayloads(t, store, activity.ID), []string{payload.ID})
	assertProviderEventIDs(t, mustListProviderImportEventsByConnection(t, store, connection.ID), []string{event.ID})
	assertProviderEventIDs(t, mustListProviderImportEventsByStatus(t, store, ProviderImportEventStatusProcessed), []string{otherEvent.ID})

	byProviderID, err := store.GetProviderActivityByProviderID(ctx, connection.ID, activity.ProviderActivityID)
	if err != nil {
		t.Fatalf("get activity by provider id: %v", err)
	}
	if byProviderID.ID != activity.ID {
		t.Fatalf("expected provider activity %q, got %q", activity.ID, byProviderID.ID)
	}
}

func TestInMemoryStoreReturnsNotFoundForProviderRecords(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	_, err := store.GetProviderConnection(ctx, "missing")
	assertNotFound(t, err)
	_, err = store.GetProviderActivity(ctx, "missing")
	assertNotFound(t, err)
	_, err = store.GetProviderActivityByProviderID(ctx, "missing", "missing")
	assertNotFound(t, err)
	_, err = store.GetProviderActivityPayload(ctx, "missing")
	assertNotFound(t, err)
	_, err = store.GetProviderImportEvent(ctx, "missing")
	assertNotFound(t, err)
}

func TestInMemoryStoreCopiesProviderPayloadBytes(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	payload := providerActivityPayload("payload-1", "provider-activity-1")
	if err := store.SaveProviderActivityPayload(ctx, payload); err != nil {
		t.Fatalf("save payload: %v", err)
	}

	payload.Payload[0] = '['
	stored := mustGetProviderActivityPayload(t, store, payload.ID)
	if string(stored.Payload) != `{"distance":8000}` {
		t.Fatalf("expected stored payload to be isolated, got %s", string(stored.Payload))
	}

	stored.Payload[0] = '['
	again := mustGetProviderActivityPayload(t, store, payload.ID)
	if string(again.Payload) != `{"distance":8000}` {
		t.Fatalf("expected fetched payload to be isolated, got %s", string(again.Payload))
	}
}

func providerConnection(id string, athleteID string, status ProviderConnectionStatus) ProviderConnection {
	now := time.Date(2026, time.June, 1, 9, 0, 0, 0, time.UTC)
	return ProviderConnection{
		ID:             id,
		AthleteID:      athleteID,
		Provider:       "garmin",
		ProviderUserID: "garmin-user-1",
		Status:         status,
		ConnectedAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func providerActivity(id string, connectionID string, athleteID string, status ProviderActivityStatus) ProviderActivity {
	now := time.Date(2026, time.June, 2, 7, 30, 0, 0, time.UTC)
	return ProviderActivity{
		ID:                   id,
		ProviderConnectionID: connectionID,
		AthleteID:            athleteID,
		Provider:             "garmin",
		ProviderActivityID:   "garmin-activity-" + id,
		ProviderActivityType: "running",
		StartedAt:            now,
		Status:               status,
		FirstSeenAt:          now.Add(30 * time.Minute),
		CreatedAt:            now.Add(30 * time.Minute),
		UpdatedAt:            now.Add(30 * time.Minute),
	}
}

func providerActivityPayload(id string, activityID string) ProviderActivityPayload {
	return ProviderActivityPayload{
		ID:                 id,
		ProviderActivityID: activityID,
		Payload:            []byte(`{"distance":8000}`),
		PayloadKind:        "activity",
		ReceivedAt:         time.Date(2026, time.June, 2, 8, 5, 0, 0, time.UTC),
	}
}

func providerImportEvent(id string, connectionID string, activityID string, status ProviderImportEventStatus) ProviderImportEvent {
	return ProviderImportEvent{
		ID:                   id,
		ProviderConnectionID: connectionID,
		ProviderActivityID:   activityID,
		Provider:             "garmin",
		EventType:            "activity_import",
		DeliveryID:           "delivery-" + id,
		Status:               status,
		ReceivedAt:           time.Date(2026, time.June, 2, 8, 10, 0, 0, time.UTC),
	}
}

func mustGetProviderConnection(t *testing.T, store *InMemoryStore, id string) ProviderConnection {
	t.Helper()
	value, err := store.GetProviderConnection(context.Background(), id)
	if err != nil {
		t.Fatalf("get provider connection: %v", err)
	}
	return value
}

func mustGetProviderActivity(t *testing.T, store *InMemoryStore, id string) ProviderActivity {
	t.Helper()
	value, err := store.GetProviderActivity(context.Background(), id)
	if err != nil {
		t.Fatalf("get provider activity: %v", err)
	}
	return value
}

func mustGetProviderActivityPayload(t *testing.T, store *InMemoryStore, id string) ProviderActivityPayload {
	t.Helper()
	value, err := store.GetProviderActivityPayload(context.Background(), id)
	if err != nil {
		t.Fatalf("get provider activity payload: %v", err)
	}
	return value
}

func mustGetProviderImportEvent(t *testing.T, store *InMemoryStore, id string) ProviderImportEvent {
	t.Helper()
	value, err := store.GetProviderImportEvent(context.Background(), id)
	if err != nil {
		t.Fatalf("get provider import event: %v", err)
	}
	return value
}

func mustListProviderConnectionsByAthlete(t *testing.T, store *InMemoryStore, athleteID string) []ProviderConnection {
	t.Helper()
	value, err := store.ListProviderConnectionsByAthlete(context.Background(), athleteID)
	if err != nil {
		t.Fatalf("list provider connections by athlete: %v", err)
	}
	return value
}

func mustListProviderConnectionsByStatus(t *testing.T, store *InMemoryStore, status ProviderConnectionStatus) []ProviderConnection {
	t.Helper()
	value, err := store.ListProviderConnectionsByStatus(context.Background(), status)
	if err != nil {
		t.Fatalf("list provider connections by status: %v", err)
	}
	return value
}

func mustListProviderActivitiesByAthlete(t *testing.T, store *InMemoryStore, athleteID string) []ProviderActivity {
	t.Helper()
	value, err := store.ListProviderActivitiesByAthlete(context.Background(), athleteID)
	if err != nil {
		t.Fatalf("list provider activities by athlete: %v", err)
	}
	return value
}

func mustListProviderActivitiesByStatus(t *testing.T, store *InMemoryStore, status ProviderActivityStatus) []ProviderActivity {
	t.Helper()
	value, err := store.ListProviderActivitiesByStatus(context.Background(), status)
	if err != nil {
		t.Fatalf("list provider activities by status: %v", err)
	}
	return value
}

func mustListProviderActivityPayloads(t *testing.T, store *InMemoryStore, activityID string) []ProviderActivityPayload {
	t.Helper()
	value, err := store.ListProviderActivityPayloads(context.Background(), activityID)
	if err != nil {
		t.Fatalf("list provider activity payloads: %v", err)
	}
	return value
}

func mustListProviderImportEventsByConnection(t *testing.T, store *InMemoryStore, connectionID string) []ProviderImportEvent {
	t.Helper()
	value, err := store.ListProviderImportEventsByConnection(context.Background(), connectionID)
	if err != nil {
		t.Fatalf("list provider import events by connection: %v", err)
	}
	return value
}

func mustListProviderImportEventsByStatus(t *testing.T, store *InMemoryStore, status ProviderImportEventStatus) []ProviderImportEvent {
	t.Helper()
	value, err := store.ListProviderImportEventsByStatus(context.Background(), status)
	if err != nil {
		t.Fatalf("list provider import events by status: %v", err)
	}
	return value
}

func assertNotFound(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func assertProviderConnectionIDs(t *testing.T, values []ProviderConnection, want []string) {
	t.Helper()
	if len(values) != len(want) {
		t.Fatalf("expected %d provider connections, got %d", len(want), len(values))
	}
	for i := range want {
		if values[i].ID != want[i] {
			t.Fatalf("expected provider connection id %q at %d, got %q", want[i], i, values[i].ID)
		}
	}
}

func assertProviderActivityIDs(t *testing.T, values []ProviderActivity, want []string) {
	t.Helper()
	if len(values) != len(want) {
		t.Fatalf("expected %d provider activities, got %d", len(want), len(values))
	}
	for i := range want {
		if values[i].ID != want[i] {
			t.Fatalf("expected provider activity id %q at %d, got %q", want[i], i, values[i].ID)
		}
	}
}

func assertProviderPayloadIDs(t *testing.T, values []ProviderActivityPayload, want []string) {
	t.Helper()
	if len(values) != len(want) {
		t.Fatalf("expected %d provider payloads, got %d", len(want), len(values))
	}
	for i := range want {
		if values[i].ID != want[i] {
			t.Fatalf("expected provider payload id %q at %d, got %q", want[i], i, values[i].ID)
		}
	}
}

func assertProviderEventIDs(t *testing.T, values []ProviderImportEvent, want []string) {
	t.Helper()
	if len(values) != len(want) {
		t.Fatalf("expected %d provider events, got %d", len(want), len(values))
	}
	for i := range want {
		if values[i].ID != want[i] {
			t.Fatalf("expected provider event id %q at %d, got %q", want[i], i, values[i].ID)
		}
	}
}
