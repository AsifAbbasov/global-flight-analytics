import assert from 'node:assert/strict'
import test from 'node:test'

const modelModuleURL = new URL(
  '../.test-dist/lib/traffic/regional-traffic-brief-model.js',
  import.meta.url
)
const importedModelModule = await import(modelModuleURL.href)
const modelModule = importedModelModule.default ?? importedModelModule
const { buildRegionalTrafficBrief } = modelModule

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
    observed_at: '2026-08-01T12:00:00Z',
    aircraft_model: 'Airbus A320',
    airline: 'Azerbaijan Airlines',
    origin_country: 'Azerbaijan',
    ...overrides,
  }
}

test('empty snapshots produce stable zero-valued brief sections', () => {
  const result = buildRegionalTrafficBrief([])

  assert.equal(result.totalCount, 0)
  assert.equal(result.airborneCount, 0)
  assert.equal(result.groundCount, 0)
  assert.equal(result.knownAltitudeCount, 0)
  assert.equal(result.altitudeCoverage, 0)
  assert.deepEqual(
    result.altitudeBands.map(item => item.count),
    [0, 0, 0, 0]
  )
  assert.deepEqual(result.topAirlines, [])
  assert.deepEqual(result.topOriginCountries, [])
})

test('altitude bands use explicit boundaries and airborne denominators', () => {
  const result = buildRegionalTrafficBrief([
    aircraft({ icao24: 'ground', on_ground: true, altitude_m: 0 }),
    aircraft({ icao24: 'low', altitude_m: 2999.9 }),
    aircraft({ icao24: 'medium-start', altitude_m: 3000 }),
    aircraft({ icao24: 'medium-end', altitude_m: 8999.9 }),
    aircraft({ icao24: 'high', altitude_m: 9000 }),
    aircraft({ icao24: 'unknown', altitude_m: null }),
  ])

  assert.equal(result.totalCount, 6)
  assert.equal(result.airborneCount, 5)
  assert.equal(result.groundCount, 1)
  assert.equal(result.knownAltitudeCount, 4)
  assert.equal(result.altitudeCoverage, 0.8)
  assert.deepEqual(
    result.altitudeBands.map(item => [item.key, item.count, item.share]),
    [
      ['low', 1, 0.2],
      ['medium', 2, 0.4],
      ['high', 1, 0.2],
      ['unknown', 1, 0.2],
    ]
  )
})

test('negative and non-finite airborne altitudes remain unknown', () => {
  const result = buildRegionalTrafficBrief([
    aircraft({ icao24: 'negative', altitude_m: -10 }),
    aircraft({ icao24: 'nan', altitude_m: Number.NaN }),
    aircraft({ icao24: 'ground', on_ground: true, altitude_m: 12000 }),
  ])

  assert.equal(result.airborneCount, 2)
  assert.equal(result.knownAltitudeCount, 0)
  assert.equal(result.altitudeCoverage, 0)
  assert.equal(
    result.altitudeBands.find(item => item.key === 'unknown').count,
    2
  )
  assert.equal(
    result.altitudeBands.find(item => item.key === 'high').count,
    0
  )
})

test('rankings normalize labels, group case variants and break ties alphabetically', () => {
  const result = buildRegionalTrafficBrief(
    [
      aircraft({ airline: ' Azerbaijan   Airlines ', origin_country: 'Azerbaijan' }),
      aircraft({ airline: 'azerbaijan airlines', origin_country: 'azerbaijan' }),
      aircraft({ airline: 'Turkish Airlines', origin_country: 'Türkiye' }),
      aircraft({ airline: 'Air Baltic', origin_country: 'Latvia' }),
      aircraft({ airline: '   ', origin_country: '' }),
    ],
    { rankingLimit: 2 }
  )

  assert.deepEqual(
    result.topAirlines.map(item => [item.label, item.count, item.share]),
    [
      ['Azerbaijan Airlines', 2, 0.4],
      ['Air Baltic', 1, 0.2],
    ]
  )
  assert.deepEqual(
    result.topOriginCountries.map(item => [item.label, item.count]),
    [
      ['Azerbaijan', 2],
      ['Latvia', 1],
    ]
  )
  assert.equal(result.unknownAirlineCount, 1)
  assert.equal(result.unknownOriginCountryCount, 1)
})

test('ranking limits are bounded to five entries', () => {
  const source = ['F', 'E', 'D', 'C', 'B', 'A'].map((label, index) =>
    aircraft({
      icao24: `aircraft-${index}`,
      airline: label,
      origin_country: label,
    })
  )

  const result = buildRegionalTrafficBrief(source, { rankingLimit: 99 })

  assert.deepEqual(
    result.topAirlines.map(item => item.label),
    ['A', 'B', 'C', 'D', 'E']
  )
  assert.equal(result.topAirlines.length, 5)
  assert.equal(result.topOriginCountries.length, 5)
})
