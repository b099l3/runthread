# Postgres Migrations

This directory contains plain SQL migration files for the Runthread schema.

No versioned migration tool is wired into the project yet. These files are intended to be adopted by the chosen migration tool later, before sqlc queries are added.

For local development, the small shell wrappers in `scripts/db` currently apply and roll back the initial core schema only:

```sh
docker compose -f infra/docker-compose.yml up -d postgres
./scripts/db/migrate-up.sh
```

To roll back the initial core schema:

```sh
./scripts/db/migrate-down.sh
```

These scripts pipe the SQL files into the local Docker Compose Postgres container with `psql`. They do not track migration versions.

Provider-specific tables are defined in `000002_provider_connections.up.sql` and `000002_provider_connections.down.sql`. They should be adopted by the same versioned migration workflow before Stage 8 implementation depends on them.
