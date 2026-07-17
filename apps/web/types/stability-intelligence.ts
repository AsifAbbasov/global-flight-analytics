import type { ProjectionIntelligenceResponse } from '@/types/projection-intelligence'

export interface StabilityIntelligenceVersion {
  version_id: string
  ordinal: number
  parent_version_id?: string
  method_name: string
  method_version: string
  policy_version: string
  implementation_version: string
  input_fingerprint: string
  output_fingerprint: string
  decision_fingerprint: string
  created_at: string
}

export interface StabilityIntelligenceTransitionMetrics {
  aligned_point_count: number
  aligned_point_share: number
  mean_horizontal_shift_kilometers: number
  maximum_horizontal_shift_kilometers: number
  aggregate_confidence_delta: number
  mean_relative_horizontal_uncertainty_change: number
  arrival_comparable: boolean
  arrival_shift_seconds: number
  method_changed: boolean
  policy_changed: boolean
  implementation_changed: boolean
  input_changed: boolean
  output_changed: boolean
}

export interface StabilityIntelligenceTransition {
  baseline_version_id: string
  candidate_version_id: string
  level: string
  score: number
  metrics: StabilityIntelligenceTransitionMetrics
  input_fingerprint: string
  evaluated_at: string
}

export interface StabilityIntelligenceAnalysisMetrics {
  version_count: number
  transition_count: number
  comparable_transition_count: number
  stable_transition_share: number
  comparable_transition_share: number
  material_change_share: number
  mean_stability_score: number
  minimum_stability_score: number
  score_standard_deviation: number
  longest_stable_run: number
  method_change_count: number
  policy_change_count: number
  implementation_change_count: number
  input_change_count: number
  output_change_count: number
  mean_horizontal_shift_kilometers: number
  maximum_horizontal_shift_kilometers: number
  latest_level: string
}

export interface StabilityIntelligenceAnalysis {
  status: string
  trend: string
  health: string
  metrics: StabilityIntelligenceAnalysisMetrics
  confidence_score: number
  confidence_level: string
  input_fingerprint: string
}

export interface StabilityIntelligenceConfidenceSummary {
  status: string
  score: number
  level: string
  target_node_id: string
  limiting_dependency_id?: string
  input_fingerprint: string
}

export interface StabilityIntelligenceFailure {
  rank: number
  code: string
  category: string
  severity: string
  classification: string
  summary: string
  detail: string
  source: string
  blocks_use: boolean
  priority_score: number
  evidence_fingerprints: string[]
}

export interface StabilityIntelligenceFailureExplanation {
  status: string
  primary_code: string
  blocking_count: number
  warning_count: number
  unknown_cause_count: number
  confidence_score: number
  confidence_level: string
  failures: StabilityIntelligenceFailure[]
  input_fingerprint: string
}

export interface StabilityIntelligenceInterventionGuard {
  status: string
  claim_kind: string
  decision: string
  confidence_score: number
  evidence_count: number
  unknown_evidence_count: number
  estimated_evidence_count: number
  evidence_completeness: number
  input_fingerprint: string
}

export interface StabilityIntelligenceScopeViolation {
  code: string
  claim_code: string
  message: string
  blocking: boolean
}

export interface StabilityIntelligenceScopeEnforcement {
  status: string
  decision: string
  claim_count: number
  allowed_count: number
  limited_count: number
  blocked_count: number
  violations: StabilityIntelligenceScopeViolation[]
  input_fingerprint: string
}

export interface StabilityIntelligenceResponse {
  version: string
  trajectory_id: string
  as_of_times: string[]
  projections: ProjectionIntelligenceResponse[]
  forecast_versions: StabilityIntelligenceVersion[]
  transitions: StabilityIntelligenceTransition[]
  forecast_analysis: StabilityIntelligenceAnalysis
  propagated_confidence: StabilityIntelligenceConfidenceSummary
  failure_explanation: StabilityIntelligenceFailureExplanation
  unknown_intervention: StabilityIntelligenceInterventionGuard
  scope_enforcement: StabilityIntelligenceScopeEnforcement
  scope_guards: string[]
  input_fingerprint: string
  generated_at: string
}

export interface StabilityIntelligenceRequest {
  trajectoryID: string
  asOfTimes: string[]
  durationSeconds: number
  signal?: AbortSignal
}
