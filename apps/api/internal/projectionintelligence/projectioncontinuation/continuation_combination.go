package projectioncontinuation

import (
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
	"time"
)

func (
	baseline *Baseline,
) combineSamples(
	samples []projectedSample,
	pattern projectionpatternconfidence.Result,
	plan projectionhorizon.Plan,
	sequence int,
	forecastTime time.Time,
) (
	projectioncontract.ProjectionPoint,
	bool,
	error,
) {
	geoPoints := make(
		[]weightedGeoPoint,
		0,
		len(samples),
	)
	totalWeight := 0.0
	for _, sample := range samples {
		geoPoints = append(
			geoPoints,
			weightedGeoPoint{
				latitude:  sample.latitude,
				longitude: sample.longitude,
				weight:    sample.weight,
			},
		)
		totalWeight += sample.weight
	}

	latitude, longitude, valid :=
		weightedMeanGeoPoint(
			geoPoints,
		)
	if !valid ||
		!positiveFinite(totalWeight) {
		return projectioncontract.
				ProjectionPoint{},
			false,
			ErrContinuationContractInvalid
	}

	offsetSeconds := forecastTime.Sub(
		plan.AsOfTime,
	).Seconds()
	horizonSeconds :=
		plan.EffectiveDuration.Seconds()
	if offsetSeconds <= 0 ||
		horizonSeconds <= 0 {
		return projectioncontract.
				ProjectionPoint{},
			false,
			ErrContinuationContractInvalid
	}

	horizontalSpreadSquared := 0.0
	altitudeWeight := 0.0
	weightedAltitude := 0.0
	for _, sample := range samples {
		distanceM :=
			greatCircleDistanceM(
				latitude,
				longitude,
				sample.latitude,
				sample.longitude,
			)
		horizontalSpreadSquared +=
			sample.weight *
				distanceM *
				distanceM

		if sample.altitudeM != nil {
			weightedAltitude +=
				sample.weight *
					*sample.altitudeM
			altitudeWeight +=
				sample.weight
		}
	}
	horizontalSpreadM := math.Sqrt(
		horizontalSpreadSquared /
			totalWeight,
	)
	configuredHorizontal :=
		baseline.config.
			InitialHorizontalUncertaintyM +
			baseline.config.
				HorizontalUncertaintyGrowthMPS*
				offsetSeconds
	horizontalUncertaintyM := math.Max(
		configuredHorizontal,
		horizontalSpreadM*
			baseline.config.
				NeighborSpreadMultiplier,
	)
	if !positiveFinite(
		horizontalUncertaintyM,
	) {
		return projectioncontract.
				ProjectionPoint{},
			false,
			ErrContinuationContractInvalid
	}

	position := projectioncontract.Position{
		Latitude:  latitude,
		Longitude: longitude,
	}
	uncertainty :=
		projectioncontract.Uncertainty{
			HorizontalRadiusM: horizontalUncertaintyM,
		}

	altitudeSampleCount := 0
	for _, sample := range samples {
		if sample.altitudeM != nil {
			altitudeSampleCount++
		}
	}
	altitudeAvailable :=
		altitudeSampleCount >=
			baseline.config.
				MinimumAltitudeSupport &&
			altitudeWeight > 0
	if altitudeAvailable {
		altitudeM :=
			weightedAltitude /
				altitudeWeight
		verticalSpreadSquared := 0.0
		for _, sample := range samples {
			if sample.altitudeM == nil {
				continue
			}
			delta :=
				*sample.altitudeM -
					altitudeM
			verticalSpreadSquared +=
				sample.weight *
					delta *
					delta
		}
		verticalSpreadM := math.Sqrt(
			verticalSpreadSquared /
				altitudeWeight,
		)
		configuredVertical :=
			baseline.config.
				InitialVerticalUncertaintyM +
				baseline.config.
					VerticalUncertaintyGrowthMPS*
					offsetSeconds
		verticalUncertaintyM := math.Max(
			configuredVertical,
			verticalSpreadM*
				baseline.config.
					NeighborSpreadMultiplier,
		)
		if finite(altitudeM) &&
			positiveFinite(
				verticalUncertaintyM,
			) {
			position.AltitudeM =
				float64Pointer(altitudeM)
			uncertainty.VerticalRadiusM =
				float64Pointer(
					verticalUncertaintyM,
				)
		} else {
			altitudeAvailable = false
		}
	}

	supportRatio := clampUnit(
		float64(len(samples)) /
			float64(
				pattern.NeighborCount,
			),
	)
	progress :=
		offsetSeconds /
			horizonSeconds
	score := pattern.Score *
		supportRatio *
		(1 -
			baseline.config.
				MaximumConfidenceLoss*
				progress)
	score = clampUnit(score)

	return projectioncontract.ProjectionPoint{
		Sequence:     sequence,
		ForecastTime: forecastTime.UTC(),
		Position:     position,
		Uncertainty:  uncertainty,
		Confidence: projectioncontract.Confidence{
			Score: score,
			Level: baseline.
				confidenceLevel(score),
			Reasons: []projectioncontract.
				ConfidenceReason{
				{
					Code:         "pattern_confidence_support_and_horizon_decay",
					Message:      "Point confidence combines pattern confidence, usable neighbor support, and configured horizon decay.",
					Contribution: score,
				},
			},
		},
	}, altitudeAvailable, nil
}
