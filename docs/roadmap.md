# Roadmap

## Stage 0: Project Foundation

Acceptance criteria:

- Monorepo structure exists.
- Product, architecture, domain, Garmin, roadmap, and decisions docs exist.
- Codex working agreement and session template exist.
- Minimal Go API service starts locally.
- Local Postgres is defined in Docker Compose.

## Stage 1: Backend Domain Model

Acceptance criteria:

- Core domain types are defined in Go.
- Domain types are provider-neutral.
- Basic validation exists for required fields and obvious invalid values.
- Unit tests cover domain construction and validation.
- Docs are updated if domain concepts change.

## Stage 2: Deterministic Planning Engine

Acceptance criteria:

- A simple deterministic planner can generate an initial training week or short plan from an athlete profile and goal.
- Planning rules are explicit and testable.
- No AI is used for training decisions.
- Tests cover common beginner/intermediate cases.

## Stage 3: Workout Completion/Missed Flow

Acceptance criteria:

- Planned workouts can be marked completed, missed, skipped, or moved.
- Workout results are represented clearly in the domain.
- Backend exposes minimal RPC endpoints or service methods for the flow.
- Tests cover the main status transitions.

## Stage 4: Garmin Activity Import Mock

Acceptance criteria:

- Mock Garmin activity payloads can be ingested locally.
- Provider data is normalised into `ImportedActivity`.
- Garmin-specific details stay inside the Garmin integration package.
- Tests cover representative activity examples.

## Stage 5: Activity-to-Workout Matching

Acceptance criteria:

- Imported activities can be matched to planned workouts using deterministic rules.
- The system can represent confident, uncertain, and rejected matches.
- Manual override is supported at the domain or service layer.
- Tests cover date, distance, duration, and type matching cases.

## Stage 6: Adaptation Engine

Acceptance criteria:

- Deterministic adaptation rules respond to missed, partial, and overperformed workouts.
- Adaptation events record what changed and why.
- Plan changes are limited, explainable, and testable.
- Tests cover common adaptation scenarios.

## Stage 7: Flutter MVP Screens

Acceptance criteria:

- Flutter app can show the current plan, workout details, imported activity status, and adaptation explanations.
- App talks to backend via ConnectRPC.
- UI supports the core loop without provider setup complexity.
- Basic loading and error states exist.

## Stage 8: Real Garmin Integration

Acceptance criteria:

- Garmin developer/API access path is validated, including current provider terms, authentication requirements, data access model, and test data options.
- Users can connect Garmin through the supported provider flow.
- Activities import into the backend.
- Imported activities are normalised and matched to planned workouts.
- Integration failures are logged and recoverable.

## Stage 9: Subscriptions and Beta Launch

Acceptance criteria:

- Subscription flow is implemented for the chosen platform and backend model.
- Beta users can onboard, connect Garmin, view a plan, and receive adaptations.
- Operational monitoring and support paths exist.
- Privacy, data deletion, and provider terms are reviewed.
