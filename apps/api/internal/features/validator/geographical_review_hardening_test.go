package validator

import (
	"context"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestValidatorAcceptsWrappedEnvelopeWithoutPathCrossing(t *testing.T) {
	input := validFeatures()
	input.Geographical.MinimumLongitude = 160
	input.Geographical.MaximumLongitude = 0
	input.Geographical.LongitudeSpanDegrees = 200
	input.Geographical.CrossesAntimeridian = false

	result, report, err := newTestValidator(t, Config{}).Validate(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if result.Quality.Status != flightfeatures.ValidationStatusValid ||
		report.ErrorCount != 0 || report.WarningCount != 0 {
		t.Fatalf("wrapped envelope was confused with path crossing: %#v", report)
	}
	if hasIssue(report, issueCodePrefix+"longitude_bounds_reversed") {
		t.Fatalf("obsolete reversed-bound issue reported: %#v", report)
	}
}

func TestValidatorAllowsDisconnectedSegmentFallbackDistances(t *testing.T) {
	input := validFeatures()
	limitation := flightfeatures.FeatureLimitation{
		Code:    flightfeatures.GeographicalLimitationSegmentEndpointFallback,
		Message: "Geographical features were reconstructed from ordered non-invalid trajectory segment endpoints; observed path distance includes only movement inside each usable segment and excludes discontinuities between segments.",
	}
	input.Geographical.Evidence.Limitations = []flightfeatures.FeatureLimitation{limitation}
	input.Geographical.GreatCircleDistanceKM = 500
	input.Geographical.ObservedPathDistanceKM = 100
	input.Geographical.MaximumDisplacementKM = 400
	input.Trajectory.PathEfficiencyRatio = 0.5
	input.Quality.Limitations = []flightfeatures.FeatureLimitation{limitation}

	result, report, err := newTestValidator(t, Config{}).Validate(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if result.Quality.Status != flightfeatures.ValidationStatusLimited ||
		report.ErrorCount != 0 {
		t.Fatalf("disconnected fallback was rejected: %#v", report)
	}
	for _, code := range []string{
		issueCodePrefix + "path_shorter_than_great_circle",
		issueCodePrefix + "displacement_exceeds_path",
		issueCodePrefix + "path_efficiency_mismatch",
	} {
		if hasIssue(report, code) {
			t.Fatalf("fallback reported obsolete issue %q: %#v", code, report)
		}
	}
}
