package flightstate

import "testing"

func TestResolveAltitudeStatusPreservesExplicitStatus(
	t *testing.T,
) {
	statuses := []AltitudeStatus{
		AltitudeStatusObserved,
		AltitudeStatusGround,
		AltitudeStatusUnknown,
		AltitudeStatusUnavailable,
		AltitudeStatusInvalid,
	}

	for _, status := range statuses {
		resolved := ResolveAltitudeStatus(
			0,
			status,
		)

		if resolved != status {
			t.Fatalf(
				"expected explicit status %q to be preserved, got %q",
				status,
				resolved,
			)
		}
	}
}

func TestResolveAltitudeStatusUsesCompatibilityRuleForLegacyStates(
	t *testing.T,
) {
	observed := ResolveAltitudeStatus(
		100,
		"",
	)
	if observed != AltitudeStatusObserved {
		t.Fatalf(
			"expected non-zero legacy altitude to resolve to observed, got %q",
			observed,
		)
	}

	unavailable := ResolveAltitudeStatus(
		0,
		"",
	)
	if unavailable != AltitudeStatusUnavailable {
		t.Fatalf(
			"expected zero legacy altitude to resolve to unavailable, got %q",
			unavailable,
		)
	}
}
