package historicalread

import (
	"errors"
	"fmt"
)

var (
	ErrContextRequired = errors.New(
		"historical read context is required",
	)
	ErrPostgresPoolRequired = errors.New(
		"historical read postgres pool is required",
	)
	ErrPostgresTransactionRequired = errors.New(
		"historical read postgres transaction is required",
	)
	ErrPostgresClientRequired = errors.New(
		"historical read postgres client is required",
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
	ErrInvalidRoutePayloadByteLimit = errors.New(
		"historical read route payload byte limit is invalid",
	)
	ErrTemporalHistoryUnavailable = errors.New(
		"historical read temporal history is unavailable for the requested as-of time",
	)
	ErrSnapshotMetadataInvalid = errors.New(
		"historical read snapshot metadata is invalid",
	)
	ErrRecordInvalid = errors.New(
		"historical read record is invalid",
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

type RecordValidationError struct {
	Dataset string
	Index   int
	Reason  string
}

func (err *RecordValidationError) Error() string {
	if err == nil {
		return "historical read record is invalid"
	}
	return fmt.Sprintf(
		"historical read %s record %d is invalid: %s",
		err.Dataset,
		err.Index,
		err.Reason,
	)
}

func (err *RecordValidationError) Unwrap() error {
	if err == nil {
		return nil
	}
	return ErrRecordInvalid
}
