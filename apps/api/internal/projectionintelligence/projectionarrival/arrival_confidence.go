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
		float64(
			computation.
				speedSampleCount,
		)/
			float64(
				estimator.config.
					MinimumSpeedSampleCount,
			),
	)
	speedStability := 0.0
	if positiveFinite(
		computation.
			estimatedGroundSpeedMPS,
	) {
		speedStability =
			1 -
				math.Min(
					1,
					computation.
						groundSpeedStdDevMPS/
						computation.
							estimatedGroundSpeedMPS,
				)
		speedStability *=
			speedSupport
	}
	speedStability =
		clampUnit(speedStability)

	score :=
		estimator.config.
			ProjectionConfidenceWeight*
			projection.Confidence.Score +
			estimator.config.
				DestinationConfidenceWeight*
				destinationConfidenceScore +
			estimator.config.
				SpeedStabilityWeight*
				speedStability

	extrapolationRatio := 0.0
	if computation.
		extrapolationDuration > 0 {
		extrapolationRatio =
			math.Min(
				1,
				float64(
					computation.
						extrapolationDuration,
				)/
					float64(
						estimator.config.
							MaximumEstimatedArrivalDuration,
					),
			)
	}
	score *= 1 -
		estimator.config.
			MaximumExtrapolationConfidenceLoss*
			extrapolationRatio
	score = clampUnit(score)

	return projectioncontract.Confidence{
		Score: score,
		Level: estimator.confidenceLevel(
			score,
		),
		Reasons: []projectioncontract.ConfidenceReason{
			{
				Code:    "position_projection_confidence",
				Message: "Arrival confidence includes the position-projection confidence.",
				Contribution: estimator.config.
					ProjectionConfidenceWeight *
					projection.
						Confidence.Score,
			},
			{
				Code:    "destination_inference_confidence",
				Message: "Arrival confidence includes Route Intelligence destination confidence.",
				Contribution: estimator.config.
					DestinationConfidenceWeight *
					destinationConfidenceScore,
			},
			{
				Code:    "projected_speed_stability",
				Message: "Arrival confidence includes projected ground-speed stability and sample support.",
				Contribution: estimator.config.
					SpeedStabilityWeight *
					speedStability,
			},
			{
				Code:    "extrapolation_confidence_decay",
				Message: "Arrival confidence decreases when the estimate extends beyond the position-projection horizon.",
				Contribution: -estimator.config.
					MaximumExtrapolationConfidenceLoss *
					extrapolationRatio,
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
	case score >= estimator.config.
		HighConfidenceMinimum:
		return projectioncontract.
			ConfidenceLevelHigh
	case score >= estimator.config.
		MediumConfidenceMinimum:
		return projectioncontract.
			ConfidenceLevelMedium
	case score > 0:
		return projectioncontract.
			ConfidenceLevelLow
	default:
		return projectioncontract.
			ConfidenceLevelNone
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

	level :=
		estimator.confidenceLevel(
			score,
		)

	return projectioncontract.Confidence{
		Score: score,
		Level: level,
		Reasons: []projectioncontract.ConfidenceReason{
			{
				Code:         "combined_projection_and_arrival_confidence",
				Message:      "Overall result confidence equals the weaker confidence between position projection and estimated arrival.",
				Contribution: score,
			},
		},
	}
}
