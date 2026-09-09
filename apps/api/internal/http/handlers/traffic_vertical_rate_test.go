package handlers

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
)

func TestToCurrentTrafficItemsPreservesVerticalRateEvidence(t *testing.T) {
	climb := 4.25
	descent := -3.5

	result := toCurrentTrafficItems([]traffic.CurrentTrafficItem{
		{ICAO24: "climb01", VerticalRateMPS: &climb},
		{ICAO24: "desc01", VerticalRateMPS: &descent},
		{ICAO24: "none01", VerticalRateMPS: nil},
	})

	if len(result) != 3 {
		t.Fatalf("traffic DTO count = %d, want 3", len(result))
	}
	if result[0].VerticalRateMPS == nil || *result[0].VerticalRateMPS != climb {
		t.Fatalf("climb vertical rate = %#v, want %v", result[0].VerticalRateMPS, climb)
	}
	if result[1].VerticalRateMPS == nil || *result[1].VerticalRateMPS != descent {
		t.Fatalf("descent vertical rate = %#v, want %v", result[1].VerticalRateMPS, descent)
	}
	if result[2].VerticalRateMPS != nil {
		t.Fatalf("unavailable vertical rate became numeric: %v", *result[2].VerticalRateMPS)
	}
}
