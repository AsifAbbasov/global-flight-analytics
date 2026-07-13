package passport

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

func TestServiceBuildsAirportPassport(t *testing.T) {
	generatedAt := time.Date(2026, time.July, 13, 14, 0, 0, 0, time.UTC)
	airports := &airportReaderStub{
		result: airport.Airport{
			ICAOCode:  "ubba",
			IATACode:  "gyd",
			Name:      "Heydar Aliyev International Airport",
			City:      "Baku",
			Country:   "Azerbaijan",
			Latitude:  40.4675,
			Longitude: 50.0467,
		},
	}
	analytics := &analyticsReaderStub{
		result: AnalyticsInput{
			Arrivals:       8,
			Departures:     7,
			ActiveAircraft: 5,
			FreshnessScore: 0.9,
			CoverageScore:  0.8,
			ObservedAt:     generatedAt.Add(-time.Minute),
		},
	}

	service, err := NewService(airports, analytics, func() time.Time { return generatedAt })
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.GetByICAO(context.Background(), " ubba ")
	if err != nil {
		t.Fatalf("GetByICAO() error = %v", err)
	}

	if airports.receivedICAO != "UBBA" {
		t.Fatalf("airport reader ICAO = %q, want UBBA", airports.receivedICAO)
	}
	if analytics.receivedICAO != "UBBA" {
		t.Fatalf("analytics reader ICAO = %q, want UBBA", analytics.receivedICAO)
	}
	if result.Identity.ICAOCode != "UBBA" {
		t.Fatalf("passport ICAO = %q, want UBBA", result.Identity.ICAOCode)
	}
	if result.Operations.Activity != 15 {
		t.Fatalf("activity = %d, want 15", result.Operations.Activity)
	}
	if !result.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("generated at = %v, want %v", result.GeneratedAt, generatedAt)
	}
}

func TestServiceRejectsEmptyICAOWithoutCallingReaders(t *testing.T) {
	airports := &airportReaderStub{}
	analytics := &analyticsReaderStub{}
	service, err := NewService(airports, analytics, time.Now)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.GetByICAO(context.Background(), "   ")
	if !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("GetByICAO() error = %v, want ErrInvalidIdentity", err)
	}
	if airports.calls != 0 {
		t.Fatalf("airport reader calls = %d, want 0", airports.calls)
	}
	if analytics.calls != 0 {
		t.Fatalf("analytics reader calls = %d, want 0", analytics.calls)
	}
}

func TestServiceStopsWhenAirportLoadFails(t *testing.T) {
	loadErr := errors.New("airport load failed")
	airports := &airportReaderStub{err: loadErr}
	analytics := &analyticsReaderStub{}
	service, err := NewService(airports, analytics, time.Now)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.GetByICAO(context.Background(), "UBBA")
	if !errors.Is(err, loadErr) {
		t.Fatalf("GetByICAO() error = %v, want wrapped airport error", err)
	}
	if analytics.calls != 0 {
		t.Fatalf("analytics reader calls = %d, want 0", analytics.calls)
	}
}

func TestServicePreservesAnalyticsError(t *testing.T) {
	analyticsErr := errors.New("analytics unavailable")
	airports := &airportReaderStub{result: validAirport()}
	analytics := &analyticsReaderStub{err: analyticsErr}
	service, err := NewService(airports, analytics, time.Now)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.GetByICAO(context.Background(), "UBBA")
	if !errors.Is(err, analyticsErr) {
		t.Fatalf("GetByICAO() error = %v, want wrapped analytics error", err)
	}
}

func TestServicePreservesBuilderValidationError(t *testing.T) {
	generatedAt := time.Date(2026, time.July, 13, 14, 0, 0, 0, time.UTC)
	airports := &airportReaderStub{result: validAirport()}
	analytics := &analyticsReaderStub{
		result: AnalyticsInput{ObservedAt: generatedAt.Add(time.Second)},
	}
	service, err := NewService(airports, analytics, func() time.Time { return generatedAt })
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.GetByICAO(context.Background(), "UBBA")
	if !errors.Is(err, ErrInvalidTime) {
		t.Fatalf("GetByICAO() error = %v, want ErrInvalidTime", err)
	}
}

func TestNewServiceValidatesDependencies(t *testing.T) {
	validAirports := &airportReaderStub{}
	validAnalytics := &analyticsReaderStub{}

	tests := []struct {
		name      string
		airports  AirportReader
		analytics AnalyticsReader
	}{
		{name: "missing airport reader", analytics: validAnalytics},
		{name: "missing analytics reader", airports: validAirports},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewService(test.airports, test.analytics, nil)
			if !errors.Is(err, ErrInvalidServiceConfiguration) {
				t.Fatalf("NewService() error = %v, want ErrInvalidServiceConfiguration", err)
			}
		})
	}
}

type airportReaderStub struct {
	result       airport.Airport
	err          error
	calls        int
	receivedICAO string
}

func (stub *airportReaderStub) GetByICAO(
	_ context.Context,
	icao string,
) (airport.Airport, error) {
	stub.calls++
	stub.receivedICAO = icao
	return stub.result, stub.err
}

type analyticsReaderStub struct {
	result       AnalyticsInput
	err          error
	calls        int
	receivedICAO string
}

func (stub *analyticsReaderStub) GetByICAO(
	_ context.Context,
	icao string,
) (AnalyticsInput, error) {
	stub.calls++
	stub.receivedICAO = icao
	return stub.result, stub.err
}
