package datasetprofiler

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

var orderedGroups = []flightfeatures.FeatureGroup{
	flightfeatures.FeatureGroupTemporal,
	flightfeatures.FeatureGroupGeographical,
	flightfeatures.FeatureGroupOperational,
	flightfeatures.FeatureGroupTrajectory,
	flightfeatures.FeatureGroupAircraft,
}

type groupAccumulator struct {
	group                     flightfeatures.FeatureGroup
	schemaFieldCount          int
	recordCount               int
	availableCount            int
	partialCount              int
	unavailableCount          int
	unknownStatusCount        int
	fieldCompletenessTotal    float64
	supportingPointCountTotal float64
	limitationOccurrenceCount int
}

func newGroupAccumulators() map[flightfeatures.FeatureGroup]*groupAccumulator {
	fieldCounts := make(
		map[flightfeatures.FeatureGroup]int,
	)
	for _, definition := range flightfeatures.CurrentSchema().Definitions {
		fieldCounts[definition.Group]++
	}

	result := make(
		map[flightfeatures.FeatureGroup]*groupAccumulator,
		len(orderedGroups),
	)
	for _, group := range orderedGroups {
		result[group] = &groupAccumulator{
			group:            group,
			schemaFieldCount: fieldCounts[group],
		}
	}

	return result
}

func (accumulator *groupAccumulator) add(
	evidence flightfeatures.GroupEvidence,
) {
	accumulator.recordCount++
	switch evidence.Status {
	case flightfeatures.AvailabilityStatusAvailable:
		accumulator.availableCount++
	case flightfeatures.AvailabilityStatusPartial:
		accumulator.partialCount++
	case flightfeatures.AvailabilityStatusUnavailable:
		accumulator.unavailableCount++
	default:
		accumulator.unknownStatusCount++
	}

	if evidence.TotalFieldCount > 0 &&
		evidence.AvailableFieldCount >= 0 {
		completeness :=
			float64(evidence.AvailableFieldCount) /
				float64(evidence.TotalFieldCount)
		if completeness < 0 {
			completeness = 0
		}
		if completeness > 1 {
			completeness = 1
		}
		accumulator.fieldCompletenessTotal += completeness
	}

	if evidence.SupportingPointCount >= 0 {
		accumulator.supportingPointCountTotal +=
			float64(evidence.SupportingPointCount)
	}
	accumulator.limitationOccurrenceCount +=
		len(evidence.Limitations)
}

func (accumulator groupAccumulator) profile() GroupProfile {
	profile := GroupProfile{
		Group:                     accumulator.group,
		SchemaFieldCount:          accumulator.schemaFieldCount,
		RecordCount:               accumulator.recordCount,
		AvailableCount:            accumulator.availableCount,
		PartialCount:              accumulator.partialCount,
		UnavailableCount:          accumulator.unavailableCount,
		UnknownStatusCount:        accumulator.unknownStatusCount,
		LimitationOccurrenceCount: accumulator.limitationOccurrenceCount,
	}
	if accumulator.recordCount == 0 {
		return profile
	}

	denominator := float64(accumulator.recordCount)
	profile.AvailableRatio =
		float64(accumulator.availableCount) / denominator
	profile.PartialRatio =
		float64(accumulator.partialCount) / denominator
	profile.UnavailableRatio =
		float64(accumulator.unavailableCount) / denominator
	profile.MeanFieldCompleteness =
		accumulator.fieldCompletenessTotal / denominator
	profile.MeanSupportingPointCount =
		accumulator.supportingPointCountTotal / denominator

	return profile
}

func groupEvidence(
	features flightfeatures.FlightFeatures,
	group flightfeatures.FeatureGroup,
) flightfeatures.GroupEvidence {
	switch group {
	case flightfeatures.FeatureGroupTemporal:
		return features.Temporal.Evidence
	case flightfeatures.FeatureGroupGeographical:
		return features.Geographical.Evidence
	case flightfeatures.FeatureGroupOperational:
		return features.Operational.Evidence
	case flightfeatures.FeatureGroupTrajectory:
		return features.Trajectory.Evidence
	case flightfeatures.FeatureGroupAircraft:
		return features.Aircraft.Evidence
	default:
		return flightfeatures.GroupEvidence{}
	}
}

func allLimitations(
	features flightfeatures.FlightFeatures,
) []flightfeatures.FeatureLimitation {
	result := make(
		[]flightfeatures.FeatureLimitation,
		0,
		len(features.Temporal.Evidence.Limitations)+
			len(features.Geographical.Evidence.Limitations)+
			len(features.Operational.Evidence.Limitations)+
			len(features.Trajectory.Evidence.Limitations)+
			len(features.Aircraft.Evidence.Limitations)+
			len(features.Quality.Limitations),
	)
	result = append(
		result,
		features.Temporal.Evidence.Limitations...,
	)
	result = append(
		result,
		features.Geographical.Evidence.Limitations...,
	)
	result = append(
		result,
		features.Operational.Evidence.Limitations...,
	)
	result = append(
		result,
		features.Trajectory.Evidence.Limitations...,
	)
	result = append(
		result,
		features.Aircraft.Evidence.Limitations...,
	)
	result = append(
		result,
		features.Quality.Limitations...,
	)

	return result
}
