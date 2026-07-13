package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricexecution"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricquery"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/gofiber/fiber/v2"
)

type analyticalQueryStub struct {
	recentItems []trajectory.FlightTrajectory
	idItems     []trajectory.FlightTrajectory
	err         error
	recent      metricquery.RecentRequest
	ids         []string
}

func (stub *analyticalQueryStub) Recent(
	ctx context.Context,
	request metricquery.RecentRequest,
) ([]trajectory.FlightTrajectory, error) {
	stub.recent = request
	return stub.recentItems, stub.err
}

func (stub *analyticalQueryStub) ByIDs(
	ctx context.Context,
	ids []string,
) ([]trajectory.FlightTrajectory, error) {
	stub.ids = append([]string(nil), ids...)
	return stub.idItems, stub.err
}

type analyticalMetricStub struct {
	activeRequest    metricexecution.ActiveAircraftRequest
	densityRequest   metricexecution.TrafficDensityRequest
	airportRequest   metricexecution.AirportActivityRequest
	coverageRequest  metricexecution.CoverageScoreRequest
	freshnessRequest metricexecution.DataFreshnessRequest
	err              error
}

func (stub *analyticalMetricStub) ActiveAircraft(
	ctx context.Context,
	request metricexecution.ActiveAircraftRequest,
) (metricexecution.Execution[int], error) {
	stub.activeRequest = request
	return successfulExecution(
		metricexecution.MetricIDActiveAircraft,
		2,
	), stub.err
}

func (stub *analyticalMetricStub) TrafficDensity(
	ctx context.Context,
	request metricexecution.TrafficDensityRequest,
) (metricexecution.Execution[float64], error) {
	stub.densityRequest = request
	return successfulExecution(
		"traffic_density",
		0.02,
	), stub.err
}

func (stub *analyticalMetricStub) AirportActivity(
	ctx context.Context,
	request metricexecution.AirportActivityRequest,
) (metricexecution.Execution[int], error) {
	stub.airportRequest = request
	return successfulExecution(
		metricexecution.MetricIDAirportActivity,
		2,
	), stub.err
}

func (stub *analyticalMetricStub) CoverageScore(
	ctx context.Context,
	request metricexecution.CoverageScoreRequest,
) (metricexecution.Execution[float64], error) {
	stub.coverageRequest = request
	return successfulExecution(
		"coverage_score",
		0.75,
	), stub.err
}

func (stub *analyticalMetricStub) DataFreshness(
	ctx context.Context,
	request metricexecution.DataFreshnessRequest,
) (metricexecution.Execution[float64], error) {
	stub.freshnessRequest = request
	return successfulExecution(
		"data_freshness",
		0.50,
	), stub.err
}

func successfulExecution[T any](
	metricID string,
	value T,
) metricexecution.Execution[T] {
	evaluatedAt := time.Date(
		2026,
		time.July,
		14,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	return metricexecution.Execution[T]{
		MetricID: metricID,
		Result: analyticalresult.Result[T]{
			Status:   analyticalresult.StatusComplete,
			Value:    value,
			HasValue: true,
			Confidence: analyticalresult.Confidence{
				Level: analyticalresult.ConfidenceLevelHigh,
				Score: 0.90,
				Reasons: []analyticalresult.Notice{
					{
						Code:    "test_confidence",
						Message: "Test confidence is high.",
					},
				},
			},
			CalculatedAt: evaluatedAt,
		},
		Scope: metricexecution.ScopeSummary{
			InputCount:   2,
			AllowedCount: 2,
			EvaluatedAt:  evaluatedAt,
		},
	}
}

func TestAnalyticalActiveAircraftEndpoint(
	t *testing.T,
) {
	query := &analyticalQueryStub{
		recentItems: []trajectory.FlightTrajectory{
			{ID: "one", SourceName: "airplanes.live"},
			{ID: "two", SourceName: "airplanes.live"},
		},
	}
	metrics := &analyticalMetricStub{}
	handler, err := NewAnalyticalMetricsHandler(
		metrics,
		query,
	)
	if err != nil {
		t.Fatalf("expected handler, got %v", err)
	}

	app := fiber.New()
	app.Get("/metric", handler.GetActiveAircraft)

	request := httptest.NewRequest(
		fiber.MethodGet,
		"/metric?window_minutes=30&limit=20",
		nil,
	)
	result, err := app.Test(request)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", result.StatusCode)
	}
	if query.recent.WindowMinutes != 30 ||
		query.recent.Limit != 20 {
		t.Fatalf("unexpected recent request: %#v", query.recent)
	}
	if len(metrics.activeRequest.Trajectories) != 2 {
		t.Fatalf("expected two trajectories, got %#v", metrics.activeRequest.Trajectories)
	}

	var body struct {
		Success bool       `json:"success"`
		Data    mapPayload `json:"data"`
	}
	if err := json.NewDecoder(result.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success ||
		body.Data.Metric != metricexecution.MetricIDActiveAircraft ||
		body.Data.Status != string(analyticalresult.StatusComplete) ||
		body.Data.Value != 2 {
		t.Fatalf("unexpected response: %#v", body)
	}
}

type mapPayload struct {
	Metric string `json:"metric"`
	Status string `json:"status"`
	Value  int    `json:"value"`
}

func TestAnalyticalActiveAircraftRejectsInvalidWindow(
	t *testing.T,
) {
	query := &analyticalQueryStub{}
	handler, err := NewAnalyticalMetricsHandler(
		&analyticalMetricStub{},
		query,
	)
	if err != nil {
		t.Fatalf("expected handler, got %v", err)
	}

	app := fiber.New()
	app.Get("/metric", handler.GetActiveAircraft)
	result, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/metric?window_minutes=181",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", result.StatusCode)
	}
	if query.recent.WindowMinutes != 0 {
		t.Fatal("expected query service not to be called")
	}
}

func TestAnalyticalTrafficDensityRequiresArea(
	t *testing.T,
) {
	handler, err := NewAnalyticalMetricsHandler(
		&analyticalMetricStub{},
		&analyticalQueryStub{},
	)
	if err != nil {
		t.Fatalf("expected handler, got %v", err)
	}

	app := fiber.New()
	app.Get("/metric", handler.GetTrafficDensity)
	result, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/metric",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", result.StatusCode)
	}
}

func TestAnalyticalAirportActivityLoadsSelectedTrajectories(
	t *testing.T,
) {
	query := &analyticalQueryStub{
		idItems: []trajectory.FlightTrajectory{
			{ID: "arrival"},
		},
	}
	metrics := &analyticalMetricStub{}
	handler, err := NewAnalyticalMetricsHandler(metrics, query)
	if err != nil {
		t.Fatalf("expected handler, got %v", err)
	}

	app := fiber.New()
	app.Get("/metric", handler.GetAirportActivity)
	result, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/metric?arrival_trajectory_ids=arrival",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", result.StatusCode)
	}
	if len(metrics.airportRequest.Arrivals) != 1 ||
		len(metrics.airportRequest.Departures) != 0 {
		t.Fatalf("unexpected airport request: %#v", metrics.airportRequest)
	}
}

func TestAnalyticalCoverageScoreParsesSnapshot(
	t *testing.T,
) {
	metrics := &analyticalMetricStub{}
	handler, err := NewAnalyticalMetricsHandler(
		metrics,
		&analyticalQueryStub{},
	)
	if err != nil {
		t.Fatalf("expected handler, got %v", err)
	}

	app := fiber.New()
	app.Get("/metric", handler.GetCoverageScore)
	result, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/metric?observed_samples=75&expected_samples=100",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", result.StatusCode)
	}
	if metrics.coverageRequest.Snapshot.ObservedSamples != 75 ||
		metrics.coverageRequest.Snapshot.ExpectedSamples != 100 {
		t.Fatalf("unexpected coverage request: %#v", metrics.coverageRequest)
	}
}

func TestAnalyticalDataFreshnessParsesTimestampAndAge(
	t *testing.T,
) {
	metrics := &analyticalMetricStub{}
	handler, err := NewAnalyticalMetricsHandler(
		metrics,
		&analyticalQueryStub{},
	)
	if err != nil {
		t.Fatalf("expected handler, got %v", err)
	}

	app := fiber.New()
	app.Get("/metric", handler.GetDataFreshness)
	result, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/metric?observed_at=2026-07-14T10:00:00Z&max_age_seconds=120",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", result.StatusCode)
	}
	if metrics.freshnessRequest.MaxAge != 120*time.Second ||
		metrics.freshnessRequest.Snapshot.Time !=
			time.Date(2026, time.July, 14, 10, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected freshness request: %#v", metrics.freshnessRequest)
	}
}

func TestAnalyticalQueryFailureReturnsServerError(
	t *testing.T,
) {
	handler, err := NewAnalyticalMetricsHandler(
		&analyticalMetricStub{},
		&analyticalQueryStub{err: errors.New("database unavailable")},
	)
	if err != nil {
		t.Fatalf("expected handler, got %v", err)
	}

	app := fiber.New()
	app.Get("/metric", handler.GetActiveAircraft)
	result, err := app.Test(
		httptest.NewRequest(fiber.MethodGet, "/metric", nil),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", result.StatusCode)
	}
}
