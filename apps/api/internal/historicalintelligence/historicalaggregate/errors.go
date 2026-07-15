package historicalaggregate

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

var (
	ErrPostgresPoolRequired = errors.New(
		"historical aggregate PostgreSQL pool is required",
	)
	ErrPostgresExecutorRequired = errors.New(
		"historical aggregate PostgreSQL executor is required",
	)
	ErrUnsupportedSchemaVersion = errors.New(
		"historical aggregate schema version is unsupported",
	)
	ErrInputFingerprintRequired = errors.New(
		"historical aggregate input fingerprint is required",
	)
	ErrInvalidListLimit = errors.New(
		"historical aggregate list limit is invalid",
	)
	ErrResultNotFound = errors.New(
		"historical aggregate result was not found",
	)
	ErrResultConflict = errors.New(
		"historical aggregate result key already exists with a different input fingerprint",
	)
	ErrScopeInvalid = errors.New(
		"historical aggregate scope is invalid",
	)
	ErrWindowRequired = errors.New(
		"historical aggregate window is required",
	)
)

type ValidationError struct {
	Report historicalcontract.ValidationReport
}

func (err *ValidationError) Error() string {
	if err == nil {
		return "historical aggregate validation failed"
	}

	return fmt.Sprintf(
		"historical aggregate validation failed: errors=%d warnings=%d",
		err.Report.ErrorCount,
		err.Report.WarningCount,
	)
}

type DatabaseError struct {
	Operation string
	Err       error
}

func (err *DatabaseError) Error() string {
	if err == nil {
		return "historical aggregate database operation failed"
	}

	return fmt.Sprintf(
		"historical aggregate database operation %s failed: %v",
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
