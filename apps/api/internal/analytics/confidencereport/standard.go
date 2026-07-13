package confidencereport

const (
	FactorCodeTrajectoryQuality   = "trajectory_quality"
	FactorCodeIdentityReliability = "identity_reliability"
	FactorCodeDataFreshness       = "data_freshness"
	FactorCodeObservationCoverage = "observation_coverage"
	FactorCodeRecentContinuity    = "recent_continuity"
	FactorCodeSourceCoverage      = "source_coverage"
	FactorCodeMethodStability     = "method_stability"

	FactorCodeCoverageGapPenalty         = "coverage_gap_penalty"
	FactorCodeProviderDegradationPenalty = "provider_degradation_penalty"
	FactorCodeFallbackSourcePenalty      = "fallback_source_penalty"
)

func Evidence(
	code string,
	weight float64,
	value float64,
	message string,
) Factor {
	return Factor{
		Code:    code,
		Kind:    FactorKindEvidence,
		Weight:  weight,
		Value:   value,
		Message: message,
	}
}

func Penalty(
	code string,
	weight float64,
	value float64,
	message string,
) Factor {
	return Factor{
		Code:    code,
		Kind:    FactorKindPenalty,
		Weight:  weight,
		Value:   value,
		Message: message,
	}
}
