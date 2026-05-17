# Readiness to Train Implementation Plan

This feature adds a lightweight pre-workout readiness check before demanding sessions, then uses workout feedback and completion patterns to adapt training and produce useful athlete summaries.

The goal is not to replace coaching judgement. The goal is to help runners notice fatigue, avoid forcing hard work on bad days, and understand which parts of the plan they consistently complete, miss, or distort.

## Product Shape

Before hard sessions, the app asks:

```text
How energised do you feel before this hard training session?
1 = exhausted
10 = fully ready
```

Initial behaviour:

- `1-4`: recommend rest or an easy replacement session.
- `5-6`: offer to delay the hard session by 24 hours and replace today with low intensity training.
- `7-10`: keep the workout as planned.

The first version should keep the decision deterministic and explainable. Automatic changes can come later once the user has seen and trusted the recommendations.

## Sessions That Need Readiness

Readiness checks should only appear before sessions with meaningful load or risk:

- High-intensity intervals.
- Threshold or tempo sessions.
- Long runs.
- Race-pace sessions.
- Key progression workouts.

Easy runs, recovery runs, rest days, and optional strength sessions should not require a readiness check by default.

## Post-Workout Feedback

After each completed session, the app should collect simple athlete feedback:

- Perceived difficulty.
- How the runner felt.
- Optional short note.

This should be independent of provider data. Garmin, Strava, and future providers can supply activity metrics, but the subjective feedback belongs to Runthread.

## Pattern Detection

The system should look for repeated patterns across recent training, including:

- Low readiness before hard sessions.
- Poor post-workout feeling after demanding sessions.
- Falling completion rate.
- Repeated partial completion of long runs.
- Missed high-intensity sessions.
- Easy runs completed too fast or at too high a heart rate.
- Runners consistently cutting short a specific session type.

These patterns should inform conservative progression or regression decisions.

Progression signals:

- High completion rate.
- Hard sessions completed as planned.
- Easy volume kept genuinely easy.
- Positive or stable subjective feedback.

Regression or recovery signals:

- Repeated tiredness or poor feeling.
- Multiple missed or shortened key sessions.
- Hard efforts completed poorly.
- Easy runs regularly completed too hard.
- Training load increasing while readiness falls.

## Monthly Training Report

The app should generate an automated summary that gives the athlete specific, practical focus areas.

Example report shape:

```text
Great month of training.

Total volume: 300 km
Low-intensity volume: 260 km
High-intensity percentage: 14%

This is a manageable training load and fits a nicely polarised model.

You had 3 long runs this month. They were planned as 18, 20, and 21 miles. You completed 18 miles in all 3 workouts.

Well done for completing the long runs. Increasing long-run distance next month is an important target as you build towards the marathon.
```

Reports should highlight:

- Total volume.
- Low-intensity volume based on heart rate and subjective feel where available.
- High-intensity percentage.
- Completion rate.
- Missed-session patterns by workout type.
- Overperformed easy sessions.
- Long-run completion against planned distance or duration.
- One to three practical focus points for the next block.

## Backend Implementation Phases

### Phase 1: Domain Model

Add provider-neutral domain concepts for:

- Workout readiness check.
- Readiness score from `1` to `10`.
- Readiness recommendation.
- Post-workout subjective feedback.
- Workout type grouping for pattern analysis.

Readiness and feedback should reference planned workouts and athlete IDs, not provider activity IDs.

### Phase 2: Readiness Recommendation Rules

Implement deterministic rules:

- `1-4`: recommend rest or easy replacement.
- `5-6`: recommend delaying the hard session by 24 hours, with a low-intensity replacement.
- `7-10`: recommend proceeding as planned.

The recommendation should include a short explanation and a proposed plan change, but applying the change can remain a separate action.

### Phase 3: App Service and Persistence

Add service methods to:

- Determine whether a workout needs readiness.
- Submit a readiness score.
- Return the readiness recommendation.
- Apply or dismiss the recommendation.
- Submit post-workout feedback.

Persist readiness checks, recommendations, user decisions, and subjective feedback.

### Phase 4: Adaptation Integration

Feed readiness and subjective feedback into the existing deterministic adaptation engine.

Initial adaptations should be conservative:

- Add rest or easy replacement when readiness is low.
- Delay a hard session after medium readiness if the user accepts.
- Reduce near-term intensity after repeated poor feedback.
- Avoid progressing long-run distance if long runs are repeatedly shortened.

### Phase 5: Pattern Analysis

Create deterministic summaries over a recent training window, likely 4 weeks by default.

The analysis should calculate:

- Completion rate.
- Completion rate by workout type.
- Planned versus completed distance or duration.
- Long-run completion trends.
- High-intensity completion trends.
- Easy-run intensity discipline.
- Readiness and feedback trends.

### Phase 6: Athlete Report

Generate a structured report from the pattern analysis.

The report should be explainable and data-backed. Any AI-generated wording, if used later, must be constrained to summarising already-computed facts and must not make training decisions.

## Acceptance Criteria

This feature is ready for a first private-beta version when:

- Hard sessions can require a pre-workout readiness score.
- Readiness scores produce deterministic recommendations.
- The athlete can accept or dismiss a recommended change.
- Post-workout difficulty and feeling can be recorded.
- Readiness and feedback are persisted separately from provider activity data.
- The adaptation engine can consume readiness and feedback signals.
- Pattern analysis can identify missed, shortened, overperformed, and poorly paced session types.
- Monthly reports summarise volume, intensity distribution, completion patterns, and practical focus areas.
- Tests cover readiness scoring, recommendation boundaries, feedback validation, pattern summaries, and adaptation behaviour.

## Assumptions

- Readiness checks are only required for demanding sessions in the first version.
- The first version recommends changes before automatically applying them.
- Training decisions remain deterministic and testable.
- Provider data can help classify intensity, but subjective feedback is captured directly in Runthread.
- Coach-facing dashboards and manual coach intervention are out of scope for the first version.
