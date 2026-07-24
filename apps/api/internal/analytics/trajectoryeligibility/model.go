package trajectoryeligibility

type Capability string

const (
	CapabilityTrafficMetrics        Capability = "traffic_metrics"
	CapabilityAirportActivity       Capability = "airport_activity"
	CapabilityRouteInference        Capability = "route_inference"
	CapabilityHistoricalAggregation Capability = "historical_aggregation"
	CapabilityProjection            Capability = "projection"
)

var orderedCapabilities = []Capability{
	CapabilityTrafficMetrics,
	CapabilityAirportActivity,
	CapabilityRouteInference,
	CapabilityHistoricalAggregation,
	CapabilityProjection,
}

func Capabilities() []Capability {
	return append(
		[]Capability(nil),
		orderedCapabilities...,
	)
}

type ReasonCode string

const (
	ReasonMissingAircraftIdentifier    ReasonCode = "missing_aircraft_identifier"
	ReasonInvalidTimeRange             ReasonCode = "invalid_time_range"
	ReasonEvaluationTimeMissing        ReasonCode = "evaluation_time_missing"
	ReasonInsufficientPoints           ReasonCode = "insufficient_points"
	ReasonLowQualityScore              ReasonCode = "low_quality_score"
	ReasonTooManyCoverageGaps          ReasonCode = "too_many_coverage_gaps"
	ReasonDurationTooShort             ReasonCode = "duration_too_short"
	ReasonDurationTooLong              ReasonCode = "duration_too_long"
	ReasonMissingIdentity              ReasonCode = "missing_identity"
	ReasonIdentityNotReliable          ReasonCode = "identity_not_reliable"
	ReasonMissingCallsign              ReasonCode = "missing_callsign"
	ReasonMissingAltitude              ReasonCode = "missing_altitude"
	ReasonFutureObservation            ReasonCode = "future_observation"
	ReasonStaleObservations            ReasonCode = "stale_observations"
	ReasonInsufficientRecentContinuity ReasonCode = "insufficient_recent_continuity"
)

type PermissionFlags struct {
	AllowTrafficMetrics        bool
	AllowAirportActivity       bool
	AllowRouteInference        bool
	AllowHistoricalAggregation bool
	AllowProjection            bool
}

func (flags PermissionFlags) Allowed(
	capability Capability,
) bool {
	switch capability {
	case CapabilityTrafficMetrics:
		return flags.AllowTrafficMetrics

	case CapabilityAirportActivity:
		return flags.AllowAirportActivity

	case CapabilityRouteInference:
		return flags.AllowRouteInference

	case CapabilityHistoricalAggregation:
		return flags.AllowHistoricalAggregation

	case CapabilityProjection:
		return flags.AllowProjection

	default:
		return false
	}
}

func (
	flags *PermissionFlags,
) set(
	capability Capability,
	allowed bool,
) {
	switch capability {
	case CapabilityTrafficMetrics:
		flags.AllowTrafficMetrics = allowed

	case CapabilityAirportActivity:
		flags.AllowAirportActivity = allowed

	case CapabilityRouteInference:
		flags.AllowRouteInference = allowed

	case CapabilityHistoricalAggregation:
		flags.AllowHistoricalAggregation = allowed

	case CapabilityProjection:
		flags.AllowProjection = allowed
	}
}

type Decision struct {
	Capability Capability
	Allowed    bool
	Reasons    []ReasonCode
}

func (
	decision Decision,
) HasReason(
	reason ReasonCode,
) bool {
	for _, current := range decision.Reasons {
		if current == reason {
			return true
		}
	}

	return false
}

type Evaluation struct {
	Permissions PermissionFlags
	Decisions   []Decision
}

func (
	evaluation Evaluation,
) Decision(
	capability Capability,
) (Decision, bool) {
	for _, decision := range evaluation.Decisions {
		if decision.Capability == capability {
			result := decision
			result.Reasons = append(
				[]ReasonCode(nil),
				decision.Reasons...,
			)

			return result, true
		}
	}

	return Decision{}, false
}

func (
	evaluation Evaluation,
) Reasons(
	capability Capability,
) []ReasonCode {
	decision, exists := evaluation.Decision(
		capability,
	)
	if !exists {
		return nil
	}

	return decision.Reasons
}
