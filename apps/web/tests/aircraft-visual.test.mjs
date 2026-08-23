import assert from 'node:assert/strict'
import test from 'node:test'

const visualModuleURL = new URL(
  '../.test-dist/lib/map/aircraft-visual.js',
  import.meta.url
)
const importedVisualModule = await import(visualModuleURL.href)
const visualModule = importedVisualModule.default ?? importedVisualModule
const { buildAircraftVisualState, normalizeAircraftHeading } = visualModule

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

test('builds an airborne selected visual state from truthful traffic fields', () => {
  const result = buildAircraftVisualState(aircraft(), 'ABC123')

  assert.deepEqual(result, {
    key: 'abc123',
    label: 'AZAL101',
    headingDegrees: 90,
    motionState: 'airborne',
    isSelected: true,
    latitude: 40.4,
    longitude: 49.9,
  })
})

test('falls back to ICAO24 label and preserves ground state', () => {
  const result = buildAircraftVisualState(
    aircraft({ callsign: '   ', on_ground: true }),
    null
  )

  assert.equal(result?.label, 'abc123')
  assert.equal(result?.motionState, 'ground')
  assert.equal(result?.isSelected, false)
})

test('normalizes headings and rejects non-finite heading values safely', () => {
  assert.equal(normalizeAircraftHeading(450), 90)
  assert.equal(normalizeAircraftHeading(-90), 270)
  assert.equal(normalizeAircraftHeading(Number.NaN), 0)
})

test('rejects invalid coordinates and blank ICAO24 keys', () => {
  assert.equal(
    buildAircraftVisualState(aircraft({ latitude: 91 }), null),
    null
  )
  assert.equal(
    buildAircraftVisualState(aircraft({ longitude: -181 }), null),
    null
  )
  assert.equal(
    buildAircraftVisualState(aircraft({ icao24: '   ' }), null),
    null
  )
})
