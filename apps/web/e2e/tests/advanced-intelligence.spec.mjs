import { expect, test } from '@playwright/test'

import { setScenario } from './helpers.mjs'

test.beforeEach(async ({ request }) => {
  await setScenario(request, 'healthy')
})

test('projection weather and stability expose server-owned evidence semantics', async ({
  page,
}) => {
  await page.goto(
    '/?region=az&aircraft=4b1801&view=intelligence#live-traffic',
    { waitUntil: 'domcontentloaded' },
  )

  await expect(
    page.getByRole('progressbar', { name: 'Projection confidence score' }),
  ).toHaveAttribute('aria-valuenow', '72')
  await expect(
    page.getByText('Unavailable. Backend status: unavailable.', {
      exact: true,
    }),
  ).toBeVisible()
  await expect(
    page.getByText('Projection is research-only.', { exact: true }),
  ).toBeVisible()

  await expect(
    page.getByRole('heading', { name: 'Weather Encounter Profile' }),
  ).toBeVisible()
  await expect(
    page.getByText('Projection uncertainty effect', { exact: true }),
  ).toBeVisible()
  await expect(
    page.getByText(
      'Context and uncertainty modifier, never proof of maneuver cause.',
      { exact: true },
    ),
  ).toBeVisible()

  await expect(
    page.getByRole('progressbar', { name: 'Mean forecast stability score' }),
  ).toHaveAttribute('aria-valuenow', '91')
  await expect(
    page.getByText(
      'A stable forecast may still be inaccurate. A changed forecast may be an improvement.',
      { exact: true },
    ),
  ).toBeVisible()
  await expect(
    page.getByRole('heading', { name: 'Attribution and scope guards' }),
  ).toBeVisible()
})
