package flightcontinuation

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestContinueUsesSameSourceFlightIdentifier(
	t *testing.T,
) {
	previous, current := sourceIdentityPair()
	result, continued := Continue(
		previous,
		current,
		Config{
			MaxGap: 5 * time.Minute,
		},
	)

	if !continued {
		t.Fatal("expected source flight identity continuation")
	}

	assertContinuedIdentity(
		t,
		previous,
		result,
	)
}

func TestContinueUsesSameCallsignWithoutSourceFlightIdentifier(
	t *testing.T,
) {
	previous, current := callsignIdentityPair()
	result, continued := Continue(
		previous,
		current,
		Config{
			MaxGap: 5 * time.Minute,
		},
	)

	if !continued {
		t.Fatal("expected callsign identity continuation")
	}

	assertContinuedIdentity(
		t,
		previous,
		result,
	)
}

func TestContinueRejectsDifferentSourceFlightIdentifiers(
	t *testing.T,
) {
	previous, current := sourceIdentityPair()
	current.FlightID =
		"22222222-2222-2222-2222-222222222222"

	_, continued := Continue(
		previous,
		current,
		Config{
			MaxGap: 5 * time.Minute,
		},
	)

	if continued {
		t.Fatal("expected different source flight identifiers to reject continuation")
	}
}

func TestContinueRejectsDifferentCallsigns(
	t *testing.T,
) {
	previous, current := callsignIdentityPair()
	current.Callsign = "AHY102"

	_, continued := Continue(
		previous,
		current,
		Config{
			MaxGap: 5 * time.Minute,
		},
	)

	if continued {
		t.Fatal("expected different callsigns to reject continuation")
	}
}

func TestContinueRejectsGapBeyondPolicy(
	t *testing.T,
) {
	previous, current := callsignIdentityPair()
	current.StartTime =
		previous.EndTime.Add(6 * time.Minute)

	_, continued := Continue(
		previous,
		current,
		Config{
			MaxGap: 5 * time.Minute,
		},
	)

	if continued {
		t.Fatal("expected an excessive time gap to reject continuation")
	}
}

func TestContinueRejectsAircraftOnlyIdentity(
	t *testing.T,
) {
	previous, current := callsignIdentityPair()
	previous.IdentityBasis =
		trajectory.FlightIdentityBasisAircraftAndStartTime
	current.IdentityBasis =
		trajectory.FlightIdentityBasisAircraftAndStartTime
	previous.Callsign = ""
	current.Callsign = ""

	_, continued := Continue(
		previous,
		current,
		Config{
			MaxGap: 5 * time.Minute,
		},
	)

	if continued {
		t.Fatal("expected aircraft-only identity to reject continuation")
	}
}

func TestContinueRejectsAlreadySplitCurrentGroup(
	t *testing.T,
) {
	previous, current := callsignIdentityPair()
	current.SplitReason =
		trajectory.FlightSplitReasonCallsignChanged

	_, continued := Continue(
		previous,
		current,
		Config{
			MaxGap: 5 * time.Minute,
		},
	)

	if continued {
		t.Fatal("expected a split current group to reject continuation")
	}
}

func TestContinueDoesNotMutateInputs(
	t *testing.T,
) {
	previous, current := callsignIdentityPair()
	originalCurrent := current

	_, continued := Continue(
		previous,
		current,
		Config{
			MaxGap: 5 * time.Minute,
		},
	)
	if !continued {
		t.Fatal("expected continuation")
	}

	if !reflect.DeepEqual(
		current,
		originalCurrent,
	) {
		t.Fatal("expected current trajectory input to remain unchanged")
	}
}

func TestConfigRejectsNegativeMaximumGap(
	t *testing.T,
) {
	err := (Config{
		MaxGap: -time.Second,
	}).Validate()

	if err == nil {
		t.Fatal("expected negative maximum gap validation error")
	}
}

func sourceIdentityPair() (
	trajectory.FlightTrajectory,
	trajectory.FlightTrajectory,
) {
	previous, current := callsignIdentityPair()
	previous.IdentityBasis =
		trajectory.FlightIdentityBasisSourceFlightID
	current.IdentityBasis =
		trajectory.FlightIdentityBasisSourceFlightID
	previous.FlightID =
		"11111111-1111-1111-1111-111111111111"
	current.FlightID =
		"11111111-1111-1111-1111-111111111111"

	return previous, current
}

func callsignIdentityPair() (
	trajectory.FlightTrajectory,
	trajectory.FlightTrajectory,
) {
	now := time.Date(
		2026,
		time.July,
		13,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	previous := trajectory.FlightTrajectory{
		IdentityKey: identityKey("a"),
		IdentityBasis: trajectory.
			FlightIdentityBasisCallsignAndStartTime,
		SplitReason: trajectory.
			FlightSplitReasonInitialObservation,
		ICAO24:   "ABC123",
		Callsign: "AHY101",
		StartTime: now.Add(
			-10 * time.Minute,
		),
		EndTime: now,
	}

	current := trajectory.FlightTrajectory{
		IdentityKey: identityKey("b"),
		IdentityBasis: trajectory.
			FlightIdentityBasisCallsignAndStartTime,
		SplitReason: trajectory.
			FlightSplitReasonInitialObservation,
		ICAO24:   "ABC123",
		Callsign: "AHY101",
		StartTime: now.Add(
			time.Minute,
		),
		EndTime: now.Add(
			2 * time.Minute,
		),
	}

	return previous, current
}

func identityKey(character string) string {
	return identityKeyPrefix +
		strings.Repeat(character, 64)
}

func assertContinuedIdentity(
	t *testing.T,
	previous trajectory.FlightTrajectory,
	actual trajectory.FlightTrajectory,
) {
	t.Helper()

	if actual.IdentityKey != previous.IdentityKey {
		t.Fatalf(
			"expected identity key %s, got %s",
			previous.IdentityKey,
			actual.IdentityKey,
		)
	}

	if actual.IdentityBasis != previous.IdentityBasis {
		t.Fatalf(
			"expected identity basis %s, got %s",
			previous.IdentityBasis,
			actual.IdentityBasis,
		)
	}

	if actual.SplitReason !=
		trajectory.FlightSplitReasonContinuedFromPreviousBatch {
		t.Fatalf(
			"expected continued split reason, got %s",
			actual.SplitReason,
		)
	}
}
