import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

test('Stage 23 keeps ETA Reliability as a real zero-poll frontend consumer', () => {
  const panel = read('components/aircraft/eta-reliability-panel.tsx')
  const projectionPanel = read('components/aircraft/projection-intelligence-panel.tsx')
  const query = read('lib/queries/eta-reliability.ts')
  const api = read('lib/api/eta-reliability.ts')

  assert.match(projectionPanel, /<ETAReliabilityPanel/)
  assert.match(panel, /Historical ETA Reliability/)
  assert.match(panel, /Median ETA error/)
  assert.match(panel, /80% error threshold/)
  assert.match(panel, /Within ±5 min/)
  assert.match(panel, /Within ±10 min/)
  assert.match(panel, /ETA window covered endpoint/)
  assert.match(panel, /observed endpoint proxy, not an official touchdown, gate or schedule timestamp/i)

  assert.match(query, /staleTime: 5 \* 60_000/)
  assert.match(query, /refetchInterval: false/)
  assert.match(query, /refetchOnWindowFocus: false/)
  assert.match(api, /\/eta-reliability/)
})

test('Stage 23 keeps bounded endpoint-proxy semantics in backend and OpenAPI', () => {
  const policy = read('../api/internal/projectionintelligence/etareliability/policy.go')
  const service = read('../api/internal/projectionintelligence/etareliability/service.go')
  const runtime = read('../api/internal/server/eta_reliability_runtime.go')
  const openAPI = JSON.parse(read('../../openapi/openapi.json'))

  assert.match(policy, /MaximumCandidateCount:\s+8/)
  assert.match(service, /trajectory_endpoint_arrival_proxy/)
  assert.match(service, /not an official touchdown, gate or schedule timestamp/i)
  assert.doesNotMatch(runtime, /projectionevaluation/)

  const operation = openAPI.paths?.['/api/v1/trajectories/{id}/eta-reliability']?.get
  assert.equal(operation?.operationId, 'getETAReliabilityByTrajectoryID')
  assert.equal(operation?.security, undefined)
  assert.ok(operation?.responses?.['200'])
  assert.ok(operation?.responses?.['422'])

  const schema = openAPI.components?.schemas?.ETAReliability
  assert.equal(
    schema?.properties?.evidence_class?.const,
    'historically_recomputed_from_persisted_observations_with_endpoint_proxy',
  )
  assert.equal(schema?.properties?.candidate_count?.maximum, 8)
})

test('Stage 23 has dedicated browser evidence and no hidden infrastructure expansion', () => {
  const browser = read('e2e/tests/advanced-intelligence.spec.mjs')
  const mock = read('e2e/mock-api.mjs')
  const document = read('../../docs/214_STAGE_23_ETA_RELIABILITY_INTELLIGENCE.md')

  assert.match(browser, /ETA reliability exposes bounded historical endpoint-proxy evidence/)
  assert.match(browser, /6 eligible \/ 8 checked/)
  assert.match(browser, /Evidence boundary/)
  assert.match(mock, /'eta-reliability'/)
  assert.match(mock, /\/api\/v1\/trajectories\/\{id\}\/eta-reliability/)

  for (const marker of [
    'STAGE_23_ADDITIONAL_COST=0_RUB',
    'STAGE_23_FRONTEND_CONSUMER=YES',
    'STAGE_23_OFFICIAL_ARRIVAL_TRUTH=NONE',
    'STAGE_23_MAX_HISTORICAL_CANDIDATES=8',
    'STAGE_23_FRONTEND_POLLING=NONE',
  ]) {
    assert.ok(document.includes(marker), marker)
  }

  assert.doesNotMatch(document, /official arrival truth is available/i)
})
