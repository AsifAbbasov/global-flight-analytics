package projectioncontinuation

import (
	"errors"
	"fmt"
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

var (
	ErrHorizonPlannerRequired = errors.New(
		"projection horizon planner is required",
	)
	ErrNeighborSelectorRequired = errors.New(
		"historical neighbor selector is required",
	)
	ErrPatternConfidenceEvaluatorRequired = errors.New(
		"pattern confidence evaluator is required",
	)
	ErrFallbackProjectorRequired = errors.New(
		"kinematic fallback projector is required",
	)
	ErrMinimumPointSupportInvalid = errors.New(
		"minimum point support must be greater than zero",
	)
	ErrMinimumAltitudeSupportInvalid = errors.New(
		"minimum altitude support must be between one and minimum point support",
	)
	ErrHorizontalUncertaintyInvalid = errors.New(
		"horizontal uncertainty policy is invalid",
	)
	ErrVerticalUncertaintyInvalid = errors.New(
		"vertical uncertainty policy is invalid",
	)
	ErrNeighborSpreadMultiplierInvalid = errors.New(
		"neighbor spread multiplier must be finite and greater than zero",
	)
	ErrMaximumConfidenceLossInvalid = errors.New(
		"maximum confidence loss must be finite and between zero and one",
	)
	ErrConfidenceThresholdInvalid = errors.New(
		"confidence thresholds must satisfy zero < medium <= high <= one",
	)
)

type HorizonPlanner interface {
	Build(
		projectionhorizon.Request,
	) (projectionhorizon.Plan, error)
}

type NeighborSelector interface {
	Select(
		projectionneighbors.Request,
	) (projectionneighbors.Result, error)
}

type PatternConfidenceEvaluator interface {
	Evaluate(
		projectionneighbors.Result,
	) (projectionpatternconfidence.Result, error)
}

type FallbackProjector interface {
	Project(
		projectionbaseline.Request,
	) (projectioncontract.Result, error)
}

type Config struct {
	HorizonPlanner             HorizonPlanner
	NeighborSelector           NeighborSelector
	PatternConfidenceEvaluator PatternConfidenceEvaluator
	FallbackProjector          FallbackProjector

	MinimumPointSupport    int
	MinimumAltitudeSupport int

	InitialHorizontalUncertaintyM  float64
	HorizontalUncertaintyGrowthMPS float64
	InitialVerticalUncertaintyM    float64
	VerticalUncertaintyGrowthMPS   float64
	NeighborSpreadMultiplier       float64

	MaximumConfidenceLoss float64

	MediumConfidenceMinimum float64
	HighConfidenceMinimum   float64
}

func (config Config) Validate() error {
	if config.HorizonPlanner == nil {
		return ErrHorizonPlannerRequired
	}
	if config.NeighborSelector == nil {
		return ErrNeighborSelectorRequired
	}
	if config.PatternConfidenceEvaluator == nil {
		return ErrPatternConfidenceEvaluatorRequired
	}
	if config.FallbackProjector == nil {
		return ErrFallbackProjectorRequired
	}
	if config.MinimumPointSupport < 1 {
		return fmt.Errorf(
			"%w: %d",
			ErrMinimumPointSupportInvalid,
			config.MinimumPointSupport,
		)
	}
	if config.MinimumAltitudeSupport < 1 ||
		config.MinimumAltitudeSupport >
			config.MinimumPointSupport {
		return fmt.Errorf(
			"%w: altitude=%d point=%d",
			ErrMinimumAltitudeSupportInvalid,
			config.MinimumAltitudeSupport,
			config.MinimumPointSupport,
		)
	}
	if !positiveFinite(
		config.InitialHorizontalUncertaintyM,
	) ||
		!nonNegativeFinite(
			config.HorizontalUncertaintyGrowthMPS,
		) {
		return fmt.Errorf(
			"%w: initial=%f growth=%f",
			ErrHorizontalUncertaintyInvalid,
			config.InitialHorizontalUncertaintyM,
			config.HorizontalUncertaintyGrowthMPS,
		)
	}
	if !positiveFinite(
		config.InitialVerticalUncertaintyM,
	) ||
		!nonNegativeFinite(
			config.VerticalUncertaintyGrowthMPS,
		) {
		return fmt.Errorf(
			"%w: initial=%f growth=%f",
			ErrVerticalUncertaintyInvalid,
			config.InitialVerticalUncertaintyM,
			config.VerticalUncertaintyGrowthMPS,
		)
	}
	if !positiveFinite(
		config.NeighborSpreadMultiplier,
	) {
		return fmt.Errorf(
			"%w: %f",
			ErrNeighborSpreadMultiplierInvalid,
			config.NeighborSpreadMultiplier,
		)
	}
	if !unitInterval(
		config.MaximumConfidenceLoss,
	) {
		return fmt.Errorf(
			"%w: %f",
			ErrMaximumConfidenceLossInvalid,
			config.MaximumConfidenceLoss,
		)
	}
	if !positiveFinite(
		config.MediumConfidenceMinimum,
	) ||
		!positiveFinite(
			config.HighConfidenceMinimum,
		) ||
		config.MediumConfidenceMinimum >
			config.HighConfidenceMinimum ||
		config.HighConfidenceMinimum > 1 {
		return fmt.Errorf(
			"%w: medium=%f high=%f",
			ErrConfidenceThresholdInvalid,
			config.MediumConfidenceMinimum,
			config.HighConfidenceMinimum,
		)
	}

	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}

func positiveFinite(value float64) bool {
	return finite(value) &&
		value > 0
}

func nonNegativeFinite(value float64) bool {
	return finite(value) &&
		value >= 0
}

func unitInterval(value float64) bool {
	return finite(value) &&
		value >= 0 &&
		value <= 1
}
