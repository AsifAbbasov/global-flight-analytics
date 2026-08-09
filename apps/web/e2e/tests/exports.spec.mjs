import { expect, test } from '@playwright/test'

import {
  readDownloadedText,
  setScenario,
} from './helpers.mjs'

test.beforeEach(async ({ request }) => {
  await setScenario(request, 'healthy')
})

test('CSV export downloads the deterministic typed traffic snapshot', async ({
  page,
}) => {
  await page.goto('/?region=world&view=aircraft#live-traffic', {
    waitUntil: 'domcontentloaded',
  })

  const exportRegion = page.getByRole('region', {
    name: 'Research snapshot export',
  })
  const downloadButton = exportRegion.getByRole('button', {
    name: 'Download CSV',
  })
  await expect(downloadButton).toBeEnabled()

  const [download] = await Promise.all([
    page.waitForEvent('download'),
    downloadButton.click(),
  ])

  expect(download.suggestedFilename()).toMatch(
    /^global-flight-analytics-world-.*\.csv$/,
  )
  const content = await readDownloadedText(download)
  expect(content).toContain('schema_version,region_code,region_name')
  expect(content).toContain(',icao24,callsign,')
  expect(content).toContain('4b1801,AZAL101')
  expect(content).toContain('4ba901,THY202')
  await expect(
    exportRegion.getByText('CSV exported · 2 records.', { exact: true }),
  ).toBeVisible()
})

test('GeoJSON export preserves coordinates and provenance metadata', async ({
  page,
}) => {
  await page.goto('/?region=world&view=aircraft#live-traffic', {
    waitUntil: 'domcontentloaded',
  })

  const exportRegion = page.getByRole('region', {
    name: 'Research snapshot export',
  })
  const [download] = await Promise.all([
    page.waitForEvent('download'),
    exportRegion
      .getByRole('button', { name: 'Download GeoJSON' })
      .click(),
  ])

  expect(download.suggestedFilename()).toMatch(
    /^global-flight-analytics-world-.*\.geojson$/,
  )
  const payload = JSON.parse(await readDownloadedText(download))
  expect(payload.type).toBe('FeatureCollection')
  expect(payload.metadata.region_code).toBe('world')
  expect(payload.metadata.aircraft_count).toBe(2)
  expect(payload.metadata.feature_count).toBe(2)
  expect(payload.metadata.excluded_invalid_coordinate_count).toBe(0)
  expect(payload.features.map(feature => feature.id)).toEqual([
    '4b1801',
    '4ba901',
  ])
  expect(payload.features[0].geometry.coordinates).toEqual([
    49.8671,
    40.4093,
  ])
  await expect(
    exportRegion.getByText(
      'GeoJSON exported with snapshot provenance metadata.',
      { exact: true },
    ),
  ).toBeVisible()
})
