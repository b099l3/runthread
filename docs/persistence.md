# Persistence Model

This document explains the initial Postgres persistence model. Plain SQL migrations live in `services/api/internal/postgres/migrations`, and small local scripts in `scripts/db` can apply or roll them back against Docker Compose Postgres. sqlc query files, generated database code, sqlc-backed repositories, and a Postgres store composition layer now exist for the current core repository interfaces. The app service still uses repository interfaces and production server startup has not been wired to Postgres.

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

Keep these out of the core tables for now:

- Garmin OAuth tokens.
- Garmin user/account identifiers.
- Garmin activity IDs.
- Garmin raw activity payloads.
- Garmin webhook delivery IDs.
- Garmin import retry state.
- Provider-specific activity type strings.

These should wait until real Garmin integration:

- `provider_connections`
- `provider_activities`
- `provider_activity_payloads`
- webhook delivery/audit tables
- token encryption and rotation strategy
- Garmin API access validation and terms review

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

`services/api/internal/postgres/store.go` composes those repositories into a `repository.Store` from a `*sql.DB`. This allows later app-service construction to choose either `repository.NewInMemoryStore()` or `postgres.NewStore(...)` without changing application service methods.

`CoreLoopService.CompleteImportedActivity` now saves both the generated `PlanWeek` and each `PlannedWorkout` through the repository boundary. This matches the Postgres shape where `plan_weeks` and `planned_workouts` are separate tables.

Server startup can read an optional `DATABASE_URL`. `services/api/internal/startup` chooses `repository.NewInMemoryStore()` when it is empty and opens a Postgres `database/sql` handle for `postgres.NewStore(db)` when it is set. Migrations are not run automatically; the configured database is expected to have the schema applied before app services use the Postgres store.

`services/api/internal/app.NewServices` builds application services from the selected store. Future RPC handlers should receive those composed services rather than selecting repositories themselves.

Current repository shape notes:

- `PlanWeek` rows require `athlete_id` and optional `goal_id`; `domain.PlanWeek` now carries those fields directly.
- `postgres.Store` no longer needs plan-week owner options. It can compose repositories from only a `*sql.DB`.
- The `DATABASE_URL` startup path still needs live database integration tests later.
- `PlanWeek` and `PlannedWorkout` are stored separately in Postgres. `GetPlanWeek` returns the week row only; loading workouts for a week should be added as an explicit read model or service method later.
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
