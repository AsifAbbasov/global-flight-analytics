package routecontract

import "strings"

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
