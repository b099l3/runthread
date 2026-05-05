# Architecture

Runthread uses a Flutter mobile app, a Go backend, ConnectRPC APIs, and Postgres.

The backend owns domain logic. The mobile app presents the plan, workout state, imported activities, adaptations, and explanations. External provider integrations are isolated so provider-specific details do not leak into the core training model.

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
          | Garmin integration   |
          +----------+-----------+
                     |
                     | sqlc queries
                     v
          +----------------------+
          |      Postgres        |
          +----------------------+

          +----------------------+
          | Garmin Provider APIs |
          +----------+-----------+
                     |
                     | Imports activities
                     v
          +----------------------+
          | Garmin integration   |
          +----------------------+
```

## Components

The Flutter app talks to the Go backend through ConnectRPC. It should not duplicate training decisions locally.

The Go backend owns the domain model, planning rules, workout matching, adaptation rules, and integration workflows.

Postgres stores app data, including athletes, goals, plans, planned workouts, imported activities, workout matches, workout results, provider connections, and adaptation events.

sqlc will generate typed Go data access code from SQL queries when the persistence layer is introduced.

Garmin integration imports activities and normalises them into Runthread's `ImportedActivity` model before the core domain evaluates them.

AI can be used to draft explanations or copy. It must not decide workouts, plan changes, training load, or adaptation rules.

