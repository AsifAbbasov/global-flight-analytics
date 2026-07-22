#!/bin/bash
set -euo pipefail

REPOSITORY_ROOT="$(
  cd "$(dirname "$0")/.." &&
    pwd
)"
API_ROOT="$REPOSITORY_ROOT/apps/api"
POSTGRES_CONTAINER=""
API_CONTAINER=""
API_IMAGE="global-flight-analytics-api:stage14-audit-$$"

# Use the exact patched toolchain required by go.mod. Hosts running an older
# Go 1.21+ command can download and select it automatically.
export GOTOOLCHAIN=go1.26.5+auto

cleanup() {
  set +e
  if [ -n "$API_CONTAINER" ]; then
    docker logs "$API_CONTAINER" >/dev/null 2>&1 || true
    docker rm --force "$API_CONTAINER" >/dev/null 2>&1 || true
  fi
  if [ -n "$POSTGRES_CONTAINER" ]; then
    docker rm --force "$POSTGRES_CONTAINER" >/dev/null 2>&1 || true
  fi
  docker image rm --force "$API_IMAGE" >/dev/null 2>&1 || true
}
trap cleanup EXIT

for command_name in git go gofmt pnpm node docker curl awk; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf '%s\n' "ERROR: required command is missing: $command_name" >&2
    exit 1
  fi
done

EFFECTIVE_GO_VERSION="$(go env GOVERSION)"
if [ "$EFFECTIVE_GO_VERSION" != "go1.26.5" ]; then
  printf '%s\n' "ERROR: expected effective Go toolchain go1.26.5, found $EFFECTIVE_GO_VERSION" >&2
  exit 1
fi
if ! grep -Fxq 'go 1.26.5' "$API_ROOT/go.mod"; then
  printf '%s\n' "ERROR: apps/api/go.mod does not pin Go 1.26.5" >&2
  exit 1
fi
if ! grep -Fq 'ARG GO_IMAGE=golang:1.26.5-alpine3.24' "$API_ROOT/Dockerfile"; then
  printf '%s\n' "ERROR: backend Dockerfile does not pin the Go 1.26.5 builder" >&2
  exit 1
fi
echo 'STAGE_14_GO_TOOLCHAIN_AUDIT=PASS'

if ! docker info >/dev/null 2>&1; then
  printf '%s\n' "ERROR: Docker is installed but the Docker daemon is not available" >&2
  exit 1
fi

cd "$REPOSITORY_ROOT"

echo '=== Repository diff validation ==='
git diff --check

echo '=== Go formatting validation ==='
unformatted_files="$(cd "$API_ROOT" && gofmt -l .)"
if [ -n "$unformatted_files" ]; then
  printf '%s\n' "The following Go files are not formatted:" >&2
  printf '%s\n' "$unformatted_files" >&2
  exit 1
fi

echo '=== Stage 14 source audit tests ==='
cd "$API_ROOT"
go test ./tools/stage14finalaudit -count=1
go run ./tools/stage14finalaudit -strict
echo 'STAGE_14_SOURCE_AUDIT=PASS'
echo 'MIGRATOR_CONTEXT_AST_AUDIT=PASS'

echo '=== PostgreSQL layer full audit closure ==='
go test ./tools/postgreslayeraudit -count=1
go run ./tools/postgreslayeraudit -strict
echo 'POSTGRESQL_LAYER_FULL_AUDIT=PASS'

echo '=== Established backend final correctness audit ==='
cd "$REPOSITORY_ROOT"
bash scripts/verify-backend-final-correctness.sh

echo '=== Pinned Go vulnerability analysis ==='
cd "$API_ROOT"
go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...

echo '=== PostgreSQL integration database ==='
POSTGRES_CONTAINER="gfa-stage14-postgres-$$"
docker run \
  --detach \
  --rm \
  --name "$POSTGRES_CONTAINER" \
  --env POSTGRES_USER=postgres \
  --env POSTGRES_PASSWORD=postgres \
  --env POSTGRES_DB=global_flight_analytics_stage14 \
  --publish 127.0.0.1::5432 \
  postgres:16.14-alpine3.24 \
  >/dev/null

postgres_ready="false"
for attempt in $(seq 1 40); do
  if docker exec "$POSTGRES_CONTAINER" \
    pg_isready \
    -U postgres \
    -d global_flight_analytics_stage14 \
    >/dev/null 2>&1; then
    postgres_ready="true"
    break
  fi
  sleep 1
done
if [ "$postgres_ready" != "true" ]; then
  printf '%s\n' "ERROR: Stage 14 PostgreSQL integration database did not become ready" >&2
  exit 1
fi

POSTGRES_PORT="$(
  docker port "$POSTGRES_CONTAINER" 5432/tcp |
    awk -F: 'NR == 1 {print $NF}'
)"
if [ -z "$POSTGRES_PORT" ]; then
  printf '%s\n' "ERROR: could not resolve PostgreSQL integration port" >&2
  exit 1
fi

TEST_DATABASE_URL="postgres://postgres:postgres@127.0.0.1:${POSTGRES_PORT}/global_flight_analytics_stage14?sslmode=disable"
export TEST_DATABASE_URL

cd "$API_ROOT"

echo '=== Production migration catalog ==='
export DATABASE_URL="$TEST_DATABASE_URL"
export DATABASE_CONNECT_TIMEOUT=10s
export MIGRATIONS_DIR="$REPOSITORY_ROOT/database/migrations"
export MIGRATION_TIMEOUT=2m
go run ./cmd/migrate
go run ./cmd/migrate
migration_status="$(go run ./cmd/migrate -status)"
printf '%s
' "$migration_status"
if printf '%s
' "$migration_status" | grep -Eq ' pending$'; then
  printf '%s
' 'ERROR: production migration catalog still contains pending migrations' >&2
  exit 1
fi
MIGRATION_FILE_COUNT="$(find "$MIGRATIONS_DIR" -maxdepth 1 -type f -name '*.sql' | wc -l | tr -d ' ')"
APPLIED_MIGRATION_COUNT="$(
  docker exec "$POSTGRES_CONTAINER" \
    psql \
    -U postgres \
    -d global_flight_analytics_stage14 \
    -Atc 'SELECT COUNT(*) FROM schema_migrations;'
)"
if [ "$APPLIED_MIGRATION_COUNT" != "$MIGRATION_FILE_COUNT" ]; then
  printf '%s\n' "ERROR: applied migration count $APPLIED_MIGRATION_COUNT does not match catalog count $MIGRATION_FILE_COUNT" >&2
  exit 1
fi
echo 'STAGE_14_PRODUCTION_MIGRATOR=PASS'

go test -count=1 \
  ./internal/repository/postgres \
  ./internal/features/featurestore \
  ./internal/routeintelligence/routestore \
  ./internal/historicalintelligence/historicalaggregate

bash "$REPOSITORY_ROOT/scripts/profile-stage-14-trajectory-queries.sh"
echo 'STAGE_14_TRAJECTORY_QUERY_PROFILING=PASS'

echo 'STAGE_14_POSTGRES_INTEGRATION=PASS'

docker rm --force "$POSTGRES_CONTAINER" >/dev/null
POSTGRES_CONTAINER=""
unset TEST_DATABASE_URL

echo '=== Frontend dependency policy ==='
cd "$REPOSITORY_ROOT"
pnpm run test:web-dependency-policy
pnpm run verify:web-dependencies
pnpm audit --prod --audit-level moderate

echo '=== Frontend quality and production build ==='
pnpm --dir apps/web lint
pnpm --dir apps/web typecheck
NEXT_PUBLIC_API_BASE_URL=http://127.0.0.1:8080 \
  pnpm --dir apps/web build
echo 'STAGE_14_FRONTEND_AUDIT=PASS'

echo '=== Backend container contract ==='
docker compose \
  --file "$REPOSITORY_ROOT/compose.yaml" \
  config \
  >/dev/null

docker build \
  --pull \
  --file "$REPOSITORY_ROOT/apps/api/Dockerfile" \
  --tag "$API_IMAGE" \
  "$REPOSITORY_ROOT"

runtime_user="$(
  docker image inspect \
    "$API_IMAGE" \
    --format '{{.Config.User}}'
)"
if [ "$runtime_user" != "10001:10001" ]; then
  printf '%s\n' "ERROR: unexpected backend container runtime user: $runtime_user" >&2
  exit 1
fi

API_CONTAINER="$(
  docker run \
    --detach \
    --publish 127.0.0.1::8080 \
    --env API_PORT=8080 \
    "$API_IMAGE"
)"

api_healthy="false"
for attempt in $(seq 1 40); do
  health_status="$(
    docker inspect \
      --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}missing{{end}}' \
      "$API_CONTAINER"
  )"
  if [ "$health_status" = "healthy" ]; then
    api_healthy="true"
    break
  fi
  if [ "$health_status" = "unhealthy" ] || [ "$health_status" = "missing" ]; then
    printf '%s\n' "ERROR: backend container health status is $health_status" >&2
    exit 1
  fi
  sleep 1
done
if [ "$api_healthy" != "true" ]; then
  printf '%s\n' "ERROR: backend container did not become healthy" >&2
  exit 1
fi

API_PORT="$(
  docker port "$API_CONTAINER" 8080/tcp |
    awk -F: 'NR == 1 {print $NF}'
)"
if [ -z "$API_PORT" ]; then
  printf '%s\n' "ERROR: could not resolve backend container port" >&2
  exit 1
fi

curl \
  --fail \
  --silent \
  --show-error \
  "http://127.0.0.1:${API_PORT}/api/v1/health" \
  >/dev/null

echo 'STAGE_14_CONTAINER_AUDIT=PASS'

docker rm --force "$API_CONTAINER" >/dev/null
API_CONTAINER=""
docker image rm --force "$API_IMAGE" >/dev/null

echo '=== Final repository validation ==='
cd "$REPOSITORY_ROOT"
git diff --check
cd "$API_ROOT"
go run ./tools/stage14finalaudit -root "$REPOSITORY_ROOT" -strict

echo 'STAGE_14_31_WRITE_REPOSITORY_DECOMPOSITION=PASS'
echo 'STAGE_14_32_AIRPORT_PAGINATION=PASS'
echo 'STAGE_14_33_EXPLICIT_CONTEXT_AND_WRITE_MODE=PASS'
echo 'STAGE_14_34_POSTGRESQL_CONTRACT_CONSOLIDATION=PASS'
echo 'STAGE_14_35_TRAJECTORY_QUERY_PROFILING=PASS'
echo 'STAGE_14_36_FINAL_CLOSURE_AUDIT=PASS'
echo 'POST_CLOSURE_MIGRATOR_CONTEXT_HARDENING=PASS'
echo 'STAGE_14_CURRENT_SCOPE_AUDIT=PASS'
echo 'STAGE_14_OVERALL_STATUS=CLOSED'
