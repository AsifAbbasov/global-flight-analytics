package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestProjectionIntelligenceHandlerAllowsDefaultDuration(t *testing.T) {
	reader := &projectionIntelligenceReaderStub{
		result: validProjectionHTTPResult(),
	}
	app := newProjectionIntelligenceTestApp(reader)
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/trajectories/"+
			"73aa02ab-7061-4e9e-a238-d32710371ee3"+
			"/projection-intelligence"+
			"?as_of_time=2026-07-16T12:00:00Z",
		nil,
	)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}
	if reader.calls != 1 {
		t.Fatalf("reader calls = %d, want 1", reader.calls)
	}
	if reader.request.RequestedDuration != 0 {
		t.Fatalf(
			"requested duration = %s, want zero default sentinel",
			reader.request.RequestedDuration,
		)
	}
}
