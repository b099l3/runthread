CREATE TABLE provider_webhook_events (
    provider text NOT NULL,
    event_id text NOT NULL,
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (provider, event_id),
    CONSTRAINT provider_webhook_events_provider_non_empty CHECK (provider <> ''),
    CONSTRAINT provider_webhook_events_event_id_non_empty CHECK (event_id <> '')
);
