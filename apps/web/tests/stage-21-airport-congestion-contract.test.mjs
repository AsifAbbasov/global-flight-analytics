import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

test('Stage 21 frontend uses the production congestion endpoint and existing query layer', () => {
  const api = read('lib/api/airport-intelligence.ts')
  const queries = read('lib/queries/airport-intelligence.ts')
  const workspace = read('components/analytics/unified-airport-analytics-workspace.tsx')

  assert.match(api, /\/intelligence\/congestion/)
  assert.match(api, /getAirportCongestionIntelligence/)
  assert.match(queries, /useAirportCongestionIntelligence/)
  assert.match(workspace, /Activity pressure/)
  assert.match(workspace, /AirportCongestionContent/)
})

test('Stage 21 UI preserves research-only evidence semantics', () => {
  const component = read('components/analytics/airport-congestion-content.tsx')

  assert.match(component, /data-airport-congestion-scope='relative-observed-activity-only'/)
  assert.match(component, /data-airport-capacity-model='none'/)
  assert.match(component, /data-airport-delay-inference='none'/)
  assert.match(component, /not airport capacity/i)
  assert.match(component, /runway occupancy/i)
  assert.match(component, /delay evidence/i)
  assert.match(component, /official operational congestion measure/i)
  assert.doesNotMatch(component, /low congestion|medium congestion|high congestion/i)
})

test('Stage 21 mock API serves congestion as a documented public operation', () => {
  const mock = read('e2e/mock-api.mjs')

  assert.match(mock, /'\/api\/v1\/airports\/\{icao\}\/intelligence\/congestion'/)
  assert.match(mock, /const airportCongestion = \{/)
  assert.match(mock, /return success\(airportCongestion\)/)
  assert.match(mock, /relative_observed_activity_only_not_airport_capacity_delay_queue_slot_or_runway_congestion/)
})
