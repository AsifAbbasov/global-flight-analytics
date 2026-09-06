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

  const timeNavigation = page.getByRole('region', {
    name: 'Historical replay time navigation',
  })
  await expect(
    timeNavigation.getByRole('heading', { name: 'Observed time navigation' }),
  ).toBeVisible()
  await expect(
    timeNavigation.getByRole('slider', { name: 'Historical replay time cursor' }),
  ).toHaveAttribute('max', '0')
  await expect(
    timeNavigation.getByText('Exact persisted observation at cursor.'),
  ).toBeVisible()
  await expect(
    timeNavigation.getByText(/Aircraft position changes only when another persisted observation is reached/),
  ).toBeVisible()

  const evidenceQuality = timeNavigation.getByLabel('Replay evidence quality profile')
  await expect(evidenceQuality).toBeVisible()
  await expect(evidenceQuality).toHaveAttribute(
    'data-flight-replay-evidence-quality',
    'descriptive-only'
  )
  await expect(evidenceQuality).toHaveAttribute(
    'data-flight-replay-evidence-score',
    'none'
  )
  await expect(
    evidenceQuality.getByText('Replay evidence quality', { exact: true }),
  ).toBeVisible()
  await expect(evidenceQuality.getByText('No synthetic score', { exact: true })).toBeVisible()
  await expect(
    evidenceQuality.getByText(/Only one persisted observation is available/),
  ).toBeVisible()
  const samplingIntervals = evidenceQuality.getByText('Sampling intervals', { exact: true }).locator('..')
  await expect(samplingIntervals.getByText('0', { exact: true })).toBeVisible()
  const altitudeEvidence = evidenceQuality.getByText('Altitude evidence', { exact: true }).locator('..')
  await expect(altitudeEvidence.getByText('100%', { exact: true })).toBeVisible()

  const evidenceProvenance = timeNavigation.getByLabel('Replay evidence provenance profile')
  await expect(evidenceProvenance).toBeVisible()
  await expect(evidenceProvenance).toHaveAttribute(
    'data-flight-replay-provenance',
    'persisted-source-labels-only'
  )
  await expect(evidenceProvenance).toHaveAttribute(
    'data-flight-replay-provider-ranking',
    'none'
  )
  await expect(evidenceProvenance).toHaveAttribute(
    'data-flight-replay-provider-accuracy-claim',
    'none'
  )
  await expect(
    evidenceProvenance.getByText('Replay evidence provenance', { exact: true }),
  ).toBeVisible()
  await expect(
    evidenceProvenance.getByText('Observed provenance', { exact: true }),
  ).toBeVisible()
  const identifiedSources = evidenceProvenance
    .getByText('Identified sources', { exact: true })
    .locator('..')
  await expect(identifiedSources.getByText('1', { exact: true })).toBeVisible()
  const unattributedSamples = evidenceProvenance
    .getByText('Unattributed samples', { exact: true })
    .locator('..')
  await expect(unattributedSamples.getByText('0', { exact: true })).toBeVisible()
  const sourceTransitions = evidenceProvenance
    .getByText('Observed source transitions', { exact: true })
    .locator('..')
  await expect(sourceTransitions.getByText('0', { exact: true })).toBeVisible()
  await expect(evidenceProvenance.getByText('playwright-fixture', { exact: true })).toBeVisible()
  await expect(
    evidenceProvenance.getByText(/Percentages are shares of persisted samples, not shares of elapsed time/),
  ).toBeVisible()

  const intervalComparison = timeNavigation.getByLabel('Observed interval comparison')
  await expect(intervalComparison).toBeVisible()
  await expect(
    intervalComparison.getByText('Observed interval comparison', { exact: true }),
  ).toBeVisible()
  await expect(intervalComparison).toHaveAttribute(
    'data-flight-replay-interval-evidence',
    'observed-endpoints-only'
  )
  await expect(intervalComparison).toHaveAttribute(
    'data-flight-replay-path-distance',
    'not-claimed'
  )
  await expect(
    intervalComparison.getByText(/At least two persisted observations are required/),
  ).toBeVisible()

  const replay = page.getByRole('region', { name: 'Historical flight replay' })
  await expect(
    replay.getByRole('heading', { name: 'Historical flight replay' }),
  ).toBeVisible()
  await expect(replay.getByText('Observed only')).toBeVisible()
  await expect(replay.getByText('No interpolation')).toBeVisible()
  await expect(replay.getByText('Sample 1 / 1')).toBeVisible()
  await expect(replay).toHaveAttribute(
    'data-replay-observation-id',
    'state-azal-101-1'
  )
  await expect(
    replay.getByRole('button', { name: 'Copy replay observation link' }),
  ).toBeVisible()
  await expect(
    replay.getByRole('slider', { name: 'Historical replay position' }),
  ).toHaveAttribute('max', '0')
  await expect(
    replay.getByLabel('Historical evidence gap timeline'),
  ).toBeVisible()
  await expect(
    replay.getByText(/Only one persisted observation is available/),
  ).toBeVisible()

  const analytics = replay.getByLabel('Observed replay analytics')
  await expect(analytics).toBeVisible()
  await expect(
    analytics.getByText('Observed replay analytics', { exact: true }),
  ).toBeVisible()
  await expect(analytics.getByText('Evidence-derived', { exact: true })).toBeVisible()
  await expect(
    analytics.getByText(/Aggregates use persisted samples only/),
  ).toBeVisible()
  await expect(
    analytics.getByText('Altitude coverage', { exact: true }),
  ).toBeVisible()
  await expect(analytics.getByText('100%', { exact: true })).toBeVisible()
  await expect(
    analytics.getByText('Observed altitude range', { exact: true }),
  ).toBeVisible()
  await expect(
    analytics.getByText('10,668–10,668 m', { exact: true }),
  ).toBeVisible()
  await expect(
    analytics.getByText('Peak observed velocity', { exact: true }),
  ).toBeVisible()
  await expect(
    analytics.getByText('230.0 m/s · 828 km/h', { exact: true }),
  ).toBeVisible()
  await expect(
    analytics.getByText('Airborne / ground samples', { exact: true }),
  ).toBeVisible()
  await expect(analytics.getByText('1 / 0', { exact: true })).toBeVisible()

  const observedChange = replay.getByLabel(
    'Observed change from previous sample'
  )
  await expect(observedChange).toBeVisible()
  await expect(
    observedChange.getByText('Observed change', { exact: true })
  ).toBeVisible()
  await expect(observedChange.getByText('Two-point evidence')).toBeVisible()
  await expect(
    observedChange.getByText(/No previous persisted observation is available/)
  ).toBeVisible()

  const velocityDatum = replay.getByText('Velocity', { exact: true }).locator('..')
  await expect(velocityDatum).toBeVisible()
  await expect(
    velocityDatum.getByText('230.0 m/s · 828 km/h', { exact: true }),
  ).toBeVisible()
  await expect(replay.getByText('Heading', { exact: true })).toBeVisible()
  await expect(replay.getByText('285°')).toBeVisible()
  await expect(replay.getByText('Vertical rate', { exact: true })).toBeVisible()
  await expect(replay.getByText('0.0 m/s', { exact: true })).toBeVisible()
  await expect(replay.getByText('Flight state', { exact: true })).toBeVisible()
  await expect(replay.getByText('Airborne', { exact: true })).toBeVisible()
  await expect(
    replay.getByText('Aircraft origin country', { exact: true }),
  ).toBeVisible()
  await expect(replay.getByText('Azerbaijan', { exact: true })).toBeVisible()
  await expect(replay.getByText('Source', { exact: true })).toBeVisible()
  await expect(replay.getByText('playwright-fixture')).toBeVisible()
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
