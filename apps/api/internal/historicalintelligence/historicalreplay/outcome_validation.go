package historicalreplay

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func validateOutcome(
	bucket historicalwindow.Bucket,
	request normalizedRequest,
	outcome historicalmaterialization.Outcome,
) error {
	if outcome.Version !=
		historicalmaterialization.Version {
		return outcomeFieldError(
			"version",
			ErrOutcomeVersionMismatch,
		)
	}
	if err := validateOutcomePlans(
		bucket,
		request,
		outcome,
	); err != nil {
		return err
	}
	if err := validateOutcomeResults(
		bucket,
		request,
		outcome,
	); err != nil {
		return err
	}
	if err := validateOutcomeRecord(
		bucket,
		request,
		outcome,
	); err != nil {
		return err
	}
	return nil
}

func validateOutcomePlans(
	bucket historicalwindow.Bucket,
	request normalizedRequest,
	outcome historicalmaterialization.Outcome,
) error {
	expectedCurrentWindow := bucketWindow(
		bucket,
		request.AsOfTime,
	)
	if err := validateSingleBucketPlan(
		"plan",
		outcome.Plan,
		bucket,
		expectedCurrentWindow,
		request.Granularity,
	); err != nil {
		return err
	}
	if outcome.Plan.PreviousWindow == nil {
		return outcomeFieldError(
			"plan.previous_window",
			ErrOutcomePlanMismatch,
		)
	}
	expectedPreviousWindow :=
		*outcome.Plan.PreviousWindow
	if !windowsEqual(
		expectedPreviousWindow,
		outcome.PreviousResult.Window,
	) {
		return outcomeFieldError(
			"plan.previous_window",
			ErrOutcomePlanMismatch,
		)
	}
	previousBucket := historicalwindow.Bucket{
		Sequence:  1,
		StartTime: expectedPreviousWindow.StartTime,
		EndTime:   expectedPreviousWindow.EndTime,
	}
	if err := validateSingleBucketPlan(
		"previous_plan",
		outcome.PreviousPlan,
		previousBucket,
		expectedPreviousWindow,
		request.Granularity,
	); err != nil {
		return err
	}

	if err := validateReadSummary(
		"read_summaries.previous",
		outcome.ReadSummaries.Previous,
		expectedPreviousWindow,
		request.DatasetLimit,
	); err != nil {
		return err
	}
	if err := validateReadSummary(
		"read_summaries.current",
		outcome.ReadSummaries.Current,
		expectedCurrentWindow,
		request.DatasetLimit,
	); err != nil {
		return err
	}
	if outcome.ReadSummaries.Previous.
		IsolationLevel !=
		outcome.ReadSummaries.Current.
			IsolationLevel {
		return outcomeFieldError(
			"read_summaries.isolation_level",
			ErrOutcomePlanMismatch,
		)
	}
	return nil
}

func validateSingleBucketPlan(
	field string,
	plan historicalwindow.Plan,
	expectedBucket historicalwindow.Bucket,
	expectedWindow historicalcontract.TimeWindow,
	granularity historicalcontract.Granularity,
) error {
	if err := historicalwindow.ValidatePlan(plan); err != nil {
		return outcomeFieldError(
			field,
			errors.Join(
				ErrOutcomePlanMismatch,
				err,
			),
		)
	}
	if plan.RequestedStartTime.Location() != time.UTC ||
		plan.RequestedEndTime.Location() != time.UTC ||
		plan.AsOfTime.Location() != time.UTC ||
		!plan.RequestedStartTime.Equal(
			expectedWindow.StartTime,
		) ||
		!plan.RequestedEndTime.Equal(
			expectedWindow.EndTime,
		) ||
		!plan.AsOfTime.Equal(
			expectedWindow.AsOfTime,
		) ||
		plan.Granularity != granularity ||
		plan.MaximumBucketCount != 1 ||
		plan.EffectiveWindow == nil ||
		!windowsEqual(
			*plan.EffectiveWindow,
			expectedWindow,
		) ||
		len(plan.Buckets) != 1 ||
		!periodBucketMatches(
			plan.Buckets[0],
			expectedBucket,
		) {
		return outcomeFieldError(
			field,
			ErrOutcomePlanMismatch,
		)
	}
	return nil
}

func validateReadSummary(
	field string,
	summary historicalmaterialization.ReadSummary,
	expectedWindow historicalcontract.TimeWindow,
	datasetLimit int,
) error {
	if !windowsEqual(summary.Window, expectedWindow) ||
		summary.DatasetLimit != datasetLimit {
		return outcomeFieldError(
			field,
			ErrOutcomePlanMismatch,
		)
	}
	switch summary.IsolationLevel {
	case historicalread.SnapshotIsolationRepeatableRead,
		historicalread.SnapshotIsolationCallerTransaction:
		return nil
	default:
		return outcomeFieldError(
			field+".isolation_level",
			ErrOutcomePlanMismatch,
		)
	}
}

func validateOutcomeResults(
	bucket historicalwindow.Bucket,
	request normalizedRequest,
	outcome historicalmaterialization.Outcome,
) error {
	expectedCurrentWindow := bucketWindow(
		bucket,
		request.AsOfTime,
	)
	if err := validateResultContract(
		"current_period_result",
		outcome.CurrentPeriodResult,
	); err != nil {
		return err
	}
	if err := validateResultContract(
		"previous_result",
		outcome.PreviousResult,
	); err != nil {
		return err
	}
	if err := validateResultContract(
		"current_result",
		outcome.CurrentResult,
	); err != nil {
		return err
	}

	if outcome.CurrentPeriodResult.
		Comparison != nil {
		return outcomeFieldError(
			"current_period_result.comparison",
			ErrOutcomeComparisonRequired,
		)
	}
	if outcome.PreviousResult.Comparison != nil {
		return outcomeFieldError(
			"previous_result.comparison",
			ErrOutcomeComparisonRequired,
		)
	}
	if outcome.CurrentResult.Comparison == nil {
		return outcomeFieldError(
			"current_result.comparison",
			ErrOutcomeComparisonRequired,
		)
	}

	if !resultMatchesRequest(
		outcome.CurrentPeriodResult,
		request,
		expectedCurrentWindow,
	) {
		return outcomeFieldError(
			"current_period_result.identity",
			ErrOutcomeRecordMismatch,
		)
	}
	if !resultMatchesRequest(
		outcome.CurrentResult,
		request,
		expectedCurrentWindow,
	) {
		return outcomeFieldError(
			"current_result.identity",
			ErrOutcomeRecordMismatch,
		)
	}

	previousWindow := outcome.PreviousResult.Window
	if !previousWindow.EndTime.UTC().Equal(
		bucket.StartTime.UTC(),
	) ||
		!previousWindow.AsOfTime.UTC().Equal(
			request.AsOfTime.UTC(),
		) ||
		previousWindow.Duration() !=
			expectedCurrentWindow.Duration() ||
		outcome.PreviousResult.Metric.Name !=
			request.MetricName ||
		!outcome.PreviousResult.Scope.Equal(
			request.Scope,
		) ||
		outcome.PreviousResult.Granularity !=
			request.Granularity ||
		!outcome.PreviousResult.GeneratedAt.UTC().
			Equal(request.GeneratedAt) {
		return outcomeFieldError(
			"previous_result.identity",
			ErrOutcomeRecordMismatch,
		)
	}

	if !windowsEqual(
		outcome.CurrentResult.Comparison.
			PreviousWindow,
		previousWindow,
	) {
		return outcomeFieldError(
			"current_result.comparison.previous_window",
			ErrOutcomeComparisonRequired,
		)
	}
	if outcome.CurrentResult.SchemaVersion !=
		outcome.CurrentPeriodResult.SchemaVersion ||
		outcome.CurrentResult.Metric !=
			outcome.CurrentPeriodResult.Metric ||
		!outcome.CurrentResult.Scope.Equal(
			outcome.CurrentPeriodResult.Scope,
		) ||
		outcome.CurrentResult.Granularity !=
			outcome.CurrentPeriodResult.Granularity ||
		!windowsEqual(
			outcome.CurrentResult.Window,
			outcome.CurrentPeriodResult.Window,
		) ||
		outcome.CurrentResult.Summary !=
			outcome.CurrentPeriodResult.Summary ||
		!outcome.CurrentResult.GeneratedAt.Equal(
			outcome.CurrentPeriodResult.GeneratedAt,
		) {
		return outcomeFieldError(
			"current_result.current_period_identity",
			ErrOutcomeRecordMismatch,
		)
	}
	previousValue, available := comparisonValue(
		outcome.PreviousResult.Metric,
		outcome.PreviousResult.Summary,
	)
	if !available || !metricValuesMatch(
		outcome.PreviousResult.Metric.Name,
		outcome.CurrentResult.Comparison.
			PreviousValue,
		previousValue,
	) {
		return outcomeFieldError(
			"current_result.comparison.previous_value",
			ErrOutcomeComparisonRequired,
		)
	}
	if !validFingerprint(
		outcome.CurrentPeriodResult.
			Provenance.InputFingerprint,
	) ||
		!validFingerprint(
			outcome.PreviousResult.
				Provenance.InputFingerprint,
		) ||
		!validFingerprint(
			outcome.CurrentResult.
				Provenance.InputFingerprint,
		) {
		return outcomeFieldError(
			"result.provenance.input_fingerprint",
			ErrOutcomeFingerprintMismatch,
		)
	}
	return nil
}

func comparisonValue(
	metric historicalcontract.Metric,
	summary historicalcontract.Summary,
) (float64, bool) {
	switch metric.Aggregation {
	case historicalcontract.AggregationCount,
		historicalcontract.AggregationSum:
		return summary.Total, true
	case historicalcontract.AggregationMinimum:
		return summary.Minimum, true
	case historicalcontract.AggregationMaximum:
		return summary.Maximum, true
	case historicalcontract.AggregationAverage,
		historicalcontract.AggregationRatio:
		return summary.Average, true
	case historicalcontract.AggregationMedian:
		return summary.Median, true
	default:
		return 0, false
	}
}

func metricValuesMatch(
	metricName historicalcontract.MetricName,
	left float64,
	right float64,
) bool {
	if math.IsNaN(left) || math.IsInf(left, 0) ||
		math.IsNaN(right) || math.IsInf(right, 0) {
		return false
	}
	specification, exists :=
		historicalcontract.MetricSpecFor(
			metricName,
		)
	if !exists {
		return false
	}
	switch specification.ValueKind {
	case historicalcontract.MetricValueKindCount:
		return left == right
	case historicalcontract.MetricValueKindRatio:
		return math.Abs(left-right) <= 1e-12
	default:
		difference := math.Abs(left - right)
		scale := math.Max(
			1,
			math.Max(
				math.Abs(left),
				math.Abs(right),
			),
		)
		return difference <= 1e-9*scale
	}
}

func validateOutcomeRecord(
	bucket historicalwindow.Bucket,
	request normalizedRequest,
	outcome historicalmaterialization.Outcome,
) error {
	record := outcome.Record
	expectedWindow := bucketWindow(
		bucket,
		request.AsOfTime,
	)
	if strings.TrimSpace(record.ID) == "" {
		return outcomeFieldError(
			"record.id",
			ErrOutcomeRecordMismatch,
		)
	}
	if !validFingerprint(
		record.InputFingerprint,
	) ||
		record.InputFingerprint !=
			record.Result.Provenance.
				InputFingerprint ||
		record.InputFingerprint !=
			outcome.CurrentResult.Provenance.
				InputFingerprint {
		return outcomeFieldError(
			"record.input_fingerprint",
			ErrOutcomeFingerprintMismatch,
		)
	}
	if !recordKeyMatchesRequest(
		record.Key,
		request,
		expectedWindow,
	) {
		return outcomeFieldError(
			"record.key",
			ErrOutcomeRecordMismatch,
		)
	}
	if !reflect.DeepEqual(
		record.Result,
		outcome.CurrentResult,
	) {
		return outcomeFieldError(
			"record.result",
			ErrOutcomeRecordMismatch,
		)
	}
	if record.StoredAt.IsZero() ||
		record.StoredAt.Location() !=
			time.UTC ||
		record.StoredAt.Before(
			record.Result.GeneratedAt,
		) {
		return outcomeFieldError(
			"record.stored_at",
			ErrOutcomeRecordMismatch,
		)
	}
	return nil
}

func validateResultContract(
	field string,
	result historicalcontract.Result,
) error {
	report := historicalcontract.Validate(result)
	if report.Status !=
		historicalcontract.ValidationStatusValid {
		return outcomeFieldError(
			field,
			errors.Join(
				ErrOutcomeResultInvalid,
				&historicalaggregate.
					ValidationError{
					Report: report.Clone(),
				},
			),
		)
	}
	return nil
}

func resultMatchesRequest(
	result historicalcontract.Result,
	request normalizedRequest,
	window historicalcontract.TimeWindow,
) bool {
	return result.Metric.Name ==
		request.MetricName &&
		result.Scope.Equal(request.Scope) &&
		result.Granularity ==
			request.Granularity &&
		windowsEqual(
			result.Window,
			window,
		) &&
		result.GeneratedAt.Location() ==
			time.UTC &&
		result.GeneratedAt.Equal(
			request.GeneratedAt,
		)
}

func recordKeyMatchesRequest(
	key historicalaggregate.ResultKey,
	request normalizedRequest,
	window historicalcontract.TimeWindow,
) bool {
	return key.SchemaVersion ==
		historicalcontract.SchemaVersionV1 &&
		key.MetricName == request.MetricName &&
		key.Scope.Equal(request.Scope) &&
		key.Granularity ==
			request.Granularity &&
		windowsEqual(key.Window, window)
}

func outcomeFieldError(
	field string,
	err error,
) error {
	return &OutcomeContractError{
		Field: field,
		Err:   err,
	}
}

func bucketWindow(
	bucket historicalwindow.Bucket,
	asOfTime time.Time,
) historicalcontract.TimeWindow {
	return historicalcontract.TimeWindow{
		StartTime: bucket.StartTime.UTC(),
		EndTime:   bucket.EndTime.UTC(),
		AsOfTime:  asOfTime.UTC(),
	}
}

func windowsEqual(
	left historicalcontract.TimeWindow,
	right historicalcontract.TimeWindow,
) bool {
	return left.StartTime.Location() ==
		time.UTC &&
		left.EndTime.Location() ==
			time.UTC &&
		left.AsOfTime.Location() ==
			time.UTC &&
		left.StartTime.Equal(
			right.StartTime,
		) &&
		left.EndTime.Equal(
			right.EndTime,
		) &&
		left.AsOfTime.Equal(
			right.AsOfTime,
		)
}

func periodBucketMatches(
	actual historicalwindow.Bucket,
	expected historicalwindow.Bucket,
) bool {
	return actual.Sequence == 1 &&
		strings.TrimSpace(actual.Key) != "" &&
		actual.StartTime.Location() == time.UTC &&
		actual.EndTime.Location() == time.UTC &&
		actual.StartTime.Equal(expected.StartTime) &&
		actual.EndTime.Equal(expected.EndTime)
}

func bucketsEqual(
	left historicalwindow.Bucket,
	right historicalwindow.Bucket,
) bool {
	return left.Key == right.Key &&
		left.Sequence == right.Sequence &&
		left.StartTime.Location() ==
			time.UTC &&
		left.EndTime.Location() ==
			time.UTC &&
		left.StartTime.Equal(
			right.StartTime,
		) &&
		left.EndTime.Equal(
			right.EndTime,
		)
}

func validFingerprint(
	value string,
) bool {
	if !strings.HasPrefix(
		value,
		"sha256:",
	) || len(value) != 71 {
		return false
	}
	for _, character := range value[7:] {
		switch {
		case character >= '0' &&
			character <= '9':
		case character >= 'a' &&
			character <= 'f':
		default:
			return false
		}
	}
	return true
}
