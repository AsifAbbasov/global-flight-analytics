package readsbcompat

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestMapItemsWithEvidenceRequiresSourceName(
	t *testing.T,
) {
	_, _, err := MapItemsWithEvidence(
		"   ",
		time.Now().UTC(),
		nil,
	)
	if !errors.Is(err, ErrSourceNameRequired) {
		t.Fatalf("expected source-name error, got %v", err)
	}
}

func TestSafeUnixSecondsPreservesFractionalReadsbTimestamp(
	t *testing.T,
) {
	actual, ok := SafeUnixSeconds(1720526400.125)
	if !ok {
		t.Fatal("expected readsb unix-seconds timestamp to be accepted")
	}
	want := time.Unix(1720526400, 125_000_000).UTC()
	if !actual.Equal(want) {
		t.Fatalf("expected %s, got %s", want, actual)
	}
}

func TestSafeUnixMillisecondsRejectsFractionalProviderTimestamp(
	t *testing.T,
) {
	if _, ok := SafeUnixMilliseconds(1720526400000.5); ok {
		t.Fatal("expected fractional millisecond timestamp to fail closed")
	}
}

func TestMapItemsWithEvidencePreservesCanonicalSemantics(
	t *testing.T,
) {
	snapshotAt := time.Date(
		2026,
		time.August,
		24,
		10,
		0,
		0,
		0,
		time.UTC,
	)
	geometricAltitudeFeet := float64(12000)

	states, evidence, err := MapItemsWithEvidence(
		"local-readsb:test-receiver",
		snapshotAt,
		[]AircraftItem{
			{
				Hex:       " abc123 ",
				Flight:    " TEST123 ",
				Latitude:  40.4093,
				Longitude: 49.8671,
				AltBaro: BarometricAltitude{
					Feet: 10000,
					Kind: BarometricAltitudeKindObserved,
				},
				AltGeom: &geometricAltitudeFeet,
				GroundSpeed: OptionalFloat64{
					Value:     250,
					Available: true,
				},
				Track: OptionalFloat64{
					Value:     90,
					Available: true,
				},
				BaroRate: OptionalFloat64{
					Value:     500,
					Available: true,
				},
				Seen: OptionalFloat64{
					Value:     2,
					Available: true,
				},
				Squawk: " 1234 ",
			},
		},
	)
	if err != nil {
		t.Fatalf("map items: %v", err)
	}
	if evidence.Received != 1 || evidence.Accepted != 1 || evidence.RejectedMalformed != 0 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
	if len(states) != 1 {
		t.Fatalf("expected one state, got %d", len(states))
	}

	state := states[0]
	if state.SourceName != "local-readsb:test-receiver" {
		t.Fatalf("unexpected source name %q", state.SourceName)
	}
	if state.ICAO24 != "ABC123" {
		t.Fatalf("unexpected ICAO24 %q", state.ICAO24)
	}
	if state.Callsign != "TEST123" {
		t.Fatalf("unexpected callsign %q", state.Callsign)
	}
	if state.SquawkCode != "1234" {
		t.Fatalf("unexpected squawk %q", state.SquawkCode)
	}
	if state.BarometricAltitudeStatus != flightstate.AltitudeStatusObserved {
		t.Fatalf("unexpected barometric altitude status %q", state.BarometricAltitudeStatus)
	}
	if state.GeometricAltitudeStatus != flightstate.AltitudeStatusObserved {
		t.Fatalf("unexpected geometric altitude status %q", state.GeometricAltitudeStatus)
	}
	if !state.VelocityAvailable || !state.HeadingAvailable || !state.VerticalRateAvailable {
		t.Fatal("expected optional telemetry to remain available")
	}
	wantObservedAt := snapshotAt.Add(-2 * time.Second)
	if !state.ObservedAt.Equal(wantObservedAt) {
		t.Fatalf("expected observed_at %s, got %s", wantObservedAt, state.ObservedAt)
	}
}

func TestMapItemsWithEvidenceRejectsMalformedItemsWithoutPoisoningBatch(
	t *testing.T,
) {
	snapshotAt := time.Date(
		2026,
		time.August,
		24,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	states, evidence, err := MapItemsWithEvidence(
		"test-source",
		snapshotAt,
		[]AircraftItem{
			{
				Hex:       "valid1",
				Latitude:  40,
				Longitude: 49,
			},
			{
				Hex:       "",
				Latitude:  40,
				Longitude: 49,
			},
		},
	)
	if err != nil {
		t.Fatalf("map mixed batch: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected one accepted state, got %d", len(states))
	}
	if evidence.Received != 2 || evidence.Accepted != 1 || evidence.RejectedMalformed != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}
