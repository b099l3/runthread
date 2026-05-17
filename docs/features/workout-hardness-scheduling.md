# Workout Hardness and Scheduling Implementation Plan

This feature assigns a before-and-after hardness profile to workouts so the scheduler can place sessions more intelligently and reschedule them safely.

The goal is to avoid bad workout stacking, especially around hard runs, long runs, strength sessions, and recovery days.

## Product Shape

Each planned workout should carry scheduling metadata:

- How hard the workout is expected to be.
- How much freshness it needs beforehand.
- How much recovery it needs afterward.
- Whether it can share a day with strength.
- Whether it can move by 24 hours.
- Which sessions it should not sit next to.

Example:

```text
Threshold session
Hardness: high
Needs: easy or rest day before
After: easy or rest day
Double-day: no heavy legs strength
```

## Scheduling Rules

Initial rules should cover:

- Rest before key hard sessions when needed.
- Rest or easy running after hard sessions.
- Avoiding hard run plus heavy strength on the same day.
- Allowing easy run plus light strength when appropriate.
- Avoiding long run immediately after high-intensity work.
- Preserving the purpose of key workouts when rescheduling.

Rules should be deterministic and testable.

## Rescheduling Behaviour

When a workout must move, the scheduler should check:

- Available days.
- Existing workout hardness.
- Required recovery windows.
- Strength session compatibility.
- Whether the moved session still fits the week.
- Whether dropping or reducing a session is safer than stacking.

The app should explain the selected change.

## Backend Implementation Phases

### Phase 1: Domain Model

Add provider-neutral concepts for:

- Workout hardness.
- Freshness requirement.
- Recovery requirement.
- Double-day compatibility.
- Scheduling constraint.

These should live with planned workout metadata and adaptation decisions, not provider activity payloads.

### Phase 2: Rule Engine

Implement deterministic scheduling rules for:

- Hard run placement.
- Long-run placement.
- Strength double days.
- Rest day protection.
- Recovery after demanding sessions.

### Phase 3: Rescheduling Service

Add service methods to:

- Evaluate whether a workout can move to a candidate day.
- Rank candidate days.
- Apply the safest reschedule.
- Explain rejected candidate days.

### Phase 4: Adaptation Integration

Use hardness rules when adapting after:

- Missed workouts.
- Low readiness.
- Injury or illness logs.
- Poor post-workout feedback.
- Schedule changes.

## Acceptance Criteria

This feature is ready for a first version when:

- Planned workouts carry expected hardness and recovery metadata.
- The scheduler can reject unsafe workout stacking.
- Strength double-day rules are represented.
- Rescheduling can choose a safer candidate day.
- Adaptation explanations include scheduling rationale.
- Tests cover hard-session spacing, long-run spacing, strength compatibility, and no-valid-day cases.

## Assumptions

- The first version uses simple categories rather than complex load modelling.
- Athlete schedule preferences are respected when available.
- If no safe reschedule exists, the system can recommend dropping or reducing the workout.
