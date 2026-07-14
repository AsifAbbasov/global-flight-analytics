package geographicalbuilder

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestNewRejectsInvalidCellPrecision(t *testing.T) {
	for _, precision := range []int{-1, 7} {
		_, err := New(Config{
			GeographicCellPrecision: precision,
		})
		if !errors.Is(
			err,
			ErrInvalidGeographicCellPrecision,
		) {
			t.Fatalf(
				"New(%d) error = %v, want %v",
				precision,
				err,
				ErrInvalidGeographicCellPrecision,
			)
		}
	}
}

func TestBuilderBuildsGeographicalFeaturesFromPoints(
	t *testing.T,
) {
	builder := newTestBuilder(t, Config{})
	item := trajectory.FlightTrajectory{
		PointCount: 4,
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:  40.0,
				Longitude: 49.0,
			},
			{
				Latitude:  40.5,
				Longitude: 49.5,
			},
			{
				Latitude:  41.0,
				Longitude: 50.0,
			},
			{
				Latitude:  41.0,
				Longitude: 50.0,
			},
		},
		Segments: []trajectory.TrajectorySegment{
			{
				SequenceNumber: 1,
				Status:         trajectory.SegmentStatusObserved,
				StartLatitude:  0,
				StartLongitude: 0,
				EndLatitude:    1,
				EndLongitude:   1,
			},
		},
	}

	features, err := builder.Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.Evidence.Status !=
		flightfeatures.AvailabilityStatusAvailable ||
		features.Evidence.AvailableFieldCount !=
			GeographicalFeatureFieldCount ||
		features.Evidence.TotalFieldCount !=
			GeographicalFeatureFieldCount ||
		features.Evidence.SupportingPointCount != 4 {
		t.Fatalf(
			"unexpected evidence: %#v",
			features.Evidence,
		)
	}
	if features.StartLatitude != 40 ||
		features.StartLongitude != 49 ||
		features.EndLatitude != 41 ||
		features.EndLongitude != 50 {
		t.Fatalf(
			"unexpected endpoints: %#v",
			features,
		)
	}
	if features.MinimumLatitude != 40 ||
		features.MaximumLatitude != 41 ||
		features.MinimumLongitude != 49 ||
		features.MaximumLongitude != 50 ||
		features.LatitudeSpanDegrees != 1 ||
		features.LongitudeSpanDegrees != 1 {
		t.Fatalf(
			"unexpected bounds: %#v",
			features,
		)
	}
	if features.GreatCircleDistanceKM <= 0 ||
		features.ObservedPathDistanceKM <
			features.GreatCircleDistanceKM ||
		features.MaximumDisplacementKM !=
			features.GreatCircleDistanceKM {
		t.Fatalf(
			"unexpected distances: %#v",
			features,
		)
	}
	if features.CrossesAntimeridian {
		t.Fatal("unexpected antimeridian crossing")
	}
	if features.UniqueGeographicCellCount != 3 ||
		features.GeographicCellPrecision !=
			DefaultGeographicCellPrecision {
		t.Fatalf(
			"unexpected cell features: %#v",
			features,
		)
	}
	if hasLimitation(
		features.Evidence.Limitations,
		"geographical_segment_endpoint_fallback",
	) {
		t.Fatal(
			"segment fallback must not be used when point evidence exists",
		)
	}
}

func TestBuilderExcludesInvalidPointCoordinates(
	t *testing.T,
) {
	builder := newTestBuilder(t, Config{})
	item := trajectory.FlightTrajectory{
		PointCount: 5,
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:  40,
				Longitude: 49,
			},
			{
				Latitude:  math.NaN(),
				Longitude: 49.5,
			},
			{
				Latitude:  91,
				Longitude: 50,
			},
			{
				Latitude:  40.5,
				Longitude: 181,
			},
			{
				Latitude:  41,
				Longitude: 50,
			},
		},
	}

	features, err := builder.Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.Evidence.SupportingPointCount != 2 {
		t.Fatalf(
			"supporting point count = %d, want 2",
			features.Evidence.SupportingPointCount,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"geographical_invalid_point_coordinates",
	) {
		t.Fatalf(
			"missing invalid-coordinate limitation: %#v",
			features.Evidence.Limitations,
		)
	}
	if features.StartLatitude != 40 ||
		features.EndLatitude != 41 {
		t.Fatalf(
			"invalid point changed endpoints: %#v",
			features,
		)
	}
}

func TestBuilderFallsBackToOrderedSegmentEndpoints(
	t *testing.T,
) {
	builder := newTestBuilder(t, Config{})
	item := trajectory.FlightTrajectory{
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:  100,
				Longitude: 0,
			},
		},
		Segments: []trajectory.TrajectorySegment{
			{
				ID:             "second",
				SequenceNumber: 2,
				Status:         trajectory.SegmentStatusObserved,
				StartLatitude:  41,
				StartLongitude: 50,
				EndLatitude:    42,
				EndLongitude:   51,
			},
			{
				ID:             "invalid",
				SequenceNumber: 3,
				Status:         trajectory.SegmentStatusInvalid,
				StartLatitude:  42,
				StartLongitude: 51,
				EndLatitude:    43,
				EndLongitude:   52,
			},
			{
				ID:             "first",
				SequenceNumber: 1,
				Status:         trajectory.SegmentStatusObserved,
				StartLatitude:  40,
				StartLongitude: 49,
				EndLatitude:    41,
				EndLongitude:   50,
			},
		},
	}

	features, err := builder.Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.StartLatitude != 40 ||
		features.StartLongitude != 49 ||
		features.EndLatitude != 42 ||
		features.EndLongitude != 51 {
		t.Fatalf(
			"unexpected fallback endpoints: %#v",
			features,
		)
	}
	if features.Evidence.SupportingPointCount != 3 {
		t.Fatalf(
			"fallback supporting count = %d, want 3",
			features.Evidence.SupportingPointCount,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"geographical_point_evidence_unusable",
	) || !hasLimitation(
		features.Evidence.Limitations,
		"geographical_segment_endpoint_fallback",
	) {
		t.Fatalf(
			"missing fallback limitations: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderReturnsUnavailableWhenCoordinatesAreAbsent(
	t *testing.T,
) {
	builder := newTestBuilder(t, Config{})

	features, err := builder.Build(
		context.Background(),
		trajectory.FlightTrajectory{},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.Evidence.Status !=
		flightfeatures.AvailabilityStatusUnavailable ||
		features.Evidence.AvailableFieldCount != 0 ||
		features.Evidence.TotalFieldCount !=
			GeographicalFeatureFieldCount ||
		features.Evidence.SupportingPointCount != 0 {
		t.Fatalf(
			"unexpected unavailable evidence: %#v",
			features.Evidence,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"geographical_point_evidence_unavailable",
	) || !hasLimitation(
		features.Evidence.Limitations,
		"geographical_coordinates_unavailable",
	) {
		t.Fatalf(
			"missing unavailable limitations: %#v",
			features.Evidence.Limitations,
		)
	}
	if features.GeographicCellPrecision !=
		DefaultGeographicCellPrecision {
		t.Fatalf(
			"cell precision = %d",
			features.GeographicCellPrecision,
		)
	}
}

func TestBuilderHandlesSingleCoordinate(t *testing.T) {
	builder := newTestBuilder(t, Config{})
	item := trajectory.FlightTrajectory{
		PointCount: 1,
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:  40,
				Longitude: 180,
			},
		},
	}

	features, err := builder.Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.StartLongitude != -180 ||
		features.EndLongitude != -180 ||
		features.MinimumLongitude != -180 ||
		features.MaximumLongitude != -180 {
		t.Fatalf(
			"longitude normalization failed: %#v",
			features,
		)
	}
	if features.GreatCircleDistanceKM != 0 ||
		features.ObservedPathDistanceKM != 0 ||
		features.MaximumDisplacementKM != 0 ||
		features.LatitudeSpanDegrees != 0 ||
		features.LongitudeSpanDegrees != 0 ||
		features.CrossesAntimeridian {
		t.Fatalf(
			"unexpected single-coordinate geometry: %#v",
			features,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"geographical_single_coordinate",
	) {
		t.Fatalf(
			"missing single-coordinate limitation: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderUsesShortestLongitudeSpanAcrossAntimeridian(
	t *testing.T,
) {
	builder := newTestBuilder(t, Config{})
	item := trajectory.FlightTrajectory{
		PointCount: 3,
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:  10,
				Longitude: 170,
			},
			{
				Latitude:  10,
				Longitude: 179,
			},
			{
				Latitude:  10,
				Longitude: -170,
			},
		},
	}

	features, err := builder.Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !features.CrossesAntimeridian {
		t.Fatal("expected antimeridian crossing")
	}
	if !approximatelyEqual(
		features.LongitudeSpanDegrees,
		20,
		1e-12,
	) {
		t.Fatalf(
			"longitude span = %v, want 20",
			features.LongitudeSpanDegrees,
		)
	}
	if features.MinimumLongitude != 170 ||
		features.MaximumLongitude != -170 {
		t.Fatalf(
			"circular longitude bounds = %v to %v",
			features.MinimumLongitude,
			features.MaximumLongitude,
		)
	}
	if features.ObservedPathDistanceKM >
		3000 {
		t.Fatalf(
			"path distance indicates a long-way-around error: %v",
			features.ObservedPathDistanceKM,
		)
	}
}

func TestBuilderHonorsCustomCellPrecision(t *testing.T) {
	builder := newTestBuilder(t, Config{
		GeographicCellPrecision: 3,
	})
	item := trajectory.FlightTrajectory{
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:  40.0001,
				Longitude: 49.0001,
			},
			{
				Latitude:  40.0004,
				Longitude: 49.0004,
			},
			{
				Latitude:  40.0006,
				Longitude: 49.0006,
			},
		},
	}

	features, err := builder.Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.GeographicCellPrecision != 3 ||
		features.UniqueGeographicCellCount != 2 {
		t.Fatalf(
			"unexpected custom cell features: %#v",
			features,
		)
	}
}

func TestBuilderReportsPointCountMetadataMismatch(
	t *testing.T,
) {
	builder := newTestBuilder(t, Config{})
	item := trajectory.FlightTrajectory{
		PointCount: 5,
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:  40,
				Longitude: 49,
			},
			{
				Latitude:  41,
				Longitude: 50,
			},
		},
	}

	features, err := builder.Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !hasLimitation(
		features.Evidence.Limitations,
		"trajectory_point_count_metadata_mismatch",
	) {
		t.Fatalf(
			"missing point-count mismatch limitation: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderDoesNotMutateInput(t *testing.T) {
	builder := newTestBuilder(t, Config{})
	item := trajectory.FlightTrajectory{
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:  40,
				Longitude: 49,
			},
		},
		Segments: []trajectory.TrajectorySegment{
			{
				ID:             "segment-two",
				SequenceNumber: 2,
				Status:         trajectory.SegmentStatusObserved,
				StartLatitude:  41,
				StartLongitude: 50,
				EndLatitude:    42,
				EndLongitude:   51,
			},
			{
				ID:             "segment-one",
				SequenceNumber: 1,
				Status:         trajectory.SegmentStatusObserved,
				StartLatitude:  40,
				StartLongitude: 49,
				EndLatitude:    41,
				EndLongitude:   50,
			},
		},
	}
	original := item
	original.Points = append(
		[]trajectory.TrackPoint4D(nil),
		item.Points...,
	)
	original.Segments = append(
		[]trajectory.TrajectorySegment(nil),
		item.Segments...,
	)

	if _, err := builder.Build(
		context.Background(),
		item,
	); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !reflect.DeepEqual(item, original) {
		t.Fatalf(
			"input mutated\ninput=%#v\noriginal=%#v",
			item,
			original,
		)
	}
}

func TestBuilderPreservesCanceledContext(t *testing.T) {
	builder := newTestBuilder(t, Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := builder.Build(
		ctx,
		trajectory.FlightTrajectory{},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Build() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestCloneFeaturesDoesNotShareLimitations(t *testing.T) {
	features := flightfeatures.GeographicalFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Limitations: []flightfeatures.FeatureLimitation{
				{
					Code: "original",
				},
			},
		},
	}

	cloned := cloneFeatures(features)
	cloned.Evidence.Limitations[0].Code = "changed"

	if features.Evidence.Limitations[0].Code != "original" {
		t.Fatal("cloneFeatures() shared limitations")
	}
}

func TestGeographicalBuilderContractConstantsRemainStable(
	t *testing.T,
) {
	if Version != "geographical-feature-builder-v1" {
		t.Fatalf("Version = %q", Version)
	}
	if GeographicalFeatureFieldCount != 11 {
		t.Fatalf(
			"GeographicalFeatureFieldCount = %d",
			GeographicalFeatureFieldCount,
		)
	}
	if DefaultGeographicCellPrecision != 2 {
		t.Fatalf(
			"DefaultGeographicCellPrecision = %d",
			DefaultGeographicCellPrecision,
		)
	}
}

func newTestBuilder(
	t *testing.T,
	config Config,
) *Builder {
	t.Helper()

	builder, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return builder
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
