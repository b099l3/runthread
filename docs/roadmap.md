# Roadmap

## Stage 0: Project Foundation

Acceptance criteria:

- Monorepo structure exists.
- Product, architecture, domain, Garmin, roadmap, and decisions docs exist.
- Codex working agreement and session template exist.
- Minimal Go API service starts locally.
- Local Postgres is defined in Docker Compose.

## Stage 1: Backend Domain Model

Status: Completed initial pass.

Acceptance criteria:

- Core domain types are defined in Go.
- Domain types are provider-neutral.
- Basic validation exists for required fields and obvious invalid values.
- Unit tests cover domain construction and validation.
- Docs are updated if domain concepts change.

Notes:

- Initial domain structs and validation live in `services/api/internal/domain`.
- Provider names, provider activity IDs, raw payloads, and OAuth details are intentionally excluded from core domain activity types.
- Future stages may refine fields as planning, matching, persistence, and RPC contracts become concrete.

## Stage 2: Deterministic Planning Engine

Status: Completed initial weekly planner pass.

Acceptance criteria:

- A simple deterministic planner can generate an initial training week or short plan from an athlete profile and goal.
- Planning rules are explicit and testable.
- No AI is used for training decisions.
- Tests cover common beginner/intermediate cases.

Notes:

- `services/api/internal/planning` can generate one seven-day `PlanWeek` from an athlete profile, training goal, and target week date.
- The initial planner supports easy runs, a Sunday-default long run, one threshold workout for intermediate runners, rest days, and an optional strength day for intermediate runners.
- Volume is based on current weekly distance and goal distance, with conservative caps to avoid aggressive increases.
- The planner is deterministic and has no persistence, Garmin, RPC, Flutter, or AI dependency.

## Stage 3: Workout Completion/Missed Flow

Status: Completed initial domain/service pass.

Acceptance criteria:

- Planned workouts can be marked completed, missed, skipped, or moved.
- Workout results are represented clearly in the domain.
- Backend exposes minimal RPC endpoints or service methods for the flow.
- Tests cover the main status transitions.

Notes:

- Stage 3 is implemented as deterministic domain helpers in `services/api/internal/domain`.
- Marking completed, missed, skipped, or moved returns an updated `PlannedWorkout` and a `WorkoutResult`.
- No RPC, database persistence, Garmin integration, Flutter app, or AI was added.
- RPC endpoints remain deferred until the API contract stage.

## Stage 4: Garmin Activity Import Mock

Status: Completed initial mock normalisation pass.

Acceptance criteria:

- Mock Garmin activity payloads can be ingested locally.
- Provider data is normalised into `ImportedActivity`.
- Garmin-specific details stay inside the Garmin integration package.
- Tests cover representative activity examples.

Notes:

- `services/api/internal/garmin` contains a mock Garmin payload type and normaliser.
- Garmin-specific activity IDs and activity labels stay inside the Garmin package.
- The normaliser outputs provider-neutral `domain.ImportedActivity` values.
- Real Garmin OAuth, webhooks, polling, database persistence, RPC, Flutter, and AI remain out of scope.

## Stage 5: Activity-to-Workout Matching

Status: Completed initial deterministic matcher pass.

Acceptance criteria:

- Imported activities can be matched to planned workouts using deterministic rules.
- The system can represent confident, uncertain, and rejected matches.
- Manual override is supported at the domain or service layer.
- Tests cover date, distance, duration, and type matching cases.

Notes:

- `services/api/internal/matching` contains provider-neutral matching rules.
- Automatic matching compares activity type, calendar date gap, target distance, and target duration.
- Matches can be high-confidence matched, medium-confidence uncertain, or low-confidence rejected.
- Manual override is represented with a manual high-confidence `WorkoutMatch`.
- No real Garmin integration, database persistence, RPC, Flutter, or AI was added.

## Stage 6: Adaptation Engine

Status: Completed initial deterministic event pass.

Acceptance criteria:

- Deterministic adaptation rules respond to missed, partial, and overperformed workouts.
- Adaptation events record what changed and why.
- Plan changes are limited, explainable, and testable.
- Tests cover common adaptation scenarios.

Notes:

- `services/api/internal/adaptation` produces `AdaptationEvent` values from workout results and the current plan week.
- Initial rules cover missed, partially completed, overperformed, and underperformed workouts.
- Changes are conservative: adjust the next relevant workout when possible, otherwise add a plan note.
- Completed-as-planned results do not produce adaptation events.
- No AI, Flutter, database persistence, ConnectRPC, or real Garmin integration was added.

## Backend Core Loop Consolidation

Status: Completed initial in-memory test harness.

Acceptance criteria:

- The backend core loop can be exercised end to end in unit tests.
- Planning, mock activity import, activity matching, workout result creation, and adaptation are wired together.
- Provider-specific mock Garmin code stays outside the provider-neutral orchestrator.
- No database persistence, ConnectRPC, Flutter, AI, or real Garmin integration is added.

Notes:

- `services/api/internal/coreloop` contains an in-memory orchestration harness for backend tests.
- The harness accepts a provider-neutral activity importer function returning `domain.ImportedActivity`.
- End-to-end tests inject the Garmin mock normaliser from the test side.
- Completed-as-planned activities produce no adaptation event; underperformed activities produce an adaptation event.

## Application Service Boundaries

Status: Completed initial in-memory boundary pass.

Acceptance criteria:

- Future API handlers have a thin application service boundary to call.
- The boundary remains in-memory and test-only.
- Provider-specific code does not leak into the application service.
- No ConnectRPC, database persistence, sqlc, Flutter, AI, or real Garmin integration is added.

Notes:

- `services/api/internal/app` contains the first API-shaped Go service boundary.
- `CoreLoopService.CompleteImportedActivity` wraps the in-memory core-loop harness with request/response structs and `context.Context`.
- This is a preparation layer for future RPC handlers, not an RPC implementation.
- Likely future RPC endpoints are documented in `docs/architecture.md`.

## Repository Boundaries

Status: Completed initial in-memory boundary pass.

Acceptance criteria:

- Repository interfaces exist for core-loop records.
- In-memory implementations can store and retrieve core records.
- In-memory repositories validate records on save.
- Tests cover save/get, missing records, validation, and copy isolation.
- No Postgres, sqlc, ConnectRPC, Flutter, AI, or real Garmin integration is added.

Notes:

- `services/api/internal/repository` defines repository interfaces and `InMemoryStore`.
- The store currently supports `AthleteProfile`, `TrainingGoal`, `PlanWeek`, `PlannedWorkout`, `ImportedActivity`, `WorkoutMatch`, `WorkoutResult`, and `AdaptationEvent`.
- These interfaces are the intended boundary for later Postgres/sqlc implementations.

## App Service Persistence Wiring

Status: Completed initial in-memory wiring pass.

Acceptance criteria:

- Successful core-loop app service calls persist core records through repository interfaces.
- Persisted records can be retrieved from the in-memory store in tests.
- The app service remains thin and does not depend on Postgres/sqlc.
- No ConnectRPC, Flutter, AI, or real Garmin integration is added.

Notes:

- `CoreLoopService.CompleteImportedActivity` now saves athlete profile, training goal, plan week, each planned workout in the week, imported activity, workout match, workout result, and adaptation event when present.
- The saved plan week and first-class planned workout records include the updated completed workout status.
- Persistence still uses repository interfaces; `InMemoryStore` is only the current test implementation.

## Postgres Persistence Model Planning

Status: Completed design document.

Acceptance criteria:

- Initial Postgres tables are documented before migrations are written.
- Table relationships, primary keys, key foreign keys, and important columns are described.
- Provider-specific Garmin data is explicitly kept out of core domain tables.
- Real Garmin persistence concerns are deferred until the Garmin integration stage.
- No migrations, sqlc, ConnectRPC, Flutter, AI, or real Garmin integration is added.

Notes:

- See `docs/persistence.md`.
- The design covers `athlete_profiles`, `training_goals`, `plan_weeks`, `planned_workouts`, `imported_activities`, `workout_matches`, `workout_results`, `adaptation_events`, and `adaptation_event_changes`.

## Initial Postgres Migrations

Status: Completed plain SQL migration pass.

Acceptance criteria:

- Initial core schema migration files exist.
- The migration creates core provider-neutral tables only.
- Provider-specific Garmin tables remain deferred.
- No migration tool, sqlc, ConnectRPC, Flutter, AI, or real Garmin integration is added.

Notes:

- Plain SQL migration files live in `services/api/internal/postgres/migrations`.
- The schema follows `docs/persistence.md`.
- A migration tool still needs to be chosen before local automated migration commands are added.

## Local Migration Workflow

Status: Completed minimal script pass.

Acceptance criteria:

- Local Postgres can be started with Docker Compose.
- The initial schema can be applied with a simple command.
- The initial schema can be rolled back with a simple command.
- No sqlc, ConnectRPC, Flutter, AI, or real Garmin integration is added.

Notes:

- `scripts/db/migrate-up.sh` applies the initial SQL schema to the Docker Compose Postgres service.
- `scripts/db/migrate-down.sh` rolls back the initial SQL schema.
- These scripts use `psql` inside the Postgres container and do not track migration versions.
- sqlc query generation and repository implementation are now complete for the current core repository interfaces. Versioned migration tooling remains deferred.

## sqlc Repository Scaffold

Status: Completed initial generated package pass.

Acceptance criteria:

- sqlc configuration exists for the API service.
- Initial read/write queries exist for athlete profiles, training goals, plan weeks, and planned workouts.
- Generated Go code is produced if sqlc is locally available.
- No ConnectRPC, Flutter, AI, or real Garmin integration is added.

Notes:

- `services/api/sqlc.yaml` points at the initial core schema migration and query directory.
- Query files live in `services/api/internal/postgres/queries`.
- sqlc `1.31.1` is pinned in `.tool-versions`.
- Generated database code lives in `services/api/internal/postgres/db`.
- Postgres `numeric` fields are mapped to `float64` or `database/sql.NullFloat64` to match the domain distance model.
- Query files now cover the current core repository tables. Provider-specific Garmin tables and future read models remain deferred.

## AthleteProfile Postgres Repository

Status: Completed first sqlc-backed repository pass.

Acceptance criteria:

- A Postgres repository implements `repository.AthleteProfileRepository`.
- The implementation uses generated sqlc queries.
- Mapping between `domain.AthleteProfile` and generated database types is explicit and tested.
- Tests do not require a live Postgres database.
- No ConnectRPC, Flutter, AI, or real Garmin integration is added.

Notes:

- `services/api/internal/postgres/AthleteProfileRepository` wraps the generated sqlc `Querier`.
- `SaveAthleteProfile` validates the domain model and performs update-then-create behavior through sqlc.
- `GetAthleteProfile` maps `sql.ErrNoRows` to `repository.ErrNotFound`.
- Mapping tests cover UUID parsing, nullable strings, weekly distance, preferred run days, and constraints.
- Later passes added sqlc-backed repositories for training goals, plan weeks, planned workouts, imported activities, matches, results, and adaptation events.
- Live database integration tests remain deferred until there is a repeatable test database workflow.

## Core Plan Postgres Repositories

Status: Completed first sqlc-backed repository pass for training goals, plan weeks, and planned workouts.

Acceptance criteria:

- Postgres repositories implement `repository.TrainingGoalRepository`, `repository.PlanWeekRepository`, and `repository.PlannedWorkoutRepository`.
- Implementations use generated sqlc queries.
- Mapping between domain types and generated database types is explicit and tested.
- Tests do not require a live Postgres database.
- No ConnectRPC, Flutter, AI, or real Garmin integration is added.

Notes:

- `services/api/internal/postgres/core_repositories.go` contains sqlc-backed repositories for training goals, plan weeks, and planned workouts.
- `repository.PlannedWorkoutRepository` was added because planned workouts are stored as first-class rows in Postgres.
- The in-memory store now supports planned workouts directly as well as embedded workouts in `PlanWeek`.
- `domain.PlanWeek` now carries `AthleteID` and optional `GoalID`, matching the Postgres `plan_weeks` ownership columns.
- `GetPlanWeek` returns the week row only; loading planned workouts for the week remains deferred.
- Later passes added sqlc-backed repositories for imported activities, workout matches, workout results, and adaptation events.

## Core Completion Postgres Repositories

Status: Completed first sqlc-backed repository pass for imported activities, workout matches, workout results, and adaptation events.

Acceptance criteria:

- sqlc query files exist for imported activities, workout matches, workout results, adaptation events, and adaptation event changes.
- sqlc generated code is refreshed.
- Postgres repositories implement `repository.ImportedActivityRepository`, `repository.WorkoutMatchRepository`, `repository.WorkoutResultRepository`, and `repository.AdaptationEventRepository`.
- Mapping between domain types and generated database types is explicit and tested.
- Tests do not require a live Postgres database.
- No ConnectRPC, Flutter, AI, or real Garmin integration is added.

Notes:

- `services/api/internal/postgres/completion_repositories.go` contains sqlc-backed repositories for the activity completion side of the loop.
- `AdaptationEventRepository.SaveAdaptationEvent` saves the event and replaces its child changes in a single transaction.
- `PlanChange` IDs are generated in the repository layer because the current domain type does not carry a change ID.
- All current core repository interfaces now have sqlc-backed Postgres implementations.
- The app service is not wired to Postgres repositories yet.
- Live database integration tests remain deferred.

## Postgres Store Composition

Status: Completed initial composition pass.

Acceptance criteria:

- A small Postgres store can provide the current repository interfaces from a `*sql.DB`.
- The store composes athlete profile, training goal, plan week, planned workout, imported activity, workout match, workout result, and adaptation event repositories.
- The existing in-memory store remains intact.
- Production server startup is not wired to Postgres yet.
- Live database integration tests are not added yet.

Notes:

- `services/api/internal/postgres.Store` implements `repository.Store`.
- `postgres.NewStore` builds all current sqlc-backed repositories from a `*sql.DB`.
- Plan-week ownership is now carried by `domain.PlanWeek`, so the store does not need temporary plan-week owner options.
- Future app-service construction can choose either `repository.NewInMemoryStore()` or `postgres.NewStore(...)` behind the same interface.
- Server handler wiring and live database tests remain deferred.

## Server Configuration Preparation

Status: Completed minimal config pass.

Acceptance criteria:

- Server startup can read an HTTP bind address from environment.
- The default local server behavior remains unchanged.
- Optional database URL configuration can be detected without making Postgres mandatory.
- No database connection is opened unless a later stage explicitly wires it.
- No ConnectRPC, Flutter, AI, real Garmin integration, or live database integration tests are added.

Notes:

- `services/api/internal/config` reads `RUNTHREAD_SERVER_ADDR`, defaulting to `:8080`.
- `DATABASE_URL` is accepted as optional future Postgres configuration.
- `cmd/server` logs whether database configuration is present.
- Later startup composition adds repository store selection while keeping HTTP handlers minimal.

## Startup Storage Composition

Status: Completed minimal composition pass.

Acceptance criteria:

- Startup defaults to `repository.NewInMemoryStore()` when `DATABASE_URL` is empty.
- Startup opens a Postgres `database/sql` handle and composes `postgres.NewStore(db)` when `DATABASE_URL` is set.
- A cleanup function is returned so opened database handles can be closed.
- Migrations are not run automatically.
- No ConnectRPC, Flutter, AI, real Garmin integration, or live database integration tests are added.

Notes:

- `services/api/internal/startup` owns storage composition.
- `cmd/server` composes storage and logs the selected storage kind, but no HTTP/RPC handler uses app services yet.
- Unit coverage verifies the no-`DATABASE_URL` path without requiring a live database.
- The `DATABASE_URL` path still needs live database integration tests once a repeatable test database workflow exists.

## Application Service Composition

Status: Completed minimal composition pass.

Acceptance criteria:

- A small application service set can be built from a selected `repository.Store`.
- `CoreLoopService` is constructed with the selected store.
- Storage composition remains separate from application service composition.
- No new HTTP endpoints, ConnectRPC, Flutter, AI, real Garmin integration, or live database integration tests are added.

Notes:

- `services/api/internal/app.NewServices` returns a `Services` struct containing `CoreLoopService`.
- `cmd/server` constructs the service set after storage composition, but HTTP handlers still only expose `/healthz`.
- Future RPC handlers should depend on this application composition layer instead of constructing stores or domain services directly.

## Initial ConnectRPC API Contract

Status: Completed contract-only pass.

Acceptance criteria:

- Initial protobuf files exist under `services/api/proto`.
- The first service contract is provider-neutral.
- The contract includes a method for completing an imported activity through the backend core loop.
- No generated RPC code, handlers, Flutter, AI, real Garmin integration, or live database integration tests are added.

Notes:

- `services/api/proto/runthread/v1/runthread.proto` defines `RunthreadService.CompleteImportedActivity`.
- Request and response messages mirror the current core-loop application boundary.
- Dates are represented as ISO date strings where the domain treats them as calendar dates; activity and event instants use `google.protobuf.Timestamp`.
- Later passes add buf config and generated Go code before handler implementation.

## Protobuf Generation Tooling

Status: Completed initial generated-code pass.

Acceptance criteria:

- buf is pinned through asdf.
- buf configuration exists under `services/api`.
- Go protobuf and ConnectRPC generation is configured.
- Generated code stays isolated under `services/api/internal/rpc`.
- Required Go module dependencies for generated code are added.
- No ConnectRPC handlers, Flutter, AI, real Garmin integration, or live database integration tests are added.

Notes:

- `.tool-versions` pins buf `1.69.0`.
- `services/api/buf.yaml` defines the proto module and lint/breaking policy.
- `services/api/buf.gen.yaml` uses pinned remote buf plugins for Go protobuf `v1.36.11` and ConnectRPC `v1.18.1` code generation.
- Generated code lives under `services/api/internal/rpc/runthread/v1`.
- Runtime dependencies are `google.golang.org/protobuf v1.36.11` and `connectrpc.com/connect v1.18.1`.
- Generate code with `asdf exec buf generate` from `services/api`.
- Handler implementation remains deferred.

## First ConnectRPC Handler

Status: Completed thin handler pass.

Acceptance criteria:

- `RunthreadService.CompleteImportedActivity` is implemented as a thin ConnectRPC handler.
- The handler depends on `app.CoreLoopService`.
- Protobuf-to-domain mapping is explicit and focused.
- The handler is wired into `cmd/server`.
- `/healthz` remains unchanged.
- No auth, additional RPC methods, Flutter, AI, real Garmin integration, or live database integration tests are added.

Notes:

- `services/api/internal/rpc/handler` owns the transport adapter and mapper code.
- `cmd/server` mounts the generated Connect handler for `RunthreadService`.
- Unit tests cover mapper behavior and the handler path using in-memory services only.
- HTTP-level tests use `httptest` and the generated Connect client to verify the mounted `CompleteImportedActivity` endpoint without starting a real listener.
- Auth, request identity, and provider-backed activity import remain deferred.

## RPC Boundary Tightening

Status: Completed documentation pass.

Acceptance criteria:

- The temporary, demo-oriented parts of `CompleteImportedActivityRequest` are identified.
- The intended source of athlete, goal, and activity data is documented for the near-term production API.
- No new endpoints are added.
- The proto remains unchanged unless a change is clearly necessary.
- No Flutter, AI, real Garmin integration, auth, or live database integration tests are added.

Notes:

- The current RPC accepts full `AthleteProfile`, `TrainingGoal`, and `ImportedActivity` payloads to keep the first handler testable before auth and stored read flows exist.
- Athlete identity should eventually come from authenticated request context, not from request-supplied profile data.
- Training goals, plan weeks, and planned workouts should eventually be loaded from repositories by ID/current-athlete scope.
- Imported activities should usually be created by Garmin/provider import and referenced by ID when completing the core loop.
- `target_week_date` is also transitional; once plans are persisted, the backend should infer the relevant week from stored plan/workout state.
- Handler code now carries a TODO marking this replacement point.

## Pre-Stage-7 Readiness

Status: Completed API readiness review.

Ready for Flutter to call:

- The API server starts locally, keeps `/healthz` unchanged, and mounts the generated ConnectRPC `RunthreadService`.
- `CompleteImportedActivity` is callable with the generated Connect client and covered by an HTTP-level `httptest`.
- `GetCurrentPlanWeek` is callable with the generated Connect client and covered by an HTTP-level `httptest`.
- The handler depends on the application service composition layer and uses repository interfaces behind the core loop.
- Default local startup uses the in-memory repository store, so the Flutter MVP can exercise the backend loop without Postgres.
- Protobuf and ConnectRPC generation are pinned through buf and asdf.

Still demo-shaped:

- `CompleteImportedActivityRequest` carries full `AthleteProfile`, `TrainingGoal`, and `ImportedActivity` payloads.
- Athlete identity is request-supplied because auth/session context does not exist yet.
- The backend generates the target week during the request instead of loading a persisted current plan.
- Imported activity data is supplied inline instead of being created by Garmin/provider import and referenced by ID.
- The in-memory default loses data on restart, and the Postgres path still lacks live database integration tests.
- The first current-plan read RPC is implemented against in-memory storage only; Postgres-backed read queries remain deferred.
- There are no separate read RPCs yet for workout detail, imported activity status, or adaptation history.

Smallest backend changes before Flutter MVP screens:

- Decide whether the first Flutter screens should call the current demo-shaped completion RPC or wait for narrower persisted write endpoints.
- Treat the current in-memory `GetCurrentPlanWeek` RPC as the initial plan-screen data source.
- Keep auth, real Garmin import, and AI explanation generation deferred if the MVP uses seeded or mock local data.
- Add live database integration tests before relying on the `DATABASE_URL` Postgres path for app development.

## Read-Side API Contract

Status: Completed contract-only pass.

Acceptance criteria:

- A provider-neutral read RPC exists for the minimum plan-week state Flutter MVP screens need.
- The request shape is small and does not include Garmin-specific fields.
- Generated protobuf and ConnectRPC code are refreshed.
- Handler implementation remains deferred except for an explicit unimplemented stub required by the generated service interface.
- No Flutter, AI, real Garmin integration, auth, or live database integration tests are added.

Notes:

- `GetCurrentPlanWeek` accepts `plan_week_id`, `athlete_id`, optional `goal_id`, and `target_week_date`.
- The response includes `PlanWeek`, imported activities, workout matches, workout results, and adaptation events.
- The request-supplied athlete and goal identifiers are demo-shaped until auth/session context and current-plan repository lookup exist.
- A later pass adds the in-memory read-side app service and handler.

## Read-Side App Service and Handler

Status: Completed in-memory implementation pass.

Acceptance criteria:

- `GetCurrentPlanWeek` is implemented as a thin ConnectRPC handler.
- The handler maps protobuf to application types explicitly and calls an app service.
- The app service uses `repository.Store` interfaces and keeps training decisions in the deterministic planner.
- The in-memory store can return a saved current plan-week snapshot with nearby activity, match, result, and adaptation state.
- When no saved week exists, the app service can generate and save a deterministic demo week from stored athlete and goal records.
- No Flutter, AI, real Garmin integration, auth, Postgres integration tests, or new RPC methods are added.

Notes:

- `GetCurrentPlanWeekRequest.plan_week_id` is supported for direct saved-week lookup.
- `athlete_id`, `goal_id`, and `target_week_date` remain transitional until auth/session context and current-plan lookup exist.
- The Postgres store still compiles behind the base repository interfaces, but the richer current-plan snapshot query is implemented only by the in-memory store for now.
- Unit tests cover the app service, handler, and mapper paths with in-memory storage only.

## Stage 7: Flutter MVP Screens

Status: Completed MVP shell pass.

Acceptance criteria:

- Flutter app can show the current plan, workout details, imported activity status, and adaptation explanations.
- App talks to backend via ConnectRPC.
- UI supports the core loop without provider setup complexity.
- Basic loading and error states exist.

Notes:

- Flutter is scaffolded under `apps/mobile` and pinned through asdf.
- The first screen renders a seven-day weekly plan with workout details, loading state, and error state.
- Tapping a planned workout opens a detail view with scheduled date, type, status, target distance, target duration, intensity, and notes.
- Imported activity, match, and workout result data from `GetCurrentPlanWeek` are parsed by the mobile client and shown as completion state when present.
- Workouts without imported activity data show an empty imported-activity state.
- Adaptation events from `GetCurrentPlanWeek` are parsed by the mobile client and shown in a calm weekly plan summary, with an empty state when no adaptations exist.
- A lightweight bottom navigation now keeps the weekly plan as the default tab and adds a history tab.
- The history tab reuses `GetCurrentPlanWeek` data to show recent imported activity and adaptation history without adding backend endpoints.
- Workout detail includes a disabled completion affordance that explains completion will come from imported Garmin activity later; the Flutter app does not call the demo-shaped completion RPC.
- The Plan tab includes a small read-only Garmin connection surface with a disabled connect action, backed by `GetProviderConnectionStatus` when available and a safe not-connected fallback when status cannot be loaded.
- `lib/src/api/runthread_api.dart` isolates the backend API boundary.
- Dart protobuf/ConnectRPC generation remains deferred; the first mobile client uses Connect's JSON protocol directly.
- The app currently assumes local backend `http://localhost:8080` and sends demo `athlete-1` / `goal-1` identifiers.
- Local in-memory API startup seeds demo `athlete-1` / `goal-1` records so the Flutter MVP can verify backend communication without manual setup.
- `lib/src/demo` provides a local demo fallback when the backend is unavailable or a non-demo backend is unseeded, and the plan and history tabs show a demo-data notice.
- Widget tests cover the weekly plan, loading/error states, demo fallback notice, workout detail navigation, disabled completion affordance, imported activity completion state, adaptation summary state, and history navigation states.
- Garmin connection screens, auth, subscriptions, and AI explanations remain deferred.

Readiness cleanup notes:

- The Stage 7 mobile surface is ready as an MVP shell for plan viewing, workout detail review, activity/adaptation read state, and demo fallback.
- It is not beta-ready yet because athlete identity, current-plan lookup, and fallback data are demo-shaped.
- The Flutter client still uses hand-written Connect JSON instead of generated Dart protobuf/ConnectRPC code.
- Mobile completion remains read-only; real completion should come from imported Garmin activities matched by the backend.
- Postgres-backed current-plan reads, auth, provider connection, and production refresh behavior remain deferred.

Stage 7 handoff:

- Acceptance criteria are met for the current MVP shell: the app shows the current plan, workout details, imported activity status when present, adaptation summaries, loading state, error state, and demo fallback.
- The weekly plan remains the default tab; history is a read-only second tab backed by the same `GetCurrentPlanWeek` response.
- The app does not call `CompleteImportedActivity` because that RPC still accepts request-supplied athlete, goal, and activity payloads for backend-loop testing.
- Remaining Stage 7 gaps are intentionally deferred: generated Dart RPC bindings, auth-backed current athlete lookup, persisted current-plan reads, live imported activity data, and production write flows.

## Stage 8: Provider-Neutral Activity Import Foundation

Acceptance criteria:

- Provider integration docs describe a provider-neutral activity import pipeline.
- Backend boundaries make clear that provider integrations, OAuth, token refresh, imports, and webhooks are backend-owned.
- Core domain logic uses `ImportedActivity` instead of Strava, Garmin, COROS, or Apple Health payloads.
- Existing provider persistence and import orchestration remain provider-oriented rather than Garmin-only.
- No real Strava or Garmin API calls are added.

Notes:

- Stage 8 supersedes the previous Garmin-first integration plan.
- Strava becomes the first likely MVP provider for validating the import loop.
- Garmin direct integration remains on the roadmap, but it should use the same import pipeline.

## Stage 9: Mock Strava Provider

Status: Completed initial mock normalisation pass.

Acceptance criteria:

- A mock Strava payload shape exists in a Strava provider package.
- Mock Strava activities normalise into `domain.ImportedActivity`.
- Tests cover representative road run, trail run, treadmill, ignored/unknown, and invalid payload cases.
- Provider-specific Strava fields stay inside Strava package tests and adapters.
- No Strava OAuth, token storage, webhooks, or real Strava API calls are added.

Notes:

- `services/api/internal/providers/strava` contains a mock Strava activity payload type, mock provider adapter, and normalisation logic.
- Mock Strava road runs, trail runs, and virtual runs map to provider-neutral `domain.ImportedActivity` values.
- Non-run Strava activity types return an unsupported activity error so a later import orchestration stage can mark them ignored.
- Strava-specific activity IDs, sport type labels, raw JSON payloads, and provider parsing remain inside the Strava provider package.
- Real Strava OAuth, token storage, backfill, webhooks, API calls, Flutter UI, and AI integration remain out of scope.

## Stage 10: Strava OAuth and Token Storage

Status: Completed backend-only skeleton pass.

Acceptance criteria:

- Backend can start a Strava OAuth flow and generate/validate state.
- Backend callback exchanges the authorization code for tokens.
- Token storage uses secure token references or encrypted storage outside core domain models.
- Provider connection status reflects pending, connected, error, and disconnected states.
- Flutter may open the backend-provided authorization URL, but does not store tokens or call Strava directly.

Notes:

- `services/api/internal/providers/strava` contains a backend-only OAuth skeleton for starting a Strava connect flow and completing a callback using test-supplied token data.
- `OAuthService.StartOAuth` creates or reuses a pending Strava provider connection, generates state, stores that state on provider connection metadata, and returns a Strava authorization URL.
- `OAuthService.CompleteOAuthCallback` validates state and code, stores token data through a `TokenStore` interface, then marks the provider connection connected with a token reference and provider user ID.
- Token values remain inside the Strava provider package and token-store boundary. Core domain models only see provider-neutral connection metadata and token references.
- Automated tests use fake token storage only. There are still no real Strava API calls, no real code exchange, no webhooks, no activity backfill, no Flutter UI, and no AI integration.

## Stage 11: Strava Activity Backfill/Import

Status: Completed backend-only skeleton pass.

Acceptance criteria:

- Backend can enqueue or run an initial Strava backfill/import job for a connected athlete.
- Import jobs fetch Strava activity details through a rate-limit-aware boundary.
- Imported Strava activities are idempotently recorded and normalised into `ImportedActivity`.
- Backfill status and errors are persisted for retry/support.
- Tests use mocks or fixtures only; no real Strava API calls run in automated tests.

Notes:

- `services/api/internal/providers/strava.BackfillService` can run an initial backfill for a connected Strava provider connection.
- Backfill uses an `ActivityFetcher` interface for listing activity summaries and fetching activity details; tests use fakes only.
- Fetched mock Strava details are normalised and persisted through the existing provider-neutral `providerimport.Service`.
- Re-importing the same Strava activity is idempotent through deterministic provider activity and imported activity IDs.
- Unsupported non-run activities are recorded as ignored provider imports.
- Strava rate-limit errors return a deferred backfill result and persist retry/support context on the provider connection.
- There are still no real Strava API calls, no webhooks, no Flutter UI, and no AI integration.

## Stage 12: Strava Webhook Handling

Status: Completed backend-only skeleton pass.

Acceptance criteria:

- Backend exposes provider-facing Strava webhook endpoints.
- Webhook verification and event deduplication are implemented according to Strava requirements.
- Activity create/update/delete events become import jobs or provider activity state changes.
- Webhook retries and failures are observable and recoverable.
- Webhook payloads do not flow directly into planning or adaptation logic.

Notes:

- `services/api/internal/providers/strava.WebhookService` parses mock Strava webhook payloads and routes activity create, update, and delete events.
- Verification and deduplication are represented by testable interfaces; tests use fakes only.
- Create and update events fetch mock activity detail through the existing `ActivityFetcher` boundary and persist through the provider-neutral `providerimport.Service`.
- Delete events record ignored provider import state without passing provider payloads to core domain logic.
- Failed fetches are recorded as failed provider import events for retry/support visibility.
- There are still no public production webhook endpoints, real Strava API calls, Flutter UI, or AI integration.

## Stage 13: Activity Matching to Planned Workouts

Status: Completed provider-neutral matching pass.

Acceptance criteria:

- Imported activities from Strava use the same matching flow as existing provider-neutral imports.
- The backend can match imported activity to the likely planned workout for the current athlete.
- Confident, uncertain, rejected, and manual match paths remain represented in `WorkoutMatch`.
- Flutter reads match state through provider-neutral plan/read APIs.
- Provider-specific matching rules are avoided unless first mapped to provider-neutral signals.

Notes:

- `services/api/internal/app.ProviderActivityMatchService` matches already-imported provider activities to planned workouts using `domain.ImportedActivity` and the existing `matching.Matcher`.
- The service persists `WorkoutMatch` records for confident, uncertain, rejected, and manual matches so manual review can build on stored match state later.
- A Strava package test proves a mock Strava backfill import can be matched to a planned workout through the provider-neutral app service.
- No Strava-specific matching rules were added; Strava payload fields remain inside `services/api/internal/providers/strava`.
- There are still no Flutter UI changes, real Strava API calls, or AI integration.

## Stage 14: Adaptation from Imported Activities

Status: Completed provider-neutral result/adaptation pass.

Acceptance criteria:

- Matched imported activities produce `WorkoutResult` records.
- Deterministic adaptation rules consume workout results from imported activities.
- Adaptation events explain what changed and why without relying on provider-specific payloads.
- Strava-derived activity data is not sent to AI prompts or used for AI model training.
- End-to-end tests cover import, match, result, and adaptation using mocked provider data.

Notes:

- `services/api/internal/app.ProviderActivityCompletionService` consumes matched provider-neutral activities and creates `WorkoutResult` records through existing workout completion helpers.
- The service runs the deterministic adaptation engine for outcomes that require adaptation and persists workout results, updated workouts/weeks, and adaptation events.
- Uncertain and rejected matches do not create workout results; they remain available for later manual review.
- A Strava package test proves mock Strava backfill import, provider-neutral matching, workout result creation, and deterministic adaptation work together without Strava-specific adaptation rules.
- Strava-derived activity data is still not sent to AI prompts or used for model training.
- There are still no Flutter UI changes, real Strava API calls, or AI integration.

## Stage 15: Garmin Direct Integration

Status: Preparation aligned with provider-neutral pipeline; real Garmin remains blocked.

Acceptance criteria:

- Garmin access, approval, data delivery, rate limits, storage, and revocation requirements are validated.
- Garmin uses the same provider-neutral connection, import, normalisation, matching, and adaptation pipeline as Strava.
- Garmin-specific fields remain inside Garmin provider packages and persistence boundaries.
- Users can connect Garmin through the supported Garmin flow when available.
- Garmin integration failures are logged, retryable where appropriate, and visible through provider-neutral status.

Notes:

- Garmin direct integration remains blocked by `docs/garmin-access-findings.md` and ADR-0007 until external findings are validated.
- Existing mock Garmin import boundaries already prove provider import, matching, workout result creation, and adaptation can run without Garmin data leaking into core packages.
- `services/api/internal/providers/garmin` now contains a mock `ActivityProvider` bridge that delegates to the existing mock Garmin normaliser.
- The bridge clarifies the eventual destination for Garmin provider-specific code without moving the legacy mock package or implementing real provider behavior.
- No real Garmin OAuth, callbacks, polling, webhooks, token refresh, API calls, Flutter connect enablement, or AI integration were added.

## Stage 16: Subscriptions and Private Beta

Status: Preparation checklist documented; implementation not started.

Acceptance criteria:

- Subscription flow is implemented for the chosen platform and backend model.
- Private beta users can onboard, connect the supported MVP provider, view a plan, and receive adaptations.
- Operational monitoring, support paths, privacy review, deletion/export behavior, and provider terms review exist.
- Strava and Garmin data access rules are documented before beta launch.
- The beta path does not depend on AI-generated training decisions.

Notes:

- Stage 16 is currently a readiness planning stage, not a payment/auth/provider implementation stage.
- See `docs/private-beta.md` for the private beta checklist, blockers, and smallest next implementation order.
- The beta provider path remains Strava-first for MVP validation. Garmin direct integration remains blocked by `docs/garmin-access-findings.md` and ADR-0007.
- Before private beta, Runthread still needs production auth/current athlete identity, live Strava OAuth and token storage, provider import operations, deletion/export and retention behavior, monitoring/support readiness, and a subscription or entitlement decision.
- No real subscriptions, payment provider calls, production auth, live Strava or Garmin API calls, Flutter UI changes, or AI integration have been added for this preparation pass.
