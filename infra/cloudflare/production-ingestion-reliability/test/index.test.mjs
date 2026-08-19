import assert from 'node:assert/strict'
import test from 'node:test'

import {
  classifyTrafficFreshness,
  classifyWorkflowRuns,
  executeScheduled,
  handleFetch,
  loadConfig,
  summarizeTrafficFreshness,
} from '../src/index.mjs'

const NOW = Date.parse('2026-08-06T07:00:00Z')

function environment(overrides = {}) {
  return {
    GITHUB_ACTIONS_TOKEN: 'test-token-never-log',
    GITHUB_API_BASE_URL: 'https://api.github.com/',
    GITHUB_API_VERSION: '2022-11-28',
    GITHUB_OWNER: 'AsifAbbasov',
    GITHUB_REPOSITORY: 'global-flight-analytics',
    GITHUB_WORKFLOW: 'production-traffic-ingestion.yml',
    GITHUB_REF: 'main',
    TRAFFIC_API_URL:
      'https://global-flight-analytics-api.onrender.com/api/v1/traffic/current',
    PRIMARY_CRON: '3,13,23,33,43,53 * * * *',
    WATCHDOG_CRON: '*/5 * * * *',
    MAX_TRAFFIC_AGE_SECONDS: '1800',
    MAX_FUTURE_SKEW_SECONDS: '60',
    DEDUPLICATION_WINDOW_SECONDS: '480',
    RECENT_FAILURE_COOLDOWN_SECONDS: '21600',
    DISPATCH_ENABLED: 'true',
    REQUEST_TIMEOUT_MILLISECONDS: '10000',
    WORKFLOW_RUNS_PER_PAGE: '20',
    ...overrides,
  }
}

function JSONResponse(payload, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      'Content-Type': 'application/json',
    },
  })
}

function response(status = 204) {
  return new Response(null, { status })
}

function sequenceFetch(handlers) {
  const calls = []
  const fetchImpl = async (url, init = {}) => {
    const call = {
      url: String(url),
      init,
    }
    calls.push(call)
    const handler = handlers.shift()
    assert.ok(handler, `unexpected fetch call ${call.url}`)
    return handler(call)
  }
  return { calls, fetchImpl }
}

function controller(cron) {
  return {
    cron,
    scheduledTime: NOW,
  }
}

test('configuration requires a secret and distinct Cron Triggers', () => {
  assert.throws(
    () => loadConfig(environment({ GITHUB_ACTIONS_TOKEN: '' })),
    /GITHUB_ACTIONS_TOKEN is required/
  )
  assert.throws(
    () =>
      loadConfig(
        environment({
          WATCHDOG_CRON: '3,13,23,33,43,53 * * * *',
        })
      ),
    /must be different/
  )
  assert.throws(
    () =>
      loadConfig(
        environment({
          TRAFFIC_API_URL: 'http://example.test/traffic',
        })
      ),
    /must be an HTTPS URL/
  )
})

test('dispatch kill switch suppresses both Cron Triggers without network calls', async () => {
  for (const cron of [
    environment().PRIMARY_CRON,
    environment().WATCHDOG_CRON,
  ]) {
    const mock = sequenceFetch([])
    const result = await executeScheduled(
      controller(cron),
      environment({ DISPATCH_ENABLED: 'false' }),
      {
        fetchImpl: mock.fetchImpl,
        nowMilliseconds: NOW,
      }
    )

    assert.equal(result.marker, 'CLOUDFLARE_DISPATCH_DISABLED')
    assert.equal(result.action, 'skipped-dispatch-disabled')
    assert.equal(mock.calls.length, 0)
  }
})

test('primary schedule dispatches one exact workflow request', async () => {
  const mock = sequenceFetch([
    () => JSONResponse({ workflow_runs: [] }),
    () => response(204),
  ])

  const result = await executeScheduled(
    controller(environment().PRIMARY_CRON),
    environment(),
    {
      fetchImpl: mock.fetchImpl,
      nowMilliseconds: NOW,
    }
  )

  assert.equal(result.marker, 'CLOUDFLARE_PRIMARY_SCHEDULE')
  assert.equal(result.action, 'dispatched')
  assert.equal(mock.calls.length, 2)
  assert.match(
    mock.calls[0].url,
    /actions\/workflows\/production-traffic-ingestion\.yml\/runs/
  )
  assert.match(mock.calls[0].url, /branch=main/)
  assert.equal(mock.calls[1].init.method, 'POST')
  assert.equal(
    mock.calls[1].init.headers.Authorization,
    'Bearer test-token-never-log'
  )
  assert.deepEqual(JSON.parse(mock.calls[1].init.body), {
    ref: 'main',
    inputs: {
      dispatch_source: 'cloudflare-primary',
    },
  })
})

test('primary schedule suppresses an active run', async () => {
  const mock = sequenceFetch([
    () =>
      JSONResponse({
        workflow_runs: [
          {
            id: 101,
            status: 'in_progress',
            event: 'workflow_dispatch',
          },
        ],
      }),
  ])

  const result = await executeScheduled(
    controller(environment().PRIMARY_CRON),
    environment(),
    {
      fetchImpl: mock.fetchImpl,
      nowMilliseconds: NOW,
    }
  )

  assert.equal(result.marker, 'CLOUDFLARE_PRIMARY_SCHEDULE')
  assert.equal(result.action, 'skipped-active-run')
  assert.equal(result.run.id, 101)
  assert.equal(mock.calls.length, 1)
})

test('primary schedule suppresses a recent successful run', async () => {
  const mock = sequenceFetch([
    () =>
      JSONResponse({
        workflow_runs: [
          {
            id: 102,
            status: 'completed',
            conclusion: 'success',
            event: 'schedule',
            created_at: '2026-08-06T06:54:00Z',
          },
        ],
      }),
  ])

  const result = await executeScheduled(
    controller(environment().PRIMARY_CRON),
    environment(),
    {
      fetchImpl: mock.fetchImpl,
      nowMilliseconds: NOW,
    }
  )

  assert.equal(result.action, 'skipped-recent-success')
  assert.equal(result.run.id, 102)
  assert.equal(mock.calls.length, 1)
})

test('recent failed workflow run opens the circuit breaker', async () => {
  const mock = sequenceFetch([
    () =>
      JSONResponse({
        workflow_runs: [
          {
            id: 105,
            status: 'completed',
            conclusion: 'failure',
            event: 'workflow_dispatch',
            created_at: '2026-08-06T06:55:00Z',
          },
        ],
      }),
  ])

  const result = await executeScheduled(
    controller(environment().PRIMARY_CRON),
    environment(),
    {
      fetchImpl: mock.fetchImpl,
      nowMilliseconds: NOW,
    }
  )

  assert.equal(result.action, 'skipped-recent-failure')
  assert.equal(result.decisionMarker, 'RECENT_FAILURE_CIRCUIT_BREAKER')
  assert.equal(result.run.id, 105)
  assert.equal(mock.calls.length, 1)
})

test('newer success resets an older failure circuit breaker', async () => {
  const mock = sequenceFetch([
    () =>
      JSONResponse({
        workflow_runs: [
          {
            id: 106,
            status: 'completed',
            conclusion: 'failure',
            event: 'workflow_dispatch',
            created_at: '2026-08-06T06:50:00Z',
          },
          {
            id: 107,
            status: 'completed',
            conclusion: 'success',
            event: 'workflow_dispatch',
            created_at: '2026-08-06T06:58:00Z',
          },
        ],
      }),
  ])

  const result = await executeScheduled(
    controller(environment().PRIMARY_CRON),
    environment(),
    {
      fetchImpl: mock.fetchImpl,
      nowMilliseconds: NOW,
    }
  )

  assert.equal(result.action, 'skipped-recent-success')
  assert.equal(result.run.id, 107)
  assert.equal(mock.calls.length, 1)
})

test('watchdog skips GitHub when traffic is fresh', async () => {
  const mock = sequenceFetch([
    () =>
      JSONResponse({
        data: [
          {
            observed_at: '2026-08-06T06:59:30Z',
          },
        ],
      }),
  ])

  const result = await executeScheduled(
    controller(environment().WATCHDOG_CRON),
    environment(),
    {
      fetchImpl: mock.fetchImpl,
      nowMilliseconds: NOW,
    }
  )

  assert.equal(result.marker, 'CLOUDFLARE_WATCHDOG')
  assert.equal(result.action, 'skipped-fresh-traffic')
  assert.equal(result.freshness.latestAgeSeconds, 30)
  assert.equal(mock.calls.length, 1)
})

test('watchdog dispatches exactly once when traffic is stale', async () => {
  const mock = sequenceFetch([
    () =>
      JSONResponse({
        data: [
          {
            observed_at: '2026-08-06T05:00:00Z',
          },
        ],
      }),
    () => JSONResponse({ workflow_runs: [] }),
    () => response(204),
  ])

  const result = await executeScheduled(
    controller(environment().WATCHDOG_CRON),
    environment(),
    {
      fetchImpl: mock.fetchImpl,
      nowMilliseconds: NOW,
    }
  )

  assert.equal(result.marker, 'STALE_TRAFFIC_RECOVERY_DISPATCH')
  assert.equal(result.action, 'dispatched')
  assert.equal(result.freshness.latestAgeSeconds, 7200)
  assert.deepEqual(JSON.parse(mock.calls[2].init.body), {
    ref: 'main',
    inputs: {
      dispatch_source: 'cloudflare-watchdog',
    },
  })
})

test('watchdog suppresses recovery while another run is active', async () => {
  const mock = sequenceFetch([
    () => JSONResponse({ data: [] }),
    () =>
      JSONResponse({
        workflow_runs: [
          {
            id: 103,
            status: 'queued',
            event: 'workflow_dispatch',
          },
        ],
      }),
  ])

  const result = await executeScheduled(
    controller(environment().WATCHDOG_CRON),
    environment(),
    {
      fetchImpl: mock.fetchImpl,
      nowMilliseconds: NOW,
    }
  )

  assert.equal(result.action, 'skipped-active-run')
  assert.equal(result.run.id, 103)
  assert.equal(mock.calls.length, 2)
})

test('future traffic fails closed without dispatching', async () => {
  const mock = sequenceFetch([
    () =>
      JSONResponse({
        data: [
          {
            observed_at: '2026-08-06T07:02:00Z',
          },
        ],
      }),
  ])

  await assert.rejects(
    executeScheduled(
      controller(environment().WATCHDOG_CRON),
      environment(),
      {
        fetchImpl: mock.fetchImpl,
        nowMilliseconds: NOW,
      }
    ),
    /too far in the future/
  )
  assert.equal(mock.calls.length, 1)
})

test('traffic and workflow classifiers preserve exact boundaries', () => {
  const summary = summarizeTrafficFreshness(
    {
      data: [
        {
          observed_at: '2026-08-06T06:30:00Z',
        },
      ],
    },
    NOW
  )
  assert.equal(summary.latestAgeSeconds, 1800)
  assert.equal(
    classifyTrafficFreshness(summary, loadConfig(environment())),
    'fresh'
  )

  const runs = classifyWorkflowRuns(
    [
      {
        id: 104,
        status: 'completed',
        conclusion: 'failure',
        event: 'workflow_dispatch',
        created_at: '2026-08-06T06:59:00Z',
      },
    ],
    NOW,
    480
  )
  assert.equal(runs.activeRun, null)
  assert.equal(runs.recentSuccessfulRun, null)
  assert.equal(runs.recentFailedRun?.id, 104)
})

test('health endpoint exposes configuration but never the token', async () => {
  const result = await handleFetch(
    new Request('https://worker.example/health'),
    environment()
  )
  assert.equal(result.status, 200)

  const body = await result.text()
  assert.match(body, /github_token_configured/)
  assert.doesNotMatch(body, /test-token-never-log/)
  assert.doesNotMatch(body, /Authorization/)
})

test('unknown Cron Trigger fails closed', async () => {
  const mock = sequenceFetch([])
  await assert.rejects(
    executeScheduled(
      controller('17 * * * *'),
      environment(),
      {
        fetchImpl: mock.fetchImpl,
        nowMilliseconds: NOW,
      }
    ),
    /unsupported Cron Trigger/
  )
  assert.equal(mock.calls.length, 0)
})
