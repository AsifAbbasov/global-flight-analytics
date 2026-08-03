#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPOSITORY_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"

fail() {
  printf '%s\n' "RECRUITER_QUICKSTART_CONTRACT=FAIL"
  printf '%s\n' "$1" >&2
  exit 1
}

require_literal() {
  file="$1"
  literal="$2"
  message="$3"

  grep -F -- "$literal" "$file" > /dev/null || fail "$message"
}

README_FILE="$REPOSITORY_ROOT/README.md"
COMPOSE_FILE="$REPOSITORY_ROOT/compose.yaml"
DOCKER_DOCUMENT="$REPOSITORY_ROOT/docs/29_REPRODUCIBLE_DOCKER.md"
WORKFLOW_FILE="$REPOSITORY_ROOT/.github/workflows/backend-ci.yml"

for required_file in \
  "$README_FILE" \
  "$COMPOSE_FILE" \
  "$DOCKER_DOCUMENT" \
  "$WORKFLOW_FILE"
do
  test -f "$required_file" || fail "required quickstart source is missing: $required_file"
done

require_literal "$README_FILE" '<!-- RECRUITER-QUICKSTART-V1 -->' \
  'README quickstart marker is missing'
require_literal "$README_FILE" 'docker compose config' \
  'README does not validate Docker Compose configuration'
require_literal "$README_FILE" 'docker compose up --build --detach' \
  'README does not start the real Compose environment'
require_literal "$README_FILE" 'http://127.0.0.1:8080/api/v1/health' \
  'README health endpoint is missing'
require_literal "$README_FILE" 'http://127.0.0.1:8080/api/v1/ready' \
  'README readiness endpoint is missing'
require_literal "$README_FILE" 'http://127.0.0.1:8080/api/v1/version' \
  'README version endpoint is missing'
require_literal "$README_FILE" 'pnpm install --frozen-lockfile' \
  'README does not require a frozen workspace install'
require_literal "$README_FILE" 'pnpm dev:web' \
  'README does not use the root frontend development command'

require_literal "$COMPOSE_FILE" 'API_MUTATION_KEY_SHA256:' \
  'Compose API service does not configure mutation authorization startup state'

compose_digest="$(
  sed -n \
    's/^[[:space:]]*API_MUTATION_KEY_SHA256: ${API_MUTATION_KEY_SHA256:-\([0-9a-f][0-9a-f]*\)}[[:space:]]*$/\1/p' \
    "$COMPOSE_FILE"
)"

if ! printf '%s\n' "$compose_digest" | grep -Eq '^[0-9a-f]{64}$'; then
  fail 'Compose local mutation-key digest must be an overridable 64-character lowercase SHA-256 value'
fi

require_literal "$COMPOSE_FILE" 'condition: service_completed_successfully' \
  'Compose API service must wait for successful migrations'
require_literal "$DOCKER_DOCUMENT" '<!-- RECRUITER-QUICKSTART-V1:DOCUMENT-29 -->' \
  'reproducible Docker documentation does not record the recruiter quickstart contract'
require_literal "$WORKFLOW_FILE" "- 'README.md'" \
  'Backend CI path filters do not include README.md'
require_literal "$WORKFLOW_FILE" 'Verify recruiter quickstart contract' \
  'Backend CI does not execute the recruiter quickstart verifier'
require_literal "$WORKFLOW_FILE" 'bash scripts/verify-recruiter-quickstart.sh' \
  'Backend CI quickstart verifier command is missing'

trigger_header="$(
  awk '
    /^on:/ { capture = 1 }
    /^permissions:/ { exit }
    capture { print }
  ' "$WORKFLOW_FILE"
)"

pull_request_count="$(
  printf '%s\n' "$trigger_header" |
    awk '
      $0 == "  pull_request:" { count += 1 }
      END { print count + 0 }
    '
)"

if [ "$pull_request_count" -ne 1 ]; then
  fail 'Backend CI must run for every pull request'
fi

pull_request_block="$(
  printf '%s\n' "$trigger_header" |
    awk '
      $0 == "  pull_request:" { capture = 1; next }
      $0 == "  push:" { exit }
      capture { print }
    '
)"

pull_request_paths_count="$(
  printf '%s\n' "$pull_request_block" |
    awk '
      {
        line = $0
        sub(/^[[:space:]]*/, "", line)
        if (line == "paths:") {
          count += 1
        }
      }
      END { print count + 0 }
    '
)"

if [ "$pull_request_paths_count" -ne 0 ]; then
  fail 'Backend CI pull_request trigger must not use path filters'
fi

readme_path_count="$(
  grep -F -c -- "- 'README.md'" "$WORKFLOW_FILE"
)"
if [ "$readme_path_count" -ne 1 ]; then
  fail 'Backend CI push path filters must include README.md exactly once'
fi

printf '%s\n' "RECRUITER_QUICKSTART_CONTRACT=PASS"
