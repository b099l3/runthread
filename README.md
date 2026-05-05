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

The backend currently exposes a health endpoint, the first implemented core-loop RPC, and the first read-side RPC for the Flutter MVP.

Prerequisites:

- `asdf`.
- Go `1.25.5` installed through `asdf`.
- sqlc `1.31.1` installed through `asdf`.
- buf `1.69.0` installed through `asdf`.
- Docker Compose for local Postgres.

Tool versions are pinned in `.tool-versions`.

```sh
asdf install
asdf current
```

Go, sqlc, and buf are pinned today. Flutter will be pinned when the mobile app is scaffolded.

```sh
cd services/api
go run ./cmd/server
```

If your shell is not using asdf shims, use `asdf exec go run ./cmd/server`.

The server defaults to `:8080`. Override it with:

```sh
RUNTHREAD_SERVER_ADDR=:9090 asdf exec go run ./cmd/server
```

Without `DATABASE_URL`, server startup uses an in-memory repository store. If `DATABASE_URL` is set, startup opens a Postgres `database/sql` handle and composes the sqlc-backed repository store. The current RPC handlers use the selected store, but the richer current-plan snapshot query is still implemented only for the in-memory store.

Then check:

```sh
curl http://localhost:8080/healthz
```

The first implemented ConnectRPC endpoint is also mounted:

```text
/runthread.v1.RunthreadService/CompleteImportedActivity
```

This first RPC is still a backend-loop test boundary. It accepts athlete, goal, and imported activity payloads directly until auth, stored plan reads, and real provider import are wired.

The API also defines `GetCurrentPlanWeek` for the Flutter MVP read path. It can return a saved plan week by ID, or generate a deterministic demo week from stored athlete and goal records when no saved week exists. Request-supplied athlete, goal, and week IDs are temporary until auth and current-plan lookup exist.

Local Postgres is defined in Docker Compose:

```sh
docker compose -f infra/docker-compose.yml up -d postgres
```

Run the initial schema migration:

```sh
./scripts/db/migrate-up.sh
```

Roll back the initial schema migration:

```sh
./scripts/db/migrate-down.sh
```

The first sqlc scaffold lives in `services/api/sqlc.yaml`, with queries under `services/api/internal/postgres/queries` and generated Go code under `services/api/internal/postgres/db`.
Generate database code from the API service directory:

```sh
cd services/api
sqlc generate
```

If your shell is not using asdf shims, run:

```sh
cd services/api
asdf exec sqlc generate
```

The initial protobuf API contract lives under `services/api/proto`. Buf configuration lives in `services/api/buf.yaml` and `services/api/buf.gen.yaml`.
Generate protobuf and ConnectRPC Go code from the API service directory:

```sh
cd services/api
asdf exec buf generate
```

Generated RPC code lives under `services/api/internal/rpc`. The generation config uses remote buf plugins for Go protobuf and ConnectRPC output, so local `protoc-gen-*` binaries are not required.

Run backend tests:

```sh
cd services/api
asdf exec go test ./...
```

Flutter has not been scaffolded yet. The mobile app directory exists as a placeholder until the MVP screens stage.

## Current Status

Backend foundation through Stage 6 is in place, plus an in-memory core-loop test harness.

This repo currently contains documentation, a minimal Go API skeleton, local Postgres infrastructure, provider-neutral domain models, deterministic planning, workout flow helpers, mock Garmin activity normalisation, activity matching, adaptation event generation, in-memory repository boundaries, app-service persistence wiring for plan weeks and first-class planned workouts, application service composition, server configuration and startup storage composition, an initial provider-neutral protobuf API contract, buf generation config, generated protobuf/ConnectRPC Go code, the first thin ConnectRPC handler, an implemented in-memory read-side RPC for the Flutter MVP, plain SQL schema migrations, local migration scripts, generated sqlc code for the core tables, sqlc-backed repositories and a Postgres store composition layer for the current core repository interfaces, and end-to-end backend core-loop tests.

It does not yet contain versioned migration tooling, live database integration tests, auth, Postgres-backed read-side query implementations, Flutter screens, real Garmin OAuth, subscriptions, or AI explanation generation.

Real Garmin API access is assumed to require validation or application before Stage 8.
