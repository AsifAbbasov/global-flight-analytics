package metricexecution

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/confidencereport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/executor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const (
	MetricIDActiveAircraft  = "traffic.active_aircraft"
	MetricIDAirportActivity = "traffic.airport_activity"
)

const (
	NoticeCodeDuplicateTrajectoriesRemoved   = "duplicate_trajectories_removed"
	NoticeCodeIneligibleTrajectoriesExcluded = "ineligible_trajectories_excluded"
	NoticeCodeNoTrajectoryObservations       = "no_trajectory_observations"
	NoticeCodeFutureObservationTime          = "future_observation_time"
)

type ReasonCount struct {
	Reason trajectoryeligibility.ReasonCode
	Count  int
}

type ScopeSummary struct {
	Capability   trajectoryeligibility.Capability
	InputCount   int
	AllowedCount int
	DeniedCount  int
	Reasons      []ReasonCount
	EvaluatedAt  time.Time
}

type Execution[T any] struct {
	MetricID         string
	Result           analyticalresult.Result[T]
	Scope            ScopeSummary
	ConfidenceReport *confidencereport.Report
}

func (
	execution Execution[T],
) IsUsable() bool {
	return execution.Result.IsUsable()
}

func (
	execution Execution[T],
) IsDenied() bool {
	return execution.Result.Status ==
		analyticalresult.StatusDenied
}

func (
	execution Execution[T],
) IsFailed() bool {
	return execution.Result.Status ==
		analyticalresult.StatusFailed
}

type PublicationMetadata struct {
	DataQuality   *dataqualitycontract.Report
	Sources       []analyticalresult.Source
	Warnings      []analyticalresult.Notice
	Limitations   []analyticalresult.Notice
	FailureMapper executor.FailureMapper
}

type ActiveAircraftRequest struct {
	Trajectories []trajectory.FlightTrajectory
	PublicationMetadata
}

type TrafficDensityRequest struct {
	Trajectories         []trajectory.FlightTrajectory
	AreaSquareKilometers float64
	PublicationMetadata
}

type AirportActivityRequest struct {
	Arrivals   []trajectory.FlightTrajectory
	Departures []trajectory.FlightTrajectory
	PublicationMetadata
}

type CoverageScoreRequest struct {
	Snapshot snapshot.Snapshot
	PublicationMetadata
}

type DataFreshnessRequest struct {
	Snapshot snapshot.Snapshot
	MaxAge   time.Duration
	PublicationMetadata
}
