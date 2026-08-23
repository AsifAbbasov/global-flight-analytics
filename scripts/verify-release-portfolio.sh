#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPOSITORY_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"

fail() {
  printf '%s\n' 'RELEASE_PORTFOLIO_CONTRACT=FAIL'
  printf '%s\n' "$1" >&2
  exit 1
}

require_file() {
  test -f "$REPOSITORY_ROOT/$1" || fail "required release file is missing: $1"
}

require_literal() {
  file="$REPOSITORY_ROOT/$1"
  literal="$2"
  message="$3"
  grep -F -- "$literal" "$file" >/dev/null || fail "$message"
}

for required_file in \
  README.md \
  package.json \
  compose.yaml \
  apps/api/Dockerfile \
  apps/api/.env.example \
  apps/web/.env.example \
  .github/workflows/backend-ci.yml \
  .github/workflows/frontend-ci.yml \
  scripts/verify-recruiter-quickstart.sh \
  scripts/verify-release.sh \
  scripts/smoke-production-release.sh \
  docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md \
  docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md \
  docs/164_RECRUITER_DEMO_SCRIPT.md \
  docs/165_SYSTEM_ARCHITECTURE_AND_DECISIONS.md \
  docs/169_RELEASE_TRUTH_AND_DEPLOYMENT_REVISION_CLOSURE.md \
  docs/170_PRODUCTION_OBSERVABILITY_AND_ALERTING_CLOSURE.md
do
  require_file "$required_file"
done

require_literal README.md '<!-- RELEASE-PORTFOLIO-CLOSURE-V1 -->' \
  'README release marker is missing'
require_literal README.md '<!-- RELEASE-TRUTH-DEPLOYMENT-REVISION-V1 -->' \
  'README release truth marker is missing'
require_literal README.md '## What Is Implemented' \
  'README implementation summary is missing'
require_literal README.md 'pnpm verify:release' \
  'README release verification command is missing'
require_literal README.md 'pnpm smoke:production' \
  'README production smoke command is missing'
require_literal README.md 'https://global-flight-analytics-web.vercel.app' \
  'README verified frontend URL is missing'
require_literal README.md 'https://global-flight-analytics-api.onrender.com' \
  'README verified API URL is missing'
require_literal README.md 'PRODUCTION_RELEASE_SMOKE=PASS' \
  'README verified full production smoke marker is missing'
require_literal README.md 'The frontend product redesign source implementation is complete' \
  'README frontend redesign source implementation boundary is missing'
require_literal README.md '<!-- CURRENT-ENGINEERING-STATUS-2026-08-V1 -->' \
  'README current engineering status marker is missing'
require_literal README.md 'OPENAPI_CONTRACT_PATHS=38' \
  'README current OpenAPI path count is missing'
require_literal README.md 'PLAYWRIGHT_E2E_BROWSER_SCENARIOS=20' \
  'README current Playwright browser scenario count is missing'
require_literal README.md 'FRONTEND_PRODUCT_SOURCE_IMPLEMENTATION=COMPLETE' \
  'README frontend product source implementation marker is missing'
require_literal README.md 'FRONTEND_PRODUCT_CLOSURE=CANDIDATE_EXACT_HEAD_CI' \
  'README frontend product closure candidate boundary is missing'
require_literal README.md 'FRONTEND_VISUAL_AND_INTERACTION_REDESIGN=IMPLEMENTED' \
  'README frontend visual redesign implementation marker is missing'
require_literal README.md 'PIXEL_GOLDEN_VISUAL_REGRESSION=OPEN' \
  'README pixel-golden visual regression boundary is missing'
require_literal README.md 'DOCUMENT_INDEX_194_195=OPEN_GOVERNANCE_DEBT' \
  'README document index governance debt boundary is missing'
require_literal README.md 'FINAL_EXACT_PRODUCTION_VALIDATION=OPEN' \
  'README final exact-production validation boundary is missing'
require_literal README.md 'V1_RELEASE=OPEN' \
  'README v1 release boundary is missing'

require_literal README.md '<!-- PRODUCTION-OBSERVABILITY-CLOSURE-V1 -->' \
  'README production observability marker is missing'

readme_portfolio_section="$(
  awk '
    /<!-- RELEASE-PORTFOLIO-CLOSURE-V1 -->/ { capture = 1 }
    /<!-- SOURCE-CONSTRAINTS-OPENSKY-V1 -->/ { exit }
    capture { print }
  ' "$REPOSITORY_ROOT/README.md"
)"

printf '%s\n' "$readme_portfolio_section" | grep -F '## MVP Focus' >/dev/null && \
  fail 'obsolete MVP Focus heading remains in the portfolio entry section'
printf '%s\n' "$readme_portfolio_section" | grep -F '## First Coding Slice' >/dev/null && \
  fail 'obsolete First Coding Slice heading remains in the portfolio entry section'
printf '%s\n' "$readme_portfolio_section" | grep -F 'Creating the Neon and Render resources still requires' >/dev/null && \
  fail 'README still claims that production infrastructure has not been created'
printf '%s\n' "$readme_portfolio_section" | grep -F 'public deployment phase is deliberately deferred' >/dev/null && \
  fail 'README still claims that public frontend deployment is deferred'
printf '%s\n' "$readme_portfolio_section" | grep -F 'The public production application is deployed from revision' >/dev/null && \
  fail 'README presents historical deployment evidence as a perpetually current revision'

require_literal package.json '"verify:release": "bash scripts/verify-release.sh"' \
  'root verify:release script is missing'
require_literal package.json '"smoke:production": "bash scripts/smoke-production-release.sh"' \
  'root production smoke script is missing'
require_literal package.json '"test:release-contract": "node --test scripts/verify-release-portfolio.test.mjs"' \
  'root release contract test script is missing'
require_literal package.json '"verify:release-contract": "bash scripts/verify-release-portfolio.sh"' \
  'root release contract verifier script is missing'

require_literal .github/workflows/backend-ci.yml "- '.github/workflows/frontend-ci.yml'" \
  'Backend CI does not trigger when the frontend workflow changes'
require_literal .github/workflows/backend-ci.yml 'Verify release and portfolio contract' \
  'Backend CI release contract step is missing'
require_literal .github/workflows/backend-ci.yml 'bash scripts/verify-release-portfolio.sh' \
  'Backend CI release verifier command is missing'
require_literal .github/workflows/frontend-ci.yml 'Verify release and portfolio contract' \
  'Frontend CI release contract step is missing'
require_literal .github/workflows/frontend-ci.yml 'pnpm run test:release-contract' \
  'Frontend CI release contract tests are missing'
require_literal .github/workflows/frontend-ci.yml 'pnpm run verify:release-contract' \
  'Frontend CI release contract verifier is missing'

backend_trigger_header="$(
  awk '
    /^on:/ { capture = 1 }
    /^permissions:/ { exit }
    capture { print }
  ' "$REPOSITORY_ROOT/.github/workflows/backend-ci.yml"
)"

backend_pull_request_count="$(
  printf '%s\n' "$backend_trigger_header" |
    awk '
      $0 == "  pull_request:" { count += 1 }
      END { print count + 0 }
    '
)"

if [ "$backend_pull_request_count" -ne 1 ]; then
  fail 'Backend CI must run for every pull request'
fi

backend_pull_request_block="$(
  printf '%s\n' "$backend_trigger_header" |
    awk '
      $0 == "  pull_request:" { capture = 1; next }
      $0 == "  push:" { exit }
      capture { print }
    '
)"

backend_pull_request_paths_count="$(
  printf '%s\n' "$backend_pull_request_block" |
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

if [ "$backend_pull_request_paths_count" -ne 0 ]; then
  fail 'Backend CI pull_request trigger must not use path filters'
fi

backend_frontend_workflow_count="$(
  grep -F -c -- "- '.github/workflows/frontend-ci.yml'" \
    "$REPOSITORY_ROOT/.github/workflows/backend-ci.yml"
)"
[ "$backend_frontend_workflow_count" -eq 1 ] || \
  fail 'Backend CI push path filters must track frontend workflow changes exactly once'

require_literal apps/api/.env.example 'Use a direct PostgreSQL connection string for migrations' \
  'API environment example does not distinguish migration connection semantics'
require_literal apps/api/.env.example 'Render-compatible deployments can set API_PORT=10000' \
  'API environment example does not document hosted port configuration'
require_literal apps/web/.env.example 'Production must use the public HTTPS API origin' \
  'frontend environment example does not document the production API origin'

require_literal docs/DOCUMENT_INDEX.md '## Document 162 — Release and Portfolio Closure' \
  'Document 162 index entry is missing'
require_literal docs/DOCUMENT_INDEX.md '## Document 163 — Production Deployment Runbook' \
  'Document 163 index entry is missing'
require_literal docs/DOCUMENT_INDEX.md '## Document 164 — Recruiter Demo Script' \
  'Document 164 index entry is missing'
require_literal docs/DOCUMENT_INDEX.md '## Document 165 — System Architecture and Decisions' \
  'Document 165 index entry is missing'
require_literal docs/DOCUMENT_INDEX.md '## Document 169 — Release Truth and Deployment Revision Closure' \
  'Document 169 index entry is missing'

require_literal docs/DOCUMENT_INDEX.md '## Document 170 — Production Observability and Alerting Closure' \
  'Document 170 index entry is missing'

require_literal docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md 'PUBLIC_API_DEPLOYMENT=CLOSED' \
  'release closure does not record verified public API deployment'
require_literal docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md 'PUBLIC_NEXTJS_DEPLOYMENT=CLOSED' \
  'release closure does not record verified public Next.js deployment'
require_literal docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md 'FULL_BROWSER_PRODUCTION_SMOKE=CLOSED' \
  'release closure does not record verified browser production smoke'
require_literal docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md 'PRODUCTION_RELEASE_SMOKE=PASS' \
  'deployment runbook does not record the verified full production smoke'
require_literal docs/169_RELEASE_TRUTH_AND_DEPLOYMENT_REVISION_CLOSURE.md 'RELEASE_TRUTH_CONTRACT=PASS' \
  'release truth closure marker is missing'
require_literal docs/169_RELEASE_TRUTH_AND_DEPLOYMENT_REVISION_CLOSURE.md 'EXPLICIT_DEPLOYMENT_REVISION_INPUT=REQUIRED' \
  'release truth closure does not require an explicit deployment revision'

require_literal docs/170_PRODUCTION_OBSERVABILITY_AND_ALERTING_CLOSURE.md 'PRODUCTION_METRICS_REMOTE_WRITE=CLOSED' \
  'production observability closure does not record remote-write completion'
require_literal docs/170_PRODUCTION_OBSERVABILITY_AND_ALERTING_CLOSURE.md 'GRAFANA_SLO_DASHBOARD=CLOSED' \
  'production observability closure does not record the SLO dashboard'
require_literal docs/170_PRODUCTION_OBSERVABILITY_AND_ALERTING_CLOSURE.md 'GRAFANA_ALERT_RULES=CLOSED' \
  'production observability closure does not record managed alert rules'
require_literal docs/170_PRODUCTION_OBSERVABILITY_AND_ALERTING_CLOSURE.md 'ALERT_NOTIFICATION_DELIVERY=CLOSED' \
  'production observability closure does not record notification delivery'
require_literal docs/170_PRODUCTION_OBSERVABILITY_AND_ALERTING_CLOSURE.md 'PRODUCTION_OBSERVABILITY_CLOSURE=PASS' \
  'production observability closure marker is missing'

for release_document in \
  docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md \
  docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md \
  docs/164_RECRUITER_DEMO_SCRIPT.md \
  docs/165_SYSTEM_ARCHITECTURE_AND_DECISIONS.md \
  docs/169_RELEASE_TRUTH_AND_DEPLOYMENT_REVISION_CLOSURE.md \
  docs/170_PRODUCTION_OBSERVABILITY_AND_ALERTING_CLOSURE.md
do
  grep -F 'REPLACE_WITH_' "$REPOSITORY_ROOT/$release_document" >/dev/null && \
    fail "unresolved release placeholder remains in $release_document"
done

if grep -F 'EXPECTED_API_REVISION="$(git rev-parse HEAD)"' \
  "$REPOSITORY_ROOT/README.md" \
  "$REPOSITORY_ROOT/docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md" >/dev/null; then
  fail 'release documentation derives deployed runtime revision from local HEAD'
fi

if grep -E 'postgres(ql)?://[^[:space:]]+:[^[:space:]]+@[^[:space:]]+\.neon\.tech' \
  "$REPOSITORY_ROOT/docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md" \
  "$REPOSITORY_ROOT/docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md" >/dev/null; then
  fail 'release documentation contains what appears to be a real Neon credential'
fi

node "$REPOSITORY_ROOT/scripts/verify-documentation-english.mjs"

printf '%s\n' 'RELEASE_PORTFOLIO_CONTRACT=PASS'
