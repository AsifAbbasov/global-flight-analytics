package server

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRegisterWeatherRouteRejectsNonPositiveOpenMeteoTimeout(
	t *testing.T,
) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "zero timeout",
			timeout: 0,
		},
		{
			name:    "negative timeout",
			timeout: -1 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				app := fiber.New()

				v1 := app.Group(
					"/api/v1",
				)

				err := registerWeatherRoute(
					v1,
					nil,
					test.timeout,
				)

				if err == nil {
					t.Fatal(
						"expected weather route registration error, got nil",
					)
				}

				expectedError := "open-meteo timeout must be greater than zero"

				if err.Error() != expectedError {
					t.Fatalf(
						"expected error %q, got %q",
						expectedError,
						err.Error(),
					)
				}
			},
		)
	}
}

func TestRegisterWeatherRouteAcceptsPositiveOpenMeteoTimeout(
	t *testing.T,
) {
	app := fiber.New()

	v1 := app.Group(
		"/api/v1",
	)

	err := registerWeatherRoute(
		v1,
		nil,
		5*time.Second,
	)
	if err != nil {
		t.Fatalf(
			"expected weather route registration to succeed, got %v",
			err,
		)
	}

	weatherRouteFound := false

	for _, route := range app.GetRoutes() {
		if route.Method == fiber.MethodGet &&
			route.Path == "/api/v1/weather/current" {
			weatherRouteFound = true
			break
		}
	}

	if !weatherRouteFound {
		t.Fatal(
			"expected current weather route to be registered",
		)
	}
}

func TestNewWithoutDatabaseDoesNotRequireOpenMeteoTimeout(
	t *testing.T,
) {
	log := newDiscardLogger()

	app, err := New(
		Config{
			Logger: log,
		},
	)
	if err != nil {
		t.Fatalf(
			"expected server without database to initialize, got error: %v",
			err,
		)
	}

	if app == nil {
		t.Fatal(
			"expected initialized application",
		)
	}
}

func TestNewPropagatesWeatherRouteInitializationError(
	t *testing.T,
) {
	log := newDiscardLogger()

	app, err := New(
		Config{
			DatabasePool: &pgxpool.Pool{},
			Logger:       log,
		},
	)

	if err == nil {
		t.Fatal(
			"expected server initialization error, got nil",
		)
	}

	if app != nil {
		t.Fatal(
			"expected nil application on initialization error",
		)
	}

	expectedError := "register weather route: open-meteo timeout must be greater than zero"

	if err.Error() != expectedError {
		t.Fatalf(
			"expected error %q, got %q",
			expectedError,
			err.Error(),
		)
	}
}

func TestNewWithDatabaseAcceptsPositiveOpenMeteoTimeout(
	t *testing.T,
) {
	log := newDiscardLogger()

	app, err := New(
		Config{
			DatabasePool:     &pgxpool.Pool{},
			Logger:           log,
			OpenMeteoTimeout: 5 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf(
			"expected server initialization to succeed, got error: %v",
			err,
		)
	}

	if app == nil {
		t.Fatal(
			"expected initialized application",
		)
	}

	weatherRouteFound := false

	for _, route := range app.GetRoutes() {
		if route.Method == fiber.MethodGet &&
			route.Path == "/api/v1/weather/current" {
			weatherRouteFound = true
			break
		}
	}

	if !weatherRouteFound {
		t.Fatal(
			"expected current weather route to be registered",
		)
	}
}

func newDiscardLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(
			io.Discard,
			nil,
		),
	)
}
