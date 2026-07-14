package datasetprofiler

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestNewRejectsUnsupportedTargetSchema(t *testing.T) {
	_, err := New(Config{
		TargetSchemaVersion: "flight-features-v2",
	})
	if !errors.Is(
		err,
		ErrUnsupportedTargetSchema,
	) {
		t.Fatalf(
			"New() error = %v, want %v",
			err,
			ErrUnsupportedTargetSchema,
		)
	}

	var typedErr *UnsupportedTargetSchemaError
	if !errors.As(err, &typedErr) {
		t.Fatalf(
			"New() error = %T, want *UnsupportedTargetSchemaError",
			err,
		)
	}
}

func TestProfilerReturnsDeterministicEmptyProfile(
	t *testing.T,
) {
	generatedAt := time.Date(
		2026,
		time.July,
		14,
		15,
		0,
		0,
		0,
		time.UTC,
	)
	profiler := newTestProfiler(t, Config{
		Now: func() time.Time {
			return generatedAt
		},
	})

	profile, err := profiler.Profile(
		context.Background(),
		Request{},
	)
	if err != nil {
		t.Fatalf("Profile() error = %v", err)
	}

	if profile.Version != Version ||
		profile.SchemaVersion !=
			flightfeatures.SchemaVersionV1 ||
		!profile.GeneratedAt.Equal(generatedAt) ||
		profile.TotalRecordCount != 0 ||
		profile.AcceptedRecordCount != 0 ||
		profile.RejectedRecordCount != 0 {
		t.Fatalf(
			"unexpected empty profile: %#v",
			profile,
		)
	}
	if len(profile.Groups) != 5 {
		t.Fatalf(
			"group count = %d, want 5",
			len(profile.Groups),
		)
	}

	wantGroups := []struct {
		group      flightfeatures.FeatureGroup
		fieldCount int
	}{
		{flightfeatures.FeatureGroupTemporal, 8},
		{flightfeatures.FeatureGroupGeographical, 11},
		{flightfeatures.FeatureGroupOperational, 11},
		{flightfeatures.FeatureGroupTrajectory, 16},
		{flightfeatures.FeatureGroupAircraft, 6},
	}
	for index, want := range wantGroups {
		if profile.Groups[index].Group != want.group ||
			profile.Groups[index].SchemaFieldCount !=
				want.fieldCount {
			t.Fatalf(
				"group[%d] = %#v, want %q with %d fields",
				index,
				profile.Groups[index],
				want.group,
				want.fieldCount,
			)
		}
	}
}

func TestProfilerAggregatesAcceptedAndRejectedRecords(
	t *testing.T,
) {
	base := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	first := validFeatures(
		"trajectory-1",
		"identity-1",
		"fingerprint-1",
		base,
	)
	first.FlightID = "flight-1"
	first.AircraftID = "aircraft-1"
	first.ICAO24 = "ABC123"
	first.Callsign = "CALL1"
	first.Quality.CompletenessScore = 0.8
	first.Quality.InputQualityScore = 0.7
	first.Quality.SupportingPointCount = 10
	first.Provenance.SourceNames = []string{
		"source-b",
		"source-a",
		"source-a",
	}
	first.Temporal.Evidence.Limitations =
		[]flightfeatures.FeatureLimitation{
			{Code: "shared_limitation"},
		}
	first.Quality.Limitations =
		[]flightfeatures.FeatureLimitation{
			{Code: "quality_limitation"},
			{Code: "shared_limitation"},
		}

	second := validFeatures(
		"trajectory-2",
		"identity-2",
		"fingerprint-2",
		base.Add(time.Hour),
	)
	second.Quality.Status =
		flightfeatures.ValidationStatusLimited
	second.FlightID = "flight-1"
	second.AircraftID = "aircraft-2"
	second.ICAO24 = "DEF456"
	second.Callsign = "CALL2"
	second.Quality.CompletenessScore = 0.6
	second.Quality.InputQualityScore = 0.9
	second.Quality.SupportingPointCount = 20
	second.Provenance.SourceNames = []string{
		"source-a",
	}
	second.Geographical.Evidence.Status =
		flightfeatures.AvailabilityStatusPartial
	second.Geographical.Evidence.AvailableFieldCount = 5
	second.Geographical.Evidence.Limitations =
		[]flightfeatures.FeatureLimitation{
			{Code: "shared_limitation"},
		}
	second.Aircraft.Evidence.Status =
		flightfeatures.AvailabilityStatusUnavailable
	second.Aircraft.Evidence.AvailableFieldCount = 0

	invalid := validFeatures(
		"trajectory-invalid",
		"identity-invalid",
		"fingerprint-invalid",
		base.Add(2*time.Hour),
	)
	invalid.Quality.Status =
		flightfeatures.ValidationStatusInvalid

	unvalidated := validFeatures(
		"trajectory-unvalidated",
		"identity-unvalidated",
		"fingerprint-unvalidated",
		base.Add(3*time.Hour),
	)
	unvalidated.Quality.Status =
		flightfeatures.ValidationStatusUnvalidated

	unsupported := validFeatures(
		"trajectory-unsupported",
		"identity-unsupported",
		"fingerprint-unsupported",
		base.Add(4*time.Hour),
	)
	unsupported.SchemaVersion = "flight-features-v2"

	unknown := validFeatures(
		"trajectory-unknown",
		"identity-unknown",
		"fingerprint-unknown",
		base.Add(5*time.Hour),
	)
	unknown.Quality.Status = "future"

	profiler := newTestProfiler(t, Config{
		Now: func() time.Time {
			return base.Add(24 * time.Hour)
		},
	})
	profile, err := profiler.Profile(
		context.Background(),
		Request{
			Features: []flightfeatures.FlightFeatures{
				first,
				second,
				invalid,
				unvalidated,
				unsupported,
				unknown,
			},
		},
	)
	if err != nil {
		t.Fatalf("Profile() error = %v", err)
	}

	if profile.TotalRecordCount != 6 ||
		profile.AcceptedRecordCount != 2 ||
		profile.RejectedRecordCount != 4 {
		t.Fatalf(
			"unexpected record counts: %#v",
			profile,
		)
	}
	if profile.Validation.ValidCount != 2 ||
		profile.Validation.LimitedCount != 1 ||
		profile.Validation.InvalidCount != 1 ||
		profile.Validation.UnvalidatedCount != 1 ||
		profile.Validation.UnknownCount != 1 {
		t.Fatalf(
			"unexpected validation profile: %#v",
			profile.Validation,
		)
	}

	wantCardinality := CardinalityProfile{
		UniqueTrajectoryCount: 2,
		UniqueIdentityCount:   2,
		UniqueFlightCount:     1,
		UniqueAircraftCount:   2,
		UniqueICAO24Count:     2,
		UniqueCallsignCount:   2,
	}
	if !reflect.DeepEqual(
		profile.Cardinality,
		wantCardinality,
	) {
		t.Fatalf(
			"cardinality = %#v, want %#v",
			profile.Cardinality,
			wantCardinality,
		)
	}

	if !profile.Time.EarliestWindowStart.Equal(
		first.Window.StartTime.UTC(),
	) || !profile.Time.LatestWindowEnd.Equal(
		second.Window.EndTime.UTC(),
	) || !profile.Time.EarliestAsOfTime.Equal(
		first.Window.AsOfTime.UTC(),
	) || !profile.Time.LatestAsOfTime.Equal(
		second.Window.AsOfTime.UTC(),
	) {
		t.Fatalf(
			"unexpected time profile: %#v",
			profile.Time,
		)
	}

	if profile.Quality.CompletenessScore.Count != 2 ||
		profile.Quality.CompletenessScore.Minimum != 0.6 ||
		profile.Quality.CompletenessScore.Maximum != 0.8 ||
		!approximatelyEqual(
			profile.Quality.CompletenessScore.Mean,
			0.7,
			1e-12,
		) ||
		profile.Quality.SupportingPoints.Mean != 15 {
		t.Fatalf(
			"unexpected quality profile: %#v",
			profile.Quality,
		)
	}

	geographical := profile.Groups[1]
	if geographical.Group !=
		flightfeatures.FeatureGroupGeographical ||
		geographical.RecordCount != 2 ||
		geographical.AvailableCount != 1 ||
		geographical.PartialCount != 1 ||
		geographical.UnavailableCount != 0 ||
		!approximatelyEqual(
			geographical.MeanFieldCompleteness,
			8.0/11.0,
			1e-12,
		) {
		t.Fatalf(
			"unexpected geographical profile: %#v",
			geographical,
		)
	}
	aircraft := profile.Groups[4]
	if aircraft.AvailableCount != 1 ||
		aircraft.UnavailableCount != 1 ||
		aircraft.MeanFieldCompleteness != 0.5 {
		t.Fatalf(
			"unexpected aircraft profile: %#v",
			aircraft,
		)
	}

	wantSources := []FrequencyProfile{
		{
			Value:               "source-a",
			OccurrenceCount:     3,
			AffectedRecordCount: 2,
		},
		{
			Value:               "source-b",
			OccurrenceCount:     1,
			AffectedRecordCount: 1,
		},
	}
	if !reflect.DeepEqual(
		profile.Sources,
		wantSources,
	) {
		t.Fatalf(
			"sources = %#v, want %#v",
			profile.Sources,
			wantSources,
		)
	}

	if len(profile.Limitations) != 2 ||
		profile.Limitations[0].Code !=
			"shared_limitation" ||
		profile.Limitations[0].OccurrenceCount != 3 ||
		profile.Limitations[0].AffectedRecordCount != 2 ||
		profile.Limitations[1].Code !=
			"quality_limitation" {
		t.Fatalf(
			"unexpected limitations: %#v",
			profile.Limitations,
		)
	}

	wantRejections := []FrequencyProfile{
		{
			Value:               "unsupported_schema_version",
			OccurrenceCount:     1,
			AffectedRecordCount: 1,
		},
		{
			Value:               "validation_status_invalid",
			OccurrenceCount:     1,
			AffectedRecordCount: 1,
		},
		{
			Value:               "validation_status_unknown",
			OccurrenceCount:     1,
			AffectedRecordCount: 1,
		},
		{
			Value:               "validation_status_unvalidated",
			OccurrenceCount:     1,
			AffectedRecordCount: 1,
		},
	}
	if !reflect.DeepEqual(
		profile.Rejections,
		wantRejections,
	) {
		t.Fatalf(
			"rejections = %#v, want %#v",
			profile.Rejections,
			wantRejections,
		)
	}
}

func TestProfilerDetectsDuplicateAndConflictingSnapshots(
	t *testing.T,
) {
	base := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	first := validFeatures(
		"trajectory-1",
		"identity-1",
		"fingerprint-a",
		base,
	)
	idempotentDuplicate := first.Clone()
	conflictingDuplicate := first.Clone()
	conflictingDuplicate.Provenance.InputFingerprint =
		"fingerprint-b"
	differentAsOf := first.Clone()
	differentAsOf.Window.AsOfTime =
		differentAsOf.Window.AsOfTime.Add(time.Minute)

	profile, err := newTestProfiler(
		t,
		Config{},
	).Profile(
		context.Background(),
		Request{
			Features: []flightfeatures.FlightFeatures{
				first,
				idempotentDuplicate,
				conflictingDuplicate,
				differentAsOf,
			},
		},
	)
	if err != nil {
		t.Fatalf("Profile() error = %v", err)
	}

	if profile.DuplicateSnapshotCount != 2 ||
		profile.ConflictingSnapshotCount != 1 {
		t.Fatalf(
			"unexpected duplicate profile: duplicates=%d conflicts=%d",
			profile.DuplicateSnapshotCount,
			profile.ConflictingSnapshotCount,
		)
	}
}

func TestProfilerTracksInvalidAcceptedQualityNumbers(
	t *testing.T,
) {
	base := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	features := validFeatures(
		"trajectory-1",
		"identity-1",
		"fingerprint-1",
		base,
	)
	features.Quality.CompletenessScore = math.NaN()
	features.Quality.InputQualityScore = 2
	features.Quality.SupportingPointCount = -1

	profile, err := newTestProfiler(
		t,
		Config{},
	).Profile(
		context.Background(),
		Request{
			Features: []flightfeatures.FlightFeatures{
				features,
			},
		},
	)
	if err != nil {
		t.Fatalf("Profile() error = %v", err)
	}

	if profile.Quality.CompletenessScore.Count != 0 ||
		profile.Quality.CompletenessScore.InvalidCount != 1 ||
		profile.Quality.InputQualityScore.InvalidCount != 1 ||
		profile.Quality.SupportingPoints.InvalidCount != 1 {
		t.Fatalf(
			"unexpected invalid numeric profile: %#v",
			profile.Quality,
		)
	}
}

func TestProfilerDoesNotMutateInput(t *testing.T) {
	base := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	features := validFeatures(
		"trajectory-1",
		"identity-1",
		"fingerprint-1",
		base,
	)
	features.Provenance.SourceNames = []string{
		"source-b",
		"source-a",
	}
	features.Quality.Limitations =
		[]flightfeatures.FeatureLimitation{
			{Code: "quality"},
		}
	original := features.Clone()

	if _, err := newTestProfiler(
		t,
		Config{},
	).Profile(
		context.Background(),
		Request{
			Features: []flightfeatures.FlightFeatures{
				features,
			},
		},
	); err != nil {
		t.Fatalf("Profile() error = %v", err)
	}

	if !reflect.DeepEqual(features, original) {
		t.Fatalf(
			"input mutated\nfeatures=%#v\noriginal=%#v",
			features,
			original,
		)
	}
}

func TestProfilerPreservesCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := newTestProfiler(
		t,
		Config{},
	).Profile(
		ctx,
		Request{},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Profile() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestProfileCloneDoesNotShareSlices(t *testing.T) {
	profile := Profile{
		Groups: []GroupProfile{
			{
				Group: flightfeatures.FeatureGroupTemporal,
			},
		},
		Sources: []FrequencyProfile{
			{Value: "source"},
		},
		Limitations: []LimitationProfile{
			{Code: "limitation"},
		},
		Rejections: []FrequencyProfile{
			{Value: "rejection"},
		},
	}

	cloned := profile.Clone()
	cloned.Groups[0].Group =
		flightfeatures.FeatureGroupAircraft
	cloned.Sources[0].Value = "changed"
	cloned.Limitations[0].Code = "changed"
	cloned.Rejections[0].Value = "changed"

	if profile.Groups[0].Group !=
		flightfeatures.FeatureGroupTemporal ||
		profile.Sources[0].Value != "source" ||
		profile.Limitations[0].Code !=
			"limitation" ||
		profile.Rejections[0].Value !=
			"rejection" {
		t.Fatal("Profile.Clone() shared slices")
	}
}

func TestDatasetProfilerVersionRemainsStable(t *testing.T) {
	if Version !=
		"flight-feature-dataset-profiler-v1" {
		t.Fatalf("Version = %q", Version)
	}
}

func newTestProfiler(
	t *testing.T,
	config Config,
) *Profiler {
	t.Helper()

	profiler, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return profiler
}

func validFeatures(
	trajectoryID string,
	identityKey string,
	fingerprint string,
	startTime time.Time,
) flightfeatures.FlightFeatures {
	endTime := startTime.Add(time.Hour)
	asOfTime := endTime.Add(time.Minute)

	return flightfeatures.FlightFeatures{
		SchemaVersion: flightfeatures.SchemaVersionV1,
		TrajectoryID:  trajectoryID,
		IdentityKey:   identityKey,
		Window: flightfeatures.FeatureWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
		ExtractedAt: asOfTime.Add(time.Second),
		Temporal: flightfeatures.TemporalFeatures{
			Evidence: availableEvidence(8, 10),
		},
		Geographical: flightfeatures.GeographicalFeatures{
			Evidence: availableEvidence(11, 10),
		},
		Operational: flightfeatures.OperationalFeatures{
			Evidence: availableEvidence(11, 10),
		},
		Trajectory: flightfeatures.TrajectoryFeatures{
			Evidence: availableEvidence(16, 10),
		},
		Aircraft: flightfeatures.AircraftFeatures{
			Evidence: availableEvidence(6, 0),
		},
		Quality: flightfeatures.FeatureQuality{
			Status:               flightfeatures.ValidationStatusValid,
			CompletenessScore:    1,
			InputQualityScore:    1,
			SupportingPointCount: 10,
		},
		Provenance: flightfeatures.FeatureProvenance{
			ExtractorVersion:    "flight-feature-extractor-v1",
			InputFingerprint:    fingerprint,
			TrajectoryUpdatedAt: endTime,
		},
	}
}

func availableEvidence(
	fieldCount int,
	supportingPointCount int,
) flightfeatures.GroupEvidence {
	return flightfeatures.GroupEvidence{
		Status:               flightfeatures.AvailabilityStatusAvailable,
		AvailableFieldCount:  fieldCount,
		TotalFieldCount:      fieldCount,
		SupportingPointCount: supportingPointCount,
	}
}

func approximatelyEqual(
	left float64,
	right float64,
	tolerance float64,
) bool {
	return math.Abs(left-right) <= tolerance
}
