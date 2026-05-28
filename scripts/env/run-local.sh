#!/usr/bin/env sh
# Run a command with local Runthread environment variables:
#   ./scripts/env/run-local.sh psql "$DATABASE_URL" -c "select 1"

set -eu

ROOT_DIR="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
ENV_FILE="${RUNTHREAD_ENV_FILE:-$ROOT_DIR/.env.local}"

if [ ! -f "$ENV_FILE" ]; then
  echo "Runthread env file not found: $ENV_FILE" >&2
  exit 1
fi

set -a
. "$ENV_FILE"
set +a

exec "$@"
