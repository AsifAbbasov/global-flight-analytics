package transponderalert

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestBuildPreservesObservedCodeWithoutConfirmingIncident(t *testing.T) {
	start := time.Date(
		2026,
		time.July,
		18,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	states := []flightstate.FlightState{
		{
			ICAO24:     "4k001",
			Callsign:   "AHY101",
			SquawkCode: "7700",
			ObservedAt: start,
			SourceName: "opensky",
		},
		{
			ICAO24:                  "4K001",
			Callsign:                "AHY101",
			SquawkCode:              "7700",
			SpecialPurposeIndicator: true,
			ObservedAt:              start.Add(20 * time.Second),
			SourceName:              "opensky",
		},
	}

	result, err := Build(
		states,
		start.Add(time.Minute),
	)
	if err != nil {
		t.Fatalf("build evidence: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("evidence count = %d, want 1", len(result))
	}
	evidence := result[0]
	if evidence.Kind != KindGeneralEmergencyCode {
		t.Fatalf("kind = %s", evidence.Kind)
	}
	if evidence.Strength != StrengthRepeatedObservation {
		t.Fatalf("strength = %s", evidence.Strength)
	}
	if evidence.MaximumClaimStrength !=
		"observed_transponder_code_only" {
		t.Fatalf(
			"claim strength = %q",
			evidence.MaximumClaimStrength,
		)
	}
	if len(evidence.Limitations) < 4 {
		t.Fatalf("limitations = %v", evidence.Limitations)
	}
	if evidence.Fingerprint == "" {
		t.Fatal("fingerprint is empty")
	}
}

func TestBuildIgnoresOrdinarySquawkCodes(t *testing.T) {
	result, err := Build(
		[]flightstate.FlightState{
			{
				ICAO24:     "4k001",
				SquawkCode: "1200",
				ObservedAt: time.Now().UTC(),
			},
		},
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("build evidence: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("evidence count = %d, want 0", len(result))
	}
}

func TestBuildRejectsInvalidSquawkCode(t *testing.T) {
	_, err := Build(
		[]flightstate.FlightState{
			{
				ICAO24:     "4k001",
				SquawkCode: "7800",
				ObservedAt: time.Now().UTC(),
			},
		},
		time.Now().UTC(),
	)
	if !errors.Is(
		err,
		flightstate.ErrSquawkCodeInvalid,
	) {
		t.Fatalf(
			"error = %v, want %v",
			err,
			flightstate.ErrSquawkCodeInvalid,
		)
	}
}
