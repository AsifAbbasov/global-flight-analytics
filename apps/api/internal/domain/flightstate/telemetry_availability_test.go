package flightstate

import "testing"

func TestLegacyTelemetryAvailabilityRemainsCompatible(
	t *testing.T,
) {
	state := FlightState{}

	if !state.HasVelocity() ||
		!state.HasHeading() ||
		!state.HasVerticalRate() ||
		!state.HasOnGroundState() ||
		!state.HasCompleteKinematics() {
		t.Fatal(
			"legacy state without explicit availability must preserve existing semantics",
		)
	}
}

func TestExplicitUnavailableTelemetryRemainsUnavailable(
	t *testing.T,
) {
	state := FlightState{
		TelemetryAvailabilityKnown: true,
	}

	if state.HasVelocity() ||
		state.HasHeading() ||
		state.HasVerticalRate() ||
		state.HasOnGroundState() ||
		state.HasCompleteKinematics() {
		t.Fatal(
			"explicitly unavailable telemetry was reported as available",
		)
	}
}

func TestExplicitZeroTelemetryRemainsAvailable(
	t *testing.T,
) {
	state := FlightState{
		TelemetryAvailabilityKnown: true,
		VelocityAvailable:          true,
		HeadingAvailable:           true,
		VerticalRateAvailable:      true,
		OnGroundAvailable:          true,
		VelocityMPS:                0,
		HeadingDegrees:             0,
		VerticalRateMPS:            0,
		OnGround:                   false,
	}

	if !state.HasCompleteKinematics() {
		t.Fatal(
			"real zero telemetry must remain distinguishable from unavailable telemetry",
		)
	}
}
