package historicalaggregate

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregatecontract"
)

var (
	ErrPostgresPoolRequired = errors.New(
		"historical aggregate PostgreSQL pool is required",
	)
	ErrPostgresExecutorRequired = errors.New(
		"historical aggregate PostgreSQL executor is required",
	)

	ErrUnsupportedSchemaVersion = historicalaggregatecontract.
					ErrUnsupportedSchemaVersion
	ErrInputFingerprintRequired = historicalaggregatecontract.
					ErrInputFingerprintRequired
	ErrInvalidListLimit = historicalaggregatecontract.
				ErrInvalidListLimit
	ErrInvalidListCursor = historicalaggregatecontract.
				ErrInvalidListCursor
	ErrResultNotFound = historicalaggregatecontract.
				ErrResultNotFound
	ErrResultConflict = historicalaggregatecontract.
				ErrResultConflict
	ErrScopeInvalid = historicalaggregatecontract.
			ErrScopeInvalid
	ErrWindowRequired = historicalaggregatecontract.
				ErrWindowRequired
	ErrCorruptResult = errors.New(
		"stored historical aggregate result is internally inconsistent",
	)
)

type ValidationError = historicalaggregatecontract.ValidationError

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
