#!/usr/bin/env node
import fs from 'node:fs'
import { execFileSync } from 'node:child_process'

function fail(message) {
  console.error(`GRAFANA_OBSERVABILITY_CONTRACT=FAIL reason=${message}`)
  process.exit(1)
}
const read = (path) => fs.readFileSync(path, 'utf8')
const parse = (path) => JSON.parse(read(path))

const dashboard = parse('monitoring/grafana-cloud/dashboard.json')
const alerts = parse('monitoring/grafana-cloud/alert-rules.json')
const workflow = read('.github/workflows/provision-grafana-observability.yml')
const provision = read('scripts/provision-grafana-observability.sh')
const packageJson = parse('package.json')
const release = read('scripts/verify-release.sh')
const backendCI = read('.github/workflows/backend-ci.yml')
const runbook = read('docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md')

if (dashboard.metadata?.name !== 'gfa-production-slo') fail('dashboard resource name is not stable')
if (dashboard.metadata?.annotations?.['grafana.app/folder'] !== 'gfa-observability') fail('dashboard folder is not stable')
if (!Array.isArray(dashboard.spec?.panels) || dashboard.spec.panels.length < 12) fail('dashboard must contain at least twelve production panels')
const variable = dashboard.spec?.templating?.list?.find((item) => item.name === 'deployment_revision')
if (!variable || variable.allValue !== '.*') fail('dashboard must expose an explicit deployment revision variable')
const dashboardText = JSON.stringify(dashboard)
for (const metric of [
  'global_flight_analytics_http_requests_total',
  'global_flight_analytics_http_request_duration_seconds_bucket',
  'global_flight_analytics_ingestion_latest_finished_age_seconds',
  'global_flight_analytics_ingestion_consecutive_failures',
  'global_flight_analytics_postgres_pool_acquired_connections',
  'global_flight_analytics_reconciliation_oldest_pending_age_seconds',
  'global_flight_analytics_collector_last_scrape_success',
]) if (!dashboardText.includes(metric)) fail(`dashboard is missing ${metric}`)

const expectedRules = new Map([
  ['gfa-api-availability', ['15m', 99.5, 'lt']],
  ['gfa-api-p95-latency', ['15m', 2, 'gt']],
  ['gfa-api-server-errors', ['10m', 0.01, 'gt']],
  ['gfa-ingestion-freshness', ['10m', 120, 'gt']],
  ['gfa-ingestion-failures', ['1m', 3, 'gt']],
  ['gfa-postgres-pool', ['10m', 0.8, 'gt']],
  ['gfa-reconciliation-backlog', ['15m', 300, 'gt']],
  ['gfa-collector-health', ['2m', 1, 'lt']],
  ['gfa-metrics-missing', ['1m', 0, 'gt']],
])
if (alerts.folderUid !== 'gfa-observability' || alerts.title !== 'global-flight-analytics-production-slo' || alerts.interval !== 60) fail('alert group identity is invalid')
if (!Array.isArray(alerts.rules) || alerts.rules.length !== expectedRules.size) fail('alert rule count is invalid')
for (const current of alerts.rules) {
  const expected = expectedRules.get(current.uid)
  if (!expected) fail(`unexpected alert rule ${current.uid}`)
  if (current.for !== expected[0]) fail(`${current.uid} has an invalid evaluation duration`)
  const condition = current.data?.find((item) => item.refId === 'B')?.model?.conditions?.[0]?.evaluator
  if (!condition || condition.params?.[0] !== expected[1] || condition.type !== expected[2]) fail(`${current.uid} has an invalid threshold`)
  if (current.labels?.environment !== 'production' || current.labels?.slo !== 'true') fail(`${current.uid} lacks bounded production labels`)
  if (current.uid !== 'gfa-metrics-missing' && current.noDataState !== 'OK') fail(`${current.uid} must avoid false no-data pages between scheduled scrapes`)
}

for (const literal of [
  'workflow_dispatch:',
  'GRAFANA_SERVICE_ACCOUNT_TOKEN: ${{ secrets.GRAFANA_SERVICE_ACCOUNT_TOKEN }}',
  'bash scripts/provision-grafana-observability.sh',
  'GRAFANA_OBSERVABILITY_WORKFLOW=PASS',
]) if (!workflow.includes(literal)) fail(`provisioning workflow is missing ${literal}`)
if (workflow.includes('schedule:')) fail('provisioning workflow must not mutate Grafana on a schedule')

for (const literal of [
  '/apis/folder.grafana.app/v1/namespaces/default/folders',
  '/apis/dashboard.grafana.app/v1/namespaces/default/dashboards',
  '/api/v1/provisioning/folder/gfa-observability/rule-groups/global-flight-analytics-production-slo',
  '/api/v1/provisioning/policies',
  'GRAFANA_NOTIFICATION_POLICY=PASS',
]) if (!provision.includes(literal)) fail(`provisioning script is missing ${literal}`)
if (/--request\s+PUT[^\n]+\/api\/v1\/provisioning\/policies/.test(provision)) fail('provisioning must not overwrite the notification policy tree')

const entries = {
  'test:grafana-observability': 'node --test scripts/verify-grafana-observability.test.mjs',
  'verify:grafana-observability': 'node scripts/verify-grafana-observability.mjs',
  'provision:grafana-observability': 'bash scripts/provision-grafana-observability.sh',
}
for (const [name, command] of Object.entries(entries)) if (packageJson.scripts?.[name] !== command) fail(`package script ${name} is missing`)
for (const command of ['pnpm run test:grafana-observability', 'pnpm run verify:grafana-observability']) {
  if (!release.includes(command)) fail(`release verification is missing ${command}`)
  if (!backendCI.includes(command)) fail(`Backend CI is missing ${command}`)
}
for (const literal of ['GRAFANA_INSTANCE_URL', 'GRAFANA_PROMETHEUS_DATASOURCE_UID', 'GRAFANA_SERVICE_ACCOUNT_TOKEN', 'GRAFANA_ALERT_RULES=PASS', 'notification delivery']) {
  if (!runbook.includes(literal)) fail(`runbook is missing ${literal}`)
}

const temp = fs.mkdtempSync('/tmp/gfa-grafana-render-')
try {
  const output = execFileSync(process.execPath, ['scripts/render-grafana-observability.mjs', temp], {encoding:'utf8', env:{...process.env, GRAFANA_PROMETHEUS_DATASOURCE_UID:'prometheus-test-uid'}})
  if (!output.includes('GRAFANA_OBSERVABILITY_RENDER=PASS')) fail('renderer did not pass')
  const renderedDashboard = JSON.parse(fs.readFileSync(`${temp}/dashboard.json`, 'utf8'))
  const renderedAlerts = JSON.parse(fs.readFileSync(`${temp}/alert-rules.json`, 'utf8'))
  if (JSON.stringify(renderedDashboard).includes('__PROMETHEUS_DATASOURCE_UID__')) fail('dashboard placeholder remains after rendering')
  if (JSON.stringify(renderedAlerts).includes('__PROMETHEUS_DATASOURCE_UID__')) fail('alert placeholder remains after rendering')
} finally { fs.rmSync(temp, {recursive:true, force:true}) }

console.log('GRAFANA_OBSERVABILITY_CONTRACT=PASS')
