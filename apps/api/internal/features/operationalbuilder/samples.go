package operationalbuilder

import (
	"context"
	"math"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type sampleCollection struct {
	altitudes                []float64
	velocities               []float64
	absoluteVerticalRates    []float64
	headings                 []float64
	groundObservationCount   int
	invalidAltitudeCount     int
	geometricFallbackCount   int
	invalidVelocityCount     int
	invalidVerticalRateCount int
	invalidHeadingCount      int
	normalizedHeadingCount   int
	nonMonotonicTimeCount    int
}

func collectSamples(
	ctx context.Context,
	points []trajectory.TrackPoint4D,
) (sampleCollection, error) {
	collection := sampleCollection{
		altitudes:             make([]float64, 0, len(points)),
		velocities:            make([]float64, 0, len(points)),
		absoluteVerticalRates: make([]float64, 0, len(points)),
		headings:              make([]float64, 0, len(points)),
	}

	var previousObservedAt time.Time
	for index, point := range points {
		if index%1024 == 0 {
			if err := ctx.Err(); err != nil {
				return sampleCollection{}, err
			}
		}

		if point.OnGround {
			collection.groundObservationCount++
		}

		if !point.ObservedAt.IsZero() {
			observedAt := point.ObservedAt.UTC()
			if !previousObservedAt.IsZero() &&
				observedAt.Before(previousObservedAt) {
				collection.nonMonotonicTimeCount++
			}
			previousObservedAt = observedAt
		}

		altitude, altitudeUsable, geometricFallback, invalidAltitude :=
			resolveAltitude(point)
		if invalidAltitude {
			collection.invalidAltitudeCount++
		}
		if altitudeUsable {
			collection.altitudes = append(
				collection.altitudes,
				altitude,
			)
			if geometricFallback {
				collection.geometricFallbackCount++
			}
		}

		if finite(point.VelocityMPS) &&
			point.VelocityMPS >= 0 {
			collection.velocities = append(
				collection.velocities,
				point.VelocityMPS,
			)
		} else {
			collection.invalidVelocityCount++
		}

		if finite(point.VerticalRateMPS) {
			collection.absoluteVerticalRates = append(
				collection.absoluteVerticalRates,
				math.Abs(point.VerticalRateMPS),
			)
		} else {
			collection.invalidVerticalRateCount++
		}

		if finite(point.HeadingDegrees) {
			heading := normalizeHeading(
				point.HeadingDegrees,
			)
			collection.headings = append(
				collection.headings,
				heading,
			)
			if point.HeadingDegrees < 0 ||
				point.HeadingDegrees >= 360 {
				collection.normalizedHeadingCount++
			}
		} else {
			collection.invalidHeadingCount++
		}
	}

	if err := ctx.Err(); err != nil {
		return sampleCollection{}, err
	}

	return collection, nil
}

func resolveAltitude(
	point trajectory.TrackPoint4D,
) (float64, bool, bool, bool) {
	barometricValue, barometricUsable, barometricInvalid :=
		altitudeValue(
			point.BarometricAltitudeM,
			point.BarometricAltitudeStatus,
		)
	if barometricUsable {
		return barometricValue, true, false, false
	}

	geometricValue, geometricUsable, geometricInvalid :=
		altitudeValue(
			point.GeometricAltitudeM,
			point.GeometricAltitudeStatus,
		)
	if geometricUsable {
		return geometricValue,
			true,
			true,
			barometricInvalid
	}

	return 0,
		false,
		false,
		barometricInvalid || geometricInvalid
}

func altitudeValue(
	value float64,
	status flightstate.AltitudeStatus,
) (float64, bool, bool) {
	resolvedStatus := flightstate.ResolveAltitudeStatus(
		value,
		status,
	)

	switch resolvedStatus {
	case flightstate.AltitudeStatusObserved:
		if !finite(value) {
			return 0, false, true
		}

		return value, true, false
	case flightstate.AltitudeStatusGround:
		return 0, true, false
	case flightstate.AltitudeStatusUnknown,
		flightstate.AltitudeStatusUnavailable:
		return 0, false, false
	case flightstate.AltitudeStatusInvalid:
		return 0, false, true
	default:
		return 0, false, true
	}
}

func (collection sampleCollection) limitations() (
	result []flightfeatures.FeatureLimitation,
) {
	if len(collection.altitudes) == 0 {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code:    "operational_altitude_unavailable",
				Message: "No usable barometric or geometric altitude observation was available.",
			},
		)
	}
	if collection.invalidAltitudeCount > 0 {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code:    "operational_invalid_altitude_observations",
				Message: "One or more altitude observations had invalid values or unsupported statuses and were excluded.",
			},
		)
	}
	if collection.geometricFallbackCount > 0 {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code:    "operational_geometric_altitude_fallback",
				Message: "Geometric altitude was used where barometric altitude was unavailable or unusable.",
			},
		)
	}

	if len(collection.velocities) == 0 {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code:    "operational_velocity_unavailable",
				Message: "No finite non-negative ground velocity observation was available.",
			},
		)
	}
	if collection.invalidVelocityCount > 0 {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code:    "operational_invalid_velocity_observations",
				Message: "One or more velocity observations were non-finite or negative and were excluded.",
			},
		)
	}

	if len(collection.absoluteVerticalRates) == 0 {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code:    "operational_vertical_rate_unavailable",
				Message: "No finite vertical-rate observation was available.",
			},
		)
	}
	if collection.invalidVerticalRateCount > 0 {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code:    "operational_invalid_vertical_rate_observations",
				Message: "One or more vertical-rate observations were non-finite and were excluded.",
			},
		)
	}

	if len(collection.headings) == 0 {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code:    "operational_heading_unavailable",
				Message: "No finite heading observation was available.",
			},
		)
	}
	if collection.invalidHeadingCount > 0 {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code:    "operational_invalid_heading_observations",
				Message: "One or more heading observations were non-finite and were excluded.",
			},
		)
	}
	if collection.normalizedHeadingCount > 0 {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code:    "operational_heading_normalized",
				Message: "One or more finite heading observations were normalized into the canonical zero-to-360-degree range.",
			},
		)
	}
	if collection.nonMonotonicTimeCount > 0 {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code:    "operational_point_order_nonmonotonic",
				Message: "One or more non-zero point timestamps decrease in input order; operational sequence metrics preserve the supplied point order.",
			},
		)
	}

	return result
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}
