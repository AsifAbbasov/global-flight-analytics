package validator

import (
	"reflect"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func validateUnavailableTemporalPayload(
	collector *issueCollector,
	item flightfeatures.TemporalFeatures,
) {
	evidence := item.Evidence
	item.Evidence = flightfeatures.GroupEvidence{}
	validateUnavailablePayload(collector, flightfeatures.FeatureGroupTemporal, "temporal", evidence, item, flightfeatures.TemporalFeatures{})
}

func validateUnavailableGeographicalPayload(
	collector *issueCollector,
	item flightfeatures.GeographicalFeatures,
) {
	evidence := item.Evidence
	item.Evidence = flightfeatures.GroupEvidence{}
	validateUnavailablePayload(collector, flightfeatures.FeatureGroupGeographical, "geographical", evidence, item, flightfeatures.GeographicalFeatures{})
}

func validateUnavailableOperationalPayload(
	collector *issueCollector,
	item flightfeatures.OperationalFeatures,
) {
	evidence := item.Evidence
	item.Evidence = flightfeatures.GroupEvidence{}
	validateUnavailablePayload(collector, flightfeatures.FeatureGroupOperational, "operational", evidence, item, flightfeatures.OperationalFeatures{})
}

func validateUnavailableTrajectoryPayload(
	collector *issueCollector,
	item flightfeatures.TrajectoryFeatures,
) {
	evidence := item.Evidence
	item.Evidence = flightfeatures.GroupEvidence{}
	validateUnavailablePayload(collector, flightfeatures.FeatureGroupTrajectory, "trajectory", evidence, item, flightfeatures.TrajectoryFeatures{})
}

func validateUnavailablePayload(
	collector *issueCollector,
	group flightfeatures.FeatureGroup,
	path string,
	evidence flightfeatures.GroupEvidence,
	payload any,
	zero any,
) {
	if evidence.Status != flightfeatures.AvailabilityStatusUnavailable {
		return
	}
	if reflect.DeepEqual(payload, zero) {
		return
	}
	collector.error(
		group,
		path,
		issueCodePrefix+"unavailable_group_payload_not_zero",
		"Unavailable feature groups must use the canonical zero-value payload so stale, non-finite, or contradictory values cannot be hidden by availability status.",
	)
}
