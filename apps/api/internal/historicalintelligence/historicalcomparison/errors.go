package historicalcomparison

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

var (
	ErrCurrentResultInvalid = errors.New(
		"current historical result is invalid",
	)
	ErrPreviousResultInvalid = errors.New(
		"previous historical result is invalid",
	)
	ErrSchemaMismatch = errors.New(
		"historical comparison schema versions do not match",
	)
	ErrMetricMismatch = errors.New(
		"historical comparison metrics do not match",
	)
	ErrScopeMismatch = errors.New(
		"historical comparison scopes do not match",
	)
	ErrGranularityMismatch = errors.New(
		"historical comparison granularities do not match",
	)
	ErrAsOfTimeMismatch = errors.New(
		"historical comparison as-of times do not match",
	)
	ErrWindowDurationMismatch = errors.New(
		"historical comparison windows must have equal duration",
	)
	ErrWindowNotAdjacent = errors.New(
		"historical comparison previous window must end when the current window begins",
	)
	ErrSeriesUnavailable = errors.New(
		"historical comparison requires available current and previous series",
	)
	ErrAggregationUnsupported = errors.New(
		"historical comparison aggregation is unsupported",
	)
)

type ResultValidationError struct {
	Kind   error
	Report historicalcontract.ValidationReport
}

func (err *ResultValidationError) Error() string {
	if err == nil {
		return "historical comparison result validation failed"
	}

	return fmt.Sprintf(
		"%v: errors=%d warnings=%d",
		err.Kind,
		err.Report.ErrorCount,
		err.Report.WarningCount,
	)
}

func (err *ResultValidationError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Kind
}
