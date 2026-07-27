package historicalcontract

import "fmt"

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
	scope Scope,
	collector *validationCollector,
) {
	specification, exists := MetricSpecFor(metric.Name)
	if !exists {
		collector.add(
			ValidationSeverityError,
			"metric_name_unsupported",
			"metric.name",
			"Metric name is not materializable by historical-intelligence-v1.",
		)
		return
	}
	if metric.Unit != specification.Unit {
		collector.add(
			ValidationSeverityError,
			"metric_unit_mismatch",
			"metric.unit",
			fmt.Sprintf("Metric unit %q does not match catalog unit %q.", metric.Unit, specification.Unit),
		)
	}
	if metric.Aggregation != specification.Aggregation {
		collector.add(
			ValidationSeverityError,
			"metric_aggregation_mismatch",
			"metric.aggregation",
			fmt.Sprintf("Metric aggregation %q does not match catalog aggregation %q.", metric.Aggregation, specification.Aggregation),
		)
	}
	if !specification.AllowsScope(scope.Type) {
		collector.add(
			ValidationSeverityError,
			"metric_scope_mismatch",
			"scope.type",
			fmt.Sprintf("Metric %q does not support scope %q.", metric.Name, scope.Type),
		)
	}
}

func isSupportedMetricName(
	value MetricName,
) bool {
	_, exists := MetricSpecFor(value)
	return exists
}
