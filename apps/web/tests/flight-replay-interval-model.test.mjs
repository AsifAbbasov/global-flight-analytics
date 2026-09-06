import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL(
  '../.test-dist/lib/replay/flight-replay-interval-model.js',
  import.meta.url
)
const importedModule = await import(moduleURL.href)
const intervalModel = importedModule.default ?? importedModule
const { buildFlightReplayObservedIntervalComparison } = intervalModel

const replay = {
  trajectory_id: 'trajectory-1',
  flight_id: 'flight-1',
  icao24: '4b1801',
  callsign: 'AZAL101',
  start_time: '2026-08-04T17:45:00Z',
  end_time: '2026-08-04T18:00:00Z',
  evidence_class: 'observed',
  interpolation_policy: 'none',
  points: [
    {
      id: 'state-1',
      flight_id: 'flight-1',
      icao24: '4b1801',
      callsign: 'AZAL101',
      latitude: 40.45,
      longitude: 50.0,
      barometric_altitude_m: 3000,
      barometric_altitude_status: 'observed',
      geometric_altitude_m: null,
      geometric_altitude_status: 'unavailable',
      velocity_mps: 185,
      heading_degrees: 350,
      vertical_rate_mps: 7.2,
      on_ground: false,
      origin_country: 'Azerbaijan',
      observed_at: '2026-08-04T17:45:00Z',
      source_name: 'fixture',
    },
    {
      id: 'state-2',
      flight_id: 'flight-1',
      icao24: '4b1801',
      callsign: 'AZAL101',
      latitude: 40.43,
      longitude: 49.94,
      barometric_altitude_m: 6000,
      barometric_altitude_status: 'observed',
      geometric_altitude_m: null,
      geometric_altitude_status: 'unavailable',
      velocity_mps: 215,
      heading_degrees: 5,
      vertical_rate_mps: 4.1,
      on_ground: false,
      origin_country: 'Azerbaijan',
      observed_at: '2026-08-04T17:52:30Z',
      source_name: 'fixture',
    },
    {
      id: 'state-3',
      flight_id: 'flight-1',
      icao24: '4b1801',
      callsign: 'AZAL101',
      latitude: 40.4093,
      longitude: 49.8671,
      barometric_altitude_m: 10668,
      barometric_altitude_status: 'observed',
      geometric_altitude_m: null,
      geometric_altitude_status: 'unavailable',
      velocity_mps: 230,
      heading_degrees: 10,
      vertical_rate_mps: 0,
      on_ground: false,
      origin_country: 'Azerbaijan',
      observed_at: '2026-08-04T18:00:00Z',
      source_name: 'fixture',
    },
  ],
}

test('interval comparison uses only selected persisted endpoints and reports interval quality', () => {
  const comparison = buildFlightReplayObservedIntervalComparison(replay, 0, 2)

  assert.ok(comparison)
  assert.equal(comparison.startIndex, 0)
  assert.equal(comparison.endIndex, 2)
  assert.equal(comparison.startPointID, 'state-1')
  assert.equal(comparison.endPointID, 'state-3')
  assert.equal(comparison.elapsedSeconds, 900)
  assert.equal(comparison.observedSampleCount, 3)
  assert.equal(comparison.intermediateSampleCount, 1)
  assert.equal(comparison.largestGapSeconds, 450)
  assert.ok(comparison.endpointGreatCircleDisplacementM > 10000)
  assert.equal(comparison.altitudeDeltaM, 7668)
  assert.equal(comparison.velocityDeltaMPS, 45)
  assert.equal(comparison.verticalRateDeltaMPS, -7.2)
  assert.equal(comparison.headingChangeDegrees, 20)
  assert.equal(comparison.startOnGround, false)
  assert.equal(comparison.endOnGround, false)
})

test('interval comparison normalizes reversed selections chronologically', () => {
  const comparison = buildFlightReplayObservedIntervalComparison(replay, 2, 0)

  assert.ok(comparison)
  assert.equal(comparison.startIndex, 0)
  assert.equal(comparison.endIndex, 2)
  assert.equal(comparison.startPointID, 'state-1')
  assert.equal(comparison.endPointID, 'state-3')
  assert.equal(comparison.elapsedSeconds, 900)
})

test('interval comparison isolates the largest gap inside the selected interval', () => {
  const unevenReplay = {
    ...replay,
    points: [
      replay.points[0],
      { ...replay.points[1], observed_at: '2026-08-04T17:46:00Z' },
      { ...replay.points[2], observed_at: '2026-08-04T17:56:00Z' },
    ],
  }

  const comparison = buildFlightReplayObservedIntervalComparison(unevenReplay, 0, 2)
  assert.ok(comparison)
  assert.equal(comparison.elapsedSeconds, 660)
  assert.equal(comparison.largestGapSeconds, 600)
})

test('interval comparison does not invent altitude evidence', () => {
  const unavailableAltitudeReplay = {
    ...replay,
    points: replay.points.map((point, index) =>
      index === 2
        ? {
            ...point,
            barometric_altitude_m: null,
            barometric_altitude_status: 'unavailable',
          }
        : point
    ),
  }

  const comparison = buildFlightReplayObservedIntervalComparison(
    unavailableAltitudeReplay,
    0,
    2
  )
  assert.ok(comparison)
  assert.equal(comparison.altitudeDeltaM, null)
})

test('interval comparison requires two different persisted observations', () => {
  assert.equal(buildFlightReplayObservedIntervalComparison(undefined, 0, 1), null)
  assert.equal(buildFlightReplayObservedIntervalComparison(replay, 1, 1), null)
  assert.equal(
    buildFlightReplayObservedIntervalComparison(
      { ...replay, points: [replay.points[0]] },
      0,
      0
    ),
    null
  )
})
