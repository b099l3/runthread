# Domain Concepts

## AthleteProfile

The runner's basic training context, such as experience level, current weekly volume, schedule preferences, constraints, and provider connections.

## TrainingGoal

The outcome the runner is training toward, such as a target race, distance, completion goal, time goal, or general fitness objective.

## TrainingPlan

The structured plan for an athlete and goal. It contains weeks, workouts, and enough metadata to understand the plan's current state.

## PlanWeek

A calendar week within a training plan. It groups planned workouts and can hold week-level intent such as recovery, build, peak, or taper.

## PlannedWorkout

A workout the runner is expected to complete. It includes date, workout type, intended duration or distance, intensity targets, and simple notes.

## ImportedActivity

A completed activity imported from a provider such as Garmin and normalised into Runthread's own model. It should contain provider-neutral data such as start time, duration, distance, pace, heart rate summaries, and activity type.

## WorkoutMatch

The link between an imported activity and a planned workout. It records whether the match was automatic, manually selected, uncertain, or rejected.

## WorkoutResult

The evaluation of what happened against the planned workout. Examples include completed as planned, partially completed, missed, overperformed, underperformed, or moved.

## AdaptationEvent

A recorded plan change caused by new information, such as a missed workout, accumulated fatigue, or a changed schedule. It should include what changed, why it changed, and enough context to explain it later.

