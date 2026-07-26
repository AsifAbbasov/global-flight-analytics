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
	requirements := flightfeatures.CurrentGroupRequirementCounts()

	requiredAvailable := 0
	requiredTotal := 0
	optionalAvailable := 0
	optionalTotal := 0
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

	limitations := make([]flightfeatures.FeatureLimitation, 0)
	seenLimitations := make(map[string]struct{})

	for _, named := range evidenceGroups {
		evidence := named.evidence
		counts, exists := requirements[named.group]
		if !exists {
			return flightfeatures.FeatureQuality{}, fmt.Errorf(
				"%w: group=%s is absent from schema",
				ErrInvalidEvidenceFieldCount,
				named.group,
			)
		}
		if counts.Required > 0 && counts.Optional > 0 {
			return flightfeatures.FeatureQuality{}, fmt.Errorf(
				"%w: group=%s required=%d optional=%d",
				ErrMixedRequirementGroupEvidence,
				named.group,
				counts.Required,
				counts.Optional,
			)
		}
		if evidence.AvailableFieldCount < 0 ||
			evidence.TotalFieldCount < 0 ||
			evidence.AvailableFieldCount > evidence.TotalFieldCount ||
			evidence.TotalFieldCount != counts.Total() {
			return flightfeatures.FeatureQuality{}, fmt.Errorf(
				"%w: group=%s available=%d total=%d schema_total=%d",
				ErrInvalidEvidenceFieldCount,
				named.group,
				evidence.AvailableFieldCount,
				evidence.TotalFieldCount,
				counts.Total(),
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

		if counts.Required > 0 {
			requiredAvailable += evidence.AvailableFieldCount
			requiredTotal += evidence.TotalFieldCount
		} else {
			optionalAvailable += evidence.AvailableFieldCount
			optionalTotal += evidence.TotalFieldCount
		}
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

	return flightfeatures.FeatureQuality{
		Status:                flightfeatures.ValidationStatusUnvalidated,
		CompletenessScore:     ratioOrOne(requiredAvailable, requiredTotal),
		OptionalCoverageScore: ratioOrOne(optionalAvailable, optionalTotal),
		InputQualityScore:     inputQualityScore,
		SupportingPointCount:  supportingPointCount,
		Limitations:           limitations,
	}, nil
}

func ratioOrOne(available int, total int) float64 {
	if total == 0 {
		return 1
	}
	return float64(available) / float64(total)
}
