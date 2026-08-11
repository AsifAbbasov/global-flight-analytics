/* global document, process */
import fs from 'node:fs/promises'

import { expect } from '@playwright/test'

export const mockAPIOrigin =
  process.env.PLAYWRIGHT_API_ORIGIN ?? 'http://127.0.0.1:8091'

export async function setScenario(request, scenario) {
  const response = await request.post(`${mockAPIOrigin}/__e2e/scenario`, {
    data: { scenario },
  })
  expect(response.ok()).toBeTruthy()
}

export async function waitForMapFirstHydration(page) {
  await expect(
    page.getByRole('region', { name: 'Live flight tracker' }),
  ).toHaveAttribute('data-hydrated', 'true')
}

export async function readDownloadedText(download) {
  const downloadedPath = await download.path()
  expect(downloadedPath).not.toBeNull()
  return fs.readFile(downloadedPath, 'utf8')
}

export async function expectNoHorizontalOverflow(page) {
  const dimensions = await page.evaluate(() => ({
    clientWidth: document.documentElement.clientWidth,
    scrollWidth: document.documentElement.scrollWidth,
  }))
  expect(dimensions.scrollWidth).toBeLessThanOrEqual(dimensions.clientWidth + 1)
}
