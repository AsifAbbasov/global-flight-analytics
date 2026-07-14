export type AnalyticalStatus = 'complete' | 'limited' | 'denied' | 'failed'

export type AnalyticalConfidenceLevel = 'none' | 'low' | 'medium' | 'high'

export type AnalyticalFactorKind = 'evidence' | 'penalty'

export interface AnalyticalNotice {
  code: string
  message: string
}

export interface AnalyticalConfidence {
  level: AnalyticalConfidenceLevel
  score: number
  reasons: AnalyticalNotice[]
}

export interface AnalyticalEligibility {
  capability: string
  allowed: boolean
  reasons: string[]
  evaluated_at: string
}

export interface AnalyticalScopeReason {
  reason: string
  count: number
}

export interface AnalyticalScope {
  capability: string
  input_count: number
  allowed_count: number
  denied_count: number
  reasons: AnalyticalScopeReason[]
  evaluated_at: string
}

export interface AnalyticalSource {
  name: string
  role: string
  observed_from?: string
  observed_to?: string
  retrieved_at?: string
  limitations: AnalyticalNotice[]
}

export interface AnalyticalFailure {
  code: string
  message: string
  retriable: boolean
}

export interface AnalyticalConfidenceFactor {
  code: string
  kind: AnalyticalFactorKind
  weight: number
  value: number
  impact: number
  message: string
}

export interface AnalyticalConfidenceReport {
  base_score: number
  penalty_score: number
  score: number
  level: AnalyticalConfidenceLevel
  factors: AnalyticalConfidenceFactor[]
  reasons: AnalyticalNotice[]
  warnings: AnalyticalNotice[]
  limitations: AnalyticalNotice[]
  evaluated_at: string
}

export interface AnalyticalMetric<TValue extends number = number> {
  metric: string
  status: AnalyticalStatus
  value?: TValue
  has_value: boolean
  confidence: AnalyticalConfidence
  eligibility?: AnalyticalEligibility
  scope: AnalyticalScope
  sources: AnalyticalSource[]
  warnings: AnalyticalNotice[]
  limitations: AnalyticalNotice[]
  calculated_at: string
  failure?: AnalyticalFailure
  confidence_report?: AnalyticalConfidenceReport
}

export interface RecentTrajectoryMetricParameters {
  windowMinutes?: number
  limit?: number
  regionCode?: string
}

export interface TrafficDensityMetricParameters
  extends RecentTrajectoryMetricParameters {
  areaSquareKilometers?: number
}

export interface AirportActivityMetricParameters {
  arrivalTrajectoryIDs?: string[]
  departureTrajectoryIDs?: string[]
}

export interface CoverageScoreMetricParameters {
  observedSamples: number
  expectedSamples: number
}

export interface DataFreshnessMetricParameters {
  observedAt: string
  maximumAgeSeconds: number
}
