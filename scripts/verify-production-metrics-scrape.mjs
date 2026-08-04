#!/usr/bin/env node
import fs from 'node:fs'

function fail(message) {
  console.error(`PRODUCTION_METRICS_SCRAPER_CONTRACT=FAIL reason=${message}`)
  process.exit(1)
}

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

const workflow = read('.github/workflows/production-metrics-scrape.yml')
const alloy = read('monitoring/grafana-cloud/config.alloy')
const packageJson = JSON.parse(read('package.json'))
const release = read('scripts/verify-release.sh')
const backendCI = read('.github/workflows/backend-ci.yml')
const runbook = read('docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md')

for (const literal of [
  "cron: '7,22,37,52 * * * *'",
  'workflow_dispatch:',
  'contents: read',
  'group: production-metrics-scrape',
  'cancel-in-progress: false',
  'PRODUCTION_METRICS_KEY: ${{ secrets.PRODUCTION_METRICS_KEY }}',
  'GRAFANA_CLOUD_PROMETHEUS_URL: ${{ secrets.GRAFANA_CLOUD_PROMETHEUS_URL }}',
  'GRAFANA_CLOUD_PROMETHEUS_QUERY_URL: ${{ secrets.GRAFANA_CLOUD_PROMETHEUS_QUERY_URL }}',
  'GRAFANA_CLOUD_PROMETHEUS_USER: ${{ secrets.GRAFANA_CLOUD_PROMETHEUS_USER }}',
  'GRAFANA_CLOUD_API_KEY: ${{ secrets.GRAFANA_CLOUD_API_KEY }}',
  'PRODUCTION_API_REVISION: ${{ vars.PRODUCTION_API_REVISION }}',
  'https://$PRODUCTION_API_HOST/internal/metrics',
  'grafana/alloy:v1.18.0',
  'GRAFANA_CLOUD_REMOTE_WRITE=PASS',
  'GRAFANA_CLOUD_QUERY_EVIDENCE=PASS',
]) {
  if (!workflow.includes(literal)) fail(`workflow is missing ${literal}`)
}
if (workflow.includes('grafana/alloy:latest')) fail('workflow uses a mutable Alloy latest tag')
if (workflow.includes('${{ github.sha }}')) fail('workflow substitutes source SHA for deployed revision')
if (/git\s+rev-parse\s+HEAD/.test(workflow)) fail('workflow derives deployed revision from local HEAD')
if (!workflow.includes('X-Internal-API-Key: $PRODUCTION_METRICS_KEY')) fail('source scrape does not use the protected metrics key')
if (!workflow.includes('deployment_revision=\\"$PRODUCTION_API_REVISION\\"')) fail('Grafana query does not bind exact deployment revision')

for (const literal of [
  'prometheus.scrape "gfa_production"',
  'metrics_path    = "/internal/metrics"',
  'scrape_interval = "15s"',
  '"X-Internal-API-Key" = [sys.env("PRODUCTION_METRICS_KEY")]',
  'prometheus.remote_write "grafana_cloud"',
  'url            = sys.env("GRAFANA_CLOUD_PROMETHEUS_URL")',
  'username = sys.env("GRAFANA_CLOUD_PROMETHEUS_USER")',
  'password = sys.env("GRAFANA_CLOUD_API_KEY")',
  'deployment_revision = sys.env("PRODUCTION_API_REVISION")',
]) {
  if (!alloy.includes(literal)) fail(`Alloy config is missing ${literal}`)
}

if (packageJson.scripts?.['test:production-metrics-scrape'] !== 'node --test scripts/verify-production-metrics-scrape.test.mjs') {
  fail('package test entry point is missing')
}
if (packageJson.scripts?.['verify:production-metrics-scrape'] !== 'node scripts/verify-production-metrics-scrape.mjs') {
  fail('package verification entry point is missing')
}
for (const command of [
  'pnpm run test:production-metrics-scrape',
  'pnpm run verify:production-metrics-scrape',
]) {
  if (!release.includes(command)) fail(`release verification is missing ${command}`)
}
for (const literal of [
  ".github/workflows/production-metrics-scrape.yml",
  "monitoring/**",
  'Verify production metrics scraper contract',
  'Validate Grafana Alloy scrape configuration',
]) {
  if (!backendCI.includes(literal)) fail(`Backend CI is missing ${literal}`)
}
for (const literal of [
  'External Grafana Cloud metrics scraper',
  'PRODUCTION_METRICS_KEY',
  'GRAFANA_CLOUD_PROMETHEUS_URL',
  'GRAFANA_CLOUD_PROMETHEUS_QUERY_URL',
  'GRAFANA_CLOUD_QUERY_EVIDENCE=PASS',
]) {
  if (!runbook.includes(literal)) fail(`runbook is missing ${literal}`)
}

console.log('PRODUCTION_METRICS_SCRAPER_CONTRACT=PASS')
