package trajectorybuilder

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestBuilderBuildsCompleteTrajectoryFeatures(t *testing.T) {
	startTime := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(time.Minute)
	item := trajectory.FlightTrajectory{
		StartTime:        startTime,
		EndTime:          endTime,
		PointCount:       4,
		SegmentCount:     4,
		CoverageGapCount: 1,
		QualityScore:     0.8,
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:   0,
				Longitude:  0,
				ObservedAt: startTime,
			},
			{
				Latitude:   0,
				Longitude:  1,
				ObservedAt: startTime.Add(10 * time.Second),
			},
			{
				Latitude:   0,
				Longitude:  2,
				ObservedAt: startTime.Add(30 * time.Second),
			},
			{
				Latitude:   0,
				Longitude:  3,
				ObservedAt: endTime,
			},
		},
		Segments: []trajectory.TrajectorySegment{
			{Status: trajectory.SegmentStatusObserved},
			{Status: trajectory.SegmentStatusInterpolated},
			{Status: trajectory.SegmentStatusEstimated},
			{Status: trajectory.SegmentStatusInvalid},
		},
		CoverageGaps: []trajectory.CoverageGap{
			{
				ID:              "gap-1",
				StartTime:       startTime.Add(20 * time.Second),
				EndTime:         startTime.Add(40 * time.Second),
				DurationSeconds: 20,
			},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.Evidence.Status !=
		flightfeatures.AvailabilityStatusAvailable ||
		features.Evidence.AvailableFieldCount !=
			TrajectoryFeatureFieldCount ||
		features.Evidence.TotalFieldCount !=
			TrajectoryFeatureFieldCount ||
		features.Evidence.SupportingPointCount != 4 {
		t.Fatalf(
			"unexpected evidence: %#v",
			features.Evidence,
		)
	}
	if features.PointCount != 4 ||
		features.SegmentCount != 4 ||
		features.CoverageGapCount != 1 ||
		features.TrajectoryQualityScore != 0.8 {
		t.Fatalf(
			"unexpected base features: %#v",
			features,
		)
	}
	if features.ObservedSegmentCount != 1 ||
		features.InterpolatedSegmentCount != 1 ||
		features.EstimatedSegmentCount != 1 ||
		features.InvalidSegmentCount != 1 {
		t.Fatalf(
			"unexpected segment counts: %#v",
			features,
		)
	}
	if features.ObservedSegmentShare != 0.25 ||
		features.InterpolatedSegmentShare != 0.25 ||
		features.EstimatedSegmentShare != 0.25 ||
		features.InvalidSegmentShare != 0.25 {
		t.Fatalf(
			"unexpected segment shares: %#v",
			features,
		)
	}
	if features.MeanSamplingIntervalSeconds != 20 ||
		features.MaximumSamplingGapSeconds != 30 {
		t.Fatalf(
			"unexpected sampling features: %#v",
			features,
		)
	}
	if !approximatelyEqual(
		features.CoverageRatio,
		2.0/3.0,
		1e-12,
	) {
		t.Fatalf(
			"coverage ratio = %v, want %v",
			features.CoverageRatio,
			2.0/3.0,
		)
	}
	if !approximatelyEqual(
		features.PathEfficiencyRatio,
		1,
		1e-12,
	) {
		t.Fatalf(
			"path efficiency = %v, want 1",
			features.PathEfficiencyRatio,
		)
	}
	if len(features.Evidence.Limitations) != 0 {
		t.Fatalf(
			"unexpected limitations: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderUsesActualCollectionLengths(t *testing.T) {
	startTime := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	item := trajectory.FlightTrajectory{
		StartTime:        startTime,
		EndTime:          startTime.Add(time.Minute),
		PointCount:       99,
		SegmentCount:     98,
		CoverageGapCount: 97,
		QualityScore:     0.5,
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:   0,
				Longitude:  0,
				ObservedAt: startTime,
			},
			{
				Latitude:   0,
				Longitude:  1,
				ObservedAt: startTime.Add(time.Minute),
			},
		},
		Segments: []trajectory.TrajectorySegment{
			{
				ID:     "unknown",
				Status: "future_status",
			},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.PointCount != 2 ||
		features.SegmentCount != 1 ||
		features.CoverageGapCount != 0 {
		t.Fatalf(
			"collection lengths were not authoritative: %#v",
			features,
		)
	}
	if features.InvalidSegmentCount != 1 ||
		features.InvalidSegmentShare != 1 {
		t.Fatalf(
			"unknown segment was not classified as invalid: %#v",
			features,
		)
	}

	for _, code := range []string{
		"trajectory_point_count_metadata_mismatch",
		"trajectory_segment_count_metadata_mismatch",
		"trajectory_coverage_gap_count_metadata_mismatch",
		"trajectory_segment_status_unknown",
	} {
		if !hasLimitation(
			features.Evidence.Limitations,
			code,
		) {
			t.Fatalf(
				"missing limitation %q in %#v",
				code,
				features.Evidence.Limitations,
			)
		}
	}
}

func TestBuilderReturnsPartialEvidenceForEmptyTrajectory(
	t *testing.T,
) {
	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.Evidence.Status !=
		flightfeatures.AvailabilityStatusPartial ||
		features.Evidence.AvailableFieldCount != 12 ||
		features.Evidence.TotalFieldCount !=
			TrajectoryFeatureFieldCount ||
		features.Evidence.SupportingPointCount != 0 {
		t.Fatalf(
			"unexpected empty evidence: %#v",
			features.Evidence,
		)
	}
	if features.PointCount != 0 ||
		features.SegmentCount != 0 ||
		features.CoverageGapCount != 0 ||
		features.TrajectoryQualityScore != 0 {
		t.Fatalf(
			"unexpected empty features: %#v",
			features,
		)
	}
	if features.ObservedSegmentShare != 0 ||
		features.InterpolatedSegmentShare != 0 ||
		features.EstimatedSegmentShare != 0 ||
		features.InvalidSegmentShare != 0 {
		t.Fatalf(
			"empty segment shares must be zero: %#v",
			features,
		)
	}

	for _, code := range []string{
		"trajectory_sampling_evidence_insufficient",
		"trajectory_coverage_window_unavailable",
		"trajectory_path_point_evidence_unavailable",
		"trajectory_path_coordinates_unavailable",
		"trajectory_path_efficiency_evidence_insufficient",
	} {
		if !hasLimitation(
			features.Evidence.Limitations,
			code,
		) {
			t.Fatalf(
				"missing limitation %q in %#v",
				code,
				features.Evidence.Limitations,
			)
		}
	}
}

func TestBuilderRejectsInvalidQualityScoreAsUnavailableField(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	item := trajectory.FlightTrajectory{
		StartTime:    startTime,
		EndTime:      startTime.Add(time.Minute),
		QualityScore: math.NaN(),
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:   0,
				Longitude:  0,
				ObservedAt: startTime,
			},
			{
				Latitude:   0,
				Longitude:  1,
				ObservedAt: startTime.Add(time.Minute),
			},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.Evidence.Status !=
		flightfeatures.AvailabilityStatusPartial ||
		features.Evidence.AvailableFieldCount != 15 {
		t.Fatalf(
			"unexpected invalid-quality evidence: %#v",
			features.Evidence,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"trajectory_quality_score_invalid",
	) {
		t.Fatalf(
			"missing quality limitation: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderDoesNotMutateInput(t *testing.T) {
	startTime := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	item := trajectory.FlightTrajectory{
		StartTime:    startTime,
		EndTime:      startTime.Add(time.Minute),
		QualityScore: 0.8,
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:   0,
				Longitude:  1,
				ObservedAt: startTime.Add(time.Minute),
			},
			{
				Latitude:   0,
				Longitude:  0,
				ObservedAt: startTime,
			},
		},
		Segments: []trajectory.TrajectorySegment{
			{
				ID:             "second",
				SequenceNumber: 2,
				Status:         trajectory.SegmentStatusObserved,
			},
			{
				ID:             "first",
				SequenceNumber: 1,
				Status:         trajectory.SegmentStatusObserved,
			},
		},
		CoverageGaps: []trajectory.CoverageGap{
			{
				ID:        "gap",
				StartTime: startTime.Add(10 * time.Second),
				EndTime:   startTime.Add(20 * time.Second),
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
	original.CoverageGaps = append(
		[]trajectory.CoverageGap(nil),
		item.CoverageGaps...,
	)

	if _, err := New().Build(
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
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := New().Build(
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
	features := flightfeatures.TrajectoryFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Limitations: []flightfeatures.FeatureLimitation{
				{Code: "original"},
			},
		},
	}

	cloned := cloneFeatures(features)
	cloned.Evidence.Limitations[0].Code = "changed"

	if features.Evidence.Limitations[0].Code !=
		"original" {
		t.Fatal("cloneFeatures() shared limitations")
	}
}

func TestTrajectoryBuilderContractConstantsRemainStable(
	t *testing.T,
) {
	if Version != "trajectory-feature-builder-v1" {
		t.Fatalf("Version = %q", Version)
	}
	if TrajectoryFeatureFieldCount != 16 {
		t.Fatalf(
			"TrajectoryFeatureFieldCount = %d",
			TrajectoryFeatureFieldCount,
		)
	}
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

func approximatelyEqual(
	left float64,
	right float64,
	tolerance float64,
) bool {
	return math.Abs(left-right) <= tolerance
}
