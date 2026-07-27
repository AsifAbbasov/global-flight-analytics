package temporalbuilder

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

const contextCheckInterval = 1024

var _ extractor.TemporalBuilder = (*Builder)(nil)

type Builder struct{}

func New() *Builder {
	return &Builder{}
}

func (builder *Builder) Build(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (flightfeatures.TemporalFeatures, error) {
	if ctx == nil {
		return flightfeatures.TemporalFeatures{}, ErrContextRequired
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.TemporalFeatures{}, err
	}

	startTime, endTime, err := normalizeWindow(item)
	if err != nil {
		return flightfeatures.TemporalFeatures{}, err
	}

	supportingPointCount, limitations, err := evaluateTemporalEvidence(
		ctx,
		item,
		startTime,
		endTime,
	)
	if err != nil {
		return flightfeatures.TemporalFeatures{}, err
	}

	durationSeconds := flightfeatures.TemporalDurationSeconds(
		startTime,
		endTime,
	)
	if item.DurationSeconds != durationSeconds {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: "trajectory_duration_metadata_mismatch",
				Message: fmt.Sprintf(
					"Trajectory duration metadata reports %d seconds while the authoritative window resolves to %d seconds under policy %q.",
					item.DurationSeconds,
					durationSeconds,
					flightfeatures.CurrentTemporalDurationRoundingPolicy,
				),
			},
		)
	}

	features := flightfeatures.TemporalFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Status:               flightfeatures.AvailabilityStatusAvailable,
			AvailableFieldCount:  TemporalFeatureFieldCount,
			TotalFieldCount:      TemporalFeatureFieldCount,
			SupportingPointCount: supportingPointCount,
			Limitations:          limitations,
		},
		DurationSeconds:     durationSeconds,
		StartHourUTC:        startTime.Hour(),
		EndHourUTC:          endTime.Hour(),
		StartWeekday:        startTime.Weekday(),
		EndWeekday:          endTime.Weekday(),
		StartMinuteOfDayUTC: startTime.Hour()*60 + startTime.Minute(),
		EndMinuteOfDayUTC:   endTime.Hour()*60 + endTime.Minute(),
		CrossesUTCMidnight: crossesUTCCalendarBoundary(
			startTime,
			endTime,
		),
	}

	if err := ctx.Err(); err != nil {
		return flightfeatures.TemporalFeatures{}, err
	}

	return cloneFeatures(features), nil
}

func normalizeWindow(
	item trajectory.FlightTrajectory,
) (time.Time, time.Time, error) {
	if item.StartTime.IsZero() {
		return time.Time{},
			time.Time{},
			ErrTrajectoryStartTimeRequired
	}
	if item.EndTime.IsZero() {
		return time.Time{},
			time.Time{},
			ErrTrajectoryEndTimeRequired
	}
	if item.EndTime.Before(item.StartTime) {
		return time.Time{},
			time.Time{},
			ErrInvalidTrajectoryWindow
	}

	return item.StartTime.UTC(),
		item.EndTime.UTC(),
		nil
}

func evaluateTemporalEvidence(
	ctx context.Context,
	item trajectory.FlightTrajectory,
	startTime time.Time,
	endTime time.Time,
) (int, []flightfeatures.FeatureLimitation, error) {
	pointCount, limitations, err := evaluatePointEvidence(
		ctx,
		item.Points,
		startTime,
		endTime,
	)
	if err != nil {
		return 0, nil, err
	}

	if len(item.Points) > 0 && item.PointCount != len(item.Points) {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: "trajectory_point_count_metadata_mismatch",
				Message: fmt.Sprintf(
					"Trajectory point-count metadata reports %d points while %d point records are present.",
					item.PointCount,
					len(item.Points),
				),
			},
		)
	}
	if pointCount > 0 {
		return pointCount, limitations, nil
	}

	segmentCount, segmentLimitations, err := evaluateSegmentEvidence(
		ctx,
		item.Segments,
		startTime,
		endTime,
	)
	if err != nil {
		return 0, nil, err
	}
	limitations = append(limitations, segmentLimitations...)
	return segmentCount, limitations, nil
}

func evaluatePointEvidence(
	ctx context.Context,
	points []trajectory.TrackPoint4D,
	startTime time.Time,
	endTime time.Time,
) (
	int,
	[]flightfeatures.FeatureLimitation,
	error,
) {
	if len(points) == 0 {
		return 0, []flightfeatures.FeatureLimitation{
			{
				Code:    "temporal_point_evidence_unavailable",
				Message: "Temporal features were derived from trajectory boundaries because no trajectory point records were available.",
			},
		}, nil
	}

	supportingPointCount := 0
	zeroTimestampCount := 0
	outOfWindowCount := 0

	for index, point := range points {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return 0, nil, err
			}
		}
		if point.ObservedAt.IsZero() {
			zeroTimestampCount++
			continue
		}

		observedAt := point.ObservedAt.UTC()
		if observedAt.Before(startTime) ||
			observedAt.After(endTime) {
			outOfWindowCount++
			continue
		}

		supportingPointCount++
	}

	limitations := make(
		[]flightfeatures.FeatureLimitation,
		0,
		3,
	)
	if zeroTimestampCount > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: "temporal_point_timestamp_missing",
				Message: fmt.Sprintf(
					"%d trajectory point timestamps were missing and were excluded from temporal evidence.",
					zeroTimestampCount,
				),
			},
		)
	}
	if outOfWindowCount > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: "temporal_point_outside_window",
				Message: fmt.Sprintf(
					"%d trajectory point timestamps were outside the authoritative trajectory window and were excluded from temporal evidence.",
					outOfWindowCount,
				),
			},
		)
	}
	if supportingPointCount == 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: "temporal_point_evidence_unusable",
				Message: fmt.Sprintf(
					"None of the %d trajectory point records contained a timestamp that could support the authoritative temporal feature window.",
					len(points),
				),
			},
		)
	}

	return supportingPointCount, limitations, nil
}

func evaluateSegmentEvidence(
	ctx context.Context,
	segments []trajectory.TrajectorySegment,
	startTime time.Time,
	endTime time.Time,
) (
	int,
	[]flightfeatures.FeatureLimitation,
	error,
) {
	if len(segments) == 0 {
		return 0, nil, nil
	}

	uniqueTimestamps := make(map[int64]struct{}, len(segments)+1)
	invalidSegmentCount := 0
	missingTimestampCount := 0
	outOfWindowCount := 0

	for index, segment := range segments {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return 0, nil, err
			}
		}
		if segment.Status == trajectory.SegmentStatusInvalid {
			invalidSegmentCount++
			continue
		}

		for _, timestamp := range []time.Time{
			segment.StartTime,
			segment.EndTime,
		} {
			if timestamp.IsZero() {
				missingTimestampCount++
				continue
			}
			normalized := timestamp.UTC()
			if normalized.Before(startTime) ||
				normalized.After(endTime) {
				outOfWindowCount++
				continue
			}
			uniqueTimestamps[normalized.UnixNano()] = struct{}{}
		}
	}

	supportingPointCount := len(uniqueTimestamps)
	limitations := []flightfeatures.FeatureLimitation{
		{
			Code: "temporal_segment_boundary_fallback",
			Message: fmt.Sprintf(
				"%d unique persisted segment-boundary timestamps were used as temporal supporting evidence because no usable trajectory point timestamp was available.",
				supportingPointCount,
			),
		},
	}
	if invalidSegmentCount > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: "temporal_invalid_segment_evidence",
				Message: fmt.Sprintf(
					"%d invalid trajectory segments were excluded from temporal fallback evidence.",
					invalidSegmentCount,
				),
			},
		)
	}
	if missingTimestampCount > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: "temporal_segment_timestamp_missing",
				Message: fmt.Sprintf(
					"%d trajectory segment boundary timestamps were missing and were excluded from temporal fallback evidence.",
					missingTimestampCount,
				),
			},
		)
	}
	if outOfWindowCount > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: "temporal_segment_outside_window",
				Message: fmt.Sprintf(
					"%d trajectory segment boundary timestamps were outside the authoritative trajectory window and were excluded from temporal fallback evidence.",
					outOfWindowCount,
				),
			},
		)
	}
	if supportingPointCount == 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "temporal_segment_evidence_unusable",
				Message: "No persisted trajectory segment boundary timestamp could support the authoritative temporal feature window.",
			},
		)
	}

	return supportingPointCount, limitations, nil
}

func crossesUTCCalendarBoundary(
	startTime time.Time,
	endTime time.Time,
) bool {
	startYear, startMonth, startDay :=
		startTime.UTC().Date()
	endYear, endMonth, endDay :=
		endTime.UTC().Date()

	return startYear != endYear ||
		startMonth != endMonth ||
		startDay != endDay
}

func cloneFeatures(
	features flightfeatures.TemporalFeatures,
) flightfeatures.TemporalFeatures {
	cloned := features
	cloned.Evidence.Limitations = append(
		[]flightfeatures.FeatureLimitation(nil),
		features.Evidence.Limitations...,
	)

	return cloned
}
