import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL(
  '../.test-dist/lib/traffic/position-provenance.js',
  import.meta.url
)
const { buildPositionProvenanceEvidence } = await import(moduleURL.href)

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
    on_ground: false,
    observed_at: '2026-09-09T12:00:00Z',
    position_observed_at: '2026-09-09T12:00:00Z',
    message_observed_at: '2026-09-09T12:00:08Z',
    aircraft_model: 'Airbus A320',
    airline: 'Azerbaijan Airlines',
    origin_country: 'Azerbaijan',
    ...overrides,
  }
}

test('keeps data feed separate from the observed position method', () => {
  const evidence = buildPositionProvenanceEvidence(
    aircraft({ position_source: 'mlat', source_name: 'adsb.lol' })
  )
  assert.equal(evidence.status, 'observed')
  assert.equal(evidence.methodLabel, 'MLAT')
  assert.equal(evidence.sourceName, 'adsb.lol')
  assert.match(evidence.description, /multilateration/i)
})

test('does not infer ADS-B from an ADSB.lol feed name', () => {
  const evidence = buildPositionProvenanceEvidence(
    aircraft({ position_source: null, source_name: 'adsb.lol' })
  )
  assert.equal(evidence.status, 'unavailable')
  assert.equal(evidence.method, null)
  assert.equal(evidence.sourceName, 'adsb.lol')
})

test('presents ADS-R, TIS-B and ADS-C as distinct methods', () => {
  assert.equal(
    buildPositionProvenanceEvidence(aircraft({ position_source: 'adsr' }))
      .methodLabel,
    'ADS-R'
  )
  assert.equal(
    buildPositionProvenanceEvidence(aircraft({ position_source: 'tisb' }))
      .methodLabel,
    'TIS-B'
  )
  assert.equal(
    buildPositionProvenanceEvidence(aircraft({ position_source: 'adsc' }))
      .methodLabel,
    'ADS-C'
  )
})

test('treats unknown runtime values as unavailable instead of fabricating provenance', () => {
  const evidence = buildPositionProvenanceEvidence(
    aircraft({ position_source: 'future-source' })
  )
  assert.equal(evidence.status, 'unavailable')
  assert.equal(evidence.method, null)
})
