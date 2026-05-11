package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	postgresdb "github.com/runthread/runthread/services/api/internal/postgres/db"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func providerConnectionToCreateParams(connection repository.ProviderConnection) (postgresdb.CreateProviderConnectionParams, error) {
	id, athleteID, err := parseRecordAndParentIDs(connection.ID, connection.AthleteID, "provider connection")
	if err != nil {
		return postgresdb.CreateProviderConnectionParams{}, err
	}

	return postgresdb.CreateProviderConnectionParams{
		ID:               id,
		AthleteID:        athleteID,
		Provider:         connection.Provider,
		ProviderUserID:   nullableString(connection.ProviderUserID),
		Status:           string(connection.Status),
		ConnectedAt:      nullableTime(connection.ConnectedAt),
		DisconnectedAt:   nullableTime(connection.DisconnectedAt),
		LastSyncAt:       nullableTime(connection.LastSyncAt),
		LastImportCursor: nullableString(connection.LastImportCursor),
		TokenReference:   nullableString(connection.TokenReference),
		TokenExpiresAt:   nullableTime(connection.TokenExpiresAt),
		LastError:        nullableString(connection.LastError),
	}, nil
}

func providerConnectionToUpdateParams(connection repository.ProviderConnection) (postgresdb.UpdateProviderConnectionParams, error) {
	id, err := uuid.Parse(connection.ID)
	if err != nil {
		return postgresdb.UpdateProviderConnectionParams{}, fmt.Errorf("parse provider connection id: %w", err)
	}

	return postgresdb.UpdateProviderConnectionParams{
		ID:               id,
		ProviderUserID:   nullableString(connection.ProviderUserID),
		Status:           string(connection.Status),
		ConnectedAt:      nullableTime(connection.ConnectedAt),
		DisconnectedAt:   nullableTime(connection.DisconnectedAt),
		LastSyncAt:       nullableTime(connection.LastSyncAt),
		LastImportCursor: nullableString(connection.LastImportCursor),
		TokenReference:   nullableString(connection.TokenReference),
		TokenExpiresAt:   nullableTime(connection.TokenExpiresAt),
		LastError:        nullableString(connection.LastError),
	}, nil
}

func providerConnectionFromDB(row postgresdb.ProviderConnection) repository.ProviderConnection {
	return repository.ProviderConnection{
		ID:               row.ID.String(),
		AthleteID:        row.AthleteID.String(),
		Provider:         row.Provider,
		ProviderUserID:   stringFromNull(row.ProviderUserID),
		Status:           repository.ProviderConnectionStatus(row.Status),
		ConnectedAt:      timeFromNull(row.ConnectedAt),
		DisconnectedAt:   timeFromNull(row.DisconnectedAt),
		LastSyncAt:       timeFromNull(row.LastSyncAt),
		LastImportCursor: stringFromNull(row.LastImportCursor),
		TokenReference:   stringFromNull(row.TokenReference),
		TokenExpiresAt:   timeFromNull(row.TokenExpiresAt),
		LastError:        stringFromNull(row.LastError),
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func providerConnectionsFromDB(rows []postgresdb.ProviderConnection) []repository.ProviderConnection {
	out := make([]repository.ProviderConnection, 0, len(rows))
	for _, row := range rows {
		out = append(out, providerConnectionFromDB(row))
	}
	return out
}

func providerActivityToCreateParams(activity repository.ProviderActivity) (postgresdb.CreateProviderActivityParams, error) {
	id, connectionID, err := parseRecordAndParentIDs(activity.ID, activity.ProviderConnectionID, "provider activity")
	if err != nil {
		return postgresdb.CreateProviderActivityParams{}, err
	}
	athleteID, err := uuid.Parse(activity.AthleteID)
	if err != nil {
		return postgresdb.CreateProviderActivityParams{}, fmt.Errorf("parse provider activity athlete id: %w", err)
	}
	importedActivityID, err := nullableUUID(activity.ImportedActivityID)
	if err != nil {
		return postgresdb.CreateProviderActivityParams{}, fmt.Errorf("parse provider activity imported activity id: %w", err)
	}

	return postgresdb.CreateProviderActivityParams{
		ID:                   id,
		ProviderConnectionID: connectionID,
		AthleteID:            athleteID,
		ImportedActivityID:   importedActivityID,
		Provider:             activity.Provider,
		ProviderActivityID:   activity.ProviderActivityID,
		ProviderActivityType: nullableString(activity.ProviderActivityType),
		StartedAt:            nullableTime(activity.StartedAt),
		Status:               string(activity.Status),
		FirstSeenAt:          activity.FirstSeenAt,
		LastSyncedAt:         nullableTime(activity.LastSyncedAt),
		LastError:            nullableString(activity.LastError),
	}, nil
}

func providerActivityToUpdateParams(activity repository.ProviderActivity) (postgresdb.UpdateProviderActivityParams, error) {
	id, err := uuid.Parse(activity.ID)
	if err != nil {
		return postgresdb.UpdateProviderActivityParams{}, fmt.Errorf("parse provider activity id: %w", err)
	}
	importedActivityID, err := nullableUUID(activity.ImportedActivityID)
	if err != nil {
		return postgresdb.UpdateProviderActivityParams{}, fmt.Errorf("parse provider activity imported activity id: %w", err)
	}

	return postgresdb.UpdateProviderActivityParams{
		ID:                   id,
		ImportedActivityID:   importedActivityID,
		ProviderActivityType: nullableString(activity.ProviderActivityType),
		StartedAt:            nullableTime(activity.StartedAt),
		Status:               string(activity.Status),
		LastSyncedAt:         nullableTime(activity.LastSyncedAt),
		LastError:            nullableString(activity.LastError),
	}, nil
}

func providerActivityFromDB(row postgresdb.ProviderActivity) repository.ProviderActivity {
	return repository.ProviderActivity{
		ID:                   row.ID.String(),
		ProviderConnectionID: row.ProviderConnectionID.String(),
		AthleteID:            row.AthleteID.String(),
		ImportedActivityID:   uuidStringFromNull(row.ImportedActivityID),
		Provider:             row.Provider,
		ProviderActivityID:   row.ProviderActivityID,
		ProviderActivityType: stringFromNull(row.ProviderActivityType),
		StartedAt:            timeFromNull(row.StartedAt),
		Status:               repository.ProviderActivityStatus(row.Status),
		FirstSeenAt:          row.FirstSeenAt,
		LastSyncedAt:         timeFromNull(row.LastSyncedAt),
		LastError:            stringFromNull(row.LastError),
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

func providerActivitiesFromDB(rows []postgresdb.ProviderActivity) []repository.ProviderActivity {
	out := make([]repository.ProviderActivity, 0, len(rows))
	for _, row := range rows {
		out = append(out, providerActivityFromDB(row))
	}
	return out
}

func providerActivityPayloadToCreateParams(payload repository.ProviderActivityPayload) (postgresdb.CreateProviderActivityPayloadParams, error) {
	id, activityID, err := parseRecordAndParentIDs(payload.ID, payload.ProviderActivityID, "provider activity payload")
	if err != nil {
		return postgresdb.CreateProviderActivityPayloadParams{}, err
	}

	return postgresdb.CreateProviderActivityPayloadParams{
		ID:                 id,
		ProviderActivityID: activityID,
		Payload:            json.RawMessage(append([]byte(nil), payload.Payload...)),
		PayloadKind:        payload.PayloadKind,
		ReceivedAt:         payload.ReceivedAt,
	}, nil
}

func providerActivityPayloadToUpdateParams(payload repository.ProviderActivityPayload) (postgresdb.UpdateProviderActivityPayloadParams, error) {
	params, err := providerActivityPayloadToCreateParams(payload)
	if err != nil {
		return postgresdb.UpdateProviderActivityPayloadParams{}, err
	}
	return postgresdb.UpdateProviderActivityPayloadParams{
		ID:                 params.ID,
		ProviderActivityID: params.ProviderActivityID,
		Payload:            params.Payload,
		PayloadKind:        params.PayloadKind,
		ReceivedAt:         params.ReceivedAt,
	}, nil
}

func providerActivityPayloadFromDB(row postgresdb.ProviderActivityPayload) repository.ProviderActivityPayload {
	return repository.ProviderActivityPayload{
		ID:                 row.ID.String(),
		ProviderActivityID: row.ProviderActivityID.String(),
		Payload:            append([]byte(nil), row.Payload...),
		PayloadKind:        row.PayloadKind,
		ReceivedAt:         row.ReceivedAt,
	}
}

func providerActivityPayloadsFromDB(rows []postgresdb.ProviderActivityPayload) []repository.ProviderActivityPayload {
	out := make([]repository.ProviderActivityPayload, 0, len(rows))
	for _, row := range rows {
		out = append(out, providerActivityPayloadFromDB(row))
	}
	return out
}

func providerImportEventToCreateParams(event repository.ProviderImportEvent) (postgresdb.CreateProviderImportEventParams, error) {
	id, err := uuid.Parse(event.ID)
	if err != nil {
		return postgresdb.CreateProviderImportEventParams{}, fmt.Errorf("parse provider import event id: %w", err)
	}
	connectionID, err := nullableUUID(event.ProviderConnectionID)
	if err != nil {
		return postgresdb.CreateProviderImportEventParams{}, fmt.Errorf("parse provider import event connection id: %w", err)
	}
	activityID, err := nullableUUID(event.ProviderActivityID)
	if err != nil {
		return postgresdb.CreateProviderImportEventParams{}, fmt.Errorf("parse provider import event activity id: %w", err)
	}

	return postgresdb.CreateProviderImportEventParams{
		ID:                   id,
		ProviderConnectionID: connectionID,
		ProviderActivityID:   activityID,
		Provider:             event.Provider,
		EventType:            event.EventType,
		DeliveryID:           nullableString(event.DeliveryID),
		Status:               string(event.Status),
		ReceivedAt:           event.ReceivedAt,
		ProcessedAt:          nullableTime(event.ProcessedAt),
		Error:                nullableString(event.Error),
	}, nil
}

func providerImportEventToUpdateParams(event repository.ProviderImportEvent) (postgresdb.UpdateProviderImportEventStatusParams, error) {
	id, err := uuid.Parse(event.ID)
	if err != nil {
		return postgresdb.UpdateProviderImportEventStatusParams{}, fmt.Errorf("parse provider import event id: %w", err)
	}

	return postgresdb.UpdateProviderImportEventStatusParams{
		ID:          id,
		Status:      string(event.Status),
		ProcessedAt: nullableTime(event.ProcessedAt),
		Error:       nullableString(event.Error),
	}, nil
}

func providerImportEventFromDB(row postgresdb.ProviderImportEvent) repository.ProviderImportEvent {
	return repository.ProviderImportEvent{
		ID:                   row.ID.String(),
		ProviderConnectionID: uuidStringFromNull(row.ProviderConnectionID),
		ProviderActivityID:   uuidStringFromNull(row.ProviderActivityID),
		Provider:             row.Provider,
		EventType:            row.EventType,
		DeliveryID:           stringFromNull(row.DeliveryID),
		Status:               repository.ProviderImportEventStatus(row.Status),
		ReceivedAt:           row.ReceivedAt,
		ProcessedAt:          timeFromNull(row.ProcessedAt),
		Error:                stringFromNull(row.Error),
	}
}

func providerImportEventsFromDB(rows []postgresdb.ProviderImportEvent) []repository.ProviderImportEvent {
	out := make([]repository.ProviderImportEvent, 0, len(rows))
	for _, row := range rows {
		out = append(out, providerImportEventFromDB(row))
	}
	return out
}
