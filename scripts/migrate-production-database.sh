#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf '%s\n' 'PRODUCTION_DATABASE_MIGRATION=FAIL'
  printf '%s\n' "$1" >&2
  exit 1
}

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPOSITORY_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"

command -v go >/dev/null 2>&1 || fail 'Go is required'
command -v git >/dev/null 2>&1 || fail 'Git is required'

: "${PRODUCTION_DATABASE_MIGRATION_URL:?Set PRODUCTION_DATABASE_MIGRATION_URL to the direct PostgreSQL connection string}"

case "$PRODUCTION_DATABASE_MIGRATION_URL" in
  postgres://*|postgresql://*) ;;
  *) fail 'PRODUCTION_DATABASE_MIGRATION_URL must be a PostgreSQL connection string' ;;
esac

case "$PRODUCTION_DATABASE_MIGRATION_URL" in
  *-pooler.*) fail 'Migrations require the direct Neon connection string, not the pooled endpoint' ;;
esac

case "$PRODUCTION_DATABASE_MIGRATION_URL" in
  *sslmode=require*|*sslmode=verify-full*) ;;
  *) fail 'Production migration connection must require TLS' ;;
esac

cd "$REPOSITORY_ROOT"
CURRENT_SHA="$(git rev-parse HEAD)"
EXPECTED_RELEASE_SHA="${EXPECTED_RELEASE_SHA:-$CURRENT_SHA}"
[ "$CURRENT_SHA" = "$EXPECTED_RELEASE_SHA" ] || \
  fail "working tree commit $CURRENT_SHA does not match EXPECTED_RELEASE_SHA $EXPECTED_RELEASE_SHA"
[ -z "$(git status --porcelain --untracked-files=all)" ] || \
  fail 'production migrations require a clean working tree'

(
  cd apps/api
  DATABASE_URL="$PRODUCTION_DATABASE_MIGRATION_URL" \
  MIGRATIONS_DIR="${MIGRATIONS_DIR:-../../database/migrations}" \
  MIGRATION_TIMEOUT="${MIGRATION_TIMEOUT:-2m}" \
  go run ./cmd/migrate
)

printf '%s\n' "PRODUCTION_DATABASE_MIGRATION_SHA=$CURRENT_SHA"
printf '%s\n' 'PRODUCTION_DATABASE_MIGRATION=PASS'
