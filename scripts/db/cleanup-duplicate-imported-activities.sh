#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=../env/incantation.sh
source "$ROOT_DIR/scripts/env/incantation.sh" >/dev/null

APPLY=0
ATHLETE_ID="${RUNTHREAD_ATHLETE_ID:-11111111-1111-1111-1111-111111111111}"

usage() {
  cat <<USAGE
Usage:
  $0 [--apply] [athlete-id]

Preview or clean duplicate imported_activities for one athlete.

Options:
  --apply      Delete duplicate imported_activities after relinking references.
  athlete-id   Defaults to RUNTHREAD_ATHLETE_ID, then local demo athlete.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --apply)
      APPLY=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      ATHLETE_ID="$1"
      shift
      ;;
  esac
done

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is not set. Check $ROOT_DIR/.env.local." >&2
  exit 1
fi

echo "Athlete: $ATHLETE_ID"

echo "Duplicate groups before cleanup:"
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -v athlete_id="$ATHLETE_ID" <<'SQL'
SELECT
  type,
  started_at,
  distance_meters,
  duration_seconds,
  count(*) AS imported_count
FROM imported_activities
WHERE athlete_id = :'athlete_id'
GROUP BY type, started_at, distance_meters, duration_seconds
HAVING count(*) > 1
ORDER BY imported_count DESC, started_at DESC
LIMIT 25;
SQL

if [[ "$APPLY" -ne 1 ]]; then
  echo "Preview only. Re-run with --apply to clean duplicates."
  exit 0
fi

echo "Applying duplicate cleanup..."
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -v athlete_id="$ATHLETE_ID" <<'SQL'
BEGIN;

CREATE TEMP TABLE duplicate_imported_activity_ids ON COMMIT DROP AS
WITH duplicate_groups AS (
  SELECT
    (array_agg(id ORDER BY created_at ASC, id ASC))[1] AS keep_id,
    array_agg(id ORDER BY created_at ASC, id ASC) AS all_ids
  FROM imported_activities
  WHERE athlete_id = :'athlete_id'
  GROUP BY athlete_id, type, started_at, distance_meters, duration_seconds
  HAVING count(*) > 1
)
SELECT keep_id, unnest(all_ids[2:]) AS delete_id
FROM duplicate_groups;

UPDATE provider_activities pa
SET imported_activity_id = d.keep_id
FROM duplicate_imported_activity_ids d
WHERE pa.imported_activity_id = d.delete_id;

UPDATE workout_results wr
SET imported_activity_id = d.keep_id
FROM duplicate_imported_activity_ids d
WHERE wr.imported_activity_id = d.delete_id;

DELETE FROM workout_matches wm
USING duplicate_imported_activity_ids d
WHERE wm.imported_activity_id = d.delete_id;

DELETE FROM imported_activities ia
USING duplicate_imported_activity_ids d
WHERE ia.id = d.delete_id;

COMMIT;
SQL

echo "Duplicate groups after cleanup:"
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -v athlete_id="$ATHLETE_ID" <<'SQL'
SELECT
  type,
  started_at,
  distance_meters,
  duration_seconds,
  count(*) AS imported_count
FROM imported_activities
WHERE athlete_id = :'athlete_id'
GROUP BY type, started_at, distance_meters, duration_seconds
HAVING count(*) > 1
ORDER BY imported_count DESC, started_at DESC
LIMIT 25;
SQL
