#!/usr/bin/env node

import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

function read(relativePath) {
  return fs.readFileSync(path.join(repositoryRoot, relativePath), 'utf8')
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

function parseScheduledMinutes(workflow) {
  const match = workflow.match(/cron: '([^']+)'/)
  assert(match, 'production ingestion workflow must define a cron schedule')
  const cron = match[1]
  const [minuteField, hourField, dayOfMonth, month, dayOfWeek] = cron.split(/\s+/)
  assert(
    hourField === '*' && dayOfMonth === '*' && month === '*' && dayOfWeek === '*',
    'production ingestion GitHub fallback must remain an every-hour minute schedule',
  )
  const minutes = minuteField.split(',').map(Number)
  assert(
    minutes.length >= 2 && minutes.every((value) => Number.isInteger(value) && value >= 0 && value <= 59),
    'production ingestion GitHub fallback must use explicit bounded minute offsets',
  )
  const sorted = [...new Set(minutes)].sort((a, b) => a - b)
  const gaps = sorted.map((minute, index) => {
    const next = sorted[(index + 1) % sorted.length]
    return index === sorted.length - 1 ? 60 - minute + next : next - minute
  })
  return { cron, minGapMinutes: Math.min(...gaps), maxGapMinutes: Math.max(...gaps) }
}

const workerPath = 'infra/cloudflare/production-ingestion-reliability'
const source = read(`${workerPath}/src/index.mjs`)
const tests = read(`${workerPath}/test/index.test.mjs`)
const freeTierTests = read(`${workerPath}/test/free-tier-budget.test.mjs`)
const configRaw = read(`${workerPath}/wrangler.jsonc`)
const config = JSON.parse(configRaw)
const infraReadme = read(`${workerPath}/README.md`)
const workflow = read('.github/workflows/production-traffic-ingestion.yml')
const fallbackSchedule = parseScheduledMinutes(workflow)
const packageJSON = JSON.parse(read('package.json'))
const releaseVerifier = read('scripts/verify-release.sh')
const backendCI = read('.github/workflows/backend-ci.yml')
const runbook = read('docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md')
const foundation = read('docs/182_ZERO_COST_PRODUCTION_INGESTION_RELIABILITY.md')
const liveEvidence = read('docs/183_CLOUDFLARE_INGESTION_LIVE_DEPLOYMENT_EVIDENCE.md')
const documentIndex = read('docs/DOCUMENT_INDEX.md')
const readme = read('README.md')

assert(config.main === 'src/index.mjs', 'Wrangler must use the repository Worker entry point')
assert(config.compatibility_date === '2026-08-01', 'Wrangler compatibility date must be explicit')
assert(config.workers_dev === true, 'production health verification requires the stable workers.dev route')
assert(config.preview_urls === false, 'production scheduler must disable public versioned preview URLs')
assert(typeof config.name === 'string' && config.name.length > 0 && config.name.length <= 63, 'workers.dev Worker name must fit the DNS label limit')
assert(Array.isArray(config.triggers?.crons) && config.triggers.crons.length === 2, 'Worker must define exactly two Cron Triggers')
assert(config.triggers.crons.includes('17,47 * * * *'), 'Worker must define the free-tier 30-minute primary schedule')
assert(config.triggers.crons.includes('19 */2 * * *'), 'Worker must define the free-tier two-hour watchdog')
assert(!config.triggers.crons.includes('3,13,23,33,43,53 * * * *'), 'Worker must not retain the ten-minute keep-awake primary')
assert(!config.triggers.crons.includes('*/5 * * * *'), 'Worker must not retain the five-minute keep-awake watchdog')
assert(!Object.hasOwn(config.vars ?? {}, 'GITHUB_ACTIONS_TOKEN'), 'GitHub token must not be stored in Wrangler variables')
assert(!configRaw.includes('Bearer '), 'Wrangler configuration must not contain a bearer token')
assert(config.vars?.DISPATCH_ENABLED === 'false', 'production dispatch must remain fail-closed during recovery')
assert(config.vars?.RECENT_FAILURE_COOLDOWN_SECONDS === '21600', 'production Worker must define the six-hour recent-failure cooldown')
assert(source.includes("primaryCron: '17,47 * * * *'"), 'Worker source default must match free-tier primary cadence')
assert(source.includes("watchdogCron: '19 */2 * * *'"), 'Worker source default must match free-tier watchdog cadence')
assert(source.includes('async scheduled(controller, environment)'), 'Worker must expose a scheduled handler')
assert(source.includes("url.pathname !== '/health'"), 'Worker must expose only the bounded health route')
assert(source.includes("status !== 'completed'"), 'unknown non-completed workflow states must block dispatch')
assert(source.includes('dispatch_source: dispatchSource'), 'workflow dispatch must preserve dispatch provenance')
assert(source.includes('response.status !== 204'), 'workflow dispatch must require the exact GitHub status')
assert(source.includes('STALE_TRAFFIC_RECOVERY_DISPATCH'), 'watchdog must publish a stable recovery marker')
assert(source.includes('ACTIVE_RUN_DEDUPLICATION'), 'Worker must publish active-run suppression evidence')
assert(source.includes('RECENT_FAILURE_CIRCUIT_BREAKER'), 'Worker must publish recent-failure circuit-breaker evidence')
assert(source.includes('CLOUDFLARE_DISPATCH_DISABLED'), 'Worker must publish fail-closed dispatch-disabled evidence')
assert(tests.includes('watchdog dispatches exactly once'), 'Worker tests must cover stale recovery')
assert(tests.includes('suppresses an active run'), 'Worker tests must cover active-run deduplication')
assert(tests.includes('dispatch kill switch suppresses both Cron Triggers'), 'Worker tests must cover the fail-closed dispatch kill switch')
assert(tests.includes('recent failed workflow run opens the circuit breaker'), 'Worker tests must cover recent-failure circuit breaking')
assert(freeTierTests.includes('Worker source defaults preserve the free-tier cadence'), 'free-tier tests must protect source defaults')
assert(fallbackSchedule.maxGapMinutes <= 15, 'temporary GitHub fallback must run at least every fifteen minutes until scheduler ownership cutover')
assert(fallbackSchedule.minGapMinutes >= 15, 'temporary GitHub fallback must not run more frequently than every fifteen minutes')
assert(!workflow.includes("cron: '*/10 * * * *'"), 'GitHub Actions must not retain the ten-minute primary')
assert(workflow.includes('dispatch_source:'), 'workflow must declare dispatch provenance input')
assert(workflow.includes('cloudflare-primary|cloudflare-watchdog'), 'workflow must validate Cloudflare dispatch sources')
assert(packageJSON.scripts['test:production-ingestion-reliability'] === 'node --test infra/cloudflare/production-ingestion-reliability/test/*.test.mjs scripts/verify-production-ingestion-reliability.test.mjs', 'package scripts must expose reliability tests')
assert(packageJSON.scripts['verify:production-ingestion-reliability'] === 'node scripts/verify-production-ingestion-reliability.mjs', 'package scripts must expose reliability verification')
assert(releaseVerifier.includes('test:production-ingestion-reliability') && releaseVerifier.includes('verify:production-ingestion-reliability'), 'release verification must enforce the Worker foundation')
assert(backendCI.includes("'infra/cloudflare/**'"), 'Backend CI push paths must include the Worker foundation')
assert(backendCI.includes('Verify zero-cost production ingestion reliability foundation'), 'Backend CI must run the Worker foundation contract')
assert(infraReadme.includes('Status: production ingestion reliability implementation is hardened;'), 'Worker README must describe current fail-closed recovery state')
assert(foundation.includes('Status: production ingestion reliability closed.'), 'historical reliability closure evidence must remain preserved')
assert(runbook.includes('Cloudflare live deployment status: verified on 2026-08-06.'), 'production runbook must record historical live deployment status')
assert(liveEvidence.includes('CLOUDFLARE_WORKER_DEPLOYMENT=PASS') && liveEvidence.includes('CLOUDFLARE_PRIMARY_REAL_DISPATCH=PASS') && liveEvidence.includes('ACTIVE_RUN_DEDUPLICATION=PASS') && liveEvidence.includes('GITHUB_SCHEDULED_FALLBACK=PASS') && liveEvidence.includes('LIVE_PRODUCTION_RUNTIME_VALIDATION=PASS') && liveEvidence.includes('PRODUCTION_INGESTION_RELIABILITY=PASS') && !liveEvidence.includes('PRODUCTION_INGESTION_RELIABILITY=PENDING'), 'historical live evidence must remain complete')

const closureEvidenceDocuments = [
  ['production runbook', runbook],
  ['reliability closure', foundation],
  ['live closure evidence', liveEvidence],
]
const expectedClosureEvidence = [
  'CLOSURE_REPOSITORY_REVISION=7dfc66685247a5a1aaea87b1391624d1014d7013',
  'WATCHDOG_RECOVERY_RUN_ID=31103550357',
  'PRIMARY_DISPATCH_RUN_ID=31112274607',
  'PRIMARY_DISPATCH_HEAD_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013',
  'ACTIVE_RUN_AND_FALLBACK_RUN_ID=31113114700',
  'ACTIVE_RUN_AND_FALLBACK_HEAD_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013',
  'FINAL_RUNTIME_VALIDATION_SHA=7dfc66685247a5a1aaea87b1391624d1014d7013',
  'FINAL_RUNTIME_VALIDATION_COMPLETED_AT=2026-08-06T15:31:58Z',
]
const forbiddenPendingEvidence = [
  'CLOUDFLARE_PRIMARY_REAL_DISPATCH=PENDING',
  'ACTIVE_RUN_DEDUPLICATION=PENDING',
  'GITHUB_SCHEDULED_FALLBACK=PENDING',
  'FINAL_EXACT_REVISION_RUNTIME_VALIDATION=PENDING',
  'PRODUCTION_INGESTION_RELIABILITY=PENDING',
]
for (const [documentName, document] of closureEvidenceDocuments) {
  for (const expectedEvidence of expectedClosureEvidence) assert(document.includes(expectedEvidence), `${documentName} must record ${expectedEvidence}`)
  for (const pendingEvidence of forbiddenPendingEvidence) assert(!document.includes(pendingEvidence), `${documentName} must not retain ${pendingEvidence}`)
  assert(document.includes('Evidence ownership and verification boundary'), `${documentName} must explain the evidence verification boundary`)
}

assert(readme.includes('repository-recorded closure evidence') && readme.includes('owner-local, non-secret supporting') && !readme.includes('Diagnosis, recovery commands, architecture boundaries, and immutable run'), 'README must describe the repository and owner-local evidence boundary accurately')
assert(readme.split('docs/182_ZERO_COST_PRODUCTION_INGESTION_RELIABILITY.md').length - 1 === 2 && readme.split('docs/183_CLOUDFLARE_INGESTION_LIVE_DEPLOYMENT_EVIDENCE.md').length - 1 === 2, 'README must link each reliability closure document exactly once')
assert(documentIndex.includes('Document 182 — Zero-Cost Production Ingestion Reliability Closure'), 'document index must register the reliability closure')
assert(documentIndex.includes('Document 183 — Cloudflare Ingestion Live Deployment and Closure Evidence'), 'document index must register the live deployment and closure evidence')
assert(!fs.existsSync(path.join(repositoryRoot, workerPath, '.dev.vars')), 'repository must not contain local Worker secrets')

console.log(`GITHUB_FALLBACK_CRON=${fallbackSchedule.cron} MIN_GAP_MINUTES=${fallbackSchedule.minGapMinutes} MAX_GAP_MINUTES=${fallbackSchedule.maxGapMinutes}`)
console.log('CLOUDFLARE_FREE_TIER_SCHEDULER=PASS')
console.log('CLOUDFLARE_RELIABILITY_SOURCE_CONTRACT=PASS')
console.log('ZERO_COST_INGESTION_RELIABILITY_FOUNDATION=PASS')
console.log('PRODUCTION_INGESTION_RELIABILITY=PASS')
