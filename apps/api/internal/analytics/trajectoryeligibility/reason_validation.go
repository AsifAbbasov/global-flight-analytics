package trajectoryeligibility

func (reason ReasonCode) IsKnown() bool {
	switch reason {
	case ReasonMissingAircraftIdentifier,
		ReasonInvalidTimeRange,
		ReasonEvaluationTimeMissing,
		ReasonInsufficientPoints,
		ReasonLowQualityScore,
		ReasonTooManyCoverageGaps,
		ReasonDurationTooShort,
		ReasonDurationTooLong,
		ReasonMissingIdentity,
		ReasonIdentityNotReliable,
		ReasonMissingCallsign,
		ReasonMissingAltitude,
		ReasonFutureObservation,
		ReasonStaleObservations,
		ReasonInsufficientRecentContinuity:
		return true
	default:
		return false
	}
}
