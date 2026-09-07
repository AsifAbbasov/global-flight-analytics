import { expect, test } from '@playwright/test'

import { setScenario } from './helpers.mjs'

test.beforeEach(async ({ request }) => {
  await setScenario(request, 'healthy')
})

test('projection weather stability airspace and ETA evolution expose server-owned evidence semantics', async ({
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

  const etaEvolution = page.getByRole('complementary', {
    name: 'Estimated Arrival Evolution',
  })
  await expect(etaEvolution).toBeVisible()
  await expect(
    etaEvolution.getByRole('heading', { name: 'Estimated Arrival Evolution' }),
  ).toBeVisible()
  await expect(etaEvolution).toHaveAttribute(
    'data-eta-evolution-evidence',
    'historically-recomputed-from-persisted-observations',
  )
  await expect(etaEvolution).toHaveAttribute(
    'data-eta-evolution-persisted-forecast-history',
    'none',
  )
  await expect(etaEvolution).toHaveAttribute(
    'data-eta-evolution-interpolation',
    'none',
  )
  await expect(etaEvolution).toHaveAttribute(
    'data-eta-evolution-cause-inference',
    'none',
  )
  await expect(
    etaEvolution.getByText(/not immutable forecast outputs stored at those past/i),
  ).toBeVisible()
  await expect(
    etaEvolution.getByText(/No ETA is interpolated between samples/i),
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

  await expect(
    page.getByRole('heading', { name: 'Airspace Intelligence — Azerbaijan' }),
  ).toBeVisible()
  await expect(
    page.getByRole('heading', { name: 'Regional occupancy' }),
  ).toBeVisible()
  await expect(
    page.getByText('Not suitable for separation or air traffic control.', {
      exact: true,
    }),
  ).toBeVisible()
  await expect(
    page.getByText(
      /They do not represent official sectors, controller workload, regulatory separation minima/,
    ),
  ).toBeVisible()
})


test('ETA reliability exposes bounded historical endpoint-proxy evidence', async ({
  page,
  request,
}) => {
  await setScenario(request, 'eta-reliability')
  await page.goto(
    '/?region=az&aircraft=4b1801&view=intelligence#live-traffic',
    { waitUntil: 'domcontentloaded' },
  )

  await expect(
    page.getByText('Historical ETA Reliability', { exact: true }),
  ).toBeVisible()
  await expect(
    page.getByRole('heading', {
      name: 'How reliable have comparable ETA estimates been?',
    }),
  ).toBeVisible()
  await expect(page.getByText('6 eligible / 8 checked', { exact: true })).toBeVisible()
  await expect(page.getByText('Median ETA error', { exact: true })).toBeVisible()
  await expect(page.getByText('4m 18s', { exact: true })).toBeVisible()
  await expect(page.getByText('80% error threshold', { exact: true })).toBeVisible()
  await expect(page.getByText('≤ 7m 42s', { exact: true })).toBeVisible()
  await expect(page.getByText('Within ±5 min', { exact: true })).toBeVisible()
  await expect(page.getByText('67%', { exact: true })).toBeVisible()
  await expect(page.getByText('Within ±10 min', { exact: true })).toBeVisible()
  await expect(page.getByText('83%', { exact: true }).first()).toBeVisible()
  await expect(
    page.getByText('ETA window covered endpoint', { exact: true }),
  ).toBeVisible()
  await expect(page.getByText('Evidence boundary', { exact: true })).toBeVisible()
  await expect(
    page.getByText(/observed endpoint proxy, not an official touchdown, gate or schedule timestamp/i),
  ).toBeVisible()
})
