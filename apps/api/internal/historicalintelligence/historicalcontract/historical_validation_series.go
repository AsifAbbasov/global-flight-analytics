package historicalcontract

import (
	"fmt"

	"time"
)

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

func totalSampleCount(
	points []Point,
) int {
	total := 0
	for _, point := range points {
		total += point.SampleCount
	}

	return total
}
