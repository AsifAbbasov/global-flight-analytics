package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
)

type AnalyticalMetricResponse struct {
	Metric           string                              `json:"metric"`
	Status           string                              `json:"status"`
	Value            any                                 `json:"value,omitempty"`
	HasValue         bool                                `json:"has_value"`
	Confidence       AnalyticalConfidenceResponse        `json:"confidence"`
	DataQuality      *dataqualitycontract.Report         `json:"data_quality,omitempty"`
	Eligibility      *AnalyticalEligibilityResponse      `json:"eligibility,omitempty"`
	Scope            AnalyticalScopeResponse             `json:"scope"`
	Sources          []AnalyticalSourceResponse          `json:"sources"`
	Warnings         []AnalyticalNoticeResponse          `json:"warnings"`
	Limitations      []AnalyticalNoticeResponse          `json:"limitations"`
	CalculatedAt     time.Time                           `json:"calculated_at"`
	Failure          *AnalyticalFailureResponse          `json:"failure,omitempty"`
	ConfidenceReport *AnalyticalConfidenceReportResponse `json:"confidence_report,omitempty"`
}

type AnalyticalConfidenceResponse struct {
	Level   string                     `json:"level"`
	Score   float64                    `json:"score"`
	Reasons []AnalyticalNoticeResponse `json:"reasons"`
}

type AnalyticalEligibilityResponse struct {
	Capability  string    `json:"capability"`
	Allowed     bool      `json:"allowed"`
	Reasons     []string  `json:"reasons"`
	EvaluatedAt time.Time `json:"evaluated_at"`
}

type AnalyticalScopeResponse struct {
	Capability   string                          `json:"capability"`
	InputCount   int                             `json:"input_count"`
	AllowedCount int                             `json:"allowed_count"`
	DeniedCount  int                             `json:"denied_count"`
	Reasons      []AnalyticalScopeReasonResponse `json:"reasons"`
	EvaluatedAt  time.Time                       `json:"evaluated_at"`
}

type AnalyticalScopeReasonResponse struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type AnalyticalSourceResponse struct {
	Name         string                     `json:"name"`
	Role         string                     `json:"role"`
	ObservedFrom time.Time                  `json:"observed_from,omitempty"`
	ObservedTo   time.Time                  `json:"observed_to,omitempty"`
	RetrievedAt  time.Time                  `json:"retrieved_at,omitempty"`
	Limitations  []AnalyticalNoticeResponse `json:"limitations"`
}

type AnalyticalNoticeResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AnalyticalFailureResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retriable bool   `json:"retriable"`
}

type AnalyticalConfidenceReportResponse struct {
	BaseScore    float64                              `json:"base_score"`
	PenaltyScore float64                              `json:"penalty_score"`
	Score        float64                              `json:"score"`
	Level        string                               `json:"level"`
	Factors      []AnalyticalConfidenceFactorResponse `json:"factors"`
	Reasons      []AnalyticalNoticeResponse           `json:"reasons"`
	Warnings     []AnalyticalNoticeResponse           `json:"warnings"`
	Limitations  []AnalyticalNoticeResponse           `json:"limitations"`
	EvaluatedAt  time.Time                            `json:"evaluated_at"`
}

type AnalyticalConfidenceFactorResponse struct {
	Code    string  `json:"code"`
	Kind    string  `json:"kind"`
	Weight  float64 `json:"weight"`
	Value   float64 `json:"value"`
	Impact  float64 `json:"impact"`
	Message string  `json:"message"`
}
