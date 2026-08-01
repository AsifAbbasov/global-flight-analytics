import assert from 'node:assert/strict'
import test from 'node:test'
const clientModuleURL = new URL(
  '../.test-dist/lib/api/client.js',
  import.meta.url
)
const importedClientModule = await import(clientModuleURL.href)
const clientModule = importedClientModule.default ?? importedClientModule

const {
  APIRequestError,
  isAbortError,
  requestAPIData,
} = clientModule

const originalFetch = globalThis.fetch
const originalAPIBaseURL = globalThis.process.env.NEXT_PUBLIC_API_BASE_URL
const originalNodeEnvironment = globalThis.process.env.NODE_ENV

function restoreEnvironment() {
  globalThis.fetch = originalFetch
  if (originalAPIBaseURL === undefined) {
    delete globalThis.process.env.NEXT_PUBLIC_API_BASE_URL
  } else {
    globalThis.process.env.NEXT_PUBLIC_API_BASE_URL = originalAPIBaseURL
  }
  if (originalNodeEnvironment === undefined) {
    delete globalThis.process.env.NODE_ENV
  } else {
    globalThis.process.env.NODE_ENV = originalNodeEnvironment
  }
}

test.afterEach(restoreEnvironment)

test('requestAPIData returns data and preserves search parameters', async () => {
  globalThis.process.env.NEXT_PUBLIC_API_BASE_URL =
    'https://api.example.test/base/'
  let capturedURL = ''
  let capturedMethod = ''

  globalThis.fetch = async (input, init) => {
    capturedURL = String(input)
    capturedMethod = init?.method ?? ''
    return new Response(
      JSON.stringify({ success: true, data: { value: 42 } }),
      {
        status: 200,
        headers: { 'content-type': 'application/json; charset=utf-8' },
      }
    )
  }

  const result = await requestAPIData('metrics', {
    method: 'POST',
    searchParams: new URLSearchParams({ region: 'AZ', limit: '25' }),
  })

  assert.deepEqual(result, { value: 42 })
  assert.equal(
    capturedURL,
    'https://api.example.test/metrics?region=AZ&limit=25'
  )
  assert.equal(capturedMethod, 'POST')
})

test('requestAPIData rejects an invalid timeout before calling fetch', async () => {
  let calls = 0
  globalThis.fetch = async () => {
    calls += 1
    throw new Error('fetch must not run')
  }

  await assert.rejects(
    requestAPIData('/health', { timeoutMilliseconds: 0 }),
    error =>
      error instanceof APIRequestError &&
      error.message === 'The API request timeout is invalid.'
  )
  assert.equal(calls, 0)
})

test('requestAPIData preserves caller cancellation when fetch settles after the timeout deadline', async () => {
  const callerController = new AbortController()

  globalThis.fetch = (_input, init) =>
    new Promise((_resolve, reject) => {
      init?.signal?.addEventListener(
        'abort',
        () => {
          setTimeout(
            () =>
              reject(
                new DOMException('The operation was aborted.', 'AbortError')
              ),
            20
          )
        },
        { once: true }
      )
    })

  const request = requestAPIData('/aircraft/live', {
    signal: callerController.signal,
    timeoutMilliseconds: 5,
  })
  callerController.abort()

  await assert.rejects(request, error => isAbortError(error))
})

test('requestAPIData reports its own deadline as a timeout', async () => {
  globalThis.fetch = (_input, init) =>
    new Promise((_resolve, reject) => {
      init?.signal?.addEventListener(
        'abort',
        () =>
          reject(new DOMException('The operation was aborted.', 'AbortError')),
        { once: true }
      )
    })

  await assert.rejects(
    requestAPIData('/aircraft/live', { timeoutMilliseconds: 5 }),
    error =>
      error instanceof APIRequestError &&
      error.message === 'The API request timed out.'
  )
})

test('requestAPIData rejects non-JSON responses', async () => {
  globalThis.fetch = async () =>
    new Response('service unavailable', {
      status: 200,
      headers: { 'content-type': 'text/plain' },
    })

  await assert.rejects(
    requestAPIData('/health'),
    error =>
      error instanceof APIRequestError &&
      error.message === 'The API returned a non-JSON response.'
  )
})

test('requestAPIData requires an explicit API base URL in production', async () => {
  delete globalThis.process.env.NEXT_PUBLIC_API_BASE_URL
  globalThis.process.env.NODE_ENV = 'production'

  await assert.rejects(
    requestAPIData('/health'),
    error =>
      error instanceof APIRequestError &&
      error.message === 'NEXT_PUBLIC_API_BASE_URL is required in production.'
  )
})
