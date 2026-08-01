// FRONTEND_UNIFIED_AIRPORT_ANALYTICS_WORKSPACE_V1

export interface AirportIntelligenceWindow {
  start_time: string
  end_time: string
  as_of_time: string
  completed_days: number
}

export interface AirportIntelligenceLimitation {
  code: string
  message: string
}

export interface AirportRankingWeights {
  movements: number
  routes: number
  observations: number
  intensity: number
  coverage: number
  freshness: number
}

export interface AirportRankedItem {
  position: number
  icao_code: string
  iata_code: string
  name: string
  city: string
  country: string
  activity_score: number
  data_confidence: number
  movements_component: number
  routes_component: number
  observations_component: number
  intensity_component: number
  coverage_score: number
  freshness_score: number
  total_movements: number
  active_routes: number
  observed_samples: number
  expected_samples: number
  movements_per_hour: number
  active_aircraft: number
}

export interface AirportIntelligenceRanking {
  version: string
  window: AirportIntelligenceWindow
  weights: AirportRankingWeights
  airports: AirportRankedItem[]
  limitations: AirportIntelligenceLimitation[]
  generated_at: string
}

export interface AirportPassportIdentity {
  icao_code: string
  iata_code: string
  name: string
}

export interface AirportPassportLocation {
  city: string
  country: string
  latitude: number
  longitude: number
  elevation_m: number | null
  elevation_status: string
  timezone: string
}

export interface AirportPassportOperations {
  arrivals: number
  departures: number
  activity: number
  active_aircraft: number
}

export interface AirportPassportDataQuality {
  freshness_score: number
  coverage_score: number
  observed_at: string
}

export interface AirportPassport {
  identity: AirportPassportIdentity
  location: AirportPassportLocation
  operations: AirportPassportOperations
  data_quality: AirportPassportDataQuality
  description: string
  generated_at: string
}

export interface AirportStatistics {
  icao_code: string
  window_start: string
  window_end: string
  arrivals: number
  departures: number
  total_movements: number
  arrival_share: number
  departure_share: number
  movements_per_hour: number
  active_aircraft: number
  active_routes: number
  observed_samples: number
  expected_samples: number
  coverage_score: number
  freshness_score: number
  latest_observation_at: string
  generated_at: string
}

export interface AirportRankingSummary {
  position: number
  total_airports: number
  activity_score: number
  data_confidence: number
  movements_component: number
  routes_component: number
  observations_component: number
  intensity_component: number
}

export interface AirportIntelligenceOverview {
  version: string
  window: AirportIntelligenceWindow
  passport: AirportPassport
  statistics: AirportStatistics
  ranking: AirportRankingSummary
  limitations: AirportIntelligenceLimitation[]
  generated_at: string
}

export interface AirportIntelligenceHistory {
  version: string
  window: AirportIntelligenceWindow
  icao_code: string
  entries: AirportStatistics[]
  limitations: AirportIntelligenceLimitation[]
  generated_at: string
}

export interface AirportTrendPoint {
  window_start: string
  window_end: string
  total_movements: number
  movements_per_hour: number
  active_routes: number
  coverage_score: number
  freshness_score: number
}

export interface AirportIntelligenceTrends {
  version: string
  window: AirportIntelligenceWindow
  icao_code: string
  compared_windows: number
  window_duration_seconds: number
  direction: string
  baseline: AirportTrendPoint
  current: AirportTrendPoint
  peak: AirportTrendPoint
  total_movements_change: number
  movements_per_hour_change: number
  movements_per_hour_change_percent: number
  movements_per_hour_change_percent_known: boolean
  active_routes_change: number
  coverage_score_change: number
  freshness_score_change: number
  gap_count: number
  gap_duration_seconds: number
  observed_duration_seconds: number
  continuity_score: number
  limitations: AirportIntelligenceLimitation[]
  generated_at: string
}

export interface AirportIntelligenceWindowOptions {
  days: number
  signal?: AbortSignal
}

export interface AirportIntelligenceRankingOptions
  extends AirportIntelligenceWindowOptions {
  limit: number
}
