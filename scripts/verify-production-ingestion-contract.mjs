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

const workflow = read('.github/workflows/production-traffic-ingestion.yml')
const main = read('apps/api/cmd/ingest/main.go')
const options = read('apps/api/cmd/ingest/command_options.go')
const runOnce = read('apps/api/cmd/ingest/run_once.go')
const freshness = read('scripts/verify-production-traffic-freshness.mjs')
const renderBlueprint = read('render.yaml')
const packageJSON = JSON.parse(read('package.json'))
const releaseVerifier = read('scripts/verify-release.sh')

assert(
  workflow.includes("cron: '*/10 * * * *'"),
  'production ingestion workflow must run every ten minutes'
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

console.log('PRODUCTION_INGESTION_CONTRACT=PASS')
