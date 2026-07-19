package routecontract

import (
	"fmt"
	"math"

	"strings"
	"time"
)

func validateStatusAndEndpoints(
	result Result,
	collector *validationCollector,
) {
	endpointCount := 0
	if result.Origin != nil {
		endpointCount++
	}
	if result.Destination != nil {
		endpointCount++
	}

	switch result.Status {
	case RouteStatusUnavailable:
		if endpointCount != 0 {
			collector.add(
				ValidationSeverityError,
				"unavailable_route_has_endpoint",
				"status",
				"Unavailable route results must not contain resolved endpoints.",
			)
		}
	case RouteStatusPartial:
		if endpointCount != 1 {
			collector.add(
				ValidationSeverityError,
				"partial_route_endpoint_count_invalid",
				"status",
				"Partial route results must contain exactly one resolved endpoint.",
			)
		}
	case RouteStatusComplete:
		if endpointCount != 2 {
			collector.add(
				ValidationSeverityError,
				"complete_route_endpoint_count_invalid",
				"status",
				"Complete route results must contain both origin and destination endpoints.",
			)
		}
	default:
		collector.add(
			ValidationSeverityError,
			"route_status_invalid",
			"status",
			"Route status must be unavailable, partial, or complete.",
		)
	}
}

func validateEndpoint(
	endpoint *EndpointInference,
	expectedRole EndpointRole,
	asOfTime time.Time,
	fieldPrefix string,
	collector *validationCollector,
) {
	if endpoint == nil {
		return
	}

	if endpoint.Role != expectedRole {
		collector.add(
			ValidationSeverityError,
			"endpoint_role_mismatch",
			fieldPrefix+".role",
			fmt.Sprintf(
				"Endpoint role must be %s.",
				expectedRole,
			),
		)
	}

	validateAirport(
		endpoint.Airport,
		fieldPrefix+".airport",
		collector,
	)
	if !finiteNonNegative(endpoint.DistanceKM) {
		collector.add(
			ValidationSeverityError,
			"endpoint_distance_invalid",
			fieldPrefix+".distance_km",
			"Endpoint distance must be finite and non-negative.",
		)
	}

	validateEvidence(
		endpoint.Evidence,
		asOfTime,
		fieldPrefix+".evidence",
		collector,
	)
	validateConfidence(
		endpoint.Confidence,
		fieldPrefix+".confidence",
		len(endpoint.Evidence),
		collector,
	)
	validateLimitations(
		endpoint.Limitations,
		fieldPrefix+".limitations",
		collector,
	)
}

func validateAirport(
	airport AirportReference,
	fieldPrefix string,
	collector *validationCollector,
) {
	if !icaoAirportPattern.MatchString(
		airport.ICAOCode,
	) {
		collector.add(
			ValidationSeverityError,
			"airport_icao_invalid",
			fieldPrefix+".icao_code",
			"Airport ICAO code must contain four uppercase alphanumeric characters.",
		)
	}
	if airport.IATACode != "" &&
		!iataAirportPattern.MatchString(
			airport.IATACode,
		) {
		collector.add(
			ValidationSeverityError,
			"airport_iata_invalid",
			fieldPrefix+".iata_code",
			"Airport IATA code must contain three uppercase alphanumeric characters when present.",
		)
	}
	if strings.TrimSpace(airport.Name) == "" {
		collector.add(
			ValidationSeverityError,
			"airport_name_required",
			fieldPrefix+".name",
			"Airport name is required.",
		)
	}
	if !validLatitude(airport.Latitude) {
		collector.add(
			ValidationSeverityError,
			"airport_latitude_invalid",
			fieldPrefix+".latitude",
			"Airport latitude must be finite and between -90 and 90 degrees.",
		)
	}
	if !validLongitude(airport.Longitude) {
		collector.add(
			ValidationSeverityError,
			"airport_longitude_invalid",
			fieldPrefix+".longitude",
			"Airport longitude must be finite and between -180 and 180 degrees.",
		)
	}
	if math.IsNaN(airport.ElevationM) ||
		math.IsInf(airport.ElevationM, 0) {
		collector.add(
			ValidationSeverityError,
			"airport_elevation_invalid",
			fieldPrefix+".elevation_m",
			"Airport elevation must be finite.",
		)
	}
}

func validateEvidence(
	items []Evidence,
	asOfTime time.Time,
	fieldPrefix string,
	collector *validationCollector,
) {
	for index, item := range items {
		itemPrefix := fmt.Sprintf(
			"%s[%d]",
			fieldPrefix,
			index,
		)
		if !validEvidenceType(item.Type) {
			collector.add(
				ValidationSeverityError,
				"evidence_type_invalid",
				itemPrefix+".type",
				"Evidence type is not supported by route-intelligence-v1.",
			)
		}
		if strings.TrimSpace(item.SourceName) == "" {
			collector.add(
				ValidationSeverityError,
				"evidence_source_required",
				itemPrefix+".source_name",
				"Evidence source name is required.",
			)
		}
		if !finiteRatio(item.Score) {
			collector.add(
				ValidationSeverityError,
				"evidence_score_invalid",
				itemPrefix+".score",
				"Evidence score must be finite and between zero and one.",
			)
		}
		if !finiteRatio(item.Weight) {
			collector.add(
				ValidationSeverityError,
				"evidence_weight_invalid",
				itemPrefix+".weight",
				"Evidence weight must be finite and between zero and one.",
			)
		}
		validateRequiredUTC(
			item.ObservedAt,
			itemPrefix+".observed_at",
			collector,
		)
		if !item.ObservedAt.IsZero() &&
			!asOfTime.IsZero() &&
			item.ObservedAt.After(asOfTime) {
			collector.add(
				ValidationSeverityError,
				"evidence_after_as_of_time",
				itemPrefix+".observed_at",
				"Evidence must not be observed after the analytical as-of time.",
			)
		}
		if strings.TrimSpace(item.Summary) == "" {
			collector.add(
				ValidationSeverityError,
				"evidence_summary_required",
				itemPrefix+".summary",
				"Evidence summary is required.",
			)
		}
		validateAttributes(
			item.Attributes,
			itemPrefix+".attributes",
			collector,
		)
	}
}

func totalEvidenceCount(
	result Result,
) int {
	total := 0
	if result.Origin != nil {
		total += len(result.Origin.Evidence)
	}
	if result.Destination != nil {
		total += len(result.Destination.Evidence)
	}

	return total
}

func validEvidenceType(
	value EvidenceType,
) bool {
	switch value {
	case EvidenceTypeTrajectoryEndpointProximity,
		EvidenceTypeGroundCycle,
		EvidenceTypeCallsignRouteToken,
		EvidenceTypeSourceFlightIdentity,
		EvidenceTypeAirportActivity,
		EvidenceTypeExternalReference:
		return true
	default:
		return false
	}
}
