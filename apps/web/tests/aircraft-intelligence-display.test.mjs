import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL(
  '../.test-dist/lib/aircraft/aircraft-intelligence-display.js',
  import.meta.url
)
const importedModule = await import(moduleURL.href)
const displayModule = importedModule.default ?? importedModule
const { buildAircraftIntelligenceDisplay } = displayModule

function aircraft(overrides = {}) {
  return {
    icao24: 'abc123',
    callsign: 'AZAL101',
    latitude: 40.4093,
    longitude: 49.8671,
    altitude_m: 10000,
    altitude_status: 'observed',
    altitude_source: 'geometric',
    velocity_mps: 220,
    heading_degrees: 90,
    on_ground: false,
    observed_at: '2026-08-23T07:00:00Z',
    aircraft_model: 'A320',
    airline: 'Azerbaijan Airlines',
    origin_country: 'Azerbaijan',
    ...overrides,
  }
}

function profile(overrides = {}) {
  return {
    icao24: 'abc123',
    registration: '4K-AZ01',
    model: 'A320-214',
    manufacturer: 'Airbus',
    aircraft_type: 'A320',
    airline: 'Azerbaijan Airlines',
    country: 'Azerbaijan',
    ...overrides,
  }
}

test('prefers observed callsign then registration then ICAO24 for title', () => {
  assert.equal(
    buildAircraftIntelligenceDisplay('abc123', aircraft(), profile()).title,
    'AZAL101'
  )
  assert.equal(
    buildAircraftIntelligenceDisplay(
      'abc123',
      aircraft({ callsign: '   ' }),
      profile()
    ).title,
    '4K-AZ01'
  )
  assert.equal(
    buildAircraftIntelligenceDisplay('abc123', undefined, undefined).title,
    'ABC123'
  )
})

test('omits unavailable profile fields instead of fabricating placeholders', () => {
  const result = buildAircraftIntelligenceDisplay(
    'abc123',
    aircraft(),
    profile({ manufacturer: '', country: '   ' })
  )

  assert.equal(
    result.profileFields.some(field => field.key === 'manufacturer'),
    false
  )
  assert.equal(
    result.profileFields.some(field => field.key === 'registration-country'),
    false
  )
  assert.equal(result.profileFields.every(field => field.value !== 'Unknown'), true)
})

test('labels evidence source for observed and profile-backed fields', () => {
  const result = buildAircraftIntelligenceDisplay('abc123', aircraft(), profile())

  assert.equal(result.observedFields.every(field => field.evidence === 'observed'), true)
  assert.equal(result.profileFields.every(field => field.evidence === 'profile'), true)
})

test('preserves observed altitude source and omits unsupported altitude evidence', () => {
  const observed = buildAircraftIntelligenceDisplay('abc123', aircraft(), undefined)
  assert.equal(
    observed.observedFields.find(field => field.key === 'altitude')?.value,
    '10,000 m (geometric)'
  )

  const ground = buildAircraftIntelligenceDisplay(
    'abc123',
    aircraft({ altitude_m: null, altitude_status: 'ground', altitude_source: 'ground' }),
    undefined
  )
  assert.equal(
    ground.observedFields.find(field => field.key === 'altitude')?.value,
    'Ground (0 m)'
  )

  const unavailable = buildAircraftIntelligenceDisplay(
    'abc123',
    aircraft({ altitude_m: null, altitude_status: 'unavailable', altitude_source: 'none' }),
    undefined
  )
  assert.equal(
    unavailable.observedFields.some(field => field.key === 'altitude'),
    false
  )
})

test('omits invalid speed heading position and timestamp', () => {
  const result = buildAircraftIntelligenceDisplay(
    'abc123',
    aircraft({
      velocity_mps: Number.NaN,
      heading_degrees: Number.NaN,
      latitude: 120,
      observed_at: 'not-a-date',
    }),
    undefined
  )

  const keys = result.observedFields.map(field => field.key)
  assert.equal(keys.includes('speed'), false)
  assert.equal(keys.includes('heading'), false)
  assert.equal(keys.includes('coordinates'), false)
  assert.equal(keys.includes('observed-at'), false)
})
