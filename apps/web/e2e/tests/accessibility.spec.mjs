import { expect, test } from '@playwright/test'

import {
  expectNoHorizontalOverflow,
  setScenario,
} from './helpers.mjs'

test.beforeEach(async ({ request }) => {
  await setScenario(request, 'healthy')
})

test('skip links landmarks and major research sections remain keyboard reachable', async ({
  page,
}) => {
  await page.goto('/', { waitUntil: 'domcontentloaded' })

  const skipMain = page.getByRole('link', { name: 'Skip to main content' })
  await expect(skipMain).toHaveAttribute('href', '#main-content')
  await expect(skipMain).not.toHaveAttribute('tabindex', '-1')
  await skipMain.focus()
  await expect(skipMain).toBeFocused()
  await page.keyboard.press('Enter')
  await expect(page).toHaveURL(/#main-content$/)

  await expect(page.getByRole('main')).toBeVisible()
  await expect(
    page.getByRole('navigation', { name: 'Primary navigation' }),
  ).toBeVisible()
  await expect(
    page.getByRole('region', { name: 'Airport Intelligence' }),
  ).toBeVisible()
  await expect(
    page.getByRole('region', {
      name: 'Compare persisted analytical evidence',
    }),
  ).toBeVisible()
  await expect(
    page.getByRole('region', { name: 'Live traffic data controls' }),
  ).toBeVisible()
})

test('mobile navigation and analytical controls keep accessible names without overflow', async ({
  page,
}) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto(
    '/?region=az&aircraft=4b1801&view=intelligence#live-traffic',
    { waitUntil: 'domcontentloaded' },
  )

  await page.getByText('Navigate', { exact: true }).click()
  const navigation = page.getByRole('navigation', {
    name: 'Mobile primary navigation',
  })
  await expect(
    navigation.getByRole('link', { name: 'Airport Intelligence' }),
  ).toBeVisible()
  await expect(
    navigation.getByRole('link', { name: 'Historical Analytics' }),
  ).toBeVisible()
  await expect(
    navigation.getByRole('link', { name: 'Live workspace' }),
  ).toBeVisible()

  await expect(
    page.getByRole('tab', { name: /Intelligence/i }),
  ).toHaveAttribute('aria-selected', 'true')
  await expect(
    page.getByRole('progressbar', { name: 'Projection confidence score' }),
  ).toHaveAttribute('aria-valuenow', '72')
  await expect(
    page.getByRole('progressbar', { name: 'Mean forecast stability score' }),
  ).toHaveAttribute('aria-valuenow', '91')

  await expectNoHorizontalOverflow(page)
})
