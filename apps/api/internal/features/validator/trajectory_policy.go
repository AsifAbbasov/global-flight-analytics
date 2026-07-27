package validator

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"

func trajectoryPathEfficiencyComparable(
	evidence flightfeatures.GroupEvidence,
) bool {
	for _, code := range []string{
		flightfeatures.TrajectoryLimitationPathEvidenceInsufficient,
		flightfeatures.TrajectoryLimitationPathZeroDistance,
		flightfeatures.TrajectoryLimitationPathAggregateNonFinite,
		flightfeatures.TrajectoryLimitationPathRatioOutOfRange,
		flightfeatures.TrajectoryLimitationPathDiscontinuityExcluded,
		flightfeatures.TrajectoryLimitationPathSegmentFallback,
		flightfeatures.TrajectoryLimitationDuplicateTimestampsCollapsed,
		flightfeatures.TrajectoryLimitationInvalidPointCoordinates,
	} {
		if flightfeatures.HasLimitationCode(evidence.Limitations, code) {
			return false
		}
	}
	return true
}
