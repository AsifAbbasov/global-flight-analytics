package statistics

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestCalculatorCalculateBuildsAirportStatistics(t *testing.T) {
	calculator, err := NewCalculator(time.Hour)
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.FixedZone("AZT", 4*60*60))
	windowEnd := windowStart.Add(4 * time.Hour)
	generatedAt := windowEnd
	latestObservationAt := generatedAt.Add(-15 * time.Minute)

	result, err := calculator.Calculate(Input{
		ICAOCode:            " ubBB ",
		WindowStart:         windowStart,
		WindowEnd:           windowEnd,
		Arrivals:            30,
		Departures:          20,
		ActiveAircraft:      12,
		ActiveRoutes:        7,
		ObservedSamples:     80,
		ExpectedSamples:     100,
		LatestObservationAt: latestObservationAt,
		GeneratedAt:         generatedAt,
	})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	if result.ICAOCode != "UBBB" {
		t.Fatalf("ICAOCode = %q, want UBBB", result.ICAOCode)
	}
	if result.TotalMovements != 50 {
		t.Fatalf("TotalMovements = %d, want 50", result.TotalMovements)
	}
	assertFloatEqual(t, result.ArrivalShare, 0.6)
	assertFloatEqual(t, result.DepartureShare, 0.4)
	assertFloatEqual(t, result.MovementsPerHour, 12.5)
	assertFloatEqual(t, result.CoverageScore, 0.8)
	assertFloatEqual(t, result.FreshnessScore, 0.75)
	if result.ActiveRoutes != 7 {
		t.Fatalf("ActiveRoutes = %d, want 7", result.ActiveRoutes)
	}

	if result.WindowStart.Location() != time.UTC || result.WindowEnd.Location() != time.UTC {
		t.Fatal("statistics window must be normalized to UTC")
	}
	if result.LatestObservationAt.Location() != time.UTC || result.GeneratedAt.Location() != time.UTC {
		t.Fatal("statistics timestamps must be normalized to UTC")
	}
}

func TestCalculatorCalculateSupportsZeroMovements(t *testing.T) {
	calculator, err := NewCalculator(time.Hour)
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	result, err := calculator.Calculate(Input{
		ICAOCode:            "UBBB",
		WindowStart:         windowStart,
		WindowEnd:           windowStart.Add(time.Hour),
		ObservedSamples:     1,
		ExpectedSamples:     1,
		LatestObservationAt: windowStart.Add(30 * time.Minute),
		GeneratedAt:         windowStart.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	if result.TotalMovements != 0 {
		t.Fatalf("TotalMovements = %d, want 0", result.TotalMovements)
	}
	assertFloatEqual(t, result.ArrivalShare, 0)
	assertFloatEqual(t, result.DepartureShare, 0)
	assertFloatEqual(t, result.MovementsPerHour, 0)
}

func TestCalculatorCalculateClampsCoverageAtOne(t *testing.T) {
	calculator, err := NewCalculator(time.Hour)
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	result, err := calculator.Calculate(Input{
		ICAOCode:            "UBBB",
		WindowStart:         windowStart,
		WindowEnd:           windowStart.Add(time.Hour),
		ObservedSamples:     120,
		ExpectedSamples:     100,
		LatestObservationAt: windowStart.Add(59 * time.Minute),
		GeneratedAt:         windowStart.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	assertFloatEqual(t, result.CoverageScore, 1)
}

func TestNewCalculatorRejectsInvalidMaximumDataAge(t *testing.T) {
	_, err := NewCalculator(0)
	if !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatalf("NewCalculator() error = %v, want ErrInvalidConfiguration", err)
	}
}

func TestCalculatorCalculateRejectsInvalidIdentity(t *testing.T) {
	calculator, err := NewCalculator(time.Hour)
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	_, err = calculator.Calculate(validInput())
	if !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("Calculate() error = %v, want ErrInvalidIdentity", err)
	}
}

func TestCalculatorCalculateRejectsInvalidWindow(t *testing.T) {
	calculator, err := NewCalculator(time.Hour)
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	input := validInput()
	input.ICAOCode = "UBBB"
	input.WindowEnd = input.WindowStart

	_, err = calculator.Calculate(input)
	if !errors.Is(err, ErrInvalidWindow) {
		t.Fatalf("Calculate() error = %v, want ErrInvalidWindow", err)
	}
}

func TestCalculatorCalculateRejectsInvalidCounters(t *testing.T) {
	calculator, err := NewCalculator(time.Hour)
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	input := validInput()
	input.ICAOCode = "UBBB"
	input.Arrivals = -1

	_, err = calculator.Calculate(input)
	if !errors.Is(err, ErrInvalidCounters) {
		t.Fatalf("Calculate() error = %v, want ErrInvalidCounters", err)
	}
}

func TestCalculatorCalculateRejectsObservationOutsideWindow(t *testing.T) {
	calculator, err := NewCalculator(time.Hour)
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	input := validInput()
	input.ICAOCode = "UBBB"
	input.LatestObservationAt = input.WindowEnd.Add(time.Second)
	input.GeneratedAt = input.LatestObservationAt

	_, err = calculator.Calculate(input)
	if !errors.Is(err, ErrInvalidTime) {
		t.Fatalf("Calculate() error = %v, want ErrInvalidTime", err)
	}
}

func TestCalculatorCalculateRejectsFutureObservation(t *testing.T) {
	calculator, err := NewCalculator(time.Hour)
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	input := validInput()
	input.ICAOCode = "UBBB"
	input.GeneratedAt = input.LatestObservationAt.Add(-time.Second)

	_, err = calculator.Calculate(input)
	if !errors.Is(err, ErrInvalidTime) {
		t.Fatalf("Calculate() error = %v, want ErrInvalidTime", err)
	}
}

func validInput() Input {
	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(time.Hour)

	return Input{
		WindowStart:         windowStart,
		WindowEnd:           windowEnd,
		ObservedSamples:     1,
		ExpectedSamples:     1,
		LatestObservationAt: windowEnd,
		GeneratedAt:         windowEnd,
	}
}

func assertFloatEqual(t *testing.T, actual, expected float64) {
	t.Helper()

	if math.Abs(actual-expected) > 1e-9 {
		t.Fatalf("value = %v, want %v", actual, expected)
	}
}
