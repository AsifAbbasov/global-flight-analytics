import { expect, test } from '@playwright/test'

import { setScenario } from './helpers.mjs'

test.beforeEach(async ({ request }) => {
  await setScenario(request, 'healthy')
})

test('airport ranking opens passport history and trend evidence', async ({
  page,
}) => {
  await page.goto('/#airport-intelligence', {
    waitUntil: 'domcontentloaded',
  })

  const workspace = page.getByRole('region', { name: 'Airport Intelligence' })
  await expect(workspace).toBeVisible()

  const search = workspace.getByRole('searchbox', { name: 'Search airport' })
  await search.fill('UBBB')

  const airport = workspace.getByRole('button', { name: /UBBB/i })
  await expect(airport).toBeVisible()
  await airport.click()

  await expect(
    workspace.getByRole('heading', {
      name: 'Heydar Aliyev International Airport',
    }),
  ).toBeVisible()
  await expect(
    workspace.getByText('Digital passport', { exact: true }),
  ).toBeVisible()
  const profileTabs = workspace.getByRole('tablist', {
    name: 'Airport Intelligence profile sections',
  })
  await expect(
    profileTabs.getByRole('tab', { name: 'Overview' }),
  ).toHaveAttribute('aria-selected', 'true')

  await profileTabs.getByRole('tab', { name: 'History' }).click()
  await expect(
    workspace.getByRole('heading', { name: 'Movement history' }),
  ).toBeVisible()
  await expect(
    workspace.getByRole('columnheader', { name: 'Movements' }),
  ).toBeVisible()

  await profileTabs.getByRole('tab', { name: 'Trends' }).click()
  await expect(
    workspace.getByText('Published trend direction', { exact: true }),
  ).toBeVisible()
  await expect(
    workspace.getByRole('heading', { name: 'Evidence movement' }),
  ).toBeVisible()

  await workspace
    .getByRole('combobox', { name: 'Completed-day window' })
    .selectOption('7')
  await expect(
    workspace.getByText('Completed 7-day analytical window', { exact: true }),
  ).toBeVisible()
})
