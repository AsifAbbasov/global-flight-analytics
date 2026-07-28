package historicalmaterialization

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalairport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcomparison"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalroute"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicaltraffic"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

var airportICAOPattern = regexp.MustCompile(
	`^[A-Z0-9]{4}$`,
)

type Materializer struct {
	repository historicalread.PeriodRepository
	store      historicalaggregate.Writer
	now        func() time.Time
}

func New(
	config Config,
) (*Materializer, error) {
	if config.Repository == nil {
		return nil, ErrReadRepositoryRequired
	}
	repository, supported := config.Repository.(historicalread.PeriodRepository)
	if !supported {
		return nil,
			ErrPeriodReadRepositoryRequired
	}
	if config.Store == nil {
		return nil, ErrAggregateStoreRequired
	}
	if config.Now == nil {
		config.Now = time.Now
	}

	return &Materializer{
		repository: repository,
		store:      config.Store,
		now:        config.Now,
	}, nil
}

func (materializer *Materializer) Materialize(
	ctx context.Context,
	request Request,
) (Outcome, error) {
	if ctx == nil {
		return Outcome{},
			stageError(
				StageRequestValidation,
				ErrContextRequired,
			)
	}
	if err := ctx.Err(); err != nil {
		return Outcome{},
			stageError(
				StageRequestValidation,
				err,
			)
	}

	normalized, family, err :=
		materializer.normalizeRequest(request)
	if err != nil {
		return Outcome{},
			stageError(
				StageRequestValidation,
				err,
			)
	}

	plans, err := buildAdjacentPlans(
		ctx,
		normalized,
	)
	if err != nil {
		return Outcome{}, err
	}
	queries := buildPeriodQueries(
		plans,
		normalized.DatasetLimit,
	)

	snapshots, err :=
		materializer.repository.ReadPeriods(
			ctx,
			queries,
		)
	if err != nil {
		return Outcome{},
			stageError(StagePeriodRead, err)
	}
	if err := ctx.Err(); err != nil {
		return Outcome{},
			stageError(StagePeriodRead, err)
	}
	if err := validatePeriodSnapshots(
		snapshots,
		queries,
	); err != nil {
		return Outcome{},
			stageError(
				StageSnapshotContract,
				err,
			)
	}

	previousResult, err := buildResult(
		family,
		snapshots.Previous,
		plans.Previous,
		normalized,
	)
	if err != nil {
		return Outcome{},
			stageError(
				StagePreviousBuild,
				err,
			)
	}
	currentPeriodResult, err := buildResult(
		family,
		snapshots.Current,
		plans.Current,
		normalized,
	)
	if err != nil {
		return Outcome{},
			stageError(
				StageCurrentBuild,
				err,
			)
	}

	compared, err := historicalcomparison.Attach(
		currentPeriodResult,
		previousResult,
	)
	if err != nil {
		return Outcome{},
			stageError(
				StageComparison,
				err,
			)
	}

	record, err := materializer.store.Put(
		ctx,
		compared,
	)
	if err != nil {
		return Outcome{},
			stageError(
				StagePersistence,
				err,
			)
	}
	if err := validatePersistedRecord(
		record,
		compared,
	); err != nil {
		return Outcome{},
			stageError(
				StagePersistenceContract,
				err,
			)
	}

	return Outcome{
		Version:      Version,
		Plan:         plans.Current.Clone(),
		PreviousPlan: plans.Previous.Clone(),
		ReadSummaries: PeriodReadSummaries{
			Previous: summarizeRead(
				snapshots.Previous,
			),
			Current: summarizeRead(
				snapshots.Current,
			),
		},
		ReadSummary: summarizeCombinedRead(
			snapshots,
		),
		CurrentPeriodResult: currentPeriodResult.Clone(),
		CurrentResult:       record.Result.Clone(),
		PreviousResult:      previousResult.Clone(),
		Record:              record.Clone(),
	}.Clone(), nil
}

type adjacentPlans struct {
	Previous historicalwindow.Plan
	Current  historicalwindow.Plan
}

func buildAdjacentPlans(
	ctx context.Context,
	request Request,
) (adjacentPlans, error) {
	current, err := historicalwindow.Build(
		ctx,
		historicalwindow.Request{
			StartTime: request.StartTime,
			EndTime:   request.EndTime,
			AsOfTime:  request.AsOfTime,
			Granularity: request.
				Granularity,
			MaximumBucketCount: request.
				MaximumBucketCount,
		},
	)
	if err != nil {
		return adjacentPlans{},
			stageError(
				StageCurrentPlanning,
				err,
			)
	}
	if current.EffectiveWindow == nil ||
		current.PreviousWindow == nil ||
		!current.HasBuckets() {
		return adjacentPlans{},
			stageError(
				StageCurrentPlanning,
				ErrNoEffectiveWindow,
			)
	}

	previous, err := historicalwindow.Build(
		ctx,
		historicalwindow.Request{
			StartTime: current.
				PreviousWindow.StartTime,
			EndTime: current.
				PreviousWindow.EndTime,
			AsOfTime: request.AsOfTime,
			Granularity: request.
				Granularity,
			MaximumBucketCount: request.
				MaximumBucketCount,
		},
	)
	if err != nil {
		return adjacentPlans{},
			stageError(
				StagePreviousPlanning,
				err,
			)
	}
	if previous.EffectiveWindow == nil ||
		!previous.HasBuckets() {
		return adjacentPlans{},
			stageError(
				StagePreviousPlanning,
				ErrNoEffectiveWindow,
			)
	}

	return adjacentPlans{
		Previous: previous.Clone(),
		Current:  current.Clone(),
	}, nil
}

func buildPeriodQueries(
	plans adjacentPlans,
	datasetLimit int,
) historicalread.PeriodQueries {
	return historicalread.PeriodQueries{
		Previous: historicalread.Query{
			Window: *plans.Previous.
				EffectiveWindow,
			Limit: datasetLimit,
			RoutePayloadByteLimit: historicalread.
				DefaultRoutePayloadByteLimit,
		},
		Current: historicalread.Query{
			Window: *plans.Current.
				EffectiveWindow,
			Limit: datasetLimit,
			RoutePayloadByteLimit: historicalread.
				DefaultRoutePayloadByteLimit,
		},
	}
}

type metricFamily string

const (
	metricFamilyTraffic metricFamily = "traffic"
	metricFamilyAirport metricFamily = "airport"
	metricFamilyRoute   metricFamily = "route"
)

func (materializer *Materializer) normalizeRequest(
	request Request,
) (Request, metricFamily, error) {
	family, ok := classifyMetric(
		request.MetricName,
	)
	if !ok {
		return Request{},
			"",
			ErrMetricUnsupported
	}

	scope, err := normalizeScope(
		request.Scope,
	)
	if err != nil {
		return Request{}, "", err
	}
	if !metricScopeAllowed(
		request.MetricName,
		scope.Type,
	) {
		return Request{},
			"",
			&MetricScopeError{
				Metric: request.MetricName,
				Scope:  scope,
			}
	}

	datasetLimit := request.DatasetLimit
	if datasetLimit == 0 {
		datasetLimit =
			historicalread.DefaultDatasetLimit
	}
	if datasetLimit < 1 ||
		datasetLimit >
			historicalread.MaximumDatasetLimit {
		return Request{},
			"",
			ErrDatasetLimitInvalid
	}

	generatedAt := request.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = materializer.now()
	}
	generatedAt = generatedAt.UTC()
	asOfTime := request.AsOfTime.UTC()
	if !asOfTime.IsZero() &&
		generatedAt.Before(asOfTime) {
		return Request{},
			"",
			ErrGeneratedAtBeforeAsOfTime
	}

	return Request{
		StartTime: request.StartTime.UTC(),
		EndTime:   request.EndTime.UTC(),
		AsOfTime:  asOfTime,

		Granularity: request.Granularity,
		MetricName:  request.MetricName,
		Scope:       scope,

		DatasetLimit: datasetLimit,
		MaximumBucketCount: request.
			MaximumBucketCount,
		GeneratedAt: generatedAt,
	}, family, nil
}

func classifyMetric(
	name historicalcontract.MetricName,
) (metricFamily, bool) {
	specification, exists := historicalcontract.MetricSpecFor(name)
	if !exists {
		return "", false
	}
	return metricFamily(specification.Family), true
}

func metricScopeAllowed(
	name historicalcontract.MetricName,
	scopeType historicalcontract.ScopeType,
) bool {
	specification, exists := historicalcontract.MetricSpecFor(name)
	return exists && specification.AllowsScope(scopeType)
}

func normalizeScope(
	scope historicalcontract.Scope,
) (historicalcontract.Scope, error) {
	normalized := historicalcontract.Scope{
		Type: scope.Type,
		RegionCode: strings.ToLower(
			strings.TrimSpace(scope.RegionCode),
		),
		AirportICAOCode: strings.ToUpper(
			strings.TrimSpace(
				scope.AirportICAOCode,
			),
		),
		OriginICAOCode: strings.ToUpper(
			strings.TrimSpace(
				scope.OriginICAOCode,
			),
		),
		DestinationICAOCode: strings.ToUpper(
			strings.TrimSpace(
				scope.DestinationICAOCode,
			),
		),
	}

	switch normalized.Type {
	case historicalcontract.ScopeTypeGlobal:
		if normalized.RegionCode != "" ||
			normalized.AirportICAOCode != "" ||
			normalized.OriginICAOCode != "" ||
			normalized.DestinationICAOCode != "" {
			return historicalcontract.Scope{},
				ErrScopeUnsupported
		}

	case historicalcontract.ScopeTypeAirport:
		if !airportICAOPattern.MatchString(
			normalized.AirportICAOCode,
		) ||
			normalized.RegionCode != "" ||
			normalized.OriginICAOCode != "" ||
			normalized.DestinationICAOCode != "" {
			return historicalcontract.Scope{},
				ErrScopeUnsupported
		}

	case historicalcontract.ScopeTypeRoute:
		if !airportICAOPattern.MatchString(
			normalized.OriginICAOCode,
		) ||
			!airportICAOPattern.MatchString(
				normalized.DestinationICAOCode,
			) ||
			normalized.RegionCode != "" ||
			normalized.AirportICAOCode != "" {
			return historicalcontract.Scope{},
				ErrScopeUnsupported
		}

	default:
		return historicalcontract.Scope{},
			ErrScopeUnsupported
	}

	return normalized, nil
}

func buildResult(
	family metricFamily,
	snapshot historicalread.Snapshot,
	plan historicalwindow.Plan,
	request Request,
) (historicalcontract.Result, error) {
	switch family {
	case metricFamilyTraffic:
		return historicaltraffic.Build(
			historicaltraffic.Request{
				Snapshot:    snapshot,
				Plan:        plan,
				MetricName:  request.MetricName,
				GeneratedAt: request.GeneratedAt,
			},
		)

	case metricFamilyAirport:
		return historicalairport.Build(
			historicalairport.Request{
				Snapshot: snapshot,
				Plan:     plan,
				AirportICAOCode: request.
					Scope.AirportICAOCode,
				MetricName:  request.MetricName,
				GeneratedAt: request.GeneratedAt,
			},
		)

	case metricFamilyRoute:
		return historicalroute.Build(
			historicalroute.Request{
				Snapshot: snapshot,
				Plan:     plan,
				OriginICAOCode: request.
					Scope.OriginICAOCode,
				DestinationICAOCode: request.
					Scope.DestinationICAOCode,
				MetricName:  request.MetricName,
				GeneratedAt: request.GeneratedAt,
			},
		)

	default:
		return historicalcontract.Result{},
			ErrMetricUnsupported
	}
}

func summarizeRead(
	snapshot historicalread.Snapshot,
) ReadSummary {
	return ReadSummary{
		Window: snapshot.Query.Window,
		IsolationLevel: snapshot.
			IsolationLevel,
		DatasetLimit: snapshot.Query.Limit,

		FlightCount:      len(snapshot.Flights),
		TrajectoryCount:  len(snapshot.Trajectories),
		ObservationCount: len(snapshot.Observations),
		RouteCount:       len(snapshot.Routes),

		FlightMatchedCount: snapshot.
			FlightMatchedCount,
		TrajectoryMatchedCount: snapshot.
			TrajectoryMatchedCount,
		ObservationMatchedCount: snapshot.
			ObservationMatchedCount,
		RouteMatchedCount: snapshot.
			RouteMatchedCount,

		RoutePayloadBytes: snapshot.
			RoutePayloadBytes,
		RouteTotalPayloadBytes: snapshot.
			RouteTotalPayloadBytes,

		FlightLimitReached: snapshot.
			FlightLimitReached,
		TrajectoryLimitReached: snapshot.
			TrajectoryLimitReached,
		ObservationLimitReached: snapshot.
			ObservationLimitReached,
		RouteLimitReached: snapshot.
			RouteLimitReached,
		RouteByteLimitReached: snapshot.
			RouteByteLimitReached,
	}
}

func summarizeCombinedRead(
	snapshots historicalread.PeriodSnapshots,
) ReadSummary {
	previous := summarizeRead(
		snapshots.Previous,
	)
	current := summarizeRead(
		snapshots.Current,
	)
	return ReadSummary{
		Window: historicalcontract.TimeWindow{
			StartTime: previous.Window.StartTime,
			EndTime:   current.Window.EndTime,
			AsOfTime:  current.Window.AsOfTime,
		},
		IsolationLevel: current.IsolationLevel,
		DatasetLimit:   current.DatasetLimit,

		FlightCount: previous.FlightCount +
			current.FlightCount,
		TrajectoryCount: previous.TrajectoryCount +
			current.TrajectoryCount,
		ObservationCount: previous.ObservationCount +
			current.ObservationCount,
		RouteCount: previous.RouteCount +
			current.RouteCount,

		FlightMatchedCount: previous.FlightMatchedCount +
			current.FlightMatchedCount,
		TrajectoryMatchedCount: previous.TrajectoryMatchedCount +
			current.TrajectoryMatchedCount,
		ObservationMatchedCount: previous.ObservationMatchedCount +
			current.ObservationMatchedCount,
		RouteMatchedCount: previous.RouteMatchedCount +
			current.RouteMatchedCount,

		RoutePayloadBytes: previous.RoutePayloadBytes +
			current.RoutePayloadBytes,
		RouteTotalPayloadBytes: previous.RouteTotalPayloadBytes +
			current.RouteTotalPayloadBytes,

		FlightLimitReached: previous.FlightLimitReached ||
			current.FlightLimitReached,
		TrajectoryLimitReached: previous.TrajectoryLimitReached ||
			current.TrajectoryLimitReached,
		ObservationLimitReached: previous.ObservationLimitReached ||
			current.ObservationLimitReached,
		RouteLimitReached: previous.RouteLimitReached ||
			current.RouteLimitReached,
		RouteByteLimitReached: previous.RouteByteLimitReached ||
			current.RouteByteLimitReached,
	}
}

func validatePeriodSnapshots(
	snapshots historicalread.PeriodSnapshots,
	queries historicalread.PeriodQueries,
) error {
	if err := validateSnapshot(
		"previous",
		snapshots.Previous,
		queries.Previous,
	); err != nil {
		return err
	}
	if err := validateSnapshot(
		"current",
		snapshots.Current,
		queries.Current,
	); err != nil {
		return err
	}
	if snapshots.Previous.IsolationLevel !=
		snapshots.Current.IsolationLevel {
		return contractError(
			ErrSnapshotIsolationMismatch,
			"both",
			"isolation_level",
		)
	}
	return nil
}

func validateSnapshot(
	period string,
	snapshot historicalread.Snapshot,
	expected historicalread.Query,
) error {
	if snapshot.Version != historicalread.Version {
		return contractError(
			ErrSnapshotVersionMismatch,
			period,
			"version",
		)
	}
	if !snapshot.Query.Equal(expected) {
		return contractError(
			ErrSnapshotQueryMismatch,
			period,
			"query",
		)
	}
	switch snapshot.IsolationLevel {
	case historicalread.
		SnapshotIsolationRepeatableRead,
		historicalread.
			SnapshotIsolationCallerTransaction:
		return nil
	default:
		return contractError(
			ErrSnapshotIsolationMismatch,
			period,
			"isolation_level",
		)
	}
}

func validatePersistedRecord(
	record historicalaggregate.Record,
	expected historicalcontract.Result,
) error {
	if strings.TrimSpace(record.ID) == "" {
		return contractError(
			ErrPersistedRecordMismatch,
			"current",
			"id",
		)
	}
	if record.InputFingerprint !=
		expected.Provenance.InputFingerprint {
		return contractError(
			ErrPersistedRecordMismatch,
			"current",
			"input_fingerprint",
		)
	}
	if record.Result.Provenance.
		InputFingerprint !=
		expected.Provenance.InputFingerprint {
		return contractError(
			ErrPersistedRecordMismatch,
			"current",
			"result_input_fingerprint",
		)
	}
	if !resultKeyMatches(
		record.Key,
		expected,
	) {
		return contractError(
			ErrPersistedRecordMismatch,
			"current",
			"key",
		)
	}
	if !resultIdentityMatches(
		record.Result,
		expected,
	) {
		return contractError(
			ErrPersistedRecordMismatch,
			"current",
			"result_identity",
		)
	}
	if record.StoredAt.IsZero() ||
		record.StoredAt.UTC().Before(
			record.Result.GeneratedAt.UTC(),
		) {
		return contractError(
			ErrPersistedRecordMismatch,
			"current",
			"stored_at",
		)
	}

	report := historicalcontract.Validate(
		record.Result,
	)
	if report.Status !=
		historicalcontract.ValidationStatusValid {
		return &ResultValidationError{
			Report: report.Clone(),
		}
	}
	return nil
}

func resultKeyMatches(
	key historicalaggregate.ResultKey,
	result historicalcontract.Result,
) bool {
	return key.SchemaVersion ==
		result.SchemaVersion &&
		key.MetricName == result.Metric.Name &&
		key.Scope.Equal(result.Scope) &&
		key.Granularity ==
			result.Granularity &&
		windowsEqual(
			key.Window,
			result.Window,
		)
}

func resultIdentityMatches(
	actual historicalcontract.Result,
	expected historicalcontract.Result,
) bool {
	return actual.SchemaVersion ==
		expected.SchemaVersion &&
		actual.Status == expected.Status &&
		actual.Metric == expected.Metric &&
		actual.Scope.Equal(expected.Scope) &&
		actual.Granularity ==
			expected.Granularity &&
		windowsEqual(
			actual.Window,
			expected.Window,
		) &&
		actual.Summary == expected.Summary &&
		comparisonsEqual(
			actual.Comparison,
			expected.Comparison,
		) &&
		actual.GeneratedAt.UTC().Equal(
			expected.GeneratedAt.UTC(),
		)
}

func comparisonsEqual(
	left *historicalcontract.PeriodComparison,
	right *historicalcontract.PeriodComparison,
) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	if !windowsEqual(
		left.PreviousWindow,
		right.PreviousWindow,
	) ||
		left.CurrentValue !=
			right.CurrentValue ||
		left.PreviousValue !=
			right.PreviousValue ||
		left.AbsoluteChange !=
			right.AbsoluteChange ||
		left.Direction != right.Direction {
		return false
	}
	if left.PercentageChange == nil ||
		right.PercentageChange == nil {
		return left.PercentageChange == nil &&
			right.PercentageChange == nil
	}
	return *left.PercentageChange ==
		*right.PercentageChange
}

func windowsEqual(
	left historicalcontract.TimeWindow,
	right historicalcontract.TimeWindow,
) bool {
	return left.StartTime.UTC().Equal(
		right.StartTime.UTC(),
	) &&
		left.EndTime.UTC().Equal(
			right.EndTime.UTC(),
		) &&
		left.AsOfTime.UTC().Equal(
			right.AsOfTime.UTC(),
		)
}
