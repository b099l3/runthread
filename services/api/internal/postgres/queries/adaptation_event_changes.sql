-- name: CreateAdaptationEventChange :one
INSERT INTO adaptation_event_changes (
    id,
    adaptation_event_id,
    planned_workout_id,
    type,
    description,
    position
) VALUES (
    sqlc.arg(id),
    sqlc.arg(adaptation_event_id),
    sqlc.arg(planned_workout_id),
    sqlc.arg(type),
    sqlc.arg(description),
    sqlc.arg(position)
)
RETURNING *;

-- name: ListAdaptationEventChanges :many
SELECT *
FROM adaptation_event_changes
WHERE adaptation_event_id = sqlc.arg(adaptation_event_id)
ORDER BY position, id;

-- name: DeleteAdaptationEventChanges :exec
DELETE FROM adaptation_event_changes
WHERE adaptation_event_id = sqlc.arg(adaptation_event_id);
