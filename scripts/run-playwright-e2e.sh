#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPOSITORY_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"
cd "$REPOSITORY_ROOT"

PLAYWRIGHT_VERSION="${PLAYWRIGHT_VERSION:-1.62.0}"
PLAYWRIGHT_INSTALL_WITH_DEPS="${PLAYWRIGHT_INSTALL_WITH_DEPS:-0}"
PLAYWRIGHT_API_ORIGIN="${PLAYWRIGHT_API_ORIGIN:-http://127.0.0.1:8091}"
PLAYWRIGHT_APP_ORIGIN="${PLAYWRIGHT_APP_ORIGIN:-http://127.0.0.1:3000}"
E2E_ROOT="$REPOSITORY_ROOT/apps/web/e2e"
PLAYWRIGHT_BIN="$E2E_ROOT/node_modules/.bin/playwright"

fail() {
  printf '%s\n' "PLAYWRIGHT_E2E=FAIL reason=${1:-unspecified}" >&2
  exit 1
}

case "$PLAYWRIGHT_VERSION" in
  1.62.0) ;;
  *) fail "expected Playwright 1.62.0, got $PLAYWRIGHT_VERSION" ;;
esac

case "$PLAYWRIGHT_INSTALL_WITH_DEPS" in
  0|1) ;;
  *) fail 'PLAYWRIGHT_INSTALL_WITH_DEPS must be 0 or 1' ;;
esac

command -v node >/dev/null 2>&1 || fail 'node is required'
command -v npm >/dev/null 2>&1 || fail 'npm is required'
command -v pnpm >/dev/null 2>&1 || fail 'pnpm is required'

rm -rf "$E2E_ROOT/test-results" "$E2E_ROOT/playwright-report"

npm install \
  --prefix "$E2E_ROOT" \
  --no-save \
  --package-lock=false \
  --ignore-scripts \
  --no-audit \
  --no-fund \
  "@playwright/test@$PLAYWRIGHT_VERSION"

test -x "$PLAYWRIGHT_BIN" || fail 'Playwright binary was not installed'
installed_version="$(
  node -p "require('./apps/web/e2e/node_modules/@playwright/test/package.json').version"
)"
test "$installed_version" = "$PLAYWRIGHT_VERSION" ||
  fail "installed Playwright version mismatch: $installed_version"

if test "$PLAYWRIGHT_INSTALL_WITH_DEPS" = "1"; then
  "$PLAYWRIGHT_BIN" install --with-deps chromium
else
  "$PLAYWRIGHT_BIN" install chromium
fi

NEXT_PUBLIC_API_BASE_URL="$PLAYWRIGHT_API_ORIGIN" \
  pnpm --dir apps/web build

export NEXT_PUBLIC_API_BASE_URL="$PLAYWRIGHT_API_ORIGIN"
export PLAYWRIGHT_API_ORIGIN
export PLAYWRIGHT_APP_ORIGIN

"$PLAYWRIGHT_BIN" test \
  --config apps/web/e2e/playwright.config.mjs \
  --project chromium

printf '%s\n' "PLAYWRIGHT_E2E_VERSION=$PLAYWRIGHT_VERSION"
printf '%s\n' 'PLAYWRIGHT_E2E_BROWSER=chromium'
printf '%s\n' 'PLAYWRIGHT_E2E_PUBLIC_NETWORK=DISABLED'
printf '%s\n' 'PLAYWRIGHT_E2E=PASS'
