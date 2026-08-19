/* global process */
import { expect, test } from '@playwright/test'

import { waitForMapFirstHydration } from './helpers.mjs'

const mockAPIOrigin =
  process.env.PLAYWRIGHT_API_ORIGIN ?? 'http://127.0.0.1:8091'

async function setScenario(request, scenario) {
  const response = await request.post(`${mockAPIOrigin}/__e2e/scenario`, {
    data: { scenario },
  })
  expect(response.ok()).toBeTruthy()
}

test('traffic failure preserves the map-first shell and recovers through Retry', async ({
  page,
  request,
}) => {
  await setScenario(request, 'traffic-error')
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  await waitForMapFirstHydration(page)

  await expect(
    page.getByText('Initial API snapshot unavailable', { exact: true }),
  ).toBeVisible()
  await expect(
    page.getByRole('region', { name: 'Live flight tracker' }),
  ).toBeVisible()

  const alert = page.getByRole('alert')
  await expect(
    alert.getByText('Live data unavailable', { exact: true }),
  ).toBeVisible()
  const retry = alert.getByRole('button', { name: 'Retry' })
  await expect(retry).toBeVisible()

  await setScenario(request, 'healthy')
  await retry.click()

  await page
    .getByLabel('Open live data controls')
    .click()
  const liveControls = page.getByRole('region', {
    name: 'Live traffic data controls',
  })
  await expect(
    liveControls.getByText('Snapshot current', { exact: true }),
  ).toBeVisible()
  await expect(
    liveControls.getByText('2 aircraft', { exact: true }),
  ).toBeVisible()
})
