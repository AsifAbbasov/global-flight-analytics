import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import test from 'node:test'

import {
  compareOperationSets,
  discoverSourceOperations,
  expectedInternalOperations,
  expectedPublicOperations,
  extractOpenAPIOperations,
  extractRouteOperationsFromGo,
  validateAuthorizationBoundaries,
} from './verify-openapi-route-inventory.mjs'

function loadSpec() {
  return JSON.parse(fs.readFileSync('openapi/openapi.json', 'utf8'))
}

test('repository route inventory passes', () => {
  const output = execFileSync(
    process.execPath,
    ['scripts/verify-openapi-route-inventory.mjs'],
    { encoding: 'utf8' },
  )
  assert.match(output, /SOURCE_PUBLIC_OPERATIONS=38/)
  assert.match(output, /OPENAPI_MISSING_OPERATIONS=30/)
  assert.match(output, /OPENAPI_ROUTE_INVENTORY=PASS/)
})

test('source inventory exposes exactly 37 GET operations and one POST operation', () => {
  const { publicOperations } = discoverSourceOperations(process.cwd())
  assert.equal(publicOperations.length, 38)
  assert.equal(publicOperations.filter(([method]) => method === 'GET').length, 37)
  assert.equal(publicOperations.filter(([method]) => method === 'POST').length, 1)
  assert.deepEqual(
    publicOperations.map(([method, routePath]) => `${method} ${routePath}`).sort(),
    expectedPublicOperations.map(([method, routePath]) => `${method} ${routePath}`).sort(),
  )
})

test('internal metrics remains the only internal HTTP operation', () => {
  const { internalOperations } = discoverSourceOperations(process.cwd())
  assert.deepEqual(internalOperations, expectedInternalOperations)
})

test('foundation OpenAPI is missing 30 source operations and invents none', () => {
  const { publicOperations } = discoverSourceOperations(process.cwd())
  const openAPIOperations = extractOpenAPIOperations(loadSpec())
  const gap = compareOperationSets(publicOperations, openAPIOperations)
  assert.equal(openAPIOperations.length, 8)
  assert.equal(gap.missing.length, 30)
  assert.deepEqual(gap.extra, [])
})

test('nested Fiber route groups are resolved into one canonical public path', () => {
  const operations = extractRouteOperationsFromGo({
    source: `
      package server
      func register(v1 Router) {
        routes := v1.Group("/analytics").Group("/metrics")
        routes.Get("/coverage-score", handler)
      }
    `,
    seedPrefixes: new Map([['v1', '/api/v1']]),
  })
  assert.deepEqual(operations, [['GET', '/api/v1/analytics/metrics/coverage-score']])
})

test('constant-backed Fiber paths are resolved without formatting dependence', () => {
  const operations = extractRouteOperationsFromGo({
    source: `
      package server
      const WeatherContextPath = "/trajectories/:id/weather-context"
      func register(v1 Router) {
        v1.Get(
          WeatherContextPath,
          handler,
        )
      }
    `,
    constants: new Map([['WeatherContextPath', '/trajectories/:id/weather-context']]),
    seedPrefixes: new Map([['v1', '/api/v1']]),
  })
  assert.deepEqual(operations, [['GET', '/api/v1/trajectories/{id}/weather-context']])
})

test('authorization boundary audit rejects unprotected mutation and metrics routes', () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'gfa-openapi-route-auth-'))
  const serverDirectory = path.join(root, 'apps/api/internal/server')
  fs.mkdirSync(serverDirectory, { recursive: true })
  fs.writeFileSync(
    path.join(serverDirectory, 'route_intelligence_database_routes.go'),
    'package server\nfunc register(v1 Router) { v1.Post("/trajectories/:id/route-intelligence", handler) }\n',
  )
  fs.writeFileSync(
    path.join(serverDirectory, 'server.go'),
    'package server\nfunc register(app Router) { app.Get(observability.MetricsPath, handler); v1 := app.Group("/api").Group("/v1"); _ = v1 }\n',
  )
  const errors = validateAuthorizationBoundaries(root)
  assert.ok(errors.some((error) => error.includes('mutationAuthorization')))
  assert.ok(errors.some((error) => error.includes('metricsAuthorization')))
})
