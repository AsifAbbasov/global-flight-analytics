package historicalmaterialization

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
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
	repository historicalread.Repository
	store      historicalaggregate.Store
	now        func() time.Time
}

func New(
	config Config,
) (*Materializer, error) {
	if config.Repository == nil {
		return nil, ErrReadRepositoryRequired
	}
	if config.Store == nil {
		return nil, ErrAggregateStoreRequired
	}
	if config.Now == nil {
		config.Now = time.Now
	}

	return &Materializer{
		repository: config.Repository,
		store:      config.Store,
		now:        config.Now,
	}, nil
}

func (materializer *Materializer) Materialize(
	ctx context.Context,
	request Request,
) (Outcome, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Outcome{}, err
	}

	normalized, family, err :=
		materializer.normalizeRequest(request)
	if err != nil {
		return Outcome{}, err
	}

	currentPlan, err := historicalwindow.Build(
		ctx,
		historicalwindow.Request{
			StartTime: normalized.StartTime,
			EndTime:   normalized.EndTime,
			AsOfTime:  normalized.AsOfTime,
			Granularity: normalized.
				Granularity,
			MaximumBucketCount: normalized.
				MaximumBucketCount,
		},
	)
	if err != nil {
		return Outcome{}, err
	}
	if currentPlan.EffectiveWindow == nil ||
		currentPlan.PreviousWindow == nil ||
		!currentPlan.HasBuckets() {
		return Outcome{}, ErrNoEffectiveWindow
	}

	previousPlan, err := historicalwindow.Build(
		ctx,
		historicalwindow.Request{
			StartTime: currentPlan.
				PreviousWindow.StartTime,
			EndTime: currentPlan.
				PreviousWindow.EndTime,
			AsOfTime: normalized.AsOfTime,
			Granularity: normalized.
				Granularity,
			MaximumBucketCount: normalized.
				MaximumBucketCount,
		},
	)
	if err != nil {
		return Outcome{}, err
	}
	if previousPlan.EffectiveWindow == nil ||
		!previousPlan.HasBuckets() {
		return Outcome{}, ErrNoEffectiveWindow
	}

	readWindow := historicalcontract.TimeWindow{
		StartTime: previousPlan.
			EffectiveWindow.StartTime,
		EndTime: currentPlan.
			EffectiveWindow.EndTime,
		AsOfTime: normalized.AsOfTime,
	}
	snapshot, err := materializer.repository.Read(
		ctx,
		historicalread.Query{
			Window: readWindow,
			Limit:  normalized.DatasetLimit,
		},
	)
	if err != nil {
		return Outcome{}, err
	}
	if err := ctx.Err(); err != nil {
		return Outcome{}, err
	}

	previousResult, err := buildResult(
		family,
		snapshot,
		previousPlan,
		normalized,
	)
	if err != nil {
		return Outcome{}, err
	}
	currentResult, err := buildResult(
		family,
		snapshot,
		currentPlan,
		normalized,
	)
	if err != nil {
		return Outcome{}, err
	}

	compared, err := historicalcomparison.Attach(
		currentResult,
		previousResult,
	)
	if err != nil {
		return Outcome{}, err
	}
	compared, err = finalizeComparedResult(
		compared,
		currentResult,
		previousResult,
	)
	if err != nil {
		return Outcome{}, err
	}

	record, err := materializer.store.Put(
		ctx,
		compared,
	)
	if err != nil {
		return Outcome{}, err
	}

	return Outcome{
		Version:      Version,
		Plan:         currentPlan.Clone(),
		PreviousPlan: previousPlan.Clone(),
		ReadSummary: summarizeRead(
			readWindow,
			snapshot,
		),
		CurrentResult:  compared.Clone(),
		PreviousResult: previousResult.Clone(),
		Record:         record.Clone(),
	}.Clone(), nil
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
	if !scopeAllowed(
		family,
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

		DatasetLimit: request.DatasetLimitOr(
			datasetLimit,
		),
		MaximumBucketCount: request.
			MaximumBucketCount,
		GeneratedAt: generatedAt,
	}, family, nil
}

func (request Request) DatasetLimitOr(
	fallback int,
) int {
	if request.DatasetLimit == 0 {
		return fallback
	}
	return request.DatasetLimit
}

func classifyMetric(
	name historicalcontract.MetricName,
) (metricFamily, bool) {
	switch name {
	case historicalcontract.MetricNameActiveAircraft,
		historicalcontract.MetricNameFlightCount,
		historicalcontract.MetricNameTrajectoryCount,
		historicalcontract.MetricNameObservationCount,
		historicalcontract.MetricNameTrafficDensity:
		return metricFamilyTraffic, true

	case historicalcontract.MetricNameAirportDepartures,
		historicalcontract.MetricNameAirportArrivals,
		historicalcontract.MetricNameAirportOperations,
		historicalcontract.MetricNameUniqueAircraft:
		return metricFamilyAirport, true

	case historicalcontract.MetricNameActiveRoutes,
		historicalcontract.MetricNameRouteObservations,
		historicalcontract.MetricNameRouteConfidence,
		historicalcontract.MetricNameCompleteRouteRatio,
		historicalcontract.MetricNamePartialRouteRatio,
		historicalcontract.MetricNameUnavailableRouteRatio,
		historicalcontract.MetricNameGreatCircleDistanceKM:
		return metricFamilyRoute, true

	default:
		return "", false
	}
}

func scopeAllowed(
	family metricFamily,
	scopeType historicalcontract.ScopeType,
) bool {
	switch family {
	case metricFamilyTraffic:
		return scopeType ==
			historicalcontract.ScopeTypeGlobal
	case metricFamilyAirport:
		return scopeType ==
			historicalcontract.ScopeTypeAirport
	case metricFamilyRoute:
		return scopeType ==
			historicalcontract.ScopeTypeGlobal ||
			scopeType ==
				historicalcontract.ScopeTypeRoute
	default:
		return false
	}
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
	window historicalcontract.TimeWindow,
	snapshot historicalread.Snapshot,
) ReadSummary {
	return ReadSummary{
		Window: window,

		FlightCount:      len(snapshot.Flights),
		TrajectoryCount:  len(snapshot.Trajectories),
		ObservationCount: len(snapshot.Observations),
		RouteCount:       len(snapshot.Routes),

		FlightLimitReached: snapshot.
			FlightLimitReached,
		TrajectoryLimitReached: snapshot.
			TrajectoryLimitReached,
		ObservationLimitReached: snapshot.
			ObservationLimitReached,
		RouteLimitReached: snapshot.
			RouteLimitReached,
	}
}

func finalizeComparedResult(
	compared historicalcontract.Result,
	current historicalcontract.Result,
	previous historicalcontract.Result,
) (historicalcontract.Result, error) {
	result := compared.Clone()
	result.Provenance.BuilderVersion = strings.Join(
		[]string{
			Version,
			historicalcomparison.Version,
			strings.TrimSpace(
				current.Provenance.BuilderVersion,
			),
			strings.TrimSpace(
				previous.Provenance.BuilderVersion,
			),
		},
		"+",
	)
	result.Provenance.InputFingerprint =
		materializationFingerprint(
			current,
			previous,
		)
	result.Provenance.SourceNames = mergeSourceNames(
		current.Provenance.SourceNames,
		previous.Provenance.SourceNames,
	)
	result.Provenance.LatestSourceUpdatedAt = laterTime(
		current.Provenance.LatestSourceUpdatedAt,
		previous.Provenance.LatestSourceUpdatedAt,
	)

	report := historicalcontract.Validate(result)
	if report.Status !=
		historicalcontract.ValidationStatusValid {
		return historicalcontract.Result{},
			&ResultValidationError{
				Report: report.Clone(),
			}
	}

	return result.Clone(), nil
}

func materializationFingerprint(
	current historicalcontract.Result,
	previous historicalcontract.Result,
) string {
	records := []string{
		Version,
		historicalcomparison.Version,
		string(current.SchemaVersion),
		string(current.Metric.Name),
		string(current.Granularity),
		scopeFingerprint(current.Scope),
		current.Window.StartTime.UTC().
			Format(time.RFC3339Nano),
		current.Window.EndTime.UTC().
			Format(time.RFC3339Nano),
		current.Window.AsOfTime.UTC().
			Format(time.RFC3339Nano),
		previous.Window.StartTime.UTC().
			Format(time.RFC3339Nano),
		previous.Window.EndTime.UTC().
			Format(time.RFC3339Nano),
		previous.Window.AsOfTime.UTC().
			Format(time.RFC3339Nano),
		strings.TrimSpace(
			current.Provenance.BuilderVersion,
		),
		strings.TrimSpace(
			current.Provenance.InputFingerprint,
		),
		strings.TrimSpace(
			previous.Provenance.BuilderVersion,
		),
		strings.TrimSpace(
			previous.Provenance.InputFingerprint,
		),
	}

	sum := sha256.Sum256(
		[]byte(strings.Join(records, "\n")),
	)
	return "sha256:" +
		hex.EncodeToString(sum[:])
}

func scopeFingerprint(
	scope historicalcontract.Scope,
) string {
	return fmt.Sprintf(
		"%s|%s|%s|%s|%s",
		scope.Type,
		scope.RegionCode,
		scope.AirportICAOCode,
		scope.OriginICAOCode,
		scope.DestinationICAOCode,
	)
}

func mergeSourceNames(
	groups ...[]string,
) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)

	for _, group := range groups {
		for _, sourceName := range group {
			normalized := strings.TrimSpace(
				sourceName,
			)
			if normalized == "" {
				continue
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			result = append(result, normalized)
		}
	}

	sort.Strings(result)
	return result
}

func laterTime(
	left time.Time,
	right time.Time,
) time.Time {
	left = left.UTC()
	right = right.UTC()
	if right.After(left) {
		return right
	}
	return left
}
