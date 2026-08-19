package live

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func (s *Store) UpsertFlightStates(
	states []flightstate.FlightState,
	receivedAt time.Time,
) UpsertResult {
	candidates := make([]Aircraft, 0, len(states))
	for _, state := range states {
		candidates = append(candidates, aircraftFromFlightState(state, receivedAt))
	}
	return s.UpsertBatch(candidates)
}

func aircraftFromFlightState(
	state flightstate.FlightState,
	receivedAt time.Time,
) Aircraft {
	return Aircraft{
		ICAO24:          state.ICAO24,
		Callsign:        state.Callsign,
		Latitude:        state.Latitude,
		Longitude:       state.Longitude,
		AltitudeM:       flightStateAltitude(state),
		VelocityMPS:     optionalFloat64(state.VelocityMPS, state.VelocityAvailable),
		HeadingDegrees:  optionalFloat64(state.HeadingDegrees, state.HeadingAvailable),
		VerticalRateMPS: optionalFloat64(state.VerticalRateMPS, state.VerticalRateAvailable),
		OnGround:        optionalBoolValue(state.OnGround, state.OnGroundAvailable),
		ObservedAt:      state.ObservedAt,
		ReceivedAt:      receivedAt,
		Source:          state.SourceName,
	}
}

func flightStateAltitude(state flightstate.FlightState) *float64 {
	if state.BarometricAltitudeStatus == flightstate.AltitudeStatusObserved ||
		state.BarometricAltitudeStatus == flightstate.AltitudeStatusGround {
		return optionalFloat64(state.BarometricAltitudeM, true)
	}
	if state.GeometricAltitudeStatus == flightstate.AltitudeStatusObserved ||
		state.GeometricAltitudeStatus == flightstate.AltitudeStatusGround {
		return optionalFloat64(state.GeometricAltitudeM, true)
	}
	return nil
}

func optionalFloat64(value float64, available bool) *float64 {
	if !available {
		return nil
	}
	result := value
	return &result
}

func optionalBoolValue(value bool, available bool) *bool {
	if !available {
		return nil
	}
	result := value
	return &result
}
