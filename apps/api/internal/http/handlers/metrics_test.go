package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

type metricsHandlerRepositoryStub struct {
	summary metrics.ActiveAircraftObservationSummary
}

func (
	r *metricsHandlerRepositoryStub,
) CountActiveAircraft(
	ctx context.Context,
	query metrics.ActiveAircraftQuery,
) (metrics.ActiveAircraftObservationSummary, error) {
	return r.summary,
		nil
}

func TestMetricsHandlerReturnsActiveAircraftMetric(
	t *testing.T,
) {
	now := time.Now().UTC()
	repository := &metricsHandlerRepositoryStub{
		summary: metrics.ActiveAircraftObservationSummary{
			Count:            3,
			FirstObservedAt:  now.Add(-5 * time.Minute),
			LatestObservedAt: now.Add(-1 * time.Minute),
			SourceNames: []string{
				"airplanes.live",
			},
			HasObservations: true,
		},
	}

	app := fiber.New()
	service := metrics.MustNewService(
		repository,
		region.NewService(),
	)
	handler := NewMetricsHandler(
		service,
	)
	app.Get(
		"/api/v1/metrics/active-aircraft",
		handler.GetActiveAircraft,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/metrics/active-aircraft?region=caucasus&window_minutes=15",
		nil,
	)
	responseMessage, err := app.Test(
		request,
	)
	if err != nil {
		t.Fatalf(
			"execute active aircraft metric request: %v",
			err,
		)
	}
	defer responseMessage.Body.Close()

	if responseMessage.StatusCode != fiber.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			responseMessage.StatusCode,
		)
	}

	var body response.SuccessResponse[dto.ActiveAircraftMetricResponse]
	if err := json.NewDecoder(
		responseMessage.Body,
	).Decode(
		&body,
	); err != nil {
		t.Fatalf(
			"decode active aircraft metric response: %v",
			err,
		)
	}

	if !body.Success {
		t.Fatal(
			"expected success response",
		)
	}

	if body.Data.Metric != "active_aircraft" {
		t.Fatalf(
			"expected active_aircraft metric, got %s",
			body.Data.Metric,
		)
	}

	if body.Data.Value != 3 {
		t.Fatalf(
			"expected metric value 3, got %d",
			body.Data.Value,
		)
	}

	if body.Data.Scope.Code != "caucasus" {
		t.Fatalf(
			"expected caucasus scope, got %s",
			body.Data.Scope.Code,
		)
	}

	if body.Data.Confidence.Level == "" {
		t.Fatal(
			"expected confidence level",
		)
	}

	if len(body.Data.Limitations) == 0 {
		t.Fatal(
			"expected public limitations",
		)
	}
}

func TestMetricsHandlerRejectsInvalidWindowMinutes(
	t *testing.T,
) {
	app := fiber.New()
	service := metrics.MustNewService(
		&metricsHandlerRepositoryStub{},
		region.NewService(),
	)
	handler := NewMetricsHandler(
		service,
	)
	app.Get(
		"/api/v1/metrics/active-aircraft",
		handler.GetActiveAircraft,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/metrics/active-aircraft?window_minutes=abc",
		nil,
	)
	responseMessage, err := app.Test(
		request,
	)
	if err != nil {
		t.Fatalf(
			"execute invalid active aircraft metric request: %v",
			err,
		)
	}
	defer responseMessage.Body.Close()

	if responseMessage.StatusCode != fiber.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			responseMessage.StatusCode,
		)
	}

	var body response.ErrorResponse
	if err := json.NewDecoder(
		responseMessage.Body,
	).Decode(
		&body,
	); err != nil {
		t.Fatalf(
			"decode invalid active aircraft metric response: %v",
			err,
		)
	}

	if body.Error.Code != "INVALID_WINDOW_MINUTES" {
		t.Fatalf(
			"expected invalid window error code, got %s",
			body.Error.Code,
		)
	}
}

func TestMetricsHandlerReturnsRegionNotFound(
	t *testing.T,
) {
	app := fiber.New()
	service := metrics.MustNewService(
		&metricsHandlerRepositoryStub{},
		region.NewService(),
	)
	handler := NewMetricsHandler(
		service,
	)
	app.Get(
		"/api/v1/metrics/active-aircraft",
		handler.GetActiveAircraft,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/metrics/active-aircraft?region=unknown-region",
		nil,
	)
	responseMessage, err := app.Test(
		request,
	)
	if err != nil {
		t.Fatalf(
			"execute unknown region metric request: %v",
			err,
		)
	}
	defer responseMessage.Body.Close()

	if responseMessage.StatusCode != fiber.StatusNotFound {
		t.Fatalf(
			"expected status 404, got %d",
			responseMessage.StatusCode,
		)
	}
}
