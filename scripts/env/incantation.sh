#!/usr/bin/env sh
# Source this file to summon local Runthread environment variables:
#   . ./scripts/env/incantation.sh

set -a

ROOT_DIR="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
ENV_FILE="${RUNTHREAD_ENV_FILE:-$ROOT_DIR/.env.local}"

if [ ! -f "$ENV_FILE" ]; then
  echo "Runthread env file not found: $ENV_FILE" >&2
  set +a
  return 1 2>/dev/null || exit 1
fi

. "$ENV_FILE"

set +a

echo "Runthread env incantation complete: $ENV_FILE"
