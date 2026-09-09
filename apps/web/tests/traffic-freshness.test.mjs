import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL('../.test-dist/lib/traffic/traffic-freshness.js', import.meta.url)
const {
  buildTrafficFreshnessEvidence,
  defaultFutureClockSkewToleranceMilliseconds,
  defaultRecentObservationWindowMilliseconds,
  formatTrafficEvidenceAge,
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

test('separates fresh position age from newer message age', () => {
  const reference = Date.parse('2026-09-09T12:00:10Z')
  const evidence = buildTrafficFreshnessEvidence(aircraft(), reference)
  assert.equal(evidence.positionStatus, 'fresh')
  assert.equal(evidence.positionAgeMilliseconds, 10_000)
  assert.equal(evidence.messageAgeMilliseconds, 2_000)
  assert.equal(formatTrafficEvidenceAge(evidence.positionAgeMilliseconds), '10s old')
})

test('marks an old position older even when the latest message is fresh', () => {
  const reference = Date.parse('2026-09-09T12:10:00Z')
  const evidence = buildTrafficFreshnessEvidence(
    aircraft({
      position_observed_at: '2026-09-09T12:00:00Z',
      observed_at: '2026-09-09T12:00:00Z',
      message_observed_at: '2026-09-09T12:09:58Z',
    }),
    reference
  )
  assert.equal(evidence.positionStatus, 'older')
  assert.equal(evidence.messageAgeMilliseconds, 2_000)
})

test('preserves the existing five-minute and one-minute project thresholds', () => {
  assert.equal(defaultRecentObservationWindowMilliseconds, 5 * 60 * 1000)
  assert.equal(defaultFutureClockSkewToleranceMilliseconds, 60 * 1000)
})

test('flags position timestamps beyond the accepted future clock skew', () => {
  const reference = Date.parse('2026-09-09T12:00:00Z')
  const evidence = buildTrafficFreshnessEvidence(
    aircraft({ position_observed_at: '2026-09-09T12:01:01Z' }),
    reference
  )
  assert.equal(evidence.positionStatus, 'future')
})

test('uses observed_at as a rolling-deployment compatibility fallback', () => {
  const reference = Date.parse('2026-09-09T12:00:10Z')
  const legacy = aircraft({ position_observed_at: '' })
  const evidence = buildTrafficFreshnessEvidence(legacy, reference)
  assert.equal(evidence.positionObservedAt, legacy.observed_at)
  assert.equal(evidence.positionStatus, 'fresh')
})

test('does not fabricate a last-message timestamp when provider evidence is absent', () => {
  const reference = Date.parse('2026-09-09T12:00:10Z')
  const evidence = buildTrafficFreshnessEvidence(
    aircraft({ message_observed_at: null }),
    reference
  )
  assert.equal(evidence.messageObservedAt, null)
  assert.equal(evidence.messageAgeMilliseconds, null)
})
