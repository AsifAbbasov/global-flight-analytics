import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { spawnSync } from 'node:child_process'
import test from 'node:test'
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

test('repository reliability verifier accepts the current checkout', () => {
  const result = spawnSync(
    process.execPath,
    ['scripts/verify-production-ingestion-reliability.mjs'],
    {
      cwd: repositoryRoot,
      encoding: 'utf8',
    }
  )

  assert.equal(
    result.status,
    0,
    `${result.stdout}\n${result.stderr}`
  )
  assert.match(
    result.stdout,
    /CLOUDFLARE_RELIABILITY_SOURCE_CONTRACT=PASS/
  )
})

test('cutover keeps Cloudflare primary and GitHub hourly fallback', () => {
  const workflow = read(
    '.github/workflows/production-traffic-ingestion.yml'
  )
  const config = JSON.parse(
    read(
      'infra/cloudflare/production-ingestion-reliability/wrangler.jsonc'
    )
  )

  assert.match(workflow, /cron: '37 \* \* \* \*'/)
  assert.doesNotMatch(workflow, /cron: '\*\/10 \* \* \* \*'/)
  assert.deepEqual(config.triggers.crons, [
    '3,13,23,33,43,53 * * * *',
    '*/5 * * * *',
  ])
})

test('production Worker disables preview URLs while keeping workers.dev', () => {
  const config = JSON.parse(
    read(
      'infra/cloudflare/production-ingestion-reliability/wrangler.jsonc'
    )
  )

  assert.equal(config.workers_dev, true)
  assert.equal(config.preview_urls, false)
  assert.equal(typeof config.name, 'string')
  assert.ok(config.name.length > 0)
  assert.ok(
    config.name.length <= 63,
    `workers.dev Worker name is too long: ${config.name.length}`
  )
})

test('repository configuration excludes the GitHub token', () => {
  const config = JSON.parse(
    read(
      'infra/cloudflare/production-ingestion-reliability/wrangler.jsonc'
    )
  )
  const source = read(
    'infra/cloudflare/production-ingestion-reliability/src/index.mjs'
  )

  assert.equal(
    Object.hasOwn(config.vars, 'GITHUB_ACTIONS_TOKEN'),
    false
  )
  assert.match(source, /GITHUB_ACTIONS_TOKEN/)
  assert.doesNotMatch(
    JSON.stringify(config),
    /Authorization|Bearer/
  )
})

test('live deployment evidence records final reliability closure', () => {
  const foundation = read(
    'docs/182_ZERO_COST_PRODUCTION_INGESTION_RELIABILITY.md'
  )
  const runbook = read(
    'docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md'
  )
  const liveEvidence = read(
    'docs/183_CLOUDFLARE_INGESTION_LIVE_DEPLOYMENT_EVIDENCE.md'
  )

  assert.match(
    foundation,
    /Status: production ingestion reliability closed\./
  )
  assert.match(
    runbook,
    /The production ingestion reliability stage is closed\./
  )
  assert.match(
    liveEvidence,
    /CLOUDFLARE_PRIMARY_REAL_DISPATCH=PASS/
  )
  assert.match(
    liveEvidence,
    /ACTIVE_RUN_DEDUPLICATION=PASS/
  )
  assert.match(
    liveEvidence,
    /GITHUB_SCHEDULED_FALLBACK=PASS/
  )
  assert.match(
    liveEvidence,
    /LIVE_PRODUCTION_RUNTIME_VALIDATION=PASS/
  )
  assert.match(
    liveEvidence,
    /PRODUCTION_INGESTION_RELIABILITY=PASS/
  )
  assert.doesNotMatch(
    liveEvidence,
    /PRODUCTION_INGESTION_RELIABILITY=PENDING/
  )
})
test('closure evidence identities remain consistent across documents', () => {
  const readme = read('README.md')
  const documents = [
    read('docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md'),
    read('docs/182_ZERO_COST_PRODUCTION_INGESTION_RELIABILITY.md'),
    read('docs/183_CLOUDFLARE_INGESTION_LIVE_DEPLOYMENT_EVIDENCE.md'),
  ]
  const expectedEvidence = [
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

  for (const document of documents) {
    for (const evidence of expectedEvidence) {
      assert.match(document, new RegExp(evidence))
    }
    for (const pendingEvidence of forbiddenPendingEvidence) {
      assert.doesNotMatch(document, new RegExp(pendingEvidence))
    }
    assert.match(
      document,
      /Evidence ownership and verification boundary/
    )
  }

  assert.match(readme, /repository-recorded closure evidence/)
  assert.match(readme, /owner-local, non-secret supporting\s+evidence/)
  assert.doesNotMatch(
    readme,
    /Diagnosis, recovery commands, architecture boundaries, and immutable run/
  )
})
