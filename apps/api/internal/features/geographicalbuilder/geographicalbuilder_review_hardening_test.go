package geographicalbuilder

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

func TestBuilderRejectsNilContext(t *testing.T) {
	_, err := newTestBuilder(t, Config{}).Build(nil, trajectory.FlightTrajectory{})
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf("Build() error = %v, want %v", err, ErrContextRequired)
	}
}

func TestBuilderOrdersEligiblePointsByObservationTime(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	end := start.Add(10 * time.Minute)
	item := trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    end,
		PointCount: 3,
		Points: []trajectory.TrackPoint4D{
			{ID: "late", Latitude: 42, Longitude: 52, ObservedAt: end},
			{ID: "early", Latitude: 40, Longitude: 50, ObservedAt: start},
			{ID: "middle", Latitude: 41, Longitude: 51, ObservedAt: start.Add(5 * time.Minute)},
		},
	}

	features, err := newTestBuilder(t, Config{}).Build(context.Background(), item)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.StartLatitude != 40 || features.EndLatitude != 42 {
		t.Fatalf("chronological endpoints = %#v", features)
	}
	if features.Evidence.SupportingPointCount != 3 {
		t.Fatalf("supporting points = %d, want 3", features.Evidence.SupportingPointCount)
	}
}

func TestBuilderTieBreaksEqualObservationTimesDeterministically(t *testing.T) {
	instant := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	points := []trajectory.TrackPoint4D{
		{Latitude: 42, Longitude: 52, ObservedAt: instant},
		{Latitude: 40, Longitude: 50, ObservedAt: instant},
		{Latitude: 41, Longitude: 51, ObservedAt: instant},
	}
	first := trajectory.FlightTrajectory{
		StartTime: instant, EndTime: instant, PointCount: len(points),
		Points: append([]trajectory.TrackPoint4D(nil), points...),
	}
	second := first
	second.Points = []trajectory.TrackPoint4D{points[2], points[0], points[1]}

	builder := newTestBuilder(t, Config{})
	left, err := builder.Build(context.Background(), first)
	if err != nil {
		t.Fatalf("first Build() error = %v", err)
	}
	right, err := builder.Build(context.Background(), second)
	if err != nil {
		t.Fatalf("second Build() error = %v", err)
	}
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("equal-time point permutation changed features\nleft=%#v\nright=%#v", left, right)
	}
}

func TestBuilderExcludesMissingAndOutsideWindowPointTimestamps(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	end := start.Add(10 * time.Minute)
	item := trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    end,
		PointCount: 5,
		Points: []trajectory.TrackPoint4D{
			{ID: "missing", Latitude: 1, Longitude: 1},
			{ID: "before", Latitude: 2, Longitude: 2, ObservedAt: start.Add(-time.Second)},
			{ID: "first", Latitude: 40, Longitude: 50, ObservedAt: start},
			{ID: "last", Latitude: 41, Longitude: 51, ObservedAt: end},
			{ID: "after", Latitude: 3, Longitude: 3, ObservedAt: end.Add(time.Second)},
		},
	}

	features, err := newTestBuilder(t, Config{}).Build(context.Background(), item)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.Evidence.SupportingPointCount != 2 ||
		features.MinimumLatitude != 40 || features.MaximumLatitude != 41 {
		t.Fatalf("ineligible points changed features: %#v", features)
	}
	for _, code := range []string{
		flightfeatures.GeographicalLimitationPointTimestampMissing,
		flightfeatures.GeographicalLimitationPointOutsideWindow,
	} {
		if !hasLimitation(features.Evidence.Limitations, code) {
			t.Fatalf("missing limitation %q in %#v", code, features.Evidence.Limitations)
		}
	}
}

func TestBuilderUsesSegmentsWhenOnlyOnePointIsEligible(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	end := start.Add(10 * time.Minute)
	item := trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    end,
		PointCount: 1,
		Points: []trajectory.TrackPoint4D{
			{ID: "only", Latitude: 80, Longitude: 80, ObservedAt: start.Add(time.Minute)},
		},
		Segments: []trajectory.TrajectorySegment{
			{
				ID: "segment", SequenceNumber: 1, Status: trajectory.SegmentStatusObserved,
				StartTime: start, EndTime: end, StartLatitude: 40, StartLongitude: 50,
				EndLatitude: 41, EndLongitude: 51, PointCount: 1,
			},
		},
	}

	features, err := newTestBuilder(t, Config{}).Build(context.Background(), item)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.StartLatitude != 40 || features.EndLatitude != 41 {
		t.Fatalf("single point blocked segment fallback: %#v", features)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		flightfeatures.GeographicalLimitationSinglePointSegmentFallback,
	) {
		t.Fatalf("missing single-point fallback limitation: %#v", features.Evidence.Limitations)
	}
}

func TestSegmentFallbackExcludesDiscontinuityFromObservedPath(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	end := start.Add(20 * time.Minute)
	item := trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    end,
		PointCount: 20,
		Segments: []trajectory.TrajectorySegment{
			{
				ID: "one", SequenceNumber: 1, Status: trajectory.SegmentStatusObserved,
				StartTime: start, EndTime: start.Add(5 * time.Minute),
				StartLatitude: 0, StartLongitude: 0, EndLatitude: 0, EndLongitude: 1,
				PointCount: 10,
			},
			{
				ID: "two", SequenceNumber: 2, Status: trajectory.SegmentStatusObserved,
				StartTime: start.Add(15 * time.Minute), EndTime: end,
				StartLatitude: 0, StartLongitude: 100, EndLatitude: 0, EndLongitude: 101,
				PointCount: 10,
			},
			{
				ID: "invalid", SequenceNumber: 3, Status: trajectory.SegmentStatusInvalid,
				StartTime: end, EndTime: end,
				StartLatitude: 0, StartLongitude: 101, EndLatitude: 0, EndLongitude: 102,
			},
		},
	}

	features, err := newTestBuilder(t, Config{}).Build(context.Background(), item)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.ObservedPathDistanceKM < 222 || features.ObservedPathDistanceKM > 224 {
		t.Fatalf("observed path includes gap: %v", features.ObservedPathDistanceKM)
	}
	if features.GreatCircleDistanceKM <= features.ObservedPathDistanceKM {
		t.Fatalf("test does not contain a disconnected envelope: %#v", features)
	}
	if features.Evidence.SupportingPointCount != 20 {
		t.Fatalf("supporting count = %d, want trajectory point count 20", features.Evidence.SupportingPointCount)
	}
	for _, code := range []string{
		flightfeatures.GeographicalLimitationSegmentDiscontinuityExcluded,
		flightfeatures.GeographicalLimitationInvalidSegmentStatus,
	} {
		if !hasLimitation(features.Evidence.Limitations, code) {
			t.Fatalf("missing segment limitation %q: %#v", code, features.Evidence.Limitations)
		}
	}
}

func TestCircularEnvelopeMayWrapWithoutPathCrossing(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	item := trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    start.Add(2 * time.Minute),
		PointCount: 3,
		Points: []trajectory.TrackPoint4D{
			{ID: "one", Latitude: 0, Longitude: 160, ObservedAt: start},
			{ID: "two", Latitude: 0, Longitude: 0, ObservedAt: start.Add(time.Minute)},
			{ID: "three", Latitude: 0, Longitude: -150, ObservedAt: start.Add(2 * time.Minute)},
		},
	}

	features, err := newTestBuilder(t, Config{}).Build(context.Background(), item)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.MinimumLongitude != 160 || features.MaximumLongitude != 0 ||
		features.LongitudeSpanDegrees != 200 || features.CrossesAntimeridian {
		t.Fatalf("circular envelope/path semantics = %#v", features)
	}
}

func TestBuilderReportsPointCountMismatchWhenCoordinatesAreUnavailable(t *testing.T) {
	item := trajectory.FlightTrajectory{
		PointCount: 2,
		Points: []trajectory.TrackPoint4D{
			{Latitude: math.NaN(), Longitude: 0},
		},
	}
	features, err := newTestBuilder(t, Config{}).Build(context.Background(), item)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.Evidence.Status != flightfeatures.AvailabilityStatusUnavailable ||
		!hasLimitation(
			features.Evidence.Limitations,
			flightfeatures.LimitationTrajectoryPointCountMetadataMismatch,
		) {
		t.Fatalf("unavailable result lost point-count mismatch: %#v", features)
	}
}

func TestGeometryPassesObserveContextCancellation(t *testing.T) {
	coordinates := make([]coordinate, 4096)
	for index := range coordinates {
		coordinates[index] = coordinate{latitude: float64(index % 90), longitude: float64(index % 180)}
	}
	ctx := &cancelAfterChecksContext{cancelAt: 2}
	_, _, err := latitudeBoundsContext(ctx, coordinates)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("latitudeBoundsContext() error = %v, want context.Canceled", err)
	}
}

type cancelAfterChecksContext struct {
	checks   int
	cancelAt int
}

func (ctx *cancelAfterChecksContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (ctx *cancelAfterChecksContext) Done() <-chan struct{}       { return nil }
func (ctx *cancelAfterChecksContext) Value(any) any               { return nil }
func (ctx *cancelAfterChecksContext) Err() error {
	ctx.checks++
	if ctx.checks >= ctx.cancelAt {
		return context.Canceled
	}
	return nil
}
