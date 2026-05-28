# Activity Match Review and Manual Completion

## Summary

Add a provider-neutral review flow for imported activities that do not automatically complete a planned workout. This phase improves the post-import experience after Strava/Garmin activities are normalised into `ImportedActivity`.

The goal is to let users inspect unmatched or uncertain imported activities, attach them to the correct planned workout, ignore them, or leave them as history-only activity. Automatic matching remains deterministic and conservative; manual review fills the gap when real training data does not line up cleanly with generated plans.

## Problem

The current import pipeline can successfully import provider activities, but many activities do not confidently match planned workouts. This is expected for historical backfills and generated plans because:

- Real activities may not match generated workout dates.
- Real distances and durations may differ from planned targets.
- Multiple activities may occur on the same day.
- One completed workout can attract multiple candidate activities.
- Some activities are useful history but should not complete a run workout.

Today these activities appear in History or on day cards, but there is no user-facing way to resolve them.

## Goals

- Show imported activities that are unmatched, uncertain, rejected, or skipped because the workout is already completed.
- Let the user manually attach an imported activity to a planned workout.
- Let the user ignore an imported activity for completion while keeping it in history.
- Prevent duplicate workout completion from multiple activities unless the user explicitly replaces the completion.
- Keep provider-specific details out of matching and completion UX.
- Preserve deterministic audit records for automatic and manual decisions.

## Non-Goals

- Do not loosen automatic matching so much that incorrect completions become common.
- Do not add AI-based matching.
- Do not expose provider tokens, raw provider payloads, or provider API calls to mobile.
- Do not implement Garmin live connection as part of this phase.
- Do not build full calendar editing or plan rescheduling in this phase.

## User Experience

### Plan Day Cards

Each day card can show:

- Planned workout details.
- Matched completion, if one exists.
- Imported activities on that day.
- A review state when an imported activity is not completing the workout.

Possible activity row states:

- `Completed this workout`
- `Possible match`
- `Not matched`
- `Ignored`
- `Workout already completed`

### Activity Review Sheet

Tapping an imported activity opens a review sheet with:

- Activity type, date/time, distance, duration, and source provider.
- Current match status and reason.
- Candidate planned workouts for the selected week and nearby days.
- Actions:
  - `Match to workout`
  - `Replace completion`
  - `Ignore for completion`
  - `Keep as history only`

### Manual Match Flow

When the user selects a planned workout:

1. Backend creates or updates a `WorkoutMatch` with `matched_by = manual`.
2. Backend creates or updates the `WorkoutResult`.
3. Planned workout status becomes completed.
4. Current plan/adaptation data refreshes.

If the selected workout already has a result, the UI must require an explicit replace action.

## Backend Design

### New App Services

Add provider-neutral services in `services/api/internal/app`:

- `ListActivityReviewItems`
  - Input: athlete ID, optional plan week ID or target week date.
  - Output: imported activities, existing matches/results, candidate workouts, review status.

- `ManuallyMatchImportedActivity`
  - Input: athlete ID, imported activity ID, planned workout ID, replacement mode.
  - Output: workout match, workout result, updated planned workout, optional adaptation event.

- `IgnoreImportedActivityForCompletion`
  - Input: athlete ID, imported activity ID, optional reason.
  - Output: review item status.

### Persistence

Add a provider-neutral review table or extend match state so manual decisions survive refreshes.

Preferred table:

```sql
CREATE TABLE imported_activity_review_decisions (
    id uuid PRIMARY KEY,
    athlete_id uuid NOT NULL REFERENCES athlete_profiles(id),
    imported_activity_id uuid NOT NULL REFERENCES imported_activities(id),
    planned_workout_id uuid REFERENCES planned_workouts(id),
    decision text NOT NULL,
    reason text,
    decided_by text NOT NULL,
    decided_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT imported_activity_review_decision_valid CHECK (
        decision IN ('manual_match', 'replace_completion', 'ignored', 'history_only')
    )
);
```

This keeps provider activity state separate from user review decisions.

### Matching Rules

Automatic matching should remain conservative:

- Confident match can auto-complete.
- Uncertain match should be reviewable.
- Rejected match should be reviewable only when the activity is on or near the selected week.
- Activities that match an already-completed workout should be marked as reviewable replacement candidates, not failures.

Manual match should:

- Use `domain.MatchSourceManual`.
- Require the activity and workout to belong to the same athlete.
- Require explicit replacement when the workout already has a result.
- Update existing deterministic match/result records instead of creating duplicates where possible.

## Mobile Design

### API Support

Add mobile API methods:

- `getActivityReviewItems(targetWeekDate)`
- `manualMatchImportedActivity(importedActivityId, plannedWorkoutId, replaceExisting)`
- `ignoreImportedActivity(importedActivityId, reason)`

### Screens

Plan tab:

- Keep day cards as the primary surface.
- Show imported activity rows inside the relevant day card.
- Use status labels and actions rather than a separate global activity card.

Activity review sheet:

- Bottom sheet or full-screen route on smaller devices.
- Candidate workouts grouped by date.
- Clear replace warning if a workout is already completed.

History tab:

- Continue showing recent imported activities.
- Add status text so history is understandable without opening the plan week.

## Test Plan

Backend:

- Lists unmatched imported activities for a week.
- Lists uncertain/rejected matches with reasons.
- Manual match creates a manual `WorkoutMatch`.
- Manual match creates a `WorkoutResult`.
- Manual match rejects cross-athlete activity/workout pairs.
- Manual match refuses to overwrite completed workouts unless replacement is explicit.
- Ignore decision prevents future retry from trying to complete the activity.
- Retry service respects ignored/history-only review decisions.

Mobile:

- Day card shows unmatched activity row.
- Review sheet opens from activity row.
- Candidate workout list renders nearby planned workouts.
- Manual match calls API and refreshes plan.
- Replace flow requires explicit confirmation.
- Ignore flow updates visible activity status.
- History shows review/completion status.

## Acceptance Criteria

- A user can resolve an unmatched Strava import without deleting or re-importing data.
- A user can manually attach a real activity to the correct planned workout.
- A user can ignore non-training or irrelevant activities while preserving history.
- Duplicate completions are prevented unless explicitly replaced.
- Automatic retry no longer repeatedly attempts ignored/history-only activities.
- All behavior is provider-neutral and works for future Garmin imports.

## Open Questions

- Should manual replacement delete the old workout result or mark it superseded?
- Should ignored activities be hidden from day cards by default?
- How many nearby days should candidate workouts include?
- Should replacement trigger adaptation recalculation immediately or defer it?
- Should review decisions be editable after submission?
