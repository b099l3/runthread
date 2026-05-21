package app

import (
	"context"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestStartProviderConnectionCreatesPendingGarminConnection(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	service := ProviderConnectionService{
		Store: store,
		Now:   providerConnectionNow,
	}

	response, err := service.StartProviderConnection(ctx, StartProviderConnectionRequest{
		AthleteID:   "athlete-1",
		Provider:    ProviderGarmin,
		RedirectURI: "runthread://provider/garmin/callback",
	})
	if err != nil {
		t.Fatalf("StartProviderConnection returned error: %v", err)
	}

	if response.Connection.ID == "" {
		t.Fatal("expected connection id")
	}
	if response.Connection.Status != repository.ProviderConnectionStatusPending {
		t.Fatalf("connection status = %q, want pending", response.Connection.Status)
	}
	if response.OAuthReady {
		t.Fatal("expected oauth ready false")
	}
	if response.AuthorizationURL != "" {
		t.Fatalf("authorization url = %q, want empty placeholder", response.AuthorizationURL)
	}

	saved, err := store.GetProviderConnection(ctx, response.Connection.ID)
	if err != nil {
		t.Fatalf("GetProviderConnection returned error: %v", err)
	}
	if saved.AthleteID != "athlete-1" {
		t.Fatalf("saved athlete id = %q, want athlete-1", saved.AthleteID)
	}
}

func TestStartProviderConnectionCreatesPendingStravaConnection(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	service := ProviderConnectionService{
		Store: store,
		Now:   providerConnectionNow,
	}

	response, err := service.StartProviderConnection(ctx, StartProviderConnectionRequest{
		AthleteID:   "athlete-1",
		Provider:    ProviderStrava,
		RedirectURI: "runthread://provider/strava/callback",
	})
	if err != nil {
		t.Fatalf("StartProviderConnection returned error: %v", err)
	}

	if response.Connection.Provider != ProviderStrava {
		t.Fatalf("connection provider = %q, want strava", response.Connection.Provider)
	}
	if response.Connection.Status != repository.ProviderConnectionStatusPending {
		t.Fatalf("connection status = %q, want pending", response.Connection.Status)
	}
	if response.OAuthReady {
		t.Fatal("expected oauth ready false until real Strava OAuth is wired")
	}
}

func TestStartProviderConnectionDelegatesConfiguredStravaStarter(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	starter := &fakeProviderConnectionStarter{
		response: StartProviderConnectionResponse{
			Connection: repository.ProviderConnection{
				ID:        "connection-1",
				AthleteID: "athlete-1",
				Provider:  ProviderStrava,
				Status:    repository.ProviderConnectionStatusPending,
			},
			AuthorizationURL: "https://www.strava.com/oauth/authorize?state=state-1",
			State:            "state-1",
			OAuthReady:       true,
		},
	}
	service := ProviderConnectionService{
		Store: store,
		ProviderStarters: map[string]ProviderConnectionStarter{
			ProviderStrava: starter,
		},
	}

	response, err := service.StartProviderConnection(ctx, StartProviderConnectionRequest{
		AthleteID:   "athlete-1",
		Provider:    ProviderStrava,
		RedirectURI: "runthread://provider/strava/callback",
	})
	if err != nil {
		t.Fatalf("StartProviderConnection returned error: %v", err)
	}

	if starter.request.Provider != ProviderStrava {
		t.Fatalf("starter provider = %q, want strava", starter.request.Provider)
	}
	if !response.OAuthReady {
		t.Fatal("OAuthReady = false, want true")
	}
	if response.AuthorizationURL == "" {
		t.Fatal("AuthorizationURL is empty")
	}
}

func TestStartProviderConnectionReusesPendingConnection(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	existing := providerConnectionRecord("connection-1", "athlete-1", repository.ProviderConnectionStatusPending)
	if err := store.SaveProviderConnection(ctx, existing); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := ProviderConnectionService{Store: store, Now: providerConnectionNow}

	response, err := service.StartProviderConnection(ctx, StartProviderConnectionRequest{
		AthleteID: "athlete-1",
		Provider:  ProviderGarmin,
	})
	if err != nil {
		t.Fatalf("StartProviderConnection returned error: %v", err)
	}

	if response.Connection.ID != existing.ID {
		t.Fatalf("connection id = %q, want existing %q", response.Connection.ID, existing.ID)
	}
}

func TestGetProviderConnectionStatusReturnsExistingConnection(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	existing := providerConnectionRecord("connection-1", "athlete-1", repository.ProviderConnectionStatusConnected)
	if err := store.SaveProviderConnection(ctx, existing); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := ProviderConnectionService{Store: store}

	response, err := service.GetProviderConnectionStatus(ctx, GetProviderConnectionStatusRequest{
		AthleteID: "athlete-1",
		Provider:  ProviderGarmin,
	})
	if err != nil {
		t.Fatalf("GetProviderConnectionStatus returned error: %v", err)
	}

	if !response.HasConnection {
		t.Fatal("expected connection")
	}
	if response.Connection.ID != existing.ID {
		t.Fatalf("connection id = %q, want %q", response.Connection.ID, existing.ID)
	}
}

func TestGetProviderConnectionStatusReturnsNoConnection(t *testing.T) {
	service := ProviderConnectionService{Store: repository.NewInMemoryStore()}

	response, err := service.GetProviderConnectionStatus(context.Background(), GetProviderConnectionStatusRequest{
		AthleteID: "athlete-1",
		Provider:  ProviderGarmin,
	})
	if err != nil {
		t.Fatalf("GetProviderConnectionStatus returned error: %v", err)
	}

	if response.HasConnection {
		t.Fatal("expected no connection")
	}
	if response.Connection.ID != "" {
		t.Fatalf("connection id = %q, want empty", response.Connection.ID)
	}
}

func TestDisconnectProviderConnectionMarksConnectionDisconnected(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	existing := providerConnectionRecord("connection-1", "athlete-1", repository.ProviderConnectionStatusConnected)
	existing.Provider = ProviderStrava
	existing.TokenReference = "token-reference-1"
	existing.TokenExpiresAt = providerConnectionNow().Add(time.Hour)
	existing.LastImportCursor = "cursor-1"
	existing.LastError = "old error"
	if err := store.SaveProviderConnection(ctx, existing); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := ProviderConnectionService{Store: store, Now: providerConnectionNow}

	response, err := service.DisconnectProviderConnection(ctx, DisconnectProviderConnectionRequest{
		AthleteID:            "athlete-1",
		Provider:             ProviderStrava,
		ProviderConnectionID: "connection-1",
	})
	if err != nil {
		t.Fatalf("DisconnectProviderConnection returned error: %v", err)
	}

	if response.Connection.Status != repository.ProviderConnectionStatusDisconnected {
		t.Fatalf("connection status = %q, want disconnected", response.Connection.Status)
	}
	if response.Connection.DisconnectedAt != providerConnectionNow() {
		t.Fatalf("disconnected at = %v, want %v", response.Connection.DisconnectedAt, providerConnectionNow())
	}
	if response.Connection.TokenReference != "" {
		t.Fatalf("token reference = %q, want empty", response.Connection.TokenReference)
	}
	if !response.Connection.TokenExpiresAt.IsZero() {
		t.Fatalf("token expires at = %v, want zero", response.Connection.TokenExpiresAt)
	}
	if response.Connection.LastImportCursor != "" {
		t.Fatalf("last import cursor = %q, want empty", response.Connection.LastImportCursor)
	}
	if response.Connection.LastError != "" {
		t.Fatalf("last error = %q, want empty", response.Connection.LastError)
	}
}

func TestDisconnectProviderConnectionRejectsOtherAthleteConnection(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	existing := providerConnectionRecord("connection-1", "athlete-2", repository.ProviderConnectionStatusConnected)
	if err := store.SaveProviderConnection(ctx, existing); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	service := ProviderConnectionService{Store: store, Now: providerConnectionNow}

	_, err := service.DisconnectProviderConnection(ctx, DisconnectProviderConnectionRequest{
		AthleteID:            "athlete-1",
		Provider:             ProviderGarmin,
		ProviderConnectionID: "connection-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProviderConnectionServiceRejectsUnsupportedProvider(t *testing.T) {
	service := ProviderConnectionService{Store: repository.NewInMemoryStore()}

	_, err := service.StartProviderConnection(context.Background(), StartProviderConnectionRequest{
		AthleteID: "athlete-1",
		Provider:  "coros",
	})
	if err == nil {
		t.Fatal("expected unsupported provider error")
	}
}

func providerConnectionRecord(id string, athleteID string, status repository.ProviderConnectionStatus) repository.ProviderConnection {
	return repository.ProviderConnection{
		ID:          id,
		AthleteID:   athleteID,
		Provider:    ProviderGarmin,
		Status:      status,
		ConnectedAt: providerConnectionNow(),
		CreatedAt:   providerConnectionNow(),
		UpdatedAt:   providerConnectionNow(),
	}
}

func providerConnectionNow() time.Time {
	return time.Date(2026, time.June, 5, 9, 0, 0, 0, time.UTC)
}

type fakeProviderConnectionStarter struct {
	request  StartProviderConnectionRequest
	response StartProviderConnectionResponse
	err      error
}

func (s *fakeProviderConnectionStarter) StartProviderConnection(ctx context.Context, req StartProviderConnectionRequest) (StartProviderConnectionResponse, error) {
	s.request = req
	if s.err != nil {
		return StartProviderConnectionResponse{}, s.err
	}
	return s.response, nil
}
