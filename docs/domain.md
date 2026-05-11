# Domain Concepts

## AthleteProfile

The runner's basic training context, such as experience level, current weekly volume, schedule preferences, and constraints.

Initial Go model:

- `ID`
- `DisplayName`
- `ExperienceLevel`: `beginner`, `intermediate`, `advanced`
- `CurrentWeeklyDistance`
- `PreferredRunDays`
- `Constraints`

Validation checks that the ID is present, experience level is recognised when set, and weekly distance is not negative.

## TrainingGoal

The outcome the runner is training toward, such as a target race, distance, completion goal, time goal, or general fitness objective.

Initial Go model:

- `ID`
- `AthleteID`
- `Type`: `general_fitness`, `race`, `distance`, `time`
- `TargetDate`
- `TargetDistance`
- `TargetDuration`
- `Notes`

Validation checks required IDs, recognised goal type, and non-negative target distance/duration.

## TrainingPlan

The structured plan for an athlete and goal. It contains weeks, workouts, and enough metadata to understand the plan's current state.

Initial Go model:

- `ID`
- `AthleteID`
- `GoalID`
- `Status`: `draft`, `active`, `completed`, `archived`
- `StartsOn`
- `EndsOn`
- `Weeks`

Validation checks required IDs, recognised status, required dates, date ordering, and nested week validity.

## PlanWeek

A calendar week within a training plan. It groups planned workouts and can hold week-level intent such as recovery, build, peak, or taper.

Initial Go model:

- `ID`
- `AthleteID`
- `GoalID`
- `PlanID`
- `WeekIndex`
- `StartsOn`
- `Focus`: `base`, `build`, `recovery`, `peak`, `taper`
- `Workouts`

Validation checks required IDs, positive week index, required start date, recognised focus when set, and nested workout validity.

`AthleteID` is required because generated weeks, repository storage, and future API reads need a stable owner. `GoalID` is optional because some future weeks may be generated without a specific race or time goal.

## PlannedWorkout

A workout the runner is expected to complete. It includes date, workout type, intended duration or distance, intensity targets, and simple notes.

Initial Go model:

- `ID`
- `PlanID`
- `PlanWeekID`
- `ScheduledFor`
- `Type`: `easy`, `long_run`, `workout`, `recovery`, `race`, `rest`, `strength`
- `Status`: `scheduled`, `completed`, `missed`, `skipped`, `moved`
- `TargetDistance`
- `TargetDuration`
- `Intensity`
- `Notes`

Validation checks required IDs/date, recognised type/status, non-negative targets, and at least one target distance or duration for run workouts. Rest and strength days may be planned without a distance target.

Stage 2 planning uses `workout` with tempo intensity as the initial threshold workout representation.

Stage 3 adds deterministic workout flow helpers for marking planned workouts completed, missed, skipped, or moved. Scheduled and moved workouts can transition; completed, missed, and skipped workouts are terminal for now.

## ImportedActivity

A completed activity imported from a provider such as Strava, Garmin, COROS, or Apple Health and normalised into Runthread's own model. It should contain provider-neutral data such as start time, duration, distance, pace, heart rate summaries, and activity type.

Initial Go model:

- `ID`
- `AthleteID`
- `Type`: `run`, `trail_run`, `treadmill`, `walk`, `other`
- `StartedAt`
- `Duration`
- `Distance`
- `AveragePace`
- `AverageHeartBPM`

Validation checks required IDs, recognised activity type, required start time, positive duration, and non-negative distance, pace, and heart rate.

Provider names, provider activity IDs, raw payloads, and OAuth details are intentionally not part of this core domain type. Those belong in integration and persistence boundaries.

Core planning, matching, workout result, and adaptation logic should use `ImportedActivity` only. Provider-specific payloads should stay in provider packages. Strava, Garmin, COROS, and Apple Health should all map into this same internal activity model before core domain logic sees the activity.

## ActivityProvider

An adapter boundary for a provider that can supply activity data. An `ActivityProvider` owns provider-specific parsing, validation, field mapping, and policy details, then emits provider-neutral imported activity data.

Initial Go interface shape:

- `ProviderName() string`
- `NormaliseActivity(ctx context.Context, payload []byte) (domain.ImportedActivity, error)`

Provider implementations must not make planning or adaptation decisions. They should only connect provider data to the provider-neutral import pipeline.

## ProviderConnection

The athlete-owned connection between Runthread and an external provider account.

It belongs outside the core domain package because it includes integration state, provider account identifiers, token references, sync status, cursors, errors, and timestamps. The backend owns provider connections, token exchange, token refresh, disconnect handling, and provider API access.

## ActivityImportJob

A backend unit of work that imports activities for a provider connection. It may represent historical backfill, webhook-triggered import, polling, manual retry, or a local mock import.

It should carry provider connection identity, job type, cursor or time-window metadata, status, retry/error state, and timestamps. It should call a provider adapter and store provider import state without exposing provider payloads to planning or adaptation logic.

## ProviderWebhookEvent

A provider-facing event received by the backend, such as a Strava activity create/update/delete event or a future Garmin delivery event.

Webhook events should be authenticated according to provider rules, stored or logged for idempotency and audit as allowed by provider terms, and converted into import jobs or provider activity updates. Core domain logic should not consume webhook payloads directly.

## WorkoutMatch

The link between an imported activity and a planned workout. It records whether the match was automatic, manually selected, uncertain, or rejected.

Initial Go model:

- `ID`
- `PlannedWorkoutID`
- `ImportedActivityID`
- `Status`: `matched`, `uncertain`, `rejected`
- `Confidence`: `high`, `medium`, `low`
- `MatchedBy`: `automatic`, `manual`
- `MatchedAt`
- `Notes`

Validation checks required IDs, recognised status/confidence/source, and required match time.

Stage 5 adds deterministic matching logic in `services/api/internal/matching`.

Initial automatic matching uses:

- activity/workout type compatibility
- calendar date gap
- target distance closeness
- target duration closeness

The matcher returns:

- `matched` with high confidence when type, date, distance, and duration line up.
- `uncertain` with medium confidence when the date is close and some effort signals line up.
- `rejected` with low confidence when the type is incompatible, the date is too far away, or too few signals match.

Manual override is represented by a manual high-confidence `WorkoutMatch` after validating the planned workout and imported activity.

## WorkoutResult

The evaluation of what happened against the planned workout. Examples include completed as planned, partially completed, missed, overperformed, underperformed, or moved.

Initial Go model:

- `ID`
- `PlannedWorkoutID`
- `ImportedActivityID`
- `Outcome`: `completed_as_planned`, `partially_completed`, `missed`, `skipped`, `moved`, `overperformed`, `underperformed`
- `CompletedAt`
- `Distance`
- `Duration`
- `Notes`

Validation checks required IDs, recognised outcome, non-negative distance/duration, and requires imported activity/completion time for outcomes that depend on an activity.

Stage 3 workout flow creates `WorkoutResult` records when a planned workout is marked completed, missed, skipped, or moved:

- Completed workouts require an imported activity ID and completion time.
- Partial, overperformed, and underperformed outcomes are treated as completed activity-backed outcomes.
- Missed and skipped workouts do not require an imported activity.
- Moved workouts record a moved result and update the planned workout date.

## AdaptationEvent

A recorded plan change caused by new information, such as a missed workout, accumulated fatigue, or a changed schedule. It should include what changed, why it changed, and enough context to explain it later.

Initial Go model:

- `ID`
- `PlanID`
- `AthleteID`
- `Type`: `missed_workout`, `partial_completion`, `overperformance`, `underperformance`, `schedule_change`
- `Reason`
- `Summary`
- `CreatedAt`
- `Changes`

Validation checks required IDs, recognised adaptation type, required reason/summary/creation time, and nested plan change validity.

`PlanChange` records a small explainable change with:

- `PlannedWorkoutID`
- `Type`: `workout_moved`, `workout_removed`, `workout_added`, `workout_adjusted`, `plan_note_added`
- `Description`

The adaptation event records what changed and why; the deterministic adaptation engine will decide when to create these events in a later stage.

Stage 6 adds a deterministic adaptation engine in `services/api/internal/adaptation`.

The initial engine responds to:

- missed workouts
- partial completions
- overperformed workouts
- underperformed workouts

It produces provider-neutral `AdaptationEvent` values with conservative `PlanChange` entries. Completed-as-planned results do not produce an adaptation event. The engine records intended changes and explanations, but it does not persist data or mutate a stored plan yet.
