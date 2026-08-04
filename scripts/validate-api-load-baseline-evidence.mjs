#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'
import { pathToFileURL } from 'node:url'

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

export function validateEvidence(evidence) {
  assert(evidence?.schema_version === 'gfa.api-load-baseline.v1', 'invalid schema version')
  assert(/^[0-9a-f]{40}$/.test(evidence?.source_sha || ''), 'invalid source SHA')
  assert(evidence?.passed === true, 'baseline is not marked as passed')
  assert(evidence?.profile?.steady_arrival_rate_per_second === 8, 'unexpected arrival rate')
  assert(evidence?.profile?.steady_duration_seconds === 30, 'unexpected steady duration')
  assert(evidence?.profile?.requests_per_iteration === 6, 'unexpected request mix')
  assert(evidence?.environment?.k6_image === 'grafana/k6:1.7.1', 'unexpected k6 image')
  assert(evidence?.environment?.postgres_image === 'postgres:16.14-alpine3.24', 'unexpected PostgreSQL image')

  const result = evidence?.results || {}
  assert(result.requests >= 1400, 'request count is below the baseline workload')
  assert(result.dropped_iterations === 0, 'load generator dropped iterations')
  assert(result.checks_rate > 0.999, 'check success rate is below objective')
  assert(result.failed_request_rate < 0.005, 'failed request rate is above objective')
  assert(result.duration_ms?.p95 < 750, 'overall p95 is above objective')
  assert(result.duration_ms?.p99 < 1500, 'overall p99 is above objective')
  assert(result.lifecycle_p95_ms < 300, 'lifecycle p95 is above objective')
  assert(result.database_read_p95_ms < 750, 'database-read p95 is above objective')

  assert(Array.isArray(evidence?.thresholds), 'threshold evidence is missing')
  assert(evidence.thresholds.length >= 7, 'threshold evidence is incomplete')
  assert(evidence.thresholds.every((threshold) => threshold.passed === true), 'a threshold is not passed')
  return true
}

function main() {
  const evidencePath = process.argv[2]
  if (!evidencePath) {
    throw new Error('usage: validate-api-load-baseline-evidence.mjs <evidence.json>')
  }
  validateEvidence(JSON.parse(fs.readFileSync(evidencePath, 'utf8')))
  console.log('API_LOAD_BASELINE_EVIDENCE=PASS')
}

const invokedPath = process.argv[1] ? pathToFileURL(path.resolve(process.argv[1])).href : ''
if (import.meta.url === invokedPath) {
  try {
    main()
  } catch (error) {
    console.error(`API_LOAD_BASELINE_EVIDENCE=FAIL reason=${error.message}`)
    process.exit(1)
  }
}
