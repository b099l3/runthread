package strava

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestTokenManagerReturnsStoredAccessToken(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryTokenStore()
	token := validOAuthToken()
	token.ExpiresAt = testOAuthNow().Add(time.Hour)
	reference, err := store.StoreToken(ctx, token)
	if err != nil {
		t.Fatalf("StoreToken returned error: %v", err)
	}

	accessToken, err := TokenManager{
		Store: store,
		Now:   testOAuthNow,
	}.AccessToken(ctx, stravaConnectionWithToken(reference))
	if err != nil {
		t.Fatalf("AccessToken returned error: %v", err)
	}

	if accessToken != "access-token" {
		t.Fatalf("access token = %q, want access-token", accessToken)
	}
}

func TestTokenManagerRefreshesExpiredToken(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryTokenStore()
	token := validOAuthToken()
	token.ProviderConnectionID = "connection-1"
	token.ExpiresAt = testOAuthNow().Add(-time.Minute)
	reference, err := store.StoreToken(ctx, token)
	if err != nil {
		t.Fatalf("StoreToken returned error: %v", err)
	}
	refresher := &fakeTokenRefresher{
		token: OAuthToken{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresAt:    testOAuthNow().Add(time.Hour),
		},
	}

	accessToken, err := TokenManager{
		Store:     store,
		Refresher: refresher,
		Now:       testOAuthNow,
	}.AccessToken(ctx, stravaConnectionWithToken(reference))
	if err != nil {
		t.Fatalf("AccessToken returned error: %v", err)
	}

	if refresher.refreshToken != "refresh-token" {
		t.Fatalf("refresh token = %q, want old refresh token", refresher.refreshToken)
	}
	if accessToken != "new-access-token" {
		t.Fatalf("access token = %q, want new-access-token", accessToken)
	}
	stored, err := store.LoadToken(ctx, reference)
	if err != nil {
		t.Fatalf("LoadToken returned error: %v", err)
	}
	if stored.ProviderConnectionID != "connection-1" {
		t.Fatalf("provider connection id = %q, want preserved connection id", stored.ProviderConnectionID)
	}
	if stored.ProviderUserID != "strava-athlete-1" {
		t.Fatalf("provider user id = %q, want preserved provider user id", stored.ProviderUserID)
	}
}

func TestTokenManagerRequiresRefresherForExpiredToken(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryTokenStore()
	token := validOAuthToken()
	token.ExpiresAt = testOAuthNow().Add(-time.Minute)
	reference, err := store.StoreToken(ctx, token)
	if err != nil {
		t.Fatalf("StoreToken returned error: %v", err)
	}

	_, err = TokenManager{
		Store: store,
		Now:   testOAuthNow,
	}.AccessToken(ctx, stravaConnectionWithToken(reference))
	if err == nil {
		t.Fatal("expected missing refresher error")
	}
}

func stravaConnectionWithToken(reference string) repository.ProviderConnection {
	return repository.ProviderConnection{
		ID:             "connection-1",
		AthleteID:      "athlete-1",
		Provider:       ProviderName,
		TokenReference: reference,
	}
}

type fakeTokenRefresher struct {
	refreshToken string
	token        OAuthToken
	err          error
}

func (r *fakeTokenRefresher) RefreshToken(ctx context.Context, refreshToken string) (OAuthToken, error) {
	if err := ctx.Err(); err != nil {
		return OAuthToken{}, err
	}
	r.refreshToken = refreshToken
	if r.err != nil {
		return OAuthToken{}, r.err
	}
	if r.token.AccessToken == "" {
		return OAuthToken{}, errors.New("missing token")
	}
	return r.token, nil
}
