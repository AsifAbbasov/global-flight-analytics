import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

function source(relativePath) {
  return readFileSync(new URL(`../${relativePath}`, import.meta.url), 'utf8')
}

test('flight replay reuses persisted flight states inside the selected trajectory window', () => {
  const api = source('lib/api/flight-replay.ts')

  assert.match(api, /\/api\/v1\/flights\/\$\{encodeURIComponent\(flightID\)\}\/states/)
  assert.match(api, /point\.flight_id !== expectedFlightID/)
  assert.match(api, /point\.icao24 !== expectedICAO24/)
  assert.match(api, /observedAt < startTime \|\| observedAt > endTime/)
  assert.match(api, /evidence_class: 'observed'/)
  assert.match(api, /interpolation_policy: 'none'/)
  assert.doesNotMatch(api, /interpolat(e|ion)\(/i)
})

test('flight replay validates telemetry already owned by the persisted flight-state contract', () => {
  const api = source('lib/api/flight-replay.ts')
  const types = source('types/flight-replay.ts')

  for (const field of [
    'velocity_mps',
    'heading_degrees',
    'vertical_rate_mps',
    'on_ground',
    'origin_country',
  ]) {
    assert.match(api, new RegExp(`record\\.${field}`))
    assert.match(types, new RegExp(`${field}:`))
  }

  assert.match(api, /requireNonNegativeNumber/)
  assert.match(api, /requireHeadingDegrees/)
  assert.match(api, /requireFiniteNumber/)
  assert.match(api, /requireBoolean/)
})

test('replay controls disclose observed-only evidence, gap semantics and no interpolation', () => {
  const control = source('components/aircraft/flight-replay-control.tsx')

  assert.match(control, /Historical flight replay/)
  assert.match(control, /Observed only/)
  assert.match(control, /No interpolation/)
  assert.match(control, /Evidence gaps/)
  assert.match(control, /Previous unobserved interval/)
  assert.match(control, /Next unobserved interval/)
  assert.match(control, /No persisted intermediate position/)
  assert.match(control, /Gaps are not filled/)
  assert.match(control, /positions between observations are not synthesized/)
  assert.match(control, /aria-label='Historical replay position'/)
  assert.match(control, /aria-label='Historical evidence gap timeline'/)
})

test('replay controls surface persisted telemetry without new external providers', () => {
  const control = source('components/aircraft/flight-replay-control.tsx')

  assert.match(control, /label='Velocity'/)
  assert.match(control, /label='Heading'/)
  assert.match(control, /label='Vertical rate'/)
  assert.match(control, /label='Flight state'/)
  assert.match(control, /label='Aircraft origin country'/)
  assert.match(control, /label='Source'/)
  assert.match(control, /currentPoint\.velocity_mps/)
  assert.match(control, /currentPoint\.heading_degrees/)
  assert.match(control, /currentPoint\.vertical_rate_mps/)
  assert.match(control, /currentPoint\.on_ground/)
  assert.match(control, /currentPoint\.origin_country/)
})

test('replay analytics remain evidence-derived and avoid inferred intermediate values', () => {
  const model = source('lib/replay/flight-replay-model.ts')
  const control = source('components/aircraft/flight-replay-control.tsx')

  assert.match(model, /buildFlightReplayAnalyticsSummary/)
  assert.match(model, /altitudeCoveragePercent/)
  assert.match(model, /medianGapSeconds/)
  assert.match(model, /peakVelocityMPS/)
  assert.match(model, /maxClimbRateMPS/)
  assert.match(model, /steepestDescentRateMPS/)
  assert.match(control, /Observed replay analytics/)
  assert.match(control, /Aggregates use persisted samples only/)
  assert.match(control, /No values are inferred between observations/)
  assert.match(control, /data-flight-replay-analytics='observed-samples-only'/)
  assert.doesNotMatch(model, /predict|forecast|interpolat/i)
})

test('replay observation sharing uses exact persisted state ids without new storage', () => {
  const model = source('lib/replay/flight-replay-model.ts')
  const workspace = source('components/map/map-evidence-workspace.tsx')
  const control = source('components/aircraft/flight-replay-control.tsx')

  assert.match(model, /flightReplayObservationParameter = 'replay_observation'/)
  assert.match(model, /resolveFlightReplayCursorFromSearch/)
  assert.match(model, /buildFlightReplayObservationShareURL/)
  assert.match(model, /point\.id === requestedPointID/)
  assert.match(workspace, /resolveFlightReplayCursorFromSearch/)
  assert.match(workspace, /window\.location\.search/)
  assert.match(control, /Copy observation link/)
  assert.match(control, /Copy replay observation link/)
  assert.match(control, /navigator\.clipboard/)
  assert.match(control, /data-replay-observation-id/)
  assert.match(control, /exact\s+persisted observation identifier/)
})

test('map renders replay as discrete point features instead of a replay line', () => {
  const map = source('components/map/traffic-map.tsx')

  assert.match(map, /selected-aircraft-flight-replay-samples/)
  assert.match(map, /selected-aircraft-flight-replay-current/)
  assert.match(map, /kind: 'sample'/)
  assert.match(map, /kind: 'current'/)
  assert.match(map, /type: 'Point'/)
  assert.doesNotMatch(
    map,
    /replaySourceID[\s\S]{0,600}type: 'line'/
  )
})

test('map workspace owns replay timing without changing trajectory or projection evidence toggles', () => {
  const workspace = source('components/map/map-evidence-workspace.tsx')

  assert.match(workspace, /useTrajectoryFlightReplay/)
  assert.match(workspace, /advanceFlightReplayCursor/)
  assert.match(workspace, /buildFlightReplayFrame/)
  assert.match(workspace, /<FlightReplayControl/)
  assert.match(workspace, /replayPoint=\{replayFrame\.point \?\? undefined\}/)
  assert.match(workspace, /replayTrail=\{replayFrame\.trailPoints\}/)
})
