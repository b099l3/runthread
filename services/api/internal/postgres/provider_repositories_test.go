package postgres

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	postgresdb "github.com/runthread/runthread/services/api/internal/postgres/db"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestProviderConnectionMapping(t *testing.T) {
	connection := providerConnection()

	createParams, err := providerConnectionToCreateParams(connection)
	if err != nil {
		t.Fatalf("connection to create params: %v", err)
	}
	if createParams.ID.String() != connection.ID {
		t.Fatalf("expected id %q, got %q", connection.ID, createParams.ID.String())
	}
	if createParams.ProviderUserID.String != connection.ProviderUserID || !createParams.ProviderUserID.Valid {
		t.Fatalf("expected provider user id to map")
	}
	if !createParams.ConnectedAt.Valid || !createParams.TokenExpiresAt.Valid {
		t.Fatalf("expected optional timestamps to map")
	}

	updateParams, err := providerConnectionToUpdateParams(connection)
	if err != nil {
		t.Fatalf("connection to update params: %v", err)
	}
	if updateParams.ID.String() != connection.ID {
		t.Fatalf("expected update id %q, got %q", connection.ID, updateParams.ID.String())
	}

	row := postgresdb.ProviderConnection{
		ID:               uuid.MustParse(connection.ID),
		AthleteID:        uuid.MustParse(connection.AthleteID),
		Provider:         connection.Provider,
		ProviderUserID:   sql.NullString{String: connection.ProviderUserID, Valid: true},
		Status:           string(connection.Status),
		ConnectedAt:      sql.NullTime{Time: connection.ConnectedAt, Valid: true},
		DisconnectedAt:   sql.NullTime{},
		LastSyncAt:       sql.NullTime{Time: connection.LastSyncAt, Valid: true},
		LastImportCursor: sql.NullString{String: connection.LastImportCursor, Valid: true},
		TokenReference:   sql.NullString{String: connection.TokenReference, Valid: true},
		TokenExpiresAt:   sql.NullTime{Time: connection.TokenExpiresAt, Valid: true},
		LastError:        sql.NullString{String: connection.LastError, Valid: true},
		CreatedAt:        connection.CreatedAt,
		UpdatedAt:        connection.UpdatedAt,
	}
	mapped := providerConnectionFromDB(row)
	if mapped != connection {
		t.Fatalf("expected mapped connection %#v, got %#v", connection, mapped)
	}
}

func TestProviderActivityMapping(t *testing.T) {
	activity := providerActivity()

	createParams, err := providerActivityToCreateParams(activity)
	if err != nil {
		t.Fatalf("activity to create params: %v", err)
	}
	if createParams.ImportedActivityID.UUID.String() != activity.ImportedActivityID || !createParams.ImportedActivityID.Valid {
		t.Fatalf("expected imported activity id to map")
	}
	if createParams.ProviderActivityType.String != activity.ProviderActivityType || !createParams.ProviderActivityType.Valid {
		t.Fatalf("expected provider activity type to map")
	}

	updateParams, err := providerActivityToUpdateParams(activity)
	if err != nil {
		t.Fatalf("activity to update params: %v", err)
	}
	if updateParams.ID.String() != activity.ID {
		t.Fatalf("expected update id %q, got %q", activity.ID, updateParams.ID.String())
	}

	row := postgresdb.ProviderActivity{
		ID:                   uuid.MustParse(activity.ID),
		ProviderConnectionID: uuid.MustParse(activity.ProviderConnectionID),
		AthleteID:            uuid.MustParse(activity.AthleteID),
		ImportedActivityID:   uuid.NullUUID{UUID: uuid.MustParse(activity.ImportedActivityID), Valid: true},
		Provider:             activity.Provider,
		ProviderActivityID:   activity.ProviderActivityID,
		ProviderActivityType: sql.NullString{String: activity.ProviderActivityType, Valid: true},
		StartedAt:            sql.NullTime{Time: activity.StartedAt, Valid: true},
		Status:               string(activity.Status),
		FirstSeenAt:          activity.FirstSeenAt,
		LastSyncedAt:         sql.NullTime{Time: activity.LastSyncedAt, Valid: true},
		LastError:            sql.NullString{String: activity.LastError, Valid: true},
		CreatedAt:            activity.CreatedAt,
		UpdatedAt:            activity.UpdatedAt,
	}
	mapped := providerActivityFromDB(row)
	if mapped != activity {
		t.Fatalf("expected mapped activity %#v, got %#v", activity, mapped)
	}
}

func TestProviderActivityPayloadMappingCopiesPayload(t *testing.T) {
	payload := providerActivityPayload()

	createParams, err := providerActivityPayloadToCreateParams(payload)
	if err != nil {
		t.Fatalf("payload to create params: %v", err)
	}
	if string(createParams.Payload) != string(payload.Payload) {
		t.Fatalf("expected payload bytes to map")
	}
	payload.Payload[0] = '['
	if string(createParams.Payload) != `{"distance":8000}` {
		t.Fatalf("expected create params payload to be isolated, got %s", string(createParams.Payload))
	}

	updatePayload := providerActivityPayload()
	updateParams, err := providerActivityPayloadToUpdateParams(updatePayload)
	if err != nil {
		t.Fatalf("payload to update params: %v", err)
	}
	if updateParams.ID.String() != updatePayload.ID || string(updateParams.Payload) != string(updatePayload.Payload) {
		t.Fatalf("expected update params to map payload fields")
	}

	row := postgresdb.ProviderActivityPayload{
		ID:                 uuid.MustParse(payload.ID),
		ProviderActivityID: uuid.MustParse(payload.ProviderActivityID),
		Payload:            json.RawMessage(`{"distance":8000}`),
		PayloadKind:        payload.PayloadKind,
		ReceivedAt:         payload.ReceivedAt,
	}
	mapped := providerActivityPayloadFromDB(row)
	if string(mapped.Payload) != `{"distance":8000}` {
		t.Fatalf("expected mapped payload bytes, got %s", string(mapped.Payload))
	}
	mapped.Payload[0] = '['
	if string(row.Payload) != `{"distance":8000}` {
		t.Fatalf("expected db payload to be isolated, got %s", string(row.Payload))
	}
}

func TestProviderImportEventMapping(t *testing.T) {
	event := providerImportEvent()

	createParams, err := providerImportEventToCreateParams(event)
	if err != nil {
		t.Fatalf("event to create params: %v", err)
	}
	if createParams.ProviderConnectionID.UUID.String() != event.ProviderConnectionID || !createParams.ProviderConnectionID.Valid {
		t.Fatalf("expected connection id to map")
	}
	if createParams.DeliveryID.String != event.DeliveryID || !createParams.DeliveryID.Valid {
		t.Fatalf("expected delivery id to map")
	}

	updateParams, err := providerImportEventToUpdateParams(event)
	if err != nil {
		t.Fatalf("event to update params: %v", err)
	}
	if updateParams.ID.String() != event.ID || updateParams.Status != string(event.Status) {
		t.Fatalf("expected update params to map")
	}

	row := postgresdb.ProviderImportEvent{
		ID:                   uuid.MustParse(event.ID),
		ProviderConnectionID: uuid.NullUUID{UUID: uuid.MustParse(event.ProviderConnectionID), Valid: true},
		ProviderActivityID:   uuid.NullUUID{UUID: uuid.MustParse(event.ProviderActivityID), Valid: true},
		Provider:             event.Provider,
		EventType:            event.EventType,
		DeliveryID:           sql.NullString{String: event.DeliveryID, Valid: true},
		Status:               string(event.Status),
		ReceivedAt:           event.ReceivedAt,
		ProcessedAt:          sql.NullTime{Time: event.ProcessedAt, Valid: true},
		Error:                sql.NullString{String: event.Error, Valid: true},
	}
	mapped := providerImportEventFromDB(row)
	if mapped != event {
		t.Fatalf("expected mapped event %#v, got %#v", event, mapped)
	}
}

func TestProviderMappingRejectsInvalidUUID(t *testing.T) {
	connection := providerConnection()
	connection.ID = "not-a-uuid"

	if _, err := providerConnectionToCreateParams(connection); err == nil {
		t.Fatalf("expected invalid provider connection id to fail")
	}

	activity := providerActivity()
	activity.ImportedActivityID = "not-a-uuid"

	if _, err := providerActivityToCreateParams(activity); err == nil {
		t.Fatalf("expected invalid imported activity id to fail")
	}
}

func providerConnection() repository.ProviderConnection {
	now := providerTime(1)
	return repository.ProviderConnection{
		ID:               "11111111-1111-1111-1111-111111111111",
		AthleteID:        "22222222-2222-2222-2222-222222222222",
		Provider:         "garmin",
		ProviderUserID:   "garmin-user-1",
		Status:           repository.ProviderConnectionStatusConnected,
		ConnectedAt:      now,
		LastSyncAt:       now.Add(time.Hour),
		LastImportCursor: "cursor-1",
		TokenReference:   "token-ref-1",
		TokenExpiresAt:   now.Add(24 * time.Hour),
		LastError:        "last error",
		CreatedAt:        now,
		UpdatedAt:        now.Add(time.Minute),
	}
}

func providerActivity() repository.ProviderActivity {
	now := providerTime(2)
	return repository.ProviderActivity{
		ID:                   "33333333-3333-3333-3333-333333333333",
		ProviderConnectionID: "11111111-1111-1111-1111-111111111111",
		AthleteID:            "22222222-2222-2222-2222-222222222222",
		ImportedActivityID:   "44444444-4444-4444-4444-444444444444",
		Provider:             "garmin",
		ProviderActivityID:   "garmin-activity-1",
		ProviderActivityType: "running",
		StartedAt:            now,
		Status:               repository.ProviderActivityStatusNormalised,
		FirstSeenAt:          now.Add(time.Minute),
		LastSyncedAt:         now.Add(2 * time.Minute),
		LastError:            "activity error",
		CreatedAt:            now.Add(time.Minute),
		UpdatedAt:            now.Add(2 * time.Minute),
	}
}

func providerActivityPayload() repository.ProviderActivityPayload {
	return repository.ProviderActivityPayload{
		ID:                 "55555555-5555-5555-5555-555555555555",
		ProviderActivityID: "33333333-3333-3333-3333-333333333333",
		Payload:            []byte(`{"distance":8000}`),
		PayloadKind:        "activity",
		ReceivedAt:         providerTime(3),
	}
}

func providerImportEvent() repository.ProviderImportEvent {
	now := providerTime(4)
	return repository.ProviderImportEvent{
		ID:                   "66666666-6666-6666-6666-666666666666",
		ProviderConnectionID: "11111111-1111-1111-1111-111111111111",
		ProviderActivityID:   "33333333-3333-3333-3333-333333333333",
		Provider:             "garmin",
		EventType:            "activity_import",
		DeliveryID:           "delivery-1",
		Status:               repository.ProviderImportEventStatusProcessed,
		ReceivedAt:           now,
		ProcessedAt:          now.Add(time.Minute),
		Error:                "event error",
	}
}

func providerTime(day int) time.Time {
	return time.Date(2026, time.June, day, 9, 0, 0, 0, time.UTC)
}
