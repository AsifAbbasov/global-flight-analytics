package historicalcontract

import "time"

func validateScope(
	scope Scope,
	collector *validationCollector,
) {
	switch scope.Type {
	case ScopeTypeGlobal:
		validateEmptyScopeField(
			scope.RegionCode,
			"scope.region_code",
			collector,
		)
		validateEmptyScopeField(
			scope.AirportICAOCode,
			"scope.airport_icao_code",
			collector,
		)
		validateEmptyScopeField(
			scope.OriginICAOCode,
			"scope.origin_icao_code",
			collector,
		)
		validateEmptyScopeField(
			scope.DestinationICAOCode,
			"scope.destination_icao_code",
			collector,
		)

	case ScopeTypeRegion:
		if !regionCodePattern.MatchString(
			scope.RegionCode,
		) {
			collector.add(
				ValidationSeverityError,
				"region_code_invalid",
				"scope.region_code",
				"Region scope requires a normalized lowercase region code.",
			)
		}
		validateEmptyScopeField(
			scope.AirportICAOCode,
			"scope.airport_icao_code",
			collector,
		)
		validateEmptyScopeField(
			scope.OriginICAOCode,
			"scope.origin_icao_code",
			collector,
		)
		validateEmptyScopeField(
			scope.DestinationICAOCode,
			"scope.destination_icao_code",
			collector,
		)

	case ScopeTypeAirport:
		validateAirportCode(
			scope.AirportICAOCode,
			"scope.airport_icao_code",
			collector,
		)
		validateEmptyScopeField(
			scope.RegionCode,
			"scope.region_code",
			collector,
		)
		validateEmptyScopeField(
			scope.OriginICAOCode,
			"scope.origin_icao_code",
			collector,
		)
		validateEmptyScopeField(
			scope.DestinationICAOCode,
			"scope.destination_icao_code",
			collector,
		)

	case ScopeTypeRoute:
		validateAirportCode(
			scope.OriginICAOCode,
			"scope.origin_icao_code",
			collector,
		)
		validateAirportCode(
			scope.DestinationICAOCode,
			"scope.destination_icao_code",
			collector,
		)
		validateEmptyScopeField(
			scope.RegionCode,
			"scope.region_code",
			collector,
		)
		validateEmptyScopeField(
			scope.AirportICAOCode,
			"scope.airport_icao_code",
			collector,
		)

	default:
		collector.add(
			ValidationSeverityError,
			"scope_type_invalid",
			"scope.type",
			"Scope type must be global, region, airport, or route.",
		)
	}
}

func validateEmptyScopeField(
	value string,
	field string,
	collector *validationCollector,
) {
	if value != "" {
		collector.add(
			ValidationSeverityError,
			"scope_field_not_applicable",
			field,
			"Scope field must be empty for the selected scope type.",
		)
	}
}

func validateAirportCode(
	value string,
	field string,
	collector *validationCollector,
) {
	if !airportICAOPattern.MatchString(value) {
		collector.add(
			ValidationSeverityError,
			"airport_icao_invalid",
			field,
			"Airport ICAO code must contain four uppercase alphanumeric characters.",
		)
	}
}

func validateTimeWindow(
	window TimeWindow,
	fieldPrefix string,
	collector *validationCollector,
) {
	validateRequiredUTC(
		window.StartTime,
		fieldPrefix+".start_time",
		collector,
	)
	validateRequiredUTC(
		window.EndTime,
		fieldPrefix+".end_time",
		collector,
	)
	validateRequiredUTC(
		window.AsOfTime,
		fieldPrefix+".as_of_time",
		collector,
	)

	if !window.StartTime.IsZero() &&
		!window.EndTime.IsZero() &&
		!window.StartTime.Before(
			window.EndTime,
		) {
		collector.add(
			ValidationSeverityError,
			"window_not_positive",
			fieldPrefix,
			"Window start time must be before end time.",
		)
	}

	if !window.EndTime.IsZero() &&
		!window.AsOfTime.IsZero() &&
		window.EndTime.After(
			window.AsOfTime,
		) {
		collector.add(
			ValidationSeverityError,
			"window_exceeds_as_of_time",
			fieldPrefix+".end_time",
			"Window end time must not exceed the analytical as-of time.",
		)
	}
}

func validateGranularity(
	granularity Granularity,
	collector *validationCollector,
) {
	switch granularity {
	case GranularityHour,
		GranularityDay,
		GranularityWeek,
		GranularityCustom:
	default:
		collector.add(
			ValidationSeverityError,
			"granularity_invalid",
			"granularity",
			"Granularity must be hour, day, week, or custom.",
		)
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
