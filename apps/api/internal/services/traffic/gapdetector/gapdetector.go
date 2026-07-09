package gapdetector

import (
	"fmt"
	"math"
	"time"

	aviationconstraints "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/constraints"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const metersPerKilometer = 1000.0

type Config struct {
	MaxTimeGap        time.Duration
	MaxGroundSpeedMPS float64
}

func (config Config) Validate() error {
	if config.MaxTimeGap < 0 {
		return fmt.Errorf(
			"max time gap must be non-negative, got %s",
			config.MaxTimeGap,
		)
	}

	if math.IsNaN(config.MaxGroundSpeedMPS) {
		return fmt.Errorf(
			"max ground speed must be a finite non-negative value",
		)
	}

	if math.IsInf(config.MaxGroundSpeedMPS, 0) {
		return fmt.Errorf(
			"max ground speed must be a finite non-negative value",
		)
	}

	if config.MaxGroundSpeedMPS < 0 {
		return fmt.Errorf(
			"max ground speed must be non-negative, got %f",
			config.MaxGroundSpeedMPS,
		)
	}

	return nil
}

type DetectionResult struct {
	HasGap            bool
	Reason            trajectory.CoverageGapReason
	Duration          time.Duration
	DistanceKm        float64
	EstimatedSpeedMPS float64
}

func DefaultConfig() Config {
	return Config{}
}

func Detect(
	previous flightstate.FlightState,
	next flightstate.FlightState,
	config Config,
) DetectionResult {
	duration := next.ObservedAt.Sub(previous.ObservedAt)

	distanceKm := HaversineDistanceKm(
		previous.Latitude,
		previous.Longitude,
		next.Latitude,
		next.Longitude,
	)

	result := DetectionResult{
		Duration:   duration,
		DistanceKm: distanceKm,
		Reason:     trajectory.CoverageGapReasonUnknown,
	}

	if duration <= 0 {
		result.HasGap = true
		result.Reason = trajectory.CoverageGapReasonUnknown
		return result
	}

	result.EstimatedSpeedMPS =
		distanceKm * metersPerKilometer / duration.Seconds()

	if config.MaxTimeGap > 0 &&
		duration > config.MaxTimeGap {
		result.HasGap = true
		result.Reason = trajectory.CoverageGapReasonTimeGap
		return result
	}

	if config.MaxGroundSpeedMPS > 0 &&
		result.EstimatedSpeedMPS > config.MaxGroundSpeedMPS {
		result.HasGap = true
		result.Reason = trajectory.CoverageGapReasonMovementJump
		return result
	}

	return result
}

func HaversineDistanceKm(
	fromLatitude float64,
	fromLongitude float64,
	toLatitude float64,
	toLongitude float64,
) float64 {
	fromLatRad := degreesToRadians(fromLatitude)
	fromLonRad := degreesToRadians(fromLongitude)
	toLatRad := degreesToRadians(toLatitude)
	toLonRad := degreesToRadians(toLongitude)

	latDelta := toLatRad - fromLatRad
	lonDelta := toLonRad - fromLonRad

	a := math.Sin(latDelta/2)*math.Sin(latDelta/2) +
		math.Cos(fromLatRad)*math.Cos(toLatRad)*
			math.Sin(lonDelta/2)*math.Sin(lonDelta/2)

	c := 2 * math.Atan2(
		math.Sqrt(a),
		math.Sqrt(1-a),
	)

	return aviationconstraints.EarthRadiusKilometers * c
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}
