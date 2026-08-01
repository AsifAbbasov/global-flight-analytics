import assert from 'node:assert/strict'
import test from 'node:test'

const statusModuleURL = new URL(
  '../.test-dist/lib/product/application-status.js',
  import.meta.url
)
const importedStatusModule = await import(statusModuleURL.href)
const statusModule = importedStatusModule.default ?? importedStatusModule
const { buildApplicationInitialStatus } = statusModule

test('ready status preserves valid startup counts', () => {
  const status = buildApplicationInitialStatus({
    initialTrafficCount: 42,
    regionCount: 4,
    trafficUnavailable: false,
    regionsUnavailable: false,
  })

  assert.equal(status.availability, 'ready')
  assert.equal(status.initialTrafficCount, 42)
  assert.equal(status.regionCount, 4)
  assert.match(status.summary, /42 aircraft/)
})

test('a valid empty snapshot remains ready instead of becoming unavailable', () => {
  const status = buildApplicationInitialStatus({
    initialTrafficCount: 0,
    regionCount: 1,
    trafficUnavailable: false,
    regionsUnavailable: false,
  })

  assert.equal(status.availability, 'ready')
  assert.match(status.summary, /valid empty world snapshot/)
})

test('region catalog failure produces a degraded startup state', () => {
  const status = buildApplicationInitialStatus({
    initialTrafficCount: 12,
    regionCount: 1,
    trafficUnavailable: false,
    regionsUnavailable: true,
  })

  assert.equal(status.availability, 'degraded')
  assert.equal(status.label, 'Region catalog degraded')
})

test('traffic snapshot failure takes precedence over a region warning', () => {
  const status = buildApplicationInitialStatus({
    initialTrafficCount: 7,
    regionCount: 3,
    trafficUnavailable: true,
    regionsUnavailable: true,
  })

  assert.equal(status.availability, 'unavailable')
  assert.equal(status.label, 'Initial API snapshot unavailable')
})

test('invalid counts are normalized before presentation', () => {
  const status = buildApplicationInitialStatus({
    initialTrafficCount: Number.NaN,
    regionCount: -4,
    trafficUnavailable: false,
    regionsUnavailable: false,
  })

  assert.equal(status.initialTrafficCount, 0)
  assert.equal(status.regionCount, 0)
})
