# Product

Runthread is a provider-neutral adaptive running training app for runners who want structured training without feeling micromanaged.

The app plans workouts, imports completed activities from connected activity providers, matches those activities to the plan, adapts future training when needed, and explains changes in calm, practical language.

Strava is the first likely MVP activity provider because it can validate the import loop quickly with a broadly available OAuth flow and activity API. Garmin direct integration remains an important premium/direct integration target after the provider-neutral import path is proven.

## Core User Problem

Many runners follow plans that assume perfect compliance. Real runners miss workouts, overperform, underperform, move days around, or arrive fatigued. Most plans either ignore this or require the runner to interpret what should change.

Runthread helps runners keep a useful plan even when real life changes the week.

## Core Product Loop

1. The app creates or maintains a training plan.
2. The runner completes a workout and records it with a connected provider.
3. Activity data is imported from Strava, Garmin, or another supported provider.
4. The imported provider activity is normalised into Runthread's `ImportedActivity` model.
5. The normalised activity is matched to the planned workout.
6. The workout result is evaluated.
7. The plan adapts if needed.
8. The runner sees a calm explanation of what changed and why.

In short:

```text
planned workout
  -> activity imported from provider
  -> activity normalised
  -> activity matched to planned workout
  -> plan adapts
```

## Initial Target User

The initial user is a recreational runner training for a race or fitness goal who already records runs with Strava, Garmin, or another common activity platform and wants structured guidance without hiring a coach.

The first version should work best for runners who:

- Run several times per week.
- Can connect a supported activity provider, with Strava likely first for MVP validation.
- Want simple adaptive guidance.
- Prefer clear explanations over complex training dashboards.

## Out of Scope for V1

- Real-time coaching during a run.
- Social features.
- Marketplace or coach management tools.
- COROS and Apple Health direct integrations.
- Full nutrition, strength, or injury management.
- AI-generated training decisions.
- Advanced workout export to watches.
- Complex multi-sport planning.
