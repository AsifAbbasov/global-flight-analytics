export type WeatherConfidenceLevel = 'none' | 'low' | 'medium' | 'high'
export interface WeatherNotice { code: string; message: string }
export interface WeatherLimitation extends WeatherNotice { scope: string }
export interface WeatherConfidence { score: number; level: WeatherConfidenceLevel; reasons: Array<WeatherNotice & { contribution: number }> }
export interface WeatherMetricSummary { present_count: number; coverage_ratio: number; minimum?: number; maximum?: number; mean?: number }
export interface WeatherDirectionSummary { present_count: number; coverage_ratio: number; mean_direction_degrees?: number; concentration?: number }
export interface WeatherSample {
  sequence: number
  position: { latitude: number; longitude: number; altitude_meters?: number; vertical_reference: string }
  source: { provider: string; dataset: string; evidence_kind: string; horizontal_resolution_kilometers?: number; temporal_resolution_seconds: number }
  features: Record<string, unknown>
  valid_at: string
  available_at: string
  retrieved_at: string
}
export interface WeatherContextResponse {
  version: string
  trajectory_id: string
  as_of_time: string
  weather: {
    schema_version: string; status: string; trajectory_id: string; as_of_time: string; samples: WeatherSample[]
    confidence: WeatherConfidence; limitations: WeatherLimitation[]; explanations: WeatherNotice[]; scope_guard: string
    input_fingerprint: string; source_names: string[]; latest_available_at: string; generated_at: string
  }
  trust: {
    version: string; decision: string; usable: boolean; as_of_time: string; score: number
    components: Array<{ name: string; score: number; weight: number }>; allowed_scopes: string[]
    limitations: WeatherNotice[]; explanations: WeatherNotice[]; input_fingerprint: string
  }
  alignment: {
    version: string; status: string; trajectory_id: string; as_of_time: string; trust_decision: string; trust_score: number
    point_count: number; aligned_count: number; unmatched_count: number; coverage_ratio: number
    matches: unknown[]; limitations: WeatherNotice[]; explanations: WeatherNotice[]; input_fingerprint: string; generated_at: string
  }
  encounter: {
    version: string; status: string; trajectory_id: string; as_of_time: string; alignment_status: string; alignment_coverage_ratio: number
    point_count: number; encounter_point_count: number; unprofiled_point_count: number; profile_coverage_ratio: number
    temperature_celsius: WeatherMetricSummary; precipitation_millimeters: WeatherMetricSummary; cloud_cover_percent: WeatherMetricSummary
    wind_speed_meters_per_second: WeatherMetricSummary; wind_gusts_meters_per_second: WeatherMetricSummary; wind_direction_degrees: WeatherDirectionSummary
    limitations: WeatherNotice[]; explanations: WeatherNotice[]; input_fingerprint: string; generated_at: string
  }
  uncertainty: {
    version: string; status: string; trajectory_id: string; as_of_time: string; severity_score: number; weather_multiplier: number
    point_adjustments: unknown[]; limitations: WeatherNotice[]; explanations: WeatherNotice[]; input_fingerprint: string; generated_at: string
  }
  input_fingerprint: string
  generated_at: string
}
export interface WeatherContextRequest { trajectoryID: string; asOfTime: string; durationSeconds: number; signal?: AbortSignal }
