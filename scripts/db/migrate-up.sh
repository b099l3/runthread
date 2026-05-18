#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
MIGRATION_DIR="$ROOT_DIR/services/api/internal/postgres/migrations"
COMPOSE_FILE="$ROOT_DIR/infra/docker-compose.yml"

DB_NAME="${POSTGRES_DB:-runthread}"
DB_USER="${POSTGRES_USER:-runthread}"

for migration_file in "$MIGRATION_DIR"/*.up.sql; do
  docker compose -f "$COMPOSE_FILE" exec -T postgres \
    psql -v ON_ERROR_STOP=1 -U "$DB_USER" -d "$DB_NAME" < "$migration_file"
done
