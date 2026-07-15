package historicalmaterialization

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

const Version = "historical-materialization-v1"

type Config struct {
	Repository historicalread.Repository
	Store      historicalaggregate.Store
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
	Window historicalcontract.TimeWindow

	FlightCount      int
	TrajectoryCount  int
	ObservationCount int
	RouteCount       int

	FlightLimitReached      bool
	TrajectoryLimitReached  bool
	ObservationLimitReached bool
	RouteLimitReached       bool
}

type Outcome struct {
	Version string

	Plan         historicalwindow.Plan
	PreviousPlan historicalwindow.Plan
	ReadSummary  ReadSummary

	CurrentResult  historicalcontract.Result
	PreviousResult historicalcontract.Result
	Record         historicalaggregate.Record
}

func (outcome Outcome) Clone() Outcome {
	cloned := outcome
	cloned.Plan = outcome.Plan.Clone()
	cloned.PreviousPlan = outcome.PreviousPlan.Clone()
	cloned.CurrentResult = outcome.CurrentResult.Clone()
	cloned.PreviousResult = outcome.PreviousResult.Clone()
	cloned.Record = outcome.Record.Clone()

	return cloned
}
