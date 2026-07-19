package server

import "github.com/gofiber/fiber/v2"

func registerCoreDatabaseRoutes(
	v1 fiber.Router,
	runtime coreDatabaseRuntime,
) {
	registerCoreRegionRoutes(v1, runtime)
	registerCoreTrafficRoutes(v1, runtime)
	registerCoreFlightRoutes(v1, runtime)
	registerCoreAircraftRoutes(v1, runtime)
	registerCoreAirportRoutes(v1, runtime)
}

func registerCoreRegionRoutes(
	v1 fiber.Router,
	runtime coreDatabaseRuntime,
) {
	v1.Get(
		"/regions",
		runtime.region.List,
	)
	v1.Get(
		"/regions/:code",
		runtime.region.GetByCode,
	)
}

func registerCoreTrafficRoutes(
	v1 fiber.Router,
	runtime coreDatabaseRuntime,
) {
	v1.Get(
		"/metrics/active-aircraft",
		runtime.metrics.GetActiveAircraft,
	)
	v1.Get(
		"/traffic/current",
		runtime.traffic.GetCurrent,
	)
	v1.Get(
		"/aircraft/:icao24/trajectory",
		runtime.trajectory.GetLatestByICAO24,
	)
	v1.Get(
		"/aircraft/:icao24/route-context",
		runtime.routeContext.GetByICAO24,
	)
	v1.Get(
		"/trajectories/:id",
		runtime.trajectory.GetByID,
	)
}

func registerCoreFlightRoutes(
	v1 fiber.Router,
	runtime coreDatabaseRuntime,
) {
	v1.Get(
		"/flights/:flightID/states",
		runtime.flightState.ListByFlightID,
	)
	v1.Get(
		"/aircraft/:icao24/latest-state",
		runtime.flightState.GetLatestByICAO24,
	)
	v1.Get(
		"/flights",
		runtime.flight.List,
	)
	v1.Get(
		"/flights/:id",
		runtime.flight.GetByID,
	)
}

func registerCoreAircraftRoutes(
	v1 fiber.Router,
	runtime coreDatabaseRuntime,
) {
	v1.Get(
		"/aircraft",
		runtime.aircraft.List,
	)
	v1.Get(
		"/aircraft/:icao24",
		runtime.aircraft.GetByICAO24,
	)
}

func registerCoreAirportRoutes(
	v1 fiber.Router,
	runtime coreDatabaseRuntime,
) {
	v1.Get(
		"/airports",
		runtime.airport.List,
	)
	v1.Get(
		"/airports/:icao",
		runtime.airport.GetByICAO,
	)
}
