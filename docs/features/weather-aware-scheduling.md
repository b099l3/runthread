# Weather-Aware Scheduling Implementation Plan

This feature uses weather forecasts to help runners and cyclists plan the best time to train on a given day.

The goal is not to cancel training automatically because the weather is imperfect. The goal is to make rain, temperature, wind, and daylight visible enough that the athlete can choose a sensible window or adjust the session safely.

## Product Shape

The app should show weather context alongside planned workouts:

- Expected rain chance and intensity.
- Temperature range during likely training windows.
- Wind speed and direction where available.
- Daylight timing where useful for early or late sessions.
- A short recommendation for the best time to run or cycle that day.
- Any practical caution, such as heat, cold, heavy rain, or strong wind.

Example:

```text
Today's run: 45 min easy

Best window: 07:00-09:00
Reason: dry, 14-16C, light wind

Avoid: 17:00-19:00
Reason: heavy rain likely
```

For cycling, the app should put more weight on wind and heavy rain because they affect safety and route choice more strongly than they do for an easy run.

## User Experience

Weather guidance should appear where the athlete is already making schedule decisions:

- Today's workout view.
- Weekly plan view.
- Rescheduling flow.
- Adaptation explanation when weather contributes to a suggested move.

The guidance should stay calm and practical:

- "Best before 09:00. Rain builds through the afternoon."
- "Cool and dry all day. Any normal training window is fine."
- "Strong winds this evening. Morning is better for cycling."
- "Hot after midday. Prefer an easy effort early or late."

The app should avoid over-precision. Forecasts are uncertain, so the copy should use plain ranges and explain the main reason for the suggestion.

## Weather Signals

Initial signals should be simple and provider-neutral:

- Temperature.
- Feels-like temperature.
- Rain probability.
- Rain intensity or precipitation amount.
- Wind speed.
- Wind gust speed where available.
- Weather condition summary.
- Forecast timestamp and location.

Later versions can add humidity, air quality, UV index, route exposure, surface conditions, and athlete-specific weather preferences.

## Recommendation Rules

The first version should use deterministic, testable rules.

For running:

- Prefer dry or low-rain windows.
- Avoid heavy rain where another suitable window exists.
- Prefer cooler windows on hot days.
- Avoid icy or unsafe conditions when known.
- Treat easy runs as more flexible than hard sessions or long runs.

For cycling:

- Prefer low wind and low-gust windows.
- Avoid heavy rain more strongly than for running.
- Flag crosswind or gust risk when route direction is known later.
- Prefer daylight for outdoor rides when available.
- Treat indoor trainer alternatives as valid suggestions when conditions are poor.

For hard workouts and long sessions:

- Prefer windows with stable conditions.
- Avoid moving demanding workouts into heat or strong wind when a safer option exists.
- If no good window exists, suggest reducing intensity, moving the session, or switching to an easier alternative.

## Scheduling Behaviour

Weather should inform scheduling without becoming the only decision.

When ranking candidate times or days, the scheduler should consider:

- Athlete availability.
- Existing workout hardness and recovery needs.
- Workout type and duration.
- Weather suitability for the activity.
- Forecast confidence and freshness.
- Whether a change preserves the purpose of the workout.

The scheduler should explain weather-related choices:

```text
Moved the easy run to the morning because rain is likely after 16:00.
```

```text
Kept the threshold session as planned. Conditions are warm but manageable, and moving it would place it too close to the long run.
```

## Backend Implementation Phases

### Phase 1: Domain Model

Add provider-neutral concepts for:

- Weather forecast.
- Weather forecast window.
- Weather condition.
- Weather suitability score.
- Weather-aware training recommendation.

These should reference athlete location and planned workout timing without coupling the core domain to a specific weather provider.

### Phase 2: Weather Provider Boundary

Add an integration boundary that can:

- Request a forecast for a location and date range.
- Normalise provider payloads into provider-neutral weather windows.
- Record forecast source and retrieval time.
- Handle missing, stale, or partial forecasts.

Provider-specific IDs, payloads, and API concepts should remain outside the domain model.

### Phase 3: Recommendation Rules

Implement deterministic rules that:

- Score forecast windows for running.
- Score forecast windows for cycling.
- Identify the best window for a planned workout.
- Identify avoidable weather risks.
- Produce a short explanation for the recommendation.

The rules should be testable without live weather API calls.

### Phase 4: Scheduling Integration

Use weather suitability when:

- Showing today's workout.
- Ranking same-day training windows.
- Suggesting a workout move.
- Explaining a weather-related schedule change.

Weather should not override recovery, injury, or readiness constraints. If training load rules say a move is unsafe, weather should not force it.

### Phase 5: App UI

Add weather guidance to the mobile app in a compact form:

- Weather summary on today's workout.
- Best training window.
- Key weather reason.
- Optional warning state.
- Weather-aware suggestion in rescheduling flows.

The UI should avoid turning the product into a weather app. It should show only the weather that changes the training decision.

## Acceptance Criteria

This feature is ready for a first private-beta version when:

- The app can show weather context for today's planned workout.
- The backend can represent provider-neutral forecast windows.
- Running and cycling have separate weather suitability rules.
- The system can recommend a best same-day training window.
- Weather recommendations include a short explanation.
- Missing or stale weather data fails gracefully.
- Weather-aware suggestions do not override safer training-load decisions.
- Tests cover rain, heat, cold, wind, stale forecasts, missing forecasts, and no-good-window cases.

## Assumptions

- The first version uses forecast data, not real-time nowcasting.
- Weather guidance is advisory unless the user chooses to reschedule.
- Athlete location can be derived from profile settings or an explicitly selected training location.
- The first version supports same-day guidance before multi-day weather-aware planning.
- AI-generated wording is not required; deterministic explanations are enough.
- Route-specific weather and surface condition modelling are out of scope for the first version.
