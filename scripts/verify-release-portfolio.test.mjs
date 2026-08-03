// RELEASE_PORTFOLIO_CLOSURE_V1

import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import path from 'node:path'

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '..'
)

async function text(relativePath) {
  return readFile(path.join(repositoryRoot, relativePath), 'utf8')
}

test('README presents the implemented product instead of an obsolete first slice', async () => {
  const readme = await text('README.md')
  const start = readme.indexOf('<!-- RELEASE-PORTFOLIO-CLOSURE-V1 -->')
  const end = readme.indexOf('<!-- SOURCE-CONSTRAINTS-OPENSKY-V1 -->')
  assert.ok(start >= 0)
  assert.ok(end > start)
  const portfolio = readme.slice(start, end)
  assert.match(portfolio, /## What Is Implemented/)
  assert.match(portfolio, /pnpm verify:release/)
  assert.match(portfolio, /pnpm smoke:production/)
  assert.match(portfolio, /https:\/\/global-flight-analytics-web\.vercel\.app/)
  assert.match(portfolio, /https:\/\/global-flight-analytics-api\.onrender\.com/)
  assert.doesNotMatch(portfolio, /## MVP Focus/)
  assert.doesNotMatch(portfolio, /## First Coding Slice/)
})

test('root package publishes one release verification and one production smoke entry point', async () => {
  const packageJSON = JSON.parse(await text('package.json'))
  assert.equal(packageJSON.scripts['test:release-contract'], 'node --test scripts/verify-release-portfolio.test.mjs')
  assert.equal(packageJSON.scripts['verify:release-contract'], 'bash scripts/verify-release-portfolio.sh')
  assert.equal(packageJSON.scripts['verify:release'], 'bash scripts/verify-release.sh')
  assert.equal(packageJSON.scripts['smoke:production'], 'bash scripts/smoke-production-release.sh')
})

test('backend and frontend CI both enforce the release contract', async () => {
  const backend = await text('.github/workflows/backend-ci.yml')
  const frontend = await text('.github/workflows/frontend-ci.yml')
  const trigger = backend.slice(0, backend.indexOf('\npermissions:\n'))
  assert.match(trigger, /on:\n  pull_request:\n\n  push:/)
  const trackedFrontendWorkflowPath = "- '.github/workflows/frontend-ci.yml'"
  assert.equal(backend.split(trackedFrontendWorkflowPath).length - 1, 1)
  assert.match(backend, /Verify release and portfolio contract/)
  assert.match(backend, /bash scripts\/verify-release-portfolio\.sh/)
  assert.match(frontend, /Verify release and portfolio contract/)
  assert.match(frontend, /pnpm run test:release-contract/)
  assert.match(frontend, /pnpm run verify:release-contract/)
})

test('production smoke contract covers frontend lifecycle CORS and build provenance', async () => {
  const smoke = await text('scripts/smoke-production-release.sh')
  assert.match(smoke, /FRONTEND_URL/)
  assert.match(smoke, /API_BASE_URL/)
  assert.match(smoke, /api\/v1\/health/)
  assert.match(smoke, /api\/v1\/ready/)
  assert.match(smoke, /api\/v1\/version/)
  assert.match(smoke, /access-control-allow-origin/)
  assert.match(smoke, /EXPECTED_API_REVISION/)
  assert.match(smoke, /PRODUCTION_RELEASE_SMOKE=PASS/)
})

test('deployment runbook separates direct migration and pooled runtime database URLs', async () => {
  const runbook = await text('docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md')
  assert.match(runbook, /direct connection string for migrations/i)
  assert.match(runbook, /pooled connection string for the running API/i)
  assert.match(runbook, /API_PORT=10000/)
  assert.match(runbook, /\/api\/v1\/ready/)
  assert.match(runbook, /Root Directory.*apps\/web/i)
  assert.match(runbook, /PRODUCTION_RELEASE_SMOKE=PASS/)
  assert.doesNotMatch(runbook, /postgres(ql)?:\/\/[^\s]+:[^\s]+@[^\s]+\.neon\.tech/i)
})

test('closure document distinguishes source CI deployment and remaining visual work', async () => {
  const closure = await text('docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md')
  assert.match(closure, /Source implementation/i)
  assert.match(closure, /Exact-commit Continuous Integration/i)
  assert.match(closure, /Public production endpoints/i)
  assert.match(closure, /PUBLIC_API_DEPLOYMENT=CLOSED/)
  assert.match(closure, /PUBLIC_NEXTJS_DEPLOYMENT=CLOSED/)
  assert.match(closure, /FULL_BROWSER_PRODUCTION_SMOKE=CLOSED/)
  assert.match(closure, /FRONTEND_VISUAL_REDESIGN=PLANNED_SEPARATE_PHASE/)
  assert.match(closure, /03ac45dc2a515c77af8d992aa6489816f1cbe927/)
})

test('recruiter guide has a bounded product and code walkthrough', async () => {
  const guide = await text('docs/164_RECRUITER_DEMO_SCRIPT.md')
  assert.match(guide, /Seven-minute walkthrough/)
  assert.match(guide, /Airport Intelligence/)
  assert.match(guide, /Historical Intelligence/)
  assert.match(guide, /Data Quality Lens/)
  assert.match(guide, /engineering decisions/i)
})

test('architecture document records modular monolith boundaries and rejected complexity', async () => {
  const architecture = await text('docs/165_SYSTEM_ARCHITECTURE_AND_DECISIONS.md')
  assert.match(architecture, /modular monolith/i)
  assert.match(architecture, /PostgreSQL/)
  assert.match(architecture, /Next\.js/)
  assert.match(architecture, /Why not microservices/i)
  assert.match(architecture, /Evidence boundary/)
})


test('release truth separates historical evidence from mutable deployment aliases', async () => {
  const readme = await text('README.md')
  const closure = await text('docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md')
  const runbook = await text('docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md')
  const operations = await text('docs/166_BACKEND_OPERATIONS_AND_CI_EVIDENCE_CLOSURE.md')
  const truth = await text('docs/169_RELEASE_TRUTH_AND_DEPLOYMENT_REVISION_CLOSURE.md')

  assert.match(readme, /RELEASE-TRUTH-DEPLOYMENT-REVISION-V1/)
  assert.doesNotMatch(readme, /The public production application is deployed from revision/)
  assert.match(runbook, /DEPLOYED_API_REVISION/)
  assert.doesNotMatch(runbook, /EXPECTED_API_REVISION="\$\(git rev-parse HEAD\)"/)
  assert.match(closure, /Historically verified production application SHA/)
  assert.match(operations, /Historically verified production application SHA/)
  assert.match(truth, /EXPLICIT_DEPLOYMENT_REVISION_INPUT=REQUIRED/)
  assert.match(truth, /RELEASE_TRUTH_CONTRACT=PASS/)
})
