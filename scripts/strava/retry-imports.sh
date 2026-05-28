#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=../env/incantation.sh
source "$ROOT_DIR/scripts/env/incantation.sh" >/dev/null

API_BASE_URL="${RUNTHREAD_API_BASE_URL:-http://localhost:8080}"
ATHLETE_ID="${RUNTHREAD_ATHLETE_ID:-11111111-1111-1111-1111-111111111111}"
GOAL_ID="${RUNTHREAD_GOAL_ID:-}"
TARGET_WEEK_DATE="${1:-}"
LIMIT="${RETRY_IMPORT_LIMIT:-100}"

payload="{\"athleteId\":\"$ATHLETE_ID\",\"goalId\":\"$GOAL_ID\",\"limit\":$LIMIT"
if [[ -n "$TARGET_WEEK_DATE" ]]; then
  payload="$payload,\"targetWeekDate\":\"$TARGET_WEEK_DATE\""
fi
payload="$payload}"

curl -sS \
  -X POST \
  -H "Content-Type: application/json" \
  -d "$payload" \
  "$API_BASE_URL/providers/strava/retry-imports"
echo
