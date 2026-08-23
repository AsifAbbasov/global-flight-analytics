// BACKEND_OPERATIONS_EVIDENCE_CLOSURE_V1
import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = resolve(import.meta.dirname, '..')
const read = path => readFileSync(resolve(root, path), 'utf8')

const releaseSHA = '49e474e929dcca5b687464f0a47ce73fcd5a52a7'
const productionSHA = '6bca02a8ed1487195b165ae9ced3ca687a373666'

test('README records exact CI closure verified public deployment and the current frontend boundary', () => {
  const source = read('README.md')
  assert.match(source, new RegExp(releaseSHA))
  assert.match(source, /Backend CI run `30715613342`/)
  assert.match(source, /Frontend CI run `30715613361`/)
  assert.match(source, new RegExp(productionSHA))
  const sourceLines = new Set(source.split('\n'))
  assert.equal(sourceLines.has('- Frontend: `https://global-flight-analytics-web.vercel.app`'), true)
  assert.equal(sourceLines.has('- API: `https://global-flight-analytics-api.onrender.com`'), true)
  assert.match(source, /FRONTEND_PRODUCT_SOURCE_IMPLEMENTATION=COMPLETE/)
  assert.match(source, /FRONTEND_PRODUCT_CLOSURE=CLOSED/)
  assert.match(source, /FRONTEND_VISUAL_AND_INTERACTION_REDESIGN=IMPLEMENTED/)
  assert.match(source, /FRONTEND_VISUAL_POLISH_V2=CLOSED/)
  assert.match(source, /STRUCTURAL_VISUAL_REGRESSION=CLOSED/)
  assert.match(source, /RETAINED_SCREENSHOT_EVIDENCE=CLOSED/)
  assert.match(source, /PIXEL_GOLDEN_VISUAL_REGRESSION=NOT_ADOPTED_NONBLOCKING/)
  assert.match(source, /FINAL_EXACT_PRODUCTION_VALIDATION=OPEN/)
})

test('root package publishes backend operations commands', () => {
  const packageJSON = JSON.parse(read('package.json'))
  assert.equal(packageJSON.scripts['test:backend-operations-contract'], 'node --test scripts/verify-backend-operations.test.mjs')
  assert.equal(packageJSON.scripts['verify:backend-operations-contract'], 'bash scripts/verify-backend-operations.sh')
  assert.equal(packageJSON.scripts['migrate:production-database'], 'bash scripts/migrate-production-database.sh')
  assert.equal(packageJSON.scripts['smoke:api-production'], 'bash scripts/smoke-api-production.sh')
})

test('Render Blueprint is a CI-gated free Docker API service', () => {
  const source = read('render.yaml')
  for (const literal of [
    'runtime: docker',
    'plan: free',
    'region: frankfurt',
    'autoDeployTrigger: checksPass',
    'dockerfilePath: ./apps/api/Dockerfile',
    'dockerContext: .',
    'healthCheckPath: /api/v1/ready',
  ]) {
    assert.ok(source.includes(literal), `missing ${literal}`)
  }
  assert.ok(!source.includes('preDeployCommand:'))
})

test('Blueprint keeps production credentials outside source control', () => {
  const source = read('render.yaml')
  for (const key of ['DATABASE_URL', 'API_ALLOWED_ORIGINS', 'API_MUTATION_KEY_SHA256']) {
    assert.match(source, new RegExp(`- key: ${key}\\n\\s+sync: false`))
  }
})

test('production migration command requires a direct TLS database connection', () => {
  const source = read('scripts/migrate-production-database.sh')
  assert.match(source, /PRODUCTION_DATABASE_MIGRATION_URL/)
  assert.match(source, /\*-pooler\.\*/)
  assert.match(source, /sslmode=require/)
  assert.match(source, /go run \.\/cmd\/migrate/)
  assert.match(source, /production migrations require a clean working tree/)
})

test('Docker and Backend CI use one Render-aware revision fallback', () => {
  const dockerfile = read('apps/api/Dockerfile')
  const runtimeStage = dockerfile.slice(dockerfile.indexOf('FROM scratch AS runtime'))
  const workflow = read('.github/workflows/backend-ci.yml')
  const containerJob = workflow.slice(workflow.indexOf('  backend-container:'))

  assert.match(dockerfile, /ARG RENDER_GIT_COMMIT/)
  assert.match(dockerfile, /effective_vcs_ref=\"\$\{VCS_REF:-\$\{RENDER_GIT_COMMIT:-unknown\}\}\"/)
  assert.match(dockerfile, /buildinfo\.revision=\$\{effective_vcs_ref\}/)
  assert.match(runtimeStage, /ARG RENDER_GIT_COMMIT/)
  assert.match(runtimeStage, /org\.opencontainers\.image\.revision=\"\$\{VCS_REF:-\$\{RENDER_GIT_COMMIT:-unknown\}\}\"/)
  assert.match(containerJob, /--build-arg RENDER_GIT_COMMIT=\"\$GITHUB_SHA\"/)
  assert.doesNotMatch(containerJob, /--build-arg VCS_REF=/)
})

test('production historical materializer ships in the backend runtime image', () => {
  const dockerfile = read('apps/api/Dockerfile')
  const workflow = read('.github/workflows/backend-ci.yml')
  const containerJob = workflow.slice(workflow.indexOf('  backend-container:'))

  assert.ok(dockerfile.includes('        materialize-historical-intelligence \\\n'))
  assert.match(containerJob, /Verify production historical materializer binary/)
  assert.match(containerJob, /\/app\/materialize-historical-intelligence/)
  assert.match(containerJob, /\n\s+-h \\\n\s+> \/dev\/null 2>&1/)
})

test('API production smoke verifies lifecycle and exact build provenance', () => {
  const source = read('scripts/smoke-api-production.sh')
  for (const endpoint of ['/api/v1/health', '/api/v1/ready', '/api/v1/version']) {
    assert.ok(source.includes(endpoint))
  }
  assert.match(source, /EXPECTED_API_REVISION/)
  assert.match(source, /PRODUCTION_API_REVISION=PASS/)
  assert.match(source, /attempt=1/)
  assert.match(source, /did not become available after four attempts/)
})

test('Backend CI permanently enforces operations contracts and Blueprint changes', () => {
  const source = read('.github/workflows/backend-ci.yml')
  const trigger = source.slice(0, source.indexOf('\npermissions:\n'))
  assert.match(trigger, /on:\n  pull_request:\n\n  push:/)
  const trackedRenderBlueprintPath = "- 'render.yaml'"
  assert.equal(source.split(trackedRenderBlueprintPath).length - 1, 1)
  assert.match(source, /Verify backend operations contract/)
  assert.match(source, /node --test scripts\/verify-backend-operations\.test\.mjs/)
  assert.match(source, /bash scripts\/verify-backend-operations\.sh/)
})

test('PostgreSQL CI executes the production replay migration integration contract', () => {
  const workflow = read('.github/workflows/backend-ci.yml')
  const replayTest = read('apps/api/internal/database/migrationfile/production_replay_migration_integration_test.go')
  const postgresJobStart = workflow.indexOf('  postgres-integration:')
  const postgresJobEnd = workflow.indexOf('\n\n  backend-ci-gate:', postgresJobStart)

  assert.notEqual(postgresJobStart, -1, 'missing PostgreSQL integration job')
  assert.notEqual(postgresJobEnd, -1, 'missing PostgreSQL integration job boundary')

  const postgresJob = workflow.slice(postgresJobStart, postgresJobEnd)
  assert.match(postgresJob, /TEST_DATABASE_URL:/)
  assert.match(postgresJob, /go test -count=1/)
  assert.match(postgresJob, /\.\/internal\/database\/migrationfile/)
  assert.match(replayTest, /func TestProductionReplayMigrationEnforcesObservationIdentity/)
  assert.match(replayTest, /os\.Getenv\("TEST_DATABASE_URL"\)/)
  assert.match(replayTest, /postgresError\.Code != "23505"/)
})

test('closure documents record exact production evidence without exposing secrets', () => {
  const hardening = read('docs/161_FRONTEND_PRODUCT_HARDENING.md')
  const release = read('docs/162_RELEASE_AND_PORTFOLIO_CLOSURE.md')
  const operations = read('docs/166_BACKEND_OPERATIONS_AND_CI_EVIDENCE_CLOSURE.md')
  assert.match(hardening, /30715613361/)
  assert.match(hardening, /PRODUCTION_RELEASE_SMOKE=PASS/)
  assert.match(release, /EXACT_COMMIT_CI_EVIDENCE=CLOSED/)
  assert.match(release, /PUBLIC_API_DEPLOYMENT=CLOSED/)
  assert.match(release, /PUBLIC_NEXTJS_DEPLOYMENT=CLOSED/)
  assert.match(operations, /PRODUCTION_DATABASE_MIGRATION=CLOSED/)
  assert.match(operations, /FULL_BROWSER_PRODUCTION_SMOKE=CLOSED/)
  assert.match(operations, /FRONTEND_VISUAL_REDESIGN=PLANNED_SEPARATE_PHASE/)
  for (const source of [hardening, release, operations]) {
    assert.doesNotMatch(source, /postgres(ql)?:\/\/[^\s]+:[^\s]+@[^\s]+\.neon\.tech/i)
  }
})

test('deployment runbook distinguishes free-plan migration from paid pre-deploy automation', () => {
  const source = read('docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md')
  assert.match(source, /free Render web service/)
  assert.match(source, /pnpm migrate:production-database/)
  assert.match(source, /paid Render service/)
  assert.match(source, /pre-deploy command/)
  assert.match(source, /PRODUCTION_RELEASE_SMOKE=PASS/)
})
