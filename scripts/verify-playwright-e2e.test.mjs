/* global fetch */
import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

import {
  createMockAPIServer,
  openAPIPaths,
  resolveMockResponse,
} from '../apps/web/e2e/mock-api.mjs'
import {
  expectedPlaywrightVersion,
  validatePlaywrightFoundation,
} from './verify-playwright-e2e.mjs'

test('repository Playwright foundation contract passes', () => {
  assert.deepEqual(validatePlaywrightFoundation(process.cwd()), [])
})

test('mock API mirrors the exact OpenAPI path surface', () => {
  const openAPI = JSON.parse(fs.readFileSync('openapi/openapi.json', 'utf8'))
  assert.deepEqual(
    [...openAPIPaths].sort(),
    Object.keys(openAPI.paths).sort(),
  )
})

test('Playwright version is pinned', () => {
  assert.equal(expectedPlaywrightVersion, '1.62.0')
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

test('core read fixtures preserve typed trajectory and metric envelopes', () => {
  const trajectory = resolveMockResponse({
    method: 'GET',
    requestURL:
      'http://127.0.0.1:8091/api/v1/aircraft/4b1801/trajectory',
    scenario: 'healthy',
  })
  assert.equal(trajectory.status, 200)
  assert.equal(trajectory.body.success, true)
  assert.equal(trajectory.body.data.id, 'trajectory-azal-101')
  assert.equal(trajectory.body.data.segments.length, 1)

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
      body: JSON.stringify({ scenario: 'traffic-error' }),
    })
    assert.equal(control.status, 200)
    assert.equal(mock.getScenario(), 'traffic-error')

    const traffic = await fetch(
      `${origin}/api/v1/traffic/current?region=world`,
    )
    assert.equal(traffic.status, 503)
    assert.equal((await traffic.json()).success, false)
  } finally {
    await mock.close()
  }
})

test('browser tests use semantic locators and no public deployment', () => {
  const source = [
    fs.readFileSync(
      'apps/web/e2e/tests/application-shell.spec.mjs',
      'utf8',
    ),
    fs.readFileSync(
      'apps/web/e2e/tests/api-recovery.spec.mjs',
      'utf8',
    ),
  ].join('\n')
  assert.match(source, /getByRole/)
  assert.doesNotMatch(source, /page\.locator\(/)
  assert.doesNotMatch(source, /data-testid/)
  assert.doesNotMatch(source, /onrender\.com|vercel\.app/)
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
