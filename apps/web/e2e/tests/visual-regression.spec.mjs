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

  const tracker = page.getByRole('region', { name: 'Live flight tracker' })
  const liveMap = page.getByRole('region', {
    name: 'Current traffic map focused on World',
  })
  const trafficWorkspace = page.getByRole('complementary', {
    name: 'Traffic workspace',
  })
  const mapTools = page.getByRole('navigation', { name: 'Map tools' })

  await expect(tracker).toBeVisible()
  await expect(liveMap).toBeVisible()
  await expect(trafficWorkspace).toBeVisible()
  await expect(mapTools).toBeVisible()
  await expect(
    page.getByRole('region', { name: 'Airport Intelligence' }),
  ).toHaveCount(0)

  const trackerBox = await tracker.boundingBox()
  const mapBox = await liveMap.boundingBox()
  const workspaceBox = await trafficWorkspace.boundingBox()

  expect(trackerBox).not.toBeNull()
  expect(mapBox).not.toBeNull()
  expect(workspaceBox).not.toBeNull()

  expect(trackerBox.y).toBeLessThan(180)
  expect(mapBox.width).toBeGreaterThan(700)
  expect(mapBox.height).toBeGreaterThan(560)
  expect(workspaceBox.width).toBeGreaterThan(320)
  expect(mapBox.x).toBeGreaterThanOrEqual(workspaceBox.x + workspaceBox.width - 2)

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
    page.getByRole('region', { name: 'Live flight tracker' }),
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
