# Runthread

Runthread is a Garmin-first adaptive running training app.

The core loop is:

1. Plan a workout.
2. The runner completes it on Garmin.
3. The activity is imported.
4. The activity is matched to the planned workout.
5. The plan adapts if needed.
6. The runner gets a calm explanation of what changed and why.

The product should become a real business, so the repo is structured for a Flutter mobile app, Go backend, ConnectRPC APIs, Postgres persistence, and a clean domain model.

## Repository Structure

```text
runthread/
  apps/
    mobile/          Flutter mobile app, to be scaffolded later
  services/
    api/             Go backend service
  docs/              Product, architecture, domain, roadmap, decisions
  infra/             Local and deployment infrastructure
  .codex/            Working agreement and reusable session prompts
```

## Local Development

The backend currently exposes only a health endpoint.

Prerequisites:

- Go 1.22 or newer.
- Docker Compose for local Postgres.

```sh
cd services/api
go run ./cmd/server
```

Then check:

```sh
curl http://localhost:8080/healthz
```

Local Postgres is defined in Docker Compose:

```sh
docker compose -f infra/docker-compose.yml up
```

Flutter has not been scaffolded yet. The mobile app directory exists as a placeholder until the MVP screens stage.

## Current Status

Stage 0: project foundation.

This repo currently contains documentation, a staged roadmap, a minimal Go API skeleton, and local Postgres infrastructure. It does not yet contain the planning engine, Garmin OAuth, activity matching, subscriptions, or AI explanation generation.

Real Garmin API access is assumed to require validation or application before Stage 8.
