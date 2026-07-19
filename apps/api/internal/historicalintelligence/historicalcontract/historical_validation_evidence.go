package historicalcontract

import (
	"fmt"
	"math"

	"strings"
	"time"
)

func validateConfidence(
	confidence Confidence,
	fieldPrefix string,
	expectedSampleCount int,
	collector *validationCollector,
) {
	if !isRatio(confidence.Score) {
		collector.add(
			ValidationSeverityError,
			"confidence_score_invalid",
			fieldPrefix+".score",
			"Confidence score must be between zero and one.",
		)
	}

	expectedLevel := ConfidenceLevelForScore(
		confidence.Score,
	)
	if confidence.Level != expectedLevel {
		collector.add(
			ValidationSeverityError,
			"confidence_level_mismatch",
			fieldPrefix+".level",
			"Confidence level must match the normalized score.",
		)
	}

	if confidence.SampleCount < 0 {
		collector.add(
			ValidationSeverityError,
			"confidence_sample_count_invalid",
			fieldPrefix+".sample_count",
			"Confidence sample count must not be negative.",
		)
	}
	if confidence.SampleCount !=
		expectedSampleCount {
		collector.add(
			ValidationSeverityError,
			"confidence_sample_count_mismatch",
			fieldPrefix+".sample_count",
			"Confidence sample count must match represented source samples.",
		)
	}

	seenCodes := make(map[string]struct{})
	for index, reason := range confidence.Reasons {
		reasonPrefix := fmt.Sprintf(
			"%s.reasons[%d]",
			fieldPrefix,
			index,
		)

		if strings.TrimSpace(reason.Code) == "" ||
			reason.Code !=
				strings.TrimSpace(reason.Code) {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_code_invalid",
				reasonPrefix+".code",
				"Confidence reason code must be normalized and non-empty.",
			)
		}
		if _, exists := seenCodes[reason.Code]; exists {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_duplicate",
				reasonPrefix+".code",
				"Confidence reason codes must be unique.",
			)
		}
		seenCodes[reason.Code] = struct{}{}

		if strings.TrimSpace(reason.Message) == "" {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_message_required",
				reasonPrefix+".message",
				"Confidence reason message is required.",
			)
		}
		if !isFinite(reason.Contribution) ||
			reason.Contribution < -1 ||
			reason.Contribution > 1 {
			collector.add(
				ValidationSeverityError,
				"confidence_reason_contribution_invalid",
				reasonPrefix+".contribution",
				"Confidence reason contribution must be between negative one and one.",
			)
		}
	}
}

func validateLimitations(
	items []Limitation,
	fieldPrefix string,
	collector *validationCollector,
) {
	seen := make(map[string]struct{})

	for index, item := range items {
		itemPrefix := fmt.Sprintf(
			"%s[%d]",
			fieldPrefix,
			index,
		)

		if strings.TrimSpace(item.Code) == "" ||
			item.Code != strings.TrimSpace(
				item.Code,
			) {
			collector.add(
				ValidationSeverityError,
				"limitation_code_invalid",
				itemPrefix+".code",
				"Limitation code must be normalized and non-empty.",
			)
		}
		if strings.TrimSpace(item.Message) == "" {
			collector.add(
				ValidationSeverityError,
				"limitation_message_required",
				itemPrefix+".message",
				"Limitation message is required.",
			)
		}
		if strings.TrimSpace(item.Scope) == "" ||
			item.Scope != strings.TrimSpace(
				item.Scope,
			) {
			collector.add(
				ValidationSeverityError,
				"limitation_scope_invalid",
				itemPrefix+".scope",
				"Limitation scope must be normalized and non-empty.",
			)
		}

		key := item.Scope + "\x00" + item.Code
		if _, exists := seen[key]; exists {
			collector.add(
				ValidationSeverityError,
				"limitation_duplicate",
				itemPrefix+".code",
				"Limitation scope and code combinations must be unique.",
			)
		}
		seen[key] = struct{}{}
	}
}

func validateProvenance(
	result Result,
	collector *validationCollector,
) {
	if strings.TrimSpace(
		result.Provenance.BuilderVersion,
	) == "" ||
		result.Provenance.BuilderVersion !=
			strings.TrimSpace(
				result.Provenance.BuilderVersion,
			) {
		collector.add(
			ValidationSeverityError,
			"builder_version_invalid",
			"provenance.builder_version",
			"Builder version must be normalized and non-empty.",
		)
	}

	if !fingerprintPattern.MatchString(
		result.Provenance.InputFingerprint,
	) {
		collector.add(
			ValidationSeverityError,
			"input_fingerprint_invalid",
			"provenance.input_fingerprint",
			"Input fingerprint must use the SHA-256 format.",
		)
	}

	if len(result.Provenance.SourceNames) == 0 {
		collector.add(
			ValidationSeverityError,
			"source_names_required",
			"provenance.source_names",
			"At least one provenance source name is required.",
		)
	}

	previousSource := ""
	seenSources := make(map[string]struct{})
	for index, sourceName := range result.Provenance.SourceNames {
		field := fmt.Sprintf(
			"provenance.source_names[%d]",
			index,
		)

		if strings.TrimSpace(sourceName) == "" ||
			sourceName != strings.TrimSpace(
				sourceName,
			) {
			collector.add(
				ValidationSeverityError,
				"source_name_invalid",
				field,
				"Source name must be normalized and non-empty.",
			)
		}
		if _, exists := seenSources[sourceName]; exists {
			collector.add(
				ValidationSeverityError,
				"source_name_duplicate",
				field,
				"Source names must be unique.",
			)
		}
		seenSources[sourceName] = struct{}{}

		if previousSource != "" &&
			sourceName < previousSource {
			collector.add(
				ValidationSeverityError,
				"source_names_not_sorted",
				field,
				"Source names must use deterministic ascending order.",
			)
		}
		previousSource = sourceName
	}

	validateRequiredUTC(
		result.Provenance.LatestSourceUpdatedAt,
		"provenance.latest_source_updated_at",
		collector,
	)
	if !result.Provenance.LatestSourceUpdatedAt.
		IsZero() &&
		!result.Window.AsOfTime.IsZero() &&
		result.Provenance.LatestSourceUpdatedAt.
			After(result.Window.AsOfTime) {
		collector.add(
			ValidationSeverityError,
			"source_updated_after_as_of_time",
			"provenance.latest_source_updated_at",
			"Latest source update time must not exceed the analytical as-of time.",
		)
	}

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
			"Generated time must not precede the analytical as-of time.",
		)
	}
}

func isHourBoundary(
	value time.Time,
) bool {
	return value.Minute() == 0 &&
		value.Second() == 0 &&
		value.Nanosecond() == 0
}

func isDayBoundary(
	value time.Time,
) bool {
	return value.Hour() == 0 &&
		isHourBoundary(value)
}

func isFinite(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}

func isRatio(
	value float64,
) bool {
	return isFinite(value) &&
		value >= 0 &&
		value <= 1
}

func almostEqual(
	left float64,
	right float64,
) bool {
	const tolerance = 1e-9

	difference := math.Abs(left - right)
	scale := math.Max(
		1,
		math.Max(
			math.Abs(left),
			math.Abs(right),
		),
	)

	return difference <= tolerance*scale
}
