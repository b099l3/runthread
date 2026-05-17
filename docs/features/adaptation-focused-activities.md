# Adaptation-Focused Activities Implementation Plan

This feature shifts planning language from activity labels to adaptation targets. Instead of thinking only in terms of run, ride, or walk, Runthread can ask what training stimulus is needed.

The goal is to make the plan more flexible while keeping the adaptation purpose clear.

## Product Shape

A planned session can express the intended adaptation:

```text
Target: 2 hours Zone 2
Preferred mode: run
Acceptable alternatives: bike, elliptical
Purpose: aerobic endurance
```

Another example:

```text
Target: 10 minutes Zone 4
Preferred mode: run intervals
Purpose: high-intensity aerobic stimulus
```

The athlete still sees practical workouts, but the backend understands the deeper requirement: time in zone, duration, intensity, and training purpose.

## Adaptation Targets

Initial target types:

- Zone 2 duration.
- Zone 2 distance.
- Zone 4 time.
- Long aerobic duration.
- Recovery movement.
- Strength stimulus.

Targets can be satisfied by different activity modes if the mode produces the intended stimulus.

## Matching Behaviour

Activity-to-workout matching should consider:

- Did the activity hit the intended intensity zone?
- Did it satisfy the target duration or distance?
- Was the mode acceptable for the adaptation?
- Did it create more fatigue than expected?

Example:

- A 2-hour easy bike ride may satisfy an aerobic Zone 2 target.
- It should not fully satisfy a marathon-specific long run if the plan needed impact tolerance and running-specific endurance.

## Backend Implementation Phases

### Phase 1: Domain Model

Add provider-neutral concepts for:

- Training stimulus.
- Adaptation target.
- Preferred activity mode.
- Acceptable activity modes.
- Target specificity.
- Stimulus completion.

Planned workouts can keep their current type while also carrying adaptation target metadata.

### Phase 2: Planning Integration

Allow the planner to attach adaptation targets to workouts.

Examples:

- Easy run: Zone 2 aerobic duration.
- Long run: aerobic duration plus running specificity.
- Intervals: Zone 4 accumulated time.
- Recovery: low-intensity movement.

### Phase 3: Matching and Completion

Extend matching and workout result logic to evaluate stimulus completion, not only activity type, distance, and duration.

The result should distinguish:

- Completed intended activity.
- Completed equivalent stimulus.
- Partially completed stimulus.
- Completed activity but missed intended stimulus.

### Phase 4: Adaptation Integration

Use stimulus completion to adapt future training.

Examples:

- If the athlete missed a run but completed equivalent Zone 2 work, reduce the penalty.
- If the athlete ran but spent too much time above Zone 2, treat the easy stimulus as poorly executed.
- If marathon-specific long runs are repeatedly replaced with cycling, flag missing specificity.

## Acceptance Criteria

This feature is ready for a first version when:

- Planned workouts can carry adaptation target metadata.
- Imported activities can be evaluated against a training stimulus.
- Matching can recognise acceptable alternative modes.
- Workout results can distinguish activity completion from stimulus completion.
- Adaptation can use stimulus completion when deciding progression or regression.
- Tests cover equivalent aerobic work, missed specificity, wrong intensity, and partial stimulus completion.

## Assumptions

- Running remains the primary activity for race-specific plans.
- Alternative modes can satisfy some aerobic targets but not all sport-specific targets.
- The first version should use deterministic rules and avoid complex physiological modelling.
