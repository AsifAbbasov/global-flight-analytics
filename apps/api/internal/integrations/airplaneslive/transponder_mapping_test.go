package airplaneslive

import "testing"

func TestMapAircraftPreservesSquawkCode(t *testing.T) {
	mapped := mapAircraft(
		AircraftItem{
			Hex:       "4k001",
			Flight:    "AHY101",
			Latitude:  40.4093,
			Longitude: 49.8671,
			Squawk:    "7700",
		},
		1760000000000,
	)
	if mapped.SquawkCode != "7700" {
		t.Fatalf(
			"squawk = %q, want 7700",
			mapped.SquawkCode,
		)
	}
}
