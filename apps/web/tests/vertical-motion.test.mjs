import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL('../.test-dist/lib/traffic/vertical-motion.js', import.meta.url)
const {
  buildVerticalMotionEvidence,
  defaultVerticalMotionDeadbandMPS,
  formatVerticalRate,
} = await import(moduleURL.href)

function aircraft(overrides = {}) {
  return {
    icao24: '4b1801',
    callsign: 'AZAL101',
    latitude: 40.4093,
    longitude: 49.8671,
    altitude_m: 10668,
    altitude_status: 'observed',
    altitude_source: 'barometric',
    velocity_mps: 230,
    heading_degrees: 285,
    vertical_rate_mps: 0,
    on_ground: false,
    observed_at: '2026-09-09T12:00:00Z',
    position_observed_at: '2026-09-09T12:00:00Z',
    message_observed_at: '2026-09-09T12:00:03Z',
    position_source: 'adsb',
    source_name: 'adsb.lol',
    aircraft_model: 'Airbus A320',
    airline: 'Azerbaijan Airlines',
    origin_country: 'Azerbaijan',
    ...overrides,
  }
}

test('classifies positive observed vertical rate as climbing', () => {
  const evidence = buildVerticalMotionEvidence(aircraft({ vertical_rate_mps: 4.25 }))
  assert.equal(evidence.status, 'climbing')
  assert.equal(evidence.label, '↑ Climbing')
  assert.equal(evidence.verticalRateMPS, 4.25)
  assert.match(evidence.displayRate, /^\+4\.3 m\/s/)
})

test('classifies negative observed vertical rate as descending', () => {
  const evidence = buildVerticalMotionEvidence(aircraft({ vertical_rate_mps: -3.5 }))
  assert.equal(evidence.status, 'descending')
  assert.equal(evidence.label, '↓ Descending')
  assert.match(evidence.displayRate, /^-3\.5 m\/s/)
})

test('uses a bounded near-level deadband', () => {
  assert.equal(defaultVerticalMotionDeadbandMPS, 0.5)
  assert.equal(buildVerticalMotionEvidence(aircraft({ vertical_rate_mps: 0.5 })).status, 'level')
  assert.equal(buildVerticalMotionEvidence(aircraft({ vertical_rate_mps: -0.5 })).status, 'level')
})

test('keeps unavailable vertical-rate evidence unavailable', () => {
  const evidence = buildVerticalMotionEvidence(aircraft({ vertical_rate_mps: null }))
  assert.equal(evidence.status, 'unavailable')
  assert.equal(evidence.verticalRateMPS, null)
  assert.equal(evidence.displayRate, 'Unavailable')
})

test('ground state takes precedence over a motion classification', () => {
  const evidence = buildVerticalMotionEvidence(
    aircraft({ on_ground: true, vertical_rate_mps: 2 })
  )
  assert.equal(evidence.status, 'ground')
  assert.equal(evidence.label, 'On ground')
})

test('formats feet per minute without changing canonical metres-per-second evidence', () => {
  assert.equal(formatVerticalRate(1), '+1.0 m/s (+197 ft/min)')
  assert.equal(formatVerticalRate(0), '0.0 m/s (0 ft/min)')
})
