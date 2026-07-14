package temporalbuilder

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

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
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.TemporalFeatures{}, err
	}

	startTime, endTime, err := normalizeWindow(item)
	if err != nil {
		return flightfeatures.TemporalFeatures{}, err
	}

	supportingPointCount, pointLimitations :=
		evaluatePointEvidence(
			item.Points,
			startTime,
			endTime,
		)
	limitations := append(
		[]flightfeatures.FeatureLimitation(nil),
		pointLimitations...,
	)

	durationSeconds := int64(
		endTime.Sub(startTime) / time.Second,
	)
	if item.DurationSeconds != 0 &&
		item.DurationSeconds != durationSeconds {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "trajectory_duration_metadata_mismatch",
				Message: "Trajectory duration metadata does not match the authoritative start and end timestamps.",
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

func evaluatePointEvidence(
	points []trajectory.TrackPoint4D,
	startTime time.Time,
	endTime time.Time,
) (
	int,
	[]flightfeatures.FeatureLimitation,
) {
	if len(points) == 0 {
		return 0, []flightfeatures.FeatureLimitation{
			{
				Code:    "temporal_point_evidence_unavailable",
				Message: "Temporal features were derived from trajectory boundaries because no trajectory points were available.",
			},
		}
	}

	supportingPointCount := 0
	zeroTimestampCount := 0
	outOfWindowCount := 0

	for _, point := range points {
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
		2,
	)
	if zeroTimestampCount > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "temporal_point_timestamp_missing",
				Message: "One or more trajectory points have no observation timestamp and were excluded from temporal evidence.",
			},
		)
	}
	if outOfWindowCount > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "temporal_point_outside_window",
				Message: "One or more trajectory point timestamps fall outside the authoritative trajectory window and were excluded from temporal evidence.",
			},
		)
	}
	if supportingPointCount == 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "temporal_point_evidence_unusable",
				Message: "No trajectory point timestamp could support the authoritative temporal feature window.",
			},
		)
	}

	return supportingPointCount, limitations
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
