package routecontract

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

const ValidationVersion = "route-intelligence-contract-validation-v1"

var (
	icao24Pattern = regexp.MustCompile(
		`^[A-F0-9]{6}$`,
	)
	icaoAirportPattern = regexp.MustCompile(
		`^[A-Z0-9]{4}$`,
	)
	iataAirportPattern = regexp.MustCompile(
		`^[A-Z0-9]{3}$`,
	)
	identityKeyPattern = regexp.MustCompile(
		`^flight-identity-[0-9a-f]{64}$`,
	)
	fingerprintPattern = regexp.MustCompile(
		`^sha256:[0-9a-f]{64}$`,
	)
)

type ValidationIssue struct {
	Severity ValidationSeverity
	Code     string
	Field    string
	Message  string
}

type ValidationReport struct {
	Version      string
	Status       ValidationStatus
	ErrorCount   int
	WarningCount int
	Issues       []ValidationIssue
}

func (report ValidationReport) Clone() ValidationReport {
	cloned := report
	cloned.Issues = append(
		[]ValidationIssue(nil),
		report.Issues...,
	)

	return cloned
}

func Validate(
	result Result,
) ValidationReport {
	collector := validationCollector{}

	validateIdentity(result, &collector)
	validateWindow(result.Window, &collector)
	validateStatusAndEndpoints(result, &collector)
	validateEndpoint(
		result.Origin,
		EndpointRoleOrigin,
		result.Window.AsOfTime,
		"origin",
		&collector,
	)
	validateEndpoint(
		result.Destination,
		EndpointRoleDestination,
		result.Window.AsOfTime,
		"destination",
		&collector,
	)
	validateSummary(result, &collector)
	validateConfidence(
		result.Confidence,
		"confidence",
		totalEvidenceCount(result),
		&collector,
	)
	validateLimitations(
		result.Limitations,
		"limitations",
		&collector,
	)
	validateProvenance(result, &collector)

	sort.SliceStable(
		collector.issues,
		func(left int, right int) bool {
			leftIssue := collector.issues[left]
			rightIssue := collector.issues[right]

			if leftIssue.Field != rightIssue.Field {
				return leftIssue.Field <
					rightIssue.Field
			}
			if leftIssue.Code != rightIssue.Code {
				return leftIssue.Code <
					rightIssue.Code
			}
			if leftIssue.Severity !=
				rightIssue.Severity {
				return leftIssue.Severity <
					rightIssue.Severity
			}

			return leftIssue.Message <
				rightIssue.Message
		},
	)

	status := ValidationStatusValid
	if collector.errorCount > 0 {
		status = ValidationStatusInvalid
	}

	return ValidationReport{
		Version:      ValidationVersion,
		Status:       status,
		ErrorCount:   collector.errorCount,
		WarningCount: collector.warningCount,
		Issues: append(
			[]ValidationIssue(nil),
			collector.issues...,
		),
	}
}

type validationCollector struct {
	issues       []ValidationIssue
	errorCount   int
	warningCount int
}

func (collector *validationCollector) add(
	severity ValidationSeverity,
	code string,
	field string,
	message string,
) {
	collector.issues = append(
		collector.issues,
		ValidationIssue{
			Severity: severity,
			Code:     code,
			Field:    field,
			Message:  message,
		},
	)

	switch severity {
	case ValidationSeverityError:
		collector.errorCount++
	case ValidationSeverityWarning:
		collector.warningCount++
	}
}

func validateIdentity(
	result Result,
	collector *validationCollector,
) {
	if result.SchemaVersion != SchemaVersionV1 {
		collector.add(
			ValidationSeverityError,
			"unsupported_schema_version",
			"schema_version",
			"Route result must use route-intelligence-v1.",
		)
	}
	if strings.TrimSpace(result.TrajectoryID) == "" {
		collector.add(
			ValidationSeverityError,
			"trajectory_id_required",
			"trajectory_id",
			"Trajectory identifier is required.",
		)
	}
	if result.IdentityKey != "" &&
		!identityKeyPattern.MatchString(
			result.IdentityKey,
		) {
		collector.add(
			ValidationSeverityError,
			"identity_key_invalid",
			"identity_key",
			"Identity key must use the stable flight-identity SHA-256 format.",
		)
	}
	if !icao24Pattern.MatchString(result.ICAO24) {
		collector.add(
			ValidationSeverityError,
			"icao24_invalid",
			"icao24",
			"ICAO24 must contain six uppercase hexadecimal characters.",
		)
	}
	if result.Callsign !=
		strings.TrimSpace(result.Callsign) {
		collector.add(
			ValidationSeverityError,
			"callsign_not_normalized",
			"callsign",
			"Callsign must not contain surrounding whitespace.",
		)
	}
}

func validateWindow(
	window RouteWindow,
	collector *validationCollector,
) {
	validateRequiredUTC(
		window.StartTime,
		"window.start_time",
		collector,
	)
	validateRequiredUTC(
		window.EndTime,
		"window.end_time",
		collector,
	)
	validateRequiredUTC(
		window.AsOfTime,
		"window.as_of_time",
		collector,
	)

	if !window.StartTime.IsZero() &&
		!window.EndTime.IsZero() &&
		window.StartTime.After(window.EndTime) {
		collector.add(
			ValidationSeverityError,
			"window_reversed",
			"window",
			"Window start time must not be after end time.",
		)
	}
	if !window.EndTime.IsZero() &&
		!window.AsOfTime.IsZero() &&
		window.EndTime.After(window.AsOfTime) {
		collector.add(
			ValidationSeverityError,
			"window_exceeds_as_of_time",
			"window.end_time",
			"Window end time must not exceed the analytical as-of time.",
		)
	}
}

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

func validateAttributes(
	items []EvidenceAttribute,
	fieldPrefix string,
	collector *validationCollector,
) {
	previousKey := ""
	for index, item := range items {
		itemPrefix := fmt.Sprintf(
			"%s[%d]",
			fieldPrefix,
			index,
		)
		normalizedKey := strings.TrimSpace(
			item.Key,
		)
		if normalizedKey == "" {
			collector.add(
				ValidationSeverityError,
				"evidence_attribute_key_required",
				itemPrefix+".key",
				"Evidence attribute key is required.",
			)
		}
		if strings.TrimSpace(item.Value) == "" {
			collector.add(
				ValidationSeverityError,
				"evidence_attribute_value_required",
				itemPrefix+".value",
				"Evidence attribute value is required.",
			)
		}
		if index > 0 &&
			normalizedKey <= previousKey {
			collector.add(
				ValidationSeverityError,
				"evidence_attributes_not_sorted_unique",
				fieldPrefix,
				"Evidence attributes must be sorted by key and contain no duplicate keys.",
			)
			break
		}
		previousKey = normalizedKey
	}
}

func validateConfidence(
	confidence Confidence,
	fieldPrefix string,
	expectedEvidenceCount int,
	collector *validationCollector,
) {
	if !finiteRatio(confidence.Score) {
		collector.add(
			ValidationSeverityError,
			"confidence_score_invalid",
			fieldPrefix+".score",
			"Confidence score must be finite and between zero and one.",
		)
	} else if confidence.Level !=
		ConfidenceLevelForScore(confidence.Score) {
		collector.add(
			ValidationSeverityError,
			"confidence_level_mismatch",
			fieldPrefix+".level",
			"Confidence level does not match the score thresholds.",
		)
	}
	if confidence.EvidenceCount !=
		expectedEvidenceCount {
		collector.add(
			ValidationSeverityError,
			"confidence_evidence_count_mismatch",
			fieldPrefix+".evidence_count",
			fmt.Sprintf(
				"Confidence evidence count must equal %d.",
				expectedEvidenceCount,
			),
		)
	}

	seenCodes := make(map[string]struct{})
	for index, reason := range confidence.Reasons {
		itemPrefix := fmt.Sprintf(
			"%s.reasons[%d]",
			fieldPrefix,
			index,
		)
		code := strings.TrimSpace(reason.Code)
		if code == "" {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_code_required",
				itemPrefix+".code",
				"Confidence reason code is required.",
			)
		} else if _, exists := seenCodes[code]; exists {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_code_duplicate",
				itemPrefix+".code",
				"Confidence reason codes must be unique.",
			)
		} else {
			seenCodes[code] = struct{}{}
		}
		if strings.TrimSpace(reason.Message) == "" {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_message_required",
				itemPrefix+".message",
				"Confidence reason message is required.",
			)
		}
		if math.IsNaN(reason.Contribution) ||
			math.IsInf(reason.Contribution, 0) ||
			reason.Contribution < -1 ||
			reason.Contribution > 1 {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_contribution_invalid",
				itemPrefix+".contribution",
				"Confidence reason contribution must be finite and between -1 and 1.",
			)
		}
	}
}

func validateSummary(
	result Result,
	collector *validationCollector,
) {
	if !finiteNonNegative(
		result.Summary.GreatCircleDistanceKM,
	) {
		collector.add(
			ValidationSeverityError,
			"route_distance_invalid",
			"summary.great_circle_distance_km",
			"Great-circle route distance must be finite and non-negative.",
		)
	}

	if result.Origin == nil ||
		result.Destination == nil {
		if result.Summary.SameAirport {
			collector.add(
				ValidationSeverityError,
				"same_airport_without_complete_route",
				"summary.same_airport",
				"Same-airport status requires both route endpoints.",
			)
		}
		if result.Summary.GreatCircleDistanceKM != 0 {
			collector.add(
				ValidationSeverityWarning,
				"route_distance_without_complete_route",
				"summary.great_circle_distance_km",
				"Route distance is normally zero until both endpoints are resolved.",
			)
		}

		return
	}

	expectedSameAirport := strings.EqualFold(
		result.Origin.Airport.ICAOCode,
		result.Destination.Airport.ICAOCode,
	)
	if result.Summary.SameAirport !=
		expectedSameAirport {
		collector.add(
			ValidationSeverityError,
			"same_airport_mismatch",
			"summary.same_airport",
			"Same-airport status must match the endpoint ICAO codes.",
		)
	}
}

func validateLimitations(
	items []Limitation,
	fieldPrefix string,
	collector *validationCollector,
) {
	seenCodes := make(map[string]struct{})
	for index, item := range items {
		itemPrefix := fmt.Sprintf(
			"%s[%d]",
			fieldPrefix,
			index,
		)
		code := strings.TrimSpace(item.Code)
		if code == "" {
			collector.add(
				ValidationSeverityError,
				"limitation_code_required",
				itemPrefix+".code",
				"Limitation code is required.",
			)
		} else if _, exists := seenCodes[code]; exists {
			collector.add(
				ValidationSeverityError,
				"limitation_code_duplicate",
				itemPrefix+".code",
				"Limitation codes must be unique within their scope.",
			)
		} else {
			seenCodes[code] = struct{}{}
		}
		if strings.TrimSpace(item.Message) == "" {
			collector.add(
				ValidationSeverityError,
				"limitation_message_required",
				itemPrefix+".message",
				"Limitation message is required.",
			)
		}
		if strings.TrimSpace(item.Scope) == "" {
			collector.add(
				ValidationSeverityError,
				"limitation_scope_required",
				itemPrefix+".scope",
				"Limitation scope is required.",
			)
		}
	}
}

func validateProvenance(
	result Result,
	collector *validationCollector,
) {
	if strings.TrimSpace(
		result.Provenance.ResolverVersion,
	) == "" {
		collector.add(
			ValidationSeverityError,
			"resolver_version_required",
			"provenance.resolver_version",
			"Resolver version is required.",
		)
	}
	if !fingerprintPattern.MatchString(
		result.Provenance.InputFingerprint,
	) {
		collector.add(
			ValidationSeverityError,
			"input_fingerprint_invalid",
			"provenance.input_fingerprint",
			"Input fingerprint must use sha256 followed by 64 lowercase hexadecimal characters.",
		)
	}
	validateRequiredUTC(
		result.Provenance.TrajectoryUpdatedAt,
		"provenance.trajectory_updated_at",
		collector,
	)
	if !result.Provenance.TrajectoryUpdatedAt.IsZero() &&
		!result.Window.AsOfTime.IsZero() &&
		result.Provenance.TrajectoryUpdatedAt.After(
			result.Window.AsOfTime,
		) {
		collector.add(
			ValidationSeverityError,
			"trajectory_updated_after_as_of_time",
			"provenance.trajectory_updated_at",
			"Trajectory update time must not exceed the analytical as-of time.",
		)
	}
	validateSortedUniqueStrings(
		result.Provenance.SourceNames,
		"provenance.source_names",
		collector,
	)
	validateRequiredUTC(
		result.GeneratedAt,
		"generated_at",
		collector,
	)
	if !result.GeneratedAt.IsZero() &&
		!result.Window.AsOfTime.IsZero() &&
		result.GeneratedAt.Before(
			result.Window.AsOfTime,
		) {
		collector.add(
			ValidationSeverityError,
			"generated_before_as_of_time",
			"generated_at",
			"Generated time must not be before the analytical as-of time.",
		)
	}
}

func validateSortedUniqueStrings(
	items []string,
	field string,
	collector *validationCollector,
) {
	if len(items) == 0 {
		collector.add(
			ValidationSeverityError,
			"source_names_required",
			field,
			"At least one provenance source name is required.",
		)

		return
	}

	previous := ""
	for index, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			collector.add(
				ValidationSeverityError,
				"source_name_empty",
				fmt.Sprintf("%s[%d]", field, index),
				"Provenance source name must not be empty.",
			)
		}
		if index > 0 && normalized <= previous {
			collector.add(
				ValidationSeverityError,
				"source_names_not_sorted_unique",
				field,
				"Provenance source names must be sorted and contain no duplicates.",
			)
			break
		}
		previous = normalized
	}
}

func validateRequiredUTC(
	value time.Time,
	field string,
	collector *validationCollector,
) {
	if value.IsZero() {
		collector.add(
			ValidationSeverityError,
			"time_required",
			field,
			"UTC timestamp is required.",
		)

		return
	}
	if value.Location() != time.UTC {
		collector.add(
			ValidationSeverityError,
			"time_not_utc",
			field,
			"Timestamp must use the UTC location.",
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

func finiteRatio(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0 &&
		value <= 1
}

func finiteNonNegative(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0
}

func validLatitude(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= -90 &&
		value <= 90
}

func validLongitude(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= -180 &&
		value <= 180
}
