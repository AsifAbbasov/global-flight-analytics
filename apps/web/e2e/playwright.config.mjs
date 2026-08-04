/* global process */
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { defineConfig, devices } from '@playwright/test'

const e2eRoot = path.dirname(fileURLToPath(import.meta.url))
const webRoot = path.resolve(e2eRoot, '..')
const appOrigin =
  process.env.PLAYWRIGHT_APP_ORIGIN ?? 'http://127.0.0.1:3000'
const apiOrigin =
  process.env.PLAYWRIGHT_API_ORIGIN ?? 'http://127.0.0.1:8091'
const isCI = process.env.CI === 'true'

export default defineConfig({
  testDir: path.join(e2eRoot, 'tests'),
  outputDir: path.join(e2eRoot, 'test-results'),
  fullyParallel: false,
  workers: 1,
  retries: isCI ? 1 : 0,
  timeout: 30_000,
  expect: {
    timeout: 7_500,
  },
  reporter: [
    ['line'],
    [
      'html',
      {
        open: 'never',
        outputFolder: path.join(e2eRoot, 'playwright-report'),
      },
    ],
  ],
  use: {
    baseURL: appOrigin,
    headless: true,
    locale: 'en-US',
    timezoneId: 'UTC',
    reducedMotion: 'reduce',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
      },
    },
  ],
  webServer: [
    {
      command: 'node mock-api.mjs',
      cwd: e2eRoot,
      url: `${apiOrigin}/api/v1/health`,
      reuseExistingServer: !isCI,
      timeout: 30_000,
    },
    {
      command: 'pnpm start --hostname 127.0.0.1 --port 3000',
      cwd: webRoot,
      env: {
        ...process.env,
        NEXT_PUBLIC_API_BASE_URL: apiOrigin,
      },
      url: appOrigin,
      reuseExistingServer: !isCI,
      timeout: 120_000,
    },
  ],
})
