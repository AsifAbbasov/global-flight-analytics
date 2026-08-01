import assert from 'node:assert/strict'
import test from 'node:test'

const modelModuleURL = new URL(
  '../.test-dist/lib/traffic/live-traffic-status-model.js',
  import.meta.url
)
const importedModelModule = await import(modelModuleURL.href)
const modelModule = importedModelModule.default ?? importedModelModule
const {
  buildLiveTrafficStatus,
  normalizeLiveTrafficRefreshInterval,
} = modelModule

test('refresh interval normalization accepts only published choices', () => {
  assert.equal(normalizeLiveTrafficRefreshInterval(30_000), 30_000)
  assert.equal(normalizeLiveTrafficRefreshInterval(60_000), 60_000)
  assert.equal(normalizeLiveTrafficRefreshInterval(120_000), 120_000)
  assert.equal(normalizeLiveTrafficRefreshInterval(45_000), 60_000)
  assert.equal(normalizeLiveTrafficRefreshInterval(undefined), 60_000)
})

test('missing timestamps remain in a stable waiting state', () => {
  const status = buildLiveTrafficStatus({
    now: 1_000_000,
    dataUpdatedAt: 0,
    refreshIntervalMilliseconds: 60_000,
    autoRefreshEnabled: true,
  })

  assert.equal(status.freshness, 'waiting')
  assert.equal(status.ageMilliseconds, null)
  assert.equal(status.nextRefreshInMilliseconds, null)
  assert.equal(status.intervalProgress, 0)
  assert.equal(status.refreshDue, false)
})

test('current snapshots expose a bounded countdown and progress', () => {
  const status = buildLiveTrafficStatus({
    now: 1_045_000,
    dataUpdatedAt: 1_000_000,
    refreshIntervalMilliseconds: 60_000,
    autoRefreshEnabled: true,
  })

  assert.equal(status.freshness, 'current')
  assert.equal(status.ageMilliseconds, 45_000)
  assert.equal(status.nextRefreshInMilliseconds, 15_000)
  assert.equal(status.intervalProgress, 0.75)
  assert.equal(status.refreshDue, false)
})

test('aging and stale states use explicit refresh-window boundaries', () => {
  const aging = buildLiveTrafficStatus({
    now: 1_060_000,
    dataUpdatedAt: 1_000_000,
    refreshIntervalMilliseconds: 60_000,
    autoRefreshEnabled: true,
  })
  const stale = buildLiveTrafficStatus({
    now: 1_120_000,
    dataUpdatedAt: 1_000_000,
    refreshIntervalMilliseconds: 60_000,
    autoRefreshEnabled: true,
  })

  assert.equal(aging.freshness, 'aging')
  assert.equal(aging.nextRefreshInMilliseconds, 0)
  assert.equal(aging.refreshDue, true)
  assert.equal(stale.freshness, 'stale')
  assert.equal(stale.intervalProgress, 1)
  assert.equal(stale.refreshDue, true)
})

test('paused automatic refresh preserves freshness but removes countdown', () => {
  const status = buildLiveTrafficStatus({
    now: 1_030_000,
    dataUpdatedAt: 1_000_000,
    refreshIntervalMilliseconds: 30_000,
    autoRefreshEnabled: false,
  })

  assert.equal(status.freshness, 'aging')
  assert.equal(status.nextRefreshInMilliseconds, null)
  assert.equal(status.refreshDue, true)
})

test('future timestamps clamp snapshot age to zero', () => {
  const status = buildLiveTrafficStatus({
    now: 1_000_000,
    dataUpdatedAt: 1_010_000,
    refreshIntervalMilliseconds: 120_000,
    autoRefreshEnabled: true,
  })

  assert.equal(status.freshness, 'current')
  assert.equal(status.ageMilliseconds, 0)
  assert.equal(status.nextRefreshInMilliseconds, 120_000)
  assert.equal(status.intervalProgress, 0)
})
