# Activity Fuel Log Implementation Plan

This feature lets athletes record fuelling around individual activities so Runthread can identify whether poor execution, fatigue, or long-run underperformance may be linked to inadequate fuel.

The goal is simple training insight, not nutrition coaching.

## Product Shape

After selected sessions, the athlete can log:

- Pre-workout fuel.
- During-workout fuel.
- Fluids.
- Caffeine.
- Any stomach issues.
- Optional notes.

Example:

```text
Long run fuel
Before: porridge and coffee
During: 2 gels, 500 ml water
Stomach: fine
```

The app can use this context when explaining difficult sessions or repeated long-run fade patterns.

## Sessions That Need Fuel Logging

Fuel logging should be optional by default and prompted more strongly after:

- Long runs.
- Race-pace workouts.
- High-intensity sessions over a configured duration.
- Races.
- Workouts that were underperformed or cut short.

Easy short runs should not require a fuel log.

## Pattern Detection

The system should look for:

- Long runs completed without during-run fuel.
- Underperformed long sessions with low fuel.
- Stomach issues associated with certain fuel types.
- Better outcomes when fuelling is consistent.
- Race-specific fuelling practice before race day.

The app should use these patterns to produce practical prompts, not prescriptive nutrition plans.

## Backend Implementation Phases

### Phase 1: Domain Model

Add provider-neutral concepts for:

- Activity fuel log.
- Fuel timing: before, during, after.
- Fuel item.
- Fluid amount.
- Caffeine flag or amount.
- Gastrointestinal comfort.
- Notes.

Fuel logs should reference workout results or imported activities, while remaining independent from provider payloads.

### Phase 2: App Service and Persistence

Add service methods to:

- Create or update a fuel log for an activity.
- Retrieve fuel logs with workout results.
- Identify activities where fuel logging should be prompted.

Persist fuel data as structured fields where useful, with notes for free-form detail.

### Phase 3: Pattern Analysis

Add deterministic analysis for:

- Fuel present or missing on long sessions.
- Fuel consistency across similar workouts.
- Fuel notes linked to underperformance.
- Stomach issue frequency.

### Phase 4: Report Integration

Training reports can include fuelling observations when they are relevant.

Example:

```text
Your long runs were more consistent when you fuelled during the run. On the two long runs without fuel, you cut the final section short.
```

## Acceptance Criteria

This feature is ready for a first version when:

- Athletes can log fuel against an activity or workout result.
- Long and demanding sessions can prompt optional fuel logging.
- Fuel logs are stored separately from provider activity payloads.
- Pattern analysis can identify basic fuel-related trends.
- Reports can include fuel observations when supported by data.
- Tests cover validation, prompt eligibility, persistence, and report summaries.

## Assumptions

- Runthread does not prescribe exact nutrition plans in the first version.
- Fuel entries can start simple and become more structured later.
- Provider integrations are not expected to supply detailed fuel data.
