package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestBuildVerificationScheduleCreatesFutureEvidenceBoundary(t *testing.T) {
	now := time.Date(2026, time.July, 16, 18, 30, 45, 987654321, time.UTC)

	schedule, err := buildVerificationSchedule(now)
	if err != nil {
		t.Fatalf("build verification schedule: %v", err)
	}

	if len(schedule.PointTimes) != storedPointCount {
		t.Fatalf("point count = %d, want %d", len(schedule.PointTimes), storedPointCount)
	}
	if !schedule.GeneratedAt.Equal(now.Truncate(time.Second)) {
		t.Fatalf("generated-at = %s, want %s", schedule.GeneratedAt, now.Truncate(time.Second))
	}
	if !schedule.PointTimes[boundedPointCount-1].Equal(schedule.AsOfTime) {
		t.Fatalf("last bounded point = %s, want as-of time %s", schedule.PointTimes[boundedPointCount-1], schedule.AsOfTime)
	}
	if !schedule.TrajectoryEnd.Equal(schedule.AsOfTime) {
		t.Fatalf("trajectory end = %s, want as-of time %s", schedule.TrajectoryEnd, schedule.AsOfTime)
	}
	if !schedule.PointTimes[storedPointCount-1].After(schedule.AsOfTime) {
		t.Fatalf("future trajectory point = %s, want after %s", schedule.PointTimes[storedPointCount-1], schedule.AsOfTime)
	}
	if schedule.WeatherObservedAt.After(schedule.AsOfTime) ||
		schedule.WeatherRetrievedAt.After(schedule.AsOfTime) {
		t.Fatalf("bounded weather evidence exceeds as-of time")
	}
	if !schedule.FutureWeatherAt.After(schedule.AsOfTime) ||
		!schedule.FutureWeatherReadAt.After(schedule.AsOfTime) {
		t.Fatalf("future weather evidence does not exceed as-of time")
	}
}

func TestWeatherContextRequestURL(t *testing.T) {
	asOfTime := time.Date(2026, time.July, 16, 12, 0, 0, 0, time.UTC)

	requestURL := weatherContextRequestURL(
		verificationTrajectoryID,
		asOfTime,
		verificationDuration,
	)
	parts := strings.SplitN(requestURL, "?", 2)
	if len(parts) != 2 {
		t.Fatalf("request URL has no query string: %q", requestURL)
	}
	if parts[0] != "/api/v1/trajectories/"+verificationTrajectoryID+"/weather-context" {
		t.Fatalf("request path = %q", parts[0])
	}

	values, err := url.ParseQuery(parts[1])
	if err != nil {
		t.Fatalf("parse request query: %v", err)
	}
	if values.Get("as_of_time") != asOfTime.Format(time.RFC3339Nano) {
		t.Fatalf("as_of_time = %q", values.Get("as_of_time"))
	}
	if values.Get("duration_seconds") != "180" {
		t.Fatalf("duration_seconds = %q, want 180", values.Get("duration_seconds"))
	}
}

func TestVerificationCoordinatesRemainNearLatestPoint(t *testing.T) {
	firstLatitude, firstLongitude := verificationCoordinates(0)
	latestLatitude, latestLongitude := verificationCoordinates(boundedPointCount - 1)

	if latestLatitude <= firstLatitude || latestLongitude <= firstLongitude {
		t.Fatalf("verification coordinates are not ordered")
	}
	if latestLatitude-firstLatitude >= 1 || latestLongitude-firstLongitude >= 1 {
		t.Fatalf("verification coordinates exceed PostgreSQL snapshot search boundary")
	}
}

func TestContainsString(t *testing.T) {
	items := []string{"surface_context", "projection_uncertainty"}
	if !containsString(items, "projection_uncertainty") {
		t.Fatalf("expected projection uncertainty scope")
	}
	if containsString(items, "missing") {
		t.Fatalf("unexpected scope match")
	}
}

func TestExecuteHTTPTestAllowsProductionLatencyAboveFiberDefault(t *testing.T) {
	app := fiber.New()
	app.Get(
		"/slow",
		func(ctx *fiber.Ctx) error {
			time.Sleep(1100 * time.Millisecond)
			return ctx.SendStatus(fiber.StatusNoContent)
		},
	)

	response, err := executeHTTPTest(
		app,
		httptest.NewRequest(
			http.MethodGet,
			"/slow",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("execute slow HTTP verification request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf(
			"status = %d, want %d",
			response.StatusCode,
			fiber.StatusNoContent,
		)
	}
}

func TestRuntimeTimeoutPolicyIsBoundedAndAboveFiberDefault(t *testing.T) {
	if runtimeHTTPTestTimeout <= time.Second {
		t.Fatalf(
			"runtime HTTP timeout = %s, must exceed Fiber default",
			runtimeHTTPTestTimeout,
		)
	}
	if runtimeVerificationTimeout <= runtimeHTTPTestTimeout {
		t.Fatalf(
			"overall runtime timeout = %s, must exceed HTTP timeout %s",
			runtimeVerificationTimeout,
			runtimeHTTPTestTimeout,
		)
	}
	if fixtureCleanupTimeout < runtimeHTTPTestTimeout {
		t.Fatalf(
			"fixture cleanup timeout = %s, must cover HTTP timeout %s",
			fixtureCleanupTimeout,
			runtimeHTTPTestTimeout,
		)
	}
}
