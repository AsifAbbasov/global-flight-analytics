package separationrisk

import (
	"fmt"
	"math"
	"time"
)

const PolicyVersionV1 = "separation-risk-policy-v1"

type RiskWeights struct {
	HorizontalProximity float64
	VerticalProximity   float64
	ClosingMotion       float64
	TemporalAlignment   float64
	EvidenceConfidence  float64
}

type ResultConfidenceWeights struct {
	ScanConfidence       float64
	AssessmentConfidence float64
	EvidenceCompleteness float64
}

type Policy struct {
	Version string

	MaximumCandidateCount     int
	MaximumPairTimeDifference time.Duration

	ElevatedRiskMinimumScore float64
	HighRiskMinimumScore     float64

	ElevatedHorizontalRadiusRatioMaximum float64
	HighHorizontalRadiusRatioMaximum     float64
	ElevatedVerticalRadiusRatioMaximum   float64
	HighVerticalRadiusRatioMaximum       float64

	ElevatedClosingRateMinimumMetersPerSecond float64
	HighClosingRateMinimumMetersPerSecond     float64

	MediumConfidenceMinimumScore float64
	HighConfidenceMinimumScore   float64

	RiskWeights             RiskWeights
	ResultConfidenceWeights ResultConfidenceWeights
}

func DefaultPolicy() Policy {
	return Policy{
		Version:                                   PolicyVersionV1,
		MaximumCandidateCount:                     499500,
		MaximumPairTimeDifference:                 30 * time.Second,
		ElevatedRiskMinimumScore:                  0.48,
		HighRiskMinimumScore:                      0.72,
		ElevatedHorizontalRadiusRatioMaximum:      0.65,
		HighHorizontalRadiusRatioMaximum:          0.35,
		ElevatedVerticalRadiusRatioMaximum:        0.75,
		HighVerticalRadiusRatioMaximum:            0.35,
		ElevatedClosingRateMinimumMetersPerSecond: 3,
		HighClosingRateMinimumMetersPerSecond:     10,
		MediumConfidenceMinimumScore:              0.50,
		HighConfidenceMinimumScore:                0.80,
		RiskWeights: RiskWeights{
			HorizontalProximity: 0.30,
			VerticalProximity:   0.25,
			ClosingMotion:       0.20,
			TemporalAlignment:   0.10,
			EvidenceConfidence:  0.15,
		},
		ResultConfidenceWeights: ResultConfidenceWeights{
			ScanConfidence:       0.45,
			AssessmentConfidence: 0.35,
			EvidenceCompleteness: 0.20,
		},
	}
}

func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 || policy.MaximumCandidateCount <= 0 || policy.MaximumPairTimeDifference <= 0 {
		return fmt.Errorf("%w: version or capacity", ErrInvalidPolicy)
	}
	if !unitInterval(policy.ElevatedRiskMinimumScore) || !unitInterval(policy.HighRiskMinimumScore) ||
		policy.HighRiskMinimumScore <= policy.ElevatedRiskMinimumScore {
		return fmt.Errorf("%w: risk score thresholds", ErrInvalidPolicy)
	}
	if !positiveFinite(policy.ElevatedHorizontalRadiusRatioMaximum) || !positiveFinite(policy.HighHorizontalRadiusRatioMaximum) ||
		policy.HighHorizontalRadiusRatioMaximum >= policy.ElevatedHorizontalRadiusRatioMaximum ||
		!positiveFinite(policy.ElevatedVerticalRadiusRatioMaximum) || !positiveFinite(policy.HighVerticalRadiusRatioMaximum) ||
		policy.HighVerticalRadiusRatioMaximum >= policy.ElevatedVerticalRadiusRatioMaximum {
		return fmt.Errorf("%w: radius ratio thresholds", ErrInvalidPolicy)
	}
	if !nonNegativeFinite(policy.ElevatedClosingRateMinimumMetersPerSecond) ||
		!positiveFinite(policy.HighClosingRateMinimumMetersPerSecond) ||
		policy.HighClosingRateMinimumMetersPerSecond <= policy.ElevatedClosingRateMinimumMetersPerSecond {
		return fmt.Errorf("%w: closing rate thresholds", ErrInvalidPolicy)
	}
	if !unitInterval(policy.MediumConfidenceMinimumScore) || !unitInterval(policy.HighConfidenceMinimumScore) ||
		policy.HighConfidenceMinimumScore <= policy.MediumConfidenceMinimumScore {
		return fmt.Errorf("%w: confidence thresholds", ErrInvalidPolicy)
	}
	if err := validateWeightTotal([]float64{
		policy.RiskWeights.HorizontalProximity,
		policy.RiskWeights.VerticalProximity,
		policy.RiskWeights.ClosingMotion,
		policy.RiskWeights.TemporalAlignment,
		policy.RiskWeights.EvidenceConfidence,
	}); err != nil {
		return fmt.Errorf("%w: risk weights", ErrInvalidPolicy)
	}
	if err := validateWeightTotal([]float64{
		policy.ResultConfidenceWeights.ScanConfidence,
		policy.ResultConfidenceWeights.AssessmentConfidence,
		policy.ResultConfidenceWeights.EvidenceCompleteness,
	}); err != nil {
		return fmt.Errorf("%w: result confidence weights", ErrInvalidPolicy)
	}
	return nil
}

func validateWeightTotal(values []float64) error {
	total := 0.0
	for _, value := range values {
		if !nonNegativeFinite(value) {
			return fmt.Errorf("invalid weight")
		}
		total += value
	}
	if math.Abs(total-1) > 1e-9 {
		return fmt.Errorf("weight total must equal one")
	}
	return nil
}
