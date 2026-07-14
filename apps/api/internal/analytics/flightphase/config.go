package flightphase

import (
	"errors"
	"fmt"
	"math"
)

const AlgorithmVersion = "basic-flight-phase-v1"

var (
	ErrGroundMaximumSpeedInvalid = errors.New(
		"ground maximum speed must be finite and non-negative",
	)
	ErrGroundMaximumAltitudeInvalid = errors.New(
		"ground maximum altitude must be finite and non-negative",
	)
	ErrTakeoffMaximumAltitudeInvalid = errors.New(
		"takeoff maximum altitude must be finite and non-negative",
	)
	ErrLandingMaximumAltitudeInvalid = errors.New(
		"landing maximum altitude must be finite and non-negative",
	)
	ErrCruiseMinimumAltitudeInvalid = errors.New(
		"cruise minimum altitude must be finite and positive",
	)
	ErrClimbMinimumVerticalRateInvalid = errors.New(
		"climb minimum vertical rate must be finite and positive",
	)
	ErrDescentMaximumVerticalRateInvalid = errors.New(
		"descent maximum vertical rate must be finite and negative",
	)
	ErrCruiseMaximumVerticalRateInvalid = errors.New(
		"cruise maximum absolute vertical rate must be finite and non-negative",
	)
	ErrAltitudeThresholdOrderInvalid = errors.New(
		"cruise minimum altitude must exceed takeoff and landing maximum altitudes",
	)
	ErrVerticalRateThresholdOrderInvalid = errors.New(
		"cruise maximum absolute vertical rate must be below climb and descent thresholds",
	)
)

type Config struct {
	GroundMaximumSpeedMPS                float64
	GroundMaximumAltitudeM               float64
	TakeoffMaximumAltitudeM              float64
	LandingMaximumAltitudeM              float64
	CruiseMinimumAltitudeM               float64
	ClimbMinimumVerticalRateMPS          float64
	DescentMaximumVerticalRateMPS        float64
	CruiseMaximumAbsoluteVerticalRateMPS float64
}

func DefaultConfig() Config {
	return Config{
		GroundMaximumSpeedMPS:                35,
		GroundMaximumAltitudeM:               150,
		TakeoffMaximumAltitudeM:              1200,
		LandingMaximumAltitudeM:              1200,
		CruiseMinimumAltitudeM:               1500,
		ClimbMinimumVerticalRateMPS:          1,
		DescentMaximumVerticalRateMPS:        -1,
		CruiseMaximumAbsoluteVerticalRateMPS: 0.75,
	}
}

func (config Config) Validate() error {
	if !finite(config.GroundMaximumSpeedMPS) ||
		config.GroundMaximumSpeedMPS < 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrGroundMaximumSpeedInvalid,
			config.GroundMaximumSpeedMPS,
		)
	}
	if !finite(config.GroundMaximumAltitudeM) ||
		config.GroundMaximumAltitudeM < 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrGroundMaximumAltitudeInvalid,
			config.GroundMaximumAltitudeM,
		)
	}
	if !finite(config.TakeoffMaximumAltitudeM) ||
		config.TakeoffMaximumAltitudeM < 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrTakeoffMaximumAltitudeInvalid,
			config.TakeoffMaximumAltitudeM,
		)
	}
	if !finite(config.LandingMaximumAltitudeM) ||
		config.LandingMaximumAltitudeM < 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrLandingMaximumAltitudeInvalid,
			config.LandingMaximumAltitudeM,
		)
	}
	if !finite(config.CruiseMinimumAltitudeM) ||
		config.CruiseMinimumAltitudeM <= 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrCruiseMinimumAltitudeInvalid,
			config.CruiseMinimumAltitudeM,
		)
	}
	if !finite(config.ClimbMinimumVerticalRateMPS) ||
		config.ClimbMinimumVerticalRateMPS <= 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrClimbMinimumVerticalRateInvalid,
			config.ClimbMinimumVerticalRateMPS,
		)
	}
	if !finite(config.DescentMaximumVerticalRateMPS) ||
		config.DescentMaximumVerticalRateMPS >= 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrDescentMaximumVerticalRateInvalid,
			config.DescentMaximumVerticalRateMPS,
		)
	}
	if !finite(config.CruiseMaximumAbsoluteVerticalRateMPS) ||
		config.CruiseMaximumAbsoluteVerticalRateMPS < 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrCruiseMaximumVerticalRateInvalid,
			config.CruiseMaximumAbsoluteVerticalRateMPS,
		)
	}

	if config.CruiseMinimumAltitudeM <=
		config.TakeoffMaximumAltitudeM ||
		config.CruiseMinimumAltitudeM <=
			config.LandingMaximumAltitudeM {
		return fmt.Errorf(
			"%w: cruise=%f takeoff=%f landing=%f",
			ErrAltitudeThresholdOrderInvalid,
			config.CruiseMinimumAltitudeM,
			config.TakeoffMaximumAltitudeM,
			config.LandingMaximumAltitudeM,
		)
	}

	if config.CruiseMaximumAbsoluteVerticalRateMPS >=
		config.ClimbMinimumVerticalRateMPS ||
		config.CruiseMaximumAbsoluteVerticalRateMPS >=
			math.Abs(config.DescentMaximumVerticalRateMPS) {
		return fmt.Errorf(
			"%w: cruise=%f climb=%f descent=%f",
			ErrVerticalRateThresholdOrderInvalid,
			config.CruiseMaximumAbsoluteVerticalRateMPS,
			config.ClimbMinimumVerticalRateMPS,
			config.DescentMaximumVerticalRateMPS,
		)
	}

	return nil
}
