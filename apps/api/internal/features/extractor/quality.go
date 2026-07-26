package extractor

import (
	"fmt"
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type namedGroupEvidence struct {
	group    flightfeatures.FeatureGroup
	evidence flightfeatures.GroupEvidence
}

func buildInitialQuality(
	features flightfeatures.FlightFeatures,
	item trajectory.FlightTrajectory,
) (flightfeatures.FeatureQuality, error) {
	evidenceGroups := []namedGroupEvidence{
		{group: flightfeatures.FeatureGroupTemporal, evidence: features.Temporal.Evidence},
		{group: flightfeatures.FeatureGroupGeographical, evidence: features.Geographical.Evidence},
		{group: flightfeatures.FeatureGroupOperational, evidence: features.Operational.Evidence},
		{group: flightfeatures.FeatureGroupTrajectory, evidence: features.Trajectory.Evidence},
		{group: flightfeatures.FeatureGroupAircraft, evidence: features.Aircraft.Evidence},
	}

	availableFieldCount := 0
	totalFieldCount := 0
	supportingPointCount := item.PointCount
	if len(item.Points) > supportingPointCount {
		supportingPointCount = len(item.Points)
	}
	if supportingPointCount < 0 {
		return flightfeatures.FeatureQuality{}, fmt.Errorf(
			"%w: trajectory point count=%d",
			ErrInvalidSupportingPointCount,
			supportingPointCount,
		)
	}

	limitations := make([]flightfeatures.FeatureLimitation, 0)
	seenLimitations := make(map[string]struct{})

	for _, named := range evidenceGroups {
		evidence := named.evidence
		if evidence.AvailableFieldCount < 0 ||
			evidence.TotalFieldCount < 0 ||
			evidence.AvailableFieldCount > evidence.TotalFieldCount {
			return flightfeatures.FeatureQuality{}, fmt.Errorf(
				"%w: group=%s available=%d total=%d",
				ErrInvalidEvidenceFieldCount,
				named.group,
				evidence.AvailableFieldCount,
				evidence.TotalFieldCount,
			)
		}
		if evidence.SupportingPointCount < 0 {
			return flightfeatures.FeatureQuality{}, fmt.Errorf(
				"%w: group=%s count=%d",
				ErrInvalidSupportingPointCount,
				named.group,
				evidence.SupportingPointCount,
			)
		}

		availableFieldCount += evidence.AvailableFieldCount
		totalFieldCount += evidence.TotalFieldCount
		if evidence.SupportingPointCount > supportingPointCount {
			supportingPointCount = evidence.SupportingPointCount
		}

		for _, limitation := range evidence.Limitations {
			key := limitation.Code + "\x00" + limitation.Message
			if _, exists := seenLimitations[key]; exists {
				continue
			}
			seenLimitations[key] = struct{}{}
			limitations = append(limitations, limitation)
		}
	}

	inputQualityScore := features.Trajectory.TrajectoryQualityScore
	if math.IsNaN(inputQualityScore) ||
		math.IsInf(inputQualityScore, 0) ||
		inputQualityScore < 0 ||
		inputQualityScore > 1 {
		return flightfeatures.FeatureQuality{}, fmt.Errorf(
			"%w: value=%v",
			ErrInvalidInputQualityScore,
			inputQualityScore,
		)
	}

	completenessScore := 0.0
	if totalFieldCount > 0 {
		completenessScore = float64(availableFieldCount) / float64(totalFieldCount)
	}

	return flightfeatures.FeatureQuality{
		Status:               flightfeatures.ValidationStatusUnvalidated,
		CompletenessScore:    completenessScore,
		InputQualityScore:    inputQualityScore,
		SupportingPointCount: supportingPointCount,
		Limitations:          limitations,
	}, nil
}
