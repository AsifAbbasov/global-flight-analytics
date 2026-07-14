package routestore

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

var (
	ErrTrajectoryIDRequired = errors.New(
		"route result trajectory id is required",
	)
	ErrUnsupportedSchemaVersion = errors.New(
		"route result schema version is unsupported",
	)
	ErrAsOfTimeRequired = errors.New(
		"route result as-of time is required",
	)
	ErrInputFingerprintRequired = errors.New(
		"route result input fingerprint is required",
	)
	ErrResultInvalid = errors.New(
		"invalid route result cannot be stored",
	)
	ErrResultNotFound = errors.New(
		"route result was not found",
	)
	ErrResultConflict = errors.New(
		"route result key already exists with different evidence",
	)
	ErrInvalidListLimit = errors.New(
		"route result list limit must be between one and one hundred",
	)
	ErrPostgresPoolRequired = errors.New(
		"route result postgres pool is required",
	)
	ErrPostgresExecutorRequired = errors.New(
		"route result postgres executor is required",
	)
	ErrInvalidTrajectoryID = errors.New(
		"route result trajectory id must be a valid UUID",
	)
	ErrCorruptResult = errors.New(
		"stored route result is internally inconsistent",
	)
)

type ValidationError struct {
	Report routecontract.ValidationReport
}

func (err *ValidationError) Error() string {
	if err == nil {
		return ErrResultInvalid.Error()
	}

	return fmt.Sprintf(
		"%s: %d error(s), %d warning(s)",
		ErrResultInvalid,
		err.Report.ErrorCount,
		err.Report.WarningCount,
	)
}

func (err *ValidationError) Unwrap() error {
	return ErrResultInvalid
}

type DatabaseError struct {
	Operation string
	Err       error
}

func (err *DatabaseError) Error() string {
	if err == nil {
		return "route result database operation failed"
	}

	return fmt.Sprintf(
		"route result database operation %s failed: %v",
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

type CorruptResultError struct {
	Field string
}

func (err *CorruptResultError) Error() string {
	if err == nil {
		return ErrCorruptResult.Error()
	}

	return fmt.Sprintf(
		"%s: field=%s",
		ErrCorruptResult,
		err.Field,
	)
}

func (err *CorruptResultError) Unwrap() error {
	return ErrCorruptResult
}
