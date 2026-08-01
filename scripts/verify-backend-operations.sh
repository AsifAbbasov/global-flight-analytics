#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPOSITORY_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"

fail() {
  printf '%s\n' 'BACKEND_OPERATIONS_CONTRACT=FAIL'
  printf '%s\n' "$1" >&2
  exit 1
}

require_literal() {
  file="$1"
  literal="$2"
  message="$3"
  grep -F -- "$literal" "$file" >/dev/null || fail "$message"
}

for required_file in \
  render.yaml \
  scripts/migrate-production-database.sh \
  scripts/smoke-api-production.sh \
  scripts/verify-backend-operations.test.mjs \
  docs/161_FRONTEND_PRODUCT_HARDENING.md \
  docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md \
  docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md \
  docs/166_BACKEND_OPERATIONS_AND_CI_EVIDENCE_CLOSURE.md \
  apps/api/Dockerfile
do
  test -f "$REPOSITORY_ROOT/$required_file" || fail "required backend operations file is missing: $required_file"
done

require_literal "$REPOSITORY_ROOT/render.yaml" 'autoDeployTrigger: checksPass' \
  'Render deployment must wait for linked checks'
require_literal "$REPOSITORY_ROOT/render.yaml" 'healthCheckPath: /api/v1/ready' \
  'Render readiness health check is missing'
require_literal "$REPOSITORY_ROOT/render.yaml" 'plan: free' \
  'free portfolio deployment plan is not explicit'
require_literal "$REPOSITORY_ROOT/scripts/migrate-production-database.sh" '*-pooler.*' \
  'production migration script does not reject pooled Neon endpoints'
require_literal "$REPOSITORY_ROOT/scripts/smoke-api-production.sh" 'PRODUCTION_API_REVISION=PASS' \
  'API smoke does not verify exact build revision'
require_literal "$REPOSITORY_ROOT/apps/api/Dockerfile" 'ARG RENDER_GIT_COMMIT' \
  'Dockerfile does not accept Render commit provenance'
require_literal "$REPOSITORY_ROOT/apps/api/Dockerfile" 'effective_vcs_ref="${VCS_REF:-${RENDER_GIT_COMMIT:-unknown}}"' \
  'Dockerfile does not preserve explicit revision precedence with Render fallback'
require_literal "$REPOSITORY_ROOT/docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md" \
  '49e474e929dcca5b687464f0a47ce73fcd5a52a7' \
  'release closure does not record the exact source release SHA'
require_literal "$REPOSITORY_ROOT/docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md" \
  '30715613342' \
  'release closure does not record Backend CI evidence'
require_literal "$REPOSITORY_ROOT/docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md" \
  '30715613361' \
  'release closure does not record Frontend CI evidence'
require_literal "$REPOSITORY_ROOT/docs/166_BACKEND_OPERATIONS_AND_CI_EVIDENCE_CLOSURE.md" \
  'Next.js visual and public deployment phase is deliberately deferred' \
  'deferred frontend phase is not explicit'
require_literal "$REPOSITORY_ROOT/.github/workflows/backend-ci.yml" \
  'Verify backend operations contract' \
  'Backend CI does not execute the operations contract'

if grep -F 'preDeployCommand:' "$REPOSITORY_ROOT/render.yaml" >/dev/null; then
  fail 'free Render Blueprint must not claim unavailable pre-deploy commands'
fi

printf '%s\n' 'BACKEND_OPERATIONS_CONTRACT=PASS'
