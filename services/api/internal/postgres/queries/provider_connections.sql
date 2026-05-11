-- name: CreateProviderConnection :one
INSERT INTO provider_connections (
    id,
    athlete_id,
    provider,
    provider_user_id,
    status,
    connected_at,
    disconnected_at,
    last_sync_at,
    last_import_cursor,
    token_reference,
    token_expires_at,
    last_error
) VALUES (
    sqlc.arg(id),
    sqlc.arg(athlete_id),
    sqlc.arg(provider),
    sqlc.arg(provider_user_id),
    sqlc.arg(status),
    sqlc.arg(connected_at),
    sqlc.arg(disconnected_at),
    sqlc.arg(last_sync_at),
    sqlc.arg(last_import_cursor),
    sqlc.arg(token_reference),
    sqlc.arg(token_expires_at),
    sqlc.arg(last_error)
)
RETURNING *;

-- name: GetProviderConnection :one
SELECT *
FROM provider_connections
WHERE id = sqlc.arg(id);

-- name: UpdateProviderConnection :one
UPDATE provider_connections
SET
    provider_user_id = sqlc.arg(provider_user_id),
    status = sqlc.arg(status),
    connected_at = sqlc.arg(connected_at),
    disconnected_at = sqlc.arg(disconnected_at),
    last_sync_at = sqlc.arg(last_sync_at),
    last_import_cursor = sqlc.arg(last_import_cursor),
    token_reference = sqlc.arg(token_reference),
    token_expires_at = sqlc.arg(token_expires_at),
    last_error = sqlc.arg(last_error),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListProviderConnectionsByAthlete :many
SELECT *
FROM provider_connections
WHERE athlete_id = sqlc.arg(athlete_id)
ORDER BY created_at DESC, id;

-- name: ListProviderConnectionsByStatus :many
SELECT *
FROM provider_connections
WHERE status = sqlc.arg(status)
ORDER BY updated_at ASC, id;
