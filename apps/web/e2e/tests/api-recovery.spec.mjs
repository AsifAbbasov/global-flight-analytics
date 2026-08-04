/* global process */
import { expect, test } from '@playwright/test'

const mockAPIOrigin =
  process.env.PLAYWRIGHT_API_ORIGIN ?? 'http://127.0.0.1:8091'

async function setScenario(request, scenario) {
  const response = await request.post(`${mockAPIOrigin}/__e2e/scenario`, {
    data: { scenario },
  })
  expect(response.ok()).toBeTruthy()
}

test('traffic failure preserves the shell and recovers through Retry', async ({
  page,
  request,
}) => {
  await setScenario(request, 'traffic-error')
  await page.goto('/', { waitUntil: 'domcontentloaded' })

  await expect(
    page.getByText('Initial API snapshot unavailable', { exact: true }),
  ).toBeVisible()
  await expect(
    page.getByRole('heading', { level: 1 }),
  ).toContainText('Observe aviation data.')

  const liveControls = page.getByRole('region', {
    name: 'Live traffic data controls',
  })
  const retry = liveControls.getByRole('button', {
    name: 'Retry traffic request',
  })
  await expect(retry).toBeVisible()
  await expect(
    liveControls.getByText('Traffic snapshot unavailable', {
      exact: true,
    }),
  ).toBeVisible()

  await setScenario(request, 'healthy')
  await retry.click()

  await expect(
    liveControls.getByText('Snapshot current', { exact: true }),
  ).toBeVisible()
  await expect(
    liveControls.getByText('2 aircraft', { exact: true }),
  ).toBeVisible()
})
