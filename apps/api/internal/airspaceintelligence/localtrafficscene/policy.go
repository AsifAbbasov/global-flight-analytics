package localtrafficscene

import (
	"fmt"
	"math"
)

const PolicyVersionV1 = "local-traffic-scene-policy-v1"

type ConfidenceWeights struct {
	RadiusDecisionConfidence float64
	SceneCoverage            float64
}

type Policy struct {
	Version string

	MinimumCompleteAircraftCount int
	MaximumInputObservationCount int

	MediumConfidenceMinimumScore float64
	HighConfidenceMinimumScore   float64

	ConfidenceWeights ConfidenceWeights
}

func DefaultPolicy() Policy {
	return Policy{
		Version:                      PolicyVersionV1,
		MinimumCompleteAircraftCount: 2,
		MaximumInputObservationCount: 5000,
		MediumConfidenceMinimumScore: 0.50,
		HighConfidenceMinimumScore:   0.80,
		ConfidenceWeights: ConfidenceWeights{
			RadiusDecisionConfidence: 0.75,
			SceneCoverage:            0.25,
		},
	}
}

func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 {
		return fmt.Errorf("%w: version", ErrInvalidPolicy)
	}
	if policy.MinimumCompleteAircraftCount < 2 ||
		policy.MaximumInputObservationCount < policy.MinimumCompleteAircraftCount {
		return fmt.Errorf("%w: observation count bounds", ErrInvalidPolicy)
	}
	if !unitInterval(policy.MediumConfidenceMinimumScore) ||
		!unitInterval(policy.HighConfidenceMinimumScore) ||
		policy.HighConfidenceMinimumScore <= policy.MediumConfidenceMinimumScore {
		return fmt.Errorf("%w: confidence thresholds", ErrInvalidPolicy)
	}
	weightTotal := policy.ConfidenceWeights.RadiusDecisionConfidence +
		policy.ConfidenceWeights.SceneCoverage
	if !unitInterval(policy.ConfidenceWeights.RadiusDecisionConfidence) ||
		!unitInterval(policy.ConfidenceWeights.SceneCoverage) ||
		math.Abs(weightTotal-1) > 1e-9 {
		return fmt.Errorf("%w: confidence weights", ErrInvalidPolicy)
	}
	return nil
}
