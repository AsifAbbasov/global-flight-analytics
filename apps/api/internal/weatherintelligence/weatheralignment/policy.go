package weatheralignment

import (
	"fmt"
	"time"
)

const PolicyVersionV1 = "weather-trajectory-alignment-policy-v1"

type Weights struct {
	Horizontal float64
	Temporal   float64
	Vertical   float64
}

type Policy struct {
	Version string

	MaximumHorizontalDistanceKilometers float64
	MaximumTemporalDistance             time.Duration
	MaximumVerticalDistanceMeters       float64

	MinimumMatchScore float64

	Weights Weights
}

func DefaultPolicy() Policy {
	return Policy{
		Version: PolicyVersionV1,

		MaximumHorizontalDistanceKilometers: 75,
		MaximumTemporalDistance:             90 * time.Minute,
		MaximumVerticalDistanceMeters:       1500,

		MinimumMatchScore: 0.35,

		Weights: Weights{
			Horizontal: 0.45,
			Temporal:   0.35,
			Vertical:   0.20,
		},
	}
}

func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 {
		return fmt.Errorf("weather alignment policy version is invalid")
	}
	if !finite(policy.MaximumHorizontalDistanceKilometers) ||
		policy.MaximumHorizontalDistanceKilometers <= 0 ||
		policy.MaximumTemporalDistance <= 0 ||
		!finite(policy.MaximumVerticalDistanceMeters) ||
		policy.MaximumVerticalDistanceMeters <= 0 ||
		!unitInterval(policy.MinimumMatchScore) {
		return fmt.Errorf("weather alignment policy thresholds are invalid")
	}

	weightTotal := policy.Weights.Horizontal + policy.Weights.Temporal + policy.Weights.Vertical
	if !finite(policy.Weights.Horizontal) ||
		!finite(policy.Weights.Temporal) ||
		!finite(policy.Weights.Vertical) ||
		policy.Weights.Horizontal < 0 ||
		policy.Weights.Temporal < 0 ||
		policy.Weights.Vertical < 0 ||
		absolute(weightTotal-1) > 1e-9 {
		return fmt.Errorf("weather alignment policy weights are invalid")
	}
	return nil
}

func (policy Policy) components(horizontalScore, temporalScore, verticalScore float64) []Component {
	return []Component{
		{Name: ComponentHorizontal, Score: clampUnit(horizontalScore), Weight: policy.Weights.Horizontal},
		{Name: ComponentTemporal, Score: clampUnit(temporalScore), Weight: policy.Weights.Temporal},
		{Name: ComponentVertical, Score: clampUnit(verticalScore), Weight: policy.Weights.Vertical},
	}
}

func weightedScore(components []Component) float64 {
	total := 0.0
	for _, component := range components {
		total += component.Score * component.Weight
	}
	return clampUnit(total)
}
