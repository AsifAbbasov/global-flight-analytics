package opensky

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestMapStateVectorPreservesObservationMetadata(t *testing.T) {
	snapshot := time.Date(
		2026,
		time.July,
		18,
		0,
		0,
		10,
		0,
		time.UTC,
	)
	positionTime := snapshot.Add(-time.Second)
	latitude := 40.4093
	longitude := 49.8671
	squawk := "7700"

	mapped, usable, err := MapStateVector(
		StateVector{
			SnapshotTime:      snapshot,
			ICAO24:            "4k001",
			OriginCountry:     "Azerbaijan",
			TimePosition:      &positionTime,
			LastContact:       snapshot,
			Latitude:          &latitude,
			Longitude:         &longitude,
			Squawk:            &squawk,
			SPI:               true,
			PositionSource:    PositionSourceMLAT,
			Category:          AircraftCategoryHeavy,
			CategoryAvailable: true,
		},
	)
	if err != nil {
		t.Fatalf("map state vector: %v", err)
	}
	if !usable {
		t.Fatal("expected usable state")
	}
	if mapped.SquawkCode != "7700" ||
		!mapped.SpecialPurposeIndicator ||
		mapped.PositionSource != flightstate.PositionSourceMLAT ||
		mapped.AircraftCategory != int(AircraftCategoryHeavy) ||
		!mapped.AircraftCategoryAvailable {
		t.Fatalf("mapped metadata = %#v", mapped)
	}
}
