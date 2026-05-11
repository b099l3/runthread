package postgres

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

func TestNewProviderStoreRequiresDB(t *testing.T) {
	_, err := NewProviderStore(nil)
	if err == nil {
		t.Fatal("NewProviderStore returned nil error, want db required error")
	}
}

func TestNewProviderStoreComposesRepositories(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://example")
	if err != nil {
		t.Fatalf("open db handle: %v", err)
	}
	defer db.Close()

	store, err := NewProviderStore(db)
	if err != nil {
		t.Fatalf("NewProviderStore returned error: %v", err)
	}

	if store.ProviderConnections == nil {
		t.Fatal("ProviderConnections repository is nil")
	}
	if store.ProviderActivities == nil {
		t.Fatal("ProviderActivities repository is nil")
	}
	if store.ProviderActivityPayloads == nil {
		t.Fatal("ProviderActivityPayloads repository is nil")
	}
	if store.ProviderImportEvents == nil {
		t.Fatal("ProviderImportEvents repository is nil")
	}
}
