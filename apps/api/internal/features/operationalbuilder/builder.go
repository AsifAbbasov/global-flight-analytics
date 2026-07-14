package operationalbuilder

import (
	"context"
	"fmt"
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

var _ extractor.OperationalBuilder = (*Builder)(nil)

type Builder struct{}

func New() *Builder {
	return &Builder{}
}

func (builder *Builder) Build(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (flightfeatures.OperationalFeatures, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.OperationalFeatures{}, err
	}

	if len(item.Points) == 0 {
		return flightfeatures.OperationalFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:          flightfeatures.AvailabilityStatusUnavailable,
				TotalFieldCount: OperationalFeatureFieldCount,
				Limitations: []flightfeatures.FeatureLimitation{
					{
						Code:    "operational_point_evidence_unavailable",
						Message: "No trajectory points were available for operational feature extraction.",
					},
				},
			},
		}, nil
	}

	samples, err := collectSamples(ctx, item.Points)
	if err != nil {
		return flightfeatures.OperationalFeatures{}, err
	}

	limitations := samples.limitations()
	if item.PointCount > 0 &&
		item.PointCount != len(item.Points) {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: "trajectory_point_count_metadata_mismatch",
				Message: fmt.Sprintf(
					"Trajectory point-count metadata reports %d points while %d point records are present.",
					item.PointCount,
					len(item.Points),
				),
			},
		)
	}

	features := flightfeatures.OperationalFeatures{
		Evidence: flightfeatures.GroupEvidence{
			TotalFieldCount:      OperationalFeatureFieldCount,
			SupportingPointCount: len(item.Points),
			Limitations:          limitations,
		},
		GroundObservationShare: float64(samples.groundObservationCount) /
			float64(len(item.Points)),
		AirborneObservationShare: float64(
			len(item.Points)-
				samples.groundObservationCount,
		) / float64(len(item.Points)),
	}

	availableFieldCount := 2

	if len(samples.altitudes) > 0 {
		minimum, maximum, mean :=
			summarize(samples.altitudes)
		features.MinimumAltitudeM = minimum
		features.MaximumAltitudeM = maximum
		features.MeanAltitudeM = mean
		features.AltitudeRangeM = maximum - minimum
		availableFieldCount += 4
	}

	if len(samples.velocities) > 0 {
		_, maximum, mean :=
			summarize(samples.velocities)
		features.MeanVelocityMPS = mean
		features.MaximumVelocityMPS = maximum
		availableFieldCount += 2
	}

	if len(samples.absoluteVerticalRates) > 0 {
		_, maximum, mean :=
			summarize(samples.absoluteVerticalRates)
		features.MeanAbsoluteVerticalRateMPS = mean
		features.MaximumAbsoluteVerticalRateMPS = maximum
		availableFieldCount += 2
	}

	if len(samples.headings) > 0 {
		features.HeadingChangeDegrees =
			cumulativeHeadingChange(samples.headings)
		availableFieldCount++
	}

	features.Evidence.AvailableFieldCount =
		availableFieldCount
	switch {
	case availableFieldCount ==
		OperationalFeatureFieldCount:
		features.Evidence.Status =
			flightfeatures.AvailabilityStatusAvailable
	case availableFieldCount > 0:
		features.Evidence.Status =
			flightfeatures.AvailabilityStatusPartial
	default:
		features.Evidence.Status =
			flightfeatures.AvailabilityStatusUnavailable
	}

	if err := ctx.Err(); err != nil {
		return flightfeatures.OperationalFeatures{}, err
	}

	return cloneFeatures(features), nil
}

func summarize(values []float64) (
	float64,
	float64,
	float64,
) {
	minimum := values[0]
	maximum := values[0]
	total := 0.0

	for _, value := range values {
		if value < minimum {
			minimum = value
		}
		if value > maximum {
			maximum = value
		}
		total += value
	}

	return minimum,
		maximum,
		total / float64(len(values))
}

func cumulativeHeadingChange(
	headings []float64,
) float64 {
	if len(headings) < 2 {
		return 0
	}

	total := 0.0
	for index := 1; index < len(headings); index++ {
		delta := normalizeHeading(
			headings[index] - headings[index-1],
		)
		if delta > 180 {
			delta = 360 - delta
		}
		total += math.Abs(delta)
	}

	return total
}

func normalizeHeading(value float64) float64 {
	normalized := math.Mod(value, 360)
	if normalized < 0 {
		normalized += 360
	}
	if normalized == 0 {
		return 0
	}

	return normalized
}

func cloneFeatures(
	features flightfeatures.OperationalFeatures,
) flightfeatures.OperationalFeatures {
	cloned := features
	cloned.Evidence.Limitations = append(
		[]flightfeatures.FeatureLimitation(nil),
		features.Evidence.Limitations...,
	)

	return cloned
}
