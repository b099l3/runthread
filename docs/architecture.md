# Architecture

Runthread uses a Flutter mobile app, a Go backend, ConnectRPC APIs, and Postgres.

The backend owns domain logic and provider integrations. The mobile app presents the plan, workout state, imported activities, adaptations, and explanations. External provider integrations are isolated so provider-specific details do not leak into the core training model.

## Intended Architecture

```text
          +----------------------+
          |   Flutter Mobile App |
          +----------+-----------+
                     |
                     | ConnectRPC
                     v
          +----------------------+
          |      Go API          |
          |----------------------|
          | RPC handlers         |
          | Domain logic         |
          | Planning engine      |
          | Adaptation engine    |
          | Provider integrations|
          +----------+-----------+
                     |
                     | sqlc queries
                     v
          +----------------------+
          |      Postgres        |
          +----------------------+

          +----------------------+
          | Provider APIs        |
          | Strava / Garmin     |
          +----------+-----------+
                     |
                     | Imports activities
                     v
          +----------------------+
          | Provider Integration |
          +----------------------+
```

Provider-neutral import flow:

```text
Flutter App
  -> Connect Strava / Garmin
  -> Go API
  -> Provider Integration Layer
  -> Normalised ImportedActivity
  -> Workout Matching
  -> Adaptation Engine
```

## Components

The Flutter app talks to the Go backend through ConnectRPC. It should not duplicate training decisions locally.

The initial Flutter MVP lives in `apps/mobile`. It starts on the weekly plan view, calls `GetCurrentPlanWeek`, and keeps transport code isolated behind `lib/src/api/runthread_api.dart`. Dart protobuf/ConnectRPC generation is deferred, so the first mobile client uses Connect's JSON protocol directly as a temporary boundary.

The Go backend owns the domain model, planning rules, workout matching, adaptation rules, and integration workflows.

Provider integrations are represented behind a provider-neutral `ActivityProvider` concept. A provider adapter knows how to validate provider-specific payloads and normalise them into `domain.ImportedActivity`. Strava, Garmin, COROS, Apple Health, or any later provider should feed the same internal import pipeline once connected.

The backend owns OAuth and import execution. The mobile app may start a connect flow and open the provider authorization URL, but the backend exchanges authorization codes, stores token references securely, refreshes tokens, imports activities, handles provider webhooks, and records import state. Mobile should not receive refresh tokens or call provider APIs directly.

The initial API contract lives in `services/api/proto/runthread/v1/runthread.proto`. It defines a provider-neutral `RunthreadService` with `CompleteImportedActivity` for the current backend core-loop application service boundary and `GetCurrentPlanWeek` as the first read-side contract for the Flutter MVP. Buf generation config lives in `services/api/buf.yaml` and `services/api/buf.gen.yaml`, and generated Go protobuf/ConnectRPC code lives under `services/api/internal/rpc`.

`services/api/internal/rpc/handler` contains the first thin ConnectRPC handler. It maps protobuf messages to domain/app types explicitly, calls application services, and maps responses back to protobuf. It does not own training decisions, persistence selection, auth, or provider-specific logic.

The current `CompleteImportedActivity` request is deliberately wider than the eventual production API. It accepts `AthleteProfile`, `TrainingGoal`, and `ImportedActivity` payloads so the first end-to-end RPC can exercise the core loop before auth, read models, and real provider import exist. Near term, this should narrow: athlete identity should come from auth/session context, goals and planned workouts should be loaded from repositories, and imported activities should usually be referenced by ID after provider import rather than supplied inline by the client. Keeping the temporary shape explicit prevents the Flutter app from treating this demo-oriented request as the long-term contract.

The current `GetCurrentPlanWeek` request is also transitional. It accepts request-supplied `plan_week_id`, `athlete_id`, optional `goal_id`, and `target_week_date` so the Flutter MVP has a small provider-neutral read shape before auth and current-plan lookup exist. The response groups the current plan week with related imported activities, matches, workout results, and adaptation events that the first MVP screens are expected to render.

The mobile app currently sends demo `athlete-1` and `goal-1` identifiers to the local backend. Local in-memory API startup seeds those records so the Flutter MVP can verify server/app communication without manual setup. If those records are missing in a non-demo backend or the backend is unavailable, the app falls back to local demo plan data from `lib/src/demo` and shows a visible demo-data notice. Onboarding, auth-backed athlete selection, and seeded beta data remain future work.

Application service boundaries sit between future RPC handlers and the domain packages. These services should expose use-case shaped methods and keep handlers thin. The first boundary is `services/api/internal/app/CoreLoopService`, which wraps the test-only core-loop harness and persists successful outputs through repository interfaces without adding ConnectRPC.

`services/api/internal/app.NewServices` is the application composition layer. It builds backend application services from a selected `repository.Store`, starting with `CoreLoopService` and `CurrentPlanWeekService`. HTTP and ConnectRPC handlers should depend on this composed service set rather than constructing domain services or repositories directly.

Repository interfaces define the persistence boundary. The initial in-memory store lives in `services/api/internal/repository` and is intended for tests and early service wiring. sqlc-backed repositories live in `services/api/internal/postgres` and now cover the current core repository interfaces. `postgres.Store` composes those repositories behind `repository.Store` from a `*sql.DB`.

`CoreLoopService` persists successful in-memory core-loop outputs through repository interfaces, including both the saved `PlanWeek` and each `PlannedWorkout` as a first-class record. This keeps future RPC handlers focused on transport concerns and keeps the eventual Postgres/sqlc implementation behind the same boundary.

`CurrentPlanWeekService` reads current plan-week state through repository interfaces. With the in-memory store, it can return a saved week and nearby completion state, or generate and save a deterministic demo week from stored athlete and goal records when no saved week exists. With the Postgres store, it reads persisted current-plan snapshots, including planned workouts, recent imported activities, workout matches, workout results, and adaptation events. Live database integration tests are still deferred.

Server startup reads configuration from the environment through `services/api/internal/config`. `RUNTHREAD_SERVER_ADDR` controls the HTTP bind address and defaults to `:8080`. `services/api/internal/startup` composes storage: it uses `repository.NewInMemoryStore()` when `DATABASE_URL` is empty, and opens a Postgres `database/sql` handle for `postgres.NewStore(db)` when `DATABASE_URL` is set. The server constructs application services from the selected store, exposes `/healthz`, and mounts the first ConnectRPC handler.

Postgres will store app data, including athletes, goals, plans, planned workouts, imported activities, workout matches, workout results, provider connections, provider import jobs/events, webhook events, token references, and adaptation events.

The initial Postgres persistence model is planned in `docs/persistence.md`. Plain SQL migrations exist under `services/api/internal/postgres/migrations`, and `scripts/db` contains minimal local up/down wrappers. sqlc config, query files, and generated database code exist for the current core and provider tables. sqlc-backed repositories and a Postgres store composition layer exist for athlete profiles, training goals, plan weeks, planned workouts, imported activities, workout matches, workout results, provider persistence, and adaptation events. Versioned migration tooling, live database integration tests, auth, and additional implemented ConnectRPC methods remain deferred.

sqlc will generate typed Go data access code from SQL queries when the persistence layer is introduced.

Provider integrations import activities and normalise them into Runthread's `ImportedActivity` model before the core domain evaluates them. Strava is the first likely real MVP provider; Garmin direct integration remains a later premium/direct integration target. Provider-specific identifiers, payloads, scopes, webhook signatures, rate-limit handling, and token details should stay in provider packages and persistence boundaries.

AI can be used to draft explanations or copy. It must not decide workouts, plan changes, training load, or adaptation rules.

## Future RPC Candidates

The first ConnectRPC services should be thin wrappers around application services, not domain packages directly.

Likely early RPC methods:

- Complete an imported activity through the core loop. This is the first proto contract.
- Fetch the current plan week and nearby activity/completion/adaptation state. This is the first read-side proto contract.
- Generate a plan week from an athlete profile, goal, and target week date if a separate planning action is needed.
- Mark a planned workout missed, skipped, moved, or manually completed.
- Match an imported activity to a planned workout.
- Fetch adaptation events and user-facing explanation data once persistence exists.

Database persistence, auth, provider connections, and real provider import should be added before these become production API endpoints.
