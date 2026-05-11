package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	postgresdb "github.com/runthread/runthread/services/api/internal/postgres/db"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type ProviderConnectionRepository struct {
	queries postgresdb.Querier
}

var _ repository.ProviderConnectionRepository = (*ProviderConnectionRepository)(nil)

func NewProviderConnectionRepository(db *sql.DB) *ProviderConnectionRepository {
	return &ProviderConnectionRepository{
		queries: postgresdb.New(db),
	}
}

func (r *ProviderConnectionRepository) SaveProviderConnection(ctx context.Context, connection repository.ProviderConnection) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := connection.Validate(); err != nil {
		return fmt.Errorf("invalid provider connection: %w", err)
	}

	updateParams, err := providerConnectionToUpdateParams(connection)
	if err != nil {
		return err
	}
	if _, err := r.queries.UpdateProviderConnection(ctx, updateParams); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("update provider connection: %w", err)
	}

	createParams, err := providerConnectionToCreateParams(connection)
	if err != nil {
		return err
	}
	if _, err := r.queries.CreateProviderConnection(ctx, createParams); err != nil {
		return fmt.Errorf("create provider connection: %w", err)
	}
	return nil
}

func (r *ProviderConnectionRepository) GetProviderConnection(ctx context.Context, id string) (repository.ProviderConnection, error) {
	if err := ctx.Err(); err != nil {
		return repository.ProviderConnection{}, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return repository.ProviderConnection{}, fmt.Errorf("parse provider connection id: %w", err)
	}

	row, err := r.queries.GetProviderConnection(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ProviderConnection{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.ProviderConnection{}, fmt.Errorf("get provider connection: %w", err)
	}
	return providerConnectionFromDB(row), nil
}

func (r *ProviderConnectionRepository) ListProviderConnectionsByAthlete(ctx context.Context, athleteID string) ([]repository.ProviderConnection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	parsedID, err := uuid.Parse(athleteID)
	if err != nil {
		return nil, fmt.Errorf("parse provider connection athlete id: %w", err)
	}
	rows, err := r.queries.ListProviderConnectionsByAthlete(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("list provider connections by athlete: %w", err)
	}
	return providerConnectionsFromDB(rows), nil
}

func (r *ProviderConnectionRepository) ListProviderConnectionsByStatus(ctx context.Context, status repository.ProviderConnectionStatus) ([]repository.ProviderConnection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !status.Valid() {
		return nil, fmt.Errorf("invalid provider connection status %q", status)
	}
	rows, err := r.queries.ListProviderConnectionsByStatus(ctx, string(status))
	if err != nil {
		return nil, fmt.Errorf("list provider connections by status: %w", err)
	}
	return providerConnectionsFromDB(rows), nil
}

type ProviderActivityRepository struct {
	queries postgresdb.Querier
}

var _ repository.ProviderActivityRepository = (*ProviderActivityRepository)(nil)

func NewProviderActivityRepository(db *sql.DB) *ProviderActivityRepository {
	return &ProviderActivityRepository{
		queries: postgresdb.New(db),
	}
}

func (r *ProviderActivityRepository) SaveProviderActivity(ctx context.Context, activity repository.ProviderActivity) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := activity.Validate(); err != nil {
		return fmt.Errorf("invalid provider activity: %w", err)
	}

	updateParams, err := providerActivityToUpdateParams(activity)
	if err != nil {
		return err
	}
	if _, err := r.queries.UpdateProviderActivity(ctx, updateParams); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("update provider activity: %w", err)
	}

	createParams, err := providerActivityToCreateParams(activity)
	if err != nil {
		return err
	}
	if _, err := r.queries.CreateProviderActivity(ctx, createParams); err != nil {
		return fmt.Errorf("create provider activity: %w", err)
	}
	return nil
}

func (r *ProviderActivityRepository) GetProviderActivity(ctx context.Context, id string) (repository.ProviderActivity, error) {
	if err := ctx.Err(); err != nil {
		return repository.ProviderActivity{}, err
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return repository.ProviderActivity{}, fmt.Errorf("parse provider activity id: %w", err)
	}
	row, err := r.queries.GetProviderActivity(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ProviderActivity{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.ProviderActivity{}, fmt.Errorf("get provider activity: %w", err)
	}
	return providerActivityFromDB(row), nil
}

func (r *ProviderActivityRepository) GetProviderActivityByProviderID(ctx context.Context, providerConnectionID string, providerActivityID string) (repository.ProviderActivity, error) {
	if err := ctx.Err(); err != nil {
		return repository.ProviderActivity{}, err
	}
	parsedConnectionID, err := uuid.Parse(providerConnectionID)
	if err != nil {
		return repository.ProviderActivity{}, fmt.Errorf("parse provider activity connection id: %w", err)
	}
	row, err := r.queries.GetProviderActivityByProviderID(ctx, postgresdb.GetProviderActivityByProviderIDParams{
		ProviderConnectionID: parsedConnectionID,
		ProviderActivityID:   providerActivityID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ProviderActivity{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.ProviderActivity{}, fmt.Errorf("get provider activity by provider id: %w", err)
	}
	return providerActivityFromDB(row), nil
}

func (r *ProviderActivityRepository) ListProviderActivitiesByAthlete(ctx context.Context, athleteID string) ([]repository.ProviderActivity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	parsedID, err := uuid.Parse(athleteID)
	if err != nil {
		return nil, fmt.Errorf("parse provider activity athlete id: %w", err)
	}
	rows, err := r.queries.ListProviderActivitiesByAthlete(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("list provider activities by athlete: %w", err)
	}
	return providerActivitiesFromDB(rows), nil
}

func (r *ProviderActivityRepository) ListProviderActivitiesByStatus(ctx context.Context, status repository.ProviderActivityStatus) ([]repository.ProviderActivity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !status.Valid() {
		return nil, fmt.Errorf("invalid provider activity status %q", status)
	}
	rows, err := r.queries.ListProviderActivitiesByStatus(ctx, string(status))
	if err != nil {
		return nil, fmt.Errorf("list provider activities by status: %w", err)
	}
	return providerActivitiesFromDB(rows), nil
}

type ProviderActivityPayloadRepository struct {
	queries postgresdb.Querier
}

var _ repository.ProviderActivityPayloadRepository = (*ProviderActivityPayloadRepository)(nil)

func NewProviderActivityPayloadRepository(db *sql.DB) *ProviderActivityPayloadRepository {
	return &ProviderActivityPayloadRepository{
		queries: postgresdb.New(db),
	}
}

func (r *ProviderActivityPayloadRepository) SaveProviderActivityPayload(ctx context.Context, payload repository.ProviderActivityPayload) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := payload.Validate(); err != nil {
		return fmt.Errorf("invalid provider activity payload: %w", err)
	}

	updateParams, err := providerActivityPayloadToUpdateParams(payload)
	if err != nil {
		return err
	}
	if _, err := r.queries.UpdateProviderActivityPayload(ctx, updateParams); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("update provider activity payload: %w", err)
	}

	createParams, err := providerActivityPayloadToCreateParams(payload)
	if err != nil {
		return err
	}
	if _, err := r.queries.CreateProviderActivityPayload(ctx, createParams); err != nil {
		return fmt.Errorf("create provider activity payload: %w", err)
	}
	return nil
}

func (r *ProviderActivityPayloadRepository) GetProviderActivityPayload(ctx context.Context, id string) (repository.ProviderActivityPayload, error) {
	if err := ctx.Err(); err != nil {
		return repository.ProviderActivityPayload{}, err
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return repository.ProviderActivityPayload{}, fmt.Errorf("parse provider activity payload id: %w", err)
	}
	row, err := r.queries.GetProviderActivityPayload(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ProviderActivityPayload{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.ProviderActivityPayload{}, fmt.Errorf("get provider activity payload: %w", err)
	}
	return providerActivityPayloadFromDB(row), nil
}

func (r *ProviderActivityPayloadRepository) ListProviderActivityPayloads(ctx context.Context, providerActivityID string) ([]repository.ProviderActivityPayload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	parsedID, err := uuid.Parse(providerActivityID)
	if err != nil {
		return nil, fmt.Errorf("parse provider activity payload activity id: %w", err)
	}
	rows, err := r.queries.ListProviderActivityPayloads(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("list provider activity payloads: %w", err)
	}
	return providerActivityPayloadsFromDB(rows), nil
}

type ProviderImportEventRepository struct {
	queries postgresdb.Querier
}

var _ repository.ProviderImportEventRepository = (*ProviderImportEventRepository)(nil)

func NewProviderImportEventRepository(db *sql.DB) *ProviderImportEventRepository {
	return &ProviderImportEventRepository{
		queries: postgresdb.New(db),
	}
}

func (r *ProviderImportEventRepository) SaveProviderImportEvent(ctx context.Context, event repository.ProviderImportEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid provider import event: %w", err)
	}

	updateParams, err := providerImportEventToUpdateParams(event)
	if err != nil {
		return err
	}
	if _, err := r.queries.UpdateProviderImportEventStatus(ctx, updateParams); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("update provider import event: %w", err)
	}

	createParams, err := providerImportEventToCreateParams(event)
	if err != nil {
		return err
	}
	if _, err := r.queries.CreateProviderImportEvent(ctx, createParams); err != nil {
		return fmt.Errorf("create provider import event: %w", err)
	}
	return nil
}

func (r *ProviderImportEventRepository) GetProviderImportEvent(ctx context.Context, id string) (repository.ProviderImportEvent, error) {
	if err := ctx.Err(); err != nil {
		return repository.ProviderImportEvent{}, err
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return repository.ProviderImportEvent{}, fmt.Errorf("parse provider import event id: %w", err)
	}
	row, err := r.queries.GetProviderImportEvent(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ProviderImportEvent{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.ProviderImportEvent{}, fmt.Errorf("get provider import event: %w", err)
	}
	return providerImportEventFromDB(row), nil
}

func (r *ProviderImportEventRepository) ListProviderImportEventsByConnection(ctx context.Context, providerConnectionID string) ([]repository.ProviderImportEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	parsedID, err := uuid.Parse(providerConnectionID)
	if err != nil {
		return nil, fmt.Errorf("parse provider import event connection id: %w", err)
	}
	rows, err := r.queries.ListProviderImportEventsByConnection(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("list provider import events by connection: %w", err)
	}
	return providerImportEventsFromDB(rows), nil
}

func (r *ProviderImportEventRepository) ListProviderImportEventsByStatus(ctx context.Context, status repository.ProviderImportEventStatus) ([]repository.ProviderImportEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !status.Valid() {
		return nil, fmt.Errorf("invalid provider import event status %q", status)
	}
	rows, err := r.queries.ListProviderImportEventsByStatus(ctx, string(status))
	if err != nil {
		return nil, fmt.Errorf("list provider import events by status: %w", err)
	}
	return providerImportEventsFromDB(rows), nil
}
