#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf '%s\n' 'PRODUCTION_RELEASE_SMOKE=FAIL'
  printf '%s\n' "$1" >&2
  exit 1
}

command -v curl >/dev/null 2>&1 || fail 'curl is required'
command -v node >/dev/null 2>&1 || fail 'node is required'

: "${FRONTEND_URL:?Set FRONTEND_URL to the deployed frontend origin}"
: "${API_BASE_URL:?Set API_BASE_URL to the deployed API origin}"

FRONTEND_URL="${FRONTEND_URL%/}"
API_BASE_URL="${API_BASE_URL%/}"
EXPECTED_API_REVISION="${EXPECTED_API_REVISION:-}"
ALLOW_HTTP_RELEASE_SMOKE="${ALLOW_HTTP_RELEASE_SMOKE:-0}"

validate_url() {
  label="$1"
  value="$2"
  case "$value" in
    https://*) ;;
    http://127.0.0.1:*|http://localhost:*)
      [ "$ALLOW_HTTP_RELEASE_SMOKE" = '1' ] || \
        fail "$label must use HTTPS outside an explicitly allowed local smoke run"
      ;;
    *) fail "$label must be an absolute HTTPS URL" ;;
  esac
}

validate_url FRONTEND_URL "$FRONTEND_URL"
validate_url API_BASE_URL "$API_BASE_URL"

FRONTEND_ORIGIN="$(node -e 'console.log(new URL(process.argv[1]).origin)' "$FRONTEND_URL")"
API_ORIGIN="$(node -e 'console.log(new URL(process.argv[1]).origin)' "$API_BASE_URL")"
[ "$FRONTEND_ORIGIN" != "$API_ORIGIN" ] || \
  fail 'frontend and API origins must be distinct for the production CORS smoke contract'

TEMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/gfa-production-smoke.XXXXXX")"
cleanup() {
  rm -rf "$TEMP_ROOT"
}
trap cleanup EXIT

fetch_api() {
  path="$1"
  label="$2"
  expected_status="$3"
  body="$TEMP_ROOT/${label}.json"
  headers="$TEMP_ROOT/${label}.headers"

  curl \
    --fail \
    --silent \
    --show-error \
    --location \
    --max-time 30 \
    --header "Accept: application/json" \
    --header "Origin: $FRONTEND_ORIGIN" \
    --dump-header "$headers" \
    --output "$body" \
    "$API_BASE_URL$path"

  node -e '
    const fs = require("node:fs")
    const [file, label, expected, expectedRevision] = process.argv.slice(1)
    let payload
    try {
      payload = JSON.parse(fs.readFileSync(file, "utf8"))
    } catch (error) {
      throw new Error(`${label} returned invalid JSON: ${error.message}`)
    }
    if (!payload || payload.success !== true || typeof payload.data !== "object" || payload.data === null) {
      throw new Error(`${label} did not return the successful API envelope`)
    }
    if (expected && payload.data.status !== expected) {
      throw new Error(`${label} status must be ${expected}, received ${payload.data.status}`)
    }
    if (label === "version") {
      for (const field of ["version", "revision", "built_at"]) {
        if (typeof payload.data[field] !== "string" || payload.data[field].trim() === "") {
          throw new Error(`version response field ${field} is missing`)
        }
      }
      if (expectedRevision && payload.data.revision !== expectedRevision) {
        throw new Error(`deployed API revision ${payload.data.revision} does not match ${expectedRevision}`)
      }
    }
  ' "$body" "$label" "$expected_status" "$EXPECTED_API_REVISION"

  cors_origin="$(
    grep -i '^access-control-allow-origin:' "$headers" \
      | tail -n 1 \
      | cut -d: -f2- \
      | tr -d '\r' \
      | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
  )"
  [ "$cors_origin" = "$FRONTEND_ORIGIN" ] || \
    fail "$label CORS origin is '$cors_origin', expected '$FRONTEND_ORIGIN'"
}

fetch_api '/api/v1/health' health ok
fetch_api '/api/v1/ready' readiness ready
fetch_api '/api/v1/version' version ''

frontend_body="$TEMP_ROOT/frontend.html"
frontend_headers="$TEMP_ROOT/frontend.headers"
curl \
  --fail \
  --silent \
  --show-error \
  --location \
  --max-time 30 \
  --dump-header "$frontend_headers" \
  --output "$frontend_body" \
  "$FRONTEND_URL/"

grep -i '^content-type:.*text/html' "$frontend_headers" >/dev/null || \
  fail 'frontend did not return an HTML content type'
grep -F 'Global Flight Analytics' "$frontend_body" >/dev/null || \
  fail 'frontend HTML does not contain the product identity'

printf '%s\n' 'PRODUCTION_FRONTEND=PASS'
printf '%s\n' 'PRODUCTION_API_HEALTH=PASS'
printf '%s\n' 'PRODUCTION_API_READINESS=PASS'
printf '%s\n' 'PRODUCTION_API_VERSION=PASS'
printf '%s\n' 'PRODUCTION_CORS=PASS'
printf '%s\n' 'PRODUCTION_RELEASE_SMOKE=PASS'
