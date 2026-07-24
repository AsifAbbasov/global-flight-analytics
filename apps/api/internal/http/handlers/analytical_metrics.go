package handlers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualityintegration"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricexecution"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricquery"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

var (
	ErrAnalyticalMetricServiceRequired = errors.New(
		"analytical metric service is required",
	)
	ErrAnalyticalQueryServiceRequired = errors.New(
		"analytical trajectory query service is required",
	)
	ErrAnalyticalAirportServiceRequired = errors.New(
		"analytical airport service is required",
	)
)

type AnalyticalMetricService interface {
	ActiveAircraft(
		ctx context.Context,
		request metricexecution.ActiveAircraftRequest,
	) (metricexecution.Execution[int], error)

	TrafficDensity(
		ctx context.Context,
		request metricexecution.TrafficDensityRequest,
	) (metricexecution.Execution[float64], error)

	AirportActivity(
		ctx context.Context,
		request metricexecution.AirportActivityRequest,
	) (metricexecution.Execution[int], error)

	CoverageScore(
		ctx context.Context,
		request metricexecution.CoverageScoreRequest,
	) (metricexecution.Execution[float64], error)

	DataFreshness(
		ctx context.Context,
		request metricexecution.DataFreshnessRequest,
	) (metricexecution.Execution[float64], error)
}

type AnalyticalAirportService interface {
	GetByICAO(
		ctx context.Context,
		icao string,
	) (airport.Airport, error)
}

type AnalyticalTrajectoryQueryService interface {
	Recent(
		ctx context.Context,
		request metricquery.RecentRequest,
	) ([]trajectory.FlightTrajectory, error)

	ByIDs(
		ctx context.Context,
		trajectoryIDs []string,
	) ([]trajectory.FlightTrajectory, error)
}

type AnalyticalMetricsHandler struct {
	metrics  AnalyticalMetricService
	query    AnalyticalTrajectoryQueryService
	airports AnalyticalAirportService
}

func NewAnalyticalMetricsHandler(
	metrics AnalyticalMetricService,
	query AnalyticalTrajectoryQueryService,
) (*AnalyticalMetricsHandler, error) {
	return newAnalyticalMetricsHandler(
		metrics,
		query,
		nil,
	)
}

func NewAnalyticalMetricsHandlerWithAirportService(
	metrics AnalyticalMetricService,
	query AnalyticalTrajectoryQueryService,
	airports AnalyticalAirportService,
) (*AnalyticalMetricsHandler, error) {
	if airports == nil {
		return nil,
			ErrAnalyticalAirportServiceRequired
	}

	return newAnalyticalMetricsHandler(
		metrics,
		query,
		airports,
	)
}

func newAnalyticalMetricsHandler(
	metrics AnalyticalMetricService,
	query AnalyticalTrajectoryQueryService,
	airports AnalyticalAirportService,
) (*AnalyticalMetricsHandler, error) {
	if metrics == nil {
		return nil, ErrAnalyticalMetricServiceRequired
	}
	if query == nil {
		return nil, ErrAnalyticalQueryServiceRequired
	}

	return &AnalyticalMetricsHandler{
		metrics:  metrics,
		query:    query,
		airports: airports,
	}, nil
}

func (handler *AnalyticalMetricsHandler) GetActiveAircraft(
	ctx *fiber.Ctx,
) error {
	recentRequest, err := parseRecentTrajectoryRequest(ctx)
	if err != nil {
		return analyticalBadRequest(ctx, err)
	}

	selectedRegion, err := resolveAnalyticalRegion(
		ctx.Query("region"),
	)
	if err != nil {
		return analyticalRegionError(ctx, err)
	}

	items, err := handler.recentTrajectoriesForRegion(
		ctx.Context(),
		recentRequest,
		selectedRegion,
	)
	if err != nil {
		return analyticalQueryError(ctx, err)
	}

	normalizedWindow, _ := recentRequest.Normalize(time.Now())

	execution, err := handler.metrics.ActiveAircraft(
		ctx.Context(),
		metricexecution.ActiveAircraftRequest{
			Trajectories: items,
			PublicationMetadata: trajectoryPublicationMetadataForRegion(
				items,
				normalizedWindow.Limit,
				selectedRegion,
			),
		},
	)
	if err != nil {
		return analyticalExecutionError(ctx, err)
	}

	return response.OK(
		ctx,
		toAnalyticalMetricResponse(execution),
	)
}

func (handler *AnalyticalMetricsHandler) GetTrafficDensity(
	ctx *fiber.Ctx,
) error {
	recentRequest, err := parseRecentTrajectoryRequest(ctx)
	if err != nil {
		return analyticalBadRequest(ctx, err)
	}

	if strings.TrimSpace(
		ctx.Query("area_square_kilometers"),
	) != "" {
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"AREA_PARAMETER_NOT_SUPPORTED",
			"area_square_kilometers is derived from region and must not be supplied",
		)
	}

	selectedRegion, err := resolveAnalyticalRegion(
		ctx.Query("region"),
	)
	if err != nil {
		return analyticalRegionError(ctx, err)
	}
	if selectedRegion == nil {
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"REGION_REQUIRED",
			"region is required for traffic density",
		)
	}

	area, err := trafficDensityAreaSquareKilometers(
		selectedRegion,
	)
	if err != nil {
		return analyticalRegionError(ctx, err)
	}

	items, err := handler.recentTrajectoriesForRegion(
		ctx.Context(),
		recentRequest,
		selectedRegion,
	)
	if err != nil {
		return analyticalQueryError(ctx, err)
	}

	normalizedWindow, _ := recentRequest.Normalize(time.Now())

	execution, err := handler.metrics.TrafficDensity(
		ctx.Context(),
		metricexecution.TrafficDensityRequest{
			Trajectories:         items,
			AreaSquareKilometers: area,
			PublicationMetadata: trajectoryPublicationMetadataForRegion(
				items,
				normalizedWindow.Limit,
				selectedRegion,
			),
		},
	)
	if err != nil {
		return analyticalExecutionError(ctx, err)
	}

	return response.OK(
		ctx,
		toAnalyticalMetricResponse(execution),
	)
}

func (handler *AnalyticalMetricsHandler) GetAirportActivity(
	ctx *fiber.Ctx,
) error {
	if handler.airports == nil {
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			"ANALYTICAL_AIRPORT_SERVICE_UNAVAILABLE",
			"Airport analytics service is unavailable",
		)
	}

	airportICAO := strings.TrimSpace(
		ctx.Query("airport_icao"),
	)
	if airportICAO == "" {
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"AIRPORT_ICAO_REQUIRED",
			"airport_icao is required",
		)
	}

	radius, err := parseAirportActivityRadius(
		ctx.Query("radius_kilometers"),
	)
	if err != nil {
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_AIRPORT_ACTIVITY_RADIUS",
			"radius_kilometers must be a positive finite number not greater than 100",
		)
	}

	recentRequest, err := parseRecentTrajectoryRequest(ctx)
	if err != nil {
		return analyticalBadRequest(ctx, err)
	}

	selectedAirport, err := handler.airports.GetByICAO(
		ctx.Context(),
		airportICAO,
	)
	if err != nil {
		return analyticalAirportError(ctx, err)
	}

	bounds, err := airportActivityQueryBounds(
		selectedAirport,
		radius,
	)
	if err != nil {
		return analyticalQueryError(ctx, err)
	}

	regionalQuery, ok := handler.query.(regionalAnalyticalTrajectoryQueryService)
	if !ok {
		return analyticalQueryError(
			ctx,
			metricquery.ErrRegionalRepositoryUnsupported,
		)
	}

	items, err := regionalQuery.RecentWithinBounds(
		ctx.Context(),
		recentRequest,
		bounds,
	)
	if err != nil {
		return analyticalQueryError(ctx, err)
	}

	normalizedWindow, _ := recentRequest.Normalize(time.Now())

	execution, err := handler.metrics.AirportActivity(
		ctx.Context(),
		metricexecution.AirportActivityRequest{
			Airport:          selectedAirport,
			Trajectories:     items,
			RadiusKilometers: radius,
			PublicationMetadata: airportActivityPublicationMetadata(
				items,
				normalizedWindow.Limit,
				selectedAirport,
				radius,
			),
		},
	)
	if err != nil {
		return analyticalExecutionError(ctx, err)
	}

	return response.OK(
		ctx,
		toAnalyticalMetricResponse(execution),
	)
}

func (handler *AnalyticalMetricsHandler) GetCoverageScore(
	ctx *fiber.Ctx,
) error {
	input, err := handler.loadProductionQualityInput(
		ctx,
		"observed_samples",
		"expected_samples",
	)
	if err != nil {
		return err
	}

	execution, err := handler.metrics.CoverageScore(
		ctx.Context(),
		metricexecution.CoverageScoreRequest{
			Snapshot:            input.snapshot,
			PublicationMetadata: input.metadata,
		},
	)
	if err != nil {
		return analyticalExecutionError(ctx, err)
	}

	return response.OK(
		ctx,
		toAnalyticalMetricResponse(execution),
	)
}

func (handler *AnalyticalMetricsHandler) GetDataFreshness(
	ctx *fiber.Ctx,
) error {
	input, err := handler.loadProductionQualityInput(
		ctx,
		"observed_at",
		"max_age_seconds",
	)
	if err != nil {
		return err
	}

	execution, err := handler.metrics.DataFreshness(
		ctx.Context(),
		metricexecution.DataFreshnessRequest{
			Snapshot: input.snapshot,
			MaxAge: dataqualityintegration.
				DefaultStaleAfter,
			PublicationMetadata: input.metadata,
		},
	)
	if err != nil {
		return analyticalExecutionError(ctx, err)
	}

	return response.OK(
		ctx,
		toAnalyticalMetricResponse(execution),
	)
}

func (handler *AnalyticalMetricsHandler) loadOptionalTrajectories(
	ctx context.Context,
	ids []string,
) ([]trajectory.FlightTrajectory, error) {
	if len(ids) == 0 {
		return []trajectory.FlightTrajectory{}, nil
	}

	return handler.query.ByIDs(ctx, ids)
}

func parseRecentTrajectoryRequest(
	ctx *fiber.Ctx,
) (metricquery.RecentRequest, error) {
	windowMinutes, err := parseOptionalInteger(
		ctx.Query("window_minutes"),
	)
	if err != nil {
		return metricquery.RecentRequest{},
			metricquery.ErrWindowMinutesInvalid
	}

	limit, err := parseOptionalInteger(
		ctx.Query("limit"),
	)
	if err != nil {
		return metricquery.RecentRequest{},
			metricquery.ErrResultLimitInvalid
	}

	request := metricquery.RecentRequest{
		WindowMinutes: windowMinutes,
		Limit:         limit,
	}

	_, err = request.Normalize(time.Now())
	if err != nil {
		return metricquery.RecentRequest{}, err
	}

	return request, nil
}

func parseOptionalInteger(
	value string,
) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, nil
	}

	return strconv.Atoi(trimmed)
}

func parseRequiredPositiveFloat(
	value string,
) (float64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, errors.New("positive float value is required")
	}

	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || parsed <= 0 || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, errors.New("positive float value is invalid")
	}

	return parsed, nil
}

func parseCSV(
	value string,
) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	parts := strings.Split(trimmed, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}

	return result
}

func trajectoryPublicationMetadata(
	items []trajectory.FlightTrajectory,
	resultLimit int,
) metricexecution.PublicationMetadata {
	limitations := []analyticalresult.Notice{
		{
			Code:    "open_data_coverage",
			Message: "Coverage depends on publicly available aviation receivers and providers.",
		},
		{
			Code:    "not_operational_air_traffic_control",
			Message: "This analytical result is not suitable for operational air traffic control.",
		},
	}

	if resultLimit > 0 && len(items) >= resultLimit {
		limitations = append(
			limitations,
			analyticalresult.Notice{
				Code:    "trajectory_result_limit_reached",
				Message: "The trajectory query reached its configured result limit; additional contributors may exist.",
			},
		)
	}

	retrievedAt := time.Now().UTC()
	sources, unattributedCount :=
		analyticalSourcesFromTrajectories(
			items,
			retrievedAt,
		)

	if unattributedCount > 0 {
		limitations = append(
			limitations,
			analyticalresult.Notice{
				Code: "unattributed_source_observations",
				Message: fmt.Sprintf(
					"%d trajectory observations did not provide a source name and were omitted from source provenance.",
					unattributedCount,
				),
			},
		)
	}

	metadata := metricexecution.PublicationMetadata{
		Sources:     sources,
		Limitations: limitations,
	}

	report, reportErr := dataqualityintegration.BuildTrajectoryReport(
		dataqualityintegration.TrajectoryReportRequest{
			Trajectories: items,
			EvaluatedAt:  retrievedAt,
		},
	)
	if reportErr != nil {
		metadata.Limitations = append(
			metadata.Limitations,
			analyticalresult.Notice{
				Code:    "data_quality_report_unavailable",
				Message: "The trajectory data-quality report could not be constructed from the retained analytical inputs.",
			},
		)
		return metadata
	}

	metadata.DataQuality = report
	return metadata
}

func analyticalSourcesFromTrajectories(
	items []trajectory.FlightTrajectory,
	retrievedAt time.Time,
) ([]analyticalresult.Source, int) {
	type sourceWindow struct {
		from     time.Time
		to       time.Time
		hasRange bool
	}

	windows := make(map[string]sourceWindow)
	unattributedCount := 0

	for _, item := range items {
		name := strings.TrimSpace(item.SourceName)
		if name == "" {
			unattributedCount++
			continue
		}

		window := windows[name]
		if !item.StartTime.IsZero() && !item.EndTime.IsZero() {
			if !window.hasRange || item.StartTime.Before(window.from) {
				window.from = item.StartTime.UTC()
			}
			if !window.hasRange || item.EndTime.After(window.to) {
				window.to = item.EndTime.UTC()
			}
			window.hasRange = true
		}
		windows[name] = window
	}

	names := make([]string, 0, len(windows))
	for name := range windows {
		names = append(names, name)
	}
	sort.Strings(names)

	sources := make([]analyticalresult.Source, 0, len(names))
	for _, name := range names {
		window := windows[name]
		source := analyticalresult.Source{
			Name:        name,
			Role:        analyticalresult.SourceRoleObservation,
			RetrievedAt: retrievedAt,
		}
		if window.hasRange {
			source.ObservedFrom = window.from
			source.ObservedTo = window.to
		}
		sources = append(sources, source)
	}

	return sources, unattributedCount
}

func analyticalBadRequest(
	ctx *fiber.Ctx,
	err error,
) error {
	switch {
	case errors.Is(err, metricquery.ErrWindowMinutesInvalid):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_WINDOW_MINUTES",
			"window_minutes must be an integer between 1 and 180",
		)
	case errors.Is(err, metricquery.ErrResultLimitInvalid):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_RESULT_LIMIT",
			"limit must be an integer between 1 and 5000",
		)
	default:
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_ANALYTICAL_REQUEST",
			"Analytical request is invalid",
		)
	}
}

func analyticalQueryError(
	ctx *fiber.Ctx,
	err error,
) error {
	if errors.Is(err, metricquery.ErrTrajectoryIDsMissing) ||
		errors.Is(err, metricquery.ErrTrajectoryIDInvalid) ||
		errors.Is(err, metricquery.ErrTrajectoryIDCountExceeded) {
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_TRAJECTORY_IDS",
			"Trajectory identifiers are invalid",
		)
	}

	return response.Error(
		ctx,
		fiber.StatusInternalServerError,
		"ANALYTICAL_QUERY_FAILED",
		"Failed to load analytical trajectory data",
	)
}

func analyticalExecutionError(
	ctx *fiber.Ctx,
	_ error,
) error {
	return response.Error(
		ctx,
		fiber.StatusInternalServerError,
		"ANALYTICAL_METRIC_FAILED",
		"Failed to execute analytical metric",
	)
}
