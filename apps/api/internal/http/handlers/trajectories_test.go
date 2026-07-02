package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	trafficquery "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/query"
	"github.com/gofiber/fiber/v2"
)

func TestTrajectoryHandlerGetLatestByICAO24(t *testing.T) {
	repository := &fakeTrajectoryHTTPRepository{
		latestByICAO24: map[string]trajectory.FlightTrajectory{
			"ABC123": makeHTTPTestTrajectory("trajectory-1", "ABC123"),
		},
	}

	handler := NewTrajectoryHandler(trafficquery.New(trafficquery.Config{
		TrajectoryRepository: repository,
	}))

	app := fiber.New()
	app.Get("/api/v1/aircraft/:icao24/trajectory", handler.GetLatestByICAO24)

	response := performTrajectoryRequest(t, app, http.MethodGet, "/api/v1/aircraft/abc123/trajectory")

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body := readTrajectoryResponseBody(t, response)
	if !strings.Contains(body, `"icao24":"ABC123"`) {
		t.Fatalf("expected normalized trajectory response, got %s", body)
	}
}

func TestTrajectoryHandlerGetLatestByICAO24RejectsInvalidICAO24(t *testing.T) {
	handler := NewTrajectoryHandler(trafficquery.New(trafficquery.Config{
		TrajectoryRepository: &fakeTrajectoryHTTPRepository{},
	}))

	app := fiber.New()
	app.Get("/api/v1/aircraft/:icao24/trajectory", handler.GetLatestByICAO24)

	response := performTrajectoryRequest(t, app, http.MethodGet, "/api/v1/aircraft/BAD/trajectory")

	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}
}

func TestTrajectoryHandlerGetLatestByICAO24ReturnsNotFound(t *testing.T) {
	handler := NewTrajectoryHandler(trafficquery.New(trafficquery.Config{
		TrajectoryRepository: &fakeTrajectoryHTTPRepository{},
	}))

	app := fiber.New()
	app.Get("/api/v1/aircraft/:icao24/trajectory", handler.GetLatestByICAO24)

	response := performTrajectoryRequest(t, app, http.MethodGet, "/api/v1/aircraft/ABC123/trajectory")

	if response.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.StatusCode)
	}
}

func TestTrajectoryHandlerGetByID(t *testing.T) {
	repository := &fakeTrajectoryHTTPRepository{
		byID: map[string]trajectory.FlightTrajectory{
			"trajectory-1": makeHTTPTestTrajectory("trajectory-1", "ABC123"),
		},
	}

	handler := NewTrajectoryHandler(trafficquery.New(trafficquery.Config{
		TrajectoryRepository: repository,
	}))

	app := fiber.New()
	app.Get("/api/v1/trajectories/:id", handler.GetByID)

	response := performTrajectoryRequest(t, app, http.MethodGet, "/api/v1/trajectories/trajectory-1")

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body := readTrajectoryResponseBody(t, response)
	if !strings.Contains(body, `"id":"trajectory-1"`) {
		t.Fatalf("expected trajectory id in response, got %s", body)
	}
}

func performTrajectoryRequest(t *testing.T, app *fiber.App, method string, path string) *http.Response {
	t.Helper()

	request := httptest.NewRequest(method, path, nil)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("expected no request error, got %v", err)
	}

	return response
}

func readTrajectoryResponseBody(t *testing.T, response *http.Response) string {
	t.Helper()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("expected readable response body, got %v", err)
	}

	return string(body)
}

type fakeTrajectoryHTTPRepository struct {
	latestByICAO24 map[string]trajectory.FlightTrajectory
	byID           map[string]trajectory.FlightTrajectory
}

func (repository *fakeTrajectoryHTTPRepository) GetLatestTrajectoryByICAO24(
	ctx context.Context,
	icao24 string,
) (trajectory.FlightTrajectory, error) {
	item, ok := repository.latestByICAO24[icao24]
	if !ok {
		return trajectory.FlightTrajectory{}, postgres.ErrTrajectoryNotFound
	}

	return item, nil
}

func (repository *fakeTrajectoryHTTPRepository) GetTrajectoryByID(
	ctx context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
	item, ok := repository.byID[trajectoryID]
	if !ok {
		return trajectory.FlightTrajectory{}, postgres.ErrTrajectoryNotFound
	}

	return item, nil
}

func makeHTTPTestTrajectory(id string, icao24 string) trajectory.FlightTrajectory {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	return trajectory.FlightTrajectory{
		ID:               id,
		ICAO24:           icao24,
		Callsign:         "AHY101",
		StartTime:        now.Add(-5 * time.Minute),
		EndTime:          now,
		DurationSeconds:  300,
		SegmentCount:     1,
		PointCount:       5,
		CoverageGapCount: 0,
		QualityScore:     0.95,
		SourceName:       "test",
		Segments: []trajectory.TrajectorySegment{
			{
				ID:              "segment-1",
				TrajectoryID:    id,
				ICAO24:          icao24,
				Callsign:        "AHY101",
				SequenceNumber:  1,
				Status:          trajectory.SegmentStatusObserved,
				QualityScore:    0.95,
				StartTime:       now.Add(-5 * time.Minute),
				EndTime:         now,
				DurationSeconds: 300,
				StartLatitude:   40.4093,
				StartLongitude:  49.8671,
				EndLatitude:     40.5000,
				EndLongitude:    50.0000,
				PointCount:      5,
				SourceName:      "test",
				CreatedAt:       now,
			},
		},
		CoverageGaps: []trajectory.CoverageGap{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
