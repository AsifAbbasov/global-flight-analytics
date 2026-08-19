#!/usr/bin/env node

import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url))
const repositoryRoot = path.resolve(scriptDirectory, '..')

function read(relativePath) {
  return fs.readFileSync(path.join(repositoryRoot, relativePath), 'utf8')
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message)
  }
}

function count(source, pattern) {
  return source.match(pattern)?.length ?? 0
}

function parseFallbackMinutes(workflowSource) {
  const cronMatches = [
    ...workflowSource.matchAll(/^\s*-\s+cron:\s+['"]([^'"]+)['"]\s*$/gm),
  ]
  assert(
    cronMatches.length === 1,
    'production ingestion workflow must declare exactly one GitHub fallback cron'
  )

  const cron = cronMatches[0][1].trim()
  const fields = cron.split(/\s+/)
  assert(
    fields.length === 5,
    'production ingestion fallback must use a five-field cron expression'
  )
  assert(
    fields.slice(1).every((field) => field === '*'),
    'production ingestion fallback must run uniformly across all hours and days'
  )

  const minuteField = fields[0]
  let minutes
  const stepMatch = /^\*\/(\d+)$/.exec(minuteField)
  if (stepMatch) {
    const step = Number(stepMatch[1])
    assert(
      Number.isInteger(step) && step > 0 && step <= 60,
      'production ingestion fallback minute step must be between 1 and 60'
    )
    minutes = []
    for (let minute = 0; minute < 60; minute += step) minutes.push(minute)
  } else {
    minutes = minuteField.split(',').map((value) => Number(value))
    assert(
      minutes.length > 0 &&
        minutes.every(
          (minute) => Number.isInteger(minute) && minute >= 0 && minute < 60
        ),
      'production ingestion fallback minutes must be valid minute values'
    )
    minutes = [...new Set(minutes)].sort((left, right) => left - right)
  }

  assert(
    minutes.length >= 2,
    'production ingestion fallback must run more than once per hour'
  )

  const gaps = minutes.map((minute, index) => {
    const next = minutes[(index + 1) % minutes.length]
    return index === minutes.length - 1 ? next + 60 - minute : next - minute
  })

  return {
    cron,
    minimumGapMinutes: Math.min(...gaps),
    maximumGapMinutes: Math.max(...gaps),
  }
}

const workflow = read('.github/workflows/production-traffic-ingestion.yml')
const main = read('apps/api/cmd/ingest/main.go')
const options = read('apps/api/cmd/ingest/command_options.go')
const runOnce = read('apps/api/cmd/ingest/run_once.go')
const freshness = read('scripts/verify-production-traffic-freshness.mjs')
const renderBlueprint = read('render.yaml')
const packageJSON = JSON.parse(read('package.json'))
const releaseVerifier = read('scripts/verify-release.sh')
const fallbackSchedule = parseFallbackMinutes(workflow)

assert(
  fallbackSchedule.maximumGapMinutes <= 15,
  'production ingestion GitHub fallback must never leave more than fifteen minutes between runs'
)
assert(
  fallbackSchedule.minimumGapMinutes >= 15,
  'GitHub Actions fallback must not become a higher-frequency primary scheduler'
)
assert(
  workflow.includes('workflow_dispatch:'),
  'production ingestion workflow must support a manual dispatch'
)
assert(
  workflow.includes('permissions:\n  contents: read'),
  'production ingestion workflow must use read-only repository permissions'
)
assert(
  workflow.includes('group: production-traffic-ingestion'),
  'production ingestion workflow must serialize production runs'
)
assert(
  workflow.includes('cancel-in-progress: false'),
  'production ingestion workflow must not cancel an active database write'
)
assert(
  workflow.includes('PRODUCTION_INGESTION_DATABASE_URL'),
  'production ingestion workflow must use the dedicated database secret'
)
assert(
  workflow.includes("TRAFFIC_INGESTION_RADIUS: '250'"),
  'production ingestion workflow must use the validated 250 nautical mile coverage radius'
)
assert(
  workflow.includes("node-version: '24.9.0'"),
  'production ingestion workflow must pin Node.js 24.9.0'
)
assert(
  workflow.includes('go run ./cmd/ingest --once'),
  'production ingestion workflow must use the bounded one-shot command'
)
assert(
  workflow.includes('verify-production-traffic-freshness.mjs'),
  'production ingestion workflow must verify end-to-end freshness'
)
assert(
  workflow.includes("MAX_TRAFFIC_AGE_SECONDS: '1800'"),
  'production ingestion workflow must retain the 1800-second freshness budget'
)
assert(
  count(workflow, /uses:\s+[^\s]+@[0-9a-f]{40}(?:\s+#\s+v\d+)?/g) ===
    count(workflow, /^\s*uses:/gm),
  'all production ingestion workflow actions must use immutable full SHAs'
)
assert(
  main.includes('runWithArgs(os.Args[1:])'),
  'ingest command must parse command-line options'
)
assert(
  main.includes('if commandOptions.Once {'),
  'ingest command must expose a bounded one-shot branch'
)
assert(
  main.includes('runSingleIngestionCycle('),
  'one-shot branch must use the tested single-cycle executor'
)
assert(
  options.includes('"once"'),
  'ingest command options must declare --once'
)
assert(
  runOnce.includes('NextDelay'),
  'single-cycle execution must publish a daemon-compatible observation result'
)
assert(
  freshness.includes('PRODUCTION_TRAFFIC_FRESHNESS=PASS'),
  'freshness verifier must publish a stable success marker'
)
assert(
  !renderBlueprint.includes('dockerCommand: /app/ingest'),
  'free production closure must not claim a paid Render ingestion worker'
)
assert(
  packageJSON.scripts['test:production-ingestion-contract'] ===
    'node --test scripts/verify-production-ingestion-contract.test.mjs scripts/verify-production-traffic-freshness.test.mjs',
  'package scripts must expose production ingestion tests'
)
assert(
  packageJSON.scripts['verify:production-ingestion-contract'] ===
    'node scripts/verify-production-ingestion-contract.mjs',
  'package scripts must expose production ingestion verification'
)
assert(
  releaseVerifier.includes('test:production-ingestion-contract'),
  'release verification must run production ingestion tests'
)
assert(
  releaseVerifier.includes('verify:production-ingestion-contract'),
  'release verification must run production ingestion source verification'
)

console.log(
  `PRODUCTION_INGESTION_FALLBACK_CRON=${fallbackSchedule.cron} ` +
    `MIN_GAP_MINUTES=${fallbackSchedule.minimumGapMinutes} ` +
    `MAX_GAP_MINUTES=${fallbackSchedule.maximumGapMinutes}`
)
console.log('PRODUCTION_INGESTION_CONTRACT=PASS')
