package ranking

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

const scoreMaximum = 100.0

var DefaultActivityWeights = ActivityWeights{
	Movements:    0.25,
	Routes:       0.25,
	Observations: 0.25,
	Intensity:    0.25,
}

var DefaultConfidenceWeights = ConfidenceWeights{
	Coverage:  0.50,
	Freshness: 0.50,
}

type Ranker struct{}

func NewRanker() Ranker {
	return Ranker{}
}

func (Ranker) Rank(input Input) (Result, error) {
	activityWeights, err := normalizeActivityWeights(input.ActivityWeights)
	if err != nil {
		return Result{}, err
	}

	confidenceWeights, err := normalizeConfidenceWeights(input.ConfidenceWeights)
	if err != nil {
		return Result{}, err
	}

	if len(input.Statistics) == 0 {
		return Result{}, fmt.Errorf("%w: at least one airport is required", ErrInvalidInput)
	}

	windowStart := input.Statistics[0].WindowStart
	windowEnd := input.Statistics[0].WindowEnd
	if windowStart.IsZero() || windowEnd.IsZero() || !windowEnd.After(windowStart) {
		return Result{}, fmt.Errorf("%w: valid statistics window is required", ErrInvalidInput)
	}

	seen := make(map[string]struct{}, len(input.Statistics))
	prepared := make([]statistics.Statistics, 0, len(input.Statistics))
	maximums := componentMaximums{}

	for _, item := range input.Statistics {
		normalized, normalizeErr := normalizeStatistics(item, windowStart, windowEnd)
		if normalizeErr != nil {
			return Result{}, normalizeErr
		}
		if _, exists := seen[normalized.ICAOCode]; exists {
			return Result{}, fmt.Errorf("%w: %s", ErrDuplicateAirport, normalized.ICAOCode)
		}
		seen[normalized.ICAOCode] = struct{}{}
		maximums.observe(normalized)
		prepared = append(prepared, normalized)
	}

	ranked := make([]AirportRank, 0, len(prepared))
	for _, item := range prepared {
		movementsComponent := normalizedScore(float64(item.TotalMovements), float64(maximums.totalMovements))
		routesComponent := normalizedScore(float64(item.ActiveRoutes), float64(maximums.activeRoutes))
		observationsComponent := normalizedScore(float64(item.ObservedSamples), float64(maximums.observedSamples))
		intensityComponent := normalizedScore(item.MovementsPerHour, maximums.movementsPerHour)

		activityScore :=
			movementsComponent*activityWeights.Movements +
				routesComponent*activityWeights.Routes +
				observationsComponent*activityWeights.Observations +
				intensityComponent*activityWeights.Intensity

		dataConfidence := scoreMaximum * (item.CoverageScore*confidenceWeights.Coverage +
			item.FreshnessScore*confidenceWeights.Freshness)

		ranked = append(ranked, AirportRank{
			ICAOCode:              item.ICAOCode,
			ActivityScore:         activityScore,
			DataConfidence:        dataConfidence,
			MovementsComponent:    movementsComponent,
			RoutesComponent:       routesComponent,
			ObservationsComponent: observationsComponent,
			IntensityComponent:    intensityComponent,
			CoverageScore:         item.CoverageScore,
			FreshnessScore:        item.FreshnessScore,
			TotalMovements:        item.TotalMovements,
			ActiveRoutes:          item.ActiveRoutes,
			ObservedSamples:       item.ObservedSamples,
			ExpectedSamples:       item.ExpectedSamples,
			MovementsPerHour:      item.MovementsPerHour,
			ActiveAircraft:        item.ActiveAircraft,
		})
	}

	sort.SliceStable(ranked, func(left, right int) bool {
		if ranked[left].ActivityScore != ranked[right].ActivityScore {
			return ranked[left].ActivityScore > ranked[right].ActivityScore
		}
		if ranked[left].TotalMovements != ranked[right].TotalMovements {
			return ranked[left].TotalMovements > ranked[right].TotalMovements
		}
		if ranked[left].ActiveRoutes != ranked[right].ActiveRoutes {
			return ranked[left].ActiveRoutes > ranked[right].ActiveRoutes
		}
		if ranked[left].ObservedSamples != ranked[right].ObservedSamples {
			return ranked[left].ObservedSamples > ranked[right].ObservedSamples
		}
		if ranked[left].MovementsPerHour != ranked[right].MovementsPerHour {
			return ranked[left].MovementsPerHour > ranked[right].MovementsPerHour
		}
		return ranked[left].ICAOCode < ranked[right].ICAOCode
	})

	for index := range ranked {
		ranked[index].Position = index + 1
	}

	return Result{
		WindowStart:       windowStart.UTC(),
		WindowEnd:         windowEnd.UTC(),
		ActivityWeights:   activityWeights,
		ConfidenceWeights: confidenceWeights,
		Airports:          ranked,
	}, nil
}

type componentMaximums struct {
	totalMovements   int
	activeRoutes     int
	observedSamples  int
	movementsPerHour float64
}

func (maximums *componentMaximums) observe(item statistics.Statistics) {
	if item.TotalMovements > maximums.totalMovements {
		maximums.totalMovements = item.TotalMovements
	}
	if item.ActiveRoutes > maximums.activeRoutes {
		maximums.activeRoutes = item.ActiveRoutes
	}
	if item.ObservedSamples > maximums.observedSamples {
		maximums.observedSamples = item.ObservedSamples
	}
	if item.MovementsPerHour > maximums.movementsPerHour {
		maximums.movementsPerHour = item.MovementsPerHour
	}
}

func normalizedScore(value, maximum float64) float64 {
	if maximum <= 0 {
		return 0
	}
	return value / maximum * scoreMaximum
}

func normalizeActivityWeights(weights ActivityWeights) (ActivityWeights, error) {
	if weights == (ActivityWeights{}) {
		weights = DefaultActivityWeights
	}
	if invalidWeight(weights.Movements) || invalidWeight(weights.Routes) ||
		invalidWeight(weights.Observations) || invalidWeight(weights.Intensity) {
		return ActivityWeights{}, fmt.Errorf("%w: activity weights must be finite and non-negative", ErrInvalidConfiguration)
	}

	total := weights.Movements + weights.Routes + weights.Observations + weights.Intensity
	if total <= 0 {
		return ActivityWeights{}, fmt.Errorf("%w: at least one positive activity weight is required", ErrInvalidConfiguration)
	}

	return ActivityWeights{
		Movements:    weights.Movements / total,
		Routes:       weights.Routes / total,
		Observations: weights.Observations / total,
		Intensity:    weights.Intensity / total,
	}, nil
}

func normalizeConfidenceWeights(weights ConfidenceWeights) (ConfidenceWeights, error) {
	if weights == (ConfidenceWeights{}) {
		weights = DefaultConfidenceWeights
	}
	if invalidWeight(weights.Coverage) || invalidWeight(weights.Freshness) {
		return ConfidenceWeights{}, fmt.Errorf("%w: confidence weights must be finite and non-negative", ErrInvalidConfiguration)
	}

	total := weights.Coverage + weights.Freshness
	if total <= 0 {
		return ConfidenceWeights{}, fmt.Errorf("%w: at least one positive confidence weight is required", ErrInvalidConfiguration)
	}

	return ConfidenceWeights{
		Coverage:  weights.Coverage / total,
		Freshness: weights.Freshness / total,
	}, nil
}

func invalidWeight(value float64) bool {
	return math.IsNaN(value) || math.IsInf(value, 0) || value < 0
}

func normalizeStatistics(
	item statistics.Statistics,
	windowStart time.Time,
	windowEnd time.Time,
) (statistics.Statistics, error) {
	item.ICAOCode = strings.ToUpper(strings.TrimSpace(item.ICAOCode))
	if item.ICAOCode == "" {
		return statistics.Statistics{}, fmt.Errorf("%w: ICAO code is required", ErrInvalidInput)
	}
	if !item.WindowStart.Equal(windowStart) || !item.WindowEnd.Equal(windowEnd) {
		return statistics.Statistics{}, fmt.Errorf("%w: %s", ErrIncomparableWindow, item.ICAOCode)
	}
	if item.TotalMovements < 0 || item.ActiveRoutes < 0 || item.ObservedSamples < 0 ||
		item.ExpectedSamples <= 0 || item.ActiveAircraft < 0 || invalidMetric(item.MovementsPerHour) {
		return statistics.Statistics{}, fmt.Errorf("%w: airport counters and intensity are invalid", ErrInvalidInput)
	}
	if !qualityScoreInRange(item.CoverageScore) {
		return statistics.Statistics{}, fmt.Errorf("%w: coverage score must be between 0 and 1", ErrInvalidInput)
	}
	if !qualityScoreInRange(item.FreshnessScore) {
		return statistics.Statistics{}, fmt.Errorf("%w: freshness score must be between 0 and 1", ErrInvalidInput)
	}

	return item, nil
}

func invalidMetric(value float64) bool {
	return math.IsNaN(value) || math.IsInf(value, 0) || value < 0
}

func qualityScoreInRange(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}
