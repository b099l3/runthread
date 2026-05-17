package strava

import (
	"context"
	"net/url"
	"testing"

	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestConnectionStarterReturnsOAuthReadyAuthorizationURL(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	oauth := testOAuthService(store, &fakeTokenStore{})
	starter := ConnectionStarter{
		OAuth:       &oauth,
		RedirectURI: "runthread://provider/strava/callback",
		Scopes:      []string{"activity:read_all"},
	}

	response, err := starter.StartProviderConnection(ctx, app.StartProviderConnectionRequest{
		AthleteID: "athlete-1",
		Provider:  app.ProviderStrava,
	})
	if err != nil {
		t.Fatalf("StartProviderConnection returned error: %v", err)
	}

	if !response.OAuthReady {
		t.Fatal("OAuthReady = false, want true")
	}
	if response.State != "state-1" {
		t.Fatalf("state = %q, want state-1", response.State)
	}
	parsed, err := url.Parse(response.AuthorizationURL)
	if err != nil {
		t.Fatalf("parse authorization url: %v", err)
	}
	if parsed.Query().Get("redirect_uri") != "runthread://provider/strava/callback" {
		t.Fatalf("redirect_uri = %q", parsed.Query().Get("redirect_uri"))
	}
}
