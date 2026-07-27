package temporalbuilder

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestBuilderRejectsNilContext(t *testing.T) {
	_, err := New().Build(nil, trajectory.FlightTrajectory{})
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf("Build() error = %v, want %v", err, ErrContextRequired)
	}
}

func TestBuilderUsesPersistedSegmentBoundariesWhenPointsAreAbsent(t *testing.T) {
	startTime := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	middleTime := startTime.Add(time.Hour)
	endTime := middleTime.Add(time.Hour)

	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{
			StartTime:       startTime,
			EndTime:         endTime,
			DurationSeconds: 7200,
			PointCount:      9,
			Segments: []trajectory.TrajectorySegment{
				{
					Status:    trajectory.SegmentStatusObserved,
					StartTime: startTime,
					EndTime:   middleTime,
				},
				{
					Status:    trajectory.SegmentStatusEstimated,
					StartTime: middleTime,
					EndTime:   endTime,
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.Evidence.SupportingPointCount != 3 {
		t.Fatalf(
			"supporting temporal timestamps = %d, want 3",
			features.Evidence.SupportingPointCount,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"temporal_point_evidence_unavailable",
	) || !hasLimitation(
		features.Evidence.Limitations,
		"temporal_segment_boundary_fallback",
	) {
		t.Fatalf(
			"segment fallback limitations = %#v",
			features.Evidence.Limitations,
		)
	}
	if hasLimitation(
		features.Evidence.Limitations,
		"temporal_segment_evidence_unusable",
	) {
		t.Fatalf(
			"usable segment timestamps were rejected: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderRejectsInvalidSegmentsFromFallbackEvidence(t *testing.T) {
	startTime := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Hour)

	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{
			StartTime:       startTime,
			EndTime:         endTime,
			DurationSeconds: 3600,
			Segments: []trajectory.TrajectorySegment{
				{
					Status:    trajectory.SegmentStatusInvalid,
					StartTime: startTime,
					EndTime:   endTime,
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.Evidence.SupportingPointCount != 0 ||
		!hasLimitation(
			features.Evidence.Limitations,
			"temporal_invalid_segment_evidence",
		) || !hasLimitation(
		features.Evidence.Limitations,
		"temporal_segment_evidence_unusable",
	) {
		t.Fatalf("invalid segment evidence = %#v", features.Evidence)
	}
}

func TestBuilderReportsZeroDurationMetadataForNonZeroWindow(t *testing.T) {
	startTime := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Second)

	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{
			StartTime: startTime,
			EndTime:   endTime,
			Points: []trajectory.TrackPoint4D{
				{ObservedAt: startTime},
				{ObservedAt: endTime},
			},
			PointCount: 2,
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"trajectory_duration_metadata_mismatch",
	) {
		t.Fatalf(
			"zero duration metadata was treated as absent: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderUsesCentralFractionalSecondPolicy(t *testing.T) {
	startTime := time.Date(
		2026,
		time.July,
		27,
		8,
		0,
		0,
		250_000_000,
		time.UTC,
	)
	endTime := startTime.Add(1750 * time.Millisecond)

	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{
			StartTime:       startTime,
			EndTime:         endTime,
			DurationSeconds: 1,
			PointCount:      2,
			Points: []trajectory.TrackPoint4D{
				{ObservedAt: startTime},
				{ObservedAt: endTime},
			},
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.DurationSeconds != 1 {
		t.Fatalf("duration seconds = %d, want 1", features.DurationSeconds)
	}
	if hasLimitation(
		features.Evidence.Limitations,
		"trajectory_duration_metadata_mismatch",
	) {
		t.Fatalf(
			"central duration policy disagreed with metadata: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderReportsPointCountMetadataMismatch(t *testing.T) {
	startTime := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Second)

	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{
			StartTime:       startTime,
			EndTime:         endTime,
			DurationSeconds: 1,
			PointCount:      3,
			Points: []trajectory.TrackPoint4D{
				{ObservedAt: startTime},
				{ObservedAt: endTime},
			},
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"trajectory_point_count_metadata_mismatch",
	) {
		t.Fatalf(
			"point-count mismatch limitation = %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderObservesCancellationDuringPointScan(t *testing.T) {
	startTime := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Hour)
	points := make([]trajectory.TrackPoint4D, 2048)
	for index := range points {
		points[index].ObservedAt = startTime.Add(
			time.Duration(index) * time.Millisecond,
		)
	}
	ctx := &cancelAfterErrCallsContext{
		Context:     context.Background(),
		cancelAfter: 3,
	}

	_, err := New().Build(
		ctx,
		trajectory.FlightTrajectory{
			StartTime:       startTime,
			EndTime:         endTime,
			DurationSeconds: 3600,
			PointCount:      len(points),
			Points:          points,
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Build() error = %v, want context.Canceled", err)
	}
}

func TestBuilderLimitationMessagesContainExactCounts(t *testing.T) {
	startTime := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Hour)

	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{
			StartTime:       startTime,
			EndTime:         endTime,
			DurationSeconds: 3600,
			PointCount:      3,
			Points: []trajectory.TrackPoint4D{
				{},
				{ObservedAt: startTime.Add(-time.Second)},
				{ObservedAt: startTime},
			},
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	messages := make(map[string]string)
	for _, limitation := range features.Evidence.Limitations {
		messages[limitation.Code] = limitation.Message
	}
	for _, code := range []string{
		"temporal_point_timestamp_missing",
		"temporal_point_outside_window",
	} {
		if !strings.Contains(messages[code], "1") {
			t.Fatalf("limitation %q does not contain exact count: %q", code, messages[code])
		}
	}
}

type cancelAfterErrCallsContext struct {
	context.Context
	errCalls    int
	cancelAfter int
}

func (ctx *cancelAfterErrCallsContext) Err() error {
	ctx.errCalls++
	if ctx.errCalls >= ctx.cancelAfter {
		return context.Canceled
	}
	return nil
}

func TestTemporalBuilderProcessingVersionIsIsolated(t *testing.T) {
	if Version != "temporal-feature-builder-v2" {
		t.Fatalf("Version = %q", Version)
	}
	if flightfeatures.CurrentProcessingVersion !=
		"flight-feature-processing-pipeline-v11" {
		t.Fatalf(
			"processing version = %q",
			flightfeatures.CurrentProcessingVersion,
		)
	}
}
