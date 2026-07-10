package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/gofiber/fiber/v2"
)

func TestFlightStateHandlerGetLatestByICAO24ReturnsNotFound(
	t *testing.T,
) {
	handler := NewFlightStateHandler(
		flightstate.NewService(
			&flightStateNotFoundRepository{},
		),
	)

	app := fiber.New()
	app.Get(
		"/api/v1/aircraft/:icao24/state/latest",
		handler.GetLatestByICAO24,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/aircraft/ABC123/state/latest",
		nil,
	)

	result, err := app.Test(
		request,
		-1,
	)
	if err != nil {
		t.Fatalf(
			"execute latest flight state request: %v",
			err,
		)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			fiber.StatusNotFound,
			result.StatusCode,
		)
	}

	body, err := io.ReadAll(
		result.Body,
	)
	if err != nil {
		t.Fatalf(
			"read latest flight state response: %v",
			err,
		)
	}

	if !strings.Contains(
		string(body),
		`"code":"FLIGHT_STATE_NOT_FOUND"`,
	) {
		t.Fatalf(
			"expected flight state not found code, got %s",
			body,
		)
	}
}

type flightStateNotFoundRepository struct{}

func (*flightStateNotFoundRepository) ListByFlightID(
	context.Context,
	string,
) ([]flightstate.FlightState, error) {
	return nil, nil
}

func (*flightStateNotFoundRepository) GetLatestByICAO24(
	context.Context,
	string,
) (flightstate.FlightState, error) {
	return flightstate.FlightState{},
		flightstate.ErrNotFound
}
