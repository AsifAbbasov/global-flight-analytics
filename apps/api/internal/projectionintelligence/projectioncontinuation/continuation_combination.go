package projectioncontinuation

import (
	"math"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

type sampleCombination struct {
	point               projectioncontract.ProjectionPoint
	altitudeAvailable   bool
	confidenceAvailable bool
}

func (
	baseline *Baseline,
) combineSamples(
	samples []projectedSample,
	pattern projectionpatternconfidence.Result,
	plan projectionhorizon.Plan,
	sequence int,
	forecastTime time.Time,
) (sampleCombination, error) {
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
		weightedMeanGeoPoint(geoPoints)
	if !valid || !positiveFinite(totalWeight) {
		return sampleCombination{},
			ErrContinuationContractInvalid
	}

	offsetSeconds := forecastTime.Sub(
		plan.AsOfTime,
	).Seconds()
	horizonSeconds := plan.EffectiveDuration.Seconds()
	if offsetSeconds <= 0 || horizonSeconds <= 0 {
		return sampleCombination{},
			ErrContinuationContractInvalid
	}

	horizontalSpreadSquared := 0.0
	altitudeWeight := 0.0
	weightedAltitude := 0.0
	for _, sample := range samples {
		distanceM := greatCircleDistanceM(
			latitude,
			longitude,
			sample.latitude,
			sample.longitude,
		)
		horizontalSpreadSquared +=
			sample.weight * distanceM * distanceM

		if sample.altitudeM != nil {
			weightedAltitude +=
				sample.weight * *sample.altitudeM
			altitudeWeight += sample.weight
		}
	}

	horizontalSpreadM := math.Sqrt(
		horizontalSpreadSquared / totalWeight,
	)
	configuredHorizontal :=
		baseline.config.InitialHorizontalUncertaintyM +
			baseline.config.HorizontalUncertaintyGrowthMPS*
				offsetSeconds
	horizontalDisagreementM :=
		horizontalSpreadM *
			baseline.config.NeighborSpreadMultiplier
	horizontalUncertaintyM, valid := composeUncertainty(
		configuredHorizontal,
		horizontalDisagreementM,
	)
	if !valid {
		return sampleCombination{},
			ErrContinuationContractInvalid
	}

	position := projectioncontract.Position{
		Latitude:  latitude,
		Longitude: longitude,
	}
	uncertainty := projectioncontract.Uncertainty{
		HorizontalRadiusM: horizontalUncertaintyM,
	}

	agreementFactor := uncertaintyAgreementFactor(
		configuredHorizontal,
		horizontalUncertaintyM,
	)
	if !positiveFinite(agreementFactor) ||
		agreementFactor > 1 {
		return sampleCombination{},
			ErrContinuationContractInvalid
	}

	altitudeSampleCount := 0
	for _, sample := range samples {
		if sample.altitudeM != nil {
			altitudeSampleCount++
		}
	}
	altitudeAvailable :=
		altitudeSampleCount >=
			baseline.config.MinimumAltitudeSupport &&
			altitudeWeight > 0
	if altitudeAvailable {
		altitudeM := weightedAltitude / altitudeWeight
		verticalSpreadSquared := 0.0
		for _, sample := range samples {
			if sample.altitudeM == nil {
				continue
			}
			delta := *sample.altitudeM - altitudeM
			verticalSpreadSquared +=
				sample.weight * delta * delta
		}
		verticalSpreadM := math.Sqrt(
			verticalSpreadSquared / altitudeWeight,
		)
		configuredVertical :=
			baseline.config.InitialVerticalUncertaintyM +
				baseline.config.VerticalUncertaintyGrowthMPS*
					offsetSeconds
		verticalDisagreementM :=
			verticalSpreadM *
				baseline.config.NeighborSpreadMultiplier
		verticalUncertaintyM, verticalValid :=
			composeUncertainty(
				configuredVertical,
				verticalDisagreementM,
			)
		if finite(altitudeM) && verticalValid {
			position.AltitudeM = float64Pointer(altitudeM)
			uncertainty.VerticalRadiusM =
				float64Pointer(verticalUncertaintyM)
			verticalAgreementFactor :=
				uncertaintyAgreementFactor(
					configuredVertical,
					verticalUncertaintyM,
				)
			agreementFactor = math.Min(
				agreementFactor,
				verticalAgreementFactor,
			)
		} else {
			altitudeAvailable = false
		}
	}

	supportRatio := effectiveWeightedSupportRatio(
		samples,
		pattern.NeighborCount,
	)
	progress := clampUnit(
		offsetSeconds / horizonSeconds,
	)
	horizonRetention := 1 -
		baseline.config.MaximumConfidenceLoss*progress
	score := clampUnit(
		pattern.Score *
			supportRatio *
			agreementFactor *
			horizonRetention,
	)

	confidence := projectioncontract.Confidence{
		Score: score,
		Level: baseline.confidenceLevel(score),
	}
	if score > 0 {
		confidence.Reasons =
			[]projectioncontract.ConfidenceReason{
				{
					Code:         "pattern_confidence",
					Message:      "Historical pattern confidence provides the upstream evidence-strength factor.",
					Contribution: clampUnit(pattern.Score),
				},
				{
					Code:         "effective_weighted_neighbor_support",
					Message:      "Neighbor support uses effective sample size so concentrated similarity weight cannot count as broad support.",
					Contribution: supportRatio,
				},
				{
					Code:         "neighbor_agreement",
					Message:      "Forecast confidence decreases as weighted horizontal or vertical disagreement expands total uncertainty.",
					Contribution: agreementFactor,
				},
				{
					Code:         "horizon_retention",
					Message:      "Confidence retention decreases monotonically across the configured forecast horizon.",
					Contribution: horizonRetention,
				},
			}
	}

	return sampleCombination{
		point: projectioncontract.ProjectionPoint{
			Sequence:     sequence,
			ForecastTime: forecastTime.UTC(),
			Position:     position,
			Uncertainty:  uncertainty,
			Confidence:   confidence,
		},
		altitudeAvailable:   altitudeAvailable,
		confidenceAvailable: score > 0,
	}, nil
}

// composeUncertainty conservatively preserves both the configured model
// uncertainty and the observed neighbor-disagreement uncertainty. Addition is
// used instead of root-sum-square because independence is not established.
func composeUncertainty(
	configured float64,
	disagreement float64,
) (float64, bool) {
	if !positiveFinite(configured) ||
		!nonNegativeFinite(disagreement) {
		return 0, false
	}

	combined := configured + disagreement
	return combined, positiveFinite(combined)
}

func uncertaintyAgreementFactor(
	configured float64,
	combined float64,
) float64 {
	if !positiveFinite(configured) ||
		!positiveFinite(combined) ||
		combined < configured {
		return 0
	}

	return clampUnit(configured / combined)
}

func effectiveWeightedSupportRatio(
	samples []projectedSample,
	neighborCount int,
) float64 {
	if len(samples) == 0 || neighborCount < 1 {
		return 0
	}

	totalWeight := 0.0
	squaredWeightSum := 0.0
	for _, sample := range samples {
		if !positiveFinite(sample.weight) {
			continue
		}
		totalWeight += sample.weight
		squaredWeightSum += sample.weight * sample.weight
	}
	if !positiveFinite(totalWeight) ||
		!positiveFinite(squaredWeightSum) {
		return 0
	}

	effectiveSampleSize :=
		totalWeight * totalWeight / squaredWeightSum
	return clampUnit(
		effectiveSampleSize / float64(neighborCount),
	)
}
