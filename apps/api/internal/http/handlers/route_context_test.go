package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/routecontext"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type routeContextServiceStub struct {
	item  routecontext.Context
	err   error
	calls int
	icao  string
}

func (stub *routeContextServiceStub) GetByICAO24(
	ctx context.Context,
	icao24 string,
) (routecontext.Context, error) {
	stub.calls++
	stub.icao = icao24

	return stub.item, stub.err
}

func TestRouteContextHandlerReturnsContext(t *testing.T) {
	service := &routeContextServiceStub{
		item: routecontext.Context{
			ICAO24:       "ABC123",
			TrajectoryID: "trajectory-one",
			Origin: &routecontext.AirportCandidate{
				Airport: airport.Airport{
					ICAOCode: "UBBB",
					IATACode: "GYD",
					Name:     "Heydar Aliyev International Airport",
				},
				DistanceKM: 4.5,
				Confidence: routecontext.Confidence{
					Score: 0.9,
					Level: routecontext.ConfidenceLevelHigh,
				},
			},
			Destination: &routecontext.AirportCandidate{
				Airport: airport.Airport{
					ICAOCode: "UGTB",
					IATACode: "TBS",
					Name:     "Tbilisi International Airport",
				},
				DistanceKM: 6.2,
				Confidence: routecontext.Confidence{
					Score: 0.85,
					Level: routecontext.ConfidenceLevelHigh,
				},
			},
			Confidence: routecontext.Confidence{
				Score: 0.85,
				Level: routecontext.ConfidenceLevelHigh,
			},
			GeneratedAt: time.Date(
				2026,
				time.July,
				14,
				10,
				0,
				0,
				0,
				time.UTC,
			),
		},
	}
	handler := NewRouteContextHandler(service)
	app := fiber.New()
	app.Get(
		"/api/v1/aircraft/:icao24/route-context",
		handler.GetByICAO24,
	)

	result, err := app.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/api/v1/aircraft/abc123/route-context",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			result.StatusCode,
		)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("expected readable body, got %v", err)
	}
	content := string(body)
	if !strings.Contains(content, `"icao24":"ABC123"`) ||
		!strings.Contains(content, `"icao_code":"UBBB"`) ||
		!strings.Contains(content, `"level":"high"`) {
		t.Fatalf("unexpected response body: %s", content)
	}
	if service.calls != 1 || service.icao != "abc123" {
		t.Fatalf(
			"unexpected service calls: calls=%d icao=%q",
			service.calls,
			service.icao,
		)
	}
}

func TestRouteContextHandlerRejectsInvalidICAO24(t *testing.T) {
	handler := NewRouteContextHandler(
		&routeContextServiceStub{
			err: routecontext.ErrInvalidICAO24,
		},
	)
	app := fiber.New()
	app.Get(
		"/api/v1/aircraft/:icao24/route-context",
		handler.GetByICAO24,
	)

	result, err := app.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/api/v1/aircraft/bad/route-context",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			result.StatusCode,
		)
	}
}

func TestRouteContextHandlerReturnsNotFound(t *testing.T) {
	handler := NewRouteContextHandler(
		&routeContextServiceStub{
			err: pgx.ErrNoRows,
		},
	)
	app := fiber.New()
	app.Get(
		"/api/v1/aircraft/:icao24/route-context",
		handler.GetByICAO24,
	)

	result, err := app.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/api/v1/aircraft/ABC123/route-context",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusNotFound {
		t.Fatalf(
			"expected status 404, got %d",
			result.StatusCode,
		)
	}
}

func TestRouteContextHandlerReturnsServerError(t *testing.T) {
	handler := NewRouteContextHandler(
		&routeContextServiceStub{
			err: errors.New("database failure"),
		},
	)
	app := fiber.New()
	app.Get(
		"/api/v1/aircraft/:icao24/route-context",
		handler.GetByICAO24,
	)

	result, err := app.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/api/v1/aircraft/ABC123/route-context",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("expected response, got %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf(
			"expected status 500, got %d",
			result.StatusCode,
		)
	}
}
