# Injury Log Implementation Plan

This feature lets athletes record injury, pain, soreness, and illness signals so the plan can adapt conservatively when training risk increases.

The goal is to capture enough context to protect training continuity without turning Runthread into a medical product.

## Product Shape

The athlete can log an injury or physical issue with:

- Body area.
- Severity.
- Start date.
- Current status.
- Impact on running.
- Optional notes.

Example:

```text
Left calf tightness
Severity: 4/10
Started: Monday
Impact: can run easy, avoiding speed
```

The app can then recommend safer plan changes, such as replacing intensity with easy running, reducing long-run load, adding rest, or asking the athlete to pause training if severity is high.

## Log Types

The first version should support broad issue types:

- Injury.
- Pain or soreness.
- Illness.
- General limitation.

The app should not diagnose conditions. It should record athlete-reported information and adapt training risk conservatively.

## Adaptation Signals

The system should respond to:

- High severity.
- Worsening severity.
- Pain that persists across multiple days.
- Pain that appears after hard sessions.
- Repeated missed workouts linked to the same issue.
- A body area that is affected by planned workouts.

Initial recommendations should be cautious and explainable.

## Backend Implementation Phases

### Phase 1: Domain Model

Add provider-neutral concepts for:

- Athlete issue log.
- Issue type.
- Body area.
- Severity from `1` to `10`.
- Status: active, improving, resolved, recurring.
- Training impact.

Injury records should belong to the athlete and can optionally reference planned workouts or workout results.

### Phase 2: App Service and Persistence

Add service methods to:

- Create an issue log entry.
- Update severity and status.
- Mark an issue resolved.
- List active issues.
- Link an issue to a missed, skipped, or modified workout.

Persist issue history rather than only the latest state, so the system can detect worsening or improvement.

### Phase 3: Adaptation Integration

Feed active injury signals into the deterministic adaptation engine.

Initial rules:

- High severity should recommend rest and avoid hard workouts.
- Moderate severity should remove or delay intensity.
- Persistent issues should block progression.
- Resolved issues should allow gradual return, not immediate aggressive progression.

### Phase 4: Reporting

Reports should summarise:

- Days affected by injury or illness.
- Workouts missed or modified because of an issue.
- Recurring issue patterns.
- Return-to-training progress.

## Acceptance Criteria

This feature is ready for a first version when:

- Athletes can create and update injury or issue logs.
- Active issues can influence training recommendations.
- The adaptation engine avoids intensity or progression when risk is high.
- Missed or modified workouts can be linked to an issue.
- Reports can summarise issue impact on training.
- Tests cover validation, status changes, severity boundaries, and adaptation behaviour.

## Assumptions

- Runthread records athlete-reported issues but does not diagnose or treat injuries.
- Medical advice, clinician workflows, and rehabilitation plans are out of scope.
- The first version should recommend changes before automatically applying them.
