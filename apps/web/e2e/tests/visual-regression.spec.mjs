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
  const overviewHeading = page.getByRole('heading', {
    name: 'Live Analytics — World',
  })
  const airport = page.getByRole('region', { name: 'Airport Intelligence' })
  const historical = page.getByRole('region', {
    name: 'Compare persisted analytical evidence',
  })
  const mapEvidenceControls = page.getByRole('group', {
    name: 'Map evidence visibility',
  })

  // The page is intentionally hydrated after the initial document response.
  // Resolve geometry from the current semantic element after visibility instead
  // of retaining a nullable bounding-box snapshot across dynamic workspace remounts.
  const liveBox = await visibleDocumentBox(liveHeading)
  const overviewBox = await visibleDocumentBox(overviewHeading)
  const airportBox = await visibleDocumentBox(airport)
  const historicalBox = await visibleDocumentBox(historical)
  await expect(mapEvidenceControls).toBeVisible()

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

test('mobile workspace preserves responsive map flow and records screenshot evidence', async ({
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

  const map = page.getByLabel('Current traffic map focused on Azerbaijan')
  const mapEvidenceControls = page.getByRole('group', {
    name: 'Map evidence visibility',
  })
  const mapBox = await visibleDocumentBox(map)
  await expect(mapEvidenceControls).toBeVisible()
  await expect(page.getByRole('button', { name: 'Trail' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Projection' })).toBeVisible()

  expect(mapBox.height).toBeGreaterThanOrEqual(380)
  expect(mapBox.height).toBeLessThan(600)
  expect(mapBox.width).toBeLessThanOrEqual(390)

  await expectNoHorizontalOverflow(page)

  await testInfo.attach('mobile-selected-aircraft-workspace.png', {
    body: await page.screenshot({
      fullPage: true,
      animations: 'disabled',
    }),
    contentType: 'image/png',
  })
})

async function visibleDocumentBox(locator) {
  await expect(locator).toBeVisible()
  return locator.evaluate(element => {
    const rect = element.getBoundingClientRect()
    return {
      width: rect.width,
      height: rect.height,
      y: rect.top + window.scrollY,
    }
  })
}
