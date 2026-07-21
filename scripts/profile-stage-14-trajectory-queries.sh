#!/bin/bash
set -euo pipefail

REPOSITORY_ROOT="$(
  cd "$(dirname "$0")/.." &&
    pwd
)"
API_ROOT="$REPOSITORY_ROOT/apps/api"

if [ -z "${TEST_DATABASE_URL:-}" ]; then
  printf '%s\n' 'ERROR: TEST_DATABASE_URL is required for trajectory query profiling' >&2
  exit 1
fi

export GOTOOLCHAIN=go1.26.5+auto
cd "$API_ROOT"
go test -count=1 -v \
  -run '^TestTrajectoryQueryProfilesUseExpectedIndexes$' \
  ./internal/repository/postgres

echo 'STAGE_14_TRAJECTORY_QUERY_PROFILING=PASS'
