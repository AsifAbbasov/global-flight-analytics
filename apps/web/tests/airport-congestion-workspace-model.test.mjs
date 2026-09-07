import test from 'node:test'
import assert from 'node:assert/strict'

const { buildAirportCongestionSummary } = await import(
  '../.test-dist/lib/analytics/airport-intelligence-workspace-model.js'
)

function congestion(overrides = {}) {
  return {
    version: 'airport-congestion-intelligence-v1',
    status: 'available',
    window: {
      start_time: '2026-07-06T00:00:00Z',
      end_time: '2026-08-05T00:00:00Z',
      as_of_time: '2026-08-05T12:00:00Z',
      completed_days: 30,
    },
    icao_code: 'UBBB',
    current: {
      icao_code: 'UBBB',
      window_start: '2026-08-04T00:00:00Z',
      window_end: '2026-08-05T00:00:00Z',
      arrivals: 21,
      departures: 19,
      total_movements: 40,
      arrival_share: 0.525,
      departure_share: 0.475,
      movements_per_hour: 1.6667,
      active_aircraft: 16,
      active_routes: 9,
      observed_samples: 1,
      expected_samples: 1,
      coverage_score: 0.96,
      freshness_score: 0.97,
      latest_observation_at: '2026-08-04T23:58:00Z',
      generated_at: '2026-08-05T00:00:00Z',
    },
    observed_window_count: 29,
    expected_window_count: 30,
    gap_window_count: 1,
    trailing_gap_window_count: 0,
    current_window_is_latest_expected: true,
    evidence_coverage: 29 / 30,
    evidence_support: 0.9,
    baseline_window_count: 28,
    baseline_median_movements_per_hour: 1.2,
    prior_peak_movements_per_hour: 1.5,
    current_to_baseline_ratio: 1.3889,
    current_to_baseline_ratio_known: true,
    current_to_prior_peak_ratio: 1.1111,
    current_to_prior_peak_ratio_known: true,
    congestion_score: 1,
    congestion_score_known: true,
    exceeds_prior_observed_activity_peak: true,
    score_semantics: 'relative observed activity only',
    scope_guard: 'relative_observed_activity_only_not_airport_capacity_delay_queue_slot_or_runway_congestion',
    explanation: 'Research-only relative activity proxy.',
    limitations: [],
    generated_at: '2026-08-05T12:00:01Z',
    ...overrides,
  }
}

test('congestion summary preserves published ratios and evidence support', () => {
  const result = buildAirportCongestionSummary(congestion())
  assert.equal(result.status, 'available')
  assert.equal(result.score, 1)
  assert.equal(result.currentToBaselineRatio, 1.3889)
  assert.equal(result.currentToPriorPeakRatio, 1.1111)
  assert.equal(result.exceedsPriorObservedActivityPeak, true)
  assert.equal(result.currentWindowIsLatestExpected, true)
  assert.equal(result.observedWindowCount, 29)
  assert.equal(result.expectedWindowCount, 30)
  assert.equal(result.gapWindowCount, 1)
  assert.equal(result.trailingGapWindowCount, 0)
  assert.equal(result.evidenceSupport, 0.9)
})

test('unknown score and ratios remain null rather than becoming zero', () => {
  const result = buildAirportCongestionSummary(
    congestion({
      status: 'unavailable',
      congestion_score: 0,
      congestion_score_known: false,
      current_to_baseline_ratio: 0,
      current_to_baseline_ratio_known: false,
      current_to_prior_peak_ratio: 0,
      current_to_prior_peak_ratio_known: false,
      current_window_is_latest_expected: false,
      trailing_gap_window_count: 1,
    })
  )

  assert.equal(result.status, 'unavailable')
  assert.equal(result.score, null)
  assert.equal(result.currentToBaselineRatio, null)
  assert.equal(result.currentToPriorPeakRatio, null)
  assert.equal(result.currentWindowIsLatestExpected, false)
  assert.equal(result.trailingGapWindowCount, 1)
})

test('evidence ratios are bounded without inventing score availability', () => {
  const result = buildAirportCongestionSummary(
    congestion({
      congestion_score_known: false,
      evidence_coverage: 1.4,
      evidence_support: -0.2,
    })
  )

  assert.equal(result.score, null)
  assert.equal(result.evidenceCoverage, 1)
  assert.equal(result.evidenceSupport, 0)
})
