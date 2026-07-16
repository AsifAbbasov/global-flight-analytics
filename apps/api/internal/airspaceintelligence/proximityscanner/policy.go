package proximityscanner

import (
	"fmt"
	"math"
)

const PolicyVersionV1 = "multi-aircraft-proximity-scanner-policy-v1"

type CandidateConfidenceWeights struct {
	PreparedEvidenceQuality  float64
	RadiusDecisionConfidence float64
	TemporalProximity        float64
	VerticalEvidence         float64
}

type ResultConfidenceWeights struct {
	SceneConfidence            float64
	MeanRadiusConfidence       float64
	PairEvaluationCompleteness float64
}

type Policy struct {
	Version string

	MaximumAircraftCount int
	MaximumPairCount     int

	ConvergingClosingRateMetersPerSecond float64
	DivergingOpeningRateMetersPerSecond  float64
	ParallelHeadingToleranceDegrees      float64

	MediumConfidenceMinimumScore float64
	HighConfidenceMinimumScore   float64

	CandidateConfidenceWeights CandidateConfidenceWeights
	ResultConfidenceWeights    ResultConfidenceWeights
}

func DefaultPolicy() Policy {
	return Policy{
		Version:                              PolicyVersionV1,
		MaximumAircraftCount:                 1000,
		MaximumPairCount:                     499500,
		ConvergingClosingRateMetersPerSecond: 5,
		DivergingOpeningRateMetersPerSecond:  5,
		ParallelHeadingToleranceDegrees:      15,
		MediumConfidenceMinimumScore:         0.50,
		HighConfidenceMinimumScore:           0.80,
		CandidateConfidenceWeights: CandidateConfidenceWeights{
			PreparedEvidenceQuality:  0.35,
			RadiusDecisionConfidence: 0.30,
			TemporalProximity:        0.20,
			VerticalEvidence:         0.15,
		},
		ResultConfidenceWeights: ResultConfidenceWeights{
			SceneConfidence:            0.45,
			MeanRadiusConfidence:       0.35,
			PairEvaluationCompleteness: 0.20,
		},
	}
}

func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 {
		return fmt.Errorf("%w: version", ErrInvalidPolicy)
	}
	if policy.MaximumAircraftCount < 2 || policy.MaximumPairCount < 1 {
		return fmt.Errorf("%w: capacity bounds", ErrInvalidPolicy)
	}
	maximumPossiblePairs := policy.MaximumAircraftCount * (policy.MaximumAircraftCount - 1) / 2
	if policy.MaximumPairCount > maximumPossiblePairs {
		return fmt.Errorf("%w: pair count exceeds aircraft capacity", ErrInvalidPolicy)
	}
	if !positiveFinite(policy.ConvergingClosingRateMetersPerSecond) ||
		!positiveFinite(policy.DivergingOpeningRateMetersPerSecond) ||
		!positiveFinite(policy.ParallelHeadingToleranceDegrees) ||
		policy.ParallelHeadingToleranceDegrees > 180 {
		return fmt.Errorf("%w: interaction classification thresholds", ErrInvalidPolicy)
	}
	if !unitInterval(policy.MediumConfidenceMinimumScore) ||
		!unitInterval(policy.HighConfidenceMinimumScore) ||
		policy.HighConfidenceMinimumScore <= policy.MediumConfidenceMinimumScore {
		return fmt.Errorf("%w: confidence thresholds", ErrInvalidPolicy)
	}
	candidateWeightTotal := policy.CandidateConfidenceWeights.PreparedEvidenceQuality +
		policy.CandidateConfidenceWeights.RadiusDecisionConfidence +
		policy.CandidateConfidenceWeights.TemporalProximity +
		policy.CandidateConfidenceWeights.VerticalEvidence
	if !nonNegativeFinite(policy.CandidateConfidenceWeights.PreparedEvidenceQuality) ||
		!nonNegativeFinite(policy.CandidateConfidenceWeights.RadiusDecisionConfidence) ||
		!nonNegativeFinite(policy.CandidateConfidenceWeights.TemporalProximity) ||
		!nonNegativeFinite(policy.CandidateConfidenceWeights.VerticalEvidence) ||
		math.Abs(candidateWeightTotal-1) > 1e-9 {
		return fmt.Errorf("%w: candidate confidence weights", ErrInvalidPolicy)
	}
	resultWeightTotal := policy.ResultConfidenceWeights.SceneConfidence +
		policy.ResultConfidenceWeights.MeanRadiusConfidence +
		policy.ResultConfidenceWeights.PairEvaluationCompleteness
	if !nonNegativeFinite(policy.ResultConfidenceWeights.SceneConfidence) ||
		!nonNegativeFinite(policy.ResultConfidenceWeights.MeanRadiusConfidence) ||
		!nonNegativeFinite(policy.ResultConfidenceWeights.PairEvaluationCompleteness) ||
		math.Abs(resultWeightTotal-1) > 1e-9 {
		return fmt.Errorf("%w: result confidence weights", ErrInvalidPolicy)
	}
	return nil
}
