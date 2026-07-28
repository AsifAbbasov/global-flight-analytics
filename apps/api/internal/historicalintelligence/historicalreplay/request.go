package historicalreplay

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

var replayAirportICAOPattern = regexp.MustCompile(
	`^[A-Z0-9]{4}$`,
)

type normalizedRequest struct {
	Request

	PlanningMaximumBucketCount int
}

func (runner *Runner) normalizeRequest(
	request Request,
	startedAt time.Time,
) (normalizedRequest, error) {
	specification, exists := historicalcontract.MetricSpecFor(
		request.MetricName,
	)
	if !exists {
		return normalizedRequest{},
			ErrMetricUnsupported
	}

	scope, err := normalizeReplayScope(
		request.Scope,
	)
	if err != nil {
		return normalizedRequest{}, err
	}
	if !specification.AllowsScope(scope.Type) {
		return normalizedRequest{},
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
		return normalizedRequest{},
			ErrDatasetLimitInvalid
	}

	maximumBucketCount :=
		request.MaximumBucketCount
	if maximumBucketCount == 0 {
		maximumBucketCount =
			historicalwindow.DefaultMaximumBucketCount
	}
	if maximumBucketCount < 1 ||
		maximumBucketCount >
			historicalwindow.MaximumBucketCount {
		return normalizedRequest{},
			ErrMaximumBucketCountInvalid
	}

	maximumWindowCount :=
		request.MaximumWindowCount
	if maximumWindowCount == 0 {
		maximumWindowCount =
			DefaultMaximumWindowCount
	}
	if maximumWindowCount < 1 ||
		maximumWindowCount >
			MaximumWindowCount {
		return normalizedRequest{},
			ErrMaximumWindowCountInvalid
	}

	generatedAt := request.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = startedAt
	}
	generatedAt = generatedAt.UTC()
	asOfTime := request.AsOfTime.UTC()
	if !asOfTime.IsZero() &&
		generatedAt.Before(asOfTime) {
		return normalizedRequest{},
			ErrGeneratedAtBeforeAsOfTime
	}
	if generatedAt.After(startedAt) {
		return normalizedRequest{},
			ErrGeneratedAtAfterStartTime
	}

	planningMaximum := maximumBucketCount
	if maximumWindowCount < planningMaximum {
		planningMaximum = maximumWindowCount
	}

	return normalizedRequest{
		Request: Request{
			StartTime: request.StartTime.UTC(),
			EndTime:   request.EndTime.UTC(),
			AsOfTime:  asOfTime,

			Granularity: request.Granularity,
			MetricName:  request.MetricName,
			Scope:       scope,

			DatasetLimit:       datasetLimit,
			MaximumBucketCount: maximumBucketCount,
			MaximumWindowCount: maximumWindowCount,
			GeneratedAt:        generatedAt,
		},
		PlanningMaximumBucketCount: planningMaximum,
	}, nil
}

func buildReplayPlan(
	ctx context.Context,
	request normalizedRequest,
) (historicalwindow.Plan, error) {
	plan, err := historicalwindow.Build(
		ctx,
		historicalwindow.Request{
			StartTime: request.StartTime,
			EndTime:   request.EndTime,
			AsOfTime:  request.AsOfTime,
			Granularity: request.
				Granularity,
			MaximumBucketCount: request.
				PlanningMaximumBucketCount,
		},
	)
	if err == nil {
		return plan.Clone(), nil
	}

	var countErr *historicalwindow.
		BucketCountExceededError
	if errors.As(err, &countErr) &&
		request.MaximumWindowCount <=
			request.MaximumBucketCount &&
		request.PlanningMaximumBucketCount ==
			request.MaximumWindowCount {
		return historicalwindow.Plan{},
			&WindowCountExceededError{
				Count:   countErr.Count,
				Maximum: request.MaximumWindowCount,
			}
	}
	return historicalwindow.Plan{}, err
}

func (request normalizedRequest) materializationRequest(
	bucket historicalwindow.Bucket,
) historicalmaterialization.Request {
	return historicalmaterialization.Request{
		StartTime: bucket.StartTime,
		EndTime:   bucket.EndTime,
		AsOfTime:  request.AsOfTime,

		Granularity: request.Granularity,
		MetricName:  request.MetricName,
		Scope:       request.Scope,

		DatasetLimit: request.DatasetLimit,

		// Every replay call represents exactly one current bucket. Using one here
		// prevents the replay-plan allocation guard from drifting into the
		// per-window materialization contract.
		MaximumBucketCount: 1,
		GeneratedAt:        request.GeneratedAt,
	}
}

func normalizeReplayScope(
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
		if !replayAirportICAOPattern.MatchString(
			normalized.AirportICAOCode,
		) ||
			normalized.RegionCode != "" ||
			normalized.OriginICAOCode != "" ||
			normalized.DestinationICAOCode != "" {
			return historicalcontract.Scope{},
				ErrScopeUnsupported
		}

	case historicalcontract.ScopeTypeRoute:
		if !replayAirportICAOPattern.MatchString(
			normalized.OriginICAOCode,
		) ||
			!replayAirportICAOPattern.MatchString(
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
