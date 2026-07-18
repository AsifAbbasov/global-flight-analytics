package handlers

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/airportproduction"
)

func TestParseAirportIntelligenceDaysDefaults(t *testing.T) {
	days, err := parseAirportIntelligenceDays("")
	if err != nil {
		t.Fatal(err)
	}
	if days != airportproduction.DefaultWindowDays {
		t.Fatalf("days = %d, want %d", days, airportproduction.DefaultWindowDays)
	}
}
func TestParseAirportIntelligenceDaysRejectsOutOfRange(t *testing.T) {
	_, err := parseAirportIntelligenceDays("366")
	if !errors.Is(err, errAirportIntelligenceDaysInvalid) {
		t.Fatalf("error = %v, want %v", err, errAirportIntelligenceDaysInvalid)
	}
}
func TestParseAirportIntelligenceRankingLimit(t *testing.T) {
	limit, err := parseAirportIntelligenceRankingLimit("25")
	if err != nil {
		t.Fatal(err)
	}
	if limit != 25 {
		t.Fatalf("limit = %d, want 25", limit)
	}
}
