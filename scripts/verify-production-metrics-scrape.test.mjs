import assert from 'node:assert/strict'
import fs from 'node:fs'
import { execFileSync } from 'node:child_process'
import test from 'node:test'

const workflow = fs.readFileSync('.github/workflows/production-metrics-scrape.yml', 'utf8')
const alloy = fs.readFileSync('monitoring/grafana-cloud/config.alloy', 'utf8')

test('production metrics scraper contract passes', () => {
  const output = execFileSync(
    process.execPath,
    ['scripts/verify-production-metrics-scrape.mjs'],
    { encoding: 'utf8' },
  )
  assert.match(output, /PRODUCTION_METRICS_SCRAPER_CONTRACT=PASS/)
})

test('external scraper preserves deployment truth and protected access', () => {
  assert.match(workflow, /PRODUCTION_API_REVISION: \$\{\{ vars\.PRODUCTION_API_REVISION \}\}/)
  assert.match(workflow, /X-Internal-API-Key: \$PRODUCTION_METRICS_KEY/)
  assert.doesNotMatch(workflow, /\$\{\{ github\.sha \}\}/)
  assert.doesNotMatch(workflow, /git\s+rev-parse\s+HEAD/)
  assert.doesNotMatch(workflow, /grafana\/alloy:latest/)
})

test('Alloy scrapes the protected endpoint and remote writes with bounded labels', () => {
  assert.match(alloy, /prometheus\.scrape "gfa_production"/)
  assert.match(alloy, /metrics_path\s+= "\/internal\/metrics"/)
  assert.match(alloy, /X-Internal-API-Key/)
  assert.match(alloy, /prometheus\.remote_write "grafana_cloud"/)
  assert.match(alloy, /deployment_revision = sys\.env\("PRODUCTION_API_REVISION"\)/)
  assert.match(alloy, /sample_limit\s+= 5000/)
  assert.match(alloy, /label_limit\s+= 16/)
})
