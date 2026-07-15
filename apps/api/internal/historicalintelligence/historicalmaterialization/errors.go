package historicalmaterialization

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

var (
	ErrReadRepositoryRequired = errors.New(
		"historical materialization read repository is required",
	)
	ErrAggregateStoreRequired = errors.New(
		"historical materialization aggregate store is required",
	)
	ErrNoEffectiveWindow = errors.New(
		"historical materialization requires at least one complete current bucket and one adjacent previous window",
	)
	ErrMetricUnsupported = errors.New(
		"historical materialization metric is unsupported",
	)
	ErrScopeUnsupported = errors.New(
		"historical materialization scope is unsupported for the selected metric",
	)
	ErrDatasetLimitInvalid = errors.New(
		"historical materialization dataset limit is invalid",
	)
	ErrGeneratedAtBeforeAsOfTime = errors.New(
		"historical materialization generated time must not precede the analytical as-of time",
	)
	ErrMaterializedResultInvalid = errors.New(
		"historical materialization produced an invalid result",
	)
)

type MetricScopeError struct {
	Metric historicalcontract.MetricName
	Scope  historicalcontract.Scope
}

func (err *MetricScopeError) Error() string {
	if err == nil {
		return "historical materialization metric and scope are incompatible"
	}

	return fmt.Sprintf(
		"%v: metric=%s scope=%s",
		ErrScopeUnsupported,
		err.Metric,
		err.Scope.Type,
	)
}

func (err *MetricScopeError) Unwrap() error {
	return ErrScopeUnsupported
}

// ResultValidationError preserves the complete Historical Intelligence
// contract report when the final comparison provenance is invalid.
type ResultValidationError struct {
	Report historicalcontract.ValidationReport
}

func (err *ResultValidationError) Error() string {
	if err == nil {
		return ErrMaterializedResultInvalid.Error()
	}

	return fmt.Sprintf(
		"%v: errors=%d warnings=%d",
		ErrMaterializedResultInvalid,
		err.Report.ErrorCount,
		err.Report.WarningCount,
	)
}

func (err *ResultValidationError) Unwrap() error {
	return ErrMaterializedResultInvalid
}
