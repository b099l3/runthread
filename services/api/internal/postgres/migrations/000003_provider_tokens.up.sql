CREATE TABLE provider_tokens (
    reference text PRIMARY KEY,
    provider text NOT NULL,
    provider_connection_id text NOT NULL,
    provider_user_id text,
    encrypted_token bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX provider_tokens_connection_idx ON provider_tokens (provider_connection_id);
