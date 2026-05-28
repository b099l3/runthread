# Persistence Model

This document explains the initial Postgres persistence model. Plain SQL migrations live in `services/api/internal/postgres/migrations`, and small local scripts in `scripts/db` can apply or roll them back against Docker Compose Postgres. sqlc query files, generated database code, sqlc-backed repositories, and Postgres store composition layers now exist for the current core and provider repository interfaces. The app service still uses core repository interfaces, and provider storage has not been wired into app services or production server startup.

The first database model should store Runthread's provider-neutral domain concepts. Provider-specific Garmin identifiers, OAuth tokens, raw payloads, webhook metadata, and provider connection state should stay out of these core tables until the real Garmin integration stage.

## Principles

- Use `uuid` primary keys for app-owned records.
- Store domain enums as text initially, with application validation owning allowed values.
- Keep provider-neutral activity data in `imported_activities`.
- Add provider-specific tables later for Garmin connections, provider activity IDs, raw payload audit data, and import delivery state.
- Prefer simple foreign keys over denormalised JSON for core relationships.
- Use timestamps for auditability: `created_at` and `updated_at` where records may change.

## Relationships

```text
athlete_profiles
  -> training_goals
  -> plan_weeks
      -> planned_workouts
  -> imported_activities

planned_workouts
  -> workout_matches <- imported_activities
  -> workout_results

adaptation_events
  -> adaptation_event_changes
```

## athlete_profiles

Stores domain type: `AthleteProfile`.

Primary key:

- `id uuid primary key`

Important columns:

- `display_name text`
- `experience_level text`
- `current_weekly_distance_meters numeric`
- `preferred_run_days smallint[]`
- `constraints text[]`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

Notes:

- `preferred_run_days` stores Go `time.Weekday` values using Postgres integers.
- Provider connections do not belong here.

## training_goals

Stores domain type: `TrainingGoal`.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `athlete_id uuid not null references athlete_profiles(id)`

Important columns:

- `type text not null`
- `target_date date`
- `target_distance_meters numeric`
- `target_duration_seconds integer`
- `notes text`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

Notes:

- A runner may have multiple goals over time.
- The active-goal concept can be added later if needed.

## plan_weeks

Stores domain type: `PlanWeek`.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `athlete_id uuid not null references athlete_profiles(id)`
- `goal_id uuid references training_goals(id)`

Important columns:

- `plan_id uuid not null`
- `week_index integer not null`
- `starts_on date not null`
- `focus text`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

Notes:

- The current domain has `TrainingPlan`, but repository work currently stores `PlanWeek` directly. `plan_id` remains as an app-owned grouping key.
- `domain.PlanWeek` carries `AthleteID` and optional `GoalID` so repository implementations do not need external owner context to save a week.
- A separate `training_plans` table can be introduced when full multi-week plan lifecycle is needed.
- Add a uniqueness constraint later on `(plan_id, week_index)`.

## planned_workouts

Stores domain type: `PlannedWorkout`.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `plan_week_id uuid not null references plan_weeks(id)`

Important columns:

- `plan_id uuid not null`
- `scheduled_for date not null`
- `type text not null`
- `status text not null`
- `target_distance_meters numeric`
- `target_duration_seconds integer`
- `intensity_kind text`
- `intensity_description text`
- `notes text`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

Notes:

- Workouts are stored as rows instead of embedding them inside `plan_weeks`.
- Rest and strength days may have no distance target.
- Run workouts should normally have either distance or duration.

## imported_activities

Stores domain type: `ImportedActivity`.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `athlete_id uuid not null references athlete_profiles(id)`

Important columns:

- `type text not null`
- `started_at timestamptz not null`
- `duration_seconds integer not null`
- `distance_meters numeric`
- `average_pace_seconds_per_kilometer integer`
- `average_heart_bpm integer`
- `created_at timestamptz not null`

Notes:

- This table is provider-neutral.
- Do not store Garmin activity IDs or Garmin activity type strings here.
- Later, real provider imports should have a separate provider activity table that maps provider IDs to `imported_activities.id`.

## workout_matches

Stores domain type: `WorkoutMatch`.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `planned_workout_id uuid not null references planned_workouts(id)`
- `imported_activity_id uuid not null references imported_activities(id)`

Important columns:

- `status text not null`
- `confidence text not null`
- `matched_by text not null`
- `matched_at timestamptz not null`
- `notes text`
- `created_at timestamptz not null`

Notes:

- `matched_by` represents automatic vs manual matching.
- Add a uniqueness rule later if the product should prevent multiple active matches for one workout or activity.

## workout_results

Stores domain type: `WorkoutResult`.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `planned_workout_id uuid not null references planned_workouts(id)`
- `imported_activity_id uuid references imported_activities(id)`

Important columns:

- `outcome text not null`
- `completed_at timestamptz`
- `distance_meters numeric`
- `duration_seconds integer`
- `notes text`
- `created_at timestamptz not null`

Notes:

- Missed and skipped results do not need an imported activity.
- Completed, partial, overperformed, and underperformed results should reference an imported activity.

## adaptation_events

Stores domain type: `AdaptationEvent`.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `plan_id uuid not null`
- `athlete_id uuid not null references athlete_profiles(id)`

Important columns:

- `type text not null`
- `reason text not null`
- `summary text not null`
- `created_at timestamptz not null`

Notes:

- `plan_id` is not yet backed by a `training_plans` table.
- This table records deterministic adaptation decisions.
- AI-generated explanation text should be stored separately later if needed.

## adaptation_event_changes

Stores domain type: `PlanChange`.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `adaptation_event_id uuid not null references adaptation_events(id) on delete cascade`
- `planned_workout_id uuid references planned_workouts(id)`

Important columns:

- `type text not null`
- `description text not null`
- `position integer not null`

Notes:

- `position` preserves change ordering.
- `planned_workout_id` is nullable for plan-level notes.

## Provider-Specific Data

Keep these out of the core tables:

- Garmin OAuth tokens.
- Garmin user/account identifiers.
- Garmin activity IDs.
- Garmin raw activity payloads.
- Garmin webhook delivery IDs.
- Garmin import retry state.
- Provider-specific activity type strings.

Stage 8 should add provider-specific tables outside the core domain model:

- `provider_connections`
- `provider_activities`
- `provider_activity_payloads`
- `provider_import_events`, if webhook/polling delivery audit is needed immediately
- token encryption and rotation strategy
- Garmin API access validation and terms review

## provider_connections

Stores integration concept: athlete-owned provider account connection.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `athlete_id uuid not null references athlete_profiles(id)`

Important columns:

- `provider text not null`
- `provider_user_id text`
- `status text not null`
- `connected_at timestamptz`
- `disconnected_at timestamptz`
- `last_sync_at timestamptz`
- `last_import_cursor text`
- `token_reference text`
- `token_expires_at timestamptz`
- `last_error text`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

Notes:

- `provider` is initially `garmin`, but this table should also work for Coros and Apple Watch later.
- `provider_user_id` is the external account identifier if Garmin provides one.
- `token_reference` should not store plaintext tokens. It should point to encrypted token material or a secret store if token storage is allowed by provider terms.
- `last_import_cursor` is optional and depends on Garmin's supported sync model.
- Suggested statuses: `pending`, `connected`, `syncing`, `error`, `disconnected`.
- A later migration should add an index on `(athlete_id, provider)` and a uniqueness rule for active connections once reconnect behavior is decided.

## provider_activities

Stores integration concept: provider-specific activity record mapped to a provider-neutral `ImportedActivity`.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `provider_connection_id uuid not null references provider_connections(id)`
- `athlete_id uuid not null references athlete_profiles(id)`
- `imported_activity_id uuid references imported_activities(id)`

Important columns:

- `provider text not null`
- `provider_activity_id text not null`
- `provider_activity_type text`
- `started_at timestamptz`
- `status text not null`
- `first_seen_at timestamptz not null`
- `last_synced_at timestamptz`
- `last_error text`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

Notes:

- This table is where Garmin activity IDs and Garmin activity type strings belong.
- `imported_activity_id` is nullable until normalisation succeeds or when an activity is intentionally ignored.
- Suggested statuses: `received`, `normalised`, `ignored`, `failed`.
- Add a uniqueness constraint on `(provider_connection_id, provider_activity_id)` for idempotency.
- Add indexes on `(athlete_id, started_at)` and `imported_activity_id`.

## provider_activity_payloads

Stores integration concept: optional raw provider payload snapshot for audit/debugging.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `provider_activity_id uuid not null references provider_activities(id) on delete cascade`

Important columns:

- `payload jsonb not null`
- `payload_kind text not null`
- `received_at timestamptz not null`

Notes:

- Store raw Garmin payloads only if allowed by Garmin terms and Runthread privacy policy.
- This table should be optional for early development if provider terms or privacy review are not complete.
- Avoid reading raw payloads from core domain logic. Normalisation should happen in the Garmin integration package.

## provider_import_events

Stores integration concept: optional webhook/polling delivery audit.

Primary key:

- `id uuid primary key`

Important foreign keys:

- `provider_connection_id uuid references provider_connections(id)`
- `provider_activity_id uuid references provider_activities(id)`

Important columns:

- `provider text not null`
- `event_type text not null`
- `delivery_id text`
- `status text not null`
- `received_at timestamptz not null`
- `processed_at timestamptz`
- `error text`

Notes:

- Add this table if Garmin uses webhook deliveries, or if polling/sync attempts need auditability from the start.
- `delivery_id` should hold provider webhook delivery identifiers when available.
- Suggested statuses: `received`, `processed`, `ignored`, `failed`.
- This is not a core training table.

## Provider Migration

The provider-specific migration creates provider tables without changing core domain tables:

- `000002_provider_connections.up.sql`
- `000002_provider_connections.down.sql`

Contents:

- create `provider_connections`
- create `provider_activities`
- create `provider_activity_payloads`
- create `provider_import_events`
- add uniqueness on `(provider_connection_id, provider_activity_id)`
- add indexes for athlete/provider lookup, activity import idempotency, and import status scans
- add a unique delivery index for provider import events when a provider delivery ID exists

Do not add Garmin-specific columns to `imported_activities`.

## Provider sqlc Scaffold

Initial sqlc query files exist for provider persistence tables:

- `provider_connections.sql`
- `provider_activities.sql`
- `provider_activity_payloads.sql`
- `provider_import_events.sql`

The queries intentionally cover small read/write operations needed by current repositories: create, update where useful, get by ID, and list by athlete/status/connection. sqlc-backed provider repositories and Postgres store composition now exist in `services/api/internal/postgres`, with tests that do not require a live database. Server startup composes the Postgres store when `DATABASE_URL` is set, and Strava provider services use those provider repositories when Strava is configured.

Provider persistence models and repository interfaces live in `services/api/internal/repository`, not in the core domain package. The in-memory store implements those interfaces for tests and early integration design. This keeps provider connection state outside `domain.ImportedActivity` and the core training logic.

Type mapping notes:

- `provider_activity_payloads.payload jsonb` generates `encoding/json.RawMessage`.
- Nullable provider timestamps and optional text fields use `database/sql` nullable types.
- Nullable provider import event foreign keys use `uuid.NullUUID`; connection-scoped list queries use a non-null UUID argument.

## Migration Notes For Later

Initial plain SQL migrations:

- `services/api/internal/postgres/migrations/000001_initial_core_schema.up.sql`
- `services/api/internal/postgres/migrations/000001_initial_core_schema.down.sql`

Local migration scripts:

- `scripts/db/migrate-up.sh`
- `scripts/db/migrate-down.sh`

Run locally:

```sh
docker compose -f infra/docker-compose.yml up -d postgres
./scripts/db/migrate-up.sh
```

Roll back locally:

```sh
./scripts/db/migrate-down.sh
```

These scripts do not track migration versions. They are a minimal local workflow around the first SQL migration only.

## sqlc Scaffold

Initial sqlc files:

- `services/api/sqlc.yaml`
- `services/api/internal/postgres/queries/athlete_profiles.sql`
- `services/api/internal/postgres/queries/training_goals.sql`
- `services/api/internal/postgres/queries/plan_weeks.sql`
- `services/api/internal/postgres/queries/planned_workouts.sql`
- `services/api/internal/postgres/queries/imported_activities.sql`
- `services/api/internal/postgres/queries/workout_matches.sql`
- `services/api/internal/postgres/queries/workout_results.sql`
- `services/api/internal/postgres/queries/adaptation_events.sql`
- `services/api/internal/postgres/queries/adaptation_event_changes.sql`
- `services/api/internal/postgres/db`

The scaffold covers the current repository core:

- `athlete_profiles`
- `training_goals`
- `plan_weeks`
- `planned_workouts`
- `imported_activities`
- `workout_matches`
- `workout_results`
- `adaptation_events`
- `adaptation_event_changes`

The query files include basic create, get, list, update, and delete operations where useful.

sqlc `1.31.1` is pinned in `.tool-versions`. Generate Go code with:

```sh
cd services/api
sqlc generate
```

If your shell is not using asdf shims, run `asdf exec sqlc generate` from `services/api`.

The generated package is configured to live at `services/api/internal/postgres/db`. That directory should remain generated-only.

The first hand-written Postgres repositories live in:

- `services/api/internal/postgres/athlete_profile_repository.go`
- `services/api/internal/postgres/core_repositories.go`
- `services/api/internal/postgres/completion_repositories.go`

They map between domain types and generated sqlc types while satisfying repository interfaces for athlete profiles, training goals, plan weeks, planned workouts, imported activities, workout matches, workout results, and adaptation events.

`services/api/internal/postgres/store.go` composes those repositories into a `repository.Store` from a `*sql.DB`. App-service construction can choose either `repository.NewInMemoryStore()` or `postgres.NewStore(...)` without changing application service methods.

`CoreLoopService.CompleteImportedActivity` now saves both the generated `PlanWeek` and each `PlannedWorkout` through the repository boundary. This matches the Postgres shape where `plan_weeks` and `planned_workouts` are separate tables.

Server startup can read an optional `DATABASE_URL`. `services/api/internal/startup` chooses `repository.NewInMemoryStore()` when it is empty and opens a Postgres `database/sql` handle for `postgres.NewStore(db)` when it is set. Migrations are not run automatically; the configured database is expected to have the schema applied before app services use the Postgres store.

`services/api/internal/app.NewServices` builds application services from the selected store. Future RPC handlers should receive those composed services rather than selecting repositories themselves.

Current repository shape notes:

- `PlanWeek` rows require `athlete_id` and optional `goal_id`; `domain.PlanWeek` now carries those fields directly.
- `postgres.Store` no longer needs plan-week owner options. It can compose repositories from only a `*sql.DB`.
- The `DATABASE_URL` startup path still needs live database integration tests later.
- `PlanWeek` and `PlannedWorkout` are stored separately in Postgres. `GetPlanWeek` returns the week row only; the current-plan snapshot reader handles the richer mobile read path by loading workouts and nearby completion/adaptation state explicitly.
- `AdaptationEventRepository.SaveAdaptationEvent` writes the event and replaces its child changes in one transaction.
- `PlanChange` does not currently carry its own ID, so Postgres adaptation-event-change IDs are generated in the repository layer.

Type mapping decisions:

- Postgres `uuid` columns generate as `github.com/google/uuid.UUID`.
- Nullable Postgres `uuid` columns generate as `uuid.NullUUID`.
- Postgres `numeric` columns generate as `float64` or `database/sql.NullFloat64` because the current numeric fields store meter values already represented as floats in the domain model.
- Postgres arrays generate through `github.com/lib/pq` because the first generated package uses `database/sql`.

## Repository Test Approach

Current Postgres repository tests cover domain-to-sqlc mapping without requiring a live database.

Live database integration tests are deferred until there is a repeatable test database workflow. Those tests should start Docker Compose Postgres, apply migrations, exercise sqlc repositories against real tables, and clean up test data deterministically.

When migration tooling is introduced:

- Choose a versioned migration tool and replace or wrap the current scripts.
- Decide whether IDs are generated by the app or by Postgres.
- Decide whether enum checks belong in database constraints or only in Go validation.
- Add indexes around common reads: athlete goals, plan weeks by plan, workouts by week/date, activities by athlete/start time, matches by workout/activity, adaptation events by athlete/plan.
