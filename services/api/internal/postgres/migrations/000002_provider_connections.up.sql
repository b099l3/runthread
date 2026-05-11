CREATE TABLE provider_connections (
    id uuid PRIMARY KEY,
    athlete_id uuid NOT NULL REFERENCES athlete_profiles(id),
    provider text NOT NULL,
    provider_user_id text,
    status text NOT NULL,
    connected_at timestamptz,
    disconnected_at timestamptz,
    last_sync_at timestamptz,
    last_import_cursor text,
    token_reference text,
    token_expires_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT provider_connections_provider_non_empty CHECK (provider <> ''),
    CONSTRAINT provider_connections_status_valid CHECK (status IN ('pending', 'connected', 'syncing', 'error', 'disconnected'))
);

CREATE TABLE provider_activities (
    id uuid PRIMARY KEY,
    provider_connection_id uuid NOT NULL REFERENCES provider_connections(id),
    athlete_id uuid NOT NULL REFERENCES athlete_profiles(id),
    imported_activity_id uuid REFERENCES imported_activities(id),
    provider text NOT NULL,
    provider_activity_id text NOT NULL,
    provider_activity_type text,
    started_at timestamptz,
    status text NOT NULL,
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_synced_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT provider_activities_provider_non_empty CHECK (provider <> ''),
    CONSTRAINT provider_activities_provider_activity_id_non_empty CHECK (provider_activity_id <> ''),
    CONSTRAINT provider_activities_status_valid CHECK (status IN ('received', 'normalised', 'ignored', 'failed')),
    CONSTRAINT provider_activities_connection_activity_unique UNIQUE (provider_connection_id, provider_activity_id)
);

CREATE TABLE provider_activity_payloads (
    id uuid PRIMARY KEY,
    provider_activity_id uuid NOT NULL REFERENCES provider_activities(id) ON DELETE CASCADE,
    payload jsonb NOT NULL,
    payload_kind text NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT provider_activity_payloads_payload_kind_non_empty CHECK (payload_kind <> '')
);

CREATE TABLE provider_import_events (
    id uuid PRIMARY KEY,
    provider_connection_id uuid REFERENCES provider_connections(id),
    provider_activity_id uuid REFERENCES provider_activities(id),
    provider text NOT NULL,
    event_type text NOT NULL,
    delivery_id text,
    status text NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    processed_at timestamptz,
    error text,
    CONSTRAINT provider_import_events_provider_non_empty CHECK (provider <> ''),
    CONSTRAINT provider_import_events_event_type_non_empty CHECK (event_type <> ''),
    CONSTRAINT provider_import_events_status_valid CHECK (status IN ('received', 'processed', 'ignored', 'failed'))
);

CREATE INDEX provider_connections_athlete_provider_idx ON provider_connections (athlete_id, provider);
CREATE INDEX provider_connections_provider_user_id_idx ON provider_connections (provider, provider_user_id) WHERE provider_user_id IS NOT NULL;
CREATE INDEX provider_connections_status_idx ON provider_connections (status);

CREATE INDEX provider_activities_athlete_started_at_idx ON provider_activities (athlete_id, started_at);
CREATE INDEX provider_activities_imported_activity_id_idx ON provider_activities (imported_activity_id);
CREATE INDEX provider_activities_status_idx ON provider_activities (status);
CREATE INDEX provider_activities_provider_activity_id_idx ON provider_activities (provider, provider_activity_id);

CREATE INDEX provider_activity_payloads_activity_id_received_at_idx ON provider_activity_payloads (provider_activity_id, received_at);

CREATE INDEX provider_import_events_connection_received_at_idx ON provider_import_events (provider_connection_id, received_at);
CREATE INDEX provider_import_events_activity_id_idx ON provider_import_events (provider_activity_id);
CREATE INDEX provider_import_events_status_idx ON provider_import_events (status);
CREATE UNIQUE INDEX provider_import_events_provider_delivery_unique_idx ON provider_import_events (provider, delivery_id) WHERE delivery_id IS NOT NULL;
