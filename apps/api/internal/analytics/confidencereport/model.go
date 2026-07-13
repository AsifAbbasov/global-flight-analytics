package confidencereport

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
)

type FactorKind string

const (
	FactorKindEvidence FactorKind = "evidence"
	FactorKindPenalty  FactorKind = "penalty"
)

func (
	kind FactorKind,
) IsKnown() bool {
	switch kind {
	case FactorKindEvidence,
		FactorKindPenalty:
		return true

	default:
		return false
	}
}

type Factor struct {
	Code    string
	Kind    FactorKind
	Weight  float64
	Value   float64
	Message string
}

type Contribution struct {
	Code    string
	Kind    FactorKind
	Weight  float64
	Value   float64
	Impact  float64
	Message string
}

type Request struct {
	Factors     []Factor
	Warnings    []analyticalresult.Notice
	Limitations []analyticalresult.Notice
	EvaluatedAt time.Time
}

type Report struct {
	BaseScore    float64
	PenaltyScore float64
	Score        float64
	Level        analyticalresult.ConfidenceLevel
	Factors      []Contribution
	Reasons      []analyticalresult.Notice
	Warnings     []analyticalresult.Notice
	Limitations  []analyticalresult.Notice
	EvaluatedAt  time.Time
}

func (
	report Report,
) HasPenalty() bool {
	return report.PenaltyScore > 0
}

func (
	report Report,
) Factor(
	code string,
) (Contribution, bool) {
	for _, factor := range report.Factors {
		if factor.Code == code {
			return factor, true
		}
	}

	return Contribution{}, false
}

func (
	report Report,
) AnalyticalConfidence() analyticalresult.Confidence {
	return analyticalresult.Confidence{
		Level: report.Level,
		Score: report.Score,
		Reasons: append(
			[]analyticalresult.Notice(nil),
			report.Reasons...,
		),
	}
}
