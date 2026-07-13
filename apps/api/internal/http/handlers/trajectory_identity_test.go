package handlers

import (
	"net/http"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	trafficquery "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/query"
	"github.com/gofiber/fiber/v2"
)

func TestTrajectoryHandlerReturnsFlightIdentityMetadata(
	t *testing.T,
) {
	item := makeHTTPTestTrajectory(
		"trajectory-identity-1",
		"ABC123",
	)
	item.IdentityKey = "flight-identity-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	item.IdentityBasis = trajectory.FlightIdentityBasisCallsignAndStartTime
	item.SplitReason = trajectory.FlightSplitReasonInitialObservation

	repository := &fakeTrajectoryHTTPRepository{
		byID: map[string]trajectory.FlightTrajectory{
			item.ID: item,
		},
	}

	handler := NewTrajectoryHandler(
		trafficquery.New(
			trafficquery.Config{
				TrajectoryRepository: repository,
			},
		),
	)

	app := fiber.New()
	app.Get(
		"/api/v1/trajectories/:id",
		handler.GetByID,
	)

	response := performTrajectoryRequest(
		t,
		app,
		http.MethodGet,
		"/api/v1/trajectories/trajectory-identity-1",
	)
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			response.StatusCode,
		)
	}

	body := readTrajectoryResponseBody(t, response)
	for _, expected := range []string{
		`"identity_key":"flight-identity-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`,
		`"identity_basis":"callsign_and_start_time"`,
		`"split_reason":"initial_observation"`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf(
				"expected response to contain %s, got %s",
				expected,
				body,
			)
		}
	}
}
