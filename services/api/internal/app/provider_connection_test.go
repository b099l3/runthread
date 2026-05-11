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
