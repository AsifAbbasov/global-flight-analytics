// FRONTEND_HISTORICAL_ANALYTICS_COMPARISON_V1
import { APIRequestError, requestAPIData, type APIRequestOptions } from '@/lib/api/client'
import type {
  HistoricalBucketStatus,
  HistoricalConfidenceLevel,
  HistoricalGranularity,
  HistoricalIntelligenceAggregateHistory,
  HistoricalIntelligenceAggregateRecord,
  HistoricalIntelligenceConfidence,
  HistoricalIntelligenceLimitation,
  HistoricalIntelligencePoint,
  HistoricalIntelligenceResult,
  HistoricalIntelligenceSelection,
  HistoricalMetricName,
  HistoricalScopeType,
  HistoricalSeriesStatus,
  HistoricalTrendDirection,
} from '@/types/historical-intelligence'

export interface HistoricalIntelligenceHistoryOptions extends APIRequestOptions {
  limit?: number
  cursor?: string
}

const metricNames = new Set<HistoricalMetricName>([
  'active_aircraft', 'flight_count', 'trajectory_count', 'observation_count',
  'traffic_density', 'airport_departures', 'airport_arrivals',
  'airport_operations', 'unique_aircraft', 'active_routes',
  'route_observations', 'route_confidence', 'complete_route_ratio',
  'partial_route_ratio', 'unavailable_route_ratio', 'great_circle_distance_km',
])
const scopeTypes = new Set<HistoricalScopeType>(['global', 'airport', 'route'])
const granularities = new Set<HistoricalGranularity>(['hour', 'day', 'week', 'custom'])
const seriesStatuses = new Set<HistoricalSeriesStatus>(['unavailable', 'partial', 'complete'])
const bucketStatuses = new Set<HistoricalBucketStatus>(['unavailable', 'partial', 'complete'])
const confidenceLevels = new Set<HistoricalConfidenceLevel>(['none', 'low', 'medium', 'high'])
const trendDirections = new Set<HistoricalTrendDirection>(['unavailable', 'down', 'flat', 'up'])
const aggregations = new Set(['count', 'sum', 'minimum', 'maximum', 'average', 'median', 'ratio'])

export async function getLatestHistoricalIntelligence(
  selection: HistoricalIntelligenceSelection,
  options: APIRequestOptions = {}
): Promise<HistoricalIntelligenceAggregateRecord> {
  return parseAggregateRecord(
    await requestAPIData<unknown>('/api/v1/historical-intelligence/aggregates/latest', {
      ...options,
      searchParams: buildSearchParams(selection),
    })
  )
}

export async function getHistoricalIntelligenceHistory(
  selection: HistoricalIntelligenceSelection,
  options: HistoricalIntelligenceHistoryOptions = {}
): Promise<HistoricalIntelligenceAggregateHistory> {
  const searchParams = buildSearchParams(selection)
  const limit = options.limit ?? 20
  if (!Number.isInteger(limit) || limit < 1 || limit > 100) {
    throw new APIRequestError('Historical Intelligence history limit must be between one and one hundred.')
  }
  searchParams.set('limit', String(limit))
  if (options.cursor?.trim()) searchParams.set('cursor', options.cursor.trim())
  return parseHistory(
    await requestAPIData<unknown>('/api/v1/historical-intelligence/aggregates/history', {
      signal: options.signal,
      timeoutMilliseconds: options.timeoutMilliseconds,
      searchParams,
    })
  )
}

function buildSearchParams(selection: HistoricalIntelligenceSelection): URLSearchParams {
  const searchParams = new URLSearchParams({
    metric: selection.metric,
    scope: selection.scope,
    granularity: selection.granularity,
  })
  if (selection.scope === 'airport') {
    const value = normalizeICAO(selection.airportICAO)
    searchParams.set('airport_icao', value)
  }
  if (selection.scope === 'route') {
    searchParams.set('origin_icao', normalizeICAO(selection.originICAO))
    searchParams.set('destination_icao', normalizeICAO(selection.destinationICAO))
  }
  return searchParams
}

function parseHistory(value: unknown): HistoricalIntelligenceAggregateHistory {
  const item = record(value, 'history')
  return {
    items: array(item.items, 'history.items').map((entry, index) =>
      parseAggregateRecord(entry, `history.items[${index}]`)
    ),
    has_more: booleanValue(item.has_more, 'history.has_more'),
    ...(item.next_cursor === undefined || item.next_cursor === null || item.next_cursor === ''
      ? {}
      : { next_cursor: stringValue(item.next_cursor, 'history.next_cursor') }),
  }
}

function parseAggregateRecord(
  value: unknown,
  field = 'aggregate'
): HistoricalIntelligenceAggregateRecord {
  const item = record(value, field)
  return {
    id: stringValue(item.id, `${field}.id`),
    input_fingerprint: fingerprint(item.input_fingerprint, `${field}.input_fingerprint`),
    stored_at: timestamp(item.stored_at, `${field}.stored_at`),
    result: parseResult(item.result, `${field}.result`),
  }
}

function parseResult(value: unknown, field: string): HistoricalIntelligenceResult {
  const item = record(value, field)
  const status = enumValue(item.status, `${field}.status`, seriesStatuses)
  const granularity = enumValue(item.granularity, `${field}.granularity`, granularities)
  const metric = record(item.metric, `${field}.metric`)
  const metricName = enumValue(metric.name, `${field}.metric.name`, metricNames)
  const aggregation = stringValue(metric.aggregation, `${field}.metric.aggregation`)
  if (!aggregations.has(aggregation)) invalid(`${field}.metric.aggregation is unsupported.`)
  const scope = parseScope(item.scope, `${field}.scope`)
  const points = array(item.points, `${field}.points`).map((point, index) =>
    parsePoint(point, `${field}.points[${index}]`)
  )
  const summary = record(item.summary, `${field}.summary`)
  const comparison = item.comparison == null
    ? undefined
    : parseComparison(item.comparison, `${field}.comparison`)
  const provenance = record(item.provenance, `${field}.provenance`)
  return {
    schema_version: stringValue(item.schema_version, `${field}.schema_version`),
    status,
    metric: {
      name: metricName,
      unit: stringValue(metric.unit, `${field}.metric.unit`),
      aggregation: aggregation as HistoricalIntelligenceResult['metric']['aggregation'],
    },
    scope,
    window: parseWindow(item.window, `${field}.window`),
    granularity,
    points,
    summary: {
      point_count: nonNegativeInteger(summary.point_count, `${field}.summary.point_count`),
      total: finiteNumber(summary.total, `${field}.summary.total`),
      minimum: finiteNumber(summary.minimum, `${field}.summary.minimum`),
      maximum: finiteNumber(summary.maximum, `${field}.summary.maximum`),
      average: finiteNumber(summary.average, `${field}.summary.average`),
      median: finiteNumber(summary.median, `${field}.summary.median`),
    },
    ...(comparison ? { comparison } : {}),
    confidence: parseConfidence(item.confidence, `${field}.confidence`),
    limitations: parseLimitations(item.limitations, `${field}.limitations`),
    provenance: {
      builder_version: stringValue(provenance.builder_version, `${field}.provenance.builder_version`),
      input_fingerprint: fingerprint(provenance.input_fingerprint, `${field}.provenance.input_fingerprint`),
      source_names: array(provenance.source_names, `${field}.provenance.source_names`).map((source, index) =>
        stringValue(source, `${field}.provenance.source_names[${index}]`)
      ),
      latest_source_updated_at: timestamp(provenance.latest_source_updated_at, `${field}.provenance.latest_source_updated_at`),
    },
    generated_at: timestamp(item.generated_at, `${field}.generated_at`),
  }
}

function parseScope(value: unknown, field: string): HistoricalIntelligenceResult['scope'] {
  const item = record(value, field)
  const type = enumValue(item.type, `${field}.type`, scopeTypes)
  if (type === 'global') return { type }
  if (type === 'airport') {
    return { type, airport_icao_code: normalizeICAO(item.airport_icao_code) }
  }
  return {
    type,
    origin_icao_code: normalizeICAO(item.origin_icao_code),
    destination_icao_code: normalizeICAO(item.destination_icao_code),
  }
}

function parsePoint(value: unknown, field: string): HistoricalIntelligencePoint {
  const item = record(value, field)
  return {
    start_time: timestamp(item.start_time, `${field}.start_time`),
    end_time: timestamp(item.end_time, `${field}.end_time`),
    status: enumValue(item.status, `${field}.status`, bucketStatuses),
    value: nonNegativeNumber(item.value, `${field}.value`),
    sample_count: nonNegativeInteger(item.sample_count, `${field}.sample_count`),
    coverage_ratio: ratio(item.coverage_ratio, `${field}.coverage_ratio`),
    confidence: parseConfidence(item.confidence, `${field}.confidence`),
    limitations: parseLimitations(item.limitations, `${field}.limitations`),
  }
}

function parseComparison(value: unknown, field: string) {
  const item = record(value, field)
  return {
    previous_window: parseWindow(item.previous_window, `${field}.previous_window`),
    current_value: finiteNumber(item.current_value, `${field}.current_value`),
    previous_value: finiteNumber(item.previous_value, `${field}.previous_value`),
    absolute_change: finiteNumber(item.absolute_change, `${field}.absolute_change`),
    ...(item.percentage_change == null
      ? {}
      : { percentage_change: finiteNumber(item.percentage_change, `${field}.percentage_change`) }),
    direction: enumValue(item.direction, `${field}.direction`, trendDirections),
  }
}

function parseConfidence(value: unknown, field: string): HistoricalIntelligenceConfidence {
  const item = record(value, field)
  return {
    score: ratio(item.score, `${field}.score`),
    level: enumValue(item.level, `${field}.level`, confidenceLevels),
    sample_count: nonNegativeInteger(item.sample_count, `${field}.sample_count`),
    reasons: array(item.reasons, `${field}.reasons`).map((reason, index) => {
      const entry = record(reason, `${field}.reasons[${index}]`)
      return {
        code: stringValue(entry.code, `${field}.reasons[${index}].code`),
        message: stringValue(entry.message, `${field}.reasons[${index}].message`),
        contribution: finiteNumber(entry.contribution, `${field}.reasons[${index}].contribution`),
      }
    }),
  }
}

function parseLimitations(value: unknown, field: string): HistoricalIntelligenceLimitation[] {
  return array(value, field).map((limitation, index) => {
    const item = record(limitation, `${field}[${index}]`)
    return {
      code: stringValue(item.code, `${field}[${index}].code`),
      message: stringValue(item.message, `${field}[${index}].message`),
      scope: stringValue(item.scope, `${field}[${index}].scope`),
    }
  })
}

function parseWindow(value: unknown, field: string) {
  const item = record(value, field)
  return {
    start_time: timestamp(item.start_time, `${field}.start_time`),
    end_time: timestamp(item.end_time, `${field}.end_time`),
    as_of_time: timestamp(item.as_of_time, `${field}.as_of_time`),
  }
}

function normalizeICAO(value: unknown): string {
  const normalized = stringValue(value, 'ICAO').trim().toUpperCase()
  if (!/^[A-Z0-9]{4}$/.test(normalized)) invalid('ICAO must contain four alphanumeric characters.')
  return normalized
}

function record(value: unknown, field: string): Record<string, unknown> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) invalid(`${field} must be an object.`)
  return value as Record<string, unknown>
}
function array(value: unknown, field: string): unknown[] {
  if (!Array.isArray(value)) invalid(`${field} must be an array.`)
  return value
}
function stringValue(value: unknown, field: string): string {
  if (typeof value !== 'string' || value.trim() === '') invalid(`${field} must be a non-empty string.`)
  return value
}
function booleanValue(value: unknown, field: string): boolean {
  if (typeof value !== 'boolean') invalid(`${field} must be a boolean.`)
  return value
}
function finiteNumber(value: unknown, field: string): number {
  if (typeof value !== 'number' || !Number.isFinite(value)) invalid(`${field} must be finite.`)
  return value
}
function nonNegativeNumber(value: unknown, field: string): number {
  const result = finiteNumber(value, field)
  if (result < 0) invalid(`${field} must not be negative.`)
  return result
}
function nonNegativeInteger(value: unknown, field: string): number {
  const result = nonNegativeNumber(value, field)
  if (!Number.isInteger(result)) invalid(`${field} must be an integer.`)
  return result
}
function ratio(value: unknown, field: string): number {
  const result = finiteNumber(value, field)
  if (result < 0 || result > 1) invalid(`${field} must be between zero and one.`)
  return result
}
function timestamp(value: unknown, field: string): string {
  const result = stringValue(value, field)
  if (Number.isNaN(Date.parse(result))) invalid(`${field} must be a timestamp.`)
  return result
}
function fingerprint(value: unknown, field: string): string {
  const result = stringValue(value, field)
  if (!/^sha256:[0-9a-f]{64}$/.test(result)) invalid(`${field} must be a SHA-256 fingerprint.`)
  return result
}
function enumValue<T extends string>(value: unknown, field: string, values: Set<T>): T {
  const result = stringValue(value, field) as T
  if (!values.has(result)) invalid(`${field} is unsupported.`)
  return result
}
function invalid(message: string): never {
  throw new APIRequestError(`The Historical Intelligence response is invalid: ${message}`)
}
