export type RouteIntelligenceStatus =
  | 'unavailable'
  | 'partial'
  | 'complete'
export type RouteIntelligenceConfidenceLevel =
  | 'none'
  | 'low'
  | 'medium'
  | 'high'
export type RouteIntelligenceEndpointRole = 'origin' | 'destination'

export interface RouteIntelligenceEvidenceAttribute { key: string; value: string }
export interface RouteIntelligenceEvidence {
  type: string
  source_name: string
  source_version: string
  score: number
  weight: number
  observed_at: string
  summary: string
  attributes: RouteIntelligenceEvidenceAttribute[]
}
export interface RouteIntelligenceConfidenceReason { code: string; message: string; contribution: number }
export interface RouteIntelligenceConfidence {
  score: number
  level: RouteIntelligenceConfidenceLevel
  evidence_count: number
  reasons: RouteIntelligenceConfidenceReason[]
}
export interface RouteIntelligenceLimitation { code: string; message: string; scope: string }
export type AirportElevationStatus = 'observed' | 'unknown' | 'invalid'
export interface RouteIntelligenceAirport {
  icao_code: string
  iata_code: string
  name: string
  city: string
  country: string
  latitude: number
  longitude: number
  elevation_m: number | null
  elevation_status: AirportElevationStatus
  timezone: string
}
export interface RouteIntelligenceEndpoint {
  role: RouteIntelligenceEndpointRole
  airport: RouteIntelligenceAirport
  distance_km: number
  confidence: RouteIntelligenceConfidence
  evidence: RouteIntelligenceEvidence[]
  limitations: RouteIntelligenceLimitation[]
}
export interface RouteIntelligenceWindow { start_time: string; end_time: string; as_of_time: string }
export interface RouteIntelligenceSummary { great_circle_distance_km: number; same_airport: boolean }
export interface RouteIntelligenceProvenance {
  resolver_version: string
  input_fingerprint: string
  trajectory_updated_at: string
  source_names: string[]
}
export interface RouteIntelligenceResult {
  schema_version: string
  status: RouteIntelligenceStatus
  trajectory_id: string
  identity_key: string
  flight_id: string
  aircraft_id: string
  icao24: string
  callsign: string
  window: RouteIntelligenceWindow
  origin?: RouteIntelligenceEndpoint
  destination?: RouteIntelligenceEndpoint
  summary: RouteIntelligenceSummary
  confidence: RouteIntelligenceConfidence
  limitations: RouteIntelligenceLimitation[]
  provenance: RouteIntelligenceProvenance
  generated_at: string
}
export interface RouteIntelligenceRecord {
  id: string
  input_fingerprint: string
  stored_at: string
  result: RouteIntelligenceResult
}
export interface RouteIntelligenceHistory {
  items: RouteIntelligenceRecord[]
  has_more: boolean
  next_before_as_of_time?: string
}
export interface RouteIntelligenceHistoryOptions {
  limit?: number
  beforeAsOfTime?: string
  signal?: AbortSignal
}
