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
	ErrResultNotFound = historicalaggregatecontract.
				ErrResultNotFound
	ErrResultConflict = historicalaggregatecontract.
				ErrResultConflict
	ErrScopeInvalid = historicalaggregatecontract.
			ErrScopeInvalid
	ErrWindowRequired = historicalaggregatecontract.
				ErrWindowRequired
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
