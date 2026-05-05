-- name: CreateAthleteProfile :one
INSERT INTO athlete_profiles (
    id,
    display_name,
    experience_level,
    current_weekly_distance_meters,
    preferred_run_days,
    constraints
) VALUES (
    sqlc.arg(id),
    sqlc.arg(display_name),
    sqlc.arg(experience_level),
    sqlc.arg(current_weekly_distance_meters),
    sqlc.arg(preferred_run_days),
    sqlc.arg(constraints)
)
RETURNING *;

-- name: GetAthleteProfile :one
SELECT *
FROM athlete_profiles
WHERE id = sqlc.arg(id);

-- name: UpdateAthleteProfile :one
UPDATE athlete_profiles
SET
    display_name = sqlc.arg(display_name),
    experience_level = sqlc.arg(experience_level),
    current_weekly_distance_meters = sqlc.arg(current_weekly_distance_meters),
    preferred_run_days = sqlc.arg(preferred_run_days),
    constraints = sqlc.arg(constraints),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteAthleteProfile :exec
DELETE FROM athlete_profiles
WHERE id = sqlc.arg(id);
