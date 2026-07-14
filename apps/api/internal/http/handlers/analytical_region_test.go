package handlers

import (
	"context"
	"math"
	"net/http/httptest"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricexecution"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricquery"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/gofiber/fiber/v2"
)

type regionalAnalyticalQueryStub struct {
	analyticalQueryStub
	regionalItems   []trajectory.FlightTrajectory
	regionalBounds  metricquery.Bounds
	regionalRequest metricquery.RecentRequest
	regionalCalls   int
	globalCalls     int
}

func (stub *regionalAnalyticalQueryStub) Recent(
	ctx context.Context,
	request metricquery.RecentRequest,
) ([]trajectory.FlightTrajectory, error) {
	stub.globalCalls++

	return stub.analyticalQueryStub.Recent(ctx, request)
}

func (stub *regionalAnalyticalQueryStub) RecentWithinBounds(
	ctx context.Context,
	request metricquery.RecentRequest,
	bounds metricquery.Bounds,
) ([]trajectory.FlightTrajectory, error) {
	stub.regionalRequest = request
	stub.regionalBounds = bounds
	stub.regionalCalls++

	return stub.regionalItems, stub.err
}

func TestAnalyticalActiveAircraftUsesSelectedRegionBounds(
	t *testing.T,
) {
	query := &regionalAnalyticalQueryStub{
		regionalItems: []trajectory.FlightTrajectory{
			{ID: "one", SourceName: "airplanes.live"},
			{ID: "two", SourceName: "airplanes.live"},
		},
	}
	metrics := &analyticalMetricStub{}
	handler, err := NewAnalyticalMetricsHandler(metrics, query)
	if err != nil {
		t.Fatalf("expected handler, got %v", err)
	}

	app := fiber.New()
	app.Get("/metric", handler.GetActiveAircraft)
	result, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/metric?region=turkey&window_minutes=30&limit=20",
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

	expectedBounds := metricquery.Bounds{
		MinLatitude:  35,
		MaxLatitude:  43,
		MinLongitude: 25,
		MaxLongitude: 45,
	}
	if query.regionalCalls != 1 ||
		query.regionalRequest.WindowMinutes != 30 ||
		query.regionalRequest.Limit != 20 ||
		query.regionalBounds != expectedBounds {
		t.Fatalf(
			"unexpected regional query: calls=%d request=%#v bounds=%#v",
			query.regionalCalls,
			query.regionalRequest,
			query.regionalBounds,
		)
	}
	if len(metrics.activeRequest.Trajectories) != 2 {
		t.Fatalf(
			"expected two regional trajectories, got %#v",
			metrics.activeRequest.Trajectories,
		)
	}
	if !containsPublicationLimitation(
		metrics.activeRequest.PublicationMetadata,
		"regional_bounding_box",
	) {
		t.Fatalf(
			"expected regional limitation, got %#v",
			metrics.activeRequest.PublicationMetadata.Limitations,
		)
	}
}

func TestAnalyticalTrafficDensityDerivesAreaFromRegion(
	t *testing.T,
) {
	query := &regionalAnalyticalQueryStub{
		regionalItems: []trajectory.FlightTrajectory{
			{ID: "one", SourceName: "airplanes.live"},
		},
	}
	metrics := &analyticalMetricStub{}
	handler, err := NewAnalyticalMetricsHandler(metrics, query)
	if err != nil {
		t.Fatalf("expected handler, got %v", err)
	}

	app := fiber.New()
	app.Get("/metric", handler.GetTrafficDensity)
	result, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/metric?region=caucasus",
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

	expectedBounds := metricquery.Bounds{
		MinLatitude:  38,
		MaxLatitude:  44,
		MinLongitude: 38,
		MaxLongitude: 51,
	}
	expectedArea, err := expectedBounds.AreaSquareKilometers()
	if err != nil {
		t.Fatalf("expected area, got %v", err)
	}

	if query.regionalBounds != expectedBounds {
		t.Fatalf(
			"expected bounds %#v, got %#v",
			expectedBounds,
			query.regionalBounds,
		)
	}
	if math.Abs(
		metrics.densityRequest.AreaSquareKilometers-expectedArea,
	) > 0.001 {
		t.Fatalf(
			"expected area %.3f, got %.3f",
			expectedArea,
			metrics.densityRequest.AreaSquareKilometers,
		)
	}
}

func TestAnalyticalMetricRejectsUnknownRegion(
	t *testing.T,
) {
	query := &regionalAnalyticalQueryStub{}
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
			"/metric?region=unknown",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status 404, got %d", result.StatusCode)
	}
	if query.regionalCalls != 0 {
		t.Fatal("expected regional query not to run")
	}
}

func TestWorldRegionUsesGlobalTrajectoryQuery(
	t *testing.T,
) {
	query := &regionalAnalyticalQueryStub{
		analyticalQueryStub: analyticalQueryStub{
			recentItems: []trajectory.FlightTrajectory{
				{ID: "global", SourceName: "airplanes.live"},
			},
		},
	}
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
			"/metric?region=world",
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
	if query.regionalCalls != 0 || query.globalCalls != 1 {
		t.Fatalf(
			"expected one global query and no regional query, got global=%d regional=%d",
			query.globalCalls,
			query.regionalCalls,
		)
	}
	if query.recent.Limit != 0 || query.recent.WindowMinutes != 0 {
		t.Fatalf("unexpected normalized request capture: %#v", query.recent)
	}
}

func TestRegionalQuerySupportFailureReturnsServerError(
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
	app.Get("/metric", handler.GetActiveAircraft)
	result, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/metric?region=turkey",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", result.StatusCode)
	}
}

func containsPublicationLimitation(
	metadata metricexecution.PublicationMetadata,
	code string,
) bool {
	for _, limitation := range metadata.Limitations {
		if limitation.Code == code {
			return true
		}
	}

	return false
}
