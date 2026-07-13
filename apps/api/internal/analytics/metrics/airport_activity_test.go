package metrics

import "testing"

func TestAirportActivityCalculate(t *testing.T) {
	metric := AirportActivity{}

	got := metric.Calculate(135, 128)

	if got != 263 {
		t.Fatalf("expected 263, got %d", got)
	}
}

func TestAirportActivityNegativeValues(t *testing.T) {
	metric := AirportActivity{}

	got := metric.Calculate(-5, 10)

	if got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
}
