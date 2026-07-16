package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheruncertainty"
	"github.com/gofiber/fiber/v2"
)

type weatherContextReaderStub struct {
	result handlers.WeatherContextReadResult
	err    error

	request handlers.WeatherContextReadRequest
	calls   int
}

func (reader *weatherContextReaderStub) GetWeatherContext(
	_ context.Context,
	request handlers.WeatherContextReadRequest,
) (handlers.WeatherContextReadResult, error) {
	reader.calls++
	reader.request = request
	return reader.result.Clone(), reader.err
}

func TestWeatherContextRouteReturnsReadOnlyAggregate(t *testing.T) {
	t.Parallel()

	reader := &weatherContextReaderStub{
		result: validUnavailableWeatherContextResult(),
	}
	app := fiber.New()
	v1 := app.Group("/api/v1")
	if err := server.RegisterWeatherContextReadRoute(v1, reader); err != nil {
		t.Fatalf("register Weather Context route: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/trajectories/"+
			"73aa02ab-7061-4e9e-a238-d32710371ee3"+
			"/weather-context"+
			"?as_of_time=2026-07-16T12:00:00%2B04:00"+
			"&duration_seconds=180",
		nil,
	)
	httpResponse, err := app.Test(request)
	if err != nil {
		t.Fatalf("execute Weather Context request: %v", err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != fiber.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			httpResponse.StatusCode,
			fiber.StatusOK,
		)
	}
	if reader.calls != 1 ||
		reader.request.TrajectoryID != "73aa02ab-7061-4e9e-a238-d32710371ee3" ||
		!reader.request.AsOfTime.Equal(
			time.Date(2026, time.July, 16, 8, 0, 0, 0, time.UTC),
		) ||
		reader.request.RequestedDuration != 3*time.Minute {
		t.Fatalf("unexpected reader request: %#v", reader.request)
	}

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			Version string `json:"version"`
			Weather struct {
				Status string `json:"status"`
			} `json:"weather"`
			Trust struct {
				Decision string `json:"decision"`
			} `json:"trust"`
			Alignment struct {
				Status string `json:"status"`
			} `json:"alignment"`
			Encounter struct {
				Status string `json:"status"`
			} `json:"encounter"`
			Uncertainty struct {
				Status string `json:"status"`
			} `json:"uncertainty"`
		} `json:"data"`
	}
	if err := json.NewDecoder(httpResponse.Body).Decode(&payload); err != nil {
		t.Fatalf("decode Weather Context response: %v", err)
	}

	if !payload.Success ||
		payload.Data.Version != "weather-context-api-v1" ||
		payload.Data.Weather.Status != "unavailable" ||
		payload.Data.Trust.Decision != "blocked" ||
		payload.Data.Alignment.Status != "unavailable" ||
		payload.Data.Encounter.Status != "unavailable" ||
		payload.Data.Uncertainty.Status != "unavailable" {
		t.Fatalf("unexpected Weather Context payload: %#v", payload)
	}
}

func TestWeatherContextRouteRejectsInvalidQueries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		code string
	}{
		{
			name: "trajectory identifier",
			url: "/api/v1/trajectories/not-a-uuid/weather-context" +
				"?as_of_time=2026-07-16T12:00:00Z&duration_seconds=180",
			code: "INVALID_WEATHER_CONTEXT_TRAJECTORY_ID",
		},
		{
			name: "as-of time",
			url: "/api/v1/trajectories/73aa02ab-7061-4e9e-a238-d32710371ee3/" +
				"weather-context?as_of_time=invalid&duration_seconds=180",
			code: "INVALID_WEATHER_CONTEXT_AS_OF_TIME",
		},
		{
			name: "duration",
			url: "/api/v1/trajectories/73aa02ab-7061-4e9e-a238-d32710371ee3/" +
				"weather-context?as_of_time=2026-07-16T12:00:00Z&duration_seconds=0",
			code: "INVALID_WEATHER_CONTEXT_DURATION",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			reader := &weatherContextReaderStub{
				result: validUnavailableWeatherContextResult(),
			}
			app := weatherContextTestApp(t, reader)

			httpResponse, err := app.Test(
				httptest.NewRequest(http.MethodGet, test.url, nil),
			)
			if err != nil {
				t.Fatalf("execute invalid request: %v", err)
			}
			defer httpResponse.Body.Close()

			if httpResponse.StatusCode != fiber.StatusBadRequest ||
				reader.calls != 0 {
				t.Fatalf(
					"status = %d calls = %d",
					httpResponse.StatusCode,
					reader.calls,
				)
			}

			var payload struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.NewDecoder(httpResponse.Body).Decode(&payload); err != nil {
				t.Fatalf("decode invalid-request response: %v", err)
			}
			if payload.Error.Code != test.code {
				t.Fatalf(
					"error code = %q, want %q",
					payload.Error.Code,
					test.code,
				)
			}
		})
	}
}

func TestWeatherContextRouteMapsReaderErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{
			name:   "not found",
			err:    handlers.ErrWeatherContextNotFound,
			status: fiber.StatusNotFound,
			code:   "WEATHER_CONTEXT_NOT_FOUND",
		},
		{
			name:   "unavailable",
			err:    handlers.ErrWeatherContextServiceUnavailable,
			status: fiber.StatusServiceUnavailable,
			code:   "WEATHER_CONTEXT_SERVICE_UNAVAILABLE",
		},
		{
			name:   "timeout",
			err:    context.DeadlineExceeded,
			status: fiber.StatusGatewayTimeout,
			code:   "WEATHER_CONTEXT_TIMEOUT",
		},
		{
			name:   "internal",
			err:    errors.New("boom"),
			status: fiber.StatusInternalServerError,
			code:   "WEATHER_CONTEXT_LOAD_FAILED",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			reader := &weatherContextReaderStub{err: test.err}
			app := weatherContextTestApp(t, reader)
			httpResponse, err := app.Test(
				httptest.NewRequest(
					http.MethodGet,
					"/api/v1/trajectories/73aa02ab-7061-4e9e-a238-d32710371ee3/"+
						"weather-context?as_of_time=2026-07-16T12:00:00Z&duration_seconds=180",
					nil,
				),
			)
			if err != nil {
				t.Fatalf("execute reader-error request: %v", err)
			}
			defer httpResponse.Body.Close()

			if httpResponse.StatusCode != test.status {
				t.Fatalf(
					"status = %d, want %d",
					httpResponse.StatusCode,
					test.status,
				)
			}

			var payload struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.NewDecoder(httpResponse.Body).Decode(&payload); err != nil {
				t.Fatalf("decode reader-error response: %v", err)
			}
			if payload.Error.Code != test.code {
				t.Fatalf(
					"error code = %q, want %q",
					payload.Error.Code,
					test.code,
				)
			}
		})
	}
}

func TestWeatherContextRouteRejectsInvalidAggregate(t *testing.T) {
	t.Parallel()

	reader := &weatherContextReaderStub{
		result: handlers.WeatherContextReadResult{},
	}
	app := weatherContextTestApp(t, reader)

	httpResponse, err := app.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/api/v1/trajectories/73aa02ab-7061-4e9e-a238-d32710371ee3/"+
				"weather-context?as_of_time=2026-07-16T12:00:00Z&duration_seconds=180",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("execute invalid-contract request: %v", err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want %d",
			httpResponse.StatusCode,
			fiber.StatusInternalServerError,
		)
	}

	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(httpResponse.Body).Decode(&payload); err != nil {
		t.Fatalf("decode invalid-contract response: %v", err)
	}
	if payload.Error.Code != "WEATHER_CONTEXT_CONTRACT_INVALID" {
		t.Fatalf("unexpected error code: %q", payload.Error.Code)
	}
}

func TestRegisterWeatherContextReadRouteRejectsNilReader(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	if err := server.RegisterWeatherContextReadRoute(
		app.Group("/api/v1"),
		nil,
	); err == nil {
		t.Fatal("expected nil-reader route registration error")
	}
}

func weatherContextTestApp(
	t *testing.T,
	reader handlers.WeatherContextReader,
) *fiber.App {
	t.Helper()

	app := fiber.New()
	if err := server.RegisterWeatherContextReadRoute(
		app.Group("/api/v1"),
		reader,
	); err != nil {
		t.Fatalf("register Weather Context route: %v", err)
	}
	return app
}

func validUnavailableWeatherContextResult() handlers.WeatherContextReadResult {
	asOfTime := time.Date(2026, time.July, 16, 12, 0, 0, 0, time.UTC)
	trajectoryID := "73aa02ab-7061-4e9e-a238-d32710371ee3"
	generatedAt := asOfTime.Add(5 * time.Minute)

	weather := weathercontract.Result{
		SchemaVersion: weathercontract.SchemaVersionV1,
		Status:        weathercontract.ResultStatusUnavailable,
		TrajectoryID:  trajectoryID,
		AsOfTime:      asOfTime,
		Confidence: weathercontract.Confidence{
			Score: 0,
			Level: weathercontract.ConfidenceLevelNone,
		},
		Limitations: []weathercontract.Limitation{
			{
				Code:    "weather_unavailable",
				Message: "Weather evidence is unavailable.",
				Scope:   "weather_context",
			},
		},
		Explanations: []weathercontract.Explanation{
			{
				Code:    "context_only",
				Message: "Weather remains contextual evidence only.",
			},
		},
		ScopeGuard: weathercontract.ScopeGuardContextOnly,
		Provenance: weathercontract.Provenance{
			InputFingerprint: weatherContextTestFingerprint("weather"),
		},
		GeneratedAt: asOfTime.Add(time.Minute),
	}

	trust := weathertrust.Result{
		Version:  weathertrust.Version,
		Decision: weathertrust.DecisionBlocked,
		Usable:   false,
		AsOfTime: asOfTime,
		Score:    0,
		Components: []weathertrust.Component{
			{Name: weathertrust.ComponentContractConfidence, Score: 0, Weight: 0.35},
			{Name: weathertrust.ComponentTemporalFreshness, Score: 0, Weight: 0.30},
			{Name: weathertrust.ComponentFeatureCompleteness, Score: 0, Weight: 0.20},
			{Name: weathertrust.ComponentVerticalApplicability, Score: 0, Weight: 0.15},
		},
		Limitations: []weathertrust.Notice{
			{Code: "weather_blocked", Message: "Weather is unavailable."},
		},
		Explanations: []weathertrust.Notice{
			{Code: "context_only", Message: "Weather is contextual."},
		},
		InputFingerprint: weatherContextTestFingerprint("trust"),
	}

	alignment := weatheralignment.Result{
		Version:       weatheralignment.Version,
		Status:        weatheralignment.StatusUnavailable,
		TrajectoryID:  trajectoryID,
		AsOfTime:      asOfTime,
		TrustDecision: weathertrust.DecisionBlocked,
		TrustScore:    0,
		Limitations: []weatheralignment.Notice{
			{Code: "alignment_unavailable", Message: "Alignment is unavailable."},
		},
		Explanations: []weatheralignment.Notice{
			{Code: "context_only", Message: "Alignment is contextual."},
		},
		InputFingerprint: weatherContextTestFingerprint("alignment"),
		GeneratedAt:      asOfTime.Add(2 * time.Minute),
	}

	encounter := weatherencounter.Result{
		Version:                weatherencounter.Version,
		Status:                 weatherencounter.StatusUnavailable,
		TrajectoryID:           trajectoryID,
		AsOfTime:               asOfTime,
		AlignmentStatus:        weatheralignment.StatusUnavailable,
		AlignmentCoverageRatio: 0,
		Limitations: []weatherencounter.Notice{
			{Code: "encounter_unavailable", Message: "Encounter is unavailable."},
		},
		Explanations: []weatherencounter.Notice{
			{Code: "context_only", Message: "Encounter is contextual."},
		},
		InputFingerprint: weatherContextTestFingerprint("encounter"),
		GeneratedAt:      asOfTime.Add(3 * time.Minute),
	}

	projection := projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        projectioncontract.ResultStatusUnavailable,
		TrajectoryID:  trajectoryID,
		Method: projectioncontract.Method{
			Name:          "weather-context-projection",
			Version:       "v1",
			DecisionClass: projectioncontract.DecisionClassProjectDerived,
		},
		Horizon: projectioncontract.Horizon{
			AsOfTime: asOfTime,
			EndTime:  asOfTime.Add(3 * time.Minute),
			Step:     time.Minute,
		},
		Confidence: projectioncontract.Confidence{
			Score: 0,
			Level: projectioncontract.ConfidenceLevelNone,
		},
		Limitations: []projectioncontract.Limitation{
			{
				Code:    "projection_unavailable",
				Message: "Projection is unavailable.",
				Scope:   "projection",
			},
		},
		ScopeGuard:  projectioncontract.ScopeGuardResearchOnly,
		GeneratedAt: asOfTime.Add(4 * time.Minute),
	}

	uncertainty := weatheruncertainty.Result{
		Version:           weatheruncertainty.Version,
		Status:            weatheruncertainty.StatusUnavailable,
		TrajectoryID:      trajectoryID,
		AsOfTime:          asOfTime,
		WeatherMultiplier: 1,
		Components: []weatheruncertainty.Component{
			{Name: weatheruncertainty.ComponentWindSpeed, Score: 0, Weight: 0.30},
			{Name: weatheruncertainty.ComponentWindGust, Score: 0, Weight: 0.20},
			{Name: weatheruncertainty.ComponentPrecipitation, Score: 0, Weight: 0.15},
			{Name: weatheruncertainty.ComponentCloudCover, Score: 0, Weight: 0.10},
			{Name: weatheruncertainty.ComponentEvidenceQuality, Score: 0, Weight: 0.25},
		},
		AdjustedProjection: projection,
		Limitations: []weatheruncertainty.Notice{
			{Code: "projection_unavailable", Message: "Projection is unavailable."},
		},
		Explanations: []weatheruncertainty.Notice{
			{Code: "context_only", Message: "Weather is contextual."},
		},
		InputFingerprint: weatherContextTestFingerprint("uncertainty"),
		GeneratedAt:      asOfTime.Add(4 * time.Minute),
	}

	return handlers.WeatherContextReadResult{
		Version:          handlers.WeatherContextReadResultVersion,
		Weather:          weather,
		Trust:            trust,
		Alignment:        alignment,
		Encounter:        encounter,
		Uncertainty:      uncertainty,
		InputFingerprint: weatherContextTestFingerprint("aggregate"),
		GeneratedAt:      generatedAt,
	}
}

func weatherContextTestFingerprint(seed string) string {
	values := map[string]string{
		"weather":     "0123456789abcdef",
		"trust":       "abcdef0123456789",
		"alignment":   "0011223344556677",
		"encounter":   "7766554433221100",
		"uncertainty": "fedcba0987654321",
		"aggregate":   "a1b2c3d4e5f60718",
	}
	value := values[seed]
	return "sha256:" + value + value + value + value
}
