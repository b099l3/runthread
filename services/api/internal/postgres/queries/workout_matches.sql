-- name: CreateWorkoutMatch :one
INSERT INTO workout_matches (
    id,
    planned_workout_id,
    imported_activity_id,
    status,
    confidence,
    matched_by,
    matched_at,
    notes
) VALUES (
    sqlc.arg(id),
    sqlc.arg(planned_workout_id),
    sqlc.arg(imported_activity_id),
    sqlc.arg(status),
    sqlc.arg(confidence),
    sqlc.arg(matched_by),
    sqlc.arg(matched_at),
    sqlc.arg(notes)
)
RETURNING *;

-- name: GetWorkoutMatch :one
SELECT *
FROM workout_matches
WHERE id = sqlc.arg(id);

-- name: UpdateWorkoutMatch :one
UPDATE workout_matches
SET
    planned_workout_id = sqlc.arg(planned_workout_id),
    imported_activity_id = sqlc.arg(imported_activity_id),
    status = sqlc.arg(status),
    confidence = sqlc.arg(confidence),
    matched_by = sqlc.arg(matched_by),
    matched_at = sqlc.arg(matched_at),
    notes = sqlc.arg(notes)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListWorkoutMatchesByPlannedWorkout :many
SELECT *
FROM workout_matches
WHERE planned_workout_id = sqlc.arg(planned_workout_id)
ORDER BY matched_at DESC, id;

-- name: ListWorkoutMatchesByImportedActivity :many
SELECT *
FROM workout_matches
WHERE imported_activity_id = sqlc.arg(imported_activity_id)
ORDER BY matched_at DESC, id;

-- name: DeleteWorkoutMatch :exec
DELETE FROM workout_matches
WHERE id = sqlc.arg(id);
