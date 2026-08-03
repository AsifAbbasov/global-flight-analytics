import assert from 'node:assert/strict'
import test from 'node:test'

import {
  assertTrafficFreshness,
  fetchJSONWithRetry,
  summarizeTrafficFreshness,
  unwrapAPIData,
} from './verify-production-traffic-freshness.mjs'

test('unwrapAPIData supports the standard response envelope', () => {
  assert.deepEqual(
    unwrapAPIData({ success: true, data: [{ icao24: 'abc123' }] }),
    [{ icao24: 'abc123' }]
  )
})

test('summarizeTrafficFreshness selects the newest observation', () => {
  const summary = summarizeTrafficFreshness(
    {
      success: true,
      data: [
        { observed_at: '2026-08-03T08:00:00.000Z' },
        { observed_at: '2026-08-03T08:09:30.000Z' },
      ],
    },
    new Date('2026-08-03T08:10:00.000Z')
  )

  assert.deepEqual(summary, {
    aircraftCount: 2,
    latestObservedAt: '2026-08-03T08:09:30.000Z',
    latestAgeSeconds: 30,
  })
})

test('assertTrafficFreshness rejects stale observations', () => {
  assert.throws(
    () =>
      assertTrafficFreshness(
        {
          aircraftCount: 1,
          latestObservedAt: '2026-08-03T08:00:00.000Z',
          latestAgeSeconds: 1801,
        },
        { maxAgeSeconds: 1800 }
      ),
    /production traffic is stale/
  )
})

test('assertTrafficFreshness accepts bounded future clock skew', () => {
  assert.doesNotThrow(() =>
    assertTrafficFreshness(
      {
        aircraftCount: 1,
        latestObservedAt: '2026-08-03T08:00:30.000Z',
        latestAgeSeconds: -30,
      },
      {
        maxAgeSeconds: 1800,
        maxFutureSkewSeconds: 60,
      }
    )
  )
})

test('summarizeTrafficFreshness rejects an empty response', () => {
  assert.throws(
    () => summarizeTrafficFreshness({ success: true, data: [] }),
    /at least one aircraft/
  )
})

test('fetchJSONWithRetry retries a transient failure', async () => {
  let attempts = 0
  const payload = await fetchJSONWithRetry('https://example.test/traffic', {
    attempts: 2,
    retryDelayMilliseconds: 0,
    requestTimeoutMilliseconds: 1000,
    sleep: async () => {},
    fetchImpl: async () => {
      attempts += 1
      if (attempts === 1) {
        throw new Error('temporary failure')
      }
      return {
        ok: true,
        async json() {
          return { success: true, data: [] }
        },
      }
    },
  })

  assert.equal(attempts, 2)
  assert.deepEqual(payload, { success: true, data: [] })
})
