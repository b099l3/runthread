#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
MIGRATION_FILE="$ROOT_DIR/services/api/internal/postgres/migrations/000001_initial_core_schema.up.sql"
COMPOSE_FILE="$ROOT_DIR/infra/docker-compose.yml"

DB_NAME="${POSTGRES_DB:-runthread}"
DB_USER="${POSTGRES_USER:-runthread}"

docker compose -f "$COMPOSE_FILE" exec -T postgres \
  psql -v ON_ERROR_STOP=1 -U "$DB_USER" -d "$DB_NAME" < "$MIGRATION_FILE"

