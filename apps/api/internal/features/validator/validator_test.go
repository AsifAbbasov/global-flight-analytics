package validator

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestNewRejectsInvalidPolicy(t *testing.T) {
	tests := []struct {
		name    string
		policy  Policy
		wantErr error
	}{
		{
			name: "minimum completeness below zero",
			policy: Policy{
				MinimumValidCompletenessScore: -0.1,
				MinimumValidInputQualityScore: 0.8,
				NumericTolerance:              1e-6,
			},
			wantErr: ErrInvalidMinimumCompleteness,
		},
		{
			name: "minimum input quality above one",
			policy: Policy{
				MinimumValidCompletenessScore: 1,
				MinimumValidInputQualityScore: 1.1,
				NumericTolerance:              1e-6,
			},
			wantErr: ErrInvalidMinimumInputQuality,
		},
		{
			name: "zero tolerance",
			policy: Policy{
				MinimumValidCompletenessScore: 1,
				MinimumValidInputQualityScore: 0.8,
				NumericTolerance:              0,
			},
			wantErr: ErrInvalidNumericTolerance,
		},
		{
			name: "non finite tolerance",
			policy: Policy{
				MinimumValidCompletenessScore: 1,
				MinimumValidInputQualityScore: 0.8,
				NumericTolerance:              math.NaN(),
			},
			wantErr: ErrInvalidNumericTolerance,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(Config{
				Policy: &test.policy,
			})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf(
					"New() error = %v, want %v",
					err,
					test.wantErr,
				)
			}
		})
	}
}

func TestValidatorMarksCompleteFeaturesValid(t *testing.T) {
	validatedAt := time.Date(
		2026,
		time.July,
		14,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	validator := newTestValidator(t, Config{
		Now: func() time.Time {
			return validatedAt
		},
	})
	input := validFeatures()

	result, report, err := validator.Validate(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.Quality.Status !=
		flightfeatures.ValidationStatusValid {
		t.Fatalf(
			"status = %q, want valid",
			result.Quality.Status,
		)
	}
	if report.Status != flightfeatures.ValidationStatusValid ||
		report.ErrorCount != 0 ||
		report.WarningCount != 0 ||
		len(report.Issues) != 0 {
		t.Fatalf("unexpected report: %#v", report)
	}
	if report.ValidatorVersion != Version ||
		!report.ValidatedAt.Equal(validatedAt) {
		t.Fatalf("unexpected report metadata: %#v", report)
	}
	if input.Quality.Status !=
		flightfeatures.ValidationStatusUnvalidated {
		t.Fatal("validator mutated input status")
	}
	if len(result.Quality.Limitations) != 0 {
		t.Fatalf(
			"valid result has limitations: %#v",
			result.Quality.Limitations,
		)
	}
}

func TestValidatorMarksIncompleteFeaturesLimited(t *testing.T) {
	validator := newTestValidator(t, Config{})
	input := validFeatures()
	input.Aircraft = flightfeatures.AircraftFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Status:          flightfeatures.AvailabilityStatusUnavailable,
			TotalFieldCount: 6,
			Limitations: []flightfeatures.FeatureLimitation{
				{
					Code:    "aircraft_metadata_unavailable",
					Message: "Aircraft metadata was not available.",
				},
			},
		},
	}
	input.Quality.CompletenessScore = float64(46) / 52

	result, report, err := validator.Validate(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.Quality.Status !=
		flightfeatures.ValidationStatusLimited {
		t.Fatalf(
			"status = %q, want limited",
			result.Quality.Status,
		)
	}
	if report.ErrorCount != 0 || report.WarningCount < 2 {
		t.Fatalf("unexpected report: %#v", report)
	}
	if !hasIssue(
		report,
		issueCodePrefix+"feature_group_unavailable",
	) {
		t.Fatalf(
			"missing unavailable-group issue: %#v",
			report.Issues,
		)
	}
	if !hasLimitation(
		result.Quality.Limitations,
		"aircraft_metadata_unavailable",
	) {
		t.Fatalf(
			"missing original limitation: %#v",
			result.Quality.Limitations,
		)
	}
}

func TestValidatorMarksContractViolationsInvalid(t *testing.T) {
	validator := newTestValidator(t, Config{})
	input := validFeatures()
	input.SchemaVersion = "future-schema"
	input.ICAO24 = "abc123"
	input.Window.AsOfTime =
		input.Window.EndTime.Add(-time.Second)
	input.Provenance.InputFingerprint = "bad"
	input.Geographical.ObservedPathDistanceKM = 100
	input.Geographical.GreatCircleDistanceKM = 140
	input.Operational.GroundObservationShare = 0.8
	input.Operational.AirborneObservationShare = 0.5
	input.Trajectory.ObservedSegmentCount = 1
	input.Quality.CompletenessScore = 0.5

	result, report, err := validator.Validate(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.Quality.Status !=
		flightfeatures.ValidationStatusInvalid ||
		report.Status != flightfeatures.ValidationStatusInvalid {
		t.Fatalf(
			"unexpected invalid status: result=%q report=%q",
			result.Quality.Status,
			report.Status,
		)
	}
	if report.ErrorCount < 7 {
		t.Fatalf(
			"expected multiple errors, got %#v",
			report,
		)
	}
	expectedCodes := []string{
		issueCodePrefix + "unsupported_schema_version",
		issueCodePrefix + "invalid_icao24",
		issueCodePrefix + "as_of_before_window_end",
		issueCodePrefix + "invalid_input_fingerprint",
		issueCodePrefix + "path_shorter_than_great_circle",
		issueCodePrefix + "observation_shares_do_not_sum_to_one",
		issueCodePrefix + "segment_status_count_mismatch",
		issueCodePrefix + "completeness_score_mismatch",
	}
	for _, code := range expectedCodes {
		if !hasIssue(report, code) {
			t.Fatalf(
				"missing issue %q in %#v",
				code,
				report.Issues,
			)
		}
	}
}

func TestValidatorTreatsPartialRelationshipFailureAsWarning(
	t *testing.T,
) {
	validator := newTestValidator(t, Config{})
	input := validFeatures()
	input.Temporal.Evidence.Status =
		flightfeatures.AvailabilityStatusPartial
	input.Temporal.Evidence.AvailableFieldCount = 7
	input.Temporal.StartHourUTC = 4
	input.Quality.CompletenessScore = float64(51) / 52

	result, report, err := validator.Validate(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.Quality.Status !=
		flightfeatures.ValidationStatusLimited ||
		report.ErrorCount != 0 {
		t.Fatalf("unexpected report: %#v", report)
	}
	if !hasIssue(
		report,
		issueCodePrefix+"start_hour_mismatch",
	) {
		t.Fatalf(
			"missing partial relationship warning: %#v",
			report.Issues,
		)
	}
}

func TestValidatorDoesNotMutateInputOrShareResultSlices(
	t *testing.T,
) {
	validator := newTestValidator(t, Config{})
	input := validFeatures()
	input.Aircraft.Evidence.Status =
		flightfeatures.AvailabilityStatusUnavailable
	input.Aircraft.Evidence.AvailableFieldCount = 0
	input.Aircraft.Registration = ""
	input.Aircraft.Manufacturer = ""
	input.Aircraft.Model = ""
	input.Aircraft.AircraftType = ""
	input.Aircraft.Airline = ""
	input.Aircraft.Country = ""
	input.Quality.CompletenessScore = float64(46) / 52
	original := input.Clone()

	result, report, err := validator.Validate(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !reflect.DeepEqual(input, original) {
		t.Fatal("validator mutated input features")
	}

	result.Quality.Limitations[0].Code = "changed"
	report.Issues[0].Code = "changed"

	if input.Quality.Limitations != nil {
		t.Fatal("result limitations share input storage")
	}
	secondResult, secondReport, err := validator.Validate(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("second Validate() error = %v", err)
	}
	if secondResult.Quality.Limitations[0].Code == "changed" ||
		secondReport.Issues[0].Code == "changed" {
		t.Fatal("validation output shares mutable slices")
	}
}

func TestValidatorIsIdempotentForValidatorLimitations(
	t *testing.T,
) {
	validator := newTestValidator(t, Config{})
	input := validFeatures()
	input.Quality.InputQualityScore = 0.5
	input.Trajectory.TrajectoryQualityScore = 0.5

	first, firstReport, err := validator.Validate(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("first Validate() error = %v", err)
	}
	second, secondReport, err := validator.Validate(
		context.Background(),
		first,
	)
	if err != nil {
		t.Fatalf("second Validate() error = %v", err)
	}

	if !reflect.DeepEqual(
		first.Quality.Limitations,
		second.Quality.Limitations,
	) {
		t.Fatalf(
			"limitations accumulated\nfirst=%#v\nsecond=%#v",
			first.Quality.Limitations,
			second.Quality.Limitations,
		)
	}
	if !reflect.DeepEqual(
		firstReport.Issues,
		secondReport.Issues,
	) {
		t.Fatalf(
			"reports differ\nfirst=%#v\nsecond=%#v",
			firstReport.Issues,
			secondReport.Issues,
		)
	}
}

func TestValidatorHonorsCustomThresholdPolicy(t *testing.T) {
	policy := DefaultPolicy()
	policy.MinimumValidCompletenessScore = 0.9
	policy.MinimumValidInputQualityScore = 0.6
	validator := newTestValidator(t, Config{
		Policy: &policy,
	})
	input := validFeatures()
	input.Aircraft.Evidence.Status =
		flightfeatures.AvailabilityStatusPartial
	input.Aircraft.Evidence.AvailableFieldCount = 4
	input.Aircraft.Airline = ""
	input.Aircraft.Country = ""
	input.Quality.CompletenessScore = float64(50) / 52
	input.Quality.InputQualityScore = 0.7
	input.Trajectory.TrajectoryQualityScore = 0.7

	result, report, err := validator.Validate(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.Quality.Status !=
		flightfeatures.ValidationStatusLimited {
		t.Fatalf("status = %q, want limited", result.Quality.Status)
	}
	if hasIssue(
		report,
		issueCodePrefix+"completeness_below_valid_threshold",
	) {
		t.Fatal("custom completeness threshold was not honored")
	}
	if hasIssue(
		report,
		issueCodePrefix+"input_quality_below_valid_threshold",
	) {
		t.Fatal("custom input-quality threshold was not honored")
	}
}

func TestValidatorPreservesContextCancellation(t *testing.T) {
	validator := newTestValidator(t, Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := validator.Validate(ctx, validFeatures())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Validate() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestReportCloneDoesNotShareIssues(t *testing.T) {
	report := Report{
		Issues: []Issue{
			{
				Code: "one",
			},
		},
	}
	cloned := report.Clone()
	cloned.Issues[0].Code = "changed"

	if report.Issues[0].Code != "one" {
		t.Fatal("Report.Clone() shares issue storage")
	}
}

func newTestValidator(
	t *testing.T,
	config Config,
) *Validator {
	t.Helper()

	if config.Now == nil {
		config.Now = func() time.Time {
			return time.Date(
				2026,
				time.July,
				14,
				12,
				0,
				0,
				0,
				time.UTC,
			)
		}
	}

	validator, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return validator
}

func validFeatures() flightfeatures.FlightFeatures {
	start := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	end := start.Add(time.Hour)
	greatCircleDistance := 140.0
	observedPathDistance := 150.0

	return flightfeatures.FlightFeatures{
		SchemaVersion: flightfeatures.SchemaVersionV1,
		TrajectoryID:  "trajectory-one",
		IdentityKey:   "flight-identity-example",
		FlightID:      "flight-one",
		AircraftID:    "aircraft-one",
		ICAO24:        "ABC123",
		Callsign:      "TEST123",
		Window: flightfeatures.FeatureWindow{
			StartTime: start,
			EndTime:   end,
			AsOfTime:  end,
		},
		ExtractedAt: end.Add(time.Minute),
		Temporal: flightfeatures.TemporalFeatures{
			Evidence:            availableEvidence(8, 4),
			DurationSeconds:     3600,
			StartHourUTC:        8,
			EndHourUTC:          9,
			StartWeekday:        start.Weekday(),
			EndWeekday:          end.Weekday(),
			StartMinuteOfDayUTC: 480,
			EndMinuteOfDayUTC:   540,
			CrossesUTCMidnight:  false,
		},
		Geographical: flightfeatures.GeographicalFeatures{
			Evidence:                  availableEvidence(11, 4),
			StartLatitude:             40,
			StartLongitude:            50,
			EndLatitude:               41,
			EndLongitude:              51,
			MinimumLatitude:           40,
			MaximumLatitude:           41,
			MinimumLongitude:          50,
			MaximumLongitude:          51,
			LatitudeSpanDegrees:       1,
			LongitudeSpanDegrees:      1,
			GreatCircleDistanceKM:     greatCircleDistance,
			ObservedPathDistanceKM:    observedPathDistance,
			MaximumDisplacementKM:     greatCircleDistance,
			CrossesAntimeridian:       false,
			UniqueGeographicCellCount: 3,
			GeographicCellPrecision:   3,
		},
		Operational: flightfeatures.OperationalFeatures{
			Evidence:                       availableEvidence(11, 4),
			MinimumAltitudeM:               1000,
			MaximumAltitudeM:               11000,
			MeanAltitudeM:                  6000,
			AltitudeRangeM:                 10000,
			MeanVelocityMPS:                210,
			MaximumVelocityMPS:             250,
			MeanAbsoluteVerticalRateMPS:    3,
			MaximumAbsoluteVerticalRateMPS: 8,
			HeadingChangeDegrees:           25,
			GroundObservationShare:         0.25,
			AirborneObservationShare:       0.75,
		},
		Trajectory: flightfeatures.TrajectoryFeatures{
			Evidence:                    availableEvidence(16, 4),
			PointCount:                  4,
			SegmentCount:                2,
			CoverageGapCount:            0,
			TrajectoryQualityScore:      0.9,
			ObservedSegmentCount:        2,
			InterpolatedSegmentCount:    0,
			EstimatedSegmentCount:       0,
			InvalidSegmentCount:         0,
			ObservedSegmentShare:        1,
			InterpolatedSegmentShare:    0,
			EstimatedSegmentShare:       0,
			InvalidSegmentShare:         0,
			MeanSamplingIntervalSeconds: 1200,
			MaximumSamplingGapSeconds:   1200,
			CoverageRatio:               1,
			PathEfficiencyRatio: greatCircleDistance /
				observedPathDistance,
		},
		Aircraft: flightfeatures.AircraftFeatures{
			Evidence:     availableEvidence(6, 0),
			Registration: "4K-AZ01",
			Manufacturer: "Example Manufacturer",
			Model:        "Example Model",
			AircraftType: "Example Type",
			Airline:      "Example Airline",
			Country:      "Azerbaijan",
		},
		Quality: flightfeatures.FeatureQuality{
			Status:               flightfeatures.ValidationStatusUnvalidated,
			CompletenessScore:    1,
			InputQualityScore:    0.9,
			SupportingPointCount: 4,
		},
		Provenance: flightfeatures.FeatureProvenance{
			ExtractorVersion:    "flight-feature-extractor-v1",
			InputFingerprint:    "sha256:" + strings.Repeat("a", 64),
			TrajectoryUpdatedAt: end,
			SourceNames: []string{
				"airplanes.live",
				"open-sky",
			},
		},
	}
}

func availableEvidence(
	totalFieldCount int,
	supportingPointCount int,
) flightfeatures.GroupEvidence {
	return flightfeatures.GroupEvidence{
		Status:               flightfeatures.AvailabilityStatusAvailable,
		AvailableFieldCount:  totalFieldCount,
		TotalFieldCount:      totalFieldCount,
		SupportingPointCount: supportingPointCount,
	}
}

func hasIssue(report Report, code string) bool {
	for _, issue := range report.Issues {
		if issue.Code == code {
			return true
		}
	}

	return false
}

func hasLimitation(
	limitations []flightfeatures.FeatureLimitation,
	code string,
) bool {
	for _, limitation := range limitations {
		if limitation.Code == code {
			return true
		}
	}

	return false
}
