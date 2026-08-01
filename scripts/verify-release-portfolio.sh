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
  docs/165_SYSTEM_ARCHITECTURE_AND_DECISIONS.md
do
  require_file "$required_file"
done

require_literal README.md '<!-- RELEASE-PORTFOLIO-CLOSURE-V1 -->' \
  'README release marker is missing'
require_literal README.md '## What Is Implemented' \
  'README implementation summary is missing'
require_literal README.md 'pnpm verify:release' \
  'README release verification command is missing'
require_literal README.md 'pnpm smoke:production' \
  'README production smoke command is missing'
require_literal README.md 'No screenshot, live URL or green check is claimed unless it has been produced for the' \
  'README exact-evidence policy is missing'

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

backend_frontend_workflow_count="$(
  grep -F -c -- "- '.github/workflows/frontend-ci.yml'" \
    "$REPOSITORY_ROOT/.github/workflows/backend-ci.yml"
)"
[ "$backend_frontend_workflow_count" -eq 2 ] || \
  fail 'Backend CI must track frontend workflow changes for push and pull_request'

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

for release_document in \
  docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md \
  docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md \
  docs/164_RECRUITER_DEMO_SCRIPT.md \
  docs/165_SYSTEM_ARCHITECTURE_AND_DECISIONS.md
do
  grep -F 'REPLACE_WITH_' "$REPOSITORY_ROOT/$release_document" >/dev/null && \
    fail "unresolved release placeholder remains in $release_document"
done

if grep -E 'postgres(ql)?://[^[:space:]]+:[^[:space:]]+@[^[:space:]]+\.neon\.tech' \
  "$REPOSITORY_ROOT/docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md" >/dev/null; then
  fail 'deployment runbook contains what appears to be a real Neon credential'
fi

printf '%s\n' 'RELEASE_PORTFOLIO_CONTRACT=PASS'
