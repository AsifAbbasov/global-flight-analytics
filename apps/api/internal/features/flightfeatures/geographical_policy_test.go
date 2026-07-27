package flightfeatures

import "testing"

func TestCircularLongitudeSpanSupportsWrappedEnvelope(t *testing.T) {
	tests := []struct {
		name string
		west float64
		east float64
		want float64
	}{
		{name: "ordinary", west: -10, east: 20, want: 30},
		{name: "wrapped", west: 160, east: 0, want: 200},
		{name: "same", west: 42, east: 42, want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CircularLongitudeSpanDegrees(test.west, test.east); got != test.want {
				t.Fatalf("span = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGeographicalPoliciesRemainVersioned(t *testing.T) {
	if CurrentGeographicalDistanceModel != "mean-earth-sphere-haversine-v1" {
		t.Fatalf("distance model = %q", CurrentGeographicalDistanceModel)
	}
	if CurrentGeographicCellPolicy != "decimal-degree-round-half-away-from-zero-v1" {
		t.Fatalf("cell policy = %q", CurrentGeographicCellPolicy)
	}
}
