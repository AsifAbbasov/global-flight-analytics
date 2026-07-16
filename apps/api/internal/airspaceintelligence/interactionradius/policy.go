package interactionradius

import (
	"fmt"
	"math"
	"time"
)

const PolicyVersionV1 = "interaction-radius-policy-v1"

type Weights struct {
	DataQuality        float64
	TemporalFreshness  float64
	MotionPlausibility float64
	VerticalEvidence   float64
}

type Policy struct {
	Version string

	MinimumHorizontalRadiusKilometers float64
	BaseHorizontalRadiusKilometers    float64
	MaximumHorizontalRadiusKilometers float64
	HorizontalLookaheadDuration       time.Duration
	QualityUncertaintyFraction        float64

	MinimumVerticalRadiusMeters float64
	BaseVerticalRadiusMeters    float64
	MaximumVerticalRadiusMeters float64
	VerticalLookaheadDuration   time.Duration

	MaximumObservationAge     time.Duration
	MaximumPairTimeDifference time.Duration

	MinimumUsableQuality  float64
	MinimumAllowedQuality float64

	LowSpeedMaximumMetersPerSecond       float64
	HighSpeedMinimumMetersPerSecond      float64
	VerticalChangeMinimumMetersPerSecond float64

	MediumConfidenceMinimumScore float64
	HighConfidenceMinimumScore   float64

	Weights Weights
}

func DefaultPolicy() Policy {
	return Policy{
		Version: PolicyVersionV1,

		MinimumHorizontalRadiusKilometers: 10,
		BaseHorizontalRadiusKilometers:    20,
		MaximumHorizontalRadiusKilometers: 80,
		HorizontalLookaheadDuration:       2 * time.Minute,
		QualityUncertaintyFraction:        0.50,

		MinimumVerticalRadiusMeters: 500,
		BaseVerticalRadiusMeters:    750,
		MaximumVerticalRadiusMeters: 3000,
		VerticalLookaheadDuration:   1 * time.Minute,

		MaximumObservationAge:     90 * time.Second,
		MaximumPairTimeDifference: 30 * time.Second,

		MinimumUsableQuality:  0.35,
		MinimumAllowedQuality: 0.65,

		LowSpeedMaximumMetersPerSecond:       90,
		HighSpeedMinimumMetersPerSecond:      240,
		VerticalChangeMinimumMetersPerSecond: 3,

		MediumConfidenceMinimumScore: 0.50,
		HighConfidenceMinimumScore:   0.80,

		Weights: Weights{
			DataQuality:        0.40,
			TemporalFreshness:  0.35,
			MotionPlausibility: 0.15,
			VerticalEvidence:   0.10,
		},
	}
}

func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 {
		return fmt.Errorf("%w: version", ErrInvalidPolicy)
	}
	if !positiveFinite(policy.MinimumHorizontalRadiusKilometers) ||
		!positiveFinite(policy.BaseHorizontalRadiusKilometers) ||
		!positiveFinite(policy.MaximumHorizontalRadiusKilometers) ||
		policy.MinimumHorizontalRadiusKilometers > policy.BaseHorizontalRadiusKilometers ||
		policy.BaseHorizontalRadiusKilometers > policy.MaximumHorizontalRadiusKilometers ||
		policy.HorizontalLookaheadDuration <= 0 ||
		!unitInterval(policy.QualityUncertaintyFraction) {
		return fmt.Errorf("%w: horizontal bounds", ErrInvalidPolicy)
	}
	if !positiveFinite(policy.MinimumVerticalRadiusMeters) ||
		!positiveFinite(policy.BaseVerticalRadiusMeters) ||
		!positiveFinite(policy.MaximumVerticalRadiusMeters) ||
		policy.MinimumVerticalRadiusMeters > policy.BaseVerticalRadiusMeters ||
		policy.BaseVerticalRadiusMeters > policy.MaximumVerticalRadiusMeters ||
		policy.VerticalLookaheadDuration <= 0 {
		return fmt.Errorf("%w: vertical bounds", ErrInvalidPolicy)
	}
	if policy.MaximumObservationAge <= 0 ||
		policy.MaximumPairTimeDifference <= 0 ||
		policy.MaximumPairTimeDifference > policy.MaximumObservationAge {
		return fmt.Errorf("%w: temporal bounds", ErrInvalidPolicy)
	}
	if !unitInterval(policy.MinimumUsableQuality) ||
		!unitInterval(policy.MinimumAllowedQuality) ||
		policy.MinimumAllowedQuality < policy.MinimumUsableQuality {
		return fmt.Errorf("%w: quality bounds", ErrInvalidPolicy)
	}
	if !nonNegativeFinite(policy.LowSpeedMaximumMetersPerSecond) ||
		!positiveFinite(policy.HighSpeedMinimumMetersPerSecond) ||
		policy.HighSpeedMinimumMetersPerSecond <= policy.LowSpeedMaximumMetersPerSecond ||
		!positiveFinite(policy.VerticalChangeMinimumMetersPerSecond) {
		return fmt.Errorf("%w: motion thresholds", ErrInvalidPolicy)
	}
	if !unitInterval(policy.MediumConfidenceMinimumScore) ||
		!unitInterval(policy.HighConfidenceMinimumScore) ||
		policy.HighConfidenceMinimumScore <= policy.MediumConfidenceMinimumScore {
		return fmt.Errorf("%w: confidence thresholds", ErrInvalidPolicy)
	}
	weightTotal := policy.Weights.DataQuality +
		policy.Weights.TemporalFreshness +
		policy.Weights.MotionPlausibility +
		policy.Weights.VerticalEvidence
	if !nonNegativeFinite(policy.Weights.DataQuality) ||
		!nonNegativeFinite(policy.Weights.TemporalFreshness) ||
		!nonNegativeFinite(policy.Weights.MotionPlausibility) ||
		!nonNegativeFinite(policy.Weights.VerticalEvidence) ||
		math.Abs(weightTotal-1) > 1e-9 {
		return fmt.Errorf("%w: weights", ErrInvalidPolicy)
	}
	return nil
}
