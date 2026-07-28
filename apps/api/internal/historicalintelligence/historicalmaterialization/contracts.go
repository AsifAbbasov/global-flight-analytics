package historicalmaterialization

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

const Version = "historical-materialization-v2"

type Config struct {
	// Repository retains the original source-compatible surface. The concrete
	// implementation must also implement historicalread.PeriodRepository so the
	// materializer can read adjacent periods atomically.
	Repository historicalread.Repository
	Store      historicalaggregate.Writer
	Now        func() time.Time
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
	GeneratedAt        time.Time
}

type ReadSummary struct {
	Window         historicalcontract.TimeWindow
	IsolationLevel string
	DatasetLimit   int

	FlightCount      int
	TrajectoryCount  int
	ObservationCount int
	RouteCount       int

	FlightMatchedCount      int64
	TrajectoryMatchedCount  int64
	ObservationMatchedCount int64
	RouteMatchedCount       int64

	RoutePayloadBytes      int64
	RouteTotalPayloadBytes int64

	FlightLimitReached      bool
	TrajectoryLimitReached  bool
	ObservationLimitReached bool
	RouteLimitReached       bool
	RouteByteLimitReached   bool
}

type PeriodReadSummaries struct {
	Previous ReadSummary
	Current  ReadSummary
}

type Outcome struct {
	Version string

	Plan          historicalwindow.Plan
	PreviousPlan  historicalwindow.Plan
	ReadSummaries PeriodReadSummaries

	// Deprecated: ReadSummary is the aggregate of both period summaries and is
	// retained only for source compatibility. Period-sensitive callers must use
	// ReadSummaries.Previous and ReadSummaries.Current.
	ReadSummary ReadSummary

	CurrentPeriodResult historicalcontract.Result
	CurrentResult       historicalcontract.Result
	PreviousResult      historicalcontract.Result
	Record              historicalaggregate.Record
}

func (outcome Outcome) Clone() Outcome {
	cloned := outcome
	cloned.Plan = outcome.Plan.Clone()
	cloned.PreviousPlan = outcome.PreviousPlan.Clone()
	cloned.CurrentPeriodResult =
		outcome.CurrentPeriodResult.Clone()
	cloned.CurrentResult = outcome.CurrentResult.Clone()
	cloned.PreviousResult = outcome.PreviousResult.Clone()
	cloned.Record = outcome.Record.Clone()

	return cloned
}
