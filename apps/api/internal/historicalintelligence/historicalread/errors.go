package historicalread

import (
	"errors"
	"fmt"
)

var (
	ErrPostgresPoolRequired = errors.New(
		"historical read postgres pool is required",
	)
	ErrPostgresExecutorRequired = errors.New(
		"historical read postgres executor is required",
	)
	ErrStartTimeRequired = errors.New(
		"historical read start time is required",
	)
	ErrEndTimeRequired = errors.New(
		"historical read end time is required",
	)
	ErrAsOfTimeRequired = errors.New(
		"historical read as-of time is required",
	)
	ErrWindowNotPositive = errors.New(
		"historical read start time must be before end time",
	)
	ErrWindowExceedsAsOfTime = errors.New(
		"historical read end time must not exceed as-of time",
	)
	ErrInvalidDatasetLimit = errors.New(
		"historical read dataset limit is invalid",
	)
)

type DatabaseError struct {
	Operation string
	Err       error
}

func (err *DatabaseError) Error() string {
	if err == nil {
		return "historical read database operation failed"
	}

	return fmt.Sprintf(
		"historical read %s: %v",
		err.Operation,
		err.Err,
	)
}

func (err *DatabaseError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Err
}
