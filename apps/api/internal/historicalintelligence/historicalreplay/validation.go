package historicalreplay

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func validateResult(
	result Result,
) error {
	if result.Version != Version {
		return resultFieldError(
			"version",
			ErrResultInvalid,
		)
	}
	switch result.Status {
	case StatusComplete,
		StatusPartial,
		StatusFailed:
	default:
		return resultFieldError(
			"status",
			ErrResultInvalid,
		)
	}

	if err := historicalwindow.ValidatePlan(
		result.Plan,
	); err != nil {
		return resultFieldError(
			"plan",
			errors.Join(
				ErrResultInvalid,
				err,
			),
		)
	}
	if result.Granularity !=
		result.Plan.Granularity {
		return resultFieldError(
			"granularity",
			ErrResultInvalid,
		)
	}
	specification, exists := historicalcontract.
		MetricSpecFor(result.MetricName)
	if !exists {
		return resultFieldError(
			"metric_name",
			ErrResultInvalid,
		)
	}
	normalizedScope, err := normalizeReplayScope(
		result.Scope,
	)
	if err != nil ||
		!normalizedScope.Equal(result.Scope) ||
		!specification.AllowsScope(
			result.Scope.Type,
		) {
		return resultFieldError(
			"scope",
			ErrResultInvalid,
		)
	}
	if result.DatasetLimit < 1 ||
		result.DatasetLimit >
			historicalread.MaximumDatasetLimit {
		return resultFieldError(
			"dataset_limit",
			ErrResultInvalid,
		)
	}
	if result.MaximumBucketCount < 1 ||
		result.MaximumBucketCount >
			historicalwindow.MaximumBucketCount {
		return resultFieldError(
			"maximum_bucket_count",
			ErrResultInvalid,
		)
	}
	if result.MaximumWindowCount < 1 ||
		result.MaximumWindowCount >
			MaximumWindowCount {
		return resultFieldError(
			"maximum_window_count",
			ErrResultInvalid,
		)
	}

	if result.PlannedWindowCount !=
		len(result.Plan.Buckets) {
		return resultFieldError(
			"planned_window_count",
			ErrResultInvalid,
		)
	}
	if result.CompletedWindowCount !=
		len(result.Windows) ||
		result.CompletedWindowCount >
			result.PlannedWindowCount {
		return resultFieldError(
			"completed_window_count",
			ErrResultInvalid,
		)
	}

	if err := validateResultTimes(result); err != nil {
		return err
	}
	if err := validateResultStatus(result); err != nil {
		return err
	}
	if err := validateResultWindows(result); err != nil {
		return err
	}
	if err := validateResultFailure(result); err != nil {
		return err
	}

	expectedFingerprint :=
		replayInputFingerprint(
			result.Plan,
			result.MetricName,
			result.Scope,
			result.Granularity,
			result.DatasetLimit,
			result.MaximumBucketCount,
			result.MaximumWindowCount,
			result.GeneratedAt,
		)
	if result.InputFingerprint !=
		expectedFingerprint {
		return resultFieldError(
			"input_fingerprint",
			ErrResultFingerprintMismatch,
		)
	}
	return nil
}

func validateResultTimes(
	result Result,
) error {
	required := []struct {
		name  string
		value time.Time
	}{
		{
			name:  "generated_at",
			value: result.GeneratedAt,
		},
		{
			name:  "started_at",
			value: result.StartedAt,
		},
		{
			name:  "completed_at",
			value: result.CompletedAt,
		},
	}
	for _, item := range required {
		if item.value.IsZero() ||
			item.value.Location() !=
				time.UTC {
			return resultFieldError(
				item.name,
				ErrResultInvalid,
			)
		}
	}
	if result.GeneratedAt.Before(
		result.Plan.AsOfTime,
	) ||
		result.GeneratedAt.After(
			result.StartedAt,
		) ||
		result.CompletedAt.Before(
			result.StartedAt,
		) {
		return resultFieldError(
			"timestamps",
			ErrResultInvalid,
		)
	}
	return nil
}

func validateResultStatus(
	result Result,
) error {
	switch result.Status {
	case StatusComplete:
		if result.PlannedWindowCount < 1 ||
			result.CompletedWindowCount !=
				result.PlannedWindowCount ||
			result.HasFailure {
			return resultFieldError(
				"status",
				ErrResultInvalid,
			)
		}

	case StatusPartial:
		if result.CompletedWindowCount < 1 ||
			result.CompletedWindowCount >=
				result.PlannedWindowCount ||
			!result.HasFailure {
			return resultFieldError(
				"status",
				ErrResultInvalid,
			)
		}

	case StatusFailed:
		if result.CompletedWindowCount != 0 ||
			!result.HasFailure {
			return resultFieldError(
				"status",
				ErrResultInvalid,
			)
		}
	}
	return nil
}

func validateResultWindows(
	result Result,
) error {
	previousCurrentFingerprint := ""
	for index, window := range result.Windows {
		if index >= len(result.Plan.Buckets) ||
			!bucketsEqual(
				window.Bucket,
				result.Plan.Buckets[index],
			) {
			return resultFieldError(
				fmt.Sprintf(
					"windows[%d].bucket",
					index,
				),
				ErrResultInvalid,
			)
		}
		if !validFingerprint(
			window.PreviousPeriodInputFingerprint,
		) ||
			!validFingerprint(
				window.CurrentPeriodInputFingerprint,
			) {
			return resultFieldError(
				fmt.Sprintf(
					"windows[%d].period_fingerprint",
					index,
				),
				ErrResultFingerprintMismatch,
			)
		}
		if previousCurrentFingerprint != "" &&
			window.PreviousPeriodInputFingerprint !=
				previousCurrentFingerprint {
			return resultFieldError(
				fmt.Sprintf(
					"windows[%d].previous_period_input_fingerprint",
					index,
				),
				ErrReplayContinuityMismatch,
			)
		}
		previousCurrentFingerprint =
			window.CurrentPeriodInputFingerprint

		expectedWindow := bucketWindow(
			window.Bucket,
			result.Plan.AsOfTime,
		)
		if !recordKeyMatchesResult(
			window.Record,
			result,
			expectedWindow,
		) {
			return resultFieldError(
				fmt.Sprintf(
					"windows[%d].record",
					index,
				),
				ErrResultInvalid,
			)
		}
		report := historicalcontract.Validate(
			window.Record.Result,
		)
		if report.Status !=
			historicalcontract.
				ValidationStatusValid {
			return resultFieldError(
				fmt.Sprintf(
					"windows[%d].record.result",
					index,
				),
				ErrOutcomeResultInvalid,
			)
		}
	}
	return nil
}

func validateResultFailure(
	result Result,
) error {
	if !result.HasFailure {
		if result.Failure !=
			(Failure{}) {
			return resultFieldError(
				"failure",
				ErrResultInvalid,
			)
		}
		return nil
	}
	if strings.TrimSpace(
		string(result.Failure.Code),
	) == "" ||
		strings.TrimSpace(
			result.Failure.Message,
		) == "" {
		return resultFieldError(
			"failure",
			ErrResultInvalid,
		)
	}
	switch result.Failure.Code {
	case FailureCodeNoReplayWindow,
		FailureCodeContextCanceled,
		FailureCodeMaterialization,
		FailureCodeOutcomeContract,
		FailureCodeContinuityMismatch:
	default:
		return resultFieldError(
			"failure.code",
			ErrResultInvalid,
		)
	}
	if result.Failure.Sequence == 0 {
		if result.Failure.Code !=
			FailureCodeNoReplayWindow ||
			!result.Failure.StartTime.IsZero() ||
			!result.Failure.EndTime.IsZero() {
			return resultFieldError(
				"failure",
				ErrResultInvalid,
			)
		}
		return nil
	}
	expectedSequence :=
		result.CompletedWindowCount + 1
	if result.Failure.Sequence !=
		expectedSequence ||
		expectedSequence >
			len(result.Plan.Buckets) {
		return resultFieldError(
			"failure.sequence",
			ErrResultInvalid,
		)
	}
	expected := result.Plan.Buckets[expectedSequence-1]
	if result.Failure.StartTime.Location() !=
		time.UTC ||
		result.Failure.EndTime.Location() !=
			time.UTC ||
		!result.Failure.StartTime.Equal(
			expected.StartTime,
		) ||
		!result.Failure.EndTime.Equal(
			expected.EndTime,
		) {
		return resultFieldError(
			"failure.window",
			ErrResultInvalid,
		)
	}
	return nil
}

func recordKeyMatchesResult(
	record historicalaggregate.Record,
	result Result,
	window historicalcontract.TimeWindow,
) bool {
	return strings.TrimSpace(record.ID) != "" &&
		validFingerprint(
			record.InputFingerprint,
		) &&
		record.InputFingerprint ==
			record.Result.Provenance.
				InputFingerprint &&
		record.Key.SchemaVersion ==
			record.Result.SchemaVersion &&
		record.Key.MetricName ==
			result.MetricName &&
		record.Key.Scope.Equal(
			result.Scope,
		) &&
		record.Key.Granularity ==
			result.Granularity &&
		windowsEqual(
			record.Key.Window,
			window,
		) &&
		record.Result.Metric.Name ==
			result.MetricName &&
		record.Result.Scope.Equal(
			result.Scope,
		) &&
		record.Result.Granularity ==
			result.Granularity &&
		record.Result.GeneratedAt.Location() ==
			time.UTC &&
		record.Result.GeneratedAt.Equal(
			result.GeneratedAt,
		) &&
		windowsEqual(
			record.Result.Window,
			window,
		) &&
		record.Result.Comparison != nil &&
		record.StoredAt.Location() ==
			time.UTC &&
		!record.StoredAt.Before(
			record.Result.GeneratedAt,
		)
}

func resultFieldError(
	field string,
	err error,
) error {
	return &ResultContractError{
		Field: field,
		Err:   err,
	}
}
