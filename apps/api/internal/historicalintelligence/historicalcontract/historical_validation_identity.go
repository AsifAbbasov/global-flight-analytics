package historicalcontract

import "strings"

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
