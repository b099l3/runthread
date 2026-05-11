-- name: CreateProviderImportEvent :one
INSERT INTO provider_import_events (
    id,
    provider_connection_id,
    provider_activity_id,
    provider,
    event_type,
    delivery_id,
    status,
    received_at,
    processed_at,
    error
) VALUES (
    sqlc.arg(id),
    sqlc.arg(provider_connection_id),
    sqlc.arg(provider_activity_id),
    sqlc.arg(provider),
    sqlc.arg(event_type),
    sqlc.arg(delivery_id),
    sqlc.arg(status),
    sqlc.arg(received_at),
    sqlc.arg(processed_at),
    sqlc.arg(error)
)
RETURNING *;

-- name: GetProviderImportEvent :one
SELECT *
FROM provider_import_events
WHERE id = sqlc.arg(id);

-- name: UpdateProviderImportEventStatus :one
UPDATE provider_import_events
SET
    status = sqlc.arg(status),
    processed_at = sqlc.arg(processed_at),
    error = sqlc.arg(error)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListProviderImportEventsByConnection :many
SELECT *
FROM provider_import_events
WHERE provider_connection_id = sqlc.arg(provider_connection_id)::uuid
ORDER BY received_at DESC, id;

-- name: ListProviderImportEventsByStatus :many
SELECT *
FROM provider_import_events
WHERE status = sqlc.arg(status)
ORDER BY received_at ASC, id;
