-- name: CreatePlanWeek :one
INSERT INTO plan_weeks (
    id,
    athlete_id,
    goal_id,
    plan_id,
    week_index,
    starts_on,
    focus
) VALUES (
    sqlc.arg(id),
    sqlc.arg(athlete_id),
    sqlc.arg(goal_id),
    sqlc.arg(plan_id),
    sqlc.arg(week_index),
    sqlc.arg(starts_on),
    sqlc.arg(focus)
)
RETURNING *;

-- name: GetPlanWeek :one
SELECT *
FROM plan_weeks
WHERE id = sqlc.arg(id);

-- name: ListPlanWeeksByPlan :many
SELECT *
FROM plan_weeks
WHERE plan_id = sqlc.arg(plan_id)
ORDER BY week_index;

-- name: ListPlanWeeksByAthlete :many
SELECT *
FROM plan_weeks
WHERE athlete_id = sqlc.arg(athlete_id)
ORDER BY starts_on DESC, week_index DESC;

-- name: UpdatePlanWeek :one
UPDATE plan_weeks
SET
    goal_id = sqlc.arg(goal_id),
    week_index = sqlc.arg(week_index),
    starts_on = sqlc.arg(starts_on),
    focus = sqlc.arg(focus),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePlanWeek :exec
DELETE FROM plan_weeks
WHERE id = sqlc.arg(id);
