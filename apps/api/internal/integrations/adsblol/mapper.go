package adsblol

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerbatch"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/readsbcompat"
)

const (
	sourceName           = "adsb.lol"
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
	snapshotAt, ok := safeUnixMilliseconds(responseTime)
	if !ok {
		return time.Time{}
	}
	return readsbcompat.ObservationTime(snapshotAt, seen)
}

func mapAircraft(
	item AircraftItem,
	responseTime float64,
) flightstate.FlightState {
	snapshotAt, _ := safeUnixMilliseconds(responseTime)
	return readsbcompat.MapAircraft(
		sourceName,
		item,
		snapshotAt,
	)
}

func aircraftItemRequiredFieldsValid(
	item AircraftItem,
	responseTime float64,
) bool {
	snapshotAt, ok := safeUnixMilliseconds(responseTime)
	if !ok {
		return false
	}
	return readsbcompat.AircraftItemRequiredFieldsValid(
		item,
		snapshotAt,
	)
}

func MapStateResponseWithEvidence(
	response *StateResponse,
) (
	[]flightstate.FlightState,
	providerbatch.Evidence,
	error,
) {
	if response == nil {
		return []flightstate.FlightState{},
			providerbatch.Evidence{},
			nil
	}

	snapshotAt, _ := safeUnixMilliseconds(response.Now)
	return readsbcompat.MapItemsWithEvidence(
		sourceName,
		snapshotAt,
		response.Aircraft,
	)
}

func MapStateResponse(
	response *StateResponse,
) []flightstate.FlightState {
	states, _, _ := MapStateResponseWithEvidence(response)
	return states
}

// OPEN-AVIATION-RESEARCH-EVIDENCE-V1-2
