import {
  APIRequestError,
  requestAPIData,
  type APIRequestOptions,
} from '@/lib/api/client'
import type {
  AirportActivityMetricParameters,
  AnalyticalConfidence,
  AnalyticalConfidenceFactor,
  AnalyticalConfidenceLevel,
  AnalyticalConfidenceReport,
  AnalyticalEligibility,
  AnalyticalFailure,
  AnalyticalMetric,
  AnalyticalNotice,
  AnalyticalScope,
  AnalyticalSource,
  AnalyticalStatus,
  CoverageScoreMetricParameters,
  DataFreshnessMetricParameters,
  RecentTrajectoryMetricParameters,
  TrafficDensityMetricParameters,
} from '@/types/analytics'

const analyticalMetricPath = '/api/v1/analytics/metrics'

const metricNames = {
  activeAircraft: 'traffic.active_aircraft',
  trafficDensity: 'traffic_density',
  airportActivity: 'traffic.airport_activity',
  coverageScore: 'coverage_score',
  dataFreshness: 'data_freshness',
} as const

export async function getAnalyticalActiveAircraft(
  parameters: RecentTrajectoryMetricParameters = {},
  options: APIRequestOptions = {}
): Promise<AnalyticalMetric<number>> {
  return requestAnalyticalMetric(
    `${analyticalMetricPath}/active-aircraft`,
    metricNames.activeAircraft,
    buildRecentTrajectorySearchParameters(parameters),
    options
  )
}

export async function getAnalyticalTrafficDensity(
  parameters: TrafficDensityMetricParameters,
  options: APIRequestOptions = {}
): Promise<AnalyticalMetric<number>> {
  const searchParameters = buildRecentTrajectorySearchParameters(parameters)
  const regionCode = parameters.regionCode?.trim()
  const areaSquareKilometers = parameters.areaSquareKilometers

  if (areaSquareKilometers !== undefined) {
    searchParameters.set(
      'area_square_kilometers',
      formatPositiveFiniteNumber(
        areaSquareKilometers,
        'areaSquareKilometers'
      )
    )
  } else if (!regionCode) {
    throw new APIRequestError(
      'Traffic density requires regionCode or areaSquareKilometers.'
    )
  }

  return requestAnalyticalMetric(
    `${analyticalMetricPath}/traffic-density`,
    metricNames.trafficDensity,
    searchParameters,
    options
  )
}

export async function getAnalyticalAirportActivity(
  parameters: AirportActivityMetricParameters,
  options: APIRequestOptions = {}
): Promise<AnalyticalMetric<number>> {
  const arrivalIDs = normalizeIdentifierList(
    parameters.arrivalTrajectoryIDs ?? []
  )
  const departureIDs = normalizeIdentifierList(
    parameters.departureTrajectoryIDs ?? []
  )

  if (arrivalIDs.length === 0 && departureIDs.length === 0) {
    throw new APIRequestError(
      'At least one arrival or departure trajectory identifier is required.'
    )
  }

  const searchParameters = new URLSearchParams()

  if (arrivalIDs.length > 0) {
    searchParameters.set('arrival_trajectory_ids', arrivalIDs.join(','))
  }

  if (departureIDs.length > 0) {
    searchParameters.set('departure_trajectory_ids', departureIDs.join(','))
  }

  return requestAnalyticalMetric(
    `${analyticalMetricPath}/airport-activity`,
    metricNames.airportActivity,
    searchParameters,
    options
  )
}

export async function getAnalyticalCoverageScore(
  parameters: CoverageScoreMetricParameters,
  options: APIRequestOptions = {}
): Promise<AnalyticalMetric<number>> {
  const searchParameters = new URLSearchParams({
    observed_samples: formatInteger(
      parameters.observedSamples,
      'observedSamples'
    ),
    expected_samples: formatInteger(
      parameters.expectedSamples,
      'expectedSamples'
    ),
  })

  return requestAnalyticalMetric(
    `${analyticalMetricPath}/coverage-score`,
    metricNames.coverageScore,
    searchParameters,
    options
  )
}

export async function getAnalyticalDataFreshness(
  parameters: DataFreshnessMetricParameters,
  options: APIRequestOptions = {}
): Promise<AnalyticalMetric<number>> {
  const observedAt = normalizeTimestamp(parameters.observedAt)

  if (
    !Number.isInteger(parameters.maximumAgeSeconds) ||
    parameters.maximumAgeSeconds < 1 ||
    parameters.maximumAgeSeconds > 86_400
  ) {
    throw new APIRequestError(
      'maximumAgeSeconds must be an integer between 1 and 86400.'
    )
  }

  const searchParameters = new URLSearchParams({
    observed_at: observedAt,
    max_age_seconds: String(parameters.maximumAgeSeconds),
  })

  return requestAnalyticalMetric(
    `${analyticalMetricPath}/data-freshness`,
    metricNames.dataFreshness,
    searchParameters,
    options
  )
}

async function requestAnalyticalMetric(
  path: string,
  expectedMetric: string,
  searchParameters: URLSearchParams,
  options: APIRequestOptions
): Promise<AnalyticalMetric<number>> {
  const payload = await requestAPIData<unknown>(path, {
    ...options,
    searchParams: searchParameters,
  })

  return parseAnalyticalMetric(payload, expectedMetric)
}

function buildRecentTrajectorySearchParameters(
  parameters: RecentTrajectoryMetricParameters
): URLSearchParams {
  const searchParameters = new URLSearchParams()

  if (parameters.windowMinutes !== undefined) {
    if (
      !Number.isInteger(parameters.windowMinutes) ||
      parameters.windowMinutes < 1 ||
      parameters.windowMinutes > 180
    ) {
      throw new APIRequestError(
        'windowMinutes must be an integer between 1 and 180.'
      )
    }

    searchParameters.set('window_minutes', String(parameters.windowMinutes))
  }

  if (parameters.limit !== undefined) {
    if (
      !Number.isInteger(parameters.limit) ||
      parameters.limit < 1 ||
      parameters.limit > 5_000
    ) {
      throw new APIRequestError(
        'limit must be an integer between 1 and 5000.'
      )
    }

    searchParameters.set('limit', String(parameters.limit))
  }

  const regionCode = parameters.regionCode?.trim()
  if (regionCode) {
    searchParameters.set('region', regionCode)
  }

  return searchParameters
}

function parseAnalyticalMetric(
  value: unknown,
  expectedMetric: string
): AnalyticalMetric<number> {
  const record = requireRecord(value, 'analytical metric')

  const metric = requireString(record.metric, 'metric')
  if (metric !== expectedMetric) {
    throw invalidPayload(
      `Expected metric ${expectedMetric}, received ${metric}.`
    )
  }

  const status = parseStatus(record.status)
  const hasValue = requireBoolean(record.has_value, 'has_value')
  const parsedValue =
    record.value === undefined ? undefined : requireFiniteNumber(record.value, 'value')

  if (hasValue && parsedValue === undefined) {
    throw invalidPayload('A usable analytical result must include value.')
  }

  if (!hasValue && parsedValue !== undefined) {
    throw invalidPayload('A non-usable analytical result must omit value.')
  }

  if (
    (status === 'complete' || status === 'limited') !== hasValue
  ) {
    throw invalidPayload('Analytical status and has_value are inconsistent.')
  }

  const result: AnalyticalMetric<number> = {
    metric,
    status,
    has_value: hasValue,
    confidence: parseConfidence(record.confidence),
    scope: parseScope(record.scope),
    sources: parseArray(record.sources, 'sources', parseSource),
    warnings: parseArray(record.warnings, 'warnings', parseNotice),
    limitations: parseArray(
      record.limitations,
      'limitations',
      parseNotice
    ),
    calculated_at: requireTimestamp(record.calculated_at, 'calculated_at'),
  }

  if (parsedValue !== undefined) {
    result.value = parsedValue
  }

  if (record.eligibility !== undefined) {
    result.eligibility = parseEligibility(record.eligibility)
  }

  if (record.failure !== undefined) {
    result.failure = parseFailure(record.failure)
  }

  if (record.confidence_report !== undefined) {
    result.confidence_report = parseConfidenceReport(
      record.confidence_report
    )
  }

  if (status === 'failed' && result.failure === undefined) {
    throw invalidPayload('A failed analytical result must include failure.')
  }

  if (status !== 'failed' && result.failure !== undefined) {
    throw invalidPayload(
      'Only a failed analytical result may include failure.'
    )
  }

  return result
}

function parseStatus(value: unknown): AnalyticalStatus {
  const status = requireString(value, 'status')

  if (
    status === 'complete' ||
    status === 'limited' ||
    status === 'denied' ||
    status === 'failed'
  ) {
    return status
  }

  throw invalidPayload(`Unknown analytical status: ${status}.`)
}

function parseConfidence(value: unknown): AnalyticalConfidence {
  const record = requireRecord(value, 'confidence')

  return {
    level: parseConfidenceLevel(record.level),
    score: requireUnitNumber(record.score, 'confidence.score'),
    reasons: parseArray(
      record.reasons,
      'confidence.reasons',
      parseNotice
    ),
  }
}

function parseConfidenceLevel(value: unknown): AnalyticalConfidenceLevel {
  const level = requireString(value, 'confidence level')

  if (
    level === 'none' ||
    level === 'low' ||
    level === 'medium' ||
    level === 'high'
  ) {
    return level
  }

  throw invalidPayload(`Unknown confidence level: ${level}.`)
}

function parseEligibility(value: unknown): AnalyticalEligibility {
  const record = requireRecord(value, 'eligibility')

  return {
    capability: requireString(record.capability, 'eligibility.capability'),
    allowed: requireBoolean(record.allowed, 'eligibility.allowed'),
    reasons: parseArray(
      record.reasons,
      'eligibility.reasons',
      item => requireString(item, 'eligibility reason')
    ),
    evaluated_at: requireTimestamp(
      record.evaluated_at,
      'eligibility.evaluated_at'
    ),
  }
}

function parseScope(value: unknown): AnalyticalScope {
  const record = requireRecord(value, 'scope')

  return {
    capability: requireString(record.capability, 'scope.capability'),
    input_count: requireNonNegativeInteger(
      record.input_count,
      'scope.input_count'
    ),
    allowed_count: requireNonNegativeInteger(
      record.allowed_count,
      'scope.allowed_count'
    ),
    denied_count: requireNonNegativeInteger(
      record.denied_count,
      'scope.denied_count'
    ),
    reasons: parseArray(record.reasons, 'scope.reasons', item => {
      const reason = requireRecord(item, 'scope reason')

      return {
        reason: requireString(reason.reason, 'scope reason code'),
        count: requireNonNegativeInteger(reason.count, 'scope reason count'),
      }
    }),
    evaluated_at: requireTimestamp(
      record.evaluated_at,
      'scope.evaluated_at'
    ),
  }
}

function parseSource(value: unknown): AnalyticalSource {
  const record = requireRecord(value, 'source')
  const source: AnalyticalSource = {
    name: requireString(record.name, 'source.name'),
    role: requireString(record.role, 'source.role'),
    limitations: parseArray(
      record.limitations,
      'source.limitations',
      parseNotice
    ),
  }

  assignOptionalTimestamp(source, 'observed_from', record.observed_from)
  assignOptionalTimestamp(source, 'observed_to', record.observed_to)
  assignOptionalTimestamp(source, 'retrieved_at', record.retrieved_at)

  return source
}

function parseNotice(value: unknown): AnalyticalNotice {
  const record = requireRecord(value, 'notice')

  return {
    code: requireString(record.code, 'notice.code'),
    message: requireString(record.message, 'notice.message'),
  }
}

function parseFailure(value: unknown): AnalyticalFailure {
  const record = requireRecord(value, 'failure')

  return {
    code: requireString(record.code, 'failure.code'),
    message: requireString(record.message, 'failure.message'),
    retriable: requireBoolean(record.retriable, 'failure.retriable'),
  }
}

function parseConfidenceReport(
  value: unknown
): AnalyticalConfidenceReport {
  const record = requireRecord(value, 'confidence_report')

  return {
    base_score: requireUnitNumber(
      record.base_score,
      'confidence_report.base_score'
    ),
    penalty_score: requireUnitNumber(
      record.penalty_score,
      'confidence_report.penalty_score'
    ),
    score: requireUnitNumber(record.score, 'confidence_report.score'),
    level: parseConfidenceLevel(record.level),
    factors: parseArray(
      record.factors,
      'confidence_report.factors',
      parseConfidenceFactor
    ),
    reasons: parseArray(
      record.reasons,
      'confidence_report.reasons',
      parseNotice
    ),
    warnings: parseArray(
      record.warnings,
      'confidence_report.warnings',
      parseNotice
    ),
    limitations: parseArray(
      record.limitations,
      'confidence_report.limitations',
      parseNotice
    ),
    evaluated_at: requireTimestamp(
      record.evaluated_at,
      'confidence_report.evaluated_at'
    ),
  }
}

function parseConfidenceFactor(
  value: unknown
): AnalyticalConfidenceFactor {
  const record = requireRecord(value, 'confidence factor')
  const kind = requireString(record.kind, 'confidence factor kind')

  if (kind !== 'evidence' && kind !== 'penalty') {
    throw invalidPayload(`Unknown confidence factor kind: ${kind}.`)
  }

  return {
    code: requireString(record.code, 'confidence factor code'),
    kind,
    weight: requireFiniteNumber(record.weight, 'confidence factor weight'),
    value: requireUnitNumber(record.value, 'confidence factor value'),
    impact: requireFiniteNumber(record.impact, 'confidence factor impact'),
    message: requireString(record.message, 'confidence factor message'),
  }
}

function parseArray<T>(
  value: unknown,
  fieldName: string,
  parser: (item: unknown) => T
): T[] {
  if (!Array.isArray(value)) {
    throw invalidPayload(`${fieldName} must be an array.`)
  }

  return value.map(parser)
}

function requireRecord(
  value: unknown,
  fieldName: string
): Record<string, unknown> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    throw invalidPayload(`${fieldName} must be an object.`)
  }

  return value as Record<string, unknown>
}

function requireString(value: unknown, fieldName: string): string {
  if (typeof value !== 'string' || value.trim() === '') {
    throw invalidPayload(`${fieldName} must be a non-empty string.`)
  }

  return value
}

function requireBoolean(value: unknown, fieldName: string): boolean {
  if (typeof value !== 'boolean') {
    throw invalidPayload(`${fieldName} must be a boolean.`)
  }

  return value
}

function requireFiniteNumber(value: unknown, fieldName: string): number {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    throw invalidPayload(`${fieldName} must be a finite number.`)
  }

  return value
}

function requireUnitNumber(value: unknown, fieldName: string): number {
  const parsed = requireFiniteNumber(value, fieldName)

  if (parsed < 0 || parsed > 1) {
    throw invalidPayload(`${fieldName} must be between zero and one.`)
  }

  return parsed
}

function requireNonNegativeInteger(
  value: unknown,
  fieldName: string
): number {
  if (
    typeof value !== 'number' ||
    !Number.isInteger(value) ||
    value < 0
  ) {
    throw invalidPayload(`${fieldName} must be a non-negative integer.`)
  }

  return value
}

function requireTimestamp(value: unknown, fieldName: string): string {
  const timestamp = requireString(value, fieldName)

  if (Number.isNaN(Date.parse(timestamp))) {
    throw invalidPayload(`${fieldName} must be a valid timestamp.`)
  }

  return timestamp
}

function assignOptionalTimestamp(
  target: AnalyticalSource,
  field:
    | 'observed_from'
    | 'observed_to'
    | 'retrieved_at',
  value: unknown
): void {
  if (value === undefined || value === null || value === '') {
    return
  }

  target[field] = requireTimestamp(value, `source.${field}`)
}

function formatInteger(value: number, fieldName: string): string {
  if (!Number.isInteger(value)) {
    throw new APIRequestError(`${fieldName} must be an integer.`)
  }

  return String(value)
}

function formatPositiveFiniteNumber(
  value: number,
  fieldName: string
): string {
  if (!Number.isFinite(value) || value <= 0) {
    throw new APIRequestError(
      `${fieldName} must be a positive finite number.`
    )
  }

  return String(value)
}

function normalizeIdentifierList(values: string[]): string[] {
  const result: string[] = []
  const seen = new Set<string>()

  for (const value of values) {
    const normalized = value.trim()

    if (normalized === '') {
      continue
    }

    if (!seen.has(normalized)) {
      seen.add(normalized)
      result.push(normalized)
    }
  }

  return result
}

function normalizeTimestamp(value: string): string {
  const timestamp = value.trim()

  if (timestamp === '' || Number.isNaN(Date.parse(timestamp))) {
    throw new APIRequestError('observedAt must be a valid timestamp.')
  }

  return new Date(timestamp).toISOString()
}

function invalidPayload(message: string): APIRequestError {
  return new APIRequestError(`The analytical API response is invalid: ${message}`)
}
