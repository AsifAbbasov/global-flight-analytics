package projectionbaseline

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

var (
	ErrHorizonPlannerRequired = errors.New(
		"projection horizon planner is required",
	)
	ErrEligibilityEvaluatorRequired = errors.New(
		"projection eligibility evaluator is required",
	)
	ErrHorizontalUncertaintyInvalid = errors.New(
		"horizontal uncertainty configuration is invalid",
	)
	ErrVerticalUncertaintyInvalid = errors.New(
		"vertical uncertainty configuration is invalid",
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

type EligibilityEvaluator interface {
	Evaluate(
		trajectory.FlightTrajectory,
		time.Time,
	) trajectoryeligibility.Evaluation
}

type Config struct {
	HorizonPlanner       HorizonPlanner
	EligibilityEvaluator EligibilityEvaluator

	InitialHorizontalUncertaintyM  float64
	HorizontalUncertaintyGrowthMPS float64
	InitialVerticalUncertaintyM    float64
	VerticalUncertaintyGrowthMPS   float64

	MaximumConfidenceLoss float64

	MediumConfidenceMinimum float64
	HighConfidenceMinimum   float64

	AllowOnGround bool
}

func (config Config) Validate() error {
	if config.HorizonPlanner == nil {
		return ErrHorizonPlannerRequired
	}
	if config.EligibilityEvaluator == nil {
		return ErrEligibilityEvaluatorRequired
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

func positiveFinite(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value > 0
}

func nonNegativeFinite(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0
}

func unitInterval(
	value float64,
) bool {
	return nonNegativeFinite(value) &&
		value <= 1
}
