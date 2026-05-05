# Garmin Integration

Garmin is the first activity provider for Runthread.

The integration should be isolated from the core domain. Garmin-specific identifiers, payloads, API details, and provider terminology should stay inside the integration and persistence layers. Core training logic should use Runthread domain models such as `ImportedActivity`, `WorkoutMatch`, and `WorkoutResult`.

Real Garmin API access is not assumed to be available during early development. Stage 8 should include validating Garmin's current access process, API capabilities, provider terms, webhook or polling model, and test data options before implementing production OAuth or import flows.

## Intended Flow

1. The user connects Garmin from the mobile app.
2. The backend completes the provider connection flow.
3. The backend stores a provider connection for the athlete.
4. Garmin activities are imported by webhook, polling, or another supported provider mechanism.
5. Raw provider data is stored only where useful for audit/debugging.
6. Activities are normalised into Runthread's `ImportedActivity` model.
7. Imported activities are matched to planned workouts.
8. Matched activities produce workout results.
9. Workout results may trigger adaptation events.

## Boundary Rule

Garmin-specific data must not leak into core domain logic.

The planning, matching, and adaptation packages should depend on provider-neutral concepts. If Coros or Apple Watch is added later, those integrations should also normalise into the same domain models.
