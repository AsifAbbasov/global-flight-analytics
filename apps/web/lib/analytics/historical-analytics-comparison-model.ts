// FRONTEND_HISTORICAL_ANALYTICS_COMPARISON_V1
import type {
  HistoricalBucketStatus,
  HistoricalGranularity,
  HistoricalIntelligenceAggregateRecord,
  HistoricalIntelligenceLimitation,
  HistoricalIntelligencePoint,
  HistoricalIntelligenceResult,
  HistoricalIntelligenceSelection,
  HistoricalMetricName,
  HistoricalScopeType,
} from '../../types/historical-intelligence'

export interface HistoricalMetricDefinition {
  name: HistoricalMetricName
  label: string
  description: string
  unit: string
  valueKind: 'count' | 'ratio' | 'continuous'
  scopes: HistoricalScopeType[]
}

export const historicalMetricCatalog: HistoricalMetricDefinition[] = [
  { name: 'active_aircraft', label: 'Active aircraft', description: 'Distinct active aircraft in each bucket.', unit: 'aircraft', valueKind: 'count', scopes: ['global'] },
  { name: 'flight_count', label: 'Flight count', description: 'Eligible flight records in each bucket.', unit: 'flights', valueKind: 'count', scopes: ['global'] },
  { name: 'trajectory_count', label: 'Trajectory count', description: 'Eligible trajectories in each bucket.', unit: 'trajectories', valueKind: 'count', scopes: ['global'] },
  { name: 'observation_count', label: 'Observation count', description: 'Eligible observations in each bucket.', unit: 'observations', valueKind: 'count', scopes: ['global'] },
  { name: 'traffic_density', label: 'Traffic density', description: 'Average observation density per hour.', unit: 'observations/hour', valueKind: 'continuous', scopes: ['global'] },
  { name: 'airport_departures', label: 'Airport departures', description: 'Probable departures associated with one airport.', unit: 'departures', valueKind: 'count', scopes: ['airport'] },
  { name: 'airport_arrivals', label: 'Airport arrivals', description: 'Probable arrivals associated with one airport.', unit: 'arrivals', valueKind: 'count', scopes: ['airport'] },
  { name: 'airport_operations', label: 'Airport operations', description: 'Probable arrivals and departures combined.', unit: 'operations', valueKind: 'count', scopes: ['airport'] },
  { name: 'unique_aircraft', label: 'Unique airport aircraft', description: 'Distinct aircraft identities associated with one airport.', unit: 'aircraft', valueKind: 'count', scopes: ['airport'] },
  { name: 'active_routes', label: 'Active routes', description: 'Distinct probable route pairs.', unit: 'route pairs', valueKind: 'count', scopes: ['global', 'route'] },
  { name: 'route_observations', label: 'Route observations', description: 'Stored Route Intelligence results.', unit: 'route results', valueKind: 'count', scopes: ['global', 'route'] },
  { name: 'route_confidence', label: 'Route confidence', description: 'Average confidence of probable route results.', unit: 'ratio', valueKind: 'ratio', scopes: ['global', 'route'] },
  { name: 'complete_route_ratio', label: 'Complete route ratio', description: 'Share of complete probable route results.', unit: 'ratio', valueKind: 'ratio', scopes: ['global'] },
  { name: 'partial_route_ratio', label: 'Partial route ratio', description: 'Share of partial probable route results.', unit: 'ratio', valueKind: 'ratio', scopes: ['global'] },
  { name: 'unavailable_route_ratio', label: 'Unavailable route ratio', description: 'Share of unavailable probable route results.', unit: 'ratio', valueKind: 'ratio', scopes: ['global'] },
  { name: 'great_circle_distance_km', label: 'Great-circle distance', description: 'Average probable route distance.', unit: 'km', valueKind: 'continuous', scopes: ['global', 'route'] },
]

export const historicalGranularities: HistoricalGranularity[] = [
  'hour',
  'day',
  'week',
  'custom',
]

export function metricsForScope(scope: HistoricalScopeType): HistoricalMetricDefinition[] {
  return historicalMetricCatalog.filter(metric => metric.scopes.includes(scope))
}

export function defaultMetricForScope(scope: HistoricalScopeType): HistoricalMetricName {
  return metricsForScope(scope)[0]?.name ?? 'active_aircraft'
}

export function normalizeICAO(value: string): string | null {
  const normalized = value.trim().toUpperCase()
  return /^[A-Z0-9]{4}$/.test(normalized) ? normalized : null
}

export function normalizeHistoricalSelection(
  selection: HistoricalIntelligenceSelection
): HistoricalIntelligenceSelection {
  const allowed = metricsForScope(selection.scope)
  const metric = allowed.some(item => item.name === selection.metric)
    ? selection.metric
    : defaultMetricForScope(selection.scope)

  if (selection.scope === 'global') {
    return { scope: 'global', metric, granularity: selection.granularity }
  }
  if (selection.scope === 'airport') {
    return {
      scope: 'airport',
      metric,
      granularity: selection.granularity,
      airportICAO: normalizeICAO(selection.airportICAO ?? '') ?? undefined,
    }
  }
  return {
    scope: 'route',
    metric,
    granularity: selection.granularity,
    originICAO: normalizeICAO(selection.originICAO ?? '') ?? undefined,
    destinationICAO: normalizeICAO(selection.destinationICAO ?? '') ?? undefined,
  }
}

export function selectionIsComplete(selection: HistoricalIntelligenceSelection): boolean {
  const normalized = normalizeHistoricalSelection(selection)
  if (normalized.scope === 'global') return true
  if (normalized.scope === 'airport') return normalized.airportICAO !== undefined
  return (
    normalized.originICAO !== undefined &&
    normalized.destinationICAO !== undefined
  )
}

export interface HistoricalSeriesPointView {
  key: string
  startTime: string
  endTime: string
  value: number
  status: HistoricalBucketStatus
  sampleCount: number
  coverageRatio: number
  confidenceScore: number
  limitationCount: number
}

export interface HistoricalSeriesView {
  points: HistoricalSeriesPointView[]
  availableCount: number
  partialCount: number
  unavailableCount: number
  minimumValue: number | null
  maximumValue: number | null
  maximumAbsoluteValue: number
}

export function buildHistoricalSeriesView(
  points: HistoricalIntelligencePoint[]
): HistoricalSeriesView {
  const ordered = [...points].sort((left, right) => {
    const timeOrder = Date.parse(left.start_time) - Date.parse(right.start_time)
    if (timeOrder !== 0) return timeOrder
    return left.end_time.localeCompare(right.end_time)
  })
  const views = ordered.map(point => ({
    key: `${point.start_time}|${point.end_time}`,
    startTime: point.start_time,
    endTime: point.end_time,
    value: point.value,
    status: point.status,
    sampleCount: point.sample_count,
    coverageRatio: point.coverage_ratio,
    confidenceScore: point.confidence.score,
    limitationCount: point.limitations.length,
  }))
  const usableValues = views
    .filter(point => point.status !== 'unavailable' && Number.isFinite(point.value))
    .map(point => point.value)
  const minimumValue = usableValues.length === 0 ? null : Math.min(...usableValues)
  const maximumValue = usableValues.length === 0 ? null : Math.max(...usableValues)
  const maximumAbsoluteValue = Math.max(
    1,
    ...usableValues.map(value => Math.abs(value))
  )
  return {
    points: views,
    availableCount: views.filter(point => point.status === 'complete').length,
    partialCount: views.filter(point => point.status === 'partial').length,
    unavailableCount: views.filter(point => point.status === 'unavailable').length,
    minimumValue,
    maximumValue,
    maximumAbsoluteValue,
  }
}

export interface HistoricalComparisonView {
  available: boolean
  direction: 'unavailable' | 'down' | 'flat' | 'up'
  currentValue: number | null
  previousValue: number | null
  absoluteChange: number | null
  percentageChange: number | null
  previousWindowLabel: string | null
}

export function buildPeriodComparisonView(
  result: HistoricalIntelligenceResult | undefined
): HistoricalComparisonView {
  const comparison = result?.comparison
  if (!comparison) {
    return {
      available: false,
      direction: 'unavailable',
      currentValue: null,
      previousValue: null,
      absoluteChange: null,
      percentageChange: null,
      previousWindowLabel: null,
    }
  }
  return {
    available: true,
    direction: comparison.direction,
    currentValue: comparison.current_value,
    previousValue: comparison.previous_value,
    absoluteChange: comparison.absolute_change,
    percentageChange: comparison.percentage_change ?? null,
    previousWindowLabel: `${comparison.previous_window.start_time}|${comparison.previous_window.end_time}`,
  }
}

export function sortAggregateHistory(
  records: HistoricalIntelligenceAggregateRecord[]
): HistoricalIntelligenceAggregateRecord[] {
  return [...records].sort((left, right) => {
    const storedOrder = Date.parse(right.stored_at) - Date.parse(left.stored_at)
    if (storedOrder !== 0) return storedOrder
    return left.id.localeCompare(right.id)
  })
}

export interface HistoricalRecordComparison {
  comparable: boolean
  reason: string | null
  leftValue: number | null
  rightValue: number | null
  absoluteChange: number | null
  percentageChange: number | null
  direction: 'unavailable' | 'down' | 'flat' | 'up'
}

export function compareHistoricalRecords(
  left: HistoricalIntelligenceAggregateRecord | undefined,
  right: HistoricalIntelligenceAggregateRecord | undefined
): HistoricalRecordComparison {
  if (!left || !right) {
    return unavailableRecordComparison('Select two stored aggregates.')
  }
  if (
    left.result.metric.name !== right.result.metric.name ||
    left.result.granularity !== right.result.granularity ||
    scopeKey(left.result) !== scopeKey(right.result)
  ) {
    return unavailableRecordComparison(
      'Stored aggregates must use the same metric, granularity and scope.'
    )
  }
  const leftValue = comparisonValue(left.result)
  const rightValue = comparisonValue(right.result)
  if (!Number.isFinite(leftValue) || !Number.isFinite(rightValue)) {
    return unavailableRecordComparison('Stored aggregate summaries are not finite.')
  }
  const absoluteChange = rightValue - leftValue
  return {
    comparable: true,
    reason: null,
    leftValue,
    rightValue,
    absoluteChange,
    percentageChange: leftValue === 0 ? null : (absoluteChange / leftValue) * 100,
    direction:
      Math.abs(absoluteChange) <= 1e-9
        ? 'flat'
        : absoluteChange > 0
          ? 'up'
          : 'down',
  }
}

export function mergeHistoricalLimitations(
  ...groups: Array<HistoricalIntelligenceLimitation[] | undefined>
): HistoricalIntelligenceLimitation[] {
  const unique = new Map<string, HistoricalIntelligenceLimitation>()
  for (const group of groups) {
    for (const item of group ?? []) {
      const key = `${item.code.trim().toLowerCase()}|${item.scope.trim().toLowerCase()}|${item.message.trim().toLowerCase()}`
      if (!unique.has(key)) unique.set(key, item)
    }
  }
  return [...unique.values()].sort((left, right) => {
    const scopeOrder = left.scope.localeCompare(right.scope)
    if (scopeOrder !== 0) return scopeOrder
    const codeOrder = left.code.localeCompare(right.code)
    if (codeOrder !== 0) return codeOrder
    return left.message.localeCompare(right.message)
  })
}

export function metricDefinition(name: HistoricalMetricName): HistoricalMetricDefinition {
  return (
    historicalMetricCatalog.find(item => item.name === name) ??
    historicalMetricCatalog[0]
  )
}

function comparisonValue(result: HistoricalIntelligenceResult): number {
  switch (result.metric.aggregation) {
    case 'count':
    case 'sum':
      return result.summary.total
    case 'minimum':
      return result.summary.minimum
    case 'maximum':
      return result.summary.maximum
    case 'median':
      return result.summary.median
    case 'average':
    case 'ratio':
      return result.summary.average
  }
}

function scopeKey(result: HistoricalIntelligenceResult): string {
  const scope = result.scope
  return [
    scope.type,
    scope.region_code ?? '',
    scope.airport_icao_code ?? '',
    scope.origin_icao_code ?? '',
    scope.destination_icao_code ?? '',
  ].join('|')
}

function unavailableRecordComparison(reason: string): HistoricalRecordComparison {
  return {
    comparable: false,
    reason,
    leftValue: null,
    rightValue: null,
    absoluteChange: null,
    percentageChange: null,
    direction: 'unavailable',
  }
}
