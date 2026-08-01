// FRONTEND_HISTORICAL_ANALYTICS_COMPARISON_V1

export type HistoricalScopeType = 'global' | 'airport' | 'route'
export type HistoricalGranularity = 'hour' | 'day' | 'week' | 'custom'
export type HistoricalSeriesStatus = 'unavailable' | 'partial' | 'complete'
export type HistoricalBucketStatus = 'unavailable' | 'partial' | 'complete'
export type HistoricalConfidenceLevel = 'none' | 'low' | 'medium' | 'high'
export type HistoricalTrendDirection = 'unavailable' | 'down' | 'flat' | 'up'

export type HistoricalMetricName =
  | 'active_aircraft'
  | 'flight_count'
  | 'trajectory_count'
  | 'observation_count'
  | 'traffic_density'
  | 'airport_departures'
  | 'airport_arrivals'
  | 'airport_operations'
  | 'unique_aircraft'
  | 'active_routes'
  | 'route_observations'
  | 'route_confidence'
  | 'complete_route_ratio'
  | 'partial_route_ratio'
  | 'unavailable_route_ratio'
  | 'great_circle_distance_km'

export interface HistoricalIntelligenceTimeWindow {
  start_time: string
  end_time: string
  as_of_time: string
}

export interface HistoricalIntelligenceScope {
  type: HistoricalScopeType
  region_code?: string
  airport_icao_code?: string
  origin_icao_code?: string
  destination_icao_code?: string
}

export interface HistoricalIntelligenceMetric {
  name: HistoricalMetricName
  unit: string
  aggregation:
    | 'count'
    | 'sum'
    | 'minimum'
    | 'maximum'
    | 'average'
    | 'median'
    | 'ratio'
}

export interface HistoricalIntelligenceConfidenceReason {
  code: string
  message: string
  contribution: number
}

export interface HistoricalIntelligenceConfidence {
  score: number
  level: HistoricalConfidenceLevel
  sample_count: number
  reasons: HistoricalIntelligenceConfidenceReason[]
}

export interface HistoricalIntelligenceLimitation {
  code: string
  message: string
  scope: string
}

export interface HistoricalIntelligencePoint {
  start_time: string
  end_time: string
  status: HistoricalBucketStatus
  value: number
  sample_count: number
  coverage_ratio: number
  confidence: HistoricalIntelligenceConfidence
  limitations: HistoricalIntelligenceLimitation[]
}

export interface HistoricalIntelligenceSummary {
  point_count: number
  total: number
  minimum: number
  maximum: number
  average: number
  median: number
}

export interface HistoricalIntelligencePeriodComparison {
  previous_window: HistoricalIntelligenceTimeWindow
  current_value: number
  previous_value: number
  absolute_change: number
  percentage_change?: number
  direction: HistoricalTrendDirection
}

export interface HistoricalIntelligenceProvenance {
  builder_version: string
  input_fingerprint: string
  source_names: string[]
  latest_source_updated_at: string
}

export interface HistoricalIntelligenceResult {
  schema_version: string
  status: HistoricalSeriesStatus
  metric: HistoricalIntelligenceMetric
  scope: HistoricalIntelligenceScope
  window: HistoricalIntelligenceTimeWindow
  granularity: HistoricalGranularity
  points: HistoricalIntelligencePoint[]
  summary: HistoricalIntelligenceSummary
  comparison?: HistoricalIntelligencePeriodComparison
  confidence: HistoricalIntelligenceConfidence
  limitations: HistoricalIntelligenceLimitation[]
  provenance: HistoricalIntelligenceProvenance
  generated_at: string
}

export interface HistoricalIntelligenceAggregateRecord {
  id: string
  input_fingerprint: string
  stored_at: string
  result: HistoricalIntelligenceResult
}

export interface HistoricalIntelligenceAggregateHistory {
  items: HistoricalIntelligenceAggregateRecord[]
  has_more: boolean
  next_cursor?: string
}

export interface HistoricalIntelligenceSelection {
  scope: HistoricalScopeType
  metric: HistoricalMetricName
  granularity: HistoricalGranularity
  airportICAO?: string
  originICAO?: string
  destinationICAO?: string
}
