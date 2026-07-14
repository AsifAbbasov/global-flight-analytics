package historicalwindow

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

var (
	ErrStartTimeRequired = errors.New(
		"historical window start time is required",
	)
	ErrEndTimeRequired = errors.New(
		"historical window end time is required",
	)
	ErrAsOfTimeRequired = errors.New(
		"historical window as-of time is required",
	)
	ErrWindowNotPositive = errors.New(
		"historical window start time must be before end time",
	)
	ErrUnsupportedGranularity = errors.New(
		"historical window granularity is unsupported",
	)
	ErrInvalidMaximumBucketCount = errors.New(
		"historical window maximum bucket count is invalid",
	)
	ErrBucketCountExceeded = errors.New(
		"historical bucket count exceeds the configured maximum",
	)
	ErrBoundarySequenceInvalid = errors.New(
		"historical window boundary sequence is invalid",
	)
)

type BucketCountExceededError struct {
	Granularity historicalcontract.Granularity
	Count       int
	Maximum     int
}

func (err *BucketCountExceededError) Error() string {
	if err == nil {
		return "historical bucket count exceeds the configured maximum"
	}

	return fmt.Sprintf(
		"historical %s bucket count %d exceeds maximum %d",
		err.Granularity,
		err.Count,
		err.Maximum,
	)
}

func (err *BucketCountExceededError) Unwrap() error {
	if err == nil {
		return nil
	}

	return ErrBucketCountExceeded
}
