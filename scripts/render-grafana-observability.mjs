#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'

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

fs.mkdirSync(outputDirectory, { recursive: true })
for (const name of ['folder.json', 'dashboard.json', 'alert-rules.json']) {
  const sourcePath = path.join('monitoring', 'grafana-cloud', name)
  let value = fs.readFileSync(sourcePath, 'utf8')
  value = value.replaceAll('__PROMETHEUS_DATASOURCE_UID__', datasourceUID)
  let parsed
  try { parsed = JSON.parse(value) } catch (error) { fail(`${name} is invalid JSON: ${error.message}`) }
  const rendered = JSON.stringify(parsed, null, 2) + '\n'
  fs.writeFileSync(path.join(outputDirectory, name), rendered)
}
console.log('GRAFANA_OBSERVABILITY_RENDER=PASS')
