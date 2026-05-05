package startup

import (
	"testing"

	"github.com/runthread/runthread/services/api/internal/config"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestComposeStorageDefaultsToInMemoryStore(t *testing.T) {
	storage, err := ComposeStorage(config.Config{})
	if err != nil {
		t.Fatalf("ComposeStorage returned error: %v", err)
	}

	if storage.Kind != StorageKindInMemory {
		t.Fatalf("Kind = %q, want %q", storage.Kind, StorageKindInMemory)
	}
	if storage.Store == nil {
		t.Fatal("Store is nil, want in-memory store")
	}
	if _, ok := storage.Store.(*repository.InMemoryStore); !ok {
		t.Fatalf("Store type = %T, want *repository.InMemoryStore", storage.Store)
	}
	if storage.Cleanup == nil {
		t.Fatal("Cleanup is nil")
	}
	if err := storage.Cleanup(); err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}
}
