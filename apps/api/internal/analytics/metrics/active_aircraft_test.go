package metrics

import "testing"

func TestActiveAircraftCalculate(t *testing.T) {
	metric := ActiveAircraft{}

	if metric.Calculate(127) != 127 {
		t.Fatal("unexpected result")
	}
}

func TestActiveAircraftNegativeValue(t *testing.T) {
	metric := ActiveAircraft{}

	if metric.Calculate(-10) != 0 {
		t.Fatal("negative value must return zero")
	}
}
