package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/etareliability"
)

type ETAReliabilityRoute struct {
	OriginICAOCode string `json:"origin_icao_code"`
	DestinationICAOCode string `json:"destination_icao_code"`
}

type ETAReliabilityMethod struct {
	Name string `json:"name"`
	Version string `json:"version"`
	DecisionClass string `json:"decision_class"`
}

type ETAReliabilityMetrics struct {
	SampleCount int `json:"sample_count"`
	MedianAbsoluteErrorSeconds float64 `json:"median_absolute_error_seconds"`
	P80AbsoluteErrorSeconds float64 `json:"p80_absolute_error_seconds"`
	WithinFiveMinutesRatio float64 `json:"within_five_minutes_ratio"`
	WithinTenMinutesRatio float64 `json:"within_ten_minutes_ratio"`
	IntervalCoverageRatio float64 `json:"interval_coverage_ratio"`
}

type ETAReliabilityNotice struct {
	Code string `json:"code"`
	Message string `json:"message"`
}

type ETAReliabilityResponse struct {
	Version string `json:"version"`
	Status string `json:"status"`
	TrajectoryID string `json:"trajectory_id"`
	Route ETAReliabilityRoute `json:"route"`
	Method ETAReliabilityMethod `json:"method"`
	TargetLeadSeconds int64 `json:"target_lead_seconds"`
	LeadToleranceSeconds int64 `json:"lead_tolerance_seconds"`
	EndpointRadiusKM float64 `json:"endpoint_radius_km"`
	CandidateCount int `json:"candidate_count"`
	EligibleSampleCount int `json:"eligible_sample_count"`
	Metrics *ETAReliabilityMetrics `json:"metrics,omitempty"`
	EvidenceClass string `json:"evidence_class"`
	Limitations []ETAReliabilityNotice `json:"limitations"`
	InputFingerprint string `json:"input_fingerprint"`
	GeneratedAt time.Time `json:"generated_at"`
}

func ToETAReliabilityResponse(result etareliability.Result) ETAReliabilityResponse {
	response := ETAReliabilityResponse{
		Version: result.Version,
		Status: string(result.Status),
		TrajectoryID: result.TrajectoryID,
		Route: ETAReliabilityRoute{
			OriginICAOCode: result.Route.OriginICAOCode,
			DestinationICAOCode: result.Route.DestinationICAOCode,
		},
		Method: ETAReliabilityMethod{
			Name: result.Method.Name,
			Version: result.Method.Version,
			DecisionClass: string(result.Method.DecisionClass),
		},
		TargetLeadSeconds: result.TargetLeadSeconds,
		LeadToleranceSeconds: result.LeadToleranceSeconds,
		EndpointRadiusKM: result.EndpointRadiusKM,
		CandidateCount: result.CandidateCount,
		EligibleSampleCount: result.EligibleSampleCount,
		EvidenceClass: result.EvidenceClass,
		InputFingerprint: result.InputFingerprint,
		GeneratedAt: result.GeneratedAt.UTC(),
		Limitations: make([]ETAReliabilityNotice, 0, len(result.Limitations)),
	}
	for _, item := range result.Limitations {
		response.Limitations = append(response.Limitations, ETAReliabilityNotice{Code: item.Code, Message: item.Message})
	}
	if result.Metrics != nil {
		response.Metrics = &ETAReliabilityMetrics{
			SampleCount: result.Metrics.SampleCount,
			MedianAbsoluteErrorSeconds: result.Metrics.MedianAbsoluteErrorSeconds,
			P80AbsoluteErrorSeconds: result.Metrics.P80AbsoluteErrorSeconds,
			WithinFiveMinutesRatio: result.Metrics.WithinFiveMinutesRatio,
			WithinTenMinutesRatio: result.Metrics.WithinTenMinutesRatio,
			IntervalCoverageRatio: result.Metrics.IntervalCoverageRatio,
		}
	}
	return response
}
