/* global fetch */
import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

import {
  createMockAPIServer,
  openAPIPaths,
  resolveMockResponse,
  supportedScenarios,
} from '../apps/web/e2e/mock-api.mjs'
import {
  countBrowserScenarios,
  expectedBrowserScenarioCount,
  expectedPlaywrightVersion,
  expectedScenarios,
  productTestFiles,
  validatePlaywrightFoundation,
} from './verify-playwright-e2e.mjs'

test('repository Playwright product coverage contract passes', () => {
  assert.deepEqual(validatePlaywrightFoundation(process.cwd()), [])
})

test('mock API mirrors the exact OpenAPI path surface', () => {
  const openAPI = JSON.parse(fs.readFileSync('openapi/openapi.json', 'utf8'))
  assert.deepEqual(
    [...openAPIPaths].sort(),
    Object.keys(openAPI.paths).sort(),
  )
})

test('Playwright version and browser scenario count are pinned', () => {
  assert.equal(expectedPlaywrightVersion, '1.62.0')
  assert.equal(expectedBrowserScenarioCount, 20)
  assert.equal(countBrowserScenarios(process.cwd()), 20)
})

test('mock failure scenario surface is explicit and bounded', () => {
  assert.deepEqual(
    [...supportedScenarios].sort(),
    [...expectedScenarios].sort(),
  )
  assert.deepEqual(
    [...supportedScenarios].sort(),
    [
      'aircraft-error',
      'airport-error',
      'healthy',
      'historical-error',
      'intelligence-error',
      'regions-error',
      'traffic-error',
    ],
  )
})

test('healthy traffic fixture is deterministic', () => {
  const result = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/traffic/current?region=az',
    scenario: 'healthy',
  })
  assert.equal(result.status, 200)
  assert.equal(result.body.success, true)
  assert.equal(result.body.data.length, 1)
  assert.equal(result.body.data[0].icao24, '4b1801')
  assert.equal('region_code' in result.body.data[0], false)
})

test('live traffic fixture mirrors the current-state snapshot contract', () => {
  const result = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/traffic/live?limit=1',
    scenario: 'healthy',
  })

  assert.equal(result.status, 200)
  assert.equal(result.body.success, true)
  assert.equal(result.body.data.server_time, '2026-08-04T18:00:10Z')
  assert.equal(result.body.data.sequence, 42)
  assert.equal(result.body.data.total_active, 2)
  assert.equal(result.body.data.matching, 2)
  assert.equal(result.body.data.truncated, true)
  assert.equal(result.body.data.aircraft.length, 1)
  assert.equal(result.body.data.aircraft[0].icao24, '4b1801')
  assert.equal(result.body.data.aircraft[0].source, 'playwright-fixture')
  assert.equal(typeof result.body.data.aircraft[0].freshness_ms, 'number')
  assert.equal('altitude_status' in result.body.data.aircraft[0], false)
})

test('core read fixtures preserve typed trajectory and metric envelopes', () => {
  const trajectory = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/aircraft/4b1801/trajectory',
    scenario: 'healthy',
  })
  assert.equal(trajectory.status, 200)
  assert.equal(trajectory.body.success, true)
  assert.equal(trajectory.body.data.id, '11111111-1111-4111-8111-111111111111')
  assert.equal(trajectory.body.data.segments.length, 1)
  assert.equal(trajectory.body.data.identity_basis, 'aircraft_and_start_time')
  assert.equal(trajectory.body.data.split_reason, 'initial_observation')
  assert.equal(trajectory.body.data.segments[0].status, 'observed')

  const metric = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/metrics/active-aircraft?region=az',
    scenario: 'healthy',
  })
  assert.equal(metric.status, 200)
  assert.deepEqual(metric.body.data.scope, { type: 'region', code: 'az' })
  assert.equal(metric.body.data.window_minutes, 15)
})

test('advanced intelligence fixtures preserve evidence pagination and uncertainty semantics', () => {
  const transponder = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/aircraft/4b1801/transponder-evidence/latest',
    scenario: 'healthy',
  })
  assert.equal(transponder.status, 200)
  assert.equal(transponder.body.data.evidence_only, true)
  assert.equal(transponder.body.data.confirmed_emergency, false)

  const weather = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/weather/current?lat=40.4&lon=49.8',
    scenario: 'healthy',
  })
  assert.equal(weather.status, 200)
  assert.equal(weather.body.data.rain_mm, null)
  assert.equal(weather.body.data.wind_gusts_mps, null)

  const history = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/historical-intelligence/aggregates/history?metric=active_aircraft&scope=global&granularity=hour',
    scenario: 'healthy',
  })
  assert.equal(history.body.data.has_more, true)
  assert.equal(history.body.data.next_cursor, 'fixture-cursor-v1')

  const projection = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/trajectories/11111111-1111-4111-8111-111111111111/projection-intelligence?as_of_time=2026-08-04T18:00:00Z',
    scenario: 'healthy',
  })
  assert.match(projection.body.data.input_fingerprint, /^sha256:[0-9a-f]{64}$/)
  assert.equal(projection.body.data.projection.points.length, 1)

  const airspace = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/airspace/regions/az/analytics?as_of_time=2026-08-04T18:00:00Z',
    scenario: 'healthy',
  })
  assert.equal(airspace.body.data.status, 'available')
  assert.equal(airspace.body.data.metrics.temporal_coverage, 1)
})

test('Stability Intelligence fixture mirrors requested analytical timestamps', () => {
  const requested = [
    '2026-08-04T17:59:00.000Z',
    '2026-08-04T17:59:30.000Z',
    '2026-08-04T18:00:00.000Z',
  ]
  const result = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/trajectories/11111111-1111-4111-8111-111111111111/stability-intelligence?as_of_times=' +
      encodeURIComponent(requested.join(',')) +
      '&duration_seconds=300',
    scenario: 'healthy',
  })

  assert.equal(result.status, 200)
  assert.deepEqual(result.body.data.as_of_times, requested)
  assert.equal(result.body.data.projections.length, requested.length)
  assert.equal(result.body.data.forecast_versions.length, requested.length)
  assert.equal(result.body.data.transitions.length, requested.length - 1)
  assert.equal(
    result.body.data.forecast_analysis.metrics.version_count,
    requested.length,
  )
  assert.equal(
    result.body.data.forecast_analysis.metrics.transition_count,
    requested.length - 1,
  )
  assert.equal(
    result.body.data.propagated_confidence.target_node_id,
    'forecast-v3',
  )
})

test('Route Intelligence fixtures preserve mutation security and history pagination', () => {
  const requestURL =
    'http://127.0.0.1:8091/api/v1/trajectories/11111111-1111-4111-8111-111111111111/route-intelligence'

  const unauthorized = resolveMockResponse({
    method: 'POST',
    requestURL,
    scenario: 'healthy',
  })
  assert.equal(unauthorized.status, 401)
  assert.equal(
    unauthorized.body.error.code,
    'MUTATION_AUTHENTICATION_REQUIRED',
  )

  const processed = resolveMockResponse({
    method: 'POST',
    requestURL,
    scenario: 'healthy',
    headers: {
      'x-internal-api-key': 'playwright-route-intelligence-key-v1',
    },
  })
  assert.equal(processed.status, 200)
  assert.equal(processed.body.data.result.status, 'complete')
  assert.equal(processed.body.data.result.origin.airport.icao_code, 'UBBB')
  assert.equal(
    processed.body.data.result.limitations[0].code,
    'inferred_not_filed',
  )

  const latest = resolveMockResponse({
    method: 'GET',
    requestURL: `${requestURL}/latest`,
    scenario: 'healthy',
  })
  assert.equal(latest.status, 200)
  assert.equal(latest.body.data.id, processed.body.data.id)

  const history = resolveMockResponse({
    method: 'GET',
    requestURL: `${requestURL}/history?limit=20`,
    scenario: 'healthy',
  })
  assert.equal(history.status, 200)
  assert.equal(history.body.data.has_more, true)
  assert.equal(
    history.body.data.next_before_as_of_time,
    '2026-08-05T18:00:00Z',
  )
})

test('all deterministic failure scenarios preserve typed 503 envelopes', () => {
  const cases = [
    [
      'aircraft-error',
      'http://127.0.0.1:8091/api/v1/aircraft/4b1801/trajectory',
      'AIRCRAFT_FIXTURE_UNAVAILABLE',
    ],
    [
      'airport-error',
      'http://127.0.0.1:8091/api/v1/airports/intelligence/ranking?days=30&limit=100',
      'AIRPORT_INTELLIGENCE_FIXTURE_UNAVAILABLE',
    ],
    [
      'historical-error',
      'http://127.0.0.1:8091/api/v1/historical-intelligence/aggregates/latest?metric=active_aircraft&scope=global&granularity=day',
      'HISTORICAL_INTELLIGENCE_FIXTURE_UNAVAILABLE',
    ],
    [
      'intelligence-error',
      'http://127.0.0.1:8091/api/v1/trajectories/11111111-1111-4111-8111-111111111111/projection-intelligence?as_of_time=2026-08-04T18:00:00Z',
      'ADVANCED_INTELLIGENCE_FIXTURE_UNAVAILABLE',
    ],
  ]

  for (const [scenario, requestURL, code] of cases) {
    const result = resolveMockResponse({
      method: 'GET',
      requestURL,
      scenario,
    })
    assert.equal(result.status, 503, scenario)
    assert.equal(result.body.success, false, scenario)
    assert.equal(result.body.error.code, code, scenario)
  }
})

test('traffic-error fixture preserves the typed error envelope', () => {
  const result = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/traffic/current?region=world',
    scenario: 'traffic-error',
  })
  assert.equal(result.status, 503)
  assert.deepEqual(result.body, {
    success: false,
    error: {
      code: 'TRAFFIC_FIXTURE_UNAVAILABLE',
      message: 'The deterministic traffic fixture is unavailable.',
    },
  })
})

test('mock control endpoint changes the live scenario', async () => {
  const mock = createMockAPIServer({ port: 0 })
  const address = await mock.listen()
  try {
    assert.equal(typeof address, 'object')
    const origin = `http://127.0.0.1:${address.port}`
    const control = await fetch(`${origin}/__e2e/scenario`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ scenario: 'intelligence-error' }),
    })
    assert.equal(control.status, 200)
    assert.equal(mock.getScenario(), 'intelligence-error')

    const projection = await fetch(
      `${origin}/api/v1/trajectories/11111111-1111-4111-8111-111111111111/projection-intelligence?as_of_time=2026-08-04T18:00:00Z`,
    )
    assert.equal(projection.status, 503)
    assert.equal((await projection.json()).success, false)
  } finally {
    await mock.close()
  }
})

test('browser product coverage uses semantic locators and no public deployment', () => {
  const source = productTestFiles
    .map(relativePath => fs.readFileSync(relativePath, 'utf8'))
    .join('\n')
  assert.match(source, /getByRole/)
  assert.doesNotMatch(source, /page\.locator\(/)
  assert.doesNotMatch(source, /data-testid/)
  assert.doesNotMatch(source, /onrender\.com|vercel\.app/)
})

test('visual regression records evidence without freezing pre-redesign pixels', () => {
  const source = fs.readFileSync(
    'apps/web/e2e/tests/visual-regression.spec.mjs',
    'utf8',
  )
  assert.match(source, /testInfo\.attach/)
  assert.match(source, /page\.screenshot/)
  assert.match(source, /boundingBox/)
  assert.match(source, /expectNoHorizontalOverflow/)
  assert.doesNotMatch(source, /toHaveScreenshot/)
})

test('CI rejects flaky browser tests and waits for hydration', () => {
  const config = fs.readFileSync(
    'apps/web/e2e/playwright.config.mjs',
    'utf8',
  )
  const applicationShell = fs.readFileSync(
    'apps/web/e2e/tests/application-shell.spec.mjs',
    'utf8',
  )

  assert.match(config, /failOnFlakyTests: isCI/)
  assert.match(applicationShell, /region=WORLD&view=invalid/)
  assert.match(applicationShell, /region=world&view=aircraft/)
  assert.ok(
    applicationShell.indexOf('region=world&view=aircraft') <
      applicationShell.indexOf("region.selectOption('az')"),
  )
})
