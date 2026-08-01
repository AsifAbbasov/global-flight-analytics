#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf '%s\n' 'PRODUCTION_API_SMOKE=FAIL'
  printf '%s\n' "$1" >&2
  exit 1
}

command -v curl >/dev/null 2>&1 || fail 'curl is required'
command -v node >/dev/null 2>&1 || fail 'node is required'

: "${API_BASE_URL:?Set API_BASE_URL to the deployed API origin}"
: "${EXPECTED_API_REVISION:?Set EXPECTED_API_REVISION to the exact deployed commit SHA}"

API_BASE_URL="${API_BASE_URL%/}"
ALLOW_HTTP_API_SMOKE="${ALLOW_HTTP_API_SMOKE:-0}"

case "$API_BASE_URL" in
  https://*) ;;
  http://127.0.0.1:*|http://localhost:*)
    [ "$ALLOW_HTTP_API_SMOKE" = '1' ] || fail 'HTTP is allowed only for an explicit local smoke run'
    ;;
  *) fail 'API_BASE_URL must be an absolute HTTPS URL' ;;
esac

printf '%s' "$EXPECTED_API_REVISION" | grep -Eq '^[0-9a-f]{40}$' || \
  fail 'EXPECTED_API_REVISION must be a full forty-character lowercase Git SHA'

TEMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/gfa-api-production-smoke.XXXXXX")"
cleanup() {
  rm -rf "$TEMP_ROOT"
}
trap cleanup EXIT

fetch_api() {
  path="$1"
  label="$2"
  expected_status="$3"
  body="$TEMP_ROOT/${label}.json"

  attempt=1
  while true; do
    if curl \
      --fail \
      --silent \
      --show-error \
      --location \
      --max-time 90 \
      --header 'Accept: application/json' \
      --output "$body" \
      "$API_BASE_URL$path"
    then
      break
    fi
    if [ "$attempt" -ge 4 ]; then
      fail "$label endpoint did not become available after four attempts"
    fi
    attempt=$((attempt + 1))
    sleep 15
  done

  node -e '
    const fs = require("node:fs")
    const [file, label, expectedStatus, expectedRevision] = process.argv.slice(1)
    let payload
    try {
      payload = JSON.parse(fs.readFileSync(file, "utf8"))
    } catch (error) {
      throw new Error(`${label} returned invalid JSON: ${error.message}`)
    }
    if (!payload || payload.success !== true || typeof payload.data !== "object" || payload.data === null) {
      throw new Error(`${label} did not return the successful API envelope`)
    }
    if (expectedStatus && payload.data.status !== expectedStatus) {
      throw new Error(`${label} status must be ${expectedStatus}, received ${payload.data.status}`)
    }
    if (label === "version") {
      for (const field of ["version", "revision", "built_at"]) {
        if (typeof payload.data[field] !== "string" || payload.data[field].trim() === "") {
          throw new Error(`version response field ${field} is missing`)
        }
      }
      if (payload.data.revision !== expectedRevision) {
        throw new Error(`deployed API revision ${payload.data.revision} does not match ${expectedRevision}`)
      }
    }
  ' "$body" "$label" "$expected_status" "$EXPECTED_API_REVISION"
}

fetch_api '/api/v1/health' health ok
fetch_api '/api/v1/ready' readiness ready
fetch_api '/api/v1/version' version ''

printf '%s\n' 'PRODUCTION_API_HEALTH=PASS'
printf '%s\n' 'PRODUCTION_API_READINESS=PASS'
printf '%s\n' 'PRODUCTION_API_VERSION=PASS'
printf '%s\n' 'PRODUCTION_API_REVISION=PASS'
printf '%s\n' 'PRODUCTION_API_SMOKE=PASS'
