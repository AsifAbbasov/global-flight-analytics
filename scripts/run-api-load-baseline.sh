#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPOSITORY_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"
cd "$REPOSITORY_ROOT"

fail() {
  printf '%s\n' "API_LOAD_BASELINE=FAIL reason=$1" >&2
  exit 1
}

for command_name in docker node git; do
  command -v "$command_name" >/dev/null 2>&1 || fail "$command_name is required"
done

API_LOAD_IMAGE="${API_LOAD_IMAGE:-global-flight-analytics-api:performance}"
K6_IMAGE="${K6_IMAGE:-grafana/k6:1.7.1}"
POSTGRES_IMAGE="${POSTGRES_IMAGE:-postgres:16.14-alpine3.24}"
API_LOAD_RATE_LIMIT_MAX="${API_LOAD_RATE_LIMIT_MAX:-10000}"
PERFORMANCE_SOURCE_SHA="${PERFORMANCE_SOURCE_SHA:-$(git rev-parse HEAD)}"

case "$API_LOAD_RATE_LIMIT_MAX" in
  *[!0-9]*|"") fail 'API_LOAD_RATE_LIMIT_MAX must be a positive integer' ;;
  0) fail 'API_LOAD_RATE_LIMIT_MAX must be greater than zero' ;;
esac

case "$PERFORMANCE_SOURCE_SHA" in
  [0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]) ;;
  *) fail 'PERFORMANCE_SOURCE_SHA must be a full lowercase SHA' ;;
esac

if ! docker image inspect "$API_LOAD_IMAGE" >/dev/null 2>&1; then
  fail "backend image is missing: $API_LOAD_IMAGE"
fi

artifact_dir="$REPOSITORY_ROOT/artifacts/performance"
rm -rf "$artifact_dir"
mkdir -p "$artifact_dir"
chmod 0777 "$artifact_dir"

suffix="${GITHUB_RUN_ID:-local}-$(date +%s)-$$"
network_name="gfa-load-$suffix"
database_name="gfa-load-db-$suffix"
api_name="gfa-load-api-$suffix"
database_id=""
api_id=""

cleanup() {
  status=$?
  if [ "$status" -ne 0 ]; then
    if [ -n "$api_id" ]; then
      docker logs "$api_id" >&2 || true
    fi
    if [ -n "$database_id" ]; then
      docker logs "$database_id" >&2 || true
    fi
  fi
  if [ -n "$api_id" ]; then
    docker rm --force "$api_id" >/dev/null 2>&1 || true
  fi
  if [ -n "$database_id" ]; then
    docker rm --force "$database_id" >/dev/null 2>&1 || true
  fi
  docker network rm "$network_name" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker network create "$network_name" >/dev/null

database_id="$(
  docker run \
    --detach \
    --name "$database_name" \
    --network "$network_name" \
    --env POSTGRES_USER=postgres \
    --env POSTGRES_PASSWORD=postgres \
    --env POSTGRES_DB=global_flight_analytics_load \
    "$POSTGRES_IMAGE"
)"

bash scripts/wait-for-postgres-container.sh \
  "$database_id" \
  postgres \
  global_flight_analytics_load \
  60 \
  1

database_url="postgres://postgres:postgres@${database_name}:5432/global_flight_analytics_load?sslmode=disable"

docker run \
  --rm \
  --network "$network_name" \
  --env DATABASE_URL="$database_url" \
  --env DATABASE_CONNECT_TIMEOUT=5s \
  --env MIGRATIONS_DIR=/app/migrations \
  --env MIGRATION_TIMEOUT=2m \
  "$API_LOAD_IMAGE" \
  /app/migrate

api_readiness_url="http://127.0.0.1:8080/api/v1/ready"

api_id="$(
  docker run \
    --detach \
    --no-healthcheck \
    --name "$api_name" \
    --network "$network_name" \
    --env API_PORT=8080 \
    --env HEALTHCHECK_URL="$api_readiness_url" \
    --env DATABASE_URL="$database_url" \
    --env DATABASE_CONNECT_TIMEOUT=5s \
    --env OPEN_METEO_TIMEOUT=5s \
    --env API_RATE_LIMIT_MAX="$API_LOAD_RATE_LIMIT_MAX" \
    --env API_MUTATION_KEY_SHA256=78fd029a7217aaf71651a747b370be0a76fccebac79bfa02457be3448b636f26 \
    --env METRICS_KEY_SHA256=3eb1bd439947eb762998e566ccc2e099c791118b2f40579cc4f7da2b5061b7f9 \
    "$API_LOAD_IMAGE"
)"

api_ready=false
for attempt in $(seq 1 60); do
  if docker exec "$api_id" /app/healthcheck >/dev/null 2>&1; then
    api_ready=true
    break
  fi

  api_running="$(
    docker inspect \
      --format '{{.State.Running}}' \
      "$api_id"
  )"
  [ "$api_running" = 'true' ] || fail 'API container exited before readiness'
  sleep 1
done

[ "$api_ready" = 'true' ] || fail "API readiness probe did not pass: $api_readiness_url"

printf '%s\n' "API_LOAD_BASELINE_READINESS_URL=$api_readiness_url"
printf '%s\n' 'API_LOAD_BASELINE_READINESS=PASS'

printf '%s\n' "API_LOAD_BASELINE_RATE_LIMIT_MAX=$API_LOAD_RATE_LIMIT_MAX"
printf '%s\n' 'API_LOAD_BASELINE_TARGET=PASS'

export PERFORMANCE_SOURCE_SHA API_LOAD_IMAGE K6_IMAGE POSTGRES_IMAGE

docker run \
  --rm \
  --network "$network_name" \
  --env API_BASE_URL="http://${api_name}:8080" \
  --env PERFORMANCE_SOURCE_SHA \
  --volume "$REPOSITORY_ROOT/load/k6:/scripts:ro" \
  --volume "$artifact_dir:/results" \
  "$K6_IMAGE" \
  run \
  --summary-mode=full \
  /scripts/api-baseline.js

test -s "$artifact_dir/k6-summary.json" || fail 'k6 summary was not created'

node scripts/summarize-api-load-baseline.mjs \
  "$artifact_dir/k6-summary.json" \
  "$artifact_dir/api-load-baseline.json" \
  "$artifact_dir/api-load-baseline.md"

node scripts/validate-api-load-baseline-evidence.mjs \
  "$artifact_dir/api-load-baseline.json"

printf '%s\n' "API_LOAD_BASELINE_SOURCE_SHA=$PERFORMANCE_SOURCE_SHA"
printf '%s\n' "API_LOAD_BASELINE_K6_IMAGE=$K6_IMAGE"
printf '%s\n' 'API_LOAD_BASELINE=PASS'
