package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualityintegration"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricexecution"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricquery"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
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

type analyticalAirportStub struct {
	item airport.Airport
	err  error
	icao string
}

func (
	stub *analyticalAirportStub,
) GetByICAO(
	ctx context.Context,
	icao string,
) (airport.Airport, error) {
	stub.icao = icao
	return stub.item, stub.err
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
		metricexecution.MetricIDTrafficDensity,
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
		metricexecution.MetricIDCoverageScore,
		0.75,
	), stub.err
}

func (stub *analyticalMetricStub) DataFreshness(
	ctx context.Context,
	request metricexecution.DataFreshnessRequest,
) (metricexecution.Execution[float64], error) {
	stub.freshnessRequest = request
	return successfulExecution(
		metricexecution.MetricIDDataFreshness,
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

func TestAnalyticalAirportActivityUsesServerAirportAndRegionalQuery(
	t *testing.T,
) {
	query := &regionalAnalyticalQueryStub{
		regionalItems: []trajectory.FlightTrajectory{
			{ID: "movement", SourceName: "airplanes.live"},
		},
	}
	metrics := &analyticalMetricStub{}
	airports := &analyticalAirportStub{
		item: airport.Airport{
			ICAOCode:  "UBBB",
			Latitude:  40.4675,
			Longitude: 50.0467,
		},
	}

	handler, err := NewAnalyticalMetricsHandlerWithAirportService(
		metrics,
		query,
		airports,
	)
	if err != nil {
		t.Fatalf("expected handler, got %v", err)
	}

	app := fiber.New()
	app.Get("/metric", handler.GetAirportActivity)
	result, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/metric?airport_icao=ubbb&radius_kilometers=20&window_minutes=30&limit=50",
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
	if airports.icao != "ubbb" {
		t.Fatalf("unexpected airport lookup %q", airports.icao)
	}
	if query.regionalCalls != 1 ||
		query.regionalRequest.WindowMinutes != 30 ||
		query.regionalRequest.Limit != 50 {
		t.Fatalf(
			"unexpected airport regional query: %#v",
			query,
		)
	}
	if metrics.airportRequest.Airport.ICAOCode != "UBBB" ||
		metrics.airportRequest.RadiusKilometers != 20 ||
		len(metrics.airportRequest.Trajectories) != 1 {
		t.Fatalf(
			"unexpected airport request: %#v",
			metrics.airportRequest,
		)
	}
}

func TestAnalyticalCoverageScoreUsesServerOwnedSnapshot(
	t *testing.T,
) {
	now := time.Now().UTC()
	query := &analyticalQueryStub{
		recentItems: []trajectory.FlightTrajectory{
			{
				SourceName: "airplanes.live",
				Points: []trajectory.TrackPoint4D{
					{
						ObservedAt: now.
							Add(-20 * time.Second),
					},
					{
						ObservedAt: now.
							Add(-10 * time.Second),
					},
				},
			},
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
	app.Get("/metric", handler.GetCoverageScore)
	result, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/metric?window_minutes=1",
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

	if query.recent.WindowMinutes != 1 ||
		query.recent.Limit !=
			metricquery.MaximumResultLimit {
		t.Fatalf(
			"unexpected production query: %#v",
			query.recent,
		)
	}

	if metrics.coverageRequest.Snapshot.
		ObservedSamples != 2 ||
		metrics.coverageRequest.Snapshot.
			ExpectedSamples != 6 {
		t.Fatalf(
			"unexpected server-owned coverage snapshot: %#v",
			metrics.coverageRequest,
		)
	}

	if !containsAnalyticalSource(
		metrics.coverageRequest.Sources,
		serverTrajectoryQuerySource,
	) {
		t.Fatalf(
			"expected server query provenance, got %#v",
			metrics.coverageRequest.Sources,
		)
	}
}

func TestAnalyticalDataFreshnessUsesServerOwnedObservation(
	t *testing.T,
) {
	observedAt := time.Now().UTC().
		Add(-30 * time.Second)
	query := &analyticalQueryStub{
		recentItems: []trajectory.FlightTrajectory{
			{
				SourceName: "airplanes.live",
				EndTime:    observedAt,
			},
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
	app.Get("/metric", handler.GetDataFreshness)
	result, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/metric?window_minutes=1",
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

	if metrics.freshnessRequest.MaxAge !=
		dataqualityintegration.DefaultStaleAfter ||
		!metrics.freshnessRequest.Snapshot.Time.Equal(
			observedAt,
		) {
		t.Fatalf(
			"unexpected server-owned freshness request: %#v",
			metrics.freshnessRequest,
		)
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

func TestTrajectoryPublicationMetadataBuildsStrictProvenance(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		24,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	metadata := trajectoryPublicationMetadata(
		[]trajectory.FlightTrajectory{
			{
				SourceName: "airplanes.live",
				StartTime:  now.Add(-time.Minute),
				EndTime:    now,
			},
			{
				SourceName: " ",
				StartTime:  now.Add(-time.Minute),
				EndTime:    now,
			},
		},
		0,
	)

	if len(metadata.Sources) != 1 {
		t.Fatalf(
			"expected one attributed source, got %#v",
			metadata.Sources,
		)
	}

	source := metadata.Sources[0]
	if source.Name != "airplanes.live" ||
		source.RetrievedAt.IsZero() {
		t.Fatalf(
			"unexpected source provenance: %#v",
			source,
		)
	}

	if containsAnalyticalSource(
		metadata.Sources,
		"unknown",
	) {
		t.Fatal(
			"placeholder source must not be published",
		)
	}

	if !containsPublicationLimitation(
		metadata,
		"unattributed_source_observations",
	) {
		t.Fatalf(
			"expected unattributed source limitation, got %#v",
			metadata.Limitations,
		)
	}
}

func TestProductionQualityEndpointsRejectCallerSnapshotParameters(
	t *testing.T,
) {
	testCases := []struct {
		name string
		path string
	}{
		{
			name: "coverage observed samples",
			path: "/coverage?observed_samples=1",
		},
		{
			name: "coverage expected samples",
			path: "/coverage?expected_samples=1",
		},
		{
			name: "coverage client limit",
			path: "/coverage?limit=10",
		},
		{
			name: "freshness observed timestamp",
			path: "/freshness?observed_at=2026-07-24T12:00:00Z",
		},
		{
			name: "freshness maximum age",
			path: "/freshness?max_age_seconds=120",
		},
		{
			name: "freshness client limit",
			path: "/freshness?limit=10",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				query := &analyticalQueryStub{}
				metrics := &analyticalMetricStub{}
				handler, err :=
					NewAnalyticalMetricsHandler(
						metrics,
						query,
					)
				if err != nil {
					t.Fatalf(
						"expected handler, got %v",
						err,
					)
				}

				app := fiber.New()
				app.Get(
					"/coverage",
					handler.GetCoverageScore,
				)
				app.Get(
					"/freshness",
					handler.GetDataFreshness,
				)

				result, requestErr := app.Test(
					httptest.NewRequest(
						fiber.MethodGet,
						testCase.path,
						nil,
					),
				)
				if requestErr != nil {
					t.Fatalf(
						"expected response for %s, got %v",
						testCase.path,
						requestErr,
					)
				}
				defer result.Body.Close()

				if result.StatusCode !=
					fiber.StatusBadRequest {
					t.Fatalf(
						"expected status 400 for %s, got %d",
						testCase.path,
						result.StatusCode,
					)
				}

				if query.recent.WindowMinutes != 0 ||
					query.recent.Limit != 0 {
					t.Fatalf(
						"query service received rejected request %s: %#v",
						testCase.path,
						query.recent,
					)
				}
			},
		)
	}
}

func containsAnalyticalSource(
	sources []analyticalresult.Source,
	name string,
) bool {
	for _, source := range sources {
		if source.Name == name {
			return true
		}
	}

	return false
}
