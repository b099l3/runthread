package strava

import (
	"context"
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/repository"
)

type TokenRepository interface {
	LoadToken(ctx context.Context, reference string) (OAuthToken, error)
	UpdateToken(ctx context.Context, reference string, token OAuthToken) error
}

type TokenRefresher interface {
	RefreshToken(ctx context.Context, refreshToken string) (OAuthToken, error)
}

type TokenManager struct {
	Store            TokenRepository
	Refresher        TokenRefresher
	RefreshThreshold time.Duration
	Now              func() time.Time
}

func (m TokenManager) AccessToken(ctx context.Context, connection repository.ProviderConnection) (string, error) {
	token, err := m.Token(ctx, connection)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (m TokenManager) Token(ctx context.Context, connection repository.ProviderConnection) (OAuthToken, error) {
	if m.Store == nil {
		return OAuthToken{}, fmt.Errorf("strava token store is required")
	}
	if connection.TokenReference == "" {
		return OAuthToken{}, fmt.Errorf("strava token reference is required")
	}
	token, err := m.Store.LoadToken(ctx, connection.TokenReference)
	if err != nil {
		return OAuthToken{}, fmt.Errorf("load Strava token: %w", err)
	}
	if !m.shouldRefresh(token) {
		return token, nil
	}
	if m.Refresher == nil {
		return OAuthToken{}, fmt.Errorf("strava token refresher is required")
	}
	refreshed, err := m.Refresher.RefreshToken(ctx, token.RefreshToken)
	if err != nil {
		return OAuthToken{}, fmt.Errorf("refresh Strava token: %w", err)
	}
	refreshed.ProviderConnectionID = token.ProviderConnectionID
	if refreshed.ProviderUserID == "" {
		refreshed.ProviderUserID = token.ProviderUserID
	}
	if err := m.Store.UpdateToken(ctx, connection.TokenReference, refreshed); err != nil {
		return OAuthToken{}, fmt.Errorf("store refreshed Strava token: %w", err)
	}
	return refreshed, nil
}

func (m TokenManager) shouldRefresh(token OAuthToken) bool {
	if token.ExpiresAt.IsZero() {
		return false
	}
	threshold := m.RefreshThreshold
	if threshold == 0 {
		threshold = time.Minute
	}
	return !token.ExpiresAt.After(m.now().Add(threshold))
}

func (m TokenManager) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now()
}
