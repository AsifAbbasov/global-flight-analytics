package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/gofiber/fiber/v2"
)

type projectionIntelligenceReaderStub struct {
	result projectionproduction.Result
	err    error

	request ProjectionIntelligenceReadRequest
	calls   int
}

func (
	reader *projectionIntelligenceReaderStub,
) GetProjectionIntelligence(
	_ context.Context,
	request ProjectionIntelligenceReadRequest,
) (projectionproduction.Result, error) {
	reader.calls++
	reader.request = request

	return reader.result.Clone(),
		reader.err
}

func TestProjectionIntelligenceHandlerReturnsProductionResult(
	t *testing.T,
) {
	reader := &projectionIntelligenceReaderStub{
		result: validProjectionHTTPResult(),
	}
	app := newProjectionIntelligenceTestApp(
		reader,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/trajectories/"+
			"73aa02ab-7061-4e9e-a238-d32710371ee3"+
			"/projection-intelligence"+
			"?as_of_time=2026-07-16T12:00:00%2B04:00"+
			"&duration_seconds=180",
		nil,
	)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf(
			"execute Projection Intelligence request: %v",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode !=
		fiber.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			response.StatusCode,
			fiber.StatusOK,
		)
	}
	if reader.calls != 1 {
		t.Fatalf(
			"reader calls = %d, want 1",
			reader.calls,
		)
	}
	if reader.request.TrajectoryID !=
		"73aa02ab-7061-4e9e-a238-d32710371ee3" ||
		!reader.request.AsOfTime.Equal(
			time.Date(
				2026,
				time.July,
				16,
				8,
				0,
				0,
				0,
				time.UTC,
			),
		) ||
		reader.request.RequestedDuration !=
			3*time.Minute {
		t.Fatalf(
			"unexpected service request: %#v",
			reader.request,
		)
	}

	var payload map[string]any
	if err := json.NewDecoder(
		response.Body,
	).Decode(&payload); err != nil {
		t.Fatalf(
			"decode Projection Intelligence response: %v",
			err,
		)
	}
	if payload["success"] != true {
		t.Fatalf(
			"unexpected response payload: %#v",
			payload,
		)
	}

	data, ok := payload["data"].(map[string]any)
	if !ok ||
		data["strategy"] !=
			"kinematic_baseline" ||
		data["arrival_status"] !=
			"skipped" {
		t.Fatalf(
			"unexpected Projection Intelligence data: %#v",
			payload,
		)
	}
}

func TestProjectionIntelligenceHandlerRejectsInvalidQueries(
	t *testing.T,
) {
	tests := []struct {
		name string
		url  string
		code string
	}{
		{
			name: "trajectory identifier",
			url: "/api/v1/trajectories/not-a-uuid/" +
				"projection-intelligence" +
				"?as_of_time=2026-07-16T12:00:00Z" +
				"&duration_seconds=180",
			code: "INVALID_PROJECTION_TRAJECTORY_ID",
		},
		{
			name: "as-of time",
			url: "/api/v1/trajectories/" +
				"73aa02ab-7061-4e9e-a238-d32710371ee3/" +
				"projection-intelligence" +
				"?as_of_time=invalid" +
				"&duration_seconds=180",
			code: "INVALID_PROJECTION_AS_OF_TIME",
		},
		{
			name: "duration",
			url: "/api/v1/trajectories/" +
				"73aa02ab-7061-4e9e-a238-d32710371ee3/" +
				"projection-intelligence" +
				"?as_of_time=2026-07-16T12:00:00Z" +
				"&duration_seconds=0",
			code: "INVALID_PROJECTION_DURATION",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				reader :=
					&projectionIntelligenceReaderStub{
						result: validProjectionHTTPResult(),
					}
				app :=
					newProjectionIntelligenceTestApp(
						reader,
					)

				request :=
					httptest.NewRequest(
						http.MethodGet,
						test.url,
						nil,
					)
				response, err :=
					app.Test(request)
				if err != nil {
					t.Fatalf(
						"execute request: %v",
						err,
					)
				}
				defer response.Body.Close()

				if response.StatusCode !=
					fiber.StatusBadRequest {
					t.Fatalf(
						"status = %d, want %d",
						response.StatusCode,
						fiber.StatusBadRequest,
					)
				}
				if reader.calls != 0 {
					t.Fatalf(
						"reader calls = %d, want 0",
						reader.calls,
					)
				}

				var payload struct {
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if err := json.NewDecoder(
					response.Body,
				).Decode(&payload); err != nil {
					t.Fatalf(
						"decode error response: %v",
						err,
					)
				}
				if payload.Error.Code !=
					test.code {
					t.Fatalf(
						"error code = %q, want %q",
						payload.Error.Code,
						test.code,
					)
				}
			},
		)
	}
}

func TestProjectionIntelligenceHandlerMapsServiceErrors(
	t *testing.T,
) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{
			name:   "not found",
			err:    ErrProjectionIntelligenceNotFound,
			status: fiber.StatusNotFound,
			code:   "PROJECTION_INTELLIGENCE_NOT_FOUND",
		},
		{
			name:   "unavailable",
			err:    ErrProjectionIntelligenceServiceUnavailable,
			status: fiber.StatusServiceUnavailable,
			code:   "PROJECTION_INTELLIGENCE_SERVICE_UNAVAILABLE",
		},
		{
			name:   "timeout",
			err:    context.DeadlineExceeded,
			status: fiber.StatusGatewayTimeout,
			code:   "PROJECTION_INTELLIGENCE_TIMEOUT",
		},
		{
			name:   "internal",
			err:    errors.New("boom"),
			status: fiber.StatusInternalServerError,
			code:   "PROJECTION_INTELLIGENCE_LOAD_FAILED",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				reader :=
					&projectionIntelligenceReaderStub{
						err: test.err,
					}
				app :=
					newProjectionIntelligenceTestApp(
						reader,
					)

				request :=
					httptest.NewRequest(
						http.MethodGet,
						"/api/v1/trajectories/"+
							"73aa02ab-7061-4e9e-a238-d32710371ee3"+
							"/projection-intelligence"+
							"?as_of_time=2026-07-16T12:00:00Z"+
							"&duration_seconds=180",
						nil,
					)
				response, err :=
					app.Test(request)
				if err != nil {
					t.Fatalf(
						"execute request: %v",
						err,
					)
				}
				defer response.Body.Close()

				if response.StatusCode !=
					test.status {
					t.Fatalf(
						"status = %d, want %d",
						response.StatusCode,
						test.status,
					)
				}

				var payload struct {
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if err := json.NewDecoder(
					response.Body,
				).Decode(&payload); err != nil {
					t.Fatalf(
						"decode error response: %v",
						err,
					)
				}
				if payload.Error.Code !=
					test.code {
					t.Fatalf(
						"error code = %q, want %q",
						payload.Error.Code,
						test.code,
					)
				}
			},
		)
	}
}

func TestProjectionIntelligenceHandlerRejectsInvalidServiceContract(
	t *testing.T,
) {
	reader := &projectionIntelligenceReaderStub{
		result: projectionproduction.Result{},
	}
	app := newProjectionIntelligenceTestApp(
		reader,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/trajectories/"+
			"73aa02ab-7061-4e9e-a238-d32710371ee3"+
			"/projection-intelligence"+
			"?as_of_time=2026-07-16T12:00:00Z"+
			"&duration_seconds=180",
		nil,
	)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf(
			"execute request: %v",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode !=
		fiber.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want %d",
			response.StatusCode,
			fiber.StatusInternalServerError,
		)
	}

	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(
		response.Body,
	).Decode(&payload); err != nil {
		t.Fatalf(
			"decode error response: %v",
			err,
		)
	}
	if payload.Error.Code !=
		"PROJECTION_INTELLIGENCE_CONTRACT_INVALID" {
		t.Fatalf(
			"unexpected error code: %q",
			payload.Error.Code,
		)
	}
}

func newProjectionIntelligenceTestApp(
	reader ProjectionIntelligenceReader,
) *fiber.App {
	app := fiber.New()
	handler :=
		NewProjectionIntelligenceHandler(
			reader,
		)
	app.Get(
		"/api/v1/trajectories/:id/projection-intelligence",
		handler.GetByTrajectoryID,
	)

	return app
}

func validProjectionHTTPResult() projectionproduction.Result {
	asOfTime := time.Date(
		2026,
		time.July,
		16,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	generatedAt := asOfTime.Add(time.Second)

	return projectionproduction.Result{
		Version: projectionproduction.Version,
		Strategy: projectionproduction.
			StrategyKinematic,
		FallbackReason: "historical_neighbors_unavailable",
		ArrivalStatus: projectionproduction.
			ArrivalStatusSkipped,
		Projection: projectioncontract.Result{
			SchemaVersion: projectioncontract.
				SchemaVersionV1,
			Status: projectioncontract.
				ResultStatusUnavailable,
			TrajectoryID: "73aa02ab-7061-4e9e-a238-d32710371ee3",
			Method: projectioncontract.Method{
				Name:    projectionbaseline.MethodName,
				Version: projectionbaseline.Version,
				DecisionClass: projectioncontract.
					DecisionClassPhysicsDerived,
			},
			Horizon: projectioncontract.Horizon{
				AsOfTime: asOfTime,
				EndTime:  asOfTime.Add(3 * time.Minute),
				Step:     time.Minute,
			},
			Points: []projectioncontract.ProjectionPoint{},
			Confidence: projectioncontract.Confidence{
				Score: 0,
				Level: projectioncontract.
					ConfidenceLevelNone,
				Reasons: []projectioncontract.
					ConfidenceReason{},
			},
			Limitations: []projectioncontract.Limitation{
				{
					Code:    "historical_neighbors_unavailable",
					Message: "No historical neighbors were available.",
					Scope:   "result",
				},
			},
			Explanations: []projectioncontract.Explanation{},
			ScopeGuard: projectioncontract.
				ScopeGuardResearchOnly,
			Provenance: projectioncontract.Provenance{
				Inputs: []projectioncontract.InputReference{},
			},
			GeneratedAt: generatedAt,
		},
		Notices: []projectionproduction.Notice{
			{
				Code:    "historical_neighbors_unavailable",
				Message: "Kinematic baseline was selected.",
			},
		},
		InputFingerprint: "sha256:" +
			strings.Repeat("a", 64),
		GeneratedAt: generatedAt,
	}
}
