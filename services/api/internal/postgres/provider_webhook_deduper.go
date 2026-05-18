package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type ProviderWebhookDeduper struct {
	db       *sql.DB
	provider string
}

func NewProviderWebhookDeduper(db *sql.DB, provider string) (*ProviderWebhookDeduper, error) {
	if db == nil {
		return nil, fmt.Errorf("provider webhook deduper db is required")
	}
	if provider == "" {
		return nil, fmt.Errorf("provider webhook deduper provider is required")
	}
	return &ProviderWebhookDeduper{db: db, provider: provider}, nil
}

func (d *ProviderWebhookDeduper) SeenWebhookEvent(ctx context.Context, eventID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if eventID == "" {
		return false, errors.New("provider webhook event id is required")
	}
	var seen bool
	if err := d.db.QueryRowContext(ctx, `
SELECT EXISTS (
    SELECT 1
    FROM provider_webhook_events
    WHERE provider = $1
      AND event_id = $2
)`, d.provider, eventID).Scan(&seen); err != nil {
		return false, fmt.Errorf("check provider webhook event: %w", err)
	}
	return seen, nil
}

func (d *ProviderWebhookDeduper) MarkWebhookEventSeen(ctx context.Context, eventID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if eventID == "" {
		return errors.New("provider webhook event id is required")
	}
	if _, err := d.db.ExecContext(ctx, `
INSERT INTO provider_webhook_events (provider, event_id)
VALUES ($1, $2)
ON CONFLICT (provider, event_id) DO NOTHING`, d.provider, eventID); err != nil {
		return fmt.Errorf("mark provider webhook event seen: %w", err)
	}
	return nil
}
