package validator

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestValidatorAcceptsSystemUpdateAfterAsOfTime(t *testing.T) {
	validator := newTestValidator(t, Config{})
	features := validFeatures()
	features.Provenance.TrajectoryUpdatedAt =
		features.Window.AsOfTime.Add(30 * time.Second)
	features.ExtractedAt = features.Window.AsOfTime.Add(time.Minute)

	result, report, err := validator.Validate(context.Background(), features)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if result.Quality.Status != flightfeatures.ValidationStatusValid ||
		report.ErrorCount != 0 || report.WarningCount != 0 {
		t.Fatalf("system provenance was confused with event time: %#v", report)
	}
}

func TestValidatorRequiresAircraftProvenanceWhenMetadataIsAvailable(t *testing.T) {
	validator := newTestValidator(t, Config{})
	features := validFeatures()
	features.Provenance.AircraftMetadataSourceName = ""
	features.Provenance.AircraftMetadataProviderVersion = ""
	features.Provenance.AircraftMetadataRetrievedAt = time.Time{}

	result, report, err := validator.Validate(context.Background(), features)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if result.Quality.Status != flightfeatures.ValidationStatusInvalid ||
		!hasIssue(report, issueCodePrefix+"aircraft_metadata_provenance_required") {
		t.Fatalf("missing aircraft provenance was not rejected: %#v", report)
	}
}

func TestValidatorReportsUnavailableTrajectoryRecordTimestamps(t *testing.T) {
	validator := newTestValidator(t, Config{})
	features := validFeatures()
	features.Provenance.TrajectoryCreatedAt = time.Time{}
	features.Provenance.TrajectoryUpdatedAt = time.Time{}

	result, report, err := validator.Validate(context.Background(), features)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if result.Quality.Status != flightfeatures.ValidationStatusLimited ||
		!hasIssue(report, issueCodePrefix+"trajectory_record_timestamps_unavailable") {
		t.Fatalf("missing record provenance was not surfaced: %#v", report)
	}
}
