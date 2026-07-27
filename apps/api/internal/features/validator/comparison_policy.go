package validator

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"

func greaterThanWithRelativeTolerance(
	left float64,
	right float64,
	tolerance float64,
) bool {
	return left > right && !approximatelyEqual(left, right, tolerance)
}

func lessThanWithRelativeTolerance(
	left float64,
	right float64,
	tolerance float64,
) bool {
	return left < right && !approximatelyEqual(left, right, tolerance)
}

func validateAvailableObservationSupport(
	collector *issueCollector,
	group flightfeatures.FeatureGroup,
	path string,
	evidence flightfeatures.GroupEvidence,
) {
	if evidence.Status != flightfeatures.AvailabilityStatusAvailable || evidence.SupportingPointCount > 0 {
		return
	}
	collector.error(
		group,
		path+".supporting_point_count",
		issueCodePrefix+"available_group_support_required",
		"An available observation-derived feature group must report at least one supporting observation.",
	)
}

func operationalGroundSharesClaimedAvailable(
	evidence flightfeatures.GroupEvidence,
) bool {
	if evidence.Status == flightfeatures.AvailabilityStatusUnavailable {
		return false
	}
	return !flightfeatures.HasLimitationCode(
		evidence.Limitations,
		flightfeatures.OperationalLimitationOnGroundUnavailable,
	) && !flightfeatures.HasLimitationCode(
		evidence.Limitations,
		flightfeatures.OperationalLimitationOnGroundMeasurementUnavailable,
	)
}
