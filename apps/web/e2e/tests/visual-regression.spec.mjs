import { expect, test } from '@playwright/test'

import {
  expectNoHorizontalOverflow,
  setScenario,
} from './helpers.mjs'

test.beforeEach(async ({ request }) => {
  await setScenario(request, 'healthy')
})

test('desktop map-first workspace stays dominant before analytical sections', async ({
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
  const liveMap = page.getByRole('region', {
    name: 'Current traffic map focused on World',
  })
  const trafficWorkspace = page.getByRole('complementary', {
    name: 'Traffic workspace',
  })
  const airport = page.getByRole('region', { name: 'Airport Intelligence' })
  const historical = page.getByRole('region', {
    name: 'Compare persisted analytical evidence',
  })

  await expect(liveHeading).toBeVisible()
  await expect(liveMap).toBeVisible()
  await expect(trafficWorkspace).toBeVisible()
  await expect(airport).toBeVisible()
  await expect(historical).toBeVisible()

  const headingBox = await liveHeading.boundingBox()
  const mapBox = await liveMap.boundingBox()
  const workspaceBox = await trafficWorkspace.boundingBox()
  const airportBox = await airport.boundingBox()
  const historicalBox = await historical.boundingBox()

  expect(headingBox).not.toBeNull()
  expect(mapBox).not.toBeNull()
  expect(workspaceBox).not.toBeNull()
  expect(airportBox).not.toBeNull()
  expect(historicalBox).not.toBeNull()

  expect(headingBox.y).toBeLessThan(180)
  expect(mapBox.width).toBeGreaterThan(700)
  expect(mapBox.height).toBeGreaterThan(560)
  expect(workspaceBox.width).toBeGreaterThan(320)
  expect(airportBox.width).toBeGreaterThan(900)
  expect(historicalBox.width).toBeGreaterThan(900)
  expect(airportBox.y).toBeGreaterThan(mapBox.y)
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

test('mobile map-first workspace preserves intelligence flow and screenshot evidence', async ({
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
    page.getByRole('region', {
      name: 'Current traffic map focused on Azerbaijan',
    }),
  ).toBeVisible()
  await expect(
    page.getByRole('heading', {
      name: 'Projection and Estimated Arrival',
    }),
  ).toBeVisible()
  await expectNoHorizontalOverflow(page)

  const mapBox = await page
    .getByRole('region', {
      name: 'Current traffic map focused on Azerbaijan',
    })
    .boundingBox()
  expect(mapBox).not.toBeNull()
  expect(mapBox.width).toBeLessThanOrEqual(390)
  expect(mapBox.height).toBeGreaterThan(420)

  await testInfo.attach('mobile-selected-aircraft-workspace.png', {
    body: await page.screenshot({
      fullPage: true,
      animations: 'disabled',
    }),
    contentType: 'image/png',
  })
})
