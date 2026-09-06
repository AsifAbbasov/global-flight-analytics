import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL(
  '../.test-dist/lib/replay/flight-replay-evidence-quality-model.js',
  import.meta.url
)
const importedModule = await import(moduleURL.href)
const evidenceQualityModel = importedModule.default ?? importedModule
const { buildFlightReplayEvidenceQualityProfile } = evidenceQualityModel

const basePoint = {
  flight_id: 'flight-1',
  icao24: '4b1801',
  callsign: 'AZAL101',
  latitude: 40.45,
  longitude: 50,
  geometric_altitude_m: null,
  geometric_altitude_status: 'unavailable',
  velocity_mps: 200,
  heading_degrees: 90,
  vertical_rate_mps: 0,
  on_ground: false,
  origin_country: 'Azerbaijan',
  source_name: 'fixture',
}

const replay = {
  trajectory_id: 'trajectory-1',
  flight_id: 'flight-1',
  icao24: '4b1801',
  callsign: 'AZAL101',
  start_time: '2026-08-04T17:45:00Z',
  end_time: '2026-08-04T17:56:00Z',
  evidence_class: 'observed',
  interpolation_policy: 'none',
  points: [
    {
      ...basePoint,
      id: 'state-1',
      barometric_altitude_m: 3000,
      barometric_altitude_status: 'observed',
      observed_at: '2026-08-04T17:45:00Z',
    },
    {
      ...basePoint,
      id: 'state-2',
      barometric_altitude_m: null,
      barometric_altitude_status: 'unavailable',
      observed_at: '2026-08-04T17:46:00Z',
    },
    {
      ...basePoint,
      id: 'state-3',
      barometric_altitude_m: 7000,
      barometric_altitude_status: 'observed',
      observed_at: '2026-08-04T17:56:00Z',
    },
  ],
}

test('evidence quality profile describes persisted sampling without assigning a score', () => {
  const profile = buildFlightReplayEvidenceQualityProfile(replay)

  assert.equal(profile.sampleCount, 3)
  assert.equal(profile.intervalCount, 2)
  assert.equal(profile.observedSpanSeconds, 660)
  assert.equal(profile.meanGapSeconds, 330)
  assert.equal(profile.medianGapSeconds, 330)
  assert.equal(profile.p90GapSeconds, 600)
  assert.equal(profile.largestGapSeconds, 600)
  assert.equal(profile.largestGapSharePercent, 91)
  assert.ok(profile.observationDensityPerHour)
  assert.ok(Math.abs(profile.observationDensityPerHour - 16.3636363636) < 0.000001)
  assert.equal(profile.altitudeCoveragePercent, 67)
  assert.equal(profile.largestGapStartObservedAt, '2026-08-04T17:46:00Z')
  assert.equal(profile.largestGapEndObservedAt, '2026-08-04T17:56:00Z')
  assert.equal(Object.hasOwn(profile, 'score'), false)
  assert.equal(Object.hasOwn(profile, 'qualityScore'), false)
})

test('evidence quality profile keeps single-sample uncertainty explicit', () => {
  const singleSample = buildFlightReplayEvidenceQualityProfile({
    ...replay,
    points: [replay.points[0]],
  })

  assert.equal(singleSample.sampleCount, 1)
  assert.equal(singleSample.intervalCount, 0)
  assert.equal(singleSample.observedSpanSeconds, 0)
  assert.equal(singleSample.meanGapSeconds, null)
  assert.equal(singleSample.medianGapSeconds, null)
  assert.equal(singleSample.p90GapSeconds, null)
  assert.equal(singleSample.largestGapSeconds, null)
  assert.equal(singleSample.largestGapSharePercent, null)
  assert.equal(singleSample.observationDensityPerHour, null)
  assert.equal(singleSample.altitudeCoveragePercent, 100)
  assert.equal(singleSample.largestGapStartObservedAt, null)
  assert.equal(singleSample.largestGapEndObservedAt, null)
})

test('evidence quality profile handles missing replay without synthetic defaults', () => {
  assert.deepEqual(buildFlightReplayEvidenceQualityProfile(undefined), {
    sampleCount: 0,
    intervalCount: 0,
    observedSpanSeconds: 0,
    meanGapSeconds: null,
    medianGapSeconds: null,
    p90GapSeconds: null,
    largestGapSeconds: null,
    largestGapSharePercent: null,
    observationDensityPerHour: null,
    altitudeCoveragePercent: 0,
    largestGapStartObservedAt: null,
    largestGapEndObservedAt: null,
  })
})