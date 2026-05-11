package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/runthread/runthread/services/api/internal/repository"
)

type ProviderStore struct {
	ProviderConnections      repository.ProviderConnectionRepository
	ProviderActivities       repository.ProviderActivityRepository
	ProviderActivityPayloads repository.ProviderActivityPayloadRepository
	ProviderImportEvents     repository.ProviderImportEventRepository
}

var _ repository.ProviderStore = (*ProviderStore)(nil)

func NewProviderStore(db *sql.DB) (*ProviderStore, error) {
	if db == nil {
		return nil, fmt.Errorf("postgres provider store db is required")
	}

	return &ProviderStore{
		ProviderConnections:      NewProviderConnectionRepository(db),
		ProviderActivities:       NewProviderActivityRepository(db),
		ProviderActivityPayloads: NewProviderActivityPayloadRepository(db),
		ProviderImportEvents:     NewProviderImportEventRepository(db),
	}, nil
}

func (s *ProviderStore) SaveProviderConnection(ctx context.Context, connection repository.ProviderConnection) error {
	return s.ProviderConnections.SaveProviderConnection(ctx, connection)
}

func (s *ProviderStore) GetProviderConnection(ctx context.Context, id string) (repository.ProviderConnection, error) {
	return s.ProviderConnections.GetProviderConnection(ctx, id)
}

func (s *ProviderStore) ListProviderConnectionsByAthlete(ctx context.Context, athleteID string) ([]repository.ProviderConnection, error) {
	return s.ProviderConnections.ListProviderConnectionsByAthlete(ctx, athleteID)
}

func (s *ProviderStore) ListProviderConnectionsByStatus(ctx context.Context, status repository.ProviderConnectionStatus) ([]repository.ProviderConnection, error) {
	return s.ProviderConnections.ListProviderConnectionsByStatus(ctx, status)
}

func (s *ProviderStore) SaveProviderActivity(ctx context.Context, activity repository.ProviderActivity) error {
	return s.ProviderActivities.SaveProviderActivity(ctx, activity)
}

func (s *ProviderStore) GetProviderActivity(ctx context.Context, id string) (repository.ProviderActivity, error) {
	return s.ProviderActivities.GetProviderActivity(ctx, id)
}

func (s *ProviderStore) GetProviderActivityByProviderID(ctx context.Context, providerConnectionID string, providerActivityID string) (repository.ProviderActivity, error) {
	return s.ProviderActivities.GetProviderActivityByProviderID(ctx, providerConnectionID, providerActivityID)
}

func (s *ProviderStore) ListProviderActivitiesByAthlete(ctx context.Context, athleteID string) ([]repository.ProviderActivity, error) {
	return s.ProviderActivities.ListProviderActivitiesByAthlete(ctx, athleteID)
}

func (s *ProviderStore) ListProviderActivitiesByStatus(ctx context.Context, status repository.ProviderActivityStatus) ([]repository.ProviderActivity, error) {
	return s.ProviderActivities.ListProviderActivitiesByStatus(ctx, status)
}

func (s *ProviderStore) SaveProviderActivityPayload(ctx context.Context, payload repository.ProviderActivityPayload) error {
	return s.ProviderActivityPayloads.SaveProviderActivityPayload(ctx, payload)
}

func (s *ProviderStore) GetProviderActivityPayload(ctx context.Context, id string) (repository.ProviderActivityPayload, error) {
	return s.ProviderActivityPayloads.GetProviderActivityPayload(ctx, id)
}

func (s *ProviderStore) ListProviderActivityPayloads(ctx context.Context, providerActivityID string) ([]repository.ProviderActivityPayload, error) {
	return s.ProviderActivityPayloads.ListProviderActivityPayloads(ctx, providerActivityID)
}

func (s *ProviderStore) SaveProviderImportEvent(ctx context.Context, event repository.ProviderImportEvent) error {
	return s.ProviderImportEvents.SaveProviderImportEvent(ctx, event)
}

func (s *ProviderStore) GetProviderImportEvent(ctx context.Context, id string) (repository.ProviderImportEvent, error) {
	return s.ProviderImportEvents.GetProviderImportEvent(ctx, id)
}

func (s *ProviderStore) ListProviderImportEventsByConnection(ctx context.Context, providerConnectionID string) ([]repository.ProviderImportEvent, error) {
	return s.ProviderImportEvents.ListProviderImportEventsByConnection(ctx, providerConnectionID)
}

func (s *ProviderStore) ListProviderImportEventsByStatus(ctx context.Context, status repository.ProviderImportEventStatus) ([]repository.ProviderImportEvent, error) {
	return s.ProviderImportEvents.ListProviderImportEventsByStatus(ctx, status)
}
