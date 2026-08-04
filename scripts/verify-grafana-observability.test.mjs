import assert from 'node:assert/strict'
import fs from 'node:fs'
import { execFileSync } from 'node:child_process'
import test from 'node:test'

const dashboard = JSON.parse(fs.readFileSync('monitoring/grafana-cloud/dashboard.json', 'utf8'))
const alerts = JSON.parse(fs.readFileSync('monitoring/grafana-cloud/alert-rules.json', 'utf8'))
const provision = fs.readFileSync('scripts/provision-grafana-observability.sh', 'utf8')

test('Grafana observability contract passes', () => {
  const output = execFileSync(process.execPath, ['scripts/verify-grafana-observability.mjs'], {encoding:'utf8'})
  assert.match(output, /GRAFANA_OBSERVABILITY_CONTRACT=PASS/)
})

test('dashboard exposes production SLO evidence by deployment revision', () => {
  assert.equal(dashboard.metadata.name, 'gfa-production-slo')
  assert.ok(dashboard.spec.panels.length >= 12)
  assert.equal(dashboard.spec.templating.list[0].name, 'deployment_revision')
})

test('alert group encodes all initial SLO thresholds', () => {
  assert.equal(alerts.rules.length, 9)
  assert.deepEqual(new Set(alerts.rules.map((rule) => rule.uid)), new Set([
    'gfa-api-availability','gfa-api-p95-latency','gfa-api-server-errors','gfa-ingestion-freshness','gfa-ingestion-failures','gfa-postgres-pool','gfa-reconciliation-backlog','gfa-collector-health','gfa-metrics-missing',
  ]))
})

test('provisioning is idempotent and preserves notification policy ownership', () => {
  assert.match(provision, /upsert_resource folder/)
  assert.match(provision, /upsert_resource dashboard/)
  assert.match(provision, /rule-groups\/global-flight-analytics-production-slo/)
  assert.doesNotMatch(provision, /--request\s+PUT[^\n]+\/api\/v1\/provisioning\/policies/)
})
