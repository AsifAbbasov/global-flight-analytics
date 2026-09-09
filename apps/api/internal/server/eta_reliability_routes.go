package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/gofiber/fiber/v2"
)

const ETAReliabilityPath = "/trajectories/:id/eta-reliability"

func RegisterETAReliabilityReadRoute(v1 fiber.Router, reader handlers.ETAReliabilityReader) error {
	if reader == nil {
		return fmt.Errorf("ETA Reliability reader is required")
	}
	handler := handlers.NewETAReliabilityHandler(reader)
	v1.Get(ETAReliabilityPath, handler.GetByTrajectoryID)
	return nil
}
