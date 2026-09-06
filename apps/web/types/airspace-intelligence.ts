export type AirspaceNotice = {
  code: string
  message: string
  scope?: string
}

export type AirspaceConfidenceReason = {
  code: string
  message: string
  contribution: number
}

export type AirspaceConfidence = {
  score: number
  level: string
  reasons: AirspaceConfidenceReason[]
}

export type AirspaceOccupancyMetrics = {
  bucket_count: number
  expected_bucket_count: number
  aircraft_observation_count: number
  unique_aircraft_count: number
  unknown_altitude_count: number
  peak_aircraft_per_bucket: number
  peak_occupied_cells: number
  mean_aircraft_per_bucket: number
  temporal_coverage: number
}

export type AirspaceRegionMetrics = {
  snapshot_count: number
  bucket_count: number
  unique_aircraft_count: number
  aircraft_observation_count: number
  occupied_cell_count: number
  sector_report_count: number
  current_aircraft_count: number
  peak_aircraft_per_bucket: number
  mean_aircraft_per_bucket: number
  mean_complexity_score: number
  peak_complexity_score: number
  airspace_pressure_index: number
  peak_airspace_pressure_index: number
  moderate_sector_count: number
  high_sector_count: number
  severe_sector_count: number
  contextual_risk_count: number
  elevated_risk_count: number
  high_risk_count: number
  indeterminate_risk_count: number
  unknown_altitude_count: number
  temporal_coverage: number
  occupancy_trend: string
  highest_complexity_level: string
}

export type AirspaceProvenance = {
  input_fingerprint: string
  source_names: string[]
  latest_observed_at: string
}

export type AirspaceRegionAnalyticsResponse = {
  version: string
  schema_version: string
  status: string
  region_code: string
  window_start: string
  window_end: string
  occupancy: {
    metrics: AirspaceOccupancyMetrics
  }
  metrics: AirspaceRegionMetrics
  confidence: AirspaceConfidence
  limitations: AirspaceNotice[]
  explanations: AirspaceNotice[]
  scope_guard: string
  provenance: AirspaceProvenance
  generated_at: string
}

export type AirspaceRegionAnalyticsRequest = {
  regionCode: string
  asOfTime: string
  windowSeconds: number
  signal?: AbortSignal
}
