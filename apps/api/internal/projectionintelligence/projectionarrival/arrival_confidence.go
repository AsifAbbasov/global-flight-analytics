package projectionarrival

import (
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func (
	estimator *Estimator,
) arrivalConfidence(
	projection projectioncontract.Result,
	destinationConfidenceScore float64,
	computation arrivalComputation,
) projectioncontract.Confidence {
	speedSupport := math.Min(
		1,
		float64(computation.speedSampleCount)/
			float64(estimator.config.MinimumSpeedSampleCount),
	)
	speedStability := 0.0
	if positiveFinite(computation.estimatedClosingSpeedMPS) {
		speedStability = 1 - math.Min(
			1,
			computation.closingSpeedStdDevMPS/
				computation.estimatedClosingSpeedMPS,
		)
		speedStability *= speedSupport
	}
	speedStability = clampUnit(speedStability)

	extrapolationRatio := 0.0
	if computation.extrapolationDuration > 0 {
		extrapolationRatio = math.Min(
			1,
			float64(computation.extrapolationDuration)/
				float64(estimator.config.MaximumEstimatedArrivalDuration),
		)
	}
	retention := clampUnit(
		1 - estimator.config.MaximumExtrapolationConfidenceLoss*
			extrapolationRatio,
	)

	projectionContribution :=
		estimator.config.ProjectionConfidenceWeight *
			projection.Confidence.Score * retention
	destinationContribution :=
		estimator.config.DestinationConfidenceWeight *
			destinationConfidenceScore * retention
	speedContribution :=
		estimator.config.SpeedStabilityWeight *
			speedStability * retention
	score := clampUnit(
		projectionContribution +
			destinationContribution +
			speedContribution,
	)

	return projectioncontract.Confidence{
		Score: score,
		Level: estimator.confidenceLevel(score),
		Reasons: []projectioncontract.ConfidenceReason{
			{
				Code:         "position_projection_confidence",
				Message:      "Position-projection confidence is weighted and reduced by the exact extrapolation-retention factor.",
				Contribution: projectionContribution,
			},
			{
				Code:         "destination_inference_confidence",
				Message:      "Route Intelligence destination confidence is weighted and reduced by the exact extrapolation-retention factor.",
				Contribution: destinationContribution,
			},
			{
				Code:         "directional_closing_speed_stability",
				Message:      "Arrival confidence uses signed closing-speed stability and complete bounded sample support.",
				Contribution: speedContribution,
			},
			{
				Code:         "extrapolation_confidence_retention",
				Message:      "The three additive contributions already include the extrapolation-retention factor, so their sum reconstructs the final score.",
				Contribution: 0,
			},
		},
	}
}

func (
	estimator *Estimator,
) confidenceLevel(
	score float64,
) projectioncontract.ConfidenceLevel {
	switch {
	case score >= estimator.config.HighConfidenceMinimum:
		return projectioncontract.ConfidenceLevelHigh
	case score >= estimator.config.MediumConfidenceMinimum:
		return projectioncontract.ConfidenceLevelMedium
	case score > 0:
		return projectioncontract.ConfidenceLevelLow
	default:
		return projectioncontract.ConfidenceLevelNone
	}
}

func (
	estimator *Estimator,
) combinedConfidence(
	projectionConfidence projectioncontract.Confidence,
	arrivalConfidence projectioncontract.Confidence,
) projectioncontract.Confidence {
	score := math.Min(
		projectionConfidence.Score,
		arrivalConfidence.Score,
	)

	return projectioncontract.Confidence{
		Score: score,
		Level: estimator.confidenceLevel(score),
		Reasons: []projectioncontract.ConfidenceReason{
			{
				Code:         "combined_projection_and_arrival_confidence",
				Message:      "Overall result confidence equals the weaker confidence between position projection and estimated arrival.",
				Contribution: score,
			},
		},
	}
}
