package airplaneslive

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerbatch"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/readsbcompat"
)

const (
	sourceName           = "airplanes.live"
	int64BoundaryFloat64 = readsbcompat.Int64BoundaryFloat64
)

func optionalGroundSpeed(
	value OptionalFloat64,
) (float64, bool) {
	return readsbcompat.OptionalGroundSpeed(value)
}

func optionalHeading(
	value OptionalFloat64,
) (float64, bool) {
	return readsbcompat.OptionalHeading(value)
}

func optionalVerticalRate(
	value OptionalFloat64,
) (float64, bool) {
	return readsbcompat.OptionalVerticalRate(value)
}

func safeUnixMilliseconds(
	value float64,
) (time.Time, bool) {
	return readsbcompat.SafeUnixMilliseconds(value)
}

func safeSeenDuration(
	value OptionalFloat64,
) (time.Duration, bool) {
	return readsbcompat.SafeSeenDuration(value)
}

func observationTime(
	responseTime float64,
	seen OptionalFloat64,
) time.Time {
	return readsbcompat.ObservationTime(responseTime, seen)
}

func mapAircraft(
	item AircraftItem,
	responseTime float64,
) flightstate.FlightState {
	return readsbcompat.MapAircraft(
		sourceName,
		item,
		responseTime,
	)
}

func aircraftItemRequiredFieldsValid(
	item AircraftItem,
	responseTime float64,
) bool {
	return readsbcompat.AircraftItemRequiredFieldsValid(
		item,
		responseTime,
	)
}

func MapStateResponseWithEvidence(
	response *StateResponse,
) (
	[]flightstate.FlightState,
	providerbatch.Evidence,
	error,
) {
	return readsbcompat.MapStateResponseWithEvidence(
		sourceName,
		response,
	)
}

func MapStateResponse(
	response *StateResponse,
) []flightstate.FlightState {
	return readsbcompat.MapStateResponse(
		sourceName,
		response,
	)
}

// OPEN-AVIATION-RESEARCH-EVIDENCE-V1-2
