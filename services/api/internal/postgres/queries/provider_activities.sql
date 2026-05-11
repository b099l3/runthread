-- name: CreateProviderActivity :one
INSERT INTO provider_activities (
    id,
    provider_connection_id,
    athlete_id,
    imported_activity_id,
    provider,
    provider_activity_id,
    provider_activity_type,
    started_at,
    status,
    first_seen_at,
    last_synced_at,
    last_error
) VALUES (
    sqlc.arg(id),
    sqlc.arg(provider_connection_id),
    sqlc.arg(athlete_id),
    sqlc.arg(imported_activity_id),
    sqlc.arg(provider),
    sqlc.arg(provider_activity_id),
    sqlc.arg(provider_activity_type),
    sqlc.arg(started_at),
    sqlc.arg(status),
    sqlc.arg(first_seen_at),
    sqlc.arg(last_synced_at),
    sqlc.arg(last_error)
)
RETURNING *;

-- name: GetProviderActivity :one
SELECT *
FROM provider_activities
WHERE id = sqlc.arg(id);

-- name: GetProviderActivityByProviderID :one
SELECT *
FROM provider_activities
WHERE provider_connection_id = sqlc.arg(provider_connection_id)
  AND provider_activity_id = sqlc.arg(provider_activity_id);

-- name: UpdateProviderActivity :one
UPDATE provider_activities
SET
    imported_activity_id = sqlc.arg(imported_activity_id),
    provider_activity_type = sqlc.arg(provider_activity_type),
    started_at = sqlc.arg(started_at),
    status = sqlc.arg(status),
    last_synced_at = sqlc.arg(last_synced_at),
    last_error = sqlc.arg(last_error),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListProviderActivitiesByAthlete :many
SELECT *
FROM provider_activities
WHERE athlete_id = sqlc.arg(athlete_id)
ORDER BY started_at DESC NULLS LAST, first_seen_at DESC, id;

-- name: ListProviderActivitiesByStatus :many
SELECT *
FROM provider_activities
WHERE status = sqlc.arg(status)
ORDER BY first_seen_at ASC, id;
