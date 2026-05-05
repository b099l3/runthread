-- name: CreateAdaptationEvent :one
INSERT INTO adaptation_events (
    id,
    plan_id,
    athlete_id,
    type,
    reason,
    summary,
    created_at
) VALUES (
    sqlc.arg(id),
    sqlc.arg(plan_id),
    sqlc.arg(athlete_id),
    sqlc.arg(type),
    sqlc.arg(reason),
    sqlc.arg(summary),
    sqlc.arg(created_at)
)
RETURNING *;

-- name: GetAdaptationEvent :one
SELECT *
FROM adaptation_events
WHERE id = sqlc.arg(id);

-- name: UpdateAdaptationEvent :one
UPDATE adaptation_events
SET
    plan_id = sqlc.arg(plan_id),
    athlete_id = sqlc.arg(athlete_id),
    type = sqlc.arg(type),
    reason = sqlc.arg(reason),
    summary = sqlc.arg(summary),
    created_at = sqlc.arg(created_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListAdaptationEventsByAthlete :many
SELECT *
FROM adaptation_events
WHERE athlete_id = sqlc.arg(athlete_id)
ORDER BY created_at DESC, id;

-- name: ListAdaptationEventsByPlan :many
SELECT *
FROM adaptation_events
WHERE plan_id = sqlc.arg(plan_id)
ORDER BY created_at DESC, id;

-- name: DeleteAdaptationEvent :exec
DELETE FROM adaptation_events
WHERE id = sqlc.arg(id);
