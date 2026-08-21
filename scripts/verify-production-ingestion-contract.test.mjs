import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '..'
)

function read(relativePath) {
  return fs.readFileSync(path.join(repositoryRoot, relativePath), 'utf8')
}

test('production ingestion remains bounded serialized and dispatch-owned', () => {
  const workflow = read('.github/workflows/production-traffic-ingestion.yml')

  assert.match(workflow, /workflow_dispatch:/)
  assert.doesNotMatch(workflow, /\bschedule:/)
  assert.doesNotMatch(workflow, /cron:/)
  assert.match(workflow, /manual\|cloudflare-primary\|cloudflare-watchdog/)
  assert.doesNotMatch(workflow, /manual\|schedule\|cloudflare-primary\|cloudflare-watchdog/)
  assert.match(workflow, /SCHEDULER-OWNERSHIP-V1/)
  assert.match(workflow, /group: production-traffic-ingestion/)
  assert.match(workflow, /cancel-in-progress: false/)
  assert.match(workflow, /go run \.\/cmd\/ingest --once/)
  assert.match(workflow, /PRODUCTION_INGESTION_DATABASE_URL/)
  assert.match(workflow, /TRAFFIC_INGESTION_RADIUS: '250'/)
  assert.match(workflow, /node-version: '24\.9\.0'/)
})

test('Cloudflare-owned ingestion keeps the 1800 second freshness budget in the dispatched run', () => {
  const workflow = read('.github/workflows/production-traffic-ingestion.yml')

  assert.doesNotMatch(workflow, /cron:/)
  assert.match(workflow, /MAX_TRAFFIC_AGE_SECONDS: '1800'/)
  assert.match(workflow, /cloudflare-primary/)
  assert.match(workflow, /cloudflare-watchdog/)
})

test('dispatched production ingestion verifies public freshness', () => {
  const workflow = read('.github/workflows/production-traffic-ingestion.yml')
  const freshness = read(
    'scripts/verify-production-traffic-freshness.mjs'
  )

  assert.match(workflow, /verify-production-traffic-freshness\.mjs/)
  assert.match(workflow, /MAX_TRAFFIC_AGE_SECONDS: '1800'/)
  assert.match(freshness, /PRODUCTION_TRAFFIC_FRESHNESS=PASS/)
  assert.match(freshness, /production traffic is stale/)
})

test('one-shot ingestion remains an explicit command mode', () => {
  const main = read('apps/api/cmd/ingest/main.go')
  const options = read('apps/api/cmd/ingest/command_options.go')
  const runOnce = read('apps/api/cmd/ingest/run_once.go')

  assert.match(main, /runWithArgs\(os\.Args\[1:\]\)/)
  assert.match(main, /if commandOptions\.Once \{/)
  assert.match(options, /"once"/)
  assert.match(runOnce, /runCycle\(ctx\)/)
})

test('free production topology does not add a Render ingestion worker', () => {
  const renderBlueprint = read('render.yaml')

  assert.equal(
    (renderBlueprint.match(/dockerCommand: \/app\/server/g) ?? []).length,
    1
  )
  assert.equal(
    (renderBlueprint.match(/dockerCommand: \/app\/ingest/g) ?? []).length,
    0
  )
})
