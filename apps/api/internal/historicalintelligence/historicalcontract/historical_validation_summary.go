package historicalcontract

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
