import assert from 'node:assert/strict'
import test from 'node:test'

const model = await import('../.test-dist/lib/analytics/historical-analytics-comparison-model.js')

const baseConfidence = { score: 0.8, level: 'high', sample_count: 2, reasons: [] }
const point = (overrides = {}) => ({
  start_time: '2026-07-01T00:00:00Z',
  end_time: '2026-07-02T00:00:00Z',
  status: 'complete',
  value: 4,
  sample_count: 2,
  coverage_ratio: 1,
  confidence: baseConfidence,
  limitations: [],
  ...overrides,
})
const result = (overrides = {}) => ({
  schema_version: 'historical-intelligence-v1',
  status: 'complete',
  metric: { name: 'active_aircraft', unit: 'aircraft', aggregation: 'count' },
  scope: { type: 'global' },
  window: {
    start_time: '2026-07-01T00:00:00Z',
    end_time: '2026-07-02T00:00:00Z',
    as_of_time: '2026-07-02T00:00:00Z',
  },
  granularity: 'day',
  points: [point()],
  summary: { point_count: 1, total: 4, minimum: 4, maximum: 4, average: 4, median: 4 },
  confidence: baseConfidence,
  limitations: [],
  provenance: {
    builder_version: 'v1',
    input_fingerprint: `sha256:${'a'.repeat(64)}`,
    source_names: ['flight_states'],
    latest_source_updated_at: '2026-07-02T00:00:00Z',
  },
  generated_at: '2026-07-02T00:00:00Z',
  ...overrides,
})
const record = (id, storedAt, overrides = {}) => ({
  id,
  input_fingerprint: `sha256:${'b'.repeat(64)}`,
  stored_at: storedAt,
  result: result(overrides),
})

test('metric catalog exposes only server-supported scopes', () => {
  assert.deepEqual(
    model.metricsForScope('airport').map(item => item.name),
    ['airport_departures', 'airport_arrivals', 'airport_operations', 'unique_aircraft']
  )
  assert.ok(model.metricsForScope('route').every(item => item.scopes.includes('route')))
})

test('scope changes replace unsupported metrics deterministically', () => {
  const normalized = model.normalizeHistoricalSelection({
    scope: 'airport',
    metric: 'active_aircraft',
    granularity: 'day',
    airportICAO: ' ubbb ',
  })
  assert.equal(normalized.metric, 'airport_departures')
  assert.equal(normalized.airportICAO, 'UBBB')
})

test('incomplete ICAO scopes remain disabled', () => {
  assert.equal(model.selectionIsComplete({ scope: 'airport', metric: 'airport_operations', granularity: 'day', airportICAO: 'UBB' }), false)
  assert.equal(model.selectionIsComplete({ scope: 'route', metric: 'active_routes', granularity: 'day', originICAO: 'UBBB', destinationICAO: 'LTFM' }), true)
})

test('series view sorts chronologically and preserves quality states', () => {
  const view = model.buildHistoricalSeriesView([
    point({ start_time: '2026-07-02T00:00:00Z', end_time: '2026-07-03T00:00:00Z', status: 'partial', value: 3 }),
    point({ start_time: '2026-07-01T00:00:00Z', status: 'unavailable', value: 0 }),
  ])
  assert.equal(view.points[0].startTime, '2026-07-01T00:00:00Z')
  assert.equal(view.partialCount, 1)
  assert.equal(view.unavailableCount, 1)
  assert.equal(view.maximumValue, 3)
})

test('missing period comparisons remain explicitly unavailable', () => {
  assert.deepEqual(model.buildPeriodComparisonView(result()), {
    available: false,
    direction: 'unavailable',
    currentValue: null,
    previousValue: null,
    absoluteChange: null,
    percentageChange: null,
    previousWindowLabel: null,
  })
})

test('record comparison uses metric aggregation semantics', () => {
  const comparison = model.compareHistoricalRecords(
    record('a', '2026-07-01T00:00:00Z', { summary: { point_count: 1, total: 10, minimum: 10, maximum: 10, average: 10, median: 10 } }),
    record('b', '2026-07-02T00:00:00Z', { summary: { point_count: 1, total: 15, minimum: 15, maximum: 15, average: 15, median: 15 } })
  )
  assert.equal(comparison.comparable, true)
  assert.equal(comparison.absoluteChange, 5)
  assert.equal(comparison.percentageChange, 50)
  assert.equal(comparison.direction, 'up')
})

test('record comparison refuses incompatible series', () => {
  const comparison = model.compareHistoricalRecords(
    record('a', '2026-07-01T00:00:00Z'),
    record('b', '2026-07-02T00:00:00Z', { metric: { name: 'flight_count', unit: 'flights', aggregation: 'count' } })
  )
  assert.equal(comparison.comparable, false)
  assert.match(comparison.reason, /same metric/)
})

test('history and limitations sort deterministically', () => {
  const history = model.sortAggregateHistory([
    record('b', '2026-07-01T00:00:00Z'),
    record('a', '2026-07-02T00:00:00Z'),
  ])
  assert.equal(history[0].id, 'a')
  const limitations = model.mergeHistoricalLimitations(
    [{ code: 'B', message: 'Second', scope: 'series' }],
    [{ code: 'a', message: 'First', scope: 'bucket' }],
    [{ code: 'A', message: 'First', scope: 'bucket' }]
  )
  assert.equal(limitations.length, 2)
  assert.equal(limitations[0].scope, 'bucket')
})
