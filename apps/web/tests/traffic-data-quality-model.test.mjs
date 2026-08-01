// FRONTEND_TRAFFIC_DATA_QUALITY_LENS_V1
import assert from 'node:assert/strict'
import test from 'node:test'

const modelModuleURL = new URL(
  '../.test-dist/lib/traffic/traffic-data-quality-model.js',
  import.meta.url
)
const importedModelModule = await import(modelModuleURL.href)
const modelModule = importedModelModule.default ?? importedModelModule
const { buildTrafficDataQualityModel } = modelModule

const referenceTime = Date.parse('2026-08-01T12:00:00Z')

function aircraft(overrides = {}) {
  return {
    icao24: 'abc123',
    callsign: 'AZAL101',
    latitude: 40.4,
    longitude: 49.9,
    altitude_m: 10000,
    altitude_status: 'observed',
    altitude_source: 'geometric',
    velocity_mps: 220,
    heading_degrees: 90,
    on_ground: false,
    observed_at: '2026-08-01T11:58:00Z',
    aircraft_model: 'Airbus A320',
    airline: 'Azerbaijan Airlines',
    origin_country: 'Azerbaijan',
    ...overrides,
  }
}

test('empty snapshots produce stable zero-valued quality sections', () => {
  const model = buildTrafficDataQualityModel([], referenceTime)

  assert.equal(model.totalCount, 0)
  assert.equal(model.coreUsableShare, 0)
  assert.equal(model.usableAirborneAltitudeShare, 0)
  assert.equal(model.identityCompleteness.length, 4)
  assert.deepEqual(model.issues, [])
})

test('complete recent records remain structurally usable without issues', () => {
  const model = buildTrafficDataQualityModel([aircraft()], referenceTime)

  assert.equal(model.coreUsableCount, 1)
  assert.equal(model.uniqueAircraftCount, 1)
  assert.equal(model.validCoordinateCount, 1)
  assert.equal(model.validMotionCount, 1)
  assert.equal(model.recentObservationCount, 1)
  assert.equal(model.issues.length, 0)
})

test('invalid coordinates, motion and timestamps are counted independently', () => {
  const model = buildTrafficDataQualityModel(
    [
      aircraft({
        latitude: 95,
        longitude: Number.NaN,
        velocity_mps: -1,
        heading_degrees: 360,
        observed_at: 'not-a-date',
      }),
    ],
    referenceTime
  )

  assert.equal(model.coreUsableCount, 0)
  assert.deepEqual(
    model.issues.slice(0, 3).map(issue => issue.key),
    ['invalid-coordinate', 'invalid-motion', 'invalid-observation-time']
  )
})

test('duplicate ICAO24 identifiers normalize case and surrounding spaces', () => {
  const model = buildTrafficDataQualityModel(
    [
      aircraft({ icao24: 'ABC123' }),
      aircraft({ icao24: ' abc123 ', callsign: 'SECOND' }),
      aircraft({ icao24: 'def456', callsign: 'THIRD' }),
    ],
    referenceTime
  )

  assert.equal(model.validIdentifierCount, 3)
  assert.equal(model.uniqueAircraftCount, 2)
  assert.equal(model.duplicateIdentifierRecordCount, 1)
  assert.equal(
    model.issues.find(issue => issue.key === 'duplicate-identifier')?.count,
    1
  )
})

test('airborne altitude completeness excludes ground records from its denominator', () => {
  const model = buildTrafficDataQualityModel(
    [
      aircraft({ on_ground: true, altitude_m: null, altitude_status: 'ground' }),
      aircraft({ icao24: 'def456', altitude_m: null, altitude_status: 'unknown' }),
      aircraft({ icao24: 'fed654', altitude_m: 9200 }),
    ],
    referenceTime
  )

  assert.equal(model.airborneCount, 2)
  assert.equal(model.usableAirborneAltitudeCount, 1)
  assert.equal(model.usableAirborneAltitudeShare, 0.5)
  assert.equal(
    model.issues.find(issue => issue.key === 'missing-airborne-altitude')
      ?.denominator,
    2
  )
})

test('identity completeness trims values and reports missing fields', () => {
  const model = buildTrafficDataQualityModel(
    [
      aircraft({ callsign: '   ', airline: '', origin_country: '  ' }),
      aircraft({ icao24: 'def456', aircraft_model: '' }),
    ],
    referenceTime
  )

  const completeness = Object.fromEntries(
    model.identityCompleteness.map(metric => [metric.key, metric])
  )
  assert.equal(completeness.callsign.presentCount, 1)
  assert.equal(completeness.airline.presentCount, 1)
  assert.equal(completeness['origin-country'].presentCount, 1)
  assert.equal(completeness['aircraft-model'].presentCount, 1)
})

test('recency boundaries and issue ordering remain deterministic', () => {
  const model = buildTrafficDataQualityModel(
    [
      aircraft({
        icao24: 'bad-id',
        observed_at: '2026-08-01T11:54:59Z',
        callsign: '',
      }),
      aircraft({
        icao24: 'def456',
        observed_at: '2026-08-01T12:01:01Z',
        airline: '',
      }),
      aircraft({
        icao24: 'fed654',
        observed_at: '2026-08-01T11:55:00Z',
      }),
    ],
    referenceTime
  )

  assert.equal(model.recentObservationCount, 1)
  assert.equal(model.staleObservationCount, 1)
  assert.equal(model.futureObservationCount, 1)
  assert.deepEqual(
    model.issues.map(issue => issue.key),
    [
      'invalid-identifier',
      'future-observation',
      'stale-observation',
      'missing-airline',
      'missing-callsign',
    ]
  )
})
