import http from 'k6/http'
import { check } from 'k6'

const apiBaseURL = (__ENV.API_BASE_URL || '').replace(/\/+$/, '')

if (!apiBaseURL) {
  throw new Error('API_BASE_URL is required')
}

const endpoints = [
  { name: 'health', path: '/api/v1/health', endpointClass: 'lifecycle' },
  { name: 'readiness', path: '/api/v1/ready', endpointClass: 'lifecycle' },
  { name: 'version', path: '/api/v1/version', endpointClass: 'lifecycle' },
  { name: 'regions', path: '/api/v1/regions', endpointClass: 'database-read' },
  { name: 'airports', path: '/api/v1/airports?limit=20', endpointClass: 'database-read' },
  { name: 'current-traffic', path: '/api/v1/traffic/current?limit=100', endpointClass: 'database-read' },
]

export const options = {
  discardResponseBodies: true,
  summaryTrendStats: ['min', 'avg', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
  scenarios: {
    warmup: {
      executor: 'shared-iterations',
      exec: 'exercisePublicReadPath',
      vus: 2,
      iterations: 10,
      maxDuration: '10s',
    },
    baseline: {
      executor: 'constant-arrival-rate',
      exec: 'exercisePublicReadPath',
      rate: 8,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 12,
      maxVUs: 40,
      startTime: '5s',
      gracefulStop: '5s',
    },
  },
  thresholds: {
    checks: ['rate>0.999'],
    http_req_failed: ['rate<0.005'],
    http_reqs: ['count>=1400'],
    http_req_duration: ['p(95)<750', 'p(99)<1500'],
    'http_req_duration{endpoint_class:lifecycle}': ['p(95)<300'],
    'http_req_duration{endpoint_class:database-read}': ['p(95)<750'],
    dropped_iterations: ['count==0'],
  },
}

export function exercisePublicReadPath() {
  const requests = endpoints.map((endpoint) => [
    'GET',
    `${apiBaseURL}${endpoint.path}`,
    null,
    {
      timeout: '5s',
      tags: {
        route: endpoint.name,
        endpoint_class: endpoint.endpointClass,
      },
    },
  ])

  const responses = http.batch(requests)
  for (let index = 0; index < endpoints.length; index += 1) {
    const endpoint = endpoints[index]
    const response = responses[index]
    check(
      response,
      {
        [`${endpoint.name} returns HTTP 200`]: (current) => current.status === 200,
      },
      {
        route: endpoint.name,
        endpoint_class: endpoint.endpointClass,
      },
    )
  }
}

function thresholdFailures(data) {
  const failures = []
  for (const [metricName, metric] of Object.entries(data.metrics || {})) {
    for (const [thresholdName, threshold] of Object.entries(metric.thresholds || {})) {
      if (threshold.ok === false) {
        failures.push(`${metricName}: ${thresholdName}`)
      }
    }
  }
  return failures
}

export function handleSummary(data) {
  const failures = thresholdFailures(data)
  const requests = data.metrics?.http_reqs?.values?.count ?? 0
  const p95 = data.metrics?.http_req_duration?.values?.['p(95)'] ?? -1
  const failedRate = data.metrics?.http_req_failed?.values?.rate ?? 1
  const status = failures.length === 0 ? 'PASS' : 'FAIL'

  return {
    '/results/k6-summary.json': `${JSON.stringify(data, null, 2)}\n`,
    stdout: [
      `API_LOAD_BASELINE_K6=${status}`,
      `API_LOAD_BASELINE_REQUESTS=${requests}`,
      `API_LOAD_BASELINE_P95_MS=${p95}`,
      `API_LOAD_BASELINE_FAILED_RATE=${failedRate}`,
      ...(failures.length > 0
        ? [`API_LOAD_BASELINE_THRESHOLD_FAILURES=${failures.join('; ')}`]
        : []),
      '',
    ].join('\n'),
  }
}
