#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'

const PRODUCTION_TRAFFIC_FRESHNESS_BUDGET_SECONDS = 1800

function fail(message) {
  console.error(`GRAFANA_OBSERVABILITY_RENDER=FAIL reason=${message}`)
  process.exit(1)
}

const outputDirectory = process.argv[2]
if (!outputDirectory) fail('output directory argument is required')
const datasourceUID = process.env.GRAFANA_PROMETHEUS_DATASOURCE_UID?.trim()
if (!datasourceUID || !/^[A-Za-z0-9_-]{3,128}$/.test(datasourceUID)) {
  fail('GRAFANA_PROMETHEUS_DATASOURCE_UID must be a bounded Grafana UID')
}

function applyProductionSLOPolicy(name, parsed) {
  if (name !== 'alert-rules.json') return parsed

  const freshnessRule = parsed.rules?.find((candidate) => candidate.uid === 'gfa-ingestion-freshness')
  if (!freshnessRule) fail('alert-rules.json is missing gfa-ingestion-freshness')

  const freshnessCondition = freshnessRule.data?.find((item) => item.refId === 'B')
  if (freshnessCondition?.model?.type !== 'math') {
    fail('gfa-ingestion-freshness must use a Grafana-managed Math condition')
  }

  freshnessCondition.model.expression = `$A > ${PRODUCTION_TRAFFIC_FRESHNESS_BUDGET_SECONDS}`
  freshnessRule.title = `Latest ingestion age above ${PRODUCTION_TRAFFIC_FRESHNESS_BUDGET_SECONDS} seconds`
  freshnessRule.annotations ??= {}
  freshnessRule.annotations.summary = `Latest successful ingestion is older than ${PRODUCTION_TRAFFIC_FRESHNESS_BUDGET_SECONDS} seconds.`
  freshnessRule.annotations.description = `Production SLO alert for latest ingestion age above ${PRODUCTION_TRAFFIC_FRESHNESS_BUDGET_SECONDS} seconds.`

  const reconciliationRule = parsed.rules?.find((candidate) => candidate.uid === 'gfa-reconciliation-backlog')
  if (!reconciliationRule) fail('alert-rules.json is missing gfa-reconciliation-backlog')

  const reconciliationQuery = reconciliationRule.data?.find((item) => item.refId === 'A')
  if (typeof reconciliationQuery?.model?.expr !== 'string') {
    fail('gfa-reconciliation-backlog must expose a Prometheus query')
  }
  reconciliationQuery.model.expr = `(${reconciliationQuery.model.expr}) or vector(0)`

  return parsed
}

fs.mkdirSync(outputDirectory, { recursive: true })
for (const name of ['folder.json', 'dashboard.json', 'alert-rules.json']) {
  const sourcePath = path.join('monitoring', 'grafana-cloud', name)
  let value = fs.readFileSync(sourcePath, 'utf8')
  value = value.replaceAll('__PROMETHEUS_DATASOURCE_UID__', datasourceUID)
  let parsed
  try { parsed = JSON.parse(value) } catch (error) { fail(`${name} is invalid JSON: ${error.message}`) }
  parsed = applyProductionSLOPolicy(name, parsed)
  const rendered = JSON.stringify(parsed, null, 2) + '\n'
  fs.writeFileSync(path.join(outputDirectory, name), rendered)
}
console.log(`GRAFANA_OBSERVABILITY_RENDER=PASS ingestion_freshness_budget_seconds=${PRODUCTION_TRAFFIC_FRESHNESS_BUDGET_SECONDS} reconciliation_no_data=zero`)
