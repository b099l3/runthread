-- name: CreateWorkoutResult :one
INSERT INTO workout_results (
    id,
    planned_workout_id,
    imported_activity_id,
    outcome,
    completed_at,
    distance_meters,
    duration_seconds,
    notes
) VALUES (
    sqlc.arg(id),
    sqlc.arg(planned_workout_id),
    sqlc.arg(imported_activity_id),
    sqlc.arg(outcome),
    sqlc.arg(completed_at),
    sqlc.arg(distance_meters),
    sqlc.arg(duration_seconds),
    sqlc.arg(notes)
)
RETURNING *;

-- name: GetWorkoutResult :one
SELECT *
FROM workout_results
WHERE id = sqlc.arg(id);

-- name: UpdateWorkoutResult :one
UPDATE workout_results
SET
    planned_workout_id = sqlc.arg(planned_workout_id),
    imported_activity_id = sqlc.arg(imported_activity_id),
    outcome = sqlc.arg(outcome),
    completed_at = sqlc.arg(completed_at),
    distance_meters = sqlc.arg(distance_meters),
    duration_seconds = sqlc.arg(duration_seconds),
    notes = sqlc.arg(notes)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListWorkoutResultsByPlannedWorkout :many
SELECT *
FROM workout_results
WHERE planned_workout_id = sqlc.arg(planned_workout_id)
ORDER BY created_at DESC, id;

-- name: ListWorkoutResultsByImportedActivity :many
SELECT *
FROM workout_results
WHERE imported_activity_id = sqlc.arg(imported_activity_id)
ORDER BY created_at DESC, id;

-- name: DeleteWorkoutResult :exec
DELETE FROM workout_results
WHERE id = sqlc.arg(id);
