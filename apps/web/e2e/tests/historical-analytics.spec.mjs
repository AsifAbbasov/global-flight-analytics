import { expect, test } from '@playwright/test'

import { setScenario } from './helpers.mjs'

test.beforeEach(async ({ request }) => {
  await setScenario(request, 'healthy')
})

test('historical workspace preserves persisted global evidence and bounded scope inputs', async ({
  page,
}) => {
  await page.goto('/#historical-analytics', {
    waitUntil: 'domcontentloaded',
  })

  const workspace = page.getByRole('region', {
    name: 'Compare persisted analytical evidence',
  })
  await expect(workspace).toBeVisible()
  await expect(
    workspace.getByRole('img', { name: /historical bar chart/i }),
  ).toBeVisible()
  await expect(
    workspace.getByRole('heading', { name: 'Evidence summary' }),
  ).toBeVisible()

  const scope = workspace.getByRole('combobox', { name: 'Scope' })
  await scope.selectOption('airport')
  await expect(
    workspace.getByText('Complete the historical scope', { exact: true }),
  ).toBeVisible()
  await expect(
    workspace.getByRole('textbox', { name: 'Airport ICAO' }),
  ).toBeVisible()

  await scope.selectOption('route')
  await expect(
    workspace.getByRole('textbox', { name: 'Origin ICAO' }),
  ).toBeVisible()
  await expect(
    workspace.getByRole('textbox', { name: 'Destination ICAO' }),
  ).toBeVisible()

  await scope.selectOption('global')
  await expect(
    workspace.getByRole('img', { name: /historical bar chart/i }),
  ).toBeVisible()
})
