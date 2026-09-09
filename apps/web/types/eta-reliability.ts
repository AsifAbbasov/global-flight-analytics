export type ETAReliabilityStatus = 'unavailable' | 'limited' | 'complete'

export interface ETAReliabilityRoute {
  origin_icao_code: string
  destination_icao_code: string
}

export interface ETAReliabilityMethod {
  name: string
  version: string
  decision_class: string
}

export interface ETAReliabilityMetrics {
  sample_count: number
  median_absolute_error_seconds: number
  p80_absolute_error_seconds: number
  within_five_minutes_ratio: number
  within_ten_minutes_ratio: number
  interval_coverage_ratio: number
}

export interface ETAReliabilityNotice {
  code: string
  message: string
}

export interface ETAReliabilityResponse {
  version: string
  status: ETAReliabilityStatus
  trajectory_id: string
  route: ETAReliabilityRoute
  method: ETAReliabilityMethod
  target_lead_seconds: number
  lead_tolerance_seconds: number
  endpoint_radius_km: number
  candidate_count: number
  eligible_sample_count: number
  metrics?: ETAReliabilityMetrics
  evidence_class: 'historically_recomputed_from_persisted_observations_with_endpoint_proxy'
  limitations: ETAReliabilityNotice[]
  input_fingerprint: string
  generated_at: string
}

export interface ETAReliabilityRequest {
  trajectoryID: string
  asOfTime: string
  durationSeconds: number
  signal?: AbortSignal
}
