package forecaststability

import (
	"fmt"
	"math"
	"strings"
)

const StabilityPolicyVersionV1 = "decision-stability-policy-v1-experimental"

type StabilityWeights struct {
	Position    float64
	Uncertainty float64
	Confidence  float64
	Arrival     float64
	Decision    float64
}

type StabilityPolicy struct {
	Version string

	MinimumAlignedPointShare float64

	StableMeanHorizontalShiftKilometers      float64
	StableMaximumHorizontalShiftKilometers   float64
	MaterialMeanHorizontalShiftKilometers    float64
	MaterialMaximumHorizontalShiftKilometers float64

	StableConfidenceDelta   float64
	MaterialConfidenceDelta float64

	StableRelativeUncertaintyChange   float64
	MaterialRelativeUncertaintyChange float64

	StableArrivalShiftSeconds   float64
	MaterialArrivalShiftSeconds float64

	Weights StabilityWeights
}

func DefaultStabilityPolicy() StabilityPolicy {
	return StabilityPolicy{
		Version:                                  StabilityPolicyVersionV1,
		MinimumAlignedPointShare:                 0.80,
		StableMeanHorizontalShiftKilometers:      3,
		StableMaximumHorizontalShiftKilometers:   8,
		MaterialMeanHorizontalShiftKilometers:    20,
		MaterialMaximumHorizontalShiftKilometers: 50,
		StableConfidenceDelta:                    0.10,
		MaterialConfidenceDelta:                  0.30,
		StableRelativeUncertaintyChange:          0.25,
		MaterialRelativeUncertaintyChange:        1.00,
		StableArrivalShiftSeconds:                120,
		MaterialArrivalShiftSeconds:              600,
		Weights: StabilityWeights{
			Position:    0.35,
			Uncertainty: 0.15,
			Confidence:  0.20,
			Arrival:     0.15,
			Decision:    0.15,
		},
	}
}

func (policy StabilityPolicy) Validate() error {
	if strings.TrimSpace(policy.Version) != StabilityPolicyVersionV1 {
		return fmt.Errorf("%w: version", ErrInvalidStabilityPolicy)
	}
	if !unitInterval(policy.MinimumAlignedPointShare) ||
		!positiveFinite(policy.StableMeanHorizontalShiftKilometers) ||
		!positiveFinite(policy.StableMaximumHorizontalShiftKilometers) ||
		!positiveFinite(policy.MaterialMeanHorizontalShiftKilometers) ||
		!positiveFinite(policy.MaterialMaximumHorizontalShiftKilometers) ||
		policy.MaterialMeanHorizontalShiftKilometers <= policy.StableMeanHorizontalShiftKilometers ||
		policy.MaterialMaximumHorizontalShiftKilometers <= policy.StableMaximumHorizontalShiftKilometers ||
		!unitInterval(policy.StableConfidenceDelta) ||
		!unitInterval(policy.MaterialConfidenceDelta) ||
		policy.MaterialConfidenceDelta <= policy.StableConfidenceDelta ||
		!nonNegativeFinite(policy.StableRelativeUncertaintyChange) ||
		!nonNegativeFinite(policy.MaterialRelativeUncertaintyChange) ||
		policy.MaterialRelativeUncertaintyChange <= policy.StableRelativeUncertaintyChange ||
		!positiveFinite(policy.StableArrivalShiftSeconds) ||
		!positiveFinite(policy.MaterialArrivalShiftSeconds) ||
		policy.MaterialArrivalShiftSeconds <= policy.StableArrivalShiftSeconds {
		return fmt.Errorf("%w: thresholds", ErrInvalidStabilityPolicy)
	}
	weights := []float64{
		policy.Weights.Position,
		policy.Weights.Uncertainty,
		policy.Weights.Confidence,
		policy.Weights.Arrival,
		policy.Weights.Decision,
	}
	total := 0.0
	for _, weight := range weights {
		if !nonNegativeFinite(weight) {
			return fmt.Errorf("%w: weights", ErrInvalidStabilityPolicy)
		}
		total += weight
	}
	if math.Abs(total-1) > 1e-9 {
		return fmt.Errorf("%w: weight total", ErrInvalidStabilityPolicy)
	}
	return nil
}
