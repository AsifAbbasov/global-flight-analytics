import { expect, test } from '@playwright/test'

import {
  setScenario,
  waitForMapFirstHydration,
} from './helpers.mjs'

test('region catalog failure preserves World and recovers on reload', async ({
  page,
  request,
}) => {
  await setScenario(request, 'regions-error')
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  await waitForMapFirstHydration(page)

  await page
    .getByLabel('Open live data controls')
    .click()
  const liveControls = page.getByRole('region', {
    name: 'Live traffic data controls',
  })
  await expect(liveControls).toBeVisible()
  await expect(
    liveControls.getByText(
      'The region list is temporarily unavailable. World view remains available; reload the page to retry.',
      { exact: true },
    ),
  ).toBeVisible()
  const region = page.getByRole('combobox', { name: 'Traffic region' })
  await expect(region).toHaveValue('world')

  await setScenario(request, 'healthy')
  await page.reload({ waitUntil: 'domcontentloaded' })

  await expect(region).toContainText('Azerbaijan')
  await expect(region).toContainText('Turkey')
})

test('aircraft profile route context and trajectory recover independently', async ({
  page,
  request,
}) => {
  await setScenario(request, 'aircraft-error')
  await page.goto(
    '/?region=az&aircraft=4b1801&view=intelligence#live-traffic',
    { waitUntil: 'domcontentloaded' },
  )

  const routeRetry = page.getByRole('button', {
    name: 'Retry route context',
  })
  const trajectoryRetry = page.getByRole('button', {
    name: 'Retry trajectory',
  })
  const profileRetry = page.getByRole('button', { name: 'Retry profile' })
  await expect(routeRetry).toBeVisible()
  await expect(trajectoryRetry).toBeVisible()
  await expect(profileRetry).toBeVisible()

  await setScenario(request, 'healthy')

  await routeRetry.click()
  await expect(
    page.getByRole('progressbar', {
      name: 'Probable route confidence score',
    }),
  ).toBeVisible()

  await trajectoryRetry.click()
  await expect(
    page.getByRole('progressbar', { name: 'Track quality score' }),
  ).toBeVisible()

  await profileRetry.click()
  await expect(page.getByText('4K-AZ01', { exact: true })).toBeVisible()
})

test('aircraft-only frontend does not issue hidden Airport Intelligence requests', async ({
  page,
  request,
}) => {
  const airportRequests = []
  page.on('request', browserRequest => {
    if (browserRequest.url().includes('/api/v1/airports/intelligence')) {
      airportRequests.push(browserRequest.url())
    }
  })

  await setScenario(request, 'airport-error')
  await page.goto('/?region=world&view=aircraft#live-traffic', {
    waitUntil: 'domcontentloaded',
  })

  await expect(
    page.getByRole('region', { name: 'Live flight tracker' }),
  ).toBeVisible()
  await expect(
    page.getByRole('heading', { name: 'Aircraft Explorer' }),
  ).toBeVisible()
  await expect(
    page.getByRole('region', { name: 'Airport Intelligence' }),
  ).toHaveCount(0)
  expect(airportRequests).toEqual([])
})

test('Historical Intelligence latest aggregate recovers through its visible Retry action', async ({
  page,
  request,
}) => {
  await setScenario(request, 'historical-error')
  await page.goto('/#historical-analytics', {
    waitUntil: 'domcontentloaded',
  })

  const workspace = page.getByRole('region', {
    name: 'Compare persisted analytical evidence',
  })
  await expect(
    workspace.getByText('Latest aggregate unavailable', { exact: true }),
  ).toBeVisible()

  await setScenario(request, 'healthy')
  await workspace.getByRole('button', { name: 'Retry' }).first().click()

  await expect(
    workspace.getByRole('img', { name: /historical bar chart/i }),
  ).toBeVisible()
})

test('advanced intelligence panels recover from one deterministic outage', async ({
  page,
  request,
}) => {
  await setScenario(request, 'intelligence-error')
  await page.goto(
    '/?region=az&aircraft=4b1801&view=intelligence#live-traffic',
    { waitUntil: 'domcontentloaded' },
  )

  const projectionRetry = page.getByRole('button', {
    name: 'Retry Projection Intelligence',
  })
  const weatherRetry = page.getByRole('button', {
    name: 'Retry Weather Context',
  })
  const stabilityRetry = page.getByRole('button', {
    name: 'Retry Stability Intelligence',
  })
  await expect(projectionRetry).toBeVisible()
  await expect(weatherRetry).toBeVisible()
  await expect(stabilityRetry).toBeVisible()

  await setScenario(request, 'healthy')

  await projectionRetry.click()
  await expect(
    page.getByRole('progressbar', { name: 'Projection confidence score' }),
  ).toBeVisible()

  await weatherRetry.click()
  await expect(
    page.getByRole('heading', { name: 'Weather Encounter Profile' }),
  ).toBeVisible()

  await stabilityRetry.click()
  await expect(
    page.getByRole('progressbar', { name: 'Mean forecast stability score' }),
  ).toBeVisible()
})
