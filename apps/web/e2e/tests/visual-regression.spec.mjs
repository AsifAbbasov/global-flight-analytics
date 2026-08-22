import { expect, test } from '@playwright/test'

import {
  expectNoHorizontalOverflow,
  setScenario,
} from './helpers.mjs'

test.beforeEach(async ({ request }) => {
  await setScenario(request, 'healthy')
})

test('desktop map-first section order and width remain structurally stable', async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 1440, height: 1100 })
  await page.goto(
    '/?region=world&aircraft=4b1801&view=intelligence#live-traffic',
    { waitUntil: 'domcontentloaded' },
  )

  const liveHeading = page.getByRole('heading', {
    name: 'Current Traffic — World',
  })
  const overview = page.locator('#overview')
  const airport = page.getByRole('region', { name: 'Airport Intelligence' })
  const historical = page.getByRole('region', {
    name: 'Compare persisted analytical evidence',
  })

  const liveBox = await liveHeading.boundingBox()
  const overviewBox = await overview.boundingBox()
  const airportBox = await airport.boundingBox()
  const historicalBox = await historical.boundingBox()
  expect(liveBox).not.toBeNull()
  expect(overviewBox).not.toBeNull()
  expect(airportBox).not.toBeNull()
  expect(historicalBox).not.toBeNull()
  expect(airportBox.width).toBeGreaterThan(900)
  expect(historicalBox.width).toBeGreaterThan(900)
  expect(overviewBox.y).toBeGreaterThan(liveBox.y)
  expect(airportBox.y).toBeGreaterThan(overviewBox.y)
  expect(historicalBox.y).toBeGreaterThan(airportBox.y)

  await expectNoHorizontalOverflow(page)

  await testInfo.attach('desktop-product-workspace.png', {
    body: await page.screenshot({
      fullPage: true,
      animations: 'disabled',
    }),
    contentType: 'image/png',
  })
})

test('mobile workspace preserves vertical flow and records screenshot evidence', async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto(
    '/?region=az&aircraft=4b1801&view=intelligence#live-traffic',
    { waitUntil: 'domcontentloaded' },
  )

  await expect(
    page.getByRole('heading', { name: 'Current Traffic — Azerbaijan' }),
  ).toBeVisible()
  await expect(
    page.getByRole('heading', {
      name: 'Projection and Estimated Arrival',
    }),
  ).toBeVisible()
  await expectNoHorizontalOverflow(page)

  await testInfo.attach('mobile-selected-aircraft-workspace.png', {
    body: await page.screenshot({
      fullPage: true,
      animations: 'disabled',
    }),
    contentType: 'image/png',
  })
})
