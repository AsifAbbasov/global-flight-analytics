import { expect, test } from '@playwright/test'

import { setScenario } from './helpers.mjs'

test.beforeEach(async ({ request }) => {
  await setScenario(request, 'healthy')
})

test('aircraft-only frontend does not mount or link the legacy Airport Intelligence workspace', async ({
  page,
}) => {
  await page.goto('/#airport-intelligence', {
    waitUntil: 'domcontentloaded',
  })

  await expect(
    page.getByRole('region', { name: 'Live flight tracker' }),
  ).toBeVisible()
  await expect(
    page.getByRole('region', {
      name: 'Current traffic map focused on World',
    }),
  ).toBeVisible()
  await expect(
    page.getByRole('region', { name: 'Airport Intelligence' }),
  ).toHaveCount(0)
  await expect(
    page.getByRole('link', { name: /Airport Intelligence/i }),
  ).toHaveCount(0)
})
