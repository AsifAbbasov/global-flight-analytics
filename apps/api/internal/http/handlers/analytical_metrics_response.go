package handlers

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/confidencereport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricexecution"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
)

func toAnalyticalMetricResponse[T any](
	execution metricexecution.Execution[T],
) dto.AnalyticalMetricResponse {
	result := execution.Result

	response := dto.AnalyticalMetricResponse{
		Metric:       execution.MetricID,
		Status:       string(result.Status),
		HasValue:     result.HasValue,
		Confidence:   toAnalyticalConfidenceResponse(result.Confidence),
		DataQuality:  toDataQualityReportResponse(result.DataQuality),
		Eligibility:  toAnalyticalEligibilityResponse(result.Eligibility),
		Scope:        toAnalyticalScopeResponse(execution.Scope),
		Sources:      toAnalyticalSourceResponses(result.Sources),
		Warnings:     toAnalyticalNoticeResponses(result.Warnings),
		Limitations:  toAnalyticalNoticeResponses(result.Limitations),
		CalculatedAt: result.CalculatedAt,
		Failure:      toAnalyticalFailureResponse(result.Failure),
		ConfidenceReport: toAnalyticalConfidenceReportResponse(
			execution.ConfidenceReport,
		),
	}

	if result.HasValue {
		response.Value = result.Value
	}

	return response
}

func toDataQualityReportResponse(
	report *dataqualitycontract.Report,
) *dataqualitycontract.Report {
	if report == nil {
		return nil
	}

	result := report.Clone()
	return &result
}

func toAnalyticalConfidenceResponse(
	confidence analyticalresult.Confidence,
) dto.AnalyticalConfidenceResponse {
	return dto.AnalyticalConfidenceResponse{
		Level:   string(confidence.Level),
		Score:   confidence.Score,
		Reasons: toAnalyticalNoticeResponses(confidence.Reasons),
	}
}

func toAnalyticalEligibilityResponse(
	eligibility *analyticalresult.Eligibility,
) *dto.AnalyticalEligibilityResponse {
	if eligibility == nil {
		return nil
	}

	reasons := make(
		[]string,
		0,
		len(eligibility.Reasons),
	)
	for _, reason := range eligibility.Reasons {
		reasons = append(reasons, string(reason))
	}

	return &dto.AnalyticalEligibilityResponse{
		Capability:  string(eligibility.Capability),
		Allowed:     eligibility.Allowed,
		Reasons:     reasons,
		EvaluatedAt: eligibility.EvaluatedAt,
	}
}

func toAnalyticalScopeResponse(
	scope metricexecution.ScopeSummary,
) dto.AnalyticalScopeResponse {
	reasons := make(
		[]dto.AnalyticalScopeReasonResponse,
		0,
		len(scope.Reasons),
	)
	for _, reason := range scope.Reasons {
		reasons = append(
			reasons,
			dto.AnalyticalScopeReasonResponse{
				Reason: string(reason.Reason),
				Count:  reason.Count,
			},
		)
	}

	return dto.AnalyticalScopeResponse{
		Capability:   string(scope.Capability),
		InputCount:   scope.InputCount,
		AllowedCount: scope.AllowedCount,
		DeniedCount:  scope.DeniedCount,
		Reasons:      reasons,
		EvaluatedAt:  scope.EvaluatedAt,
	}
}

func toAnalyticalSourceResponses(
	sources []analyticalresult.Source,
) []dto.AnalyticalSourceResponse {
	result := make(
		[]dto.AnalyticalSourceResponse,
		0,
		len(sources),
	)

	for _, source := range sources {
		result = append(
			result,
			dto.AnalyticalSourceResponse{
				Name:         source.Name,
				Role:         string(source.Role),
				ObservedFrom: source.ObservedFrom,
				ObservedTo:   source.ObservedTo,
				RetrievedAt:  source.RetrievedAt,
				Limitations: toAnalyticalNoticeResponses(
					source.Limitations,
				),
			},
		)
	}

	return result
}

func toAnalyticalNoticeResponses(
	notices []analyticalresult.Notice,
) []dto.AnalyticalNoticeResponse {
	result := make(
		[]dto.AnalyticalNoticeResponse,
		0,
		len(notices),
	)

	for _, notice := range notices {
		result = append(
			result,
			dto.AnalyticalNoticeResponse{
				Code:    notice.Code,
				Message: notice.Message,
			},
		)
	}

	return result
}

func toAnalyticalFailureResponse(
	failure *analyticalresult.Failure,
) *dto.AnalyticalFailureResponse {
	if failure == nil {
		return nil
	}

	return &dto.AnalyticalFailureResponse{
		Code:      failure.Code,
		Message:   failure.Message,
		Retriable: failure.Retriable,
	}
}

func toAnalyticalConfidenceReportResponse(
	report *confidencereport.Report,
) *dto.AnalyticalConfidenceReportResponse {
	if report == nil {
		return nil
	}

	factors := make(
		[]dto.AnalyticalConfidenceFactorResponse,
		0,
		len(report.Factors),
	)
	for _, factor := range report.Factors {
		factors = append(
			factors,
			dto.AnalyticalConfidenceFactorResponse{
				Code:    factor.Code,
				Kind:    string(factor.Kind),
				Weight:  factor.Weight,
				Value:   factor.Value,
				Impact:  factor.Impact,
				Message: factor.Message,
			},
		)
	}

	return &dto.AnalyticalConfidenceReportResponse{
		BaseScore:    report.BaseScore,
		PenaltyScore: report.PenaltyScore,
		Score:        report.Score,
		Level:        string(report.Level),
		Factors:      factors,
		Reasons:      toAnalyticalNoticeResponses(report.Reasons),
		Warnings:     toAnalyticalNoticeResponses(report.Warnings),
		Limitations:  toAnalyticalNoticeResponses(report.Limitations),
		EvaluatedAt:  report.EvaluatedAt,
	}
}
