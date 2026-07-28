package historicalreplay

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

const Version = "historical-replay-v2"

const FingerprintVersion = "historical-replay-input-fingerprint-v1"

const (
	DefaultMaximumWindowCount = 1_000
	MaximumWindowCount        = 10_000
)

type Status string

const (
	StatusComplete Status = "complete"
	StatusPartial  Status = "partial"
	StatusFailed   Status = "failed"
)

type FailureCode string

const (
	FailureCodeNoReplayWindow     FailureCode = "no_replay_window"
	FailureCodeContextCanceled    FailureCode = "context_canceled"
	FailureCodeMaterialization    FailureCode = "materialization_failed"
	FailureCodeOutcomeContract    FailureCode = "outcome_contract_invalid"
	FailureCodeContinuityMismatch FailureCode = "continuity_mismatch"
)

type Materializer interface {
	Materialize(
		context.Context,
		historicalmaterialization.Request,
	) (historicalmaterialization.Outcome, error)
}

type Config struct {
	Materializer Materializer
	Now          func() time.Time
}

type Request struct {
	StartTime time.Time
	EndTime   time.Time
	AsOfTime  time.Time

	Granularity historicalcontract.Granularity
	MetricName  historicalcontract.MetricName
	Scope       historicalcontract.Scope

	DatasetLimit       int
	MaximumBucketCount int
	MaximumWindowCount int
	GeneratedAt        time.Time
}

type Failure struct {
	Sequence int

	StartTime time.Time
	EndTime   time.Time

	Code    FailureCode
	Message string
}

type WindowResult struct {
	Bucket historicalwindow.Bucket
	Record historicalaggregate.Record

	PreviousPeriodInputFingerprint string
	CurrentPeriodInputFingerprint  string
}

func (result WindowResult) Clone() WindowResult {
	cloned := result
	cloned.Record = result.Record.Clone()
	return cloned
}

type Result struct {
	Version string
	Status  Status

	Plan historicalwindow.Plan

	MetricName  historicalcontract.MetricName
	Scope       historicalcontract.Scope
	Granularity historicalcontract.Granularity

	DatasetLimit       int
	MaximumBucketCount int
	MaximumWindowCount int

	PlannedWindowCount   int
	CompletedWindowCount int
	Windows              []WindowResult

	HasFailure bool
	Failure    Failure

	GeneratedAt time.Time
	StartedAt   time.Time
	CompletedAt time.Time

	InputFingerprint string
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Plan = result.Plan.Clone()
	cloned.Windows = make(
		[]WindowResult,
		0,
		len(result.Windows),
	)
	for _, window := range result.Windows {
		cloned.Windows = append(
			cloned.Windows,
			window.Clone(),
		)
	}
	return cloned
}

func (result Result) Validate() error {
	return validateResult(result)
}
