import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import fs from 'node:fs'
import test from 'node:test'

import {
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
  assert.match(output, /OPENAPI_CONTRACT=PASS/)
})

test('contract exposes exactly eighteen stable GET operations', () => {
  const spec = loadSpec()
  assert.equal(Object.keys(spec.paths).length, 18)
  assert.deepEqual(
    Object.keys(spec.paths).sort(),
    [...requiredOperations.keys()].sort(),
  )
  for (const pathItem of Object.values(spec.paths)) {
    assert.ok(pathItem.get)
    assert.equal(pathItem.post, undefined)
    assert.equal(pathItem.patch, undefined)
    assert.equal(pathItem.delete, undefined)
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
