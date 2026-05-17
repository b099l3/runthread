# Workout Rationale Implementation Plan

This feature explains why a workout exists in the athlete's plan, what it is intended to improve, and how it fits their individual goal.

The goal is to make the plan feel purposeful without overwhelming the athlete with coaching theory.

## Product Shape

Each planned workout can include a short rationale:

```text
This threshold session improves your ability to hold a strong pace without tipping into unsustainable effort. It supports your 10K goal by building controlled speed endurance.
```

The rationale should explain:

- Why this workout is scheduled.
- What system or ability it improves.
- How hard it should feel.
- How it relates to the athlete's current goal.
- What matters most for execution.

## Rationale Types

Initial rationale can be attached to broad workout types:

- Easy run.
- Recovery run.
- Long run.
- Threshold or tempo.
- Intervals.
- Race-pace session.
- Strength.
- Rest day.

The explanation should become more individualised when the app knows the athlete's goal, training phase, recent history, and adaptation context.

## Backend Implementation Phases

### Phase 1: Rationale Templates

Add deterministic rationale templates for common workout types and training phases.

Templates should support variables such as:

- Goal type.
- Race distance.
- Plan week focus.
- Workout intensity.
- Target distance or duration.
- Recent adaptation reason.

### Phase 2: Domain Model

Add provider-neutral concepts for:

- Workout rationale.
- Rationale category.
- Training benefit.
- Execution focus.
- Source: planned, adapted, or manually assigned.

Rationale should be stored or generated from stable plan facts, not from provider-specific activity data.

### Phase 3: App Service

Add service methods to:

- Get rationale for a planned workout.
- Regenerate rationale after a plan adaptation.
- Include rationale in current plan week responses.

The mobile app should receive concise text and structured benefit tags.

### Phase 4: Adaptation Explanations

When a workout changes, the rationale should explain both:

- Why the workout exists.
- Why it changed.

Example:

```text
This was moved because yesterday's readiness score was low. Keeping the session tomorrow protects the purpose of the workout without stacking hard efforts.
```

## Acceptance Criteria

This feature is ready for a first version when:

- Planned workouts can expose a concise rationale.
- Rationale varies by workout type, goal, and training phase.
- Adapted workouts can explain why they changed.
- Mobile can display rationale without extra provider calls.
- Tests cover template selection, missing context, adapted workouts, and stable wording inputs.

## Assumptions

- The first version uses deterministic templates.
- AI-generated wording, if added later, should only summarise already-computed plan facts.
- Long coaching education content is out of scope for the first version.
