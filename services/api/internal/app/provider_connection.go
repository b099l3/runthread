package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/runthread/runthread/services/api/internal/repository"
)

const (
	ProviderGarmin = "garmin"
)

type ProviderConnectionService struct {
	Store repository.ProviderStore
	Now   func() time.Time
}

type GetProviderConnectionStatusRequest struct {
	AthleteID            string
	Provider             string
	ProviderConnectionID string
}

type GetProviderConnectionStatusResponse struct {
	Connection    repository.ProviderConnection
	HasConnection bool
}

type StartProviderConnectionRequest struct {
	AthleteID   string
	Provider    string
	RedirectURI string
}

type StartProviderConnectionResponse struct {
	Connection       repository.ProviderConnection
	AuthorizationURL string
	State            string
	OAuthReady       bool
}

func (s ProviderConnectionService) GetProviderConnectionStatus(ctx context.Context, req GetProviderConnectionStatusRequest) (GetProviderConnectionStatusResponse, error) {
	if err := s.validateProviderRequest(req.AthleteID, req.Provider); err != nil {
		return GetProviderConnectionStatusResponse{}, err
	}
	if req.ProviderConnectionID != "" {
		connection, err := s.Store.GetProviderConnection(ctx, req.ProviderConnectionID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return GetProviderConnectionStatusResponse{
					HasConnection: false,
				}, nil
			}
			return GetProviderConnectionStatusResponse{}, fmt.Errorf("get provider connection: %w", err)
		}
		if connection.AthleteID != req.AthleteID {
			return GetProviderConnectionStatusResponse{}, fmt.Errorf("provider connection does not belong to athlete")
		}
		if connection.Provider != req.Provider {
			return GetProviderConnectionStatusResponse{}, fmt.Errorf("provider connection provider %q does not match request provider %q", connection.Provider, req.Provider)
		}
		return GetProviderConnectionStatusResponse{
			Connection:    connection,
			HasConnection: true,
		}, nil
	}

	connection, found, err := s.bestConnection(ctx, req.AthleteID, req.Provider)
	if err != nil {
		return GetProviderConnectionStatusResponse{}, err
	}
	return GetProviderConnectionStatusResponse{
		Connection:    connection,
		HasConnection: found,
	}, nil
}

func (s ProviderConnectionService) StartProviderConnection(ctx context.Context, req StartProviderConnectionRequest) (StartProviderConnectionResponse, error) {
	if err := s.validateProviderRequest(req.AthleteID, req.Provider); err != nil {
		return StartProviderConnectionResponse{}, err
	}

	pending, found, err := s.pendingConnection(ctx, req.AthleteID, req.Provider)
	if err != nil {
		return StartProviderConnectionResponse{}, err
	}
	if found {
		return StartProviderConnectionResponse{
			Connection: pending,
			OAuthReady: false,
		}, nil
	}

	now := s.now()
	connection := repository.ProviderConnection{
		ID:        uuid.NewString(),
		AthleteID: req.AthleteID,
		Provider:  req.Provider,
		Status:    repository.ProviderConnectionStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.Store.SaveProviderConnection(ctx, connection); err != nil {
		return StartProviderConnectionResponse{}, fmt.Errorf("save provider connection: %w", err)
	}
	return StartProviderConnectionResponse{
		Connection: connection,
		OAuthReady: false,
	}, nil
}

func (s ProviderConnectionService) validateProviderRequest(athleteID string, provider string) error {
	if s.Store == nil {
		return fmt.Errorf("provider store is required")
	}
	if athleteID == "" {
		return fmt.Errorf("athlete id is required")
	}
	if provider == "" {
		return fmt.Errorf("provider is required")
	}
	if provider != ProviderGarmin {
		return fmt.Errorf("unsupported provider %q", provider)
	}
	return nil
}

func (s ProviderConnectionService) pendingConnection(ctx context.Context, athleteID string, provider string) (repository.ProviderConnection, bool, error) {
	connections, err := s.Store.ListProviderConnectionsByAthlete(ctx, athleteID)
	if err != nil {
		return repository.ProviderConnection{}, false, fmt.Errorf("list provider connections: %w", err)
	}
	for _, connection := range connections {
		if connection.Provider == provider && connection.Status == repository.ProviderConnectionStatusPending {
			return connection, true, nil
		}
	}
	return repository.ProviderConnection{}, false, nil
}

func (s ProviderConnectionService) bestConnection(ctx context.Context, athleteID string, provider string) (repository.ProviderConnection, bool, error) {
	connections, err := s.Store.ListProviderConnectionsByAthlete(ctx, athleteID)
	if err != nil {
		return repository.ProviderConnection{}, false, fmt.Errorf("list provider connections: %w", err)
	}

	var best repository.ProviderConnection
	bestRank := 0
	for _, connection := range connections {
		if connection.Provider != provider {
			continue
		}
		rank := providerConnectionStatusRank(connection.Status)
		if rank > bestRank {
			best = connection
			bestRank = rank
		}
	}
	return best, bestRank > 0, nil
}

func providerConnectionStatusRank(status repository.ProviderConnectionStatus) int {
	switch status {
	case repository.ProviderConnectionStatusConnected:
		return 5
	case repository.ProviderConnectionStatusSyncing:
		return 4
	case repository.ProviderConnectionStatusPending:
		return 3
	case repository.ProviderConnectionStatusError:
		return 2
	case repository.ProviderConnectionStatusDisconnected:
		return 1
	default:
		return 0
	}
}

func (s ProviderConnectionService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}
