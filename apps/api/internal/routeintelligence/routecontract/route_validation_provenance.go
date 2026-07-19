package routecontract

import (
	"fmt"
	"math"

	"strings"
	"time"
)

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
