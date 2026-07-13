package passport

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

func TestBuilderBuildsAirportPassport(t *testing.T) {
	generatedAt := time.Date(2026, time.July, 13, 10, 0, 0, 0, time.FixedZone("AZT", 4*60*60))
	observedAt := generatedAt.Add(-2 * time.Minute)

	result, err := NewBuilder().Build(
		airport.Airport{
			ICAOCode:    " ubba ",
			IATACode:    " gyd ",
			Name:        " Heydar Aliyev International Airport ",
			City:        " Baku ",
			Country:     " Azerbaijan ",
			Latitude:    40.4675,
			Longitude:   50.0467,
			ElevationM:  3,
			Timezone:    " Asia/Baku ",
			Description: " Main international airport ",
		},
		AnalyticsInput{
			Arrivals:       12,
			Departures:     9,
			ActiveAircraft: 7,
			FreshnessScore: 0.95,
			CoverageScore:  0.8,
			ObservedAt:     observedAt,
		},
		generatedAt,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if result.Identity.ICAOCode != "UBBA" {
		t.Fatalf("ICAOCode = %q, want UBBA", result.Identity.ICAOCode)
	}
	if result.Identity.IATACode != "GYD" {
		t.Fatalf("IATACode = %q, want GYD", result.Identity.IATACode)
	}
	if result.Operations.Activity != 21 {
		t.Fatalf("Activity = %d, want 21", result.Operations.Activity)
	}
	if result.DataQuality.FreshnessScore != 0.95 {
		t.Fatalf("FreshnessScore = %v, want 0.95", result.DataQuality.FreshnessScore)
	}
	if result.GeneratedAt.Location() != time.UTC {
		t.Fatalf("GeneratedAt location = %v, want UTC", result.GeneratedAt.Location())
	}
	if result.DataQuality.ObservedAt.Location() != time.UTC {
		t.Fatalf("ObservedAt location = %v, want UTC", result.DataQuality.ObservedAt.Location())
	}
}

func TestBuilderAllowsMissingOptionalAirportFields(t *testing.T) {
	now := time.Date(2026, time.July, 13, 9, 0, 0, 0, time.UTC)

	result, err := NewBuilder().Build(
		airport.Airport{
			ICAOCode:  "UBBY",
			Name:      "Zabrat Airport",
			Latitude:  40.4955,
			Longitude: 49.9768,
		},
		AnalyticsInput{
			ObservedAt: now,
		},
		now,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Identity.IATACode != "" {
		t.Fatalf("IATACode = %q, want empty", result.Identity.IATACode)
	}
}

func TestBuilderRejectsInvalidIdentity(t *testing.T) {
	now := time.Now().UTC()
	tests := []airport.Airport{
		{Name: "Airport", Latitude: 40, Longitude: 50},
		{ICAOCode: "UBBA", Latitude: 40, Longitude: 50},
	}

	for _, source := range tests {
		_, err := NewBuilder().Build(source, AnalyticsInput{ObservedAt: now}, now)
		if !errors.Is(err, ErrInvalidIdentity) {
			t.Fatalf("Build() error = %v, want ErrInvalidIdentity", err)
		}
	}
}

func TestBuilderRejectsInvalidCoordinates(t *testing.T) {
	now := time.Now().UTC()
	tests := []airport.Airport{
		{ICAOCode: "UBBA", Name: "Airport", Latitude: 91, Longitude: 50},
		{ICAOCode: "UBBA", Name: "Airport", Latitude: 40, Longitude: 181},
	}

	for _, source := range tests {
		_, err := NewBuilder().Build(source, AnalyticsInput{ObservedAt: now}, now)
		if !errors.Is(err, ErrInvalidCoordinates) {
			t.Fatalf("Build() error = %v, want ErrInvalidCoordinates", err)
		}
	}
}

func TestBuilderRejectsInvalidOperations(t *testing.T) {
	now := time.Now().UTC()
	source := validAirport()
	tests := []AnalyticsInput{
		{Arrivals: -1, ObservedAt: now},
		{Departures: -1, ObservedAt: now},
		{ActiveAircraft: -1, ObservedAt: now},
	}

	for _, analytics := range tests {
		_, err := NewBuilder().Build(source, analytics, now)
		if !errors.Is(err, ErrInvalidOperations) {
			t.Fatalf("Build() error = %v, want ErrInvalidOperations", err)
		}
	}
}

func TestBuilderRejectsInvalidDataQuality(t *testing.T) {
	now := time.Now().UTC()
	tests := []AnalyticsInput{
		{FreshnessScore: -0.1, ObservedAt: now},
		{FreshnessScore: 1.1, ObservedAt: now},
		{CoverageScore: -0.1, ObservedAt: now},
		{CoverageScore: 1.1, ObservedAt: now},
	}

	for _, analytics := range tests {
		_, err := NewBuilder().Build(validAirport(), analytics, now)
		if !errors.Is(err, ErrInvalidDataQuality) {
			t.Fatalf("Build() error = %v, want ErrInvalidDataQuality", err)
		}
	}
}

func TestBuilderRejectsInvalidTimes(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		analytics   AnalyticsInput
		generatedAt time.Time
	}{
		{analytics: AnalyticsInput{ObservedAt: now}, generatedAt: time.Time{}},
		{analytics: AnalyticsInput{}, generatedAt: now},
		{analytics: AnalyticsInput{ObservedAt: now.Add(time.Second)}, generatedAt: now},
	}

	for _, test := range tests {
		_, err := NewBuilder().Build(validAirport(), test.analytics, test.generatedAt)
		if !errors.Is(err, ErrInvalidTime) {
			t.Fatalf("Build() error = %v, want ErrInvalidTime", err)
		}
	}
}

func validAirport() airport.Airport {
	return airport.Airport{
		ICAOCode:  "UBBA",
		Name:      "Heydar Aliyev International Airport",
		Latitude:  40.4675,
		Longitude: 50.0467,
	}
}
