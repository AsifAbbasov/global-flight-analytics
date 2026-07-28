package main

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalreplay"
)

const commandVersion = "historical-intelligence-production-runner-v2"

type operationMode string

const (
	operationModeMaterialize operationMode = "materialize"
	operationModeReplay      operationMode = "replay"
)

type commandOptions struct {
	Mode operationMode

	StartTime time.Time
	EndTime   time.Time
	AsOfTime  time.Time

	Granularity historicalcontract.Granularity
	MetricName  historicalcontract.MetricName
	Scope       historicalcontract.Scope

	DatasetLimit       int
	MaximumBucketCount int
	MaximumWindowCount int
}

type reportWindow struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	AsOfTime  time.Time `json:"as_of_time"`
}

type reportScope struct {
	Type                string `json:"type"`
	AirportICAOCode     string `json:"airport_icao_code,omitempty"`
	OriginICAOCode      string `json:"origin_icao_code,omitempty"`
	DestinationICAOCode string `json:"destination_icao_code,omitempty"`
}

type reportRecord struct {
	ID               string       `json:"id"`
	InputFingerprint string       `json:"input_fingerprint"`
	Window           reportWindow `json:"window"`
	Status           string       `json:"status"`
	ConfidenceLevel  string       `json:"confidence_level"`
	PointCount       int          `json:"point_count"`
	Total            float64      `json:"total"`
	StoredAt         time.Time    `json:"stored_at"`
}

type reportReadSummary struct {
	Window reportWindow `json:"window"`

	FlightCount      int `json:"flight_count"`
	TrajectoryCount  int `json:"trajectory_count"`
	ObservationCount int `json:"observation_count"`
	RouteCount       int `json:"route_count"`

	FlightLimitReached      bool `json:"flight_limit_reached"`
	TrajectoryLimitReached  bool `json:"trajectory_limit_reached"`
	ObservationLimitReached bool `json:"observation_limit_reached"`
	RouteLimitReached       bool `json:"route_limit_reached"`
}

type reportReplayFailure struct {
	Sequence int `json:"sequence"`

	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`

	Code    string `json:"code"`
	Message string `json:"message"`
}

type commandReport struct {
	Version string `json:"version"`
	Mode    string `json:"mode"`
	Status  string `json:"status"`

	MetricName  string      `json:"metric_name"`
	Scope       reportScope `json:"scope"`
	Granularity string      `json:"granularity"`

	RequestedWindow reportWindow `json:"requested_window"`

	DatasetLimit       int `json:"dataset_limit"`
	MaximumBucketCount int `json:"maximum_bucket_count"`
	MaximumWindowCount int `json:"maximum_window_count,omitempty"`

	MaterializedRecordCount    int            `json:"materialized_record_count"`
	ReplayWindowCount          int            `json:"replay_window_count,omitempty"`
	PlannedReplayWindowCount   int            `json:"planned_replay_window_count,omitempty"`
	CompletedReplayWindowCount int            `json:"completed_replay_window_count,omitempty"`
	Records                    []reportRecord `json:"records"`

	ReplayInputFingerprint string               `json:"replay_input_fingerprint,omitempty"`
	ReplayFailure          *reportReplayFailure `json:"replay_failure,omitempty"`
	ReadSummary            *reportReadSummary   `json:"read_summary,omitempty"`

	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

func reportFromMaterialization(
	options commandOptions,
	outcome historicalmaterialization.Outcome,
	completedAt time.Time,
) commandReport {
	readSummary := reportReadSummary{
		Window: reportWindowFromContract(
			outcome.ReadSummary.Window,
		),
		FlightCount: outcome.ReadSummary.
			FlightCount,
		TrajectoryCount: outcome.ReadSummary.
			TrajectoryCount,
		ObservationCount: outcome.ReadSummary.
			ObservationCount,
		RouteCount: outcome.ReadSummary.
			RouteCount,
		FlightLimitReached: outcome.ReadSummary.
			FlightLimitReached,
		TrajectoryLimitReached: outcome.ReadSummary.
			TrajectoryLimitReached,
		ObservationLimitReached: outcome.ReadSummary.
			ObservationLimitReached,
		RouteLimitReached: outcome.ReadSummary.
			RouteLimitReached,
	}

	return commandReport{
		Version: commandVersion,
		Mode:    string(options.Mode),
		Status:  string(historicalreplay.StatusComplete),

		MetricName: string(options.MetricName),
		Scope: reportScopeFromContract(
			options.Scope,
		),
		Granularity: string(
			options.Granularity,
		),
		RequestedWindow: reportWindow{
			StartTime: options.StartTime.UTC(),
			EndTime:   options.EndTime.UTC(),
			AsOfTime:  options.AsOfTime.UTC(),
		},
		DatasetLimit: options.DatasetLimit,
		MaximumBucketCount: options.
			MaximumBucketCount,
		MaterializedRecordCount: 1,
		Records: []reportRecord{
			reportRecordFromAggregate(
				outcome.Record,
			),
		},
		ReadSummary: &readSummary,
		StartedAt:   outcome.Record.Result.GeneratedAt.UTC(),
		CompletedAt: completedAt.UTC(),
	}
}

func reportFromReplay(
	options commandOptions,
	result historicalreplay.Result,
	fallbackCompletedAt time.Time,
) commandReport {
	records := make(
		[]reportRecord,
		0,
		len(result.Windows),
	)
	for _, window := range result.Windows {
		records = append(
			records,
			reportRecordFromAggregate(
				window.Record,
			),
		)
	}

	status := result.Status
	if status == "" {
		if result.HasFailure && len(result.Windows) > 0 {
			status = historicalreplay.StatusPartial
		} else if result.HasFailure {
			status = historicalreplay.StatusFailed
		} else {
			status = historicalreplay.StatusComplete
		}
	}
	plannedWindowCount := result.PlannedWindowCount
	if plannedWindowCount == 0 {
		plannedWindowCount = len(result.Plan.Buckets)
	}
	if plannedWindowCount == 0 {
		plannedWindowCount = len(result.Windows)
	}
	completedWindowCount := result.CompletedWindowCount
	if completedWindowCount == 0 && len(result.Windows) > 0 {
		completedWindowCount = len(result.Windows)
	}
	startedAt := result.StartedAt.UTC()
	if startedAt.IsZero() {
		startedAt = result.GeneratedAt.UTC()
	}
	if startedAt.IsZero() {
		startedAt = fallbackCompletedAt.UTC()
	}
	completedAt := result.CompletedAt.UTC()
	if completedAt.IsZero() {
		completedAt = fallbackCompletedAt.UTC()
	}

	var failure *reportReplayFailure
	if result.HasFailure {
		failure = &reportReplayFailure{
			Sequence:  result.Failure.Sequence,
			StartTime: result.Failure.StartTime.UTC(),
			EndTime:   result.Failure.EndTime.UTC(),
			Code:      string(result.Failure.Code),
			Message:   result.Failure.Message,
		}
	}

	return commandReport{
		Version: commandVersion,
		Mode:    string(options.Mode),
		Status:  string(status),

		MetricName: string(options.MetricName),
		Scope: reportScopeFromContract(
			options.Scope,
		),
		Granularity: string(
			options.Granularity,
		),
		RequestedWindow: reportWindow{
			StartTime: options.StartTime.UTC(),
			EndTime:   options.EndTime.UTC(),
			AsOfTime:  options.AsOfTime.UTC(),
		},
		DatasetLimit: options.DatasetLimit,
		MaximumBucketCount: options.
			MaximumBucketCount,
		MaximumWindowCount: options.
			MaximumWindowCount,
		MaterializedRecordCount:    len(records),
		ReplayWindowCount:          completedWindowCount,
		PlannedReplayWindowCount:   plannedWindowCount,
		CompletedReplayWindowCount: completedWindowCount,
		Records:                    records,
		ReplayInputFingerprint:     result.InputFingerprint,
		ReplayFailure:              failure,
		StartedAt:                  startedAt,
		CompletedAt:                completedAt,
	}
}

func reportWindowFromContract(
	window historicalcontract.TimeWindow,
) reportWindow {
	return reportWindow{
		StartTime: window.StartTime.UTC(),
		EndTime:   window.EndTime.UTC(),
		AsOfTime:  window.AsOfTime.UTC(),
	}
}

func reportScopeFromContract(
	scope historicalcontract.Scope,
) reportScope {
	return reportScope{
		Type:            string(scope.Type),
		AirportICAOCode: scope.AirportICAOCode,
		OriginICAOCode:  scope.OriginICAOCode,
		DestinationICAOCode: scope.
			DestinationICAOCode,
	}
}
