package postgres

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

func TestNewProviderWebhookDeduperRequiresDB(t *testing.T) {
	_, err := NewProviderWebhookDeduper(nil, "strava")
	if err == nil {
		t.Fatal("NewProviderWebhookDeduper returned nil error, want db required error")
	}
}

func TestNewProviderWebhookDeduperRequiresProvider(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://example")
	if err != nil {
		t.Fatalf("sql.Open returned error: %v", err)
	}
	defer db.Close()

	_, err = NewProviderWebhookDeduper(db, "")
	if err == nil {
		t.Fatal("NewProviderWebhookDeduper returned nil error, want provider required error")
	}
}

func TestNewProviderWebhookDeduper(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://example")
	if err != nil {
		t.Fatalf("sql.Open returned error: %v", err)
	}
	defer db.Close()

	deduper, err := NewProviderWebhookDeduper(db, "strava")
	if err != nil {
		t.Fatalf("NewProviderWebhookDeduper returned error: %v", err)
	}
	if deduper.provider != "strava" {
		t.Fatalf("provider = %q, want strava", deduper.provider)
	}
}
