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

test('foundation keeps GitHub primary cadence until live cutover', () => {
  const workflow = read(
    '.github/workflows/production-traffic-ingestion.yml'
  )
  const config = JSON.parse(
    read(
      'infra/cloudflare/production-ingestion-reliability/wrangler.jsonc'
    )
  )

  assert.match(workflow, /cron: '\*\/10 \* \* \* \*'/)
  assert.deepEqual(config.triggers.crons, [
    '3,13,23,33,43,53 * * * *',
    '*/5 * * * *',
  ])
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

test('deployment truth remains explicitly open', () => {
  const foundation = read(
    'docs/182_ZERO_COST_PRODUCTION_INGESTION_RELIABILITY.md'
  )
  const runbook = read(
    'docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md'
  )

  assert.match(
    foundation,
    /deployment and production cutover not yet verified/
  )
  assert.match(
    runbook,
    /implemented, not deployed/
  )
  assert.match(
    foundation,
    /PRODUCTION_INGESTION_RELIABILITY=PASS` must not be\s+claimed/
  )
})
