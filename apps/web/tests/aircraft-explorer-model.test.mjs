import assert from 'node:assert/strict'
import test from 'node:test'

const modelModuleURL = new URL(
  '../.test-dist/lib/traffic/aircraft-explorer-model.js',
  import.meta.url
)
const importedModelModule = await import(modelModuleURL.href)
const modelModule = importedModelModule.default ?? importedModelModule
const { buildAircraftExplorerModel } = modelModule

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

test('search matches callsign, ICAO24, model, airline and country', () => {
  const source = [
    aircraft(),
    aircraft({
      icao24: 'def456',
      callsign: 'THY7',
      aircraft_model: 'Boeing 737',
      airline: 'Turkish Airlines',
      origin_country: 'Türkiye',
    }),
  ]

  for (const query of ['thy7', 'DEF456', 'boeing', 'turkish', 'türkiye']) {
    const result = buildAircraftExplorerModel(source, { query })
    assert.equal(result.matchedCount, 1)
    assert.equal(result.items[0].icao24, 'def456')
  }
})

test('altitude sorting places highest observed altitude first and null last', () => {
  const result = buildAircraftExplorerModel(
    [
      aircraft({ icao24: 'low', altitude_m: 1200 }),
      aircraft({ icao24: 'unknown', altitude_m: null }),
      aircraft({ icao24: 'high', altitude_m: 11200 }),
    ],
    { sort: 'altitude-descending' }
  )

  assert.deepEqual(
    result.items.map(item => item.icao24),
    ['high', 'low', 'unknown']
  )
})

test('speed sorting is descending with deterministic ICAO24 tie-breaking', () => {
  const result = buildAircraftExplorerModel(
    [
      aircraft({ icao24: 'z-last', velocity_mps: 250 }),
      aircraft({ icao24: 'a-first', velocity_mps: 250 }),
      aircraft({ icao24: 'slow', velocity_mps: 100 }),
    ],
    { sort: 'speed-descending' }
  )

  assert.deepEqual(
    result.items.map(item => item.icao24),
    ['a-first', 'z-last', 'slow']
  )
})

test('recent sorting puts invalid timestamps after valid observations', () => {
  const result = buildAircraftExplorerModel(
    [
      aircraft({ icao24: 'old', observed_at: '2026-08-01T10:00:00Z' }),
      aircraft({ icao24: 'invalid', observed_at: 'not-a-date' }),
      aircraft({ icao24: 'new', observed_at: '2026-08-01T12:00:00Z' }),
    ],
    { sort: 'recent' }
  )

  assert.deepEqual(
    result.items.map(item => item.icao24),
    ['new', 'old', 'invalid']
  )
})

test('summary counts describe matched aircraft while limit controls display only', () => {
  const result = buildAircraftExplorerModel(
    [
      aircraft({ icao24: 'air-known' }),
      aircraft({ icao24: 'air-unknown', altitude_m: null }),
      aircraft({ icao24: 'ground', on_ground: true, altitude_m: null }),
    ],
    { limit: 2, sort: 'callsign' }
  )

  assert.equal(result.totalCount, 3)
  assert.equal(result.matchedCount, 3)
  assert.equal(result.displayedCount, 2)
  assert.equal(result.items.length, 2)
  assert.equal(result.airborneCount, 2)
  assert.equal(result.groundCount, 1)
  assert.equal(result.unknownAltitudeCount, 2)
})
