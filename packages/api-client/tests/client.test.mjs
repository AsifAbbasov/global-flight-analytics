import assert from 'node:assert/strict'
import test from 'node:test'
import {
  APIError,
  GlobalFlightAnalyticsClient,
  operationDefinitions,
} from '../dist/index.js'

function jsonResponse(body, init = {}) {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: {
      'content-type': 'application/json; charset=utf-8',
      'x-request-id': 'request-test-1',
      ...init.headers,
    },
  })
}

test('public GET uses generated path and query metadata', async () => {
  const calls = []
  const client = new GlobalFlightAnalyticsClient({
    baseURL: 'https://api.example.test/',
    fetch: async (input, init) => {
      calls.push({ input: String(input), init })
      return jsonResponse({ success: true, data: { items: [] } })
    },
  })

  await client.request('listRouteIntelligenceHistoryByTrajectoryID', {
    path: { id: '550e8400-e29b-41d4-a716-446655440000' },
    query: { limit: 20, before_as_of_time: '2026-08-05T12:00:00Z' },
  })

  assert.equal(calls.length, 1)
  assert.equal(
    calls[0].input,
    'https://api.example.test/api/v1/trajectories/550e8400-e29b-41d4-a716-446655440000/route-intelligence/history?limit=20&before_as_of_time=2026-08-05T12%3A00%3A00Z',
  )
  assert.equal(calls[0].init.method, 'GET')
  assert.equal(new Headers(calls[0].init.headers).get('X-Internal-API-Key'), null)
})

test('protected mutation requires and attaches the internal API key', async () => {
  let called = false
  const withoutKey = new GlobalFlightAnalyticsClient({
    baseURL: 'https://api.example.test/',
    fetch: async () => {
      called = true
      return jsonResponse({ success: true })
    },
  })

  await assert.rejects(
    withoutKey.request('processRouteIntelligenceByTrajectoryID', {
      path: { id: '550e8400-e29b-41d4-a716-446655440000' },
    }),
    /requires internalApiKey/,
  )
  assert.equal(called, false)

  const calls = []
  const withKey = new GlobalFlightAnalyticsClient({
    baseURL: 'https://api.example.test/',
    internalApiKey: 'test-only-internal-key-000000000000',
    fetch: async (input, init) => {
      calls.push({ input: String(input), init })
      return jsonResponse({ success: true, data: {} })
    },
  })
  await withKey.request('processRouteIntelligenceByTrajectoryID', {
    path: { id: '550e8400-e29b-41d4-a716-446655440000' },
  })
  assert.equal(calls[0].init.method, 'POST')
  assert.equal(
    new Headers(calls[0].init.headers).get('X-Internal-API-Key'),
    'test-only-internal-key-000000000000',
  )
})

test('typed API errors preserve status code request ID and details', async () => {
  const client = new GlobalFlightAnalyticsClient({
    baseURL: 'https://api.example.test/',
    fetch: async () => jsonResponse({
      success: false,
      error: {
        code: 'SERVICE_UNAVAILABLE',
        message: 'backend unavailable',
        details: { dependency: 'postgres' },
      },
    }, { status: 503 }),
  })

  await assert.rejects(
    client.request('getHealth', {}),
    (error) => {
      assert(error instanceof APIError)
      assert.equal(error.status, 503)
      assert.equal(error.code, 'SERVICE_UNAVAILABLE')
      assert.equal(error.requestId, 'request-test-1')
      assert.deepEqual(error.details, { dependency: 'postgres' })
      return true
    },
  )
})

test('unresolved path validation rejects adversarial templates before fetch', async () => {
  const definition = operationDefinitions.getHealth
  const originalPath = definition.path
  let called = false

  try {
    definition.path = '{'.repeat(100_000)
    const client = new GlobalFlightAnalyticsClient({
      baseURL: 'https://api.example.test/',
      fetch: async () => {
        called = true
        return jsonResponse({ success: true })
      },
    })

    await assert.rejects(
      client.request('getHealth', {}),
      /unresolved path parameter/,
    )
    assert.equal(called, false)
  } finally {
    definition.path = originalPath
  }
})

test('base URL rejects credentials paths queries and fragments', () => {
  for (const baseURL of [
    'https://user:secret@api.example.test/',
    'https://api.example.test/prefix',
    'https://api.example.test/?environment=test',
    'https://api.example.test/#fragment',
    'file:///tmp/api',
  ]) {
    assert.throws(() => new GlobalFlightAnalyticsClient({ baseURL }), TypeError)
  }
})
