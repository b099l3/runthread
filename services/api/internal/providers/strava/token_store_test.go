package strava

import (
	"context"
	"strings"
	"testing"
)

func TestInMemoryTokenStoreStoresTokenBehindReference(t *testing.T) {
	store := NewInMemoryTokenStore()

	reference, err := store.StoreToken(context.Background(), validOAuthToken())
	if err != nil {
		t.Fatalf("StoreToken returned error: %v", err)
	}

	if !strings.HasPrefix(reference, "strava-token-") {
		t.Fatalf("reference = %q, want strava-token prefix", reference)
	}
	if store.tokens[reference].AccessToken != "access-token" {
		t.Fatalf("stored access token = %q, want access-token", store.tokens[reference].AccessToken)
	}
}

func TestInMemoryTokenStoreLoadsAndUpdatesToken(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryTokenStore()
	reference, err := store.StoreToken(ctx, validOAuthToken())
	if err != nil {
		t.Fatalf("StoreToken returned error: %v", err)
	}

	token, err := store.LoadToken(ctx, reference)
	if err != nil {
		t.Fatalf("LoadToken returned error: %v", err)
	}
	if token.AccessToken != "access-token" {
		t.Fatalf("access token = %q, want access-token", token.AccessToken)
	}

	token.AccessToken = "new-access-token"
	if err := store.UpdateToken(ctx, reference, token); err != nil {
		t.Fatalf("UpdateToken returned error: %v", err)
	}
	updated, err := store.LoadToken(ctx, reference)
	if err != nil {
		t.Fatalf("LoadToken returned error: %v", err)
	}
	if updated.AccessToken != "new-access-token" {
		t.Fatalf("updated access token = %q, want new-access-token", updated.AccessToken)
	}
}
