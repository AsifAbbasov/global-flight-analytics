package trajectory

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"

// TrackPoint4DFromFlightState preserves both telemetry values and their
// availability contract when a validated flight state enters trajectory
// processing or is reconstructed from PostgreSQL.
func TrackPoint4DFromFlightState(state flightstate.FlightState) TrackPoint4D {
	return TrackPoint4D{
		ID:                  state.ID,
		FlightStateID:       state.ID,
		FlightID:            state.FlightID,
		AircraftID:          state.AircraftID,
		ICAO24:              state.ICAO24,
		Callsign:            state.Callsign,
		Latitude:            state.Latitude,
		Longitude:           state.Longitude,
		BarometricAltitudeM: state.BarometricAltitudeM,
		BarometricAltitudeStatus: flightstate.ResolveAltitudeStatus(
			state.BarometricAltitudeM,
			state.BarometricAltitudeStatus,
		),
		GeometricAltitudeM: state.GeometricAltitudeM,
		GeometricAltitudeStatus: flightstate.ResolveAltitudeStatus(
			state.GeometricAltitudeM,
			state.GeometricAltitudeStatus,
		),
		VelocityMPS:                state.VelocityMPS,
		VelocityAvailable:          state.VelocityAvailable,
		HeadingDegrees:             state.HeadingDegrees,
		HeadingAvailable:           state.HeadingAvailable,
		VerticalRateMPS:            state.VerticalRateMPS,
		VerticalRateAvailable:      state.VerticalRateAvailable,
		OnGround:                   state.OnGround,
		OnGroundAvailable:          state.OnGroundAvailable,
		TelemetryAvailabilityKnown: state.TelemetryAvailabilityKnown,
		OriginCountry:              state.OriginCountry,
		ObservedAt:                 state.ObservedAt,
		SourceName:                 state.SourceName,
	}
}

func (point TrackPoint4D) HasVelocity() bool {
	if !point.TelemetryAvailabilityKnown {
		return true
	}
	return point.VelocityAvailable
}

func (point TrackPoint4D) HasHeading() bool {
	if !point.TelemetryAvailabilityKnown {
		return true
	}
	return point.HeadingAvailable
}

func (point TrackPoint4D) HasVerticalRate() bool {
	if !point.TelemetryAvailabilityKnown {
		return true
	}
	return point.VerticalRateAvailable
}

func (point TrackPoint4D) HasOnGroundState() bool {
	if !point.TelemetryAvailabilityKnown {
		return true
	}
	return point.OnGroundAvailable
}
