export type ProjectionConfidenceLevel = 'none' | 'low' | 'medium' | 'high'

export interface ProjectionNotice { code: string; message: string }
export interface ProjectionMethod { name: string; version: string; decision_class: string }
export interface ProjectionHorizon { as_of_time: string; end_time: string; step_seconds: number; duration_seconds: number }
export interface ProjectionPosition { latitude: number; longitude: number; altitude_m?: number }
export interface ProjectionUncertainty { horizontal_radius_m: number; vertical_radius_m?: number }
export interface ProjectionConfidenceReason { code: string; message: string; contribution: number }
export interface ProjectionConfidence { score: number; level: ProjectionConfidenceLevel; reasons: ProjectionConfidenceReason[] }
export interface ProjectionLimitation { code: string; message: string; scope: string }
export interface ProjectionExplanation { code: string; message: string }
export interface ProjectionPoint { sequence: number; forecast_time: string; position: ProjectionPosition; uncertainty: ProjectionUncertainty; confidence: ProjectionConfidence }
export interface ProjectionArrivalEstimate {
  airport_icao_code: string
  earliest_time: string
  estimated_time: string
  latest_time: string
  confidence: ProjectionConfidence
  limitations: ProjectionLimitation[]
}
export interface ProjectionInputReference {
  name: string
  classification: string
  source_name: string
  observed_at: string
  retrieved_at: string
  limitation?: string
}
export interface ProjectionProvenance {
  input_fingerprint: string
  inputs: ProjectionInputReference[]
  latest_input_observed_at: string
}
export interface ProjectionResult {
  schema_version: string
  status: string
  trajectory_id: string
  flight_id: string
  aircraft_id: string
  icao24: string
  callsign: string
  method: ProjectionMethod
  horizon: ProjectionHorizon
  points: ProjectionPoint[]
  arrival?: ProjectionArrivalEstimate
  confidence: ProjectionConfidence
  limitations: ProjectionLimitation[]
  explanations: ProjectionExplanation[]
  scope_guard: string
  provenance: ProjectionProvenance
  generated_at: string
}

// Evidence fields are strategy-specific and are not rendered by Stage 13.1.
// Keeping their values unknown prevents the client from inventing a narrower
// contract while preserving the complete backend response for later panels.
export interface ProjectionEvidence {
  neighbor_selection?: Record<string, unknown>
  pattern_confidence?: Record<string, unknown>
  freshness?: Record<string, unknown>
  route_frequency?: Record<string, unknown>
}

export interface ProjectionIntelligenceResponse {
  version: string
  strategy: string
  fallback_reason: string
  arrival_status: string
  projection: ProjectionResult
  evidence: ProjectionEvidence
  notices: ProjectionNotice[]
  input_fingerprint: string
  generated_at: string
}

export interface ProjectionIntelligenceRequest {
  trajectoryID: string
  asOfTime: string
  durationSeconds: number
  signal?: AbortSignal
}
