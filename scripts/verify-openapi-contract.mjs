#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

export const requiredOperations = new Map([
  ['/api/v1/health', 'getHealth'],
  ['/api/v1/ready', 'getReadiness'],
  ['/api/v1/version', 'getVersion'],
  ['/api/v1/regions', 'listRegions'],
  ['/api/v1/regions/{code}', 'getRegionByCode'],
  ['/api/v1/airports', 'listAirports'],
  ['/api/v1/airports/{icao}', 'getAirportByICAO'],
  ['/api/v1/traffic/current', 'getCurrentTraffic'],
])

const allowedOperationKeys = new Set(['get', 'parameters', 'summary', 'description'])
const forbiddenMethods = new Set(['post', 'put', 'patch', 'delete', 'trace', 'connect'])

function hasOwn(value, key) {
  return Object.prototype.hasOwnProperty.call(value, key)
}

function pointerResolve(document, reference) {
  if (!reference.startsWith('#/')) return undefined
  return reference
    .slice(2)
    .split('/')
    .map((segment) => segment.replaceAll('~1', '/').replaceAll('~0', '~'))
    .reduce((current, segment) => current?.[segment], document)
}

function collectReferences(value, output = []) {
  if (Array.isArray(value)) {
    for (const item of value) collectReferences(item, output)
    return output
  }
  if (!value || typeof value !== 'object') return output
  if (typeof value.$ref === 'string') output.push(value.$ref)
  for (const child of Object.values(value)) collectReferences(child, output)
  return output
}

function parameterMap(operation) {
  return new Map(
    (operation.parameters ?? []).map((parameter) => [
      `${parameter.in}:${parameter.name}`,
      parameter,
    ]),
  )
}

export function validateOpenAPIContract(spec) {
  const errors = []

  if (spec.openapi !== '3.1.0') {
    errors.push('openapi version must be 3.1.0')
  }
  if (spec.jsonSchemaDialect !== 'https://json-schema.org/draft/2020-12/schema') {
    errors.push('JSON Schema dialect must be draft 2020-12')
  }
  if (spec.info?.title !== 'Global Flight Analytics Public Read API') {
    errors.push('info.title is not the canonical public read API title')
  }

  if (!Array.isArray(spec.servers) || spec.servers.length !== 1 || spec.servers[0]?.url !== '/') {
    errors.push('servers must contain exactly one same-origin relative URL')
  }
  for (const server of spec.servers ?? []) {
    if (/^[a-z][a-z0-9+.-]*:\/\//i.test(server.url ?? '')) {
      errors.push(`absolute server URL is forbidden: ${server.url}`)
    }
  }

  const paths = spec.paths ?? {}
  const actualPaths = Object.keys(paths).sort()
  const expectedPaths = [...requiredOperations.keys()].sort()
  if (JSON.stringify(actualPaths) !== JSON.stringify(expectedPaths)) {
    errors.push(`path surface mismatch: expected ${expectedPaths.join(', ')}`)
  }

  const operationIds = new Set()
  for (const [route, expectedOperationID] of requiredOperations) {
    const pathItem = paths[route]
    if (!pathItem) continue

    for (const key of Object.keys(pathItem)) {
      if (forbiddenMethods.has(key)) {
        errors.push(`foundation route ${route} must remain read-only; found ${key}`)
      } else if (!allowedOperationKeys.has(key)) {
        errors.push(`unsupported path item key ${key} on ${route}`)
      }
    }

    const operation = pathItem.get
    if (!operation) {
      errors.push(`GET operation is missing for ${route}`)
      continue
    }
    if (operation.operationId !== expectedOperationID) {
      errors.push(`operationId mismatch for ${route}`)
    }
    if (operationIds.has(operation.operationId)) {
      errors.push(`duplicate operationId ${operation.operationId}`)
    }
    operationIds.add(operation.operationId)

    if (!operation.responses?.['200']) {
      errors.push(`200 response is missing for ${route}`)
    }

    const params = parameterMap(operation)
    for (const match of route.matchAll(/\{([^}]+)\}/g)) {
      const name = match[1]
      const parameter = params.get(`path:${name}`)
      if (!parameter?.required) {
        errors.push(`required path parameter ${name} is missing for ${route}`)
      }
    }
  }

  for (const ref of collectReferences(spec)) {
    if (!ref.startsWith('#/')) {
      errors.push(`external reference is forbidden: ${ref}`)
      continue
    }
    if (pointerResolve(spec, ref) === undefined) {
      errors.push(`unresolved reference: ${ref}`)
    }
  }

  const requiredSchemas = [
    'ErrorResponse',
    'HealthResponse',
    'ReadinessResponse',
    'VersionResponse',
    'RegionItem',
    'AirportListItem',
    'AirportProfile',
    'CurrentTrafficItem',
  ]
  for (const schema of requiredSchemas) {
    if (!spec.components?.schemas?.[schema]) {
      errors.push(`required schema is missing: ${schema}`)
    }
  }

  if (spec.components?.schemas?.ErrorResponse?.properties?.success?.const !== false) {
    errors.push('ErrorResponse.success must be const false')
  }

  for (const responseName of [
    'HealthResponse',
    'ReadinessResponse',
    'VersionResponse',
    'RegionResponse',
    'RegionsResponse',
    'AirportResponse',
    'AirportsResponse',
    'CurrentTrafficResponse',
  ]) {
    const schema = spec.components?.schemas?.[responseName]
    if (schema?.properties?.success?.const !== true || !schema?.properties?.data) {
      errors.push(`${responseName} must use the typed success envelope`)
    }
  }

  return errors
}

function read(root, relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8')
}

function requireIncludes(errors, content, literals, label) {
  for (const literal of literals) {
    if (!content.includes(literal)) {
      errors.push(`${label} is missing ${literal}`)
    }
  }
}

export function validateRepository(root) {
  const spec = JSON.parse(read(root, 'openapi/openapi.json'))
  const errors = validateOpenAPIContract(spec)

  const server = read(root, 'apps/api/internal/server/server.go')
  requireIncludes(errors, server, ['"/health"', '"/ready"', '"/version"'], 'server routes')

  const coreRoutes = read(root, 'apps/api/internal/server/core_database_routes.go')
  requireIncludes(
    errors,
    coreRoutes,
    ['"/regions"', '"/regions/:code"', '"/airports"', '"/airports/:icao"', '"/traffic/current"'],
    'core database routes',
  )

  const response = read(root, 'apps/api/internal/http/response/response.go')
  requireIncludes(
    errors,
    response,
    ['json:"success"', 'json:"data"', 'json:"error"', 'json:"code"', 'json:"message"'],
    'response envelope',
  )

  const systemDTO = read(root, 'apps/api/internal/http/dto/system.go')
  requireIncludes(errors, systemDTO, ['json:"status"', 'json:"version"', 'json:"revision"', 'json:"built_at"'], 'system DTO')

  const regionDTO = read(root, 'apps/api/internal/http/dto/region.go')
  requireIncludes(
    errors,
    regionDTO,
    ['json:"code"', 'json:"name"', 'json:"description"', 'json:"bounds"', 'json:"min_latitude"', 'json:"max_latitude"', 'json:"min_longitude"', 'json:"max_longitude"'],
    'region DTO',
  )

  const airportDTO = read(root, 'apps/api/internal/http/dto/airport.go')
  requireIncludes(
    errors,
    airportDTO,
    ['json:"icao_code"', 'json:"iata_code"', 'json:"elevation_m"', 'json:"elevation_status"', 'json:"timezone"'],
    'airport DTO',
  )

  const trafficDTO = read(root, 'apps/api/internal/http/dto/traffic.go')
  requireIncludes(
    errors,
    trafficDTO,
    ['json:"icao24"', 'json:"altitude_m"', 'json:"altitude_status"', 'json:"altitude_source"', 'json:"observed_at"'],
    'traffic DTO',
  )

  const trafficHandler = read(root, 'apps/api/internal/http/handlers/traffic.go')
  requireIncludes(errors, trafficHandler, ['c.Query("region")', '"REGION_NOT_FOUND"'], 'traffic handler')

  const readiness = read(root, 'apps/api/internal/http/handlers/readiness.go')
  requireIncludes(errors, readiness, ['Status: "ready"', '"SERVICE_NOT_READY"'], 'readiness handler')

  const packageJSON = JSON.parse(read(root, 'package.json'))
  const scripts = {
    'test:openapi-contract': 'node --test scripts/verify-openapi-contract.test.mjs',
    'verify:openapi-contract': 'node scripts/verify-openapi-contract.mjs',
  }
  for (const [name, command] of Object.entries(scripts)) {
    if (packageJSON.scripts?.[name] !== command) {
      errors.push(`package script is missing or incorrect: ${name}`)
    }
  }

  const release = read(root, 'scripts/verify-release.sh')
  requireIncludes(
    errors,
    release,
    ['pnpm run test:openapi-contract', 'pnpm run verify:openapi-contract', 'RELEASE_OPENAPI_CONTRACT=PASS'],
    'release gate',
  )

  const workflow = read(root, '.github/workflows/api-contract.yml')
  requireIncludes(
    errors,
    workflow,
    [
      'name: OpenAPI Contract',
      'node --test scripts/verify-openapi-contract.test.mjs',
      'node scripts/verify-openapi-contract.mjs',
      'openapi/openapi.json',
      'actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a',
    ],
    'OpenAPI workflow',
  )

  const document = read(root, 'docs/175_OPENAPI_CONTRACT_FOUNDATION.md')
  requireIncludes(
    errors,
    document,
    ['Status: IMPLEMENTED', 'eight stable GET operations', 'Playwright', 'OPENAPI_CONTRACT=PASS'],
    'Document 175',
  )

  const documentIndex = read(root, 'docs/DOCUMENT_INDEX.md')
  requireIncludes(
    errors,
    documentIndex,
    ['OPENAPI-CONTRACT-FOUNDATION-V1:DOCUMENT-INDEX', '175_OPENAPI_CONTRACT_FOUNDATION.md'],
    'documentation index',
  )

  return errors
}

function main() {
  const errors = validateRepository(process.cwd())
  if (errors.length > 0) {
    for (const error of errors) {
      console.error(`OPENAPI_CONTRACT=FAIL reason=${error}`)
    }
    process.exit(1)
  }
  console.log('OPENAPI_CONTRACT_PATHS=8')
  console.log('OPENAPI_CONTRACT_SCHEMAS=PASS')
  console.log('OPENAPI_CONTRACT_ROUTE_DRIFT=PASS')
  console.log('OPENAPI_CONTRACT=PASS')
}

const currentFile = fileURLToPath(import.meta.url)
if (process.argv[1] && path.resolve(process.argv[1]) === currentFile) {
  main()
}
