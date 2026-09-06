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

test('replay controls disclose observed-only evidence and no interpolation', () => {
  const control = source('components/aircraft/flight-replay-control.tsx')

  assert.match(control, /Historical flight replay/)
  assert.match(control, /Observed only/)
  assert.match(control, /No interpolation/)
  assert.match(control, /Gaps are not filled/)
  assert.match(control, /positions between observations are not synthesized/)
  assert.match(control, /aria-label='Historical replay position'/)
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
