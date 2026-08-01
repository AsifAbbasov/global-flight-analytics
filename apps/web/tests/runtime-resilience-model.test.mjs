// FRONTEND_PRODUCT_HARDENING_V1

import test from 'node:test'
import assert from 'node:assert/strict'

const {
  buildRuntimeConnectivityView,
  frontendRetryDelayMilliseconds,
  maximumFrontendQueryRetries,
  maximumFrontendRetryDelayMilliseconds,
  runtimeConnectivityState,
  shouldRetryFrontendQuery,
} = await import('../.test-dist/lib/product/runtime-resilience-model.js')

test('network and unknown failures are retried only while online', () => {
  assert.equal(
    shouldRetryFrontendQuery({ failureCount: 0, status: null, online: true }),
    true
  )
  assert.equal(
    shouldRetryFrontendQuery({ failureCount: 0, status: null, online: false }),
    false
  )
})

test('ordinary client errors are not retried', () => {
  for (const status of [400, 401, 403, 404, 409, 422]) {
    assert.equal(
      shouldRetryFrontendQuery({ failureCount: 0, status, online: true }),
      false
    )
  }
})

test('timeouts throttling and server failures remain retryable', () => {
  for (const status of [408, 425, 429, 500, 502, 503, 504]) {
    assert.equal(
      shouldRetryFrontendQuery({ failureCount: 0, status, online: true }),
      true
    )
  }
})

test('retry attempts are strictly bounded', () => {
  assert.equal(maximumFrontendQueryRetries, 2)
  assert.equal(
    shouldRetryFrontendQuery({
      failureCount: maximumFrontendQueryRetries,
      status: 503,
      online: true,
    }),
    false
  )
  assert.equal(
    shouldRetryFrontendQuery({ failureCount: -1, status: 503, online: true }),
    false
  )
})

test('retry delay uses bounded exponential backoff', () => {
  assert.equal(frontendRetryDelayMilliseconds(0), 1_000)
  assert.equal(frontendRetryDelayMilliseconds(1), 2_000)
  assert.equal(frontendRetryDelayMilliseconds(2), 4_000)
  assert.equal(frontendRetryDelayMilliseconds(20), maximumFrontendRetryDelayMilliseconds)
  assert.equal(frontendRetryDelayMilliseconds(Number.NaN), 1_000)
})

test('connectivity transitions remain explicit', () => {
  assert.equal(runtimeConnectivityState(null, false), 'unknown')
  assert.equal(runtimeConnectivityState(false, false), 'offline')
  assert.equal(runtimeConnectivityState(true, false), 'online')
  assert.equal(runtimeConnectivityState(true, true), 'recovered')
})

test('only offline and recovered states publish visible announcements', () => {
  const unknown = buildRuntimeConnectivityView('unknown')
  const online = buildRuntimeConnectivityView('online')
  const offline = buildRuntimeConnectivityView('offline')
  const recovered = buildRuntimeConnectivityView('recovered')

  assert.equal(unknown.visible, false)
  assert.equal(online.visible, false)
  assert.equal(offline.visible, true)
  assert.equal(offline.liveMode, 'assertive')
  assert.match(offline.message, /Existing evidence remains visible/)
  assert.equal(recovered.visible, true)
  assert.equal(recovered.liveMode, 'polite')
})
