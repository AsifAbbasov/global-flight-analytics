#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'
import { pathToFileURL } from 'node:url'

function requiredMetric(summary, name) {
  const metric = summary?.metrics?.[name]
  if (!metric || typeof metric !== 'object' || !metric.values) {
    throw new Error(`required k6 metric is missing: ${name}`)
  }
  return metric
}

function finite(value, label) {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    throw new Error(`invalid numeric metric: ${label}`)
  }
  return value
}

function thresholdEvidence(summary) {
  const evidence = []
  for (const [metricName, metric] of Object.entries(summary.metrics || {})) {
    for (const [expression, result] of Object.entries(metric.thresholds || {})) {
      evidence.push({
        metric: metricName,
        expression,
        passed: result?.ok === true,
      })
    }
  }
  evidence.sort((left, right) =>
    `${left.metric}:${left.expression}`.localeCompare(`${right.metric}:${right.expression}`),
  )
  if (evidence.length < 7) {
    throw new Error('k6 summary does not contain the complete threshold set')
  }
  return evidence
}

export function buildEvidence(summary, environment = process.env) {
  const requests = requiredMetric(summary, 'http_reqs')
  const checks = requiredMetric(summary, 'checks')
  const failed = requiredMetric(summary, 'http_req_failed')
  const duration = requiredMetric(summary, 'http_req_duration')
  const lifecycle = requiredMetric(
    summary,
    'http_req_duration{endpoint_class:lifecycle}',
  )
  const databaseRead = requiredMetric(
    summary,
    'http_req_duration{endpoint_class:database-read}',
  )
  const iterations = requiredMetric(summary, 'iterations')
  const dropped = requiredMetric(summary, 'dropped_iterations')
  const thresholds = thresholdEvidence(summary)

  const sourceSHA = environment.PERFORMANCE_SOURCE_SHA || ''
  if (!/^[0-9a-f]{40}$/.test(sourceSHA)) {
    throw new Error('PERFORMANCE_SOURCE_SHA must be a full lowercase SHA')
  }

  const evidence = {
    schema_version: 'gfa.api-load-baseline.v1',
    generated_at: new Date().toISOString(),
    source_sha: sourceSHA,
    scope: 'containerized Go API plus PostgreSQL on one GitHub-hosted Ubuntu runner',
    profile: {
      warmup_iterations: 10,
      steady_arrival_rate_per_second: 8,
      steady_duration_seconds: 30,
      requests_per_iteration: 6,
      minimum_expected_requests: 1400,
      endpoints: [
        '/api/v1/health',
        '/api/v1/ready',
        '/api/v1/version',
        '/api/v1/regions',
        '/api/v1/airports?limit=20',
        '/api/v1/traffic/current?limit=100',
      ],
    },
    environment: {
      runner: environment.GITHUB_ACTIONS === 'true' ? 'github-hosted-ubuntu' : 'local-docker',
      api_image: environment.API_LOAD_IMAGE || 'global-flight-analytics-api:performance',
      postgres_image: environment.POSTGRES_IMAGE || 'postgres:16.14-alpine3.24',
      k6_image: environment.K6_IMAGE || 'grafana/k6:1.7.1',
    },
    results: {
      requests: finite(requests.values.count, 'http_reqs.count'),
      request_rate_per_second: finite(requests.values.rate, 'http_reqs.rate'),
      iterations: finite(iterations.values.count, 'iterations.count'),
      dropped_iterations: finite(dropped.values.count, 'dropped_iterations.count'),
      checks_rate: finite(checks.values.rate, 'checks.rate'),
      failed_request_rate: finite(failed.values.rate, 'http_req_failed.rate'),
      duration_ms: {
        average: finite(duration.values.avg, 'http_req_duration.avg'),
        median: finite(duration.values.med, 'http_req_duration.med'),
        p90: finite(duration.values['p(90)'], 'http_req_duration.p(90)'),
        p95: finite(duration.values['p(95)'], 'http_req_duration.p(95)'),
        p99: finite(duration.values['p(99)'], 'http_req_duration.p(99)'),
        maximum: finite(duration.values.max, 'http_req_duration.max'),
      },
      lifecycle_p95_ms: finite(
        lifecycle.values['p(95)'],
        'lifecycle.p(95)',
      ),
      database_read_p95_ms: finite(
        databaseRead.values['p(95)'],
        'database-read.p(95)',
      ),
    },
    thresholds,
    passed: thresholds.every((threshold) => threshold.passed),
  }

  if (!evidence.passed) {
    throw new Error('one or more k6 thresholds failed')
  }
  return evidence
}

export function renderMarkdown(evidence) {
  const result = evidence.results
  return [
    '# API Load Baseline',
    '',
    `- Source SHA: \`${evidence.source_sha}\``,
    `- Scope: ${evidence.scope}`,
    `- Requests: ${result.requests}`,
    `- Request rate: ${result.request_rate_per_second.toFixed(2)} requests/second`,
    `- Failed request rate: ${(result.failed_request_rate * 100).toFixed(3)}%`,
    `- Check success rate: ${(result.checks_rate * 100).toFixed(3)}%`,
    `- Overall p95: ${result.duration_ms.p95.toFixed(2)} milliseconds`,
    `- Overall p99: ${result.duration_ms.p99.toFixed(2)} milliseconds`,
    `- Lifecycle p95: ${result.lifecycle_p95_ms.toFixed(2)} milliseconds`,
    `- Database-read p95: ${result.database_read_p95_ms.toFixed(2)} milliseconds`,
    `- Dropped iterations: ${result.dropped_iterations}`,
    '',
    'This is a bounded Continuous Integration baseline, not a public Render capacity claim.',
    '',
  ].join('\n')
}

function main() {
  const [inputPath, jsonOutputPath, markdownOutputPath] = process.argv.slice(2)
  if (!inputPath || !jsonOutputPath || !markdownOutputPath) {
    throw new Error(
      'usage: summarize-api-load-baseline.mjs <k6-summary.json> <evidence.json> <evidence.md>',
    )
  }

  const summary = JSON.parse(fs.readFileSync(inputPath, 'utf8'))
  const evidence = buildEvidence(summary)
  fs.mkdirSync(path.dirname(jsonOutputPath), { recursive: true })
  fs.writeFileSync(jsonOutputPath, `${JSON.stringify(evidence, null, 2)}\n`)
  fs.writeFileSync(markdownOutputPath, renderMarkdown(evidence))
  console.log('API_LOAD_BASELINE_SUMMARY=PASS')
}

const invokedPath = process.argv[1] ? pathToFileURL(path.resolve(process.argv[1])).href : ''
if (import.meta.url === invokedPath) {
  try {
    main()
  } catch (error) {
    console.error(`API_LOAD_BASELINE_SUMMARY=FAIL reason=${error.message}`)
    process.exit(1)
  }
}
