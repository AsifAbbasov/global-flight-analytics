import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import fs from 'node:fs'
import test from 'node:test'

import {
  requiredMutationOperations,
  requiredOperations,
  validateOpenAPIContract,
} from './verify-openapi-contract.mjs'

function loadSpec() {
  return JSON.parse(fs.readFileSync('openapi/openapi.json', 'utf8'))
}

test('repository OpenAPI contract passes', () => {
  const output = execFileSync(
    process.execPath,
    ['scripts/verify-openapi-contract.mjs'],
    { encoding: 'utf8' },
  )
  assert.match(output, /OPENAPI_CONTRACT_PATHS=39/)
  assert.match(output, /OPENAPI_PUBLIC_READ_OPERATIONS=38/)
  assert.match(output, /OPENAPI_PROTECTED_MUTATION_OPERATIONS=1/)
  assert.match(output, /OPENAPI_CONTRACT=PASS/)
})

test('contract exposes 38 reads and one protected mutation across 39 paths', () => {
  const spec = loadSpec()
  assert.equal(Object.keys(spec.paths).length, 39)
  assert.deepEqual(
    Object.keys(spec.paths).sort(),
    [
      ...requiredOperations.keys(),
      ...requiredMutationOperations.keys(),
    ].sort(),
  )
  for (const route of requiredOperations.keys()) {
    assert.ok(spec.paths[route].get)
    assert.equal(spec.paths[route].post, undefined)
  }
  for (const route of requiredMutationOperations.keys()) {
    assert.ok(spec.paths[route].post)
    assert.equal(spec.paths[route].get, undefined)
  }
})

test('active-aircraft query bounds remain source aligned', () => {
  const spec = loadSpec()
  const parameters = spec.paths['/api/v1/metrics/active-aircraft'].get.parameters
  const window = parameters.find(parameter => parameter.name === 'window_minutes')
  assert.deepEqual(window.schema, {
    type: 'integer',
    minimum: 1,
    maximum: 180,
    default: 15,
  })
})

test('flight-state altitude values remain explicitly nullable', () => {
  const spec = loadSpec()
  const schema = spec.components.schemas.FlightStateItem
  assert.deepEqual(schema.properties.barometric_altitude_m.type, [
    'number',
    'null',
  ])
  assert.deepEqual(schema.properties.geometric_altitude_m.type, [
    'number',
    'null',
  ])
})

test('advanced intelligence query contracts remain source aligned', () => {
  const spec = loadSpec()

  const weather = spec.paths['/api/v1/weather/current'].get.parameters
  assert.equal(weather.find(parameter => parameter.name === 'lat').required, true)
  assert.equal(weather.find(parameter => parameter.name === 'lon').required, true)

  const historical =
    spec.paths['/api/v1/historical-intelligence/aggregates/latest'].get
      .parameters
  assert.deepEqual(
    historical.find(parameter => parameter.name === 'scope').schema.enum,
    ['global', 'region', 'airport', 'route'],
  )

  const airspace =
    spec.paths['/api/v1/airspace/regions/{code}/analytics'].get.parameters
  assert.deepEqual(
    airspace.find(parameter => parameter.name === 'window_seconds').schema,
    {
      type: 'integer',
      minimum: 60,
      maximum: 3600,
      multipleOf: 60,
      default: 300,
    },
  )
})

test('server-owned analytical quality inputs are not published as client parameters', () => {
  const spec = loadSpec()
  for (const route of [
    '/api/v1/analytics/metrics/coverage-score',
    '/api/v1/analytics/metrics/data-freshness',
  ]) {
    const names = spec.paths[route].get.parameters.map(parameter => parameter.name)
    assert.deepEqual(names.sort(), ['region', 'window_minutes'])
    assert.equal(names.includes('limit'), false)
    assert.equal(names.includes('observed_samples'), false)
    assert.equal(names.includes('expected_samples'), false)
    assert.equal(names.includes('observed_at'), false)
    assert.equal(names.includes('max_age_seconds'), false)
  }
})

test('only intentionally open strategy evidence allows unknown properties', () => {
  const spec = loadSpec()
  assert.equal(spec.components.schemas.OpenObject.additionalProperties, true)
  assert.equal(
    spec.components.schemas.ProjectionIntelligence.additionalProperties,
    false,
  )
  assert.equal(
    spec.components.schemas.AirspaceRegionAnalytics.additionalProperties,
    false,
  )
})

test('Route Intelligence security and history bounds remain source aligned', () => {
  const spec = loadSpec()
  assert.deepEqual(spec.components.securitySchemes.InternalAPIKey, {
    type: 'apiKey',
    in: 'header',
    name: 'X-Internal-API-Key',
    description:
      'Internal mutation credential. The server stores only a SHA-256 digest; clients send the raw 32–256 character key.',
  })

  const mutation =
    spec.paths['/api/v1/trajectories/{id}/route-intelligence'].post
  assert.deepEqual(mutation.security, [{ InternalAPIKey: [] }])
  assert.equal(mutation.requestBody, undefined)
  assert.deepEqual(spec.components.headers.CacheControlNoStore, {
    description: 'Mutation responses are not cacheable.',
    schema: { type: 'string', const: 'no-store' },
  })
  assert.ok(mutation.responses['401'])
  assert.ok(mutation.responses['503'])

  const latest =
    spec.paths['/api/v1/trajectories/{id}/route-intelligence/latest'].get
  assert.equal(latest.security, undefined)

  const history =
    spec.paths['/api/v1/trajectories/{id}/route-intelligence/history'].get
  assert.deepEqual(
    history.parameters.find(parameter => parameter.name === 'limit').schema,
    { type: 'integer', minimum: 1, maximum: 100, default: 20 },
  )
  assert.deepEqual(
    history.parameters.find(
      parameter => parameter.name === 'before_as_of_time',
    ).schema,
    { type: 'string', format: 'date-time' },
  )

  const result = spec.components.schemas.RouteIntelligenceResult
  assert.equal(result.properties.icao24.pattern, '^[A-F0-9]{6}$')
  assert.equal(
    result.properties.identity_key.pattern,
    '^(?:|flight-identity-[0-9a-f]{64})$',
  )
  assert.deepEqual(
    spec.components.schemas.RouteIntelligenceAirport.properties.elevation_status.enum,
    ['observed', 'unknown', 'invalid'],
  )
  assert.deepEqual(
    spec.components.schemas.RouteIntelligenceHistory.properties.next_before_as_of_time,
    { type: 'string', format: 'date-time' },
  )
})

test('protected mutation security cannot be removed', () => {
  const spec = loadSpec()
  spec.paths['/api/v1/trajectories/{id}/route-intelligence'].post.security = []
  assert.ok(
    validateOpenAPIContract(spec).some(error =>
      error.includes('InternalAPIKey security requirement is missing'),
    ),
  )
})

test('absolute deployment origins are rejected', () => {
  const spec = loadSpec()
  spec.servers = [{ url: 'https://global-flight-analytics-api.onrender.com' }]
  assert.ok(
    validateOpenAPIContract(spec).some((error) =>
      error.includes('absolute server URL is forbidden'),
    ),
  )
})

test('unresolved local references are rejected', () => {
  const spec = loadSpec()
  spec.paths['/api/v1/health'].get.responses['200'].content[
    'application/json'
  ].schema.$ref = '#/components/schemas/MissingSchema'
  assert.ok(
    validateOpenAPIContract(spec).some((error) =>
      error.includes('unresolved reference'),
    ),
  )
})

test('mutation methods are rejected from the public read contract', () => {
  const spec = loadSpec()
  spec.paths['/api/v1/regions'].post = {
    operationId: 'createRegion',
    responses: { 204: { description: 'unexpected' } },
  }
  assert.ok(
    validateOpenAPIContract(spec).some((error) =>
      error.includes('must remain read-only'),
    ),
  )
})

test('path variables require matching required path parameters', () => {
  const spec = loadSpec()
  spec.paths['/api/v1/regions/{code}'].get.parameters = []
  assert.ok(
    validateOpenAPIContract(spec).some((error) =>
      error.includes('required path parameter code is missing'),
    ),
  )
})

test('operation identifiers must remain unique', () => {
  const spec = loadSpec()
  spec.paths['/api/v1/airports'].get.operationId = 'listRegions'
  assert.ok(
    validateOpenAPIContract(spec).some((error) =>
      error.includes('duplicate operationId'),
    ),
  )
})

test('live traffic query contract remains bounded', () => {
  const spec = loadSpec()
  const operation = spec.paths['/api/v1/traffic/live'].get
  assert.equal(operation.operationId, 'getLiveTraffic')
  const parameters = new Map(operation.parameters.map((parameter) => [parameter.name, parameter]))
  assert.deepEqual(parameters.get('limit').schema, {
    type: 'integer',
    minimum: 1,
    maximum: 5000,
    default: 1500,
  })
  assert.equal(parameters.get('selected').required, false)
  assert.deepEqual(spec.components.schemas.LiveTrafficItem.properties.altitude_m.type, ['number', 'null'])
  assert.deepEqual(spec.components.schemas.LiveTrafficItem.properties.on_ground.type, ['boolean', 'null'])
})
