#!/usr/bin/env node
import fs from 'node:fs'

function fail(message) {
  console.error(`API_LOAD_BASELINE_CONTRACT=FAIL reason=${message}`)
  process.exit(1)
}

function read(relativePath) {
  return fs.readFileSync(relativePath, 'utf8')
}

const workflow = read('.github/workflows/api-load-baseline.yml')
const k6Script = read('load/k6/api-baseline.js')
const runner = read('scripts/run-api-load-baseline.sh')
const summarizer = read('scripts/summarize-api-load-baseline.mjs')
const validator = read('scripts/validate-api-load-baseline-evidence.mjs')
const packageJson = JSON.parse(read('package.json'))
const release = read('scripts/verify-release.sh')
const document = read('docs/174_API_LOAD_BASELINE.md')

for (const literal of [
  'pull_request:',
  'PERFORMANCE_SOURCE_SHA: ${{ github.event.pull_request.head.sha || github.sha }}',
  'grafana/k6:1.7.1',
  'postgres:16.14-alpine3.24',
  'bash scripts/run-api-load-baseline.sh',
  'actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a # v7.0.1',
  'artifacts/performance/api-load-baseline.json',
  'retention-days: 30',
]) {
  if (!workflow.includes(literal)) fail(`workflow is missing ${literal}`)
}
if (workflow.includes('global-flight-analytics-api.onrender.com')) {
  fail('load baseline must not target the public Render service')
}
if (workflow.includes('schedule:')) {
  fail('load baseline must not run on a schedule')
}

for (const literal of [
  "executor: 'constant-arrival-rate'",
  'rate: 8',
  "duration: '30s'",
  "http_reqs: ['count>=1400']",
  "http_req_failed: ['rate<0.005']",
  "http_req_duration: ['p(95)<750', 'p(99)<1500']",
  "'http_req_duration{endpoint_class:lifecycle}': ['p(95)<300']",
  "'http_req_duration{endpoint_class:database-read}': ['p(95)<750']",
  "dropped_iterations: ['count==0']",
  '/api/v1/traffic/current?limit=100',
  '/results/k6-summary.json',
]) {
  if (!k6Script.includes(literal)) fail(`k6 script is missing ${literal}`)
}
if (/https?:\/\//.test(k6Script.replace("const apiBaseURL = (__ENV.API_BASE_URL || '').replace(/\\\/+$/, '')", ''))) {
  fail('k6 script contains a hard-coded remote HTTP origin')
}

for (const literal of [
  'docker network create',
  'scripts/wait-for-postgres-container.sh',
  '/app/migrate',
  '--no-healthcheck',
  '--env HEALTHCHECK_URL="$api_readiness_url"',
  'docker exec "$api_id" /app/healthcheck',
  'API_LOAD_BASELINE_READINESS=PASS',
  'API_LOAD_BASELINE_TARGET=PASS',
  'node scripts/summarize-api-load-baseline.mjs',
  'node scripts/validate-api-load-baseline-evidence.mjs',
  'API_LOAD_BASELINE=PASS',
]) {
  if (!runner.includes(literal)) fail(`runner is missing ${literal}`)
}
if (runner.includes('.State.Health')) {
  fail('runner must not trust image health metadata for readiness')
}
if (runner.includes('--publish')) fail('load target must remain isolated from host ports')

for (const literal of [
  "schema_version: 'gfa.api-load-baseline.v1'",
  'minimum_expected_requests: 1400',
  'This is a bounded Continuous Integration baseline, not a public Render capacity claim.',
]) {
  if (!summarizer.includes(literal)) fail(`summarizer is missing ${literal}`)
}
for (const literal of [
  "result.requests >= 1400",
  "result.dropped_iterations === 0",
  "result.duration_ms?.p95 < 750",
  "result.duration_ms?.p99 < 1500",
  "result.lifecycle_p95_ms < 300",
  "result.database_read_p95_ms < 750",
]) {
  if (!validator.includes(literal)) fail(`evidence validator is missing ${literal}`)
}

const expectedScripts = {
  'test:api-load-baseline': 'node --test scripts/verify-api-load-baseline.test.mjs',
  'verify:api-load-baseline': 'node scripts/verify-api-load-baseline.mjs',
  'run:api-load-baseline': 'bash scripts/run-api-load-baseline.sh',
}
for (const [name, command] of Object.entries(expectedScripts)) {
  if (packageJson.scripts?.[name] !== command) fail(`package script is missing: ${name}`)
}
for (const command of [
  'pnpm run test:api-load-baseline',
  'pnpm run verify:api-load-baseline',
]) {
  if (!release.includes(command)) fail(`release gate is missing ${command}`)
}

for (const literal of [
  'Status: IMPLEMENTED',
  'not a Render capacity claim',
  'api-load-baseline.json',
  'API_LOAD_BASELINE=PASS',
]) {
  if (!document.includes(literal)) fail(`Document 174 is missing ${literal}`)
}
if (!/forty-eight HTTP\s+requests per second/.test(document)) {
  fail('Document 174 is missing the steady request-rate explanation')
}

console.log('API_LOAD_BASELINE_CONTRACT=PASS')
