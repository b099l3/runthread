package startup

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/runthread/runthread/services/api/internal/config"
	"github.com/runthread/runthread/services/api/internal/postgres"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type StorageKind string

const (
	StorageKindInMemory StorageKind = "in_memory"
	StorageKindPostgres StorageKind = "postgres"
)

type Storage struct {
	Store   repository.Store
	Kind    StorageKind
	Cleanup func() error
}

func ComposeStorage(cfg config.Config) (Storage, error) {
	if !cfg.DatabaseConfigured() {
		store := repository.NewInMemoryStore()
		if err := SeedDemoData(context.Background(), store); err != nil {
			return Storage{}, fmt.Errorf("seed demo data: %w", err)
		}
		return Storage{
			Store:   store,
			Kind:    StorageKindInMemory,
			Cleanup: func() error { return nil },
		}, nil
	}

	db, err := OpenPostgresDB(cfg.DatabaseURL)
	if err != nil {
		return Storage{}, err
	}

	store, err := postgres.NewStore(db)
	if err != nil {
		_ = db.Close()
		return Storage{}, fmt.Errorf("compose postgres store: %w", err)
	}

	return Storage{
		Store:   store,
		Kind:    StorageKindPostgres,
		Cleanup: db.Close,
	}, nil
}

func OpenPostgresDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres db: %w", err)
	}
	return db, nil
}
