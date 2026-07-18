package airportproduction

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/history"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/overview"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/passport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/ranking"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/trends"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

const defaultMaximumDataAge = 48 * time.Hour

type Config struct {
	AirportRepository airport.Repository
	ObservationReader ObservationReader
	MaximumDataAge    time.Duration
	Now               func() time.Time
}

type Service struct {
	airportRepository airport.Repository
	observationReader ObservationReader
	statistics        statistics.Calculator
	passport          passport.Builder
	ranker            ranking.Ranker
	overview          overview.Assembler
	history           history.Builder
	trends            trends.Analyzer
	now               func() time.Time
}

func New(config Config) (*Service, error) {
	if config.AirportRepository == nil {
		return nil, fmt.Errorf("%w: airport repository is required", ErrInvalidConfiguration)
	}
	if config.ObservationReader == nil {
		return nil, fmt.Errorf("%w: observation reader is required", ErrInvalidConfiguration)
	}
	maximumDataAge := config.MaximumDataAge
	if maximumDataAge == 0 {
		maximumDataAge = defaultMaximumDataAge
	}
	statisticsCalculator, err := statistics.NewCalculator(maximumDataAge)
	if err != nil {
		return nil, fmt.Errorf("%w: create statistics calculator: %v", ErrInvalidConfiguration, err)
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return &Service{
		airportRepository: config.AirportRepository,
		observationReader: config.ObservationReader,
		statistics:        statisticsCalculator,
		passport:          passport.NewBuilder(),
		ranker:            ranking.NewRanker(),
		overview:          overview.NewAssembler(),
		history:           history.NewBuilder(),
		trends:            trends.NewAnalyzer(),
		now:               now,
	}, nil
}

func (service *Service) GetOverview(ctx context.Context, icaoCode string, request WindowRequest) (OverviewResult, error) {
	if service == nil {
		return OverviewResult{}, ErrInvalidConfiguration
	}
	ctx = nonNilContext(ctx)
	window, err := service.normalizeWindow(request)
	if err != nil {
		return OverviewResult{}, err
	}
	normalizedICAO, err := normalizeICAO(icaoCode)
	if err != nil {
		return OverviewResult{}, err
	}
	airportValue, err := service.airportRepository.GetByICAO(ctx, normalizedICAO)
	if err != nil {
		return OverviewResult{}, err
	}
	statisticsByICAO, err := service.loadAggregateStatistics(ctx, window)
	if err != nil {
		return OverviewResult{}, err
	}
	selectedStatistics, exists := statisticsByICAO[normalizedICAO]
	if !exists {
		return OverviewResult{}, fmt.Errorf("%w: %s", ErrObservationsNotFound, normalizedICAO)
	}
	rankingResult, err := service.rankStatistics(statisticsByICAO)
	if err != nil {
		return OverviewResult{}, err
	}
	generatedAt := service.now().UTC()
	passportValue, err := service.passport.Build(airportValue, passport.AnalyticsInput{
		Arrivals:       selectedStatistics.Arrivals,
		Departures:     selectedStatistics.Departures,
		ActiveAircraft: selectedStatistics.ActiveAircraft,
		FreshnessScore: selectedStatistics.FreshnessScore,
		CoverageScore:  selectedStatistics.CoverageScore,
		ObservedAt:     selectedStatistics.LatestObservationAt,
	}, generatedAt)
	if err != nil {
		return OverviewResult{}, fmt.Errorf("build Airport Passport: %w", err)
	}
	overviewValue, err := service.overview.Assemble(overview.Input{
		Passport:      passportValue,
		Statistics:    selectedStatistics,
		RankingResult: rankingResult,
		GeneratedAt:   generatedAt,
	})
	if err != nil {
		return OverviewResult{}, fmt.Errorf("assemble Airport Overview: %w", err)
	}
	return OverviewResult{Version: Version, Window: window, Overview: overviewValue, Limitations: productionLimitations(), GeneratedAt: generatedAt}, nil
}

func (service *Service) GetHistory(ctx context.Context, icaoCode string, request WindowRequest) (HistoryResult, error) {
	if service == nil {
		return HistoryResult{}, ErrInvalidConfiguration
	}
	ctx = nonNilContext(ctx)
	window, err := service.normalizeWindow(request)
	if err != nil {
		return HistoryResult{}, err
	}
	normalizedICAO, err := normalizeICAO(icaoCode)
	if err != nil {
		return HistoryResult{}, err
	}
	if _, err := service.airportRepository.GetByICAO(ctx, normalizedICAO); err != nil {
		return HistoryResult{}, err
	}
	observations, err := service.observationReader.ListDaily(ctx, DailyQuery{ICAOCode: normalizedICAO, WindowStart: window.StartTime, WindowEnd: window.EndTime})
	if err != nil {
		return HistoryResult{}, fmt.Errorf("read Airport Intelligence observations: %w", err)
	}
	if len(observations) == 0 {
		return HistoryResult{}, fmt.Errorf("%w: %s", ErrObservationsNotFound, normalizedICAO)
	}
	entries := make([]statistics.Statistics, 0, len(observations))
	for _, observation := range observations {
		entry, calculateErr := service.calculateDailyStatistics(observation)
		if calculateErr != nil {
			return HistoryResult{}, calculateErr
		}
		entries = append(entries, entry)
	}
	generatedAt := service.now().UTC()
	historyValue, err := service.history.Build(history.Input{ICAOCode: normalizedICAO, Entries: entries, GeneratedAt: generatedAt})
	if err != nil {
		return HistoryResult{}, fmt.Errorf("build Airport History: %w", err)
	}
	return HistoryResult{Version: Version, Window: window, History: historyValue, Limitations: productionLimitations(), GeneratedAt: generatedAt}, nil
}

func (service *Service) GetTrends(ctx context.Context, icaoCode string, request WindowRequest) (TrendsResult, error) {
	historyResult, err := service.GetHistory(ctx, icaoCode, request)
	if err != nil {
		return TrendsResult{}, err
	}
	if len(historyResult.History.Entries) < 2 {
		return TrendsResult{}, fmt.Errorf("%w: at least two observed daily windows are required", ErrInsufficientHistory)
	}
	generatedAt := service.now().UTC()
	trendValue, err := service.trends.Analyze(trends.Input{History: historyResult.History, GeneratedAt: generatedAt})
	if err != nil {
		if errors.Is(err, trends.ErrInsufficientHistory) {
			return TrendsResult{}, fmt.Errorf("%w: %v", ErrInsufficientHistory, err)
		}
		return TrendsResult{}, fmt.Errorf("analyze Airport Trends: %w", err)
	}
	return TrendsResult{Version: Version, Window: historyResult.Window, Trends: trendValue, Limitations: productionLimitations(), GeneratedAt: generatedAt}, nil
}

func (service *Service) GetRanking(ctx context.Context, request WindowRequest) (RankingResult, error) {
	if service == nil {
		return RankingResult{}, ErrInvalidConfiguration
	}
	ctx = nonNilContext(ctx)
	window, err := service.normalizeWindow(request)
	if err != nil {
		return RankingResult{}, err
	}
	statisticsByICAO, err := service.loadAggregateStatistics(ctx, window)
	if err != nil {
		return RankingResult{}, err
	}
	rankingResult, err := service.rankStatistics(statisticsByICAO)
	if err != nil {
		return RankingResult{}, err
	}
	airportValues, err := service.airportRepository.List(ctx)
	if err != nil {
		return RankingResult{}, fmt.Errorf("list airports for ranking: %w", err)
	}
	airportsByICAO := make(map[string]airport.Airport, len(airportValues))
	for _, airportValue := range airportValues {
		normalized := strings.ToUpper(strings.TrimSpace(airportValue.ICAOCode))
		if normalized == "" {
			continue
		}
		airportValue.ICAOCode = normalized
		airportsByICAO[normalized] = airportValue
	}
	return RankingResult{Version: Version, Window: window, Ranking: rankingResult, Airports: airportsByICAO, Limitations: productionLimitations(), GeneratedAt: service.now().UTC()}, nil
}

func (service *Service) loadAggregateStatistics(ctx context.Context, window Window) (map[string]statistics.Statistics, error) {
	observations, err := service.observationReader.ListDaily(ctx, DailyQuery{WindowStart: window.StartTime, WindowEnd: window.EndTime})
	if err != nil {
		return nil, fmt.Errorf("read Airport Intelligence observations: %w", err)
	}
	if len(observations) == 0 {
		return nil, ErrObservationsNotFound
	}
	grouped := make(map[string][]DailyObservation)
	for _, observation := range observations {
		normalizedICAO, normalizeErr := normalizeICAO(observation.ICAOCode)
		if normalizeErr != nil {
			return nil, fmt.Errorf("normalize Airport Intelligence observation: %w", normalizeErr)
		}
		observation.ICAOCode = normalizedICAO
		grouped[normalizedICAO] = append(grouped[normalizedICAO], observation)
	}
	result := make(map[string]statistics.Statistics, len(grouped))
	for icaoCode, items := range grouped {
		aggregate, aggregateErr := service.calculateAggregateStatistics(icaoCode, items, window)
		if aggregateErr != nil {
			return nil, aggregateErr
		}
		result[icaoCode] = aggregate
	}
	return result, nil
}

func (service *Service) calculateAggregateStatistics(icaoCode string, observations []DailyObservation, window Window) (statistics.Statistics, error) {
	if len(observations) == 0 {
		return statistics.Statistics{}, ErrObservationsNotFound
	}
	sort.Slice(observations, func(left, right int) bool {
		return observations[left].WindowStart.Before(observations[right].WindowStart)
	})
	arrivals, departures, peakActiveAircraft, peakActiveRoutes := 0, 0, 0, 0
	latestObservationAt := time.Time{}
	for _, observation := range observations {
		if err := validateDailyObservation(observation, window); err != nil {
			return statistics.Statistics{}, err
		}
		arrivals += observation.Arrivals
		departures += observation.Departures
		if observation.ActiveAircraft > peakActiveAircraft {
			peakActiveAircraft = observation.ActiveAircraft
		}
		if observation.ActiveRoutes > peakActiveRoutes {
			peakActiveRoutes = observation.ActiveRoutes
		}
		if observation.ObservedAt.After(latestObservationAt) {
			latestObservationAt = observation.ObservedAt
		}
	}
	value, err := service.statistics.Calculate(statistics.Input{
		ICAOCode:            icaoCode,
		WindowStart:         window.StartTime,
		WindowEnd:           window.EndTime,
		Arrivals:            arrivals,
		Departures:          departures,
		ActiveAircraft:      peakActiveAircraft,
		ActiveRoutes:        peakActiveRoutes,
		ObservedSamples:     len(observations),
		ExpectedSamples:     window.CompletedDays,
		LatestObservationAt: latestObservationAt,
		GeneratedAt:         window.EndTime,
	})
	if err != nil {
		return statistics.Statistics{}, fmt.Errorf("calculate aggregate Airport Statistics for %s: %w", icaoCode, err)
	}
	return value, nil
}

func (service *Service) calculateDailyStatistics(observation DailyObservation) (statistics.Statistics, error) {
	dailyWindow := Window{StartTime: observation.WindowStart, EndTime: observation.WindowEnd, AsOfTime: observation.WindowEnd, CompletedDays: 1}
	if err := validateDailyObservation(observation, dailyWindow); err != nil {
		return statistics.Statistics{}, err
	}
	value, err := service.statistics.Calculate(statistics.Input{
		ICAOCode:            observation.ICAOCode,
		WindowStart:         observation.WindowStart,
		WindowEnd:           observation.WindowEnd,
		Arrivals:            observation.Arrivals,
		Departures:          observation.Departures,
		ActiveAircraft:      observation.ActiveAircraft,
		ActiveRoutes:        observation.ActiveRoutes,
		ObservedSamples:     1,
		ExpectedSamples:     1,
		LatestObservationAt: observation.ObservedAt,
		GeneratedAt:         observation.WindowEnd,
	})
	if err != nil {
		return statistics.Statistics{}, fmt.Errorf("calculate daily Airport Statistics for %s: %w", observation.ICAOCode, err)
	}
	return value, nil
}

func (service *Service) rankStatistics(statisticsByICAO map[string]statistics.Statistics) (ranking.Result, error) {
	if len(statisticsByICAO) == 0 {
		return ranking.Result{}, ErrObservationsNotFound
	}
	items := make([]statistics.Statistics, 0, len(statisticsByICAO))
	for _, item := range statisticsByICAO {
		items = append(items, item)
	}
	sort.Slice(items, func(left, right int) bool { return items[left].ICAOCode < items[right].ICAOCode })
	result, err := service.ranker.Rank(ranking.Input{Statistics: items})
	if err != nil {
		return ranking.Result{}, fmt.Errorf("rank airports: %w", err)
	}
	return result, nil
}

func (service *Service) normalizeWindow(request WindowRequest) (Window, error) {
	days := request.Days
	if days == 0 {
		days = DefaultWindowDays
	}
	if days < MinimumWindowDays || days > MaximumWindowDays {
		return Window{}, fmt.Errorf("%w: days must be between %d and %d", ErrInvalidRequest, MinimumWindowDays, MaximumWindowDays)
	}
	asOfTime := request.AsOfTime
	if asOfTime.IsZero() {
		asOfTime = service.now()
	}
	asOfTime = asOfTime.UTC()
	if asOfTime.Year() < 1970 {
		return Window{}, fmt.Errorf("%w: as-of time must not precede 1970", ErrInvalidRequest)
	}
	windowEnd := time.Date(asOfTime.Year(), asOfTime.Month(), asOfTime.Day(), 0, 0, 0, 0, time.UTC)
	windowStart := windowEnd.Add(-time.Duration(days) * 24 * time.Hour)
	return Window{StartTime: windowStart, EndTime: windowEnd, AsOfTime: asOfTime, CompletedDays: days}, nil
}

func validateDailyObservation(observation DailyObservation, parentWindow Window) error {
	normalizedICAO, err := normalizeICAO(observation.ICAOCode)
	if err != nil {
		return err
	}
	if observation.WindowStart.IsZero() || observation.WindowEnd.IsZero() || !observation.WindowEnd.After(observation.WindowStart) {
		return fmt.Errorf("%w: %s has an invalid daily window", ErrInvalidRequest, normalizedICAO)
	}
	if observation.WindowStart.Before(parentWindow.StartTime) || observation.WindowEnd.After(parentWindow.EndTime) {
		return fmt.Errorf("%w: %s observation is outside the request window", ErrInvalidRequest, normalizedICAO)
	}
	if observation.Arrivals < 0 || observation.Departures < 0 || observation.ActiveAircraft < 0 || observation.ActiveRoutes < 0 {
		return fmt.Errorf("%w: %s observation counters cannot be negative", ErrInvalidRequest, normalizedICAO)
	}
	if observation.ObservedAt.IsZero() || observation.ObservedAt.Before(observation.WindowStart) || !observation.ObservedAt.Before(observation.WindowEnd) {
		return fmt.Errorf("%w: %s observed time is outside its daily window", ErrInvalidRequest, normalizedICAO)
	}
	return nil
}

func normalizeICAO(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if len(normalized) != 4 {
		return "", fmt.Errorf("%w: ICAO code must contain four characters", ErrInvalidRequest)
	}
	for _, character := range normalized {
		if character < 'A' || character > 'Z' {
			return "", fmt.Errorf("%w: ICAO code must contain only Latin letters", ErrInvalidRequest)
		}
	}
	return normalized, nil
}

func productionLimitations() []Limitation {
	return []Limitation{
		{Code: "OPEN_DATA_NOT_OFFICIAL_OPERATIONS", Message: "Airport Intelligence uses project open-data tables and does not represent official airport operations."},
		{Code: "COMPLETED_UTC_DAYS_ONLY", Message: "The requested window contains completed Coordinated Universal Time days and excludes the current partial day."},
		{Code: "RELATIVE_RANKING", Message: "Activity scores are relative to airports with observations in the same comparison window."},
		{Code: "ROUTE_DERIVED_CONTEXT_MAY_BE_INCOMPLETE", Message: "Active-route and active-aircraft context depends on available project route records and can be incomplete."},
	}
}

func nonNilContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
