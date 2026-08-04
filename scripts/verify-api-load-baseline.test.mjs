import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import fs from 'node:fs'
import test from 'node:test'

import {
  buildEvidence,
  renderMarkdown,
} from './summarize-api-load-baseline.mjs'
import { validateEvidence } from './validate-api-load-baseline-evidence.mjs'

function metric(type, values, thresholds = {}) {
  return { type, contains: 'default', values, thresholds }
}

function validSummary() {
  return {
    state: { testRunDurationMs: 35100 },
    metrics: {
      checks: metric('rate', { rate: 1, passes: 1500, fails: 0 }, {
        'rate>0.999': { ok: true },
      }),
      http_req_failed: metric('rate', { rate: 0, passes: 0, fails: 1500 }, {
        'rate<0.005': { ok: true },
      }),
      http_reqs: metric('counter', { count: 1500, rate: 49.2 }, {
        'count>=1400': { ok: true },
      }),
      iterations: metric('counter', { count: 250, rate: 8.2 }),
      dropped_iterations: metric('counter', { count: 0, rate: 0 }, {
        'count==0': { ok: true },
      }),
      http_req_duration: metric('trend', {
        min: 2,
        avg: 40,
        med: 30,
        'p(90)': 80,
        'p(95)': 110,
        'p(99)': 220,
        max: 400,
      }, {
        'p(95)<750': { ok: true },
        'p(99)<1500': { ok: true },
      }),
      'http_req_duration{endpoint_class:lifecycle}': metric('trend', {
        min: 1,
        avg: 12,
        med: 10,
        'p(90)': 22,
        'p(95)': 30,
        'p(99)': 45,
        max: 70,
      }, {
        'p(95)<300': { ok: true },
      }),
      'http_req_duration{endpoint_class:database-read}': metric('trend', {
        min: 3,
        avg: 68,
        med: 55,
        'p(90)': 130,
        'p(95)': 170,
        'p(99)': 280,
        max: 420,
      }, {
        'p(95)<750': { ok: true },
      }),
    },
  }
}

const environment = {
  PERFORMANCE_SOURCE_SHA: '0123456789abcdef0123456789abcdef01234567',
  GITHUB_ACTIONS: 'true',
  API_LOAD_IMAGE: 'global-flight-analytics-api:performance',
  K6_IMAGE: 'grafana/k6:1.7.1',
  POSTGRES_IMAGE: 'postgres:16.14-alpine3.24',
}

test('API load baseline contract passes', () => {
  const output = execFileSync(
    process.execPath,
    ['scripts/verify-api-load-baseline.mjs'],
    { encoding: 'utf8' },
  )
  assert.match(output, /API_LOAD_BASELINE_CONTRACT=PASS/)
})

test('summary normalizes into bounded performance evidence', () => {
  const evidence = buildEvidence(validSummary(), environment)
  assert.equal(evidence.passed, true)
  assert.equal(evidence.results.requests, 1500)
  assert.equal(evidence.results.duration_ms.p95, 110)
  assert.equal(evidence.environment.runner, 'github-hosted-ubuntu')
  assert.equal(validateEvidence(evidence), true)
  assert.match(renderMarkdown(evidence), /not a public Render capacity claim/)
})

test('failed k6 thresholds cannot become passing evidence', () => {
  const summary = validSummary()
  summary.metrics.http_req_duration.thresholds['p(95)<750'].ok = false
  assert.throws(
    () => buildEvidence(summary, environment),
    /one or more k6 thresholds failed/,
  )
})

test('evidence validation rejects understated request volume', () => {
  const evidence = buildEvidence(validSummary(), environment)
  evidence.results.requests = 1399
  assert.throws(() => validateEvidence(evidence), /request count/)
})

test('load runner keeps the target inside an isolated Docker network', () => {
  const runner = fs.readFileSync('scripts/run-api-load-baseline.sh', 'utf8')
  assert.match(runner, /docker network create/)
  assert.match(runner, /--network "\$network_name"/)
  assert.doesNotMatch(runner, /--publish/)
})

test('load runner executes the bundled readiness probe instead of trusting image metadata', () => {
  const runner = fs.readFileSync('scripts/run-api-load-baseline.sh', 'utf8')
  assert.match(runner, /--no-healthcheck/)
  assert.match(runner, /--env HEALTHCHECK_URL="\$api_readiness_url"/)
  assert.match(runner, /docker exec "\$api_id" \/app\/healthcheck/)
  assert.match(runner, /API_LOAD_BASELINE_READINESS=PASS/)
  assert.doesNotMatch(runner, /\.State\.Health/)
})
