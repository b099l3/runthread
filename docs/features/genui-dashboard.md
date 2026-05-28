# GenUI Dashboard Implementation Plan

This feature adds an athlete home dashboard composed through Flutter GenUI.

The goal is to give the runner a calm, useful first screen that adapts to their current training context without becoming a complex analytics dashboard. The dashboard should show what matters now, why it matters, and what the athlete can do next.

The implementation should take inspiration from Very Good Ventures' GenUI home-screen and architecture guidance: expose a curated catalog of approved Flutter widgets, let GenUI compose those widgets at runtime, and keep product logic, training decisions, and data access outside the generated UI layer.

## Product Shape

The dashboard is the app home screen for an athlete.

It should answer:

- What am I meant to do today?
- Has anything changed in my plan?
- Do I need to review a matched or unmatched activity?
- Do I need to answer a readiness or feedback prompt?
- Is my current week broadly on track?
- What is the most useful next action?

The first version should stay action-oriented. It should not become a dense training analytics surface.

Example dashboard modules:

- Today's workout.
- Current week timeline.
- Recent adaptation summary.
- Readiness prompt for demanding sessions.
- Activity match review prompt.
- Provider connection or sync status.
- One concise training insight.
- One primary next action.

## GenUI Role

GenUI is a dynamic presentation layer, not the training engine.

Runthread should compute structured dashboard context from deterministic services, then provide that context to a GenUI surface. GenUI can choose which approved dashboard components to render and how to arrange them, but it must not make training decisions or invent plan changes.

The core concepts are:

- Catalog: the approved Runthread dashboard widgets GenUI may use.
- Surface: the home-screen region where generated dashboard content appears.
- Conversation or context: the current athlete, plan, activity, readiness, and adaptation state used to compose the dashboard.

The catalog should be domain-specific. GenUI should not receive a broad generic widget set when composing the dashboard.

## Dashboard Context

The backend or app service should produce a provider-neutral dashboard context.

Initial context should include:

- Athlete ID and coaching detail level.
- Current date and local timezone.
- Current plan week.
- Today's planned workout, if any.
- Recent imported activities.
- Recent workout matches and unmatched activities.
- Recent adaptation events and explanations.
- Readiness requirement for today's workout.
- Pending post-workout feedback prompts.
- Provider connection and sync state.
- Known data freshness or degraded-data states.

The context should contain computed facts, not raw provider payloads. Provider-specific Garmin, Strava, or future-provider data should already be normalised before it reaches the dashboard surface.

## Component Catalog

The first dashboard catalog should include a small set of stable components:

- `WorkoutCard`: displays the next or current planned workout.
- `WeekTimeline`: displays the current training week at a glance.
- `AdaptationSummary`: explains a recent plan change.
- `ReadinessPrompt`: collects readiness before a demanding session.
- `ActivityMatchReview`: asks the athlete to confirm or correct an uncertain match.
- `TrainingInsightCard`: shows one computed, data-backed insight.
- `ProviderSyncStatus`: shows provider connection, sync, or degraded-data status.
- `NextAction`: presents the most useful action to take now.

Each catalog item should define:

- A clear name.
- A narrow JSON schema.
- Required and optional fields.
- Allowed actions.
- A normal Flutter widget builder.

The model may compose catalog items, but the app must validate generated payloads before rendering them.

## Interaction Model

Interactive dashboard widgets should route through normal Runthread app actions.

Examples:

- Submitting readiness calls the readiness service.
- Accepting or dismissing an adaptation recommendation calls the adaptation flow.
- Reviewing an activity match opens the match review flow.
- Connecting a provider opens the provider connection flow.

GenUI should not directly mutate training state. It can present actions and collect structured input, but mutations should pass through existing app services with validation and tests.

## Backend Implementation Phases

### Phase 1: Dashboard Context

Add a deterministic dashboard context builder that gathers the current athlete home-screen state.

The builder should assemble data from existing plan week, provider import, activity matching, adaptation, readiness, and profile services. Missing or stale data should be represented explicitly so the dashboard can render a useful fallback state.

### Phase 2: Mobile Static Fallback

Build a fixed Flutter dashboard using the same context and visual components planned for the GenUI catalog.

This fallback is required for:

- GenUI generation failure.
- Invalid generated payloads.
- Offline or degraded network states.
- Private-beta users excluded from the experiment.

### Phase 3: GenUI Catalog

Create a dashboard-specific Flutter GenUI catalog from the fallback dashboard components.

The catalog should expose only approved dashboard modules. It should avoid generic free-form layout widgets unless they are necessary and tightly constrained.

### Phase 4: GenUI Surface

Add a GenUI surface to the athlete home screen.

The surface should receive the dashboard context and system instructions that describe:

- The dashboard's purpose.
- The allowed catalog components.
- The priority order for urgent states.
- The requirement to stay concise.
- The rule that training decisions come only from provided context.

### Phase 5: Observability and Safety

Record dashboard generation metadata:

- Context version.
- Catalog version.
- Generated module list.
- Validation failures.
- Fallback reason.
- Generation latency.

Logs should support debugging without storing sensitive raw provider payloads or unnecessary personal data.

## Mobile Implementation Phases

### Phase 1: Shared Dashboard Widgets

Create reusable Flutter widgets for the static dashboard and the GenUI catalog builders.

The widgets should follow the app's existing navigation and state-management patterns. GenUI types should stay near the dashboard surface and catalog layer rather than leaking through app-wide models.

### Phase 2: Home Screen Integration

Make the dashboard the main signed-in home screen.

The screen should load dashboard context, render the fallback immediately when needed, and render the GenUI surface only when the feature is enabled and context is valid.

### Phase 3: Actions and Routing

Wire dashboard actions into existing flows:

- Workout detail.
- Readiness submission.
- Activity match review.
- Provider connection.
- Adaptation explanation.
- History or plan week detail.

The generated UI should never be the only path to a required action.

## Acceptance Criteria

This feature is ready for a first private-beta version when:

- The app can build a provider-neutral dashboard context.
- The athlete home screen can render a useful static fallback dashboard.
- The GenUI dashboard uses a curated Runthread component catalog.
- Generated payloads are validated before rendering.
- Invalid, slow, or failed generation falls back to the static dashboard.
- Dashboard actions route through deterministic app services.
- The dashboard can show today's workout, current week, pending readiness, activity match review, provider sync state, and recent adaptation explanations.
- GenUI cannot create training decisions not present in the dashboard context.
- Tests cover context construction, catalog schemas, fallback rendering, validation failures, and dashboard actions.

## Assumptions

- The first version is athlete-facing only.
- Coach-facing dashboards are out of scope.
- GenUI personalises presentation and module selection, not training logic.
- The first production-like version keeps a static dashboard fallback.
- Raw provider payloads are not sent to the GenUI layer.
- API keys or model credentials must not be embedded in shipped mobile clients.
- GenUI is treated as experimental during private beta until reliability, latency, and cost are understood.
