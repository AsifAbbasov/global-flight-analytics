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

test('renders an OpenAPI-backed server snapshot', async ({ page }) => {
  await page.goto('/', { waitUntil: 'domcontentloaded' })

  await expect(page.getByRole('heading', { level: 1 })).toContainText(
    'Observe aviation data.',
  )
  await expect(
    page.getByText('Initial snapshot ready', { exact: true }),
  ).toBeVisible()

  const initialAircraftCard = page
    .getByRole('article')
    .filter({ hasText: 'Initial aircraft' })
  await expect(initialAircraftCard).toContainText('2')

  const region = page.getByRole('combobox', { name: 'Region' })
  await expect(region).toHaveValue('world')
  await expect(region).toContainText('World')
  await expect(region).toContainText('Azerbaijan')
  await expect(region).toContainText('Turkey')

  await expect(
    page.getByRole('heading', { name: 'Current Traffic — World' }),
  ).toBeVisible()
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

  const region = page.getByRole('combobox', { name: 'Region' })
  await expect(region).toHaveValue('world')
  await region.selectOption('az')

  await expect(page).toHaveURL(
    /\/\?region=az&view=aircraft#live-traffic$/,
  )
  await expect(region).toHaveValue('az')
  await expect(
    page.getByRole('heading', {
      name: 'Current Traffic — Azerbaijan',
    }),
  ).toBeVisible()
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
    mobileNavigation.getByRole('link', { name: 'Live workspace' }),
  ).toBeVisible()
  await expect(
    mobileNavigation.getByRole('link', { name: 'Research scope' }),
  ).toBeVisible()

  const canonicalURL = new URL(page.url())
  expect(canonicalURL.searchParams.get('region')).toBe('world')
  expect(canonicalURL.searchParams.get('view')).toBe('aircraft')
})
