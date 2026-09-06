import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL(
  '../.test-dist/lib/replay/flight-replay-model.js',
  import.meta.url
)
const importedModule = await import(moduleURL.href)
const replayModel = importedModule.default ?? importedModule
const {
  advanceFlightReplayCursor,
  buildFlightReplayFrame,
  buildFlightReplayGaps,
  buildFlightReplayGapSummary,
  clampFlightReplayCursorIndex,
} = replayModel

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
      heading_degrees: 240,
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
      heading_degrees: 265,
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
      heading_degrees: 285,
      vertical_rate_mps: 0,
      on_ground: false,
      origin_country: 'Azerbaijan',
      observed_at: '2026-08-04T18:00:00Z',
      source_name: 'fixture',
    },
  ],
}

test('replay cursor clamps to persisted sample bounds', () => {
  assert.equal(clampFlightReplayCursorIndex(-4, 3), 0)
  assert.equal(clampFlightReplayCursorIndex(99, 3), 2)
  assert.equal(clampFlightReplayCursorIndex(1.9, 3), 1)
  assert.equal(clampFlightReplayCursorIndex(Number.NaN, 3), 0)
  assert.equal(clampFlightReplayCursorIndex(1, 0), 0)
})

test('replay frame exposes only the observed prefix through the current sample', () => {
  const frame = buildFlightReplayFrame(replay, 1)

  assert.equal(frame.cursorIndex, 1)
  assert.equal(frame.point.id, 'state-2')
  assert.deepEqual(
    frame.trailPoints.map(point => point.id),
    ['state-1', 'state-2']
  )
  assert.equal(frame.progress, 0.5)
  assert.equal(replay.points.length, 3)
})

test('replay gaps preserve exact elapsed time between persisted observations', () => {
  const gaps = buildFlightReplayGaps(replay)

  assert.deepEqual(
    gaps.map(gap => ({
      fromIndex: gap.fromIndex,
      toIndex: gap.toIndex,
      durationSeconds: gap.durationSeconds,
    })),
    [
      { fromIndex: 0, toIndex: 1, durationSeconds: 450 },
      { fromIndex: 1, toIndex: 2, durationSeconds: 450 },
    ]
  )
})

test('replay gap summary describes the current observed sample without filling gaps', () => {
  assert.deepEqual(buildFlightReplayGapSummary(replay, 1), {
    gaps: buildFlightReplayGaps(replay),
    previousGapSeconds: 450,
    nextGapSeconds: 450,
    largestGapSeconds: 450,
    totalObservedSpanSeconds: 900,
  })

  const finalSummary = buildFlightReplayGapSummary(replay, 2)
  assert.equal(finalSummary.previousGapSeconds, 450)
  assert.equal(finalSummary.nextGapSeconds, null)
  assert.equal(finalSummary.totalObservedSpanSeconds, 900)
})

test('replay cursor advances discretely and stops at the final observed sample', () => {
  assert.deepEqual(advanceFlightReplayCursor(0, 3), {
    cursorIndex: 1,
    completed: false,
  })
  assert.deepEqual(advanceFlightReplayCursor(1, 3), {
    cursorIndex: 2,
    completed: true,
  })
  assert.deepEqual(advanceFlightReplayCursor(2, 3), {
    cursorIndex: 2,
    completed: true,
  })
})
