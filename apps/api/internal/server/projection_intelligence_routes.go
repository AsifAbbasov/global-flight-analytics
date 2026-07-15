package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/gofiber/fiber/v2"
)

const ProjectionIntelligencePath = "/trajectories/:id/projection-intelligence"

// RegisterProjectionIntelligenceReadRoute composes the read-only Projection
// Intelligence endpoint with an already constructed query service. Production
// server wiring and runtime verification may supply different readers while
// preserving the same HTTP contract.
func RegisterProjectionIntelligenceReadRoute(
	v1 fiber.Router,
	reader handlers.ProjectionIntelligenceReader,
) error {
	if reader == nil {
		return fmt.Errorf(
			"Projection Intelligence reader is required",
		)
	}

	handler :=
		handlers.NewProjectionIntelligenceHandler(
			reader,
		)
	v1.Get(
		ProjectionIntelligencePath,
		handler.GetByTrajectoryID,
	)

	return nil
}
