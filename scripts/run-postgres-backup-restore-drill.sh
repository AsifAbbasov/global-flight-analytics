#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPOSITORY_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"
cd "$REPOSITORY_ROOT"

fail() {
  printf '%s\n' "POSTGRES_BACKUP_RESTORE_DRILL=FAIL reason=$1" >&2
  exit 1
}

for command_name in docker; do
  command -v "$command_name" >/dev/null 2>&1 || fail "$command_name is required"
done

BACKEND_IMAGE="${BACKEND_IMAGE:-global-flight-analytics-api:security-hardening}"
POSTGRES_IMAGE="${POSTGRES_IMAGE:-postgres:16.14-alpine3.24}"
SOURCE_SHA="${SOURCE_SHA:-local}"
ARTIFACT_DIR="${ARTIFACT_DIR:-$REPOSITORY_ROOT/artifacts/security/postgres-backup-restore}"

case "$SOURCE_SHA" in
  local|[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]) ;;
  *) fail "SOURCE_SHA must be 'local' or a full lowercase Git SHA" ;;
esac

docker image inspect "$BACKEND_IMAGE" >/dev/null 2>&1 ||
  fail "backend image is missing: $BACKEND_IMAGE"

rm -rf "$ARTIFACT_DIR"
mkdir -p "$ARTIFACT_DIR"
chmod 0700 "$ARTIFACT_DIR"

suffix="${GITHUB_RUN_ID:-local}-$(date +%s)-$$"
network_name="gfa-backup-$suffix"
postgres_name="gfa-backup-postgres-$suffix"
source_database="gfa_backup_source"
restore_database="gfa_backup_restore"
postgres_id=""

cleanup() {
  status=$?
  if [ "$status" -ne 0 ] && [ -n "$postgres_id" ]; then
    docker logs "$postgres_id" >&2 || true
  fi
  if [ -n "$postgres_id" ]; then
    docker rm --force "$postgres_id" >/dev/null 2>&1 || true
  fi
  docker network rm "$network_name" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker network create "$network_name" >/dev/null

postgres_id="$(
  docker run \
    --detach \
    --name "$postgres_name" \
    --network "$network_name" \
    --env POSTGRES_USER=postgres \
    --env POSTGRES_PASSWORD=postgres \
    --env POSTGRES_DB="$source_database" \
    "$POSTGRES_IMAGE"
)"

bash scripts/wait-for-postgres-container.sh \
  "$postgres_id" \
  postgres \
  "$source_database" \
  60 \
  1

source_database_url="postgres://postgres:postgres@${postgres_name}:5432/${source_database}?sslmode=disable"

docker run \
  --rm \
  --network "$network_name" \
  --env DATABASE_URL="$source_database_url" \
  --env DATABASE_CONNECT_TIMEOUT=5s \
  --env MIGRATIONS_DIR=/app/migrations \
  --env MIGRATION_TIMEOUT=2m \
  "$BACKEND_IMAGE" \
  /app/migrate

docker exec \
  "$postgres_id" \
  psql \
  --username postgres \
  --dbname "$source_database" \
  --set ON_ERROR_STOP=1 \
  --command "
    INSERT INTO airports (
      icao_code,
      name,
      latitude,
      longitude,
      source_name
    )
    VALUES (
      'ZZZZ',
      'Backup Drill Airport',
      0,
      0,
      'backup-drill'
    );
  " >/dev/null

source_marker_count="$(
  docker exec \
    "$postgres_id" \
    psql \
    --username postgres \
    --dbname "$source_database" \
    --no-align \
    --tuples-only \
    --set ON_ERROR_STOP=1 \
    --command "SELECT COUNT(*) FROM airports WHERE source_name = 'backup-drill';" |
    tr -d '[:space:]'
)"
[ "$source_marker_count" = "1" ] ||
  fail "source fixture was not created exactly once"

container_backup="/tmp/gfa-backup-restore-drill.dump"
host_backup="$ARTIFACT_DIR/gfa-backup-restore-drill.dump"

docker exec \
  "$postgres_id" \
  pg_dump \
  --username postgres \
  --dbname "$source_database" \
  --format custom \
  --no-owner \
  --no-privileges \
  --file "$container_backup"

docker cp \
  "$postgres_id:$container_backup" \
  "$host_backup" >/dev/null

test -s "$host_backup" || fail "backup file is empty"
chmod 0600 "$host_backup"

if command -v sha256sum >/dev/null 2>&1; then
  checksum="$(sha256sum "$host_backup" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  checksum="$(shasum -a 256 "$host_backup" | awk '{print $1}')"
else
  fail "sha256sum or shasum is required"
fi

printf '%s  %s\n' \
  "$checksum" \
  "$(basename "$host_backup")" \
  > "$host_backup.sha256"
chmod 0600 "$host_backup.sha256"

case "$checksum" in
  [0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]) ;;
  *) fail "unexpected SHA-256 checksum format" ;;
esac

docker exec \
  "$postgres_id" \
  createdb \
  --username postgres \
  "$restore_database"

docker exec \
  "$postgres_id" \
  pg_restore \
  --username postgres \
  --dbname "$restore_database" \
  --no-owner \
  --no-privileges \
  --exit-on-error \
  "$container_backup"

critical_tables=(
  airports
  ingestion_runs
  flight_states
  flight_trajectories
)

for table_name in "${critical_tables[@]}"; do
  present="$(
    docker exec \
      "$postgres_id" \
      psql \
      --username postgres \
      --dbname "$restore_database" \
      --no-align \
      --tuples-only \
      --set ON_ERROR_STOP=1 \
      --command "SELECT CASE WHEN to_regclass('public.${table_name}') IS NULL THEN 0 ELSE 1 END;" |
      tr -d '[:space:]'
  )"
  [ "$present" = "1" ] ||
    fail "critical table is missing after restore: $table_name"
done

declare -A source_counts
declare -A restore_counts

for table_name in "${critical_tables[@]}"; do
  source_counts["$table_name"]="$(
    docker exec \
      "$postgres_id" \
      psql \
      --username postgres \
      --dbname "$source_database" \
      --no-align \
      --tuples-only \
      --set ON_ERROR_STOP=1 \
      --command "SELECT COUNT(*) FROM ${table_name};" |
      tr -d '[:space:]'
  )"
  restore_counts["$table_name"]="$(
    docker exec \
      "$postgres_id" \
      psql \
      --username postgres \
      --dbname "$restore_database" \
      --no-align \
      --tuples-only \
      --set ON_ERROR_STOP=1 \
      --command "SELECT COUNT(*) FROM ${table_name};" |
      tr -d '[:space:]'
  )"

  [ "${source_counts[$table_name]}" = "${restore_counts[$table_name]}" ] ||
    fail "row count mismatch for $table_name"
done

restore_marker_count="$(
  docker exec \
    "$postgres_id" \
    psql \
    --username postgres \
    --dbname "$restore_database" \
    --no-align \
    --tuples-only \
    --set ON_ERROR_STOP=1 \
    --command "SELECT COUNT(*) FROM airports WHERE source_name = 'backup-drill';" |
    tr -d '[:space:]'
)"
[ "$restore_marker_count" = "1" ] ||
  fail "restored fixture count is not exactly one"

source_marker_after_restore="$(
  docker exec \
    "$postgres_id" \
    psql \
    --username postgres \
    --dbname "$source_database" \
    --no-align \
    --tuples-only \
    --set ON_ERROR_STOP=1 \
    --command "SELECT COUNT(*) FROM airports WHERE source_name = 'backup-drill';" |
    tr -d '[:space:]'
)"
[ "$source_marker_after_restore" = "$source_marker_count" ] ||
  fail "source database changed during restore drill"

evidence_file="$ARTIFACT_DIR/evidence.json"
cat > "$evidence_file" <<EOF
{
  "source_sha": "$SOURCE_SHA",
  "postgres_image": "$POSTGRES_IMAGE",
  "backup_format": "postgres-custom",
  "backup_sha256": "$checksum",
  "source_database_unchanged": true,
  "critical_tables_verified": [
    "airports",
    "ingestion_runs",
    "flight_states",
    "flight_trajectories"
  ],
  "row_counts": {
    "airports": ${restore_counts[airports]},
    "ingestion_runs": ${restore_counts[ingestion_runs]},
    "flight_states": ${restore_counts[flight_states]},
    "flight_trajectories": ${restore_counts[flight_trajectories]}
  },
  "fixture_airports_restored": $restore_marker_count
}
EOF
chmod 0600 "$evidence_file"

printf '%s\n' "POSTGRES_BACKUP_RESTORE_SOURCE_SHA=$SOURCE_SHA"
printf '%s\n' "POSTGRES_BACKUP_RESTORE_SHA256=$checksum"
printf '%s\n' "POSTGRES_BACKUP_RESTORE_AIRPORT_ROWS=${restore_counts[airports]}"
printf '%s\n' "POSTGRES_BACKUP_RESTORE_SOURCE_UNCHANGED=PASS"
printf '%s\n' "POSTGRES_BACKUP_RESTORE_DRILL=PASS"
