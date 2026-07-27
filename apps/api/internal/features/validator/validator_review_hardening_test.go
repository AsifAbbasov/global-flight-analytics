package validator

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestValidatorRejectsNilContext(t *testing.T) {
	validator := newTestValidator(t, Config{})
	_, _, err := validator.Validate(nil, validFeatures())
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrContextRequired)
	}
}

func TestValidatorRejectsNonFiniteValueInPartialGroup(t *testing.T) {
	validator := newTestValidator(t, Config{})
	input := validFeatures()
	input.Operational.Evidence.Status = flightfeatures.AvailabilityStatusPartial
	input.Operational.Evidence.AvailableFieldCount = 9
	input.Operational.Evidence.Limitations = []flightfeatures.FeatureLimitation{
		{
			Code:    flightfeatures.OperationalLimitationVelocityUnavailable,
			Message: "Velocity evidence is unavailable.",
		},
	}
	input.Operational.HeadingChangeDegrees = math.NaN()
	input.Quality.CompletenessScore = float64(48) / 50

	result, report, err := validator.Validate(context.Background(), input)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if result.Quality.Status != flightfeatures.ValidationStatusInvalid || report.ErrorCount == 0 {
		t.Fatalf("partial NaN was not rejected: %#v", report)
	}
	if !hasIssue(report, issueCodePrefix+"negative_or_non_finite_value") {
		t.Fatalf("missing non-finite integrity issue: %#v", report.Issues)
	}
}

func TestValidatorRebuildsQualityLimitationsFromCurrentEvidence(t *testing.T) {
	validator := newTestValidator(t, Config{})
	input := validFeatures()
	input.Operational.Evidence.Status = flightfeatures.AvailabilityStatusPartial
	input.Operational.Evidence.AvailableFieldCount = 9
	input.Operational.Evidence.Limitations = []flightfeatures.FeatureLimitation{
		{
			Code:    flightfeatures.OperationalLimitationVelocityUnavailable,
			Message: "Velocity evidence is unavailable.",
		},
	}
	input.Operational.HeadingChangeDegrees = math.NaN()
	input.Quality.CompletenessScore = float64(48) / 50

	first, firstReport, err := validator.Validate(context.Background(), input)
	if err != nil {
		t.Fatalf("first Validate() error = %v", err)
	}
	if firstReport.Status != flightfeatures.ValidationStatusInvalid {
		t.Fatalf("first status = %q", firstReport.Status)
	}

	corrected := first
	reference := validFeatures()
	corrected.Operational = reference.Operational
	corrected.Quality.CompletenessScore = 1
	corrected.Quality.OptionalCoverageScore = 1
	corrected.Quality.InputQualityScore = reference.Quality.InputQualityScore
	corrected.Quality.SupportingPointCount = reference.Quality.SupportingPointCount

	second, secondReport, err := validator.Validate(context.Background(), corrected)
	if err != nil {
		t.Fatalf("second Validate() error = %v", err)
	}
	if secondReport.Status != flightfeatures.ValidationStatusValid || len(second.Quality.Limitations) != 0 {
		t.Fatalf("stale limitations survived correction: result=%#v report=%#v", second.Quality.Limitations, secondReport)
	}
}

func TestValidatorRejectsResidualPayloadInUnavailableGroup(t *testing.T) {
	validator := newTestValidator(t, Config{})
	input := validFeatures()
	input.Operational.Evidence.Status = flightfeatures.AvailabilityStatusUnavailable
	input.Operational.Evidence.AvailableFieldCount = 0
	input.Operational.Evidence.SupportingPointCount = 0
	input.Operational.Evidence.Limitations = []flightfeatures.FeatureLimitation{
		{Code: "operational_unavailable", Message: "Operational evidence is unavailable."},
	}
	input.Quality.CompletenessScore = float64(39) / 50
	input.Quality.SupportingPointCount = 4

	_, report, err := validator.Validate(context.Background(), input)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if report.Status != flightfeatures.ValidationStatusInvalid ||
		!hasIssue(report, issueCodePrefix+"unavailable_group_payload_not_zero") {
		t.Fatalf("residual unavailable payload was not rejected: %#v", report)
	}
}

func TestValidatorRejectsAvailableOperationalGroupWithoutSupport(t *testing.T) {
	validator := newTestValidator(t, Config{})
	input := validFeatures()
	input.Operational.Evidence.SupportingPointCount = 0

	_, report, err := validator.Validate(context.Background(), input)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if report.Status != flightfeatures.ValidationStatusInvalid ||
		!hasIssue(report, issueCodePrefix+"available_group_support_required") {
		t.Fatalf("zero-support available group was not rejected: %#v", report)
	}
}

func TestValidatorRequiresExplanationForPartialEvidence(t *testing.T) {
	validator := newTestValidator(t, Config{})
	input := validFeatures()
	input.Operational.Evidence.Status = flightfeatures.AvailabilityStatusPartial
	input.Operational.Evidence.AvailableFieldCount = 10
	input.Quality.CompletenessScore = float64(49) / 50

	_, report, err := validator.Validate(context.Background(), input)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if report.Status != flightfeatures.ValidationStatusInvalid ||
		!hasIssue(report, issueCodePrefix+"evidence_limitation_required") {
		t.Fatalf("unexplained partial evidence was not rejected: %#v", report)
	}
}

func TestValidatorRejectsLongitudeSpanMismatch(t *testing.T) {
	validator := newTestValidator(t, Config{})
	input := validFeatures()
	input.Geographical.LongitudeSpanDegrees = 200

	_, report, err := validator.Validate(context.Background(), input)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if report.Status != flightfeatures.ValidationStatusInvalid ||
		!hasIssue(report, issueCodePrefix+"longitude_span_mismatch") {
		t.Fatalf("longitude span mismatch was not rejected: %#v", report)
	}
}
