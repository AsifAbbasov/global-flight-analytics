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

test('live deployment evidence is recorded without claiming final closure', () => {
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
    /Cloudflare Worker deployed and initial live evidence verified/
  )
  assert.match(
    runbook,
    /Cloudflare live deployment status: verified on 2026-08-06/
  )
  assert.match(
    liveEvidence,
    /PRODUCTION_INGESTION_RELIABILITY=PENDING/
  )
  assert.doesNotMatch(
    liveEvidence,
    /PRODUCTION_INGESTION_RELIABILITY=PASS/
  )
})
