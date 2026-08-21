/* global URL */
import { expect, test } from '@playwright/test'

import { setScenario } from './helpers.mjs'

test.beforeEach(async ({ request }) => {
  await setScenario(request, 'healthy')
})

test('aircraft explorer selection opens the complete intelligence workspace', async ({
  page,
}) => {
  await page.goto('/?region=world&view=aircraft#live-traffic', {
    waitUntil: 'domcontentloaded',
  })

  const search = page.getByRole('searchbox', { name: 'Search aircraft' })
  await search.fill('AZAL101')

  const results = page.getByRole('list', { name: 'Aircraft search results' })
  const aircraft = results.getByRole('button', { name: /AZAL101/i })
  await expect(aircraft).toBeVisible()
  await aircraft.click()

  const url = new URL(page.url())
  expect(url.searchParams.get('region')).toBe('world')
  expect(url.searchParams.get('aircraft')).toBe('4b1801')
  expect(url.searchParams.get('view')).toBe('intelligence')

  await expect(
    page.getByRole('tab', { name: /Intelligence/i }),
  ).toHaveAttribute('aria-selected', 'true')
  await expect(page.getByRole('heading', { name: 'AZAL101' })).toBeVisible()
  await expect(
    page.getByRole('heading', {
      name: 'Probable route and airport context',
    }),
  ).toBeVisible()
  await expect(
    page.getByRole('heading', { name: 'Latest trajectory' }),
  ).toBeVisible()
  await expect(
    page.getByRole('heading', { name: 'Registered profile' }),
  ).toBeVisible()
  await expect(
    page.getByRole('progressbar', { name: 'Track quality score' }),
  ).toHaveAttribute('aria-valuenow', '96')
  await expect(
    page.getByRole('heading', {
      name: 'Projection and Estimated Arrival',
    }),
  ).toBeVisible()
  await expect(
    page.getByRole('heading', { name: 'Weather Context' }),
  ).toBeVisible()
  await expect(
    page.getByRole('heading', {
      name: 'Stability and Explainability',
    }),
  ).toBeVisible()
})

test('aircraft deep link restores intelligence and clearing selection returns to explorer', async ({
  page,
}) => {
  await page.goto(
    '/?region=AZ&aircraft=4B1801&view=intelligence#live-traffic',
    { waitUntil: 'domcontentloaded' },
  )

  await expect(page).toHaveURL(
    /\/\?region=az&aircraft=4b1801&view=intelligence#live-traffic$/,
  )

  const intelligenceTab = page.getByRole('tab', { name: /Intelligence/i })
  await expect(intelligenceTab).toHaveAttribute('aria-selected', 'true')

  const intelligencePanel = page.getByRole('tabpanel')
  await expect(
    intelligencePanel.getByText('Selected aircraft', { exact: true }),
  ).toBeVisible()
  await expect(intelligencePanel.getByText(/^4b1801$/i)).toBeVisible()

  const clearSelection = intelligencePanel.getByRole('button', {
    name: 'Clear selection',
  })
  await expect(clearSelection).toBeVisible()
  await clearSelection.click()

  const url = new URL(page.url())
  expect(url.searchParams.get('region')).toBe('az')
  expect(url.searchParams.get('aircraft')).toBeNull()
  expect(url.searchParams.get('view')).toBe('aircraft')
  await expect(
    page.getByRole('heading', { name: 'Aircraft Explorer' }),
  ).toBeVisible()
})
