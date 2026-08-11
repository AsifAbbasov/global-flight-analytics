#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import {
  openAPIPaths,
  resolveMockResponse,
  supportedScenarios,
} from '../apps/web/e2e/mock-api.mjs'

export const expectedPlaywrightVersion = '1.62.0'
export const expectedBrowserScenarioCount = 20
export const expectedScenarios = new Set([
  'healthy',
  'traffic-error',
  'regions-error',
  'aircraft-error',
  'airport-error',
  'historical-error',
  'intelligence-error',
])

export const productTestFiles = Object.freeze([
  'apps/web/e2e/tests/application-shell.spec.mjs',
  'apps/web/e2e/tests/api-recovery.spec.mjs',
  'apps/web/e2e/tests/aircraft-intelligence.spec.mjs',
  'apps/web/e2e/tests/airport-intelligence.spec.mjs',
  'apps/web/e2e/tests/historical-analytics.spec.mjs',
  'apps/web/e2e/tests/advanced-intelligence.spec.mjs',
  'apps/web/e2e/tests/exports.spec.mjs',
  'apps/web/e2e/tests/product-recovery.spec.mjs',
  'apps/web/e2e/tests/accessibility.spec.mjs',
  'apps/web/e2e/tests/visual-regression.spec.mjs',
])

function read(root, relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8')
}

function fail(errors, message) {
  errors.push(message)
}

function requireIncludes(errors, content, literals, label) {
  for (const literal of literals) {
    if (!content.includes(literal)) {
      fail(errors, `${label} is missing ${literal}`)
    }
  }
}

function sorted(values) {
  return [...values].sort()
}

export function countBrowserScenarios(root) {
  return productTestFiles.reduce((total, relativePath) => {
    const source = read(root, relativePath)
    return total + (source.match(/\btest\s*\(\s*['"`]/g) ?? []).length
  }, 0)
}

export function validatePlaywrightFoundation(root) {
  const errors = []
  const openAPI = JSON.parse(read(root, 'openapi/openapi.json'))
  const documentedPaths = new Set(Object.keys(openAPI.paths ?? {}))

  if (
    JSON.stringify(sorted(documentedPaths)) !==
    JSON.stringify(sorted(openAPIPaths))
  ) {
    fail(errors, 'mock API path surface does not match OpenAPI')
  }

  if (
    JSON.stringify(sorted(supportedScenarios)) !==
    JSON.stringify(sorted(expectedScenarios))
  ) {
    fail(errors, 'mock API scenario surface is not canonical')
  }

  for (const relativePath of productTestFiles) {
    if (!fs.existsSync(path.join(root, relativePath))) {
      fail(errors, `missing Playwright product test: ${relativePath}`)
    }
  }

  const browserScenarioCount = countBrowserScenarios(root)
  if (browserScenarioCount !== expectedBrowserScenarioCount) {
    fail(
      errors,
      `expected ${expectedBrowserScenarioCount} browser scenarios, got ${browserScenarioCount}`,
    )
  }

  const runner = read(root, 'scripts/run-playwright-e2e.sh')
  requireIncludes(
    errors,
    runner,
    [
      'PLAYWRIGHT_VERSION="${PLAYWRIGHT_VERSION:-1.62.0}"',
      '--prefix "$E2E_ROOT"',
      '--no-save',
      '--package-lock=false',
      '--ignore-scripts',
      '"@playwright/test@$PLAYWRIGHT_VERSION"',
      'install --with-deps chromium',
      'pnpm --dir apps/web build',
      '--project chromium',
      'PLAYWRIGHT_E2E_PUBLIC_NETWORK=DISABLED',
      'PLAYWRIGHT_E2E=PASS',
    ],
    'Playwright runner',
  )
  if (/onrender\.com|vercel\.app/.test(runner)) {
    fail(errors, 'Playwright runner must not target a public deployment')
  }
  if (runner.includes('pnpm add') || runner.includes('pnpm install --lockfile-only')) {
    fail(errors, 'Playwright runner must not mutate the pnpm lockfile')
  }

  const config = read(root, 'apps/web/e2e/playwright.config.mjs')
  requireIncludes(
    errors,
    config,
    [
      'workers: 1',
      'fullyParallel: false',
      'failOnFlakyTests: isCI',
      "trace: 'retain-on-failure'",
      "screenshot: 'only-on-failure'",
      "video: 'retain-on-failure'",
      "name: 'chromium'",
      "command: 'node mock-api.mjs'",
      "command: 'pnpm start --hostname 127.0.0.1 --port 3000'",
      'NEXT_PUBLIC_API_BASE_URL: apiOrigin',
    ],
    'Playwright config',
  )
  if (/onrender\.com|vercel\.app/.test(config)) {
    fail(errors, 'Playwright config must not target a public deployment')
  }

  const testSource = productTestFiles
    .map(relativePath => read(root, relativePath))
    .join('\n')
  requireIncludes(
    errors,
    testSource,
    [
      "getByRole('heading'",
      "getByRole('combobox', { name: 'Traffic region' })",
      "getByRole('region', { name: 'Live traffic data controls' })",
      "getByLabel('Open live data controls')",
      'waitForMapFirstHydration',
      'page.setViewportSize({ width: 390, height: 844 })',
      'region=WORLD&view=invalid',
      'region=world&view=aircraft',
      'region=az&view=aircraft',
      "name: 'Aircraft search results'",
      "name: 'Live flight tracker'",
      "name: 'Map tools'",
      "name: 'Compare persisted analytical evidence'",
      "name: 'Projection confidence score'",
      "name: 'Mean forecast stability score'",
      "name: 'Download CSV'",
      "name: 'Download GeoJSON'",
      "name: 'Skip to main content'",
      'testInfo.attach',
      'boundingBox()',
    ],
    'Playwright product tests',
  )
  if (testSource.includes('data-testid') || testSource.includes('page.locator(')) {
    fail(errors, 'Playwright tests must prefer semantic locators')
  }
  if (/onrender\.com|vercel\.app/.test(testSource)) {
    fail(errors, 'Playwright tests must not target a public deployment')
  }

  const aircraftTests = read(
    root,
    'apps/web/e2e/tests/aircraft-intelligence.spec.mjs',
  )
  requireIncludes(
    errors,
    aircraftTests,
    [
      'Aircraft search results',
      'Probable route and airport context',
      'Latest trajectory',
      'Projection and Estimated Arrival',
      'Weather Context',
      'Stability and Explainability',
      'Clear selection',
    ],
    'aircraft journey',
  )

  const airportTests = read(
    root,
    'apps/web/e2e/tests/airport-intelligence.spec.mjs',
  )
  requireIncludes(
    errors,
    airportTests,
    [
      'aircraft-only frontend does not mount or link the legacy Airport Intelligence workspace',
      "name: 'Live flight tracker'",
      "name: 'Current traffic map focused on World'",
      "name: 'Airport Intelligence'",
      'toHaveCount(0)',
    ],
    'aircraft-only frontend boundary journey',
  )

  const historicalTests = read(
    root,
    'apps/web/e2e/tests/historical-analytics.spec.mjs',
  )
  requireIncludes(
    errors,
    historicalTests,
    [
      'Compare persisted analytical evidence',
      'historical bar chart',
      'Complete the historical scope',
      'Airport ICAO',
      'Origin ICAO',
      'Destination ICAO',
    ],
    'historical journey',
  )

  const advancedTests = read(
    root,
    'apps/web/e2e/tests/advanced-intelligence.spec.mjs',
  )
  requireIncludes(
    errors,
    advancedTests,
    [
      'Projection confidence score',
      'Weather Encounter Profile',
      'Projection uncertainty effect',
      'Mean forecast stability score',
      'Attribution and scope guards',
    ],
    'advanced intelligence journey',
  )

  const exportTests = read(root, 'apps/web/e2e/tests/exports.spec.mjs')
  requireIncludes(
    errors,
    exportTests,
    [
      'Download CSV',
      'Download GeoJSON',
      'FeatureCollection',
      'excluded_invalid_coordinate_count',
      '4b1801,AZAL101',
    ],
    'export journey',
  )

  const recoveryTests = read(
    root,
    'apps/web/e2e/tests/product-recovery.spec.mjs',
  )
  requireIncludes(
    errors,
    recoveryTests,
    [
      'regions-error',
      'aircraft-error',
      'airport-error',
      'historical-error',
      'intelligence-error',
      'Retry route context',
      'Retry trajectory',
      'aircraft-only frontend does not issue hidden Airport Intelligence requests',
      'Retry Projection Intelligence',
      'Retry Weather Context',
      'Retry Stability Intelligence',
    ],
    'product recovery journeys',
  )

  const accessibilityTests = read(
    root,
    'apps/web/e2e/tests/accessibility.spec.mjs',
  )
  requireIncludes(
    errors,
    accessibilityTests,
    [
      'Skip to main content',
      'Tracker tools',
      'Map tools',
      'Mobile primary navigation',
      'aria-selected',
      'expectNoHorizontalOverflow',
    ],
    'accessibility journey',
  )

  const visualTests = read(
    root,
    'apps/web/e2e/tests/visual-regression.spec.mjs',
  )
  requireIncludes(
    errors,
    visualTests,
    [
      'width: 1440, height: 1100',
      'width: 390, height: 844',
      'boundingBox()',
      'expectNoHorizontalOverflow',
      "testInfo.attach('desktop-product-workspace.png'",
      "testInfo.attach('mobile-selected-aircraft-workspace.png'",
      'page.screenshot({',
    ],
    'visual layout regression',
  )
  if (visualTests.includes('toHaveScreenshot(')) {
    fail(
      errors,
      'pre-redesign visual coverage must not fossilize current pixels with golden baselines',
    )
  }

  const mockSource = read(root, 'apps/web/e2e/mock-api.mjs')
  requireIncludes(
    errors,
    mockSource,
    [
      "scenario === 'traffic-error'",
      "scenario === 'regions-error'",
      "scenario === 'aircraft-error'",
      "scenario === 'airport-error'",
      "scenario === 'historical-error'",
      "scenario === 'intelligence-error'",
      "'/__e2e/scenario'",
      "'Access-Control-Allow-Origin': 'http://127.0.0.1:3000'",
      "'Cache-Control': 'no-store'",
      "'/api/v1/traffic/current'",
      "'/api/v1/aircraft/{icao24}/trajectory'",
      "'/api/v1/metrics/active-aircraft'",
      "'/api/v1/analytics/metrics/coverage-score'",
      "'/api/v1/historical-intelligence/aggregates/history'",
      "'/api/v1/trajectories/{id}/projection-intelligence'",
      "'/api/v1/trajectories/{id}/stability-intelligence'",
      'stabilityIntelligenceFixture(requestedAsOfTimes)',
      "url.searchParams.get('as_of_times')",
      "'/api/v1/airspace/regions/{code}/analytics'",
      "'/api/v1/trajectories/{id}/route-intelligence'",
      "'/api/v1/trajectories/{id}/route-intelligence/latest'",
      "'/api/v1/trajectories/{id}/route-intelligence/history'",
      "'X-Internal-API-Key'",
    ],
    'mock API',
  )

  const packageJSON = JSON.parse(read(root, 'package.json'))
  const expectedScripts = {
    'test:playwright-e2e-contract':
      'node --test scripts/verify-playwright-e2e.test.mjs',
    'verify:playwright-e2e':
      'node scripts/verify-playwright-e2e.mjs',
    'run:playwright-e2e': 'bash scripts/run-playwright-e2e.sh',
  }
  for (const [name, command] of Object.entries(expectedScripts)) {
    if (packageJSON.scripts?.[name] !== command) {
      fail(errors, `package script is missing or incorrect: ${name}`)
    }
  }

  const workflow = read(root, '.github/workflows/playwright-e2e.yml')
  requireIncludes(
    errors,
    workflow,
    [
      'name: Playwright E2E',
      "PLAYWRIGHT_VERSION: '1.62.0'",
      'pnpm install --frozen-lockfile',
      'node --test scripts/verify-playwright-e2e.test.mjs',
      'bash scripts/run-playwright-e2e.sh',
      'apps/web/e2e/playwright-report',
      'apps/web/e2e/test-results',
      'docs/186_PLAYWRIGHT_PRODUCT_COVERAGE.md',
      'actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a',
    ],
    'Playwright workflow',
  )

  const release = read(root, 'scripts/verify-release.sh')
  requireIncludes(
    errors,
    release,
    [
      'pnpm run test:playwright-e2e-contract',
      'pnpm run verify:playwright-e2e',
      'RELEASE_PLAYWRIGHT_E2E_CONTRACT=PASS',
    ],
    'release gate',
  )
  if (release.includes('pnpm run run:playwright-e2e')) {
    fail(errors, 'full browser suite must remain outside the general release gate')
  }

  const gitignore = read(root, '.gitignore')
  requireIncludes(
    errors,
    gitignore,
    [
      '/apps/web/e2e/test-results/',
      '/apps/web/e2e/playwright-report/',
    ],
    '.gitignore',
  )

  const foundationDocument = read(
    root,
    'docs/176_PLAYWRIGHT_E2E_FOUNDATION.md',
  )
  requireIncludes(
    errors,
    foundationDocument,
    [
      'Status: IMPLEMENTED',
      'original foundation contained four Chromium end-to-end scenarios',
      'OpenAPI',
      'PLAYWRIGHT_E2E=PASS',
      'public Render deployment is never targeted',
      'thirty-eight paths',
      'Route Intelligence',
      '186_PLAYWRIGHT_PRODUCT_COVERAGE.md',
    ],
    'Document 176',
  )

  const coverageDocument = read(
    root,
    'docs/186_PLAYWRIGHT_PRODUCT_COVERAGE.md',
  )
  requireIncludes(
    errors,
    coverageDocument,
    [
      'Status: IMPLEMENTED',
      'twenty deterministic Chromium browser scenarios',
      'PLAYWRIGHT_E2E_MOCK_SCENARIOS=7',
      'PLAYWRIGHT_E2E_BROWSER_SCENARIOS=20',
      'Pixel-golden `toHaveScreenshot` baselines are intentionally deferred',
      'This closes the current functional Playwright product-journey debt.',
    ],
    'Document 186',
  )

  const documentIndex = read(root, 'docs/DOCUMENT_INDEX.md')
  requireIncludes(
    errors,
    documentIndex,
    [
      'PLAYWRIGHT-E2E-FOUNDATION-V1:DOCUMENT-INDEX',
      '176_PLAYWRIGHT_E2E_FOUNDATION.md',
      'PLAYWRIGHT-PRODUCT-COVERAGE-V1:DOCUMENT-INDEX',
      '186_PLAYWRIGHT_PRODUCT_COVERAGE.md',
    ],
    'documentation index',
  )

  const healthyTraffic = resolveMockResponse({
    method: 'GET',
    requestURL: 'http://127.0.0.1:8091/api/v1/traffic/current?region=world',
    scenario: 'healthy',
  })
  if (
    healthyTraffic.status !== 200 ||
    healthyTraffic.body.success !== true ||
    healthyTraffic.body.data.length !== 2
  ) {
    fail(errors, 'healthy mock traffic response is invalid')
  }

  const failedTraffic = resolveMockResponse({
    method: 'GET',
    requestURL: 'http://127.0.0.1:8091/api/v1/traffic/current?region=world',
    scenario: 'traffic-error',
  })
  if (
    failedTraffic.status !== 503 ||
    failedTraffic.body.success !== false
  ) {
    fail(errors, 'traffic-error mock response is invalid')
  }

  return errors
}

function main() {
  const errors = validatePlaywrightFoundation(process.cwd())
  if (errors.length > 0) {
    for (const error of errors) {
      console.error(`PLAYWRIGHT_E2E_CONTRACT=FAIL reason=${error}`)
    }
    process.exit(1)
  }

  console.log(`PLAYWRIGHT_E2E_VERSION=${expectedPlaywrightVersion}`)
  console.log('PLAYWRIGHT_E2E_OPENAPI_PATHS=38')
  console.log(`PLAYWRIGHT_E2E_MOCK_SCENARIOS=${expectedScenarios.size}`)
  console.log(`PLAYWRIGHT_E2E_BROWSER_SCENARIOS=${expectedBrowserScenarioCount}`)
  console.log(`PLAYWRIGHT_E2E_SCENARIOS=${expectedBrowserScenarioCount}`)
  console.log('PLAYWRIGHT_E2E_PRODUCT_COVERAGE=PASS')
  console.log('PLAYWRIGHT_E2E_VISUAL_LAYOUT_REGRESSION=PASS')
  console.log('PLAYWRIGHT_E2E_MOCK_API=PASS')
  console.log('PLAYWRIGHT_E2E_CONTRACT=PASS')
}

const currentFile = fileURLToPath(import.meta.url)
if (process.argv[1] && path.resolve(process.argv[1]) === currentFile) {
  main()
}
