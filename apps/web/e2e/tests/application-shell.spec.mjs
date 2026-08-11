/* global URL, process */
import { expect, test } from '@playwright/test'

const mockAPIOrigin =
  process.env.PLAYWRIGHT_API_ORIGIN ?? 'http://127.0.0.1:8091'

async function setScenario(request, scenario) {
  const response = await request.post(`${mockAPIOrigin}/__e2e/scenario`, {
    data: { scenario },
  })
  expect(response.ok()).toBeTruthy()
}

test.beforeEach(async ({ request }) => {
  await setScenario(request, 'healthy')
})

test('renders an OpenAPI-backed server snapshot in the map-first shell', async ({ page }) => {
  await page.goto('/', { waitUntil: 'domcontentloaded' })

  await expect(
    page.getByRole('region', { name: 'Live flight tracker' }),
  ).toBeVisible()
  await expect(
    page.getByText('Initial snapshot ready', { exact: true }),
  ).toBeVisible()
  const initialStatus = page.getByLabel('Initial platform status')
  await expect(
    initialStatus.getByText('2', { exact: true }),
  ).toBeVisible()
  await expect(
    initialStatus.getByText('aircraft', { exact: true }),
  ).toBeVisible()

  const region = page.getByRole('combobox', { name: 'Traffic region' })
  await expect(region).toHaveValue('world')
  await expect(region).toContainText('World')
  await expect(region).toContainText('Azerbaijan')
  await expect(region).toContainText('Turkey')

  await expect(
    page.getByRole('region', { name: 'Current traffic map focused on World' }),
  ).toBeVisible()
  await page
    .getByLabel('Open live data controls')
    .click()
  await expect(
    page
      .getByRole('region', { name: 'Live traffic data controls' })
      .getByText('2 aircraft', { exact: true }),
  ).toBeVisible()
})

test('changing region updates the shareable workspace URL', async ({
  page,
}) => {
  await page.goto('/?region=WORLD&view=invalid#live-traffic', {
    waitUntil: 'domcontentloaded',
  })

  await expect(page).toHaveURL(
    /\/\?region=world&view=aircraft#live-traffic$/,
  )

  const region = page.getByRole('combobox', { name: 'Traffic region' })
  await expect(region).toHaveValue('world')
  await region.selectOption('az')

  await expect(page).toHaveURL(
    /\/\?region=az&view=aircraft#live-traffic$/,
  )
  await expect(region).toHaveValue('az')
  await expect(
    page.getByRole('region', {
      name: 'Current traffic map focused on Azerbaijan',
    }),
  ).toBeVisible()
  await page
    .getByLabel('Open live data controls')
    .click()
  await expect(
    page
      .getByRole('region', { name: 'Live traffic data controls' })
      .getByText('1 aircraft', { exact: true }),
  ).toBeVisible()
})

test('mobile navigation keeps product sections reachable', async ({
  page,
}) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto('/', { waitUntil: 'domcontentloaded' })

  await page.getByText('Navigate', { exact: true }).click()
  const mobileNavigation = page.getByRole('navigation', {
    name: 'Mobile primary navigation',
  })
  await expect(
    mobileNavigation.getByRole('link', { name: 'Live map' }),
  ).toBeVisible()
  await expect(
    mobileNavigation.getByRole('link', { name: 'Analytics' }),
  ).toBeVisible()
  await expect(
    mobileNavigation.getByRole('link', { name: 'History' }),
  ).toBeVisible()
  await expect(
    mobileNavigation.getByRole('link', { name: 'About' }),
  ).toBeVisible()

  const canonicalURL = new URL(page.url())
  expect(canonicalURL.searchParams.get('region')).toBe('world')
  expect(canonicalURL.searchParams.get('view')).toBe('aircraft')
})
