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
		return flightfeatures.OperationalFeatures{}, ErrContextRequired
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.OperationalFeatures{}, err
	}

	samples, err := collectSamples(ctx, item)
	if err != nil {
		return flightfeatures.OperationalFeatures{}, err
	}
	limitations := samples.limitations()
	if item.PointCount != len(item.Points) {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: flightfeatures.LimitationTrajectoryPointCountMetadataMismatch,
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
			SupportingPointCount: samples.supportingPointCount,
			Limitations:          limitations,
		},
	}
	availableFieldCount := 0

	if len(samples.altitudes) > 0 {
		minimum, maximum, mean, available, aggregateErr := summarizeContext(
			ctx,
			samples.altitudes,
		)
		if aggregateErr != nil {
			return flightfeatures.OperationalFeatures{}, aggregateErr
		}
		if available {
			features.MinimumAltitudeM = minimum
			features.MaximumAltitudeM = maximum
			features.MeanAltitudeM = mean
			features.AltitudeRangeM = maximum - minimum
			availableFieldCount += 4
		} else {
			appendAggregateLimitation(&features, "altitude")
		}
	}

	if len(samples.velocities) > 0 {
		_, maximum, mean, available, aggregateErr := summarizeContext(
			ctx,
			samples.velocities,
		)
		if aggregateErr != nil {
			return flightfeatures.OperationalFeatures{}, aggregateErr
		}
		if available {
			features.MeanVelocityMPS = mean
			features.MaximumVelocityMPS = maximum
			availableFieldCount += 2
		} else {
			appendAggregateLimitation(&features, "velocity")
		}
	}

	if len(samples.absoluteVerticalRates) > 0 {
		_, maximum, mean, available, aggregateErr := summarizeContext(
			ctx,
			samples.absoluteVerticalRates,
		)
		if aggregateErr != nil {
			return flightfeatures.OperationalFeatures{}, aggregateErr
		}
		if available {
			features.MeanAbsoluteVerticalRateMPS = mean
			features.MaximumAbsoluteVerticalRateMPS = maximum
			availableFieldCount += 2
		} else {
			appendAggregateLimitation(&features, "vertical rate")
		}
	}

	if samples.headingSampleCount > 0 {
		headingChange, available, aggregateErr := sumContext(
			ctx,
			samples.headingChanges,
		)
		if aggregateErr != nil {
			return flightfeatures.OperationalFeatures{}, aggregateErr
		}
		if available {
			features.HeadingChangeDegrees = headingChange
			availableFieldCount++
		} else {
			appendAggregateLimitation(&features, "heading change")
		}
	}

	if samples.groundStateCount > 0 {
		denominator := float64(samples.groundStateCount)
		features.GroundObservationShare =
			float64(samples.groundObservationCount) / denominator
		features.AirborneObservationShare =
			float64(samples.airborneObservationCount) / denominator
		availableFieldCount += 2
	}

	features.Evidence.AvailableFieldCount = availableFieldCount
	switch {
	case availableFieldCount == OperationalFeatureFieldCount:
		features.Evidence.Status = flightfeatures.AvailabilityStatusAvailable
	case availableFieldCount > 0:
		features.Evidence.Status = flightfeatures.AvailabilityStatusPartial
	default:
		features.Evidence.Status = flightfeatures.AvailabilityStatusUnavailable
	}

	if err := ctx.Err(); err != nil {
		return flightfeatures.OperationalFeatures{}, err
	}
	return cloneFeatures(features), nil
}

func appendAggregateLimitation(
	features *flightfeatures.OperationalFeatures,
	signal string,
) {
	features.Evidence.Limitations = append(
		features.Evidence.Limitations,
		flightfeatures.FeatureLimitation{
			Code: flightfeatures.OperationalLimitationAggregateNonFinite,
			Message: fmt.Sprintf(
				"The %s aggregate became non-finite and was excluded from operational evidence.",
				signal,
			),
		},
	)
}

func summarizeContext(
	ctx context.Context,
	values []float64,
) (float64, float64, float64, bool, error) {
	if len(values) == 0 {
		return 0, 0, 0, false, nil
	}
	minimum := values[0]
	maximum := values[0]
	total := 0.0
	compensation := 0.0
	for index, value := range values {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return 0, 0, 0, false, err
			}
		}
		if !finite(value) {
			return 0, 0, 0, false, nil
		}
		if value < minimum {
			minimum = value
		}
		if value > maximum {
			maximum = value
		}
		corrected := value - compensation
		next := total + corrected
		compensation = (next - total) - corrected
		total = next
		if !finite(total) || !finite(compensation) {
			return 0, 0, 0, false, nil
		}
	}
	mean := total / float64(len(values))
	if !finite(minimum) || !finite(maximum) || !finite(mean) {
		return 0, 0, 0, false, nil
	}
	return minimum, maximum, mean, true, nil
}

func sumContext(
	ctx context.Context,
	values []float64,
) (float64, bool, error) {
	total := 0.0
	compensation := 0.0
	for index, value := range values {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return 0, false, err
			}
		}
		if !finite(value) {
			return 0, false, nil
		}
		corrected := value - compensation
		next := total + corrected
		compensation = (next - total) - corrected
		total = next
		if !finite(total) || !finite(compensation) {
			return 0, false, nil
		}
	}
	return total, true, nil
}

// summarize remains as a narrow package-level compatibility helper for focused
// mathematical tests. Production aggregation uses summarizeContext so context
// cancellation and non-finite accumulation are observable.
func summarize(values []float64) (float64, float64, float64) {
	minimum, maximum, mean, available, _ := summarizeContext(
		context.Background(),
		values,
	)
	if !available {
		return 0, 0, 0
	}
	return minimum, maximum, mean
}

// cumulativeHeadingChange accepts an already-valid contiguous heading sequence.
// Production collection breaks sequences when a measurement is unavailable or
// invalid and sums the resulting per-run changes directly.
func cumulativeHeadingChange(headings []float64) float64 {
	if len(headings) < 2 {
		return 0
	}
	changes := make([]float64, 0, len(headings)-1)
	for index := 1; index < len(headings); index++ {
		changes = append(
			changes,
			shortestHeadingChange(headings[index-1], headings[index]),
		)
	}
	total, available, _ := sumContext(context.Background(), changes)
	if !available || math.IsNaN(total) || math.IsInf(total, 0) {
		return 0
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
