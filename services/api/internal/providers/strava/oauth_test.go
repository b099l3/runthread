package strava

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestStartOAuthCreatesPendingConnectionAndAuthorizationURL(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	tokens := &fakeTokenStore{}
	service := testOAuthService(store, tokens)

	response, err := service.StartOAuth(ctx, StartOAuthRequest{
		AthleteID:   "athlete-1",
		RedirectURI: "runthread://provider/strava/callback",
		Scopes:      []string{"read", "activity:read_all", "read"},
	})
	if err != nil {
		t.Fatalf("StartOAuth returned error: %v", err)
	}

	if response.Connection.ID == "" {
		t.Fatal("expected connection id")
	}
	if response.Connection.Provider != ProviderName {
		t.Fatalf("provider = %q, want strava", response.Connection.Provider)
	}
	if response.Connection.Status != repository.ProviderConnectionStatusPending {
		t.Fatalf("status = %q, want pending", response.Connection.Status)
	}
	if response.State != "state-1" {
		t.Fatalf("state = %q, want state-1", response.State)
	}
	if response.Connection.LastImportCursor != "state-1" {
		t.Fatalf("connection state cursor = %q, want state-1", response.Connection.LastImportCursor)
	}

	parsed, err := url.Parse(response.AuthorizationURL)
	if err != nil {
		t.Fatalf("parse authorization url: %v", err)
	}
	if parsed.Host != "www.strava.com" {
		t.Fatalf("authorization host = %q, want www.strava.com", parsed.Host)
	}
	query := parsed.Query()
	if query.Get("client_id") != "client-1" {
		t.Fatalf("client_id = %q, want client-1", query.Get("client_id"))
	}
	if query.Get("redirect_uri") != "runthread://provider/strava/callback" {
		t.Fatalf("redirect_uri = %q", query.Get("redirect_uri"))
	}
	if query.Get("state") != "state-1" {
		t.Fatalf("state query = %q, want state-1", query.Get("state"))
	}
	if query.Get("scope") != "read,activity:read_all" {
		t.Fatalf("scope = %q, want read,activity:read_all", query.Get("scope"))
	}

	saved, err := store.GetProviderConnection(ctx, response.Connection.ID)
	if err != nil {
		t.Fatalf("GetProviderConnection returned error: %v", err)
	}
	if saved.LastImportCursor != "state-1" {
		t.Fatalf("saved state cursor = %q, want state-1", saved.LastImportCursor)
	}
}

func TestStartOAuthReusesPendingConnectionWithNewState(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	existing := repository.ProviderConnection{
		ID:               "connection-1",
		AthleteID:        "athlete-1",
		Provider:         ProviderName,
		Status:           repository.ProviderConnectionStatusPending,
		LastImportCursor: "old-state",
		CreatedAt:        testOAuthNow().Add(-time.Hour),
		UpdatedAt:        testOAuthNow().Add(-time.Hour),
	}
	if err := store.SaveProviderConnection(ctx, existing); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testOAuthService(store, &fakeTokenStore{})

	response, err := service.StartOAuth(ctx, StartOAuthRequest{
		AthleteID:   "athlete-1",
		RedirectURI: "runthread://provider/strava/callback",
	})
	if err != nil {
		t.Fatalf("StartOAuth returned error: %v", err)
	}

	if response.Connection.ID != existing.ID {
		t.Fatalf("connection id = %q, want %q", response.Connection.ID, existing.ID)
	}
	if response.State != "state-1" {
		t.Fatalf("state = %q, want state-1", response.State)
	}
	if response.Connection.LastImportCursor != "state-1" {
		t.Fatalf("connection state cursor = %q, want state-1", response.Connection.LastImportCursor)
	}
}

func TestCompleteOAuthCallbackStoresTokenReferenceAndConnectsConnection(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	tokens := &fakeTokenStore{reference: "token-ref-1"}
	exchanger := &fakeCodeExchanger{token: validOAuthToken()}
	pending := pendingStravaConnection("connection-1", "athlete-1", "state-1")
	if err := store.SaveProviderConnection(ctx, pending); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testOAuthServiceWithExchanger(store, tokens, exchanger)

	response, err := service.CompleteOAuthCallback(ctx, CompleteOAuthCallbackRequest{
		State:       "state-1",
		Code:        "auth-code-1",
		RedirectURI: "runthread://provider/strava/callback",
	})
	if err != nil {
		t.Fatalf("CompleteOAuthCallback returned error: %v", err)
	}

	if response.Connection.Status != repository.ProviderConnectionStatusConnected {
		t.Fatalf("status = %q, want connected", response.Connection.Status)
	}
	if response.Connection.ProviderUserID != "strava-athlete-1" {
		t.Fatalf("provider user id = %q, want strava-athlete-1", response.Connection.ProviderUserID)
	}
	if response.Connection.TokenReference != "token-ref-1" {
		t.Fatalf("token reference = %q, want token-ref-1", response.Connection.TokenReference)
	}
	if response.Connection.LastImportCursor != "" {
		t.Fatalf("state cursor = %q, want cleared", response.Connection.LastImportCursor)
	}
	if tokens.stored.ProviderConnectionID != "connection-1" {
		t.Fatalf("stored token connection = %q, want connection-1", tokens.stored.ProviderConnectionID)
	}
	if tokens.stored.ProviderUserID != "strava-athlete-1" {
		t.Fatalf("stored token provider user = %q, want strava-athlete-1", tokens.stored.ProviderUserID)
	}
	if exchanger.request.Code != "auth-code-1" {
		t.Fatalf("exchange code = %q, want auth-code-1", exchanger.request.Code)
	}
	if exchanger.request.RedirectURI != "runthread://provider/strava/callback" {
		t.Fatalf("exchange redirect uri = %q", exchanger.request.RedirectURI)
	}

	saved, err := store.GetProviderConnection(ctx, pending.ID)
	if err != nil {
		t.Fatalf("GetProviderConnection returned error: %v", err)
	}
	if saved.Status != repository.ProviderConnectionStatusConnected {
		t.Fatalf("saved status = %q, want connected", saved.Status)
	}
}

func TestCompleteOAuthCallbackRejectsInvalidState(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	if err := store.SaveProviderConnection(ctx, pendingStravaConnection("connection-1", "athlete-1", "state-1")); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testOAuthService(store, &fakeTokenStore{})

	_, err := service.CompleteOAuthCallback(ctx, CompleteOAuthCallbackRequest{
		State: "wrong-state",
		Code:  "auth-code-1",
	})

	assertOAuthError(t, err, "state not found")
}

func TestCompleteOAuthCallbackMarksConnectionErrorWhenTokenStorageFails(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	pending := pendingStravaConnection("connection-1", "athlete-1", "state-1")
	if err := store.SaveProviderConnection(ctx, pending); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testOAuthServiceWithExchanger(store, &fakeTokenStore{err: errors.New("vault unavailable")}, &fakeCodeExchanger{token: validOAuthToken()})

	_, err := service.CompleteOAuthCallback(ctx, CompleteOAuthCallbackRequest{
		State: "state-1",
		Code:  "auth-code-1",
	})

	assertOAuthError(t, err, "store Strava token")
	saved, err := store.GetProviderConnection(ctx, pending.ID)
	if err != nil {
		t.Fatalf("GetProviderConnection returned error: %v", err)
	}
	if saved.Status != repository.ProviderConnectionStatusError {
		t.Fatalf("saved status = %q, want error", saved.Status)
	}
	if !strings.Contains(saved.LastError, "vault unavailable") {
		t.Fatalf("last error = %q, want token storage error", saved.LastError)
	}
}

func TestCompleteOAuthCallbackMarksConnectionErrorWhenCodeExchangeFails(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	pending := pendingStravaConnection("connection-1", "athlete-1", "state-1")
	if err := store.SaveProviderConnection(ctx, pending); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := testOAuthServiceWithExchanger(store, &fakeTokenStore{}, &fakeCodeExchanger{err: errors.New("strava unavailable")})

	_, err := service.CompleteOAuthCallback(ctx, CompleteOAuthCallbackRequest{
		State: "state-1",
		Code:  "auth-code-1",
	})

	assertOAuthError(t, err, "exchange Strava authorization code")
	saved, err := store.GetProviderConnection(ctx, pending.ID)
	if err != nil {
		t.Fatalf("GetProviderConnection returned error: %v", err)
	}
	if saved.Status != repository.ProviderConnectionStatusError {
		t.Fatalf("saved status = %q, want error", saved.Status)
	}
}

func TestStartOAuthRequiresBackendOnlyInputs(t *testing.T) {
	service := OAuthService{Store: repository.NewInMemoryStore(), Tokens: &fakeTokenStore{}}

	_, err := service.StartOAuth(context.Background(), StartOAuthRequest{
		AthleteID:   "athlete-1",
		RedirectURI: "runthread://provider/strava/callback",
	})

	assertOAuthError(t, err, "client id")
}

func testOAuthService(store repository.ProviderStore, tokens TokenStore) OAuthService {
	return testOAuthServiceWithExchanger(store, tokens, &fakeCodeExchanger{token: validOAuthToken()})
}

func testOAuthServiceWithExchanger(store repository.ProviderStore, tokens TokenStore, exchanger CodeExchanger) OAuthService {
	return OAuthService{
		Store:     store,
		Exchanger: exchanger,
		Tokens:    tokens,
		ClientID:  "client-1",
		Now:       testOAuthNow,
		NewState: func() (string, error) {
			return "state-1", nil
		},
	}
}

func validOAuthToken() OAuthToken {
	return OAuthToken{
		ProviderUserID: "strava-athlete-1",
		AccessToken:    "access-token",
		RefreshToken:   "refresh-token",
		Scopes:         []string{"activity:read_all"},
		ExpiresAt:      testOAuthNow().Add(6 * time.Hour),
	}
}

type fakeCodeExchanger struct {
	request OAuthCodeExchangeRequest
	token   OAuthToken
	err     error
}

func (e *fakeCodeExchanger) ExchangeCode(ctx context.Context, req OAuthCodeExchangeRequest) (OAuthToken, error) {
	if err := ctx.Err(); err != nil {
		return OAuthToken{}, err
	}
	e.request = req
	if e.err != nil {
		return OAuthToken{}, e.err
	}
	return e.token, nil
}

func pendingStravaConnection(id string, athleteID string, state string) repository.ProviderConnection {
	return repository.ProviderConnection{
		ID:               id,
		AthleteID:        athleteID,
		Provider:         ProviderName,
		Status:           repository.ProviderConnectionStatusPending,
		LastImportCursor: state,
		CreatedAt:        testOAuthNow(),
		UpdatedAt:        testOAuthNow(),
	}
}

func testOAuthNow() time.Time {
	return time.Date(2026, time.June, 9, 10, 0, 0, 0, time.UTC)
}

type fakeTokenStore struct {
	reference string
	stored    OAuthToken
	err       error
}

func (s *fakeTokenStore) StoreToken(ctx context.Context, token OAuthToken) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if s.err != nil {
		return "", s.err
	}
	s.stored = token
	if s.reference != "" {
		return s.reference, nil
	}
	return "token-ref-default", nil
}

func assertOAuthError(t *testing.T, err error, contains string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected error containing %q, got %q", contains, err.Error())
	}
}
