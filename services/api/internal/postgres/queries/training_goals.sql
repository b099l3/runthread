-- name: CreateTrainingGoal :one
INSERT INTO training_goals (
    id,
    athlete_id,
    type,
    target_date,
    target_distance_meters,
    target_duration_seconds,
    notes
) VALUES (
    sqlc.arg(id),
    sqlc.arg(athlete_id),
    sqlc.arg(type),
    sqlc.arg(target_date),
    sqlc.arg(target_distance_meters),
    sqlc.arg(target_duration_seconds),
    sqlc.arg(notes)
)
RETURNING *;

-- name: GetTrainingGoal :one
SELECT *
FROM training_goals
WHERE id = sqlc.arg(id);

-- name: ListTrainingGoalsByAthlete :many
SELECT *
FROM training_goals
WHERE athlete_id = sqlc.arg(athlete_id)
ORDER BY created_at DESC, id;

-- name: UpdateTrainingGoal :one
UPDATE training_goals
SET
    type = sqlc.arg(type),
    target_date = sqlc.arg(target_date),
    target_distance_meters = sqlc.arg(target_distance_meters),
    target_duration_seconds = sqlc.arg(target_duration_seconds),
    notes = sqlc.arg(notes),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteTrainingGoal :exec
DELETE FROM training_goals
WHERE id = sqlc.arg(id);
