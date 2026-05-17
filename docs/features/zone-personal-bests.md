# Zone Personal Bests Implementation Plan

This feature recognises personal bests within specific training zones, such as the athlete's fastest Zone 2 run over a distance or their longest Zone 2 duration.

The goal is to celebrate aerobic progress without encouraging every easy run to become a race.

## Product Shape

The app can surface achievements like:

```text
This was your fastest Zone 2 run at 8:30/mi.
```

Supported first-version examples:

- Fastest pace while staying in Zone 2.
- Longest Zone 2 run by distance.
- Longest Zone 2 run by time.
- Best average pace for a configured Zone 2 distance band.

These should be framed as progress signals, not instructions to chase speed during easy sessions.

## Zone Definition

The system should use the athlete's configured heart-rate zones where available.

If reliable heart-rate zones are missing, the app can avoid zone PB detection rather than guessing.

## Guardrails

Zone PBs should avoid rewarding bad execution:

- Do not award a Zone 2 PB if too much of the activity was above Zone 2.
- Do not award on short activities below a minimum duration or distance.
- Do not award if heart-rate data is missing or clearly unreliable.
- Do not make the PB messaging more prominent than plan compliance or recovery guidance.

## Backend Implementation Phases

### Phase 1: Domain Model

Add provider-neutral concepts for:

- Athlete training zones.
- Zone summary for imported activities.
- Zone personal best.
- PB category.
- PB eligibility.

Zone PBs should reference imported activities but should not depend on provider-specific activity IDs.

### Phase 2: Zone Summary Calculation

Calculate activity-level zone summaries from imported heart-rate data when available.

Initial outputs:

- Time in each zone.
- Percentage in target zone.
- Average pace.
- Distance.
- Duration.

If the current `ImportedActivity` model does not yet include enough samples, this feature should wait for a richer activity metrics model instead of forcing the calculation from averages only.

### Phase 3: PB Detection Rules

Implement deterministic PB rules:

- Fastest average pace for eligible Zone 2 run.
- Longest Zone 2 distance.
- Longest Zone 2 duration.

Eligibility should require a minimum Zone 2 percentage and minimum activity size.

### Phase 4: Product Surfacing

Show PBs in activity detail and periodic reports.

Reports can use these PBs as evidence of aerobic improvement.

## Acceptance Criteria

This feature is ready for a first version when:

- Athlete heart-rate zones can be configured or imported.
- Activity zone summaries can be calculated reliably.
- Zone PBs are only awarded when eligibility rules pass.
- PBs can be stored and compared against historical activities.
- Activity detail can show relevant Zone 2 PBs.
- Tests cover zone boundaries, eligibility, missing data, and PB replacement.

## Assumptions

- The first version focuses on running activities.
- Zone PB detection should be skipped when heart-rate data is not reliable enough.
- The feature should support aerobic progress recognition without incentivising overtraining.
