import assert from 'node:assert/strict'
import fs from 'node:fs'
import { execFileSync } from 'node:child_process'
import test from 'node:test'

const dashboard = JSON.parse(fs.readFileSync('monitoring/grafana-cloud/dashboard.json', 'utf8'))
const alerts = JSON.parse(fs.readFileSync('monitoring/grafana-cloud/alert-rules.json', 'utf8'))
const provision = fs.readFileSync('scripts/provision-grafana-observability.sh', 'utf8')
const workflow = fs.readFileSync('.github/workflows/provision-grafana-observability.yml', 'utf8')
const backendCI = fs.readFileSync('.github/workflows/backend-ci.yml', 'utf8')

test('Grafana observability contract passes', () => {
  const output = execFileSync(process.execPath, ['scripts/verify-grafana-observability.mjs'], {encoding:'utf8'})
  assert.match(output, /GRAFANA_OBSERVABILITY_CONTRACT=PASS/)
})

test('dashboard exposes production SLO evidence by deployment revision', () => {
  assert.equal(dashboard.metadata.name, 'gfa-production-slo')
  assert.ok(dashboard.spec.panels.length >= 12)
  assert.equal(dashboard.spec.templating.list[0].name, 'deployment_revision')
})

test('alert group uses modern Math conditions with exact SLO semantics', () => {
  assert.equal(alerts.rules.length, 9)
  const expressions = new Map(alerts.rules.map((rule) => [rule.uid, rule.data.find((item) => item.refId === 'B')?.model?.expression]))
  assert.equal(expressions.get('gfa-api-server-errors'), '$A >= 0.01')
  assert.equal(expressions.get('gfa-ingestion-failures'), '$A >= 3')
  for (const rule of alerts.rules) {
    const condition = rule.data.find((item) => item.refId === 'B')
    assert.equal(condition?.model?.type, 'math')
    assert.doesNotMatch(JSON.stringify(condition), /classic_conditions/)
  }
})

test('Grafana Cloud provisioning uses the stack namespace', () => {
  assert.match(workflow, /GRAFANA_STACK_ID: \$\{\{ vars\.GRAFANA_STACK_ID \}\}/)
  assert.match(provision, /GRAFANA_NAMESPACE="stacks-\$GRAFANA_STACK_ID"/)
  assert.match(provision, /namespaces\/\$GRAFANA_NAMESPACE\/folders/)
  assert.match(provision, /namespaces\/\$GRAFANA_NAMESPACE\/dashboards/)
  assert.doesNotMatch(provision, /namespaces\/default/)
})

test('backend container contract is self-contained without pnpm', () => {
  const start = backendCI.indexOf('      - name: Verify Grafana SLO dashboard and alert contract')
  const end = backendCI.indexOf('      - name: Verify recruiter quickstart contract', start)
  assert.ok(start >= 0 && end > start)
  const step = backendCI.slice(start, end)
  assert.match(step, /node --test scripts\/verify-grafana-observability\.test\.mjs/)
  assert.match(step, /node scripts\/verify-grafana-observability\.mjs/)
  assert.doesNotMatch(step, /pnpm /)
})

test('provisioning is idempotent and preserves notification policy ownership', () => {
  assert.match(provision, /upsert_resource folder/)
  assert.match(provision, /upsert_resource dashboard/)
  assert.match(provision, /rule-groups\/global-flight-analytics-production-slo/)
  assert.doesNotMatch(provision, /--request\s+PUT[^\n]+\/api\/v1\/provisioning\/policies/)
})
