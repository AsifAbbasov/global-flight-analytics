package airspaceregionanalytics

import (
	"fmt"
	"math"
	"time"
)

const PolicyVersionV1 = "airspace-region-analytics-policy-v1"

type ComplexityWeights struct {
	Density           float64
	PairInteraction   float64
	DeterminateRisk   float64
	HeadingDispersion float64
	SpeedVariability  float64
	AltitudeMixing    float64
}

type ConfidenceWeights struct {
	SceneConfidence  float64
	ScanConfidence   float64
	RiskConfidence   float64
	DataQuality      float64
	TemporalCoverage float64
}

type Policy struct {
	Version                              string
	TimeBucketDuration                   time.Duration
	LatitudeCellDegrees                  float64
	LongitudeCellDegrees                 float64
	AltitudeBandMeters                   float64
	MaximumSnapshots                     int
	MaximumAircraftObservations          int
	DenseAircraftCount                   int
	MixedAltitudeBandCount               int
	SpeedVariabilityScaleMetersPerSecond float64
	ModerateComplexityMinimumScore       float64
	HighComplexityMinimumScore           float64
	SevereComplexityMinimumScore         float64
	MediumConfidenceMinimumScore         float64
	HighConfidenceMinimumScore           float64
	OccupancyTrendChangeThreshold        float64
	ComplexityWeights                    ComplexityWeights
	ConfidenceWeights                    ConfidenceWeights
}

func DefaultPolicy() Policy {
	return Policy{
		Version:                              PolicyVersionV1,
		TimeBucketDuration:                   60 * time.Second,
		LatitudeCellDegrees:                  1,
		LongitudeCellDegrees:                 1,
		AltitudeBandMeters:                   3000,
		MaximumSnapshots:                     1440,
		MaximumAircraftObservations:          250000,
		DenseAircraftCount:                   8,
		MixedAltitudeBandCount:               3,
		SpeedVariabilityScaleMetersPerSecond: 80,
		ModerateComplexityMinimumScore:       0.35,
		HighComplexityMinimumScore:           0.60,
		SevereComplexityMinimumScore:         0.80,
		MediumConfidenceMinimumScore:         0.50,
		HighConfidenceMinimumScore:           0.80,
		OccupancyTrendChangeThreshold:        0.10,
		ComplexityWeights: ComplexityWeights{
			Density:           0.24,
			PairInteraction:   0.20,
			DeterminateRisk:   0.24,
			HeadingDispersion: 0.12,
			SpeedVariability:  0.08,
			AltitudeMixing:    0.12,
		},
		ConfidenceWeights: ConfidenceWeights{
			SceneConfidence:  0.25,
			ScanConfidence:   0.20,
			RiskConfidence:   0.20,
			DataQuality:      0.20,
			TemporalCoverage: 0.15,
		},
	}
}

func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 {
		return fmt.Errorf("%w: version", ErrInvalidPolicy)
	}
	if policy.TimeBucketDuration <= 0 ||
		!positiveFinite(policy.LatitudeCellDegrees) || policy.LatitudeCellDegrees > 180 ||
		!positiveFinite(policy.LongitudeCellDegrees) || policy.LongitudeCellDegrees > 360 ||
		!positiveFinite(policy.AltitudeBandMeters) {
		return fmt.Errorf("%w: occupancy grid", ErrInvalidPolicy)
	}
	if policy.MaximumSnapshots <= 0 || policy.MaximumAircraftObservations <= 0 ||
		policy.DenseAircraftCount <= 0 || policy.MixedAltitudeBandCount < 2 ||
		!positiveFinite(policy.SpeedVariabilityScaleMetersPerSecond) {
		return fmt.Errorf("%w: capacity and normalization", ErrInvalidPolicy)
	}
	if !orderedUnitThresholds(
		policy.ModerateComplexityMinimumScore,
		policy.HighComplexityMinimumScore,
		policy.SevereComplexityMinimumScore,
	) ||
		!unitInterval(policy.MediumConfidenceMinimumScore) ||
		!unitInterval(policy.HighConfidenceMinimumScore) ||
		policy.HighConfidenceMinimumScore <= policy.MediumConfidenceMinimumScore ||
		!unitInterval(policy.OccupancyTrendChangeThreshold) {
		return fmt.Errorf("%w: thresholds", ErrInvalidPolicy)
	}
	complexityTotal := policy.ComplexityWeights.Density +
		policy.ComplexityWeights.PairInteraction +
		policy.ComplexityWeights.DeterminateRisk +
		policy.ComplexityWeights.HeadingDispersion +
		policy.ComplexityWeights.SpeedVariability +
		policy.ComplexityWeights.AltitudeMixing
	if !validWeightTotal(complexityTotal, []float64{
		policy.ComplexityWeights.Density,
		policy.ComplexityWeights.PairInteraction,
		policy.ComplexityWeights.DeterminateRisk,
		policy.ComplexityWeights.HeadingDispersion,
		policy.ComplexityWeights.SpeedVariability,
		policy.ComplexityWeights.AltitudeMixing,
	}) {
		return fmt.Errorf("%w: complexity weights", ErrInvalidPolicy)
	}
	confidenceTotal := policy.ConfidenceWeights.SceneConfidence +
		policy.ConfidenceWeights.ScanConfidence +
		policy.ConfidenceWeights.RiskConfidence +
		policy.ConfidenceWeights.DataQuality +
		policy.ConfidenceWeights.TemporalCoverage
	if !validWeightTotal(confidenceTotal, []float64{
		policy.ConfidenceWeights.SceneConfidence,
		policy.ConfidenceWeights.ScanConfidence,
		policy.ConfidenceWeights.RiskConfidence,
		policy.ConfidenceWeights.DataQuality,
		policy.ConfidenceWeights.TemporalCoverage,
	}) {
		return fmt.Errorf("%w: confidence weights", ErrInvalidPolicy)
	}
	return nil
}

func orderedUnitThresholds(values ...float64) bool {
	if len(values) == 0 {
		return false
	}
	previous := -1.0
	for _, value := range values {
		if !unitInterval(value) || value <= previous {
			return false
		}
		previous = value
	}
	return true
}

func validWeightTotal(total float64, values []float64) bool {
	for _, value := range values {
		if !nonNegativeFinite(value) {
			return false
		}
	}
	return math.Abs(total-1) <= 1e-9
}
