#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const HTTP_METHODS = new Set(['GET', 'POST', 'PUT', 'PATCH', 'DELETE'])

export const expectedPublicOperations = Object.freeze([
  ['GET', '/api/v1/health'],
  ['GET', '/api/v1/ready'],
  ['GET', '/api/v1/version'],
  ['GET', '/api/v1/regions'],
  ['GET', '/api/v1/regions/{code}'],
  ['GET', '/api/v1/metrics/active-aircraft'],
  ['GET', '/api/v1/traffic/current'],
  ['GET', '/api/v1/aircraft/{icao24}/trajectory'],
  ['GET', '/api/v1/aircraft/{icao24}/route-context'],
  ['GET', '/api/v1/trajectories/{id}'],
  ['GET', '/api/v1/flights/{flightID}/states'],
  ['GET', '/api/v1/aircraft/{icao24}/latest-state'],
  ['GET', '/api/v1/flights'],
  ['GET', '/api/v1/flights/{id}'],
  ['GET', '/api/v1/aircraft'],
  ['GET', '/api/v1/aircraft/{icao24}'],
  ['GET', '/api/v1/airports'],
  ['GET', '/api/v1/airports/{icao}'],
  ['GET', '/api/v1/aircraft/{icao24}/transponder-evidence/latest'],
  ['POST', '/api/v1/trajectories/{id}/route-intelligence'],
  ['GET', '/api/v1/trajectories/{id}/route-intelligence/latest'],
  ['GET', '/api/v1/trajectories/{id}/route-intelligence/history'],
  ['GET', '/api/v1/weather/current'],
  ['GET', '/api/v1/analytics/metrics/active-aircraft'],
  ['GET', '/api/v1/analytics/metrics/traffic-density'],
  ['GET', '/api/v1/analytics/metrics/airport-activity'],
  ['GET', '/api/v1/analytics/metrics/coverage-score'],
  ['GET', '/api/v1/analytics/metrics/data-freshness'],
  ['GET', '/api/v1/airports/intelligence/ranking'],
  ['GET', '/api/v1/airports/{icao}/intelligence/overview'],
  ['GET', '/api/v1/airports/{icao}/intelligence/history'],
  ['GET', '/api/v1/airports/{icao}/intelligence/trends'],
  ['GET', '/api/v1/historical-intelligence/aggregates/latest'],
  ['GET', '/api/v1/historical-intelligence/aggregates/history'],
  ['GET', '/api/v1/trajectories/{id}/projection-intelligence'],
  ['GET', '/api/v1/trajectories/{id}/stability-intelligence'],
  ['GET', '/api/v1/trajectories/{id}/weather-context'],
  ['GET', '/api/v1/airspace/regions/{code}/analytics'],
])

export const expectedInternalOperations = Object.freeze([
  ['GET', '/internal/metrics'],
])

const serverDirectory = 'apps/api/internal/server'
const observabilityRegistry = 'apps/api/internal/observability/registry.go'
const openAPIPath = 'openapi/openapi.json'

function operationKey(method, routePath) {
  return `${method.toUpperCase()} ${routePath}`
}

function normalizeRoutePath(routePath) {
  const withParameters = routePath.replace(/:([A-Za-z_][A-Za-z0-9_]*)/g, '{$1}')
  const normalized = withParameters.replace(/\/+/g, '/')
  return normalized.length > 1 ? normalized.replace(/\/$/, '') : normalized
}

function stripGoComments(source) {
  return source
    .replace(/\/\*[\s\S]*?\*\//g, '')
    .replace(/(^|\s)\/\/.*$/gm, '$1')
}

function parseStringConstants(source) {
  const constants = new Map()
  const cleaned = stripGoComments(source)
  const declarationPattern = /\b([A-Za-z_][A-Za-z0-9_]*)\s*(?:[A-Za-z_][A-Za-z0-9_.]*)?\s*=\s*"([^"\\]*(?:\\.[^"\\]*)*)"/g

  for (const match of cleaned.matchAll(/\bconst\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:[A-Za-z_][A-Za-z0-9_.]*)?\s*=\s*"([^"\\]*(?:\\.[^"\\]*)*)"/g)) {
    constants.set(match[1], JSON.parse(`"${match[2]}"`))
  }
  for (const block of cleaned.matchAll(/\bconst\s*\(([\s\S]*?)\)/g)) {
    for (const match of block[1].matchAll(declarationPattern)) {
      constants.set(match[1], JSON.parse(`"${match[2]}"`))
    }
  }
  return constants
}

function parseGroupPrefixes(source, seedPrefixes) {
  const prefixes = new Map(seedPrefixes)
  const cleaned = stripGoComments(source)
  const assignmentPattern = /\b([A-Za-z_][A-Za-z0-9_]*)\s*:=\s*([A-Za-z_][A-Za-z0-9_]*)((?:\s*\.Group\s*\(\s*"[^"\\]*(?:\\.[^"\\]*)*"\s*,?\s*\))+)/g

  let changed = true
  while (changed) {
    changed = false
    for (const match of cleaned.matchAll(assignmentPattern)) {
      const [, target, parent, chain] = match
      if (!prefixes.has(parent) || prefixes.has(target)) continue
      const segments = [...chain.matchAll(/\.Group\s*\(\s*"([^"\\]*(?:\\.[^"\\]*)*)"/g)]
        .map((segment) => JSON.parse(`"${segment[1]}"`))
      prefixes.set(target, normalizeRoutePath([prefixes.get(parent), ...segments].join('/')))
      changed = true
    }
  }
  return prefixes
}

function resolveRouteArgument(argument, constants) {
  const trimmed = argument.trim()
  const stringMatch = trimmed.match(/^"([^"\\]*(?:\\.[^"\\]*)*)"$/)
  if (stringMatch) return JSON.parse(`"${stringMatch[1]}"`)

  const identifierMatch = trimmed.match(/^(?:[A-Za-z_][A-Za-z0-9_]*\.)*([A-Za-z_][A-Za-z0-9_]*)$/)
  if (!identifierMatch) return undefined
  return constants.get(identifierMatch[1])
}

export function extractRouteOperationsFromGo({ source, constants = new Map(), seedPrefixes = new Map() }) {
  const prefixes = parseGroupPrefixes(source, seedPrefixes)
  const operations = []
  const cleaned = stripGoComments(source)
  const callPattern = /\b([A-Za-z_][A-Za-z0-9_]*)\.(Get|Post|Put|Patch|Delete)\s*\(\s*((?:"[^"\\]*(?:\\.[^"\\]*)*")|(?:(?:[A-Za-z_][A-Za-z0-9_]*\.)*[A-Za-z_][A-Za-z0-9_]*))\s*,/g

  for (const match of cleaned.matchAll(callPattern)) {
    const [, receiver, methodName, argument] = match
    if (!prefixes.has(receiver)) continue
    const relativePath = resolveRouteArgument(argument, constants)
    if (relativePath === undefined) continue
    const fullPath = normalizeRoutePath(`${prefixes.get(receiver)}/${relativePath}`)
    operations.push([methodName.toUpperCase(), fullPath])
  }

  return operations
}

function listProductionGoFiles(root, relativeDirectory) {
  const absoluteDirectory = path.join(root, relativeDirectory)
  return fs
    .readdirSync(absoluteDirectory, { withFileTypes: true })
    .filter((entry) => entry.isFile() && entry.name.endsWith('.go') && !entry.name.endsWith('_test.go'))
    .map((entry) => path.join(relativeDirectory, entry.name))
    .sort()
}

function read(root, relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8')
}

export function discoverSourceOperations(root) {
  const sourceFiles = listProductionGoFiles(root, serverDirectory)
  const registrySource = read(root, observabilityRegistry)
  const constants = parseStringConstants(registrySource)

  for (const relativePath of sourceFiles) {
    for (const [name, value] of parseStringConstants(read(root, relativePath))) {
      if (constants.has(name) && constants.get(name) !== value) {
        throw new Error(`conflicting route constant ${name}`)
      }
      constants.set(name, value)
    }
  }

  const operations = []
  for (const relativePath of sourceFiles) {
    const source = read(root, relativePath)
    operations.push(
      ...extractRouteOperationsFromGo({
        source,
        constants,
        seedPrefixes: new Map([
          ['app', ''],
          ['v1', '/api/v1'],
        ]),
      }),
    )
  }

  const publicOperations = operations.filter(([, routePath]) => routePath.startsWith('/api/v1/'))
  const internalOperations = operations.filter(([, routePath]) => routePath.startsWith('/internal/'))
  return { publicOperations, internalOperations }
}

export function extractOpenAPIOperations(spec) {
  const operations = []
  for (const [routePath, pathItem] of Object.entries(spec.paths ?? {})) {
    for (const [method, operation] of Object.entries(pathItem ?? {})) {
      const upperMethod = method.toUpperCase()
      if (!HTTP_METHODS.has(upperMethod) || !operation || typeof operation !== 'object') continue
      operations.push([upperMethod, normalizeRoutePath(routePath)])
    }
  }
  return operations
}

export function compareOperationSets(sourceOperations, documentedOperations) {
  const sourceKeys = new Set(sourceOperations.map(([method, routePath]) => operationKey(method, routePath)))
  const documentedKeys = new Set(documentedOperations.map(([method, routePath]) => operationKey(method, routePath)))
  return {
    missing: [...sourceKeys].filter((key) => !documentedKeys.has(key)).sort(),
    extra: [...documentedKeys].filter((key) => !sourceKeys.has(key)).sort(),
  }
}

function compareExactInventory(actual, expected, label) {
  const actualKeys = actual.map(([method, routePath]) => operationKey(method, routePath)).sort()
  const expectedKeys = expected.map(([method, routePath]) => operationKey(method, routePath)).sort()
  const errors = []

  const duplicates = actualKeys.filter((key, index) => actualKeys.indexOf(key) !== index)
  for (const duplicate of [...new Set(duplicates)]) {
    errors.push(`${label} contains duplicate source registration: ${duplicate}`)
  }

  const actualSet = new Set(actualKeys)
  const expectedSet = new Set(expectedKeys)
  for (const key of expectedKeys) {
    if (!actualSet.has(key)) errors.push(`${label} is missing source operation: ${key}`)
  }
  for (const key of actualKeys) {
    if (!expectedSet.has(key)) errors.push(`${label} contains unclassified source operation: ${key}`)
  }
  return errors
}

export function validateAuthorizationBoundaries(root) {
  const errors = []
  const routeIntelligence = stripGoComments(read(root, 'apps/api/internal/server/route_intelligence_database_routes.go'))
  if (!/v1\.Post\s*\(\s*"\/trajectories\/:id\/route-intelligence"\s*,\s*mutationAuthorization\s*,/.test(routeIntelligence)) {
    errors.push('route-intelligence mutation must be registered with mutationAuthorization before the handler')
  }

  const server = stripGoComments(read(root, 'apps/api/internal/server/server.go'))
  if (!/app\.Get\s*\(\s*observability\.MetricsPath\s*,\s*metricsAuthorization\s*,/.test(server)) {
    errors.push('internal metrics route must be registered with metricsAuthorization before the handler')
  }
  if (!/app\.Group\s*\(\s*"\/api"\s*,?\s*\)\s*\.Group\s*\(\s*"\/v1"\s*,?\s*\)/.test(server)) {
    errors.push('public API must preserve the nested /api and /v1 Fiber groups')
  }
  return errors
}

export function validateRepository(root) {
  const errors = []
  const { publicOperations, internalOperations } = discoverSourceOperations(root)
  errors.push(...compareExactInventory(publicOperations, expectedPublicOperations, 'public route inventory'))
  errors.push(...compareExactInventory(internalOperations, expectedInternalOperations, 'internal route inventory'))
  errors.push(...validateAuthorizationBoundaries(root))

  const spec = JSON.parse(read(root, openAPIPath))
  const openAPIOperations = extractOpenAPIOperations(spec)
  const gap = compareOperationSets(publicOperations, openAPIOperations)

  if (publicOperations.length !== 38) errors.push(`public operation count must be 38; found ${publicOperations.length}`)
  if (publicOperations.filter(([method]) => method === 'GET').length !== 37) errors.push('public GET operation count must be 37')
  if (publicOperations.filter(([method]) => method === 'POST').length !== 1) errors.push('public POST operation count must be 1')
  if (internalOperations.length !== 1) errors.push(`internal operation count must be 1; found ${internalOperations.length}`)
  if (openAPIOperations.length !== 35) errors.push(`current OpenAPI operation count must be 35; found ${openAPIOperations.length}`)
  if (gap.missing.length !== 3) errors.push(`OpenAPI missing operation count must be 3; found ${gap.missing.length}`)
  const expectedRemaining = [
    'GET /api/v1/trajectories/{id}/route-intelligence/history',
    'GET /api/v1/trajectories/{id}/route-intelligence/latest',
    'POST /api/v1/trajectories/{id}/route-intelligence',
  ]
  if (JSON.stringify(gap.missing) !== JSON.stringify(expectedRemaining)) errors.push(`remaining OpenAPI gap is not the exact Route Intelligence slice: ${gap.missing.join(', ')}`)
  if (gap.extra.length !== 0) errors.push(`OpenAPI contains operations absent from the source: ${gap.extra.join(', ')}`)
  if (openAPIOperations.some(([, routePath]) => routePath.startsWith('/internal/'))) {
    errors.push('public OpenAPI must not expose /internal operations')
  }

  return { errors, publicOperations, internalOperations, openAPIOperations, gap }
}

function main() {
  const result = validateRepository(process.cwd())
  if (result.errors.length > 0) {
    for (const error of result.errors) console.error(`OPENAPI_ROUTE_INVENTORY=FAIL reason=${error}`)
    process.exit(1)
  }

  console.log(`SOURCE_PUBLIC_OPERATIONS=${result.publicOperations.length}`)
  console.log(`SOURCE_PUBLIC_GET_OPERATIONS=${result.publicOperations.filter(([method]) => method === 'GET').length}`)
  console.log(`SOURCE_PUBLIC_POST_OPERATIONS=${result.publicOperations.filter(([method]) => method === 'POST').length}`)
  console.log(`SOURCE_INTERNAL_OPERATIONS=${result.internalOperations.length}`)
  console.log(`OPENAPI_DOCUMENTED_OPERATIONS=${result.openAPIOperations.length}`)
  console.log(`OPENAPI_MISSING_OPERATIONS=${result.gap.missing.length}`)
  console.log(`OPENAPI_EXTRA_OPERATIONS=${result.gap.extra.length}`)
  console.log('OPENAPI_MUTATION_AUTHORIZATION=PASS')
  console.log('OPENAPI_INTERNAL_METRICS_BOUNDARY=PASS')
  console.log('OPENAPI_ROUTE_INVENTORY=PASS')
}

const currentFile = fileURLToPath(import.meta.url)
if (process.argv[1] && path.resolve(process.argv[1]) === currentFile) main()
