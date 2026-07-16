package localtrafficscene

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactionradius"
)

func TestValidateAcceptsBuiltScene(t *testing.T) {
	request := validRequest()
	request.Observations = []ObservationInput{
		validObservation("trajectory:a", "A1B2C3", 40.40, 49.80, 10*time.Second),
		validObservation("trajectory:b", "D4E5F6", 40.45, 49.85, 15*time.Second),
	}
	result, err := Build(request, DefaultPolicy(), interactionradius.DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	report := Validate(result, DefaultPolicy())
	if report.Status != ValidationStatusValid {
		t.Fatalf("Validate() = %+v", report)
	}
}

func TestValidateRejectsTamperedMetrics(t *testing.T) {
	request := validRequest()
	request.Observations = []ObservationInput{
		validObservation("trajectory:a", "A1B2C3", 40.40, 49.80, 10*time.Second),
	}
	result, err := Build(request, DefaultPolicy(), interactionradius.DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	result.Metrics.IncludedAircraftCount++
	report := Validate(result, DefaultPolicy())
	if report.Status != ValidationStatusInvalid {
		t.Fatalf("Validate() = %+v, want invalid", report)
	}
}

func TestValidateRejectsMissingScopeGuard(t *testing.T) {
	request := validRequest()
	request.Observations = []ObservationInput{
		validObservation("trajectory:a", "A1B2C3", 40.40, 49.80, 10*time.Second),
	}
	result, err := Build(request, DefaultPolicy(), interactionradius.DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	result.ScopeGuard = ""
	report := Validate(result, DefaultPolicy())
	if report.Status != ValidationStatusInvalid {
		t.Fatalf("Validate() = %+v, want invalid", report)
	}
}
