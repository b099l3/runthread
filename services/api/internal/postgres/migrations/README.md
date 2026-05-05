# Postgres Migrations

This directory contains plain SQL migration files for the initial Runthread schema.

No versioned migration tool is wired into the project yet. These files are intended to be adopted by the chosen migration tool later, before sqlc queries are added.

For local development, use the small shell wrappers in `scripts/db`:

```sh
docker compose -f infra/docker-compose.yml up -d postgres
./scripts/db/migrate-up.sh
```

To roll back the initial schema:

```sh
./scripts/db/migrate-down.sh
```

These scripts pipe the SQL files into the local Docker Compose Postgres container with `psql`. They do not track migration versions.

Provider-specific Garmin tables are intentionally not included yet.
