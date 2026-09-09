#!/usr/bin/env node
import fs from 'node:fs'

const path = 'apps/api/internal/repository/postgres/traffic_position_provenance_integration_test.go'
const current = fs.readFileSync(path, 'utf8')
const pattern = /VALUES \(\$1, \$2, 'success', \$2\)\n/g
const matches = current.match(pattern) ?? []

if (matches.length !== 1) {
  throw new Error(`${path}: expected one ingestion_runs VALUES terminator candidate, found ${matches.length}`)
}

fs.writeFileSync(
  path,
  current.replace(pattern, "VALUES ($1, $2, 'success', $2);\n"),
)
