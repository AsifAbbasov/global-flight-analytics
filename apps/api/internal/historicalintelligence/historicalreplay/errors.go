package historicalreplay

import (
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

var (
	ErrContextRequired = errors.New(
		"historical replay context is required",
	)
	ErrMaterializerRequired = errors.New(
		"historical replay materializer is required",
	)
	ErrClockRequired = errors.New(
		"historical replay clock must return a non-zero time",
	)
	ErrMaximumWindowCountInvalid = errors.New(
		"historical replay maximum window count is invalid",
	)
	ErrMaximumBucketCountInvalid = errors.New(
		"historical replay maximum bucket count is invalid",
	)
	ErrNoReplayWindow = errors.New(
		"historical replay contains no complete window",
	)
	ErrMetricUnsupported = errors.New(
		"historical replay metric is unsupported",
	)
	ErrScopeUnsupported = errors.New(
		"historical replay scope is unsupported",
	)
	ErrDatasetLimitInvalid = errors.New(
		"historical replay dataset limit is invalid",
	)
	ErrGeneratedAtBeforeAsOfTime = errors.New(
		"historical replay generation time must not precede the analytical as-of time",
	)
	ErrGeneratedAtAfterStartTime = errors.New(
		"historical replay generation time must not be after replay start time",
	)
	ErrOutcomeVersionMismatch = errors.New(
		"historical replay materialization outcome version is invalid",
	)
	ErrOutcomePlanMismatch = errors.New(
		"historical replay materialization outcome plan is invalid",
	)
	ErrOutcomeRecordMismatch = errors.New(
		"historical replay materialization record is invalid",
	)
	ErrOutcomeResultInvalid = errors.New(
		"historical replay materialization result is invalid",
	)
	ErrOutcomeComparisonRequired = errors.New(
		"historical replay materialization result requires a period comparison",
	)
	ErrOutcomeFingerprintMismatch = errors.New(
		"historical replay materialization fingerprint is invalid",
	)
	ErrReplayContinuityMismatch = errors.New(
		"historical replay adjacent materializations disagree about their shared period",
	)
	ErrResultInvalid = errors.New(
		"historical replay result is invalid",
	)
	ErrResultFingerprintMismatch = errors.New(
		"historical replay input fingerprint is invalid",
	)
)

type MetricScopeError struct {
	Metric historicalcontract.MetricName
	Scope  historicalcontract.Scope
}

func (err *MetricScopeError) Error() string {
	if err == nil {
		return "historical replay metric scope is unsupported"
	}
	return fmt.Sprintf(
		"historical replay metric %q does not support scope %q",
		err.Metric,
		err.Scope.Type,
	)
}

func (err *MetricScopeError) Unwrap() error {
	if err == nil {
		return nil
	}
	return ErrScopeUnsupported
}

type WindowCountExceededError struct {
	Count   int
	Maximum int
}

func (err *WindowCountExceededError) Error() string {
	if err == nil {
		return "historical replay window count exceeded"
	}
	return fmt.Sprintf(
		"historical replay window count %d exceeds maximum %d",
		err.Count,
		err.Maximum,
	)
}

type WindowError struct {
	Sequence  int
	StartTime time.Time
	EndTime   time.Time
	Err       error
}

func (err *WindowError) Error() string {
	if err == nil {
		return "historical replay window failed"
	}
	return fmt.Sprintf(
		"historical replay window %d [%s,%s) failed: %v",
		err.Sequence,
		err.StartTime.UTC().Format(time.RFC3339),
		err.EndTime.UTC().Format(time.RFC3339),
		err.Err,
	)
}

func (err *WindowError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

type OutcomeContractError struct {
	Field string
	Err   error
}

func (err *OutcomeContractError) Error() string {
	if err == nil {
		return "historical replay materialization outcome contract failed"
	}
	return fmt.Sprintf(
		"historical replay materialization outcome field %s failed: %v",
		err.Field,
		err.Err,
	)
}

func (err *OutcomeContractError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

type ResultContractError struct {
	Field string
	Err   error
}

func (err *ResultContractError) Error() string {
	if err == nil {
		return "historical replay result contract failed"
	}
	return fmt.Sprintf(
		"historical replay result field %s failed: %v",
		err.Field,
		err.Err,
	)
}

func (err *ResultContractError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}
