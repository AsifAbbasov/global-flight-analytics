package historicalcontract

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

const ValidationVersion = "historical-intelligence-contract-validation-v1"

var (
	airportICAOPattern = regexp.MustCompile(
		`^[A-Z0-9]{4}$`,
	)
	regionCodePattern = regexp.MustCompile(
		`^[a-z0-9][a-z0-9_-]{1,31}$`,
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

	validateContractIdentity(result, &collector)
	validateMetric(result.Metric, &collector)
	validateScope(result.Scope, &collector)
	validateTimeWindow(
		result.Window,
		"window",
		&collector,
	)
	validateGranularity(
		result.Granularity,
		&collector,
	)
	validatePoints(result, &collector)
	validateSeriesStatus(result, &collector)
	validateSummary(result, &collector)
	validateComparison(result, &collector)
	validateConfidence(
		result.Confidence,
		"confidence",
		totalSampleCount(result.Points),
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

func validateContractIdentity(
	result Result,
	collector *validationCollector,
) {
	if result.SchemaVersion != SchemaVersionV1 {
		collector.add(
			ValidationSeverityError,
			"unsupported_schema_version",
			"schema_version",
			"Historical result must use historical-intelligence-v1.",
		)
	}

	switch result.Status {
	case SeriesStatusUnavailable,
		SeriesStatusPartial,
		SeriesStatusComplete:
	default:
		collector.add(
			ValidationSeverityError,
			"series_status_invalid",
			"status",
			"Series status must be unavailable, partial, or complete.",
		)
	}
}

func validateMetric(
	metric Metric,
	collector *validationCollector,
) {
	if !isSupportedMetricName(metric.Name) {
		collector.add(
			ValidationSeverityError,
			"metric_name_unsupported",
			"metric.name",
			"Metric name is not part of historical-intelligence-v1.",
		)
	}

	if strings.TrimSpace(metric.Unit) == "" {
		collector.add(
			ValidationSeverityError,
			"metric_unit_required",
			"metric.unit",
			"Metric unit is required.",
		)
	} else if metric.Unit !=
		strings.TrimSpace(metric.Unit) {
		collector.add(
			ValidationSeverityError,
			"metric_unit_not_normalized",
			"metric.unit",
			"Metric unit must not contain surrounding whitespace.",
		)
	}

	switch metric.Aggregation {
	case AggregationCount,
		AggregationSum,
		AggregationMinimum,
		AggregationMaximum,
		AggregationAverage,
		AggregationMedian,
		AggregationRatio:
	default:
		collector.add(
			ValidationSeverityError,
			"metric_aggregation_invalid",
			"metric.aggregation",
			"Metric aggregation is unsupported.",
		)
	}
}

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

func validatePoints(
	result Result,
	collector *validationCollector,
) {
	for index, point := range result.Points {
		fieldPrefix := fmt.Sprintf(
			"points[%d]",
			index,
		)

		validateRequiredUTC(
			point.StartTime,
			fieldPrefix+".start_time",
			collector,
		)
		validateRequiredUTC(
			point.EndTime,
			fieldPrefix+".end_time",
			collector,
		)

		if !point.StartTime.IsZero() &&
			!point.EndTime.IsZero() &&
			!point.StartTime.Before(
				point.EndTime,
			) {
			collector.add(
				ValidationSeverityError,
				"bucket_not_positive",
				fieldPrefix,
				"Bucket start time must be before end time.",
			)
		}

		if !result.Window.StartTime.IsZero() &&
			point.StartTime.Before(
				result.Window.StartTime,
			) {
			collector.add(
				ValidationSeverityError,
				"bucket_before_window",
				fieldPrefix+".start_time",
				"Bucket must not start before the historical window.",
			)
		}
		if !result.Window.EndTime.IsZero() &&
			point.EndTime.After(
				result.Window.EndTime,
			) {
			collector.add(
				ValidationSeverityError,
				"bucket_after_window",
				fieldPrefix+".end_time",
				"Bucket must not end after the historical window.",
			)
		}
		if !result.Window.AsOfTime.IsZero() &&
			point.EndTime.After(
				result.Window.AsOfTime,
			) {
			collector.add(
				ValidationSeverityError,
				"bucket_future_evidence",
				fieldPrefix+".end_time",
				"Bucket must not contain evidence after the analytical as-of time.",
			)
		}

		validateBucketAlignment(
			point,
			result.Granularity,
			fieldPrefix,
			collector,
		)
		validatePointValue(
			point,
			fieldPrefix,
			collector,
		)
		validateConfidence(
			point.Confidence,
			fieldPrefix+".confidence",
			point.SampleCount,
			collector,
		)
		validateLimitations(
			point.Limitations,
			fieldPrefix+".limitations",
			collector,
		)

		if index > 0 {
			previous := result.Points[index-1]
			if !point.StartTime.After(
				previous.StartTime,
			) {
				collector.add(
					ValidationSeverityError,
					"bucket_order_invalid",
					fieldPrefix+".start_time",
					"Buckets must be ordered by strictly increasing start time.",
				)
			}
			if point.StartTime.Before(
				previous.EndTime,
			) {
				collector.add(
					ValidationSeverityError,
					"bucket_overlap",
					fieldPrefix+".start_time",
					"Buckets must not overlap.",
				)
			}
		}
	}

	if result.Granularity == GranularityCustom &&
		len(result.Points) > 1 {
		collector.add(
			ValidationSeverityError,
			"custom_granularity_multiple_buckets",
			"points",
			"Custom granularity supports exactly one bucket.",
		)
	}
}

func validateBucketAlignment(
	point Point,
	granularity Granularity,
	fieldPrefix string,
	collector *validationCollector,
) {
	if point.StartTime.IsZero() ||
		point.EndTime.IsZero() {
		return
	}

	duration := point.EndTime.Sub(
		point.StartTime,
	)

	switch granularity {
	case GranularityHour:
		if !isHourBoundary(point.StartTime) ||
			duration != time.Hour {
			collector.add(
				ValidationSeverityError,
				"hour_bucket_misaligned",
				fieldPrefix,
				"Hourly buckets must begin on a UTC hour boundary and last exactly one hour.",
			)
		}

	case GranularityDay:
		if !isDayBoundary(point.StartTime) ||
			duration != 24*time.Hour {
			collector.add(
				ValidationSeverityError,
				"day_bucket_misaligned",
				fieldPrefix,
				"Daily buckets must begin at UTC midnight and last exactly twenty-four hours.",
			)
		}

	case GranularityWeek:
		if !isDayBoundary(point.StartTime) ||
			point.StartTime.Weekday() !=
				time.Monday ||
			duration != 7*24*time.Hour {
			collector.add(
				ValidationSeverityError,
				"week_bucket_misaligned",
				fieldPrefix,
				"Weekly buckets must begin on Monday at UTC midnight and last exactly seven days.",
			)
		}

	case GranularityCustom:
	}
}

func validatePointValue(
	point Point,
	fieldPrefix string,
	collector *validationCollector,
) {
	switch point.Status {
	case BucketStatusUnavailable,
		BucketStatusPartial,
		BucketStatusComplete:
	default:
		collector.add(
			ValidationSeverityError,
			"bucket_status_invalid",
			fieldPrefix+".status",
			"Bucket status must be unavailable, partial, or complete.",
		)
	}

	if !isFinite(point.Value) ||
		point.Value < 0 {
		collector.add(
			ValidationSeverityError,
			"bucket_value_invalid",
			fieldPrefix+".value",
			"Historical metric value must be a finite non-negative number.",
		)
	}
	if point.SampleCount < 0 {
		collector.add(
			ValidationSeverityError,
			"bucket_sample_count_invalid",
			fieldPrefix+".sample_count",
			"Bucket sample count must not be negative.",
		)
	}
	if !isRatio(point.CoverageRatio) {
		collector.add(
			ValidationSeverityError,
			"bucket_coverage_invalid",
			fieldPrefix+".coverage_ratio",
			"Bucket coverage ratio must be between zero and one.",
		)
	}

	switch point.Status {
	case BucketStatusUnavailable:
		if point.Value != 0 ||
			point.SampleCount != 0 ||
			point.CoverageRatio != 0 {
			collector.add(
				ValidationSeverityError,
				"unavailable_bucket_has_data",
				fieldPrefix,
				"Unavailable bucket must have zero value, samples, and coverage.",
			)
		}

	case BucketStatusPartial:
		if point.CoverageRatio <= 0 ||
			point.CoverageRatio >= 1 {
			collector.add(
				ValidationSeverityError,
				"partial_bucket_coverage_invalid",
				fieldPrefix+".coverage_ratio",
				"Partial bucket coverage must be greater than zero and less than one.",
			)
		}

	case BucketStatusComplete:
		if point.CoverageRatio != 1 {
			collector.add(
				ValidationSeverityError,
				"complete_bucket_coverage_invalid",
				fieldPrefix+".coverage_ratio",
				"Complete bucket coverage must equal one.",
			)
		}
	}
}

func validateSeriesStatus(
	result Result,
	collector *validationCollector,
) {
	switch result.Status {
	case SeriesStatusUnavailable:
		if len(result.Points) != 0 {
			collector.add(
				ValidationSeverityError,
				"unavailable_series_has_points",
				"points",
				"Unavailable series must not contain points.",
			)
		}
		if len(result.Limitations) == 0 {
			collector.add(
				ValidationSeverityWarning,
				"unavailable_series_without_limitation",
				"limitations",
				"Unavailable series should explain why historical data is unavailable.",
			)
		}

	case SeriesStatusPartial:
		if len(result.Points) == 0 {
			collector.add(
				ValidationSeverityError,
				"partial_series_without_points",
				"points",
				"Partial series must contain at least one point.",
			)
		}
		if isCompleteCoverage(result) {
			collector.add(
				ValidationSeverityError,
				"partial_series_is_complete",
				"status",
				"Partial series must not represent complete contiguous coverage.",
			)
		}
		if len(result.Limitations) == 0 {
			collector.add(
				ValidationSeverityWarning,
				"partial_series_without_limitation",
				"limitations",
				"Partial series should explain incomplete historical coverage.",
			)
		}

	case SeriesStatusComplete:
		if len(result.Points) == 0 {
			collector.add(
				ValidationSeverityError,
				"complete_series_without_points",
				"points",
				"Complete series must contain at least one point.",
			)
		}
		if !isCompleteCoverage(result) {
			collector.add(
				ValidationSeverityError,
				"complete_series_coverage_invalid",
				"points",
				"Complete series must cover the entire window with contiguous complete buckets.",
			)
		}
	}
}

func isCompleteCoverage(
	result Result,
) bool {
	if len(result.Points) == 0 ||
		result.Window.StartTime.IsZero() ||
		result.Window.EndTime.IsZero() {
		return false
	}

	if !result.Points[0].StartTime.Equal(
		result.Window.StartTime,
	) ||
		!result.Points[len(result.Points)-1].
			EndTime.Equal(
			result.Window.EndTime,
		) {
		return false
	}

	for index, point := range result.Points {
		if point.Status !=
			BucketStatusComplete {
			return false
		}
		if index > 0 &&
			!point.StartTime.Equal(
				result.Points[index-1].
					EndTime,
			) {
			return false
		}
	}

	return true
}

func validateSummary(
	result Result,
	collector *validationCollector,
) {
	expected := Summarize(result.Points)
	actual := result.Summary

	if actual.PointCount !=
		expected.PointCount {
		collector.add(
			ValidationSeverityError,
			"summary_point_count_mismatch",
			"summary.point_count",
			"Summary point count must match available and partial points.",
		)
	}

	validateSummaryValue(
		actual.Total,
		expected.Total,
		"summary.total",
		collector,
	)
	validateSummaryValue(
		actual.Minimum,
		expected.Minimum,
		"summary.minimum",
		collector,
	)
	validateSummaryValue(
		actual.Maximum,
		expected.Maximum,
		"summary.maximum",
		collector,
	)
	validateSummaryValue(
		actual.Average,
		expected.Average,
		"summary.average",
		collector,
	)
	validateSummaryValue(
		actual.Median,
		expected.Median,
		"summary.median",
		collector,
	)
}

func validateSummaryValue(
	actual float64,
	expected float64,
	field string,
	collector *validationCollector,
) {
	if !isFinite(actual) ||
		actual < 0 {
		collector.add(
			ValidationSeverityError,
			"summary_value_invalid",
			field,
			"Summary value must be a finite non-negative number.",
		)
		return
	}

	if !almostEqual(actual, expected) {
		collector.add(
			ValidationSeverityError,
			"summary_value_mismatch",
			field,
			"Summary value must match deterministic point aggregation.",
		)
	}
}

func validateComparison(
	result Result,
	collector *validationCollector,
) {
	comparison := result.Comparison
	if comparison == nil {
		return
	}

	validateTimeWindow(
		comparison.PreviousWindow,
		"comparison.previous_window",
		collector,
	)

	if !comparison.PreviousWindow.AsOfTime.IsZero() &&
		!result.Window.AsOfTime.IsZero() &&
		!comparison.PreviousWindow.AsOfTime.Equal(
			result.Window.AsOfTime,
		) {
		collector.add(
			ValidationSeverityError,
			"comparison_as_of_time_mismatch",
			"comparison.previous_window.as_of_time",
			"Previous and current periods must use the same analytical as-of time.",
		)
	}

	if comparison.PreviousWindow.Duration() !=
		result.Window.Duration() {
		collector.add(
			ValidationSeverityError,
			"comparison_duration_mismatch",
			"comparison.previous_window",
			"Previous and current periods must have equal duration.",
		)
	}

	if !comparison.PreviousWindow.EndTime.IsZero() &&
		!result.Window.StartTime.IsZero() &&
		!comparison.PreviousWindow.EndTime.Equal(
			result.Window.StartTime,
		) {
		collector.add(
			ValidationSeverityError,
			"comparison_period_not_adjacent",
			"comparison.previous_window.end_time",
			"Previous period must end exactly when the current period begins.",
		)
	}

	if !isFinite(comparison.CurrentValue) ||
		comparison.CurrentValue < 0 {
		collector.add(
			ValidationSeverityError,
			"comparison_current_value_invalid",
			"comparison.current_value",
			"Current comparison value must be a finite non-negative number.",
		)
	}
	if !isFinite(comparison.PreviousValue) ||
		comparison.PreviousValue < 0 {
		collector.add(
			ValidationSeverityError,
			"comparison_previous_value_invalid",
			"comparison.previous_value",
			"Previous comparison value must be a finite non-negative number.",
		)
	}

	expectedChange := comparison.CurrentValue -
		comparison.PreviousValue
	if !isFinite(comparison.AbsoluteChange) ||
		!almostEqual(
			comparison.AbsoluteChange,
			expectedChange,
		) {
		collector.add(
			ValidationSeverityError,
			"comparison_absolute_change_mismatch",
			"comparison.absolute_change",
			"Absolute change must equal current value minus previous value.",
		)
	}

	if comparison.PreviousValue == 0 {
		if comparison.PercentageChange != nil {
			collector.add(
				ValidationSeverityError,
				"comparison_percentage_change_undefined",
				"comparison.percentage_change",
				"Percentage change must be omitted when the previous value is zero.",
			)
		}
	} else {
		expectedPercentage := (expectedChange /
			comparison.PreviousValue) * 100

		if comparison.PercentageChange == nil ||
			!isFinite(
				*comparison.PercentageChange,
			) ||
			!almostEqual(
				*comparison.PercentageChange,
				expectedPercentage,
			) {
			collector.add(
				ValidationSeverityError,
				"comparison_percentage_change_mismatch",
				"comparison.percentage_change",
				"Percentage change must match the current and previous values.",
			)
		}
	}

	expectedDirection := TrendDirectionForChange(
		expectedChange,
	)
	if comparison.Direction !=
		expectedDirection {
		collector.add(
			ValidationSeverityError,
			"comparison_direction_mismatch",
			"comparison.direction",
			"Trend direction must match the absolute change.",
		)
	}
}

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

func totalSampleCount(
	points []Point,
) int {
	total := 0
	for _, point := range points {
		total += point.SampleCount
	}

	return total
}

func isSupportedMetricName(
	value MetricName,
) bool {
	for _, candidate := range supportedMetricNames {
		if candidate == value {
			return true
		}
	}

	return false
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
