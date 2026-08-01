// FRONTEND_UNIFIED_AIRPORT_ANALYTICS_WORKSPACE_V1

import test from 'node:test'
import assert from 'node:assert/strict'

const {
  buildAirportHistorySeries,
  buildAirportRankingView,
  buildAirportTrendSummary,
  mergeAirportLimitations,
  normalizeAirportICAOCode,
} = await import('../.test-dist/lib/analytics/airport-intelligence-workspace-model.js')

function airport(overrides = {}) {
  return {
    position: 1,
    icao_code: 'UBBB',
    iata_code: 'GYD',
    name: 'Heydar Aliyev International Airport',
    city: 'Baku',
    country: 'Azerbaijan',
    activity_score: 0.8,
    data_confidence: 0.9,
    movements_component: 0.7,
    routes_component: 0.6,
    observations_component: 0.8,
    intensity_component: 0.9,
    coverage_score: 0.95,
    freshness_score: 0.92,
    total_movements: 120,
    active_routes: 18,
    observed_samples: 100,
    expected_samples: 105,
    movements_per_hour: 5,
    active_aircraft: 20,
    ...overrides,
  }
}

function history(overrides = {}) {
  return {
    icao_code: 'UBBB',
    window_start: '2026-07-01T00:00:00Z',
    window_end: '2026-07-02T00:00:00Z',
    arrivals: 40,
    departures: 42,
    total_movements: 82,
    arrival_share: 40 / 82,
    departure_share: 42 / 82,
    movements_per_hour: 3.4,
    active_aircraft: 15,
    active_routes: 10,
    observed_samples: 90,
    expected_samples: 96,
    coverage_score: 0.94,
    freshness_score: 0.91,
    latest_observation_at: '2026-07-01T23:58:00Z',
    generated_at: '2026-07-02T00:05:00Z',
    ...overrides,
  }
}

function trends(overrides = {}) {
  const point = {
    window_start: '2026-07-01T00:00:00Z',
    window_end: '2026-07-02T00:00:00Z',
    total_movements: 82,
    movements_per_hour: 3.4,
    active_routes: 10,
    coverage_score: 0.94,
    freshness_score: 0.91,
  }
  return {
    version: '1',
    window: {
      start_time: '2026-07-01T00:00:00Z',
      end_time: '2026-07-31T00:00:00Z',
      as_of_time: '2026-08-01T00:00:00Z',
      completed_days: 30,
    },
    icao_code: 'UBBB',
    compared_windows: 30,
    window_duration_seconds: 86400,
    direction: 'increasing',
    baseline: point,
    current: { ...point, total_movements: 100 },
    peak: { ...point, total_movements: 120 },
    total_movements_change: 18,
    movements_per_hour_change: 0.75,
    movements_per_hour_change_percent: 22,
    movements_per_hour_change_percent_known: true,
    active_routes_change: 2,
    coverage_score_change: 0.02,
    freshness_score_change: -0.01,
    gap_count: 0,
    gap_duration_seconds: 0,
    observed_duration_seconds: 2592000,
    continuity_score: 0.98,
    limitations: [],
    generated_at: '2026-08-01T00:05:00Z',
    ...overrides,
  }
}

test('ICAO normalization accepts four alphanumeric characters only', () => {
  assert.equal(normalizeAirportICAOCode(' ubbb '), 'UBBB')
  assert.equal(normalizeAirportICAOCode('KJFK'), 'KJFK')
  assert.equal(normalizeAirportICAOCode('ABC'), null)
  assert.equal(normalizeAirportICAOCode('ABCDE'), null)
})

test('ranking search matches airport identity and location fields', () => {
  const items = [
    airport(),
    airport({
      position: 2,
      icao_code: 'LTFM',
      iata_code: 'IST',
      name: 'Istanbul Airport',
      city: 'Istanbul',
      country: 'Türkiye',
    }),
  ]
  assert.deepEqual(
    buildAirportRankingView(items, { search: 'istanbul' }).items.map(item => item.icao_code),
    ['LTFM']
  )
  assert.deepEqual(
    buildAirportRankingView(items, { search: 'gyd' }).items.map(item => item.icao_code),
    ['UBBB']
  )
})

test('ranking sorts metrics descending with deterministic ICAO tie breaking', () => {
  const items = [
    airport({ position: 3, icao_code: 'ZZZZ', activity_score: 0.7 }),
    airport({ position: 2, icao_code: 'AAAA', activity_score: 0.9 }),
    airport({ position: 1, icao_code: 'BBBB', activity_score: 0.9 }),
  ]
  assert.deepEqual(
    buildAirportRankingView(items, { sort: 'activity' }).items.map(item => item.icao_code),
    ['AAAA', 'BBBB', 'ZZZZ']
  )
})

test('ranking limit controls display without changing matched count', () => {
  const items = [airport(), airport({ icao_code: 'LTFM' }), airport({ icao_code: 'KJFK' })]
  const result = buildAirportRankingView(items, { limit: 2 })
  assert.equal(result.totalCount, 3)
  assert.equal(result.matchedCount, 3)
  assert.equal(result.visibleCount, 2)
})

test('history series sorts chronologically and uses visible peak denominator', () => {
  const result = buildAirportHistorySeries([
    history({ window_start: '2026-07-03T00:00:00Z', window_end: '2026-07-04T00:00:00Z', total_movements: 60 }),
    history({ window_start: '2026-07-01T00:00:00Z', window_end: '2026-07-02T00:00:00Z', total_movements: 100 }),
    history({ window_start: '2026-07-02T00:00:00Z', window_end: '2026-07-03T00:00:00Z', total_movements: 80 }),
  ], 2)
  assert.deepEqual(result.points.map(point => point.totalMovements), [80, 60])
  assert.equal(result.peakMovements, 80)
  assert.equal(result.points[0].movementShareOfPeak, 1)
  assert.equal(result.points[1].movementShareOfPeak, 0.75)
})

test('empty history remains stable and zero valued', () => {
  assert.deepEqual(buildAirportHistorySeries([]), {
    totalEntryCount: 0,
    visibleEntryCount: 0,
    peakMovements: 0,
    peakMovementsPerHour: 0,
    points: [],
  })
})

test('trend summary preserves explicit unknown percentages and gap evidence', () => {
  const result = buildAirportTrendSummary(
    trends({
      direction: 'decline',
      movements_per_hour_change_percent_known: false,
      movements_per_hour_change_percent: 0,
      gap_count: 2,
      continuity_score: 1.4,
    })
  )
  assert.equal(result.direction, 'decreasing')
  assert.equal(result.directionLabel, 'Activity decreasing')
  assert.equal(result.movementRateDeltaPercent, null)
  assert.equal(result.hasGaps, true)
  assert.equal(result.continuityScore, 1)
})

test('limitations merge case-insensitive duplicates and sort deterministically', () => {
  const result = mergeAirportLimitations(
    [{ code: 'B', message: 'Second' }],
    [
      { code: 'a', message: 'First' },
      { code: 'A', message: 'first' },
      { code: '', message: '' },
    ]
  )
  assert.deepEqual(result, [
    { code: 'a', message: 'First' },
    { code: 'B', message: 'Second' },
  ])
})
