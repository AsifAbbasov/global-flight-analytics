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
  ['/api/v1/traffic/live', 'getLiveTraffic'],
  ['/api/v1/metrics/active-aircraft', 'getActiveAircraftMetric'],
  ['/api/v1/aircraft/{icao24}/trajectory', 'getLatestTrajectoryByICAO24'],
  ['/api/v1/aircraft/{icao24}/route-context', 'getAircraftRouteContextByICAO24'],
  ['/api/v1/trajectories/{id}', 'getTrajectoryByID'],
  ['/api/v1/flights/{flightID}/states', 'listFlightStatesByFlightID'],
  ['/api/v1/aircraft/{icao24}/latest-state', 'getLatestFlightStateByICAO24'],
  ['/api/v1/flights', 'listFlights'],
  ['/api/v1/flights/{id}', 'getFlightByID'],
  ['/api/v1/aircraft', 'listAircraft'],
  ['/api/v1/aircraft/{icao24}', 'getAircraftByICAO24'],
  ['/api/v1/aircraft/{icao24}/transponder-evidence/latest', 'getLatestTransponderEvidence'],
  ['/api/v1/weather/current', 'getCurrentWeather'],
  ['/api/v1/analytics/metrics/active-aircraft', 'getAnalyticalActiveAircraft'],
  ['/api/v1/analytics/metrics/traffic-density', 'getAnalyticalTrafficDensity'],
  ['/api/v1/analytics/metrics/airport-activity', 'getAnalyticalAirportActivity'],
  ['/api/v1/analytics/metrics/coverage-score', 'getAnalyticalCoverageScore'],
  ['/api/v1/analytics/metrics/data-freshness', 'getAnalyticalDataFreshness'],
  ['/api/v1/airports/intelligence/ranking', 'getAirportIntelligenceRanking'],
  ['/api/v1/airports/{icao}/intelligence/overview', 'getAirportIntelligenceOverview'],
  ['/api/v1/airports/{icao}/intelligence/history', 'getAirportIntelligenceHistory'],
  ['/api/v1/airports/{icao}/intelligence/trends', 'getAirportIntelligenceTrends'],
  ['/api/v1/historical-intelligence/aggregates/latest', 'getLatestHistoricalIntelligenceAggregate'],
  ['/api/v1/historical-intelligence/aggregates/history', 'listHistoricalIntelligenceAggregateHistory'],
  ['/api/v1/trajectories/{id}/projection-intelligence', 'getProjectionIntelligenceByTrajectoryID'],
  ['/api/v1/trajectories/{id}/stability-intelligence', 'getStabilityIntelligenceByTrajectoryID'],
  ['/api/v1/trajectories/{id}/weather-context', 'getWeatherContextByTrajectoryID'],
  ['/api/v1/airspace/regions/{code}/analytics', 'getAirspaceRegionAnalytics'],
  ['/api/v1/trajectories/{id}/route-intelligence/latest', 'getLatestRouteIntelligenceByTrajectoryID'],
  ['/api/v1/trajectories/{id}/route-intelligence/history', 'listRouteIntelligenceHistoryByTrajectoryID'],
])

export const requiredMutationOperations = new Map([
  ['/api/v1/trajectories/{id}/route-intelligence', 'processRouteIntelligenceByTrajectoryID'],
])

const allowedReadPathItemKeys = new Set(['get', 'parameters', 'summary', 'description'])
const allowedMutationPathItemKeys = new Set(['post', 'parameters', 'summary', 'description'])
const forbiddenReadMethods = new Set(['post', 'put', 'patch', 'delete', 'trace', 'connect'])

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
  if (spec.info?.title !== 'Global Flight Analytics Public API') {
    errors.push('info.title is not the canonical public API title')
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
  const expectedPaths = [
    ...requiredOperations.keys(),
    ...requiredMutationOperations.keys(),
  ].sort()
  if (JSON.stringify(actualPaths) !== JSON.stringify(expectedPaths)) {
    errors.push(`path surface mismatch: expected ${expectedPaths.join(', ')}`)
  }

  const operationIds = new Set()
  for (const [route, expectedOperationID] of requiredOperations) {
    const pathItem = paths[route]
    if (!pathItem) continue

    for (const key of Object.keys(pathItem)) {
      if (forbiddenReadMethods.has(key)) {
        errors.push(`public read route ${route} must remain read-only; found ${key}`)
      } else if (!allowedReadPathItemKeys.has(key)) {
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

  for (const [route, expectedOperationID] of requiredMutationOperations) {
    const pathItem = paths[route]
    if (!pathItem) continue

    for (const key of Object.keys(pathItem)) {
      if (!allowedMutationPathItemKeys.has(key)) {
        errors.push(`unsupported mutation path item key ${key} on ${route}`)
      }
    }

    const operation = pathItem.post
    if (!operation) {
      errors.push(`POST operation is missing for ${route}`)
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
    if (!operation.responses?.['401']) {
      errors.push(`401 response is missing for protected mutation ${route}`)
    }
    if (!operation.responses?.['503']) {
      errors.push(`503 response is missing for protected mutation ${route}`)
    }
    if (JSON.stringify(operation.security) !== JSON.stringify([{ InternalAPIKey: [] }])) {
      errors.push(`InternalAPIKey security requirement is missing for ${route}`)
    }
    if (operation.requestBody !== undefined) {
      errors.push(`Route Intelligence mutation must not publish a request body`)
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

  const internalAPIKey = spec.components?.securitySchemes?.InternalAPIKey
  if (
    internalAPIKey?.type !== 'apiKey' ||
    internalAPIKey?.in !== 'header' ||
    internalAPIKey?.name !== 'X-Internal-API-Key'
  ) {
    errors.push('InternalAPIKey must be an X-Internal-API-Key header apiKey scheme')
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
    'LiveTrafficItem',
    'LiveTrafficSnapshot',
    'LiveTrafficResponse',
    'ActiveAircraftMetric',
    'AircraftListItem',
    'AircraftProfile',
    'FlightListItem',
    'FlightProfile',
    'FlightStateItem',
    'Trajectory',
    'AircraftRouteContext',
    'ActiveAircraftMetricResponse',
    'AircraftListResponse',
    'AircraftResponse',
    'FlightListResponse',
    'FlightResponse',
    'FlightStateListResponse',
    'FlightStateResponse',
    'TrajectoryResponse',
    'AircraftRouteContextResponse',
    'TransponderEvidence',
    'CurrentWeather',
    'AnalyticalMetric',
    'AirportIntelligenceRanking',
    'AirportIntelligenceOverview',
    'AirportIntelligenceHistory',
    'AirportIntelligenceTrends',
    'HistoricalAggregateRecord',
    'HistoricalAggregateHistory',
    'ProjectionIntelligence',
    'StabilityIntelligence',
    'WeatherContext',
    'AirspaceRegionAnalytics',
    'OpenObject',
    'TransponderEvidenceResponse',
    'CurrentWeatherResponse',
    'AnalyticalMetricResponse',
    'AirportIntelligenceRankingResponse',
    'AirportIntelligenceOverviewResponse',
    'AirportIntelligenceHistoryResponse',
    'AirportIntelligenceTrendsResponse',
    'HistoricalAggregateRecordResponse',
    'HistoricalAggregateHistoryResponse',
    'ProjectionIntelligenceResponse',
    'StabilityIntelligenceResponse',
    'WeatherContextResponse',
    'AirspaceRegionAnalyticsResponse',
    'RouteIntelligenceEvidenceAttribute',
    'RouteIntelligenceEvidence',
    'RouteIntelligenceConfidenceReason',
    'RouteIntelligenceConfidence',
    'RouteIntelligenceLimitation',
    'RouteIntelligenceAirport',
    'RouteIntelligenceEndpoint',
    'RouteIntelligenceWindow',
    'RouteIntelligenceSummary',
    'RouteIntelligenceProvenance',
    'RouteIntelligenceResult',
    'RouteIntelligenceRecord',
    'RouteIntelligenceHistory',
    'RouteIntelligenceRecordResponse',
    'RouteIntelligenceHistoryResponse',
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
    'LiveTrafficResponse',
    'ActiveAircraftMetricResponse',
    'AircraftListResponse',
    'AircraftResponse',
    'FlightListResponse',
    'FlightResponse',
    'FlightStateListResponse',
    'FlightStateResponse',
    'TrajectoryResponse',
    'AircraftRouteContextResponse',
    'TransponderEvidenceResponse',
    'CurrentWeatherResponse',
    'AnalyticalMetricResponse',
    'AirportIntelligenceRankingResponse',
    'AirportIntelligenceOverviewResponse',
    'AirportIntelligenceHistoryResponse',
    'AirportIntelligenceTrendsResponse',
    'HistoricalAggregateRecordResponse',
    'HistoricalAggregateHistoryResponse',
    'ProjectionIntelligenceResponse',
    'StabilityIntelligenceResponse',
    'WeatherContextResponse',
    'AirspaceRegionAnalyticsResponse',
    'RouteIntelligenceRecordResponse',
    'RouteIntelligenceHistoryResponse',
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
    [
      '"/regions"',
      '"/regions/:code"',
      '"/metrics/active-aircraft"',
      '"/traffic/current"',
      '"/aircraft/:icao24/trajectory"',
      '"/aircraft/:icao24/route-context"',
      '"/trajectories/:id"',
      '"/flights/:flightID/states"',
      '"/aircraft/:icao24/latest-state"',
      '"/flights"',
      '"/flights/:id"',
      '"/aircraft"',
      '"/aircraft/:icao24"',
      '"/airports"',
      '"/airports/:icao"',
    ],
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

  const metricsDTO = read(root, 'apps/api/internal/http/dto/metrics.go')
  requireIncludes(errors, metricsDTO, ['json:"metric"', 'json:"window_minutes"', 'json:"confidence"', 'json:"sources"', 'json:"limitations"'], 'metrics DTO')

  const aircraftDTO = read(root, 'apps/api/internal/http/dto/aircraft.go')
  requireIncludes(errors, aircraftDTO, ['json:"icao24"', 'json:"registration"', 'json:"aircraft_type"'], 'aircraft DTO')

  const flightDTO = read(root, 'apps/api/internal/http/dto/flight.go')
  requireIncludes(errors, flightDTO, ['json:"aircraft_id"', 'json:"first_seen_at"', 'json:"last_seen_at"'], 'flight DTO')

  const flightStateDTO = read(root, 'apps/api/internal/http/dto/flightstate.go')
  requireIncludes(errors, flightStateDTO, ['json:"barometric_altitude_m"', 'json:"barometric_altitude_status"', 'json:"geometric_altitude_m"', 'json:"geometric_altitude_status"'], 'flight-state DTO')

  const trajectoryDTO = read(root, 'apps/api/internal/http/dto/trajectory.go')
  requireIncludes(errors, trajectoryDTO, ['json:"identity_key"', 'json:"segments"', 'json:"coverage_gaps"', 'json:"quality_score"'], 'trajectory DTO')

  const routeContextDTO = read(root, 'apps/api/internal/http/dto/route_context.go')
  requireIncludes(errors, routeContextDTO, ['json:"trajectory_id"', 'json:"origin,omitempty"', 'json:"destination,omitempty"', 'json:"limitations"'], 'route-context DTO')

  const metricsHandler = read(root, 'apps/api/internal/http/handlers/metrics.go')
  requireIncludes(errors, metricsHandler, ['c.Query("window_minutes")', 'c.Query("region")', '"INVALID_WINDOW_MINUTES"'], 'metrics handler')

  const trajectoryHandler = read(root, 'apps/api/internal/http/handlers/trajectories.go')
  requireIncludes(errors, trajectoryHandler, ['"INVALID_ICAO24"', '"INVALID_TRAJECTORY_ID"', '"TRAJECTORY_SERVICE_UNAVAILABLE"'], 'trajectory handler')

  const routeContextHandler = read(root, 'apps/api/internal/http/handlers/route_context.go')
  requireIncludes(errors, routeContextHandler, ['"INVALID_ICAO24"', '"ROUTE_CONTEXT_NOT_FOUND"', '"ROUTE_CONTEXT_SERVICE_UNAVAILABLE"'], 'route-context handler')

  const transponderDTO = read(root, 'apps/api/internal/http/dto/transponder_evidence.go')
  requireIncludes(errors, transponderDTO, ['json:"evidence_only"', 'json:"confirmed_emergency"', 'json:"maximum_claim_strength"'], 'transponder-evidence DTO')

  const weatherDTO = read(root, 'apps/api/internal/http/dto/weather.go')
  requireIncludes(errors, weatherDTO, ['json:"temperature_celsius"', 'json:"rain_mm"', 'json:"wind_gusts_mps"'], 'current-weather DTO')

  const analyticalDTO = read(root, 'apps/api/internal/http/dto/analytical_metric.go')
  requireIncludes(errors, analyticalDTO, ['json:"has_value"', 'json:"data_quality,omitempty"', 'json:"confidence_report,omitempty"'], 'analytical-metric DTO')

  const airportIntelligenceHandler = read(root, 'apps/api/internal/http/handlers/airport_intelligence.go')
  requireIncludes(errors, airportIntelligenceHandler, ['ctx.Query(airportIntelligenceDaysQuery)', 'ctx.Query(airportIntelligenceAsOfTimeQuery)', 'maximumAirportIntelligenceRankingLimit = 200'], 'Airport Intelligence handler')

  const historicalHandler = read(root, 'apps/api/internal/http/handlers/historical_intelligence.go')
  requireIncludes(errors, historicalHandler, ['historicalMetricQuery', 'historicalScopeQuery', 'historicalGranularityQuery', 'historicalCursorQuery'], 'Historical Intelligence handler')

  const projectionHandler = read(root, 'apps/api/internal/http/handlers/projection_intelligence.go')
  requireIncludes(errors, projectionHandler, ['projectionIntelligenceAsOfTimeQuery', 'projectionIntelligenceDurationSecondsQuery', '"INVALID_PROJECTION_TRAJECTORY_ID"'], 'Projection Intelligence handler')

  const stabilityHandler = read(root, 'apps/api/internal/http/handlers/stability_intelligence.go')
  requireIncludes(errors, stabilityHandler, ['stabilityIntelligenceAsOfTimesQuery', 'MinimumAsOfTimeCount', '"INVALID_STABILITY_DURATION"'], 'Stability Intelligence handler')

  const weatherContextHandler = read(root, 'apps/api/internal/http/handlers/weather_context.go')
  requireIncludes(errors, weatherContextHandler, ['weatherContextAsOfTimeQuery', 'weatherContextDurationSecondsQuery', '"WEATHER_CONTEXT_CONTRACT_INVALID"'], 'Weather Context handler')

  const airspaceHandler = read(root, 'apps/api/internal/http/handlers/airspace_region_analytics.go')
  requireIncludes(errors, airspaceHandler, ['minimumAirspaceAnalyticsWindowSeconds = 60', 'maximumAirspaceAnalyticsWindowSeconds = 3600', 'windowSeconds%60 != 0'], 'Airspace Intelligence handler')

  const analyticalHandler = read(root, 'apps/api/internal/http/handlers/analytical_metrics.go')
  requireIncludes(errors, analyticalHandler, ['ctx.Query("window_minutes")', 'ctx.Query("airport_icao")', '"AREA_PARAMETER_NOT_SUPPORTED"'], 'advanced analytical handler')

  const routeIntelligenceRoutes = read(root, 'apps/api/internal/server/route_intelligence_database_routes.go')
  requireIncludes(errors, routeIntelligenceRoutes, ['"/trajectories/:id/route-intelligence"', 'mutationAuthorization', '"/trajectories/:id/route-intelligence/latest"', '"/trajectories/:id/route-intelligence/history"'], 'Route Intelligence routes')

  const routeIntelligenceDTO = read(root, 'apps/api/internal/http/dto/route_intelligence.go')
  requireIncludes(errors, routeIntelligenceDTO, ['json:"schema_version"', 'json:"origin,omitempty"', 'json:"destination,omitempty"', 'json:"next_before_as_of_time,omitempty"'], 'Route Intelligence DTO')

  const routeIntelligenceHandler = read(root, 'apps/api/internal/http/handlers/route_intelligence.go')
  requireIncludes(errors, routeIntelligenceHandler, ['routeIntelligenceHistoryLimitQuery', 'routeIntelligenceHistoryBeforeQuery', 'routestore.DefaultListLimit', 'routestore.MaximumListLimit', '"INVALID_ROUTE_INTELLIGENCE_CURSOR"'], 'Route Intelligence handler')

  const mutationAuthorization = read(root, 'apps/api/internal/middleware/mutation_authorization.go')
  requireIncludes(errors, mutationAuthorization, ['"MUTATION_AUTHENTICATION_REQUIRED"', '"MUTATION_AUTHENTICATION_UNAVAILABLE"', 'fiber.StatusUnauthorized', 'fiber.StatusServiceUnavailable'], 'mutation authorization middleware')

  const internalAPIKey = read(root, 'apps/api/internal/security/internalapikey/key.go')
  requireIncludes(errors, internalAPIKey, ['HeaderName = "X-Internal-API-Key"', 'MinimumCandidateLength = 32', 'MaximumCandidateLength = 256', 'subtle.ConstantTimeCompare'], 'internal API key contract')

  const routeIdentityValidation = read(root, 'apps/api/internal/routeintelligence/routecontract/route_validation_identifiers.go')
  requireIncludes(errors, routeIdentityValidation, ['^[A-F0-9]{6}$', '^flight-identity-[0-9a-f]{64}$', '^sha256:[0-9a-f]{64}$'], 'Route Intelligence identifier validation')

  const routeAssessmentValidation = read(root, 'apps/api/internal/routeintelligence/routecontract/route_validation_assessment.go')
  requireIncludes(errors, routeAssessmentValidation, ['confidence.EvidenceCount', 'expectedEvidenceCount', 'confidence_evidence_count_mismatch'], 'Route Intelligence confidence validation')

  const airportElevation = read(root, 'apps/api/internal/domain/airport/elevation.go')
  requireIncludes(errors, airportElevation, ['ElevationStatusObserved', 'ElevationStatusUnknown', 'ElevationStatusInvalid'], 'Route Intelligence airport elevation vocabulary')

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
    ['Status: IMPLEMENTED', 'thirty-eight production public operations', 'InternalAPIKey', 'OPENAPI_CONTRACT=PASS'],
    'Document 175',
  )

  const closureDocument = read(root, 'docs/180_OPENAPI_ROUTE_INTELLIGENCE_CONTRACT_CLOSURE.md')
  requireIncludes(
    errors,
    closureDocument,
    ['Status: IMPLEMENTED', 'InternalAPIKey', 'OPENAPI_MISSING_OPERATIONS=0', 'b67005b7ff6a944cffc2f4d846aec1009ea69e53'],
    'Document 180',
  )

  const documentIndex = read(root, 'docs/DOCUMENT_INDEX.md')
  requireIncludes(
    errors,
    documentIndex,
    ['OPENAPI-CONTRACT-FOUNDATION-V1:DOCUMENT-INDEX', '175_OPENAPI_CONTRACT_FOUNDATION.md', 'OPENAPI-ROUTE-INTELLIGENCE-CONTRACT-CLOSURE-V1:DOCUMENT-INDEX', '180_OPENAPI_ROUTE_INTELLIGENCE_CONTRACT_CLOSURE.md'],
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
  console.log(`OPENAPI_CONTRACT_PATHS=${requiredOperations.size + requiredMutationOperations.size}`)
  console.log(`OPENAPI_PUBLIC_READ_OPERATIONS=${requiredOperations.size}`)
  console.log(`OPENAPI_PROTECTED_MUTATION_OPERATIONS=${requiredMutationOperations.size}`)
  console.log('OPENAPI_ROUTE_INTELLIGENCE_OPERATIONS=3')
  console.log('OPENAPI_CONTRACT_SCHEMAS=PASS')
  console.log('OPENAPI_CONTRACT_SECURITY=PASS')
  console.log('OPENAPI_CONTRACT_ROUTE_DRIFT=PASS')
  console.log('OPENAPI_CONTRACT=PASS')
}

const currentFile = fileURLToPath(import.meta.url)
if (process.argv[1] && path.resolve(process.argv[1]) === currentFile) {
  main()
}
