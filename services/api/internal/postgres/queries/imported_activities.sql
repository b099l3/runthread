-- name: CreateImportedActivity :one
INSERT INTO imported_activities (
    id,
    athlete_id,
    type,
    started_at,
    duration_seconds,
    distance_meters,
    average_pace_seconds_per_kilometer,
    average_heart_bpm
) VALUES (
    sqlc.arg(id),
    sqlc.arg(athlete_id),
    sqlc.arg(type),
    sqlc.arg(started_at),
    sqlc.arg(duration_seconds),
    sqlc.arg(distance_meters),
    sqlc.arg(average_pace_seconds_per_kilometer),
    sqlc.arg(average_heart_bpm)
)
RETURNING *;

-- name: GetImportedActivity :one
SELECT *
FROM imported_activities
WHERE id = sqlc.arg(id);

-- name: UpdateImportedActivity :one
UPDATE imported_activities
SET
    athlete_id = sqlc.arg(athlete_id),
    type = sqlc.arg(type),
    started_at = sqlc.arg(started_at),
    duration_seconds = sqlc.arg(duration_seconds),
    distance_meters = sqlc.arg(distance_meters),
    average_pace_seconds_per_kilometer = sqlc.arg(average_pace_seconds_per_kilometer),
    average_heart_bpm = sqlc.arg(average_heart_bpm)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListImportedActivitiesByAthlete :many
SELECT *
FROM imported_activities
WHERE athlete_id = sqlc.arg(athlete_id)
ORDER BY started_at DESC, id;

-- name: DeleteImportedActivity :exec
DELETE FROM imported_activities
WHERE id = sqlc.arg(id);
