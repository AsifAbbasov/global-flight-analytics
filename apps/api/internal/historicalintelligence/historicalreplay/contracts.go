package historicalreplay

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

const Version = "historical-replay-v1"

const (
	DefaultMaximumWindowCount = 1_000
	MaximumWindowCount        = 10_000
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

type WindowResult struct {
	Bucket historicalwindow.Bucket
	Record historicalaggregate.Record
}

func (result WindowResult) Clone() WindowResult {
	cloned := result
	cloned.Record = result.Record.Clone()
	return cloned
}

type Result struct {
	Version string
	Plan    historicalwindow.Plan
	Windows []WindowResult
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
