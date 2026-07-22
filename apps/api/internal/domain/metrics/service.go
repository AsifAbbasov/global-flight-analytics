package metrics

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

var ErrInvalidWindowMinutes = errors.New("invalid active aircraft window minutes")

type RegionResolver interface {
	GetByCode(string) (region.Region, error)
}

type Service struct {
	repository     Repository
	regionResolver RegionResolver
	now            func() time.Time
}

func NewService(
	repository Repository,
	regionResolver RegionResolver,
) *Service {
	return newServiceWithClock(
		repository,
		regionResolver,
		time.Now,
	)
}

func newServiceWithClock(
	repository Repository,
	regionResolver RegionResolver,
	now func() time.Time,
) *Service {
	dependency.Must("metrics repository", repository)
	dependency.Must("metrics region resolver", regionResolver)
	dependency.Must("metrics clock", now)

	return &Service{
		repository:     repository,
		regionResolver: regionResolver,
		now:            now,
	}
}

func (
	s *Service,
) CalculateActiveAircraft(
	ctx context.Context,
	request ActiveAircraftRequest,
) (ActiveAircraftMetric, error) {
	windowMinutes, err := normalizeActiveAircraftWindowMinutes(
		request.WindowMinutes,
	)
	if err != nil {
		return ActiveAircraftMetric{},
			err
	}

	calculatedAt := s.now().UTC()
	window := time.Duration(windowMinutes) * time.Minute

	query := ActiveAircraftQuery{
		ObservedFrom: calculatedAt.Add(-window),
		ObservedTo:   calculatedAt,
	}

	scope := MetricScope{
		Type: MetricScopeGlobal,
		Code: "world",
	}

	regionCode := strings.TrimSpace(
		strings.ToLower(
			request.RegionCode,
		),
	)
	if regionCode != "" {
		selectedRegion, err := s.regionResolver.GetByCode(
			regionCode,
		)
		if err != nil {
			return ActiveAircraftMetric{},
				err
		}

		query.UseBounds = true
		query.Bounds = Bounds{
			MinLatitude:  selectedRegion.Bounds.MinLatitude,
			MaxLatitude:  selectedRegion.Bounds.MaxLatitude,
			MinLongitude: selectedRegion.Bounds.MinLongitude,
			MaxLongitude: selectedRegion.Bounds.MaxLongitude,
		}

		scope = MetricScope{
			Type: MetricScopeRegion,
			Code: selectedRegion.Code,
		}
	}

	summary, err := s.repository.CountActiveAircraft(
		ctx,
		query,
	)
	if err != nil {
		return ActiveAircraftMetric{},
			err
	}

	return ActiveAircraftMetric{
		Metric:        ActiveAircraftMetricName,
		Value:         summary.Count,
		WindowMinutes: windowMinutes,
		Scope:         scope,
		ObservedFrom:  query.ObservedFrom,
		ObservedTo:    query.ObservedTo,
		CalculatedAt:  calculatedAt,
		Confidence: calculateActiveAircraftConfidence(
			summary,
			window,
			calculatedAt,
		),
		Sources: buildMetricSources(
			summary.SourceNames,
		),
		Limitations: defaultOpenDataLimitations(),
	}, nil
}

func normalizeActiveAircraftWindowMinutes(
	value int,
) (int, error) {
	if value == 0 {
		return DefaultActiveAircraftWindowMinutes,
			nil
	}

	if value < MinimumActiveAircraftWindowMinutes ||
		value > MaximumActiveAircraftWindowMinutes {
		return 0,
			ErrInvalidWindowMinutes
	}

	return value,
		nil
}

func calculateActiveAircraftConfidence(
	summary ActiveAircraftObservationSummary,
	window time.Duration,
	calculatedAt time.Time,
) MetricConfidence {
	if !summary.HasObservations ||
		summary.Count == 0 {
		return MetricConfidence{
			Level: ConfidenceLevelNone,
			Score: 0,
			Reasons: []string{
				"no_recent_open_data_observations",
				"not_official_air_traffic_control_data",
			},
		}
	}

	if summary.LatestObservedAt.After(calculatedAt) {
		return MetricConfidence{
			Level: ConfidenceLevelNone,
			Score: 0,
			Reasons: []string{
				"latest_observation_is_in_future",
				"temporal_evidence_is_invalid",
				"not_official_air_traffic_control_data",
			},
		}
	}

	latestAge := calculatedAt.Sub(
		summary.LatestObservedAt,
	)

	confidence := MetricConfidence{
		Level: ConfidenceLevelLow,
		Score: 0.45,
		Reasons: []string{
			"recent_open_data_observations_available",
			"latest_observation_within_metric_window",
		},
	}

	if latestAge <= 2*time.Minute {
		confidence.Level = ConfidenceLevelHigh
		confidence.Score = 0.9
		confidence.Reasons = append(
			confidence.Reasons,
			"latest_observation_within_two_minutes",
		)
	} else if latestAge <= window/2 {
		confidence.Level = ConfidenceLevelMedium
		confidence.Score = 0.7
		confidence.Reasons = append(
			confidence.Reasons,
			"latest_observation_within_half_window",
		)
	} else {
		confidence.Reasons = append(
			confidence.Reasons,
			"latest_observation_older_than_half_window",
		)
	}

	if len(uniqueSortedStrings(summary.SourceNames)) > 1 {
		confidence.Reasons = append(
			confidence.Reasons,
			"multi_provider_coverage",
		)
	} else {
		confidence.Reasons = append(
			confidence.Reasons,
			"single_provider_coverage",
		)
	}

	confidence.Reasons = append(
		confidence.Reasons,
		"not_official_air_traffic_control_data",
	)

	return confidence
}

func buildMetricSources(
	sourceNames []string,
) []MetricSource {
	uniqueNames := uniqueSortedStrings(
		sourceNames,
	)

	sources := make(
		[]MetricSource,
		0,
		len(uniqueNames),
	)
	for _, name := range uniqueNames {
		sources = append(
			sources,
			MetricSource{
				Name: name,
				Role: "live_position_source",
			},
		)
	}

	return sources
}

func uniqueSortedStrings(
	values []string,
) []string {
	seen := make(
		map[string]struct{},
		len(values),
	)
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		seen[trimmed] = struct{}{}
	}

	result := make(
		[]string,
		0,
		len(seen),
	)
	for value := range seen {
		result = append(
			result,
			value,
		)
	}

	sort.Strings(
		result,
	)

	return result
}

func defaultOpenDataLimitations() []string {
	return []string{
		"open_data_estimate",
		"coverage_depends_on_public_receivers",
		"not_suitable_for_operational_air_traffic_control",
	}
}
