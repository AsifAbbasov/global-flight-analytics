package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/airportproduction"
)

type AirportCongestionIntelligenceResponse struct {
	Version string `json:"version"`
	Status  string `json:"status"`

	Window   AirportIntelligenceWindow `json:"window"`
	ICAOCode string                    `json:"icao_code"`
	Current  AirportStatisticsResponse `json:"current"`

	ObservedWindowCount           int     `json:"observed_window_count"`
	ExpectedWindowCount           int     `json:"expected_window_count"`
	GapWindowCount                int     `json:"gap_window_count"`
	TrailingGapWindowCount        int     `json:"trailing_gap_window_count"`
	CurrentWindowIsLatestExpected bool    `json:"current_window_is_latest_expected"`
	EvidenceCoverage              float64 `json:"evidence_coverage"`
	EvidenceSupport               float64 `json:"evidence_support"`

	BaselineWindowCount            int     `json:"baseline_window_count"`
	BaselineMedianMovementsPerHour float64 `json:"baseline_median_movements_per_hour"`
	PriorPeakMovementsPerHour      float64 `json:"prior_peak_movements_per_hour"`

	CurrentToBaselineRatio       float64 `json:"current_to_baseline_ratio"`
	CurrentToBaselineRatioKnown  bool    `json:"current_to_baseline_ratio_known"`
	CurrentToPriorPeakRatio      float64 `json:"current_to_prior_peak_ratio"`
	CurrentToPriorPeakRatioKnown bool    `json:"current_to_prior_peak_ratio_known"`

	CongestionScore                  float64 `json:"congestion_score"`
	CongestionScoreKnown             bool    `json:"congestion_score_known"`
	ExceedsPriorObservedActivityPeak bool    `json:"exceeds_prior_observed_activity_peak"`

	ScoreSemantics string `json:"score_semantics"`
	ScopeGuard     string `json:"scope_guard"`
	Explanation    string `json:"explanation"`

	Limitations []AirportIntelligenceLimitation `json:"limitations"`
	GeneratedAt time.Time                       `json:"generated_at"`
}

func ToAirportCongestionIntelligenceResponse(
	result airportproduction.CongestionResult,
) AirportCongestionIntelligenceResponse {
	value := result.Congestion
	return AirportCongestionIntelligenceResponse{
		Version:                          result.Version,
		Status:                           string(value.Status),
		Window:                           toAirportIntelligenceWindow(result.Window),
		ICAOCode:                         value.ICAOCode,
		Current:                          toAirportStatisticsResponse(value.Current),
		ObservedWindowCount:              value.ObservedWindowCount,
		ExpectedWindowCount:              value.ExpectedWindowCount,
		GapWindowCount:                   value.GapWindowCount,
		TrailingGapWindowCount:           value.TrailingGapWindowCount,
		CurrentWindowIsLatestExpected:    value.CurrentWindowIsLatestExpected,
		EvidenceCoverage:                 value.EvidenceCoverage,
		EvidenceSupport:                  value.EvidenceSupport,
		BaselineWindowCount:              value.BaselineWindowCount,
		BaselineMedianMovementsPerHour:   value.BaselineMedianMovementsPerHour,
		PriorPeakMovementsPerHour:        value.PriorPeakMovementsPerHour,
		CurrentToBaselineRatio:           value.CurrentToBaselineRatio,
		CurrentToBaselineRatioKnown:      value.CurrentToBaselineRatioKnown,
		CurrentToPriorPeakRatio:          value.CurrentToPriorPeakRatio,
		CurrentToPriorPeakRatioKnown:     value.CurrentToPriorPeakRatioKnown,
		CongestionScore:                  value.CongestionScore,
		CongestionScoreKnown:             value.CongestionScoreKnown,
		ExceedsPriorObservedActivityPeak: value.ExceedsPriorObservedActivityPeak,
		ScoreSemantics:                   value.ScoreSemantics,
		ScopeGuard:                       value.ScopeGuard,
		Explanation:                      value.Explanation,
		Limitations:                      toAirportIntelligenceLimitations(result.Limitations),
		GeneratedAt:                      result.GeneratedAt,
	}
}
