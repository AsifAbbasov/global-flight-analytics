import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL(
  '../.test-dist/lib/projection/eta-evolution-model.js',
  import.meta.url
)
const importedModule = await import(moduleURL.href)
const model = importedModule.default ?? importedModule
const {
  etaEvolutionMaximumSamples,
  selectETAEvolutionAsOfTimes,
  buildETAEvolutionSummary,
} = model

function replayWithTimes(times) {
  return {
    trajectory_id: '11111111-1111-4111-8111-111111111111',
    flight_id: 'flight-1',
    icao24: '4b1801',
    callsign: 'AZAL101',
    start_time: times[0] ?? '2026-08-04T10:00:00Z',
    end_time: times[times.length - 1] ?? '2026-08-04T10:00:00Z',
    evidence_class: 'observed',
    interpolation_policy: 'none',
    points: times.map((observed_at, index) => ({
      id: `state-${index}`,
      flight_id: 'flight-1',
      icao24: '4b1801',
      callsign: 'AZAL101',
      latitude: 40,
      longitude: 50,
      barometric_altitude_m: 3000,
      barometric_altitude_status: 'observed',
      geometric_altitude_m: null,
      geometric_altitude_status: 'unavailable',
      velocity_mps: 200,
      heading_degrees: 90,
      vertical_rate_mps: 0,
      on_ground: false,
      origin_country: 'Azerbaijan',
      observed_at,
      source_name: 'fixture',
    })),
  }
}

function projection(asOfTime, estimatedTime, options = {}) {
  const earliestTime = new Date(Date.parse(estimatedTime) - 5 * 60_000).toISOString()
  const latestTime = new Date(Date.parse(estimatedTime) + 5 * 60_000).toISOString()
  return {
    version: 'projection-production-composition-v2',
    strategy: options.strategy ?? 'kinematic_baseline',
    fallback_reason: '',
    arrival_status: options.arrivalStatus ?? 'attached',
    projection: {
      schema_version: 'projection-intelligence-v1',
      status: 'complete',
      trajectory_id: '11111111-1111-4111-8111-111111111111',
      flight_id: 'flight-1',
      aircraft_id: 'aircraft-1',
      icao24: '4b1801',
      callsign: 'AZAL101',
      method: { name: 'fixture-method', version: 'v1', decision_class: 'project_derived' },
      horizon: { as_of_time: asOfTime, end_time: estimatedTime, step_seconds: 60, duration_seconds: 300 },
      points: [],
      ...(options.withoutArrival
        ? {}
        : {
            arrival: {
              airport_icao_code: 'UBBB',
              earliest_time: earliestTime,
              estimated_time: estimatedTime,
              latest_time: latestTime,
              confidence: { score: 0.8, level: 'high', reasons: [] },
              limitations: [],
            },
          }),
      confidence: { score: 0.8, level: 'high', reasons: [] },
      limitations: [],
      explanations: [],
      scope_guard: 'research_only_not_for_operational_use',
      provenance: {
        input_fingerprint: `sha256:${'a'.repeat(64)}`,
        inputs: [],
        latest_input_observed_at: asOfTime,
      },
      generated_at: '2026-09-07T00:00:00Z',
    },
    evidence: {},
    notices: [],
    input_fingerprint: `sha256:${'b'.repeat(64)}`,
    generated_at: '2026-09-07T00:00:00Z',
  }
}

test('sampler is bounded and uses only persisted observation timestamps', () => {
  const times = Array.from({ length: 20 }, (_, index) =>
    new Date(Date.UTC(2026, 7, 4, 10, index, 0)).toISOString()
  )
  const selected = selectETAEvolutionAsOfTimes(replayWithTimes(times))

  assert.equal(selected.length, etaEvolutionMaximumSamples)
  assert.equal(selected[0], times[0])
  assert.equal(selected[selected.length - 1], times[times.length - 1])
  for (const timestamp of selected) assert.ok(times.includes(timestamp))
})

test('sampler preserves every valid timestamp when replay is already small', () => {
  const times = [
    '2026-08-04T10:00:00Z',
    '2026-08-04T10:01:00Z',
    '2026-08-04T10:02:00Z',
  ]
  assert.deepEqual(selectETAEvolutionAsOfTimes(replayWithTimes(times)), times)
})

test('ETA evolution reports transparent adjacent and net changes', () => {
  const inputs = [
    {
      asOfTime: '2026-08-04T10:00:00Z',
      result: projection('2026-08-04T10:00:00Z', '2026-08-04T11:00:00Z'),
    },
    {
      asOfTime: '2026-08-04T10:05:00Z',
      result: projection('2026-08-04T10:05:00Z', '2026-08-04T11:04:00Z'),
    },
    {
      asOfTime: '2026-08-04T10:10:00Z',
      result: projection('2026-08-04T10:10:00Z', '2026-08-04T11:02:00Z'),
    },
  ]

  const summary = buildETAEvolutionSummary(inputs)
  assert.equal(summary.sampledObservationCount, 3)
  assert.equal(summary.availableETACount, 3)
  assert.equal(summary.unavailableETACount, 0)
  assert.equal(summary.points[1].etaChangeSeconds, 240)
  assert.equal(summary.points[2].etaChangeSeconds, -120)
  assert.equal(summary.netETAChangeSeconds, 120)
  assert.equal(summary.largestAbsoluteAdjacentETAChangeSeconds, 240)
  assert.equal(summary.points[0].arrivalWindowSeconds, 600)
})

test('unavailable point breaks adjacent ETA comparison instead of carrying forecast forward', () => {
  const summary = buildETAEvolutionSummary([
    {
      asOfTime: '2026-08-04T10:00:00Z',
      result: projection('2026-08-04T10:00:00Z', '2026-08-04T11:00:00Z'),
    },
    {
      asOfTime: '2026-08-04T10:05:00Z',
      result: projection('2026-08-04T10:05:00Z', '2026-08-04T11:00:00Z', {
        withoutArrival: true,
        arrivalStatus: 'withheld',
      }),
    },
    {
      asOfTime: '2026-08-04T10:10:00Z',
      result: projection('2026-08-04T10:10:00Z', '2026-08-04T11:05:00Z'),
    },
  ])

  assert.equal(summary.availableETACount, 2)
  assert.equal(summary.unavailableETACount, 1)
  assert.equal(summary.points[1].available, false)
  assert.equal(summary.points[1].estimatedTime, null)
  assert.equal(summary.points[2].etaChangeSeconds, null)
  assert.equal(summary.netETAChangeSeconds, 300)
  assert.equal(summary.largestAbsoluteAdjacentETAChangeSeconds, null)
})

test('transport error remains unavailable evidence rather than a synthetic ETA', () => {
  const summary = buildETAEvolutionSummary([
    {
      asOfTime: '2026-08-04T10:00:00Z',
      error: 'Projection Intelligence unavailable',
    },
  ])

  assert.equal(summary.availableETACount, 0)
  assert.equal(summary.unavailableETACount, 1)
  assert.equal(summary.points[0].estimatedTime, null)
  assert.equal(summary.points[0].etaChangeSeconds, null)
  assert.equal(summary.points[0].error, 'Projection Intelligence unavailable')
})
