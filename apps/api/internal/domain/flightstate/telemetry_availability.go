package flightstate

func (state FlightState) HasVelocity() bool {
	if !state.TelemetryAvailabilityKnown {
		return true
	}
	return state.VelocityAvailable
}

func (state FlightState) HasHeading() bool {
	if !state.TelemetryAvailabilityKnown {
		return true
	}
	return state.HeadingAvailable
}

func (state FlightState) HasVerticalRate() bool {
	if !state.TelemetryAvailabilityKnown {
		return true
	}
	return state.VerticalRateAvailable
}

func (state FlightState) HasOnGroundState() bool {
	if !state.TelemetryAvailabilityKnown {
		return true
	}
	return state.OnGroundAvailable
}

func (state FlightState) HasCompleteKinematics() bool {
	return state.HasVelocity() &&
		state.HasHeading() &&
		state.HasVerticalRate() &&
		state.HasOnGroundState()
}
