import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

const roadmap = read('../../docs/24_MVP_VERSION_ROADMAP.md')
const sequence = read('../../docs/25_IMPLEMENTATION_SEQUENCE.md')
const index = read('../../docs/DOCUMENT_INDEX.md')
const findings = read('../../docs/FINDING_REGISTER.md')
const reconciliation = read('../../docs/212_VERSION_2_RECONCILIATION_AUDIT.md')
const readme = read('../../README.md')

test('Version 2 reconciliation preserves the original 18-item roadmap and records the bounded release disposition', () => {
  assert.match(roadmap, /VERSION_2_ORIGINAL_SCOPE_ITEMS=18/)
  assert.match(roadmap, /VERSION_2_IMPLEMENTED_CAPABILITIES=14/)
  assert.match(roadmap, /VERSION_2_BOUNDED_UNCALIBRATED_CAPABILITIES=1/)
  assert.match(roadmap, /VERSION_2_DEFERRED_RESEARCH_CAPABILITIES=3/)
  assert.match(roadmap, /VERSION_2_DISCRETE_FRECHET=DEFERRED_RESEARCH/)
  assert.match(roadmap, /VERSION_2_TRAJECTORY_SPATIAL_INDEX=DEFERRED_RESEARCH/)
  assert.match(roadmap, /VERSION_2_WEATHER_GRID=DEFERRED_RESEARCH/)
  assert.match(
    roadmap,
    /VERSION_2_SIMILARITY_THRESHOLD_POLICY=IMPLEMENTED_BOUNDED_UNCALIBRATED/,
  )
})

test('Version 2 reconciliation keeps advanced research out of unsupported production claims', () => {
  assert.match(reconciliation, /VERSION_2_DISCRETE_FRECHET=DEFERRED_RESEARCH/)
  assert.match(reconciliation, /VERSION_2_TRAJECTORY_SPATIAL_INDEX=DEFERRED_RESEARCH/)
  assert.match(reconciliation, /VERSION_2_WEATHER_GRID=DEFERRED_RESEARCH/)
  assert.match(reconciliation, /VERSION_2_NEW_ANALYTICAL_ENGINE=NONE/)
  assert.match(reconciliation, /VERSION_2_NEW_BACKEND_ENDPOINT=NONE/)
  assert.match(reconciliation, /VERSION_2_NEW_DATABASE_TABLE=NONE/)
  assert.match(reconciliation, /VERSION_2_NEW_MIGRATION=NONE/)
  assert.match(reconciliation, /VERSION_2_ADDITIONAL_COST=0_RUB/)
})

test('Implementation Sequence records the repository-real Stage 15 through Stage 22 progression', () => {
  for (const stage of [15, 16, 17, 18, 19, 20, 21, 22]) {
    assert.match(sequence, new RegExp(`Stage ${stage}`))
  }
  assert.match(sequence, /VERSION_2_RECONCILIATION=CANDIDATE/)
  assert.match(sequence, /DISCRETE_FRECHET=DEFERRED_RESEARCH/)
  assert.match(sequence, /TRAJECTORY_SPATIAL_INDEX=DEFERRED_RESEARCH/)
  assert.match(sequence, /WEATHER_GRID=DEFERRED_RESEARCH/)
})

test('Documentation Index registers every post-196 evidence document and reconciliation audit', () => {
  for (const documentNumber of [197, 198, 199, 200, 201, 202, 203, 204, 205, 206, 207, 208, 209, 210, 211, 212]) {
    assert.match(index, new RegExp(`Document ${documentNumber}`))
  }
  assert.match(index, /212_VERSION_2_RECONCILIATION_AUDIT\.md/)
})

test('Finding Register owns the documentation drift without closing the independent security-settings finding', () => {
  assert.match(findings, /GFA-GOV-457/)
  assert.match(findings, /Version 2 roadmap, index, implementation sequence and README drift/)
  assert.match(findings, /GFA-SEC-445[^\n]*IN_PROGRESS/)
  assert.match(findings, /Canonical finding register covers 457 findings/)
})

test('README publishes the current OpenAPI surface and bounded Version 2 candidate state', () => {
  assert.match(readme, /OPENAPI_CONTRACT_OPERATIONS=39/)
  assert.match(readme, /OPENAPI_PUBLIC_READ_OPERATIONS=38/)
  assert.match(readme, /OPENAPI_PROTECTED_MUTATION_OPERATIONS=1/)
  assert.match(readme, /VERSION_2_RECONCILIATION=CANDIDATE/)
  assert.match(readme, /Discrete Fréchet[^\n]*DEFERRED_RESEARCH/)
  assert.match(readme, /Weather Grid[^\n]*DEFERRED_RESEARCH/)
})

test('README preserves historical v1 release-contract markers without presenting them as the current Version 2 surface', () => {
  assert.match(readme, /Historical v1\.0\.0 release-contract baseline/)
  assert.match(readme, /OPENAPI_CONTRACT_PATHS=38/)
  assert.match(readme, /PLAYWRIGHT_E2E_BROWSER_SCENARIOS=20/)
  assert.match(readme, /PLAYWRIGHT_E2E_MOCK_SCENARIOS=7/)
  const currentSurfaceStart = readme.indexOf('## Contract and Test Surface')
  const historicalBaselineStart = readme.indexOf('### Historical v1.0.0 release-contract baseline')
  assert.ok(currentSurfaceStart >= 0)
  assert.ok(historicalBaselineStart > currentSurfaceStart)
  const currentSurface = readme.slice(currentSurfaceStart, historicalBaselineStart)
  assert.match(currentSurface, /OPENAPI_CONTRACT_OPERATIONS=39/)
  assert.doesNotMatch(currentSurface, /OPENAPI_CONTRACT_PATHS=38/)
})

test('The reconciliation document cannot self-close before exact-head and post-merge evidence exist', () => {
  assert.match(reconciliation, /VERSION_2_RECONCILIATION=CANDIDATE/)
  assert.match(
    reconciliation,
    /VERSION_2_CANONICAL_CLOSURE=PENDING_EXACT_HEAD_CI_MERGE_AND_POST_MERGE_VALIDATION/,
  )
  assert.match(reconciliation, /GFA_GOV_457=IN_PROGRESS_VERSION_2_DOCUMENTATION_RECONCILIATION/)
  assert.doesNotMatch(reconciliation, /VERSION_2_RECONCILIATION=CLOSED/)
})
