#!/usr/bin/env node

import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '..'
)

function read(relativePath) {
  return fs.readFileSync(
    path.join(repositoryRoot, relativePath),
    'utf8'
  )
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message)
  }
}

const workerPath =
  'infra/cloudflare/production-ingestion-reliability'
const source = read(`${workerPath}/src/index.mjs`)
const tests = read(`${workerPath}/test/index.test.mjs`)
const configRaw = read(`${workerPath}/wrangler.jsonc`)
const config = JSON.parse(configRaw)
const infraReadme = read(`${workerPath}/README.md`)
const workflow = read(
  '.github/workflows/production-traffic-ingestion.yml'
)
const packageJSON = JSON.parse(read('package.json'))
const releaseVerifier = read('scripts/verify-release.sh')
const backendCI = read('.github/workflows/backend-ci.yml')
const runbook = read('docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md')
const foundation = read(
  'docs/182_ZERO_COST_PRODUCTION_INGESTION_RELIABILITY.md'
)
const documentIndex = read('docs/DOCUMENT_INDEX.md')

assert(
  config.main === 'src/index.mjs',
  'Wrangler must use the repository Worker entry point'
)
assert(
  config.compatibility_date === '2026-08-01',
  'Wrangler compatibility date must be explicit'
)
assert(
  config.workers_dev === true,
  'production health verification requires the stable workers.dev route'
)
assert(
  config.preview_urls === false,
  'production scheduler must disable public versioned preview URLs'
)
assert(
  typeof config.name === 'string' &&
    config.name.length > 0 &&
    config.name.length <= 63,
  'workers.dev Worker name must fit the DNS label limit'
)
assert(
  Array.isArray(config.triggers?.crons) &&
    config.triggers.crons.length === 2,
  'Worker must define exactly two Cron Triggers'
)
assert(
  config.triggers.crons.includes(
    '3,13,23,33,43,53 * * * *'
  ),
  'Worker must define the offset ten-minute primary schedule'
)
assert(
  config.triggers.crons.includes('*/5 * * * *'),
  'Worker must define the five-minute watchdog'
)
assert(
  !Object.hasOwn(config.vars ?? {}, 'GITHUB_ACTIONS_TOKEN'),
  'GitHub token must not be stored in Wrangler variables'
)
assert(
  !configRaw.includes('Bearer '),
  'Wrangler configuration must not contain a bearer token'
)
assert(
  source.includes('async scheduled(controller, environment)'),
  'Worker must expose a scheduled handler'
)
assert(
  source.includes("url.pathname !== '/health'"),
  'Worker must expose only the bounded health route'
)
assert(
  source.includes("status !== 'completed'"),
  'unknown non-completed workflow states must block dispatch'
)
assert(
  source.includes('dispatch_source: dispatchSource'),
  'workflow dispatch must preserve dispatch provenance'
)
assert(
  source.includes('response.status !== 204'),
  'workflow dispatch must require the exact GitHub status'
)
assert(
  source.includes('STALE_TRAFFIC_RECOVERY_DISPATCH'),
  'watchdog must publish a stable recovery marker'
)
assert(
  source.includes('ACTIVE_RUN_DEDUPLICATION'),
  'Worker must publish active-run suppression evidence'
)
assert(
  tests.includes('watchdog dispatches exactly once'),
  'Worker tests must cover stale recovery'
)
assert(
  tests.includes('suppresses an active run'),
  'Worker tests must cover active-run deduplication'
)
assert(
  workflow.includes("cron: '*/10 * * * *'"),
  'foundation must not weaken the active GitHub schedule before deployment'
)
assert(
  workflow.includes('dispatch_source:'),
  'workflow must declare dispatch provenance input'
)
assert(
  workflow.includes('cloudflare-primary|cloudflare-watchdog'),
  'workflow must validate Cloudflare dispatch sources'
)
assert(
  packageJSON.scripts[
    'test:production-ingestion-reliability'
  ] ===
    'node --test infra/cloudflare/production-ingestion-reliability/test/*.test.mjs scripts/verify-production-ingestion-reliability.test.mjs',
  'package scripts must expose reliability tests'
)
assert(
  packageJSON.scripts[
    'verify:production-ingestion-reliability'
  ] ===
    'node scripts/verify-production-ingestion-reliability.mjs',
  'package scripts must expose reliability verification'
)
assert(
  releaseVerifier.includes(
    'test:production-ingestion-reliability'
  ) &&
    releaseVerifier.includes(
      'verify:production-ingestion-reliability'
    ),
  'release verification must enforce the Worker foundation'
)
assert(
  backendCI.includes("'infra/cloudflare/**'"),
  'Backend CI push paths must include the Worker foundation'
)
assert(
  backendCI.includes(
    'Verify zero-cost production ingestion reliability foundation'
  ),
  'Backend CI must run the Worker foundation contract'
)
assert(
  infraReadme.includes('deployment not yet verified') &&
    foundation.includes(
      'cloud deployment and production cutover not yet verified'
    ),
  'documentation must preserve deployment truth'
)
assert(
  runbook.includes(
    'Repository foundation status: implemented, not deployed.'
  ),
  'production runbook must record foundation status'
)
assert(
  documentIndex.includes(
    'Document 182 — Zero-Cost Production Ingestion Reliability Foundation'
  ),
  'document index must register the foundation'
)
assert(
  !fs.existsSync(
    path.join(repositoryRoot, workerPath, '.dev.vars')
  ),
  'repository must not contain local Worker secrets'
)

console.log('CLOUDFLARE_RELIABILITY_SOURCE_CONTRACT=PASS')
console.log('ZERO_COST_INGESTION_RELIABILITY_FOUNDATION=PASS')
