package server

import "github.com/gofiber/fiber/v2"

func registerRouteIntelligenceDatabaseRoutes(
	v1 fiber.Router,
	runtime routeIntelligenceDatabaseRuntime,
	mutationAuthorization fiber.Handler,
) {
	v1.Post(
		"/trajectories/:id/route-intelligence",
		mutationAuthorization,
		runtime.handler.ProcessByTrajectoryID,
	)
	v1.Get(
		"/trajectories/:id/route-intelligence/latest",
		runtime.handler.GetLatestByTrajectoryID,
	)
	v1.Get(
		"/trajectories/:id/route-intelligence/history",
		runtime.handler.ListHistoryByTrajectoryID,
	)
}
