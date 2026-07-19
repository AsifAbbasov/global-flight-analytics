package server

import "github.com/gofiber/fiber/v2"

const TransponderEvidencePath = "/aircraft/:icao24/transponder-evidence/latest"

func registerTransponderEvidenceDatabaseRoutes(
	v1 fiber.Router,
	runtime transponderEvidenceDatabaseRuntime,
) {
	v1.Get(
		TransponderEvidencePath,
		runtime.handler.GetLatest,
	)
}
