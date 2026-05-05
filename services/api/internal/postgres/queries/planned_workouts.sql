-- name: CreatePlannedWorkout :one
INSERT INTO planned_workouts (
    id,
    plan_week_id,
    plan_id,
    scheduled_for,
    type,
    status,
    target_distance_meters,
    target_duration_seconds,
    intensity_kind,
    intensity_description,
    notes
) VALUES (
    sqlc.arg(id),
    sqlc.arg(plan_week_id),
    sqlc.arg(plan_id),
    sqlc.arg(scheduled_for),
    sqlc.arg(type),
    sqlc.arg(status),
    sqlc.arg(target_distance_meters),
    sqlc.arg(target_duration_seconds),
    sqlc.arg(intensity_kind),
    sqlc.arg(intensity_description),
    sqlc.arg(notes)
)
RETURNING *;

-- name: GetPlannedWorkout :one
SELECT *
FROM planned_workouts
WHERE id = sqlc.arg(id);

-- name: ListPlannedWorkoutsByWeek :many
SELECT *
FROM planned_workouts
WHERE plan_week_id = sqlc.arg(plan_week_id)
ORDER BY scheduled_for, id;

-- name: ListPlannedWorkoutsByPlan :many
SELECT *
FROM planned_workouts
WHERE plan_id = sqlc.arg(plan_id)
ORDER BY scheduled_for, id;

-- name: UpdatePlannedWorkout :one
UPDATE planned_workouts
SET
    scheduled_for = sqlc.arg(scheduled_for),
    type = sqlc.arg(type),
    status = sqlc.arg(status),
    target_distance_meters = sqlc.arg(target_distance_meters),
    target_duration_seconds = sqlc.arg(target_duration_seconds),
    intensity_kind = sqlc.arg(intensity_kind),
    intensity_description = sqlc.arg(intensity_description),
    notes = sqlc.arg(notes),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdatePlannedWorkoutStatus :one
UPDATE planned_workouts
SET
    status = sqlc.arg(status),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePlannedWorkout :exec
DELETE FROM planned_workouts
WHERE id = sqlc.arg(id);
