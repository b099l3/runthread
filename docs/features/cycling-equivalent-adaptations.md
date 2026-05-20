# Cycling Equivalent Adaptations Implementation Plan

This feature allows Runthread to support cycling activities as first-class training inputs and, where appropriate, adapt planned runs into cycling equivalents.

The goal is to give athletes useful alternatives when running is not possible while keeping the training purpose honest. Cycling can replace some aerobic run stimulus, but it should not be treated as identical to running.

## Product Shape

When a planned run is eligible for a cycling equivalent, the app can present an alternative:

```text
Original: 45 minute easy run
Alternative: 60 minute easy ride
Purpose: aerobic endurance
Reason: maintains low-intensity aerobic work while reducing impact
```

Another example:

```text
Original: 90 minute long run
Alternative: 2 hour Zone 2 ride
Purpose: aerobic endurance
Limitation: does not fully replace running-specific endurance or impact tolerance
```

The athlete should understand both what the ride preserves and what it does not preserve.

## Supported Activity Modes

Initial cycling support should include:

- Outdoor ride.
- Indoor ride.
- Stationary bike.
- Smart trainer ride.

Provider-specific activity names should map into provider-neutral cycling modes before adaptation or matching logic uses them.

## Run-to-Cycling Equivalents

Cycling equivalents should be allowed for run sessions where the main purpose is transferable aerobic stimulus:

- Easy aerobic runs.
- Recovery runs.
- Low-intensity base runs.
- Some general aerobic duration targets.
- Low-readiness replacements for harder sessions.

Cycling equivalents should be restricted or marked partial for sessions with strong running specificity:

- Race-pace runs.
- Marathon-specific long runs.
- Hill sprints.
- Running drills.
- Speed sessions where mechanics, cadence, or impact tolerance matter.
- Late-stage race preparation workouts.

The first version should prefer conservative equivalence rules over aggressive conversion.

## Equivalence Rules

Initial deterministic rules can consider:

- Intended training purpose.
- Planned run duration.
- Planned run intensity.
- Athlete goal and race distance.
- Training phase.
- Recent running volume.
- Injury or readiness context.

Example starting rules:

- Easy run to easy ride: allowed when intensity stays low.
- Recovery run to easy ride: allowed when the goal is movement and circulation.
- Long run to long ride: partial credit unless the plan explicitly allows non-running aerobic alternatives.
- Interval run to bike intervals: only allowed when the target is cardiovascular intensity, not running mechanics or race specificity.

Equivalence should usually be based on time and intensity, not distance.

## Matching Behaviour

When an athlete completes a ride against a planned run, matching should answer:

- Was cycling an acceptable alternative for this planned workout?
- Did the ride hit the intended intensity?
- Did the ride provide enough duration for the target stimulus?
- Should the workout count as complete, equivalent, partial, or missed?
- Did the substitution leave a running-specific gap that should affect future planning?

Example outcomes:

- Completed equivalent stimulus: easy run replaced by a low-intensity ride of sufficient duration.
- Partially completed stimulus: long run replaced by a long ride, with aerobic work preserved but running specificity missed.
- Missed intended stimulus: race-pace run replaced by an easy ride.

## Adaptation Behaviour

Adaptation should use cycling substitutions conservatively:

- Reduce penalties for missed easy aerobic runs when an equivalent ride was completed.
- Preserve recovery intent when a ride replaces a recovery run.
- Avoid progressing running-specific load based only on cycling completion.
- Flag repeated ride substitutions when the athlete is preparing for a running race.
- Use cycling as a lower-impact alternative when injury risk, soreness, or readiness suggests running is not appropriate.

## Backend Implementation Phases

### Phase 1: Activity Mode Normalisation

Add provider-neutral cycling activity modes and map Garmin, Strava, and future provider ride types into them.

### Phase 2: Equivalence Model

Add deterministic run-to-cycling equivalence rules based on purpose, intensity, duration, specificity, and athlete context.

### Phase 3: Matching Integration

Allow completed rides to match planned runs when cycling is an acceptable alternative or partial substitute.

### Phase 4: Adaptation Integration

Use ride completion when adapting future training while keeping running-specific progression separate from general aerobic progression.

### Phase 5: Athlete Explanation

Explain substitutions clearly in the app, including whether the ride fully satisfied the workout, partially satisfied it, or preserved only the aerobic purpose.

## Acceptance Criteria

This feature is ready for a first version when:

- Ride and cycling provider activities map into provider-neutral activity modes.
- Planned runs can declare whether cycling is an acceptable alternative.
- Matching can classify rides against planned runs as equivalent, partial, or not equivalent.
- Adaptation can reduce missed-workout penalties for acceptable cycling equivalents.
- Running-specific progression is not advanced solely from cycling substitutions.
- Athlete-facing explanations describe both the benefit and limitation of the substitution.
- Tests cover easy run replacement, recovery replacement, long-run partial credit, race-specific rejection, wrong intensity, and repeated substitutions.

## Assumptions

- Running remains the primary sport for running race plans.
- Cycling can preserve aerobic stimulus but does not fully replace impact tolerance, running economy, or race-specific mechanics.
- The first version uses deterministic rules rather than physiological load modelling.
- Equivalence rules should be intentionally conservative until real athlete outcomes justify broader substitution.
