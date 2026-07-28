package historicalmaterialization

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

var (
	ErrContextRequired = errors.New(
		"historical materialization context is required",
	)
	ErrReadRepositoryRequired = errors.New(
		"historical materialization period read repository is required",
	)
	ErrPeriodReadRepositoryRequired = errors.New(
		"historical materialization repository must support atomic adjacent-period reads",
	)
	ErrAggregateStoreRequired = errors.New(
		"historical materialization aggregate writer is required",
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
	ErrSnapshotVersionMismatch = errors.New(
		"historical materialization snapshot version does not match the Historical Read contract",
	)
	ErrSnapshotQueryMismatch = errors.New(
		"historical materialization snapshot query does not match the requested period",
	)
	ErrSnapshotIsolationMismatch = errors.New(
		"historical materialization period snapshots do not share one supported isolation boundary",
	)
	ErrPersistedRecordMismatch = errors.New(
		"historical materialization persisted record does not match the compared result",
	)
)

type Stage string

const (
	StageRequestValidation   Stage = "request_validation"
	StageCurrentPlanning     Stage = "current_planning"
	StagePreviousPlanning    Stage = "previous_planning"
	StagePeriodRead          Stage = "period_read"
	StageSnapshotContract    Stage = "snapshot_contract"
	StagePreviousBuild       Stage = "previous_build"
	StageCurrentBuild        Stage = "current_build"
	StageComparison          Stage = "comparison"
	StagePersistence         Stage = "persistence"
	StagePersistenceContract Stage = "persistence_contract"
)

type StageError struct {
	Stage Stage
	Err   error
}

func (err *StageError) Error() string {
	if err == nil {
		return "historical materialization stage failed"
	}
	return fmt.Sprintf(
		"historical materialization stage %s failed: %v",
		err.Stage,
		err.Err,
	)
}

func (err *StageError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

func stageError(stage Stage, err error) error {
	if err == nil {
		return nil
	}
	return &StageError{
		Stage: stage,
		Err:   err,
	}
}

type ContractError struct {
	Kind   error
	Period string
	Field  string
}

func (err *ContractError) Error() string {
	if err == nil {
		return "historical materialization contract mismatch"
	}
	return fmt.Sprintf(
		"%v: period=%s field=%s",
		err.Kind,
		err.Period,
		err.Field,
	)
}

func (err *ContractError) Unwrap() error {
	if err == nil || err.Kind == nil {
		return nil
	}
	return err.Kind
}

func contractError(
	kind error,
	period string,
	field string,
) error {
	return &ContractError{
		Kind:   kind,
		Period: period,
		Field:  field,
	}
}

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
// contract report when the final comparison or persisted record is invalid.
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
