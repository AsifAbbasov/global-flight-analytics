package trajectorybuilder

import (
	"fmt"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type ratioMetric struct {
	available bool
	value     float64
}

func calculateCoverageRatio(
	item trajectory.FlightTrajectory,
) (
	ratioMetric,
	[]flightfeatures.FeatureLimitation,
) {
	if item.StartTime.IsZero() ||
		item.EndTime.IsZero() {
		return ratioMetric{},
			[]flightfeatures.FeatureLimitation{
				{
					Code:    "trajectory_coverage_window_unavailable",
					Message: "Trajectory start and end timestamps are required for coverage ratio calculation.",
				},
			}
	}
	if item.EndTime.Before(item.StartTime) {
		return ratioMetric{},
			[]flightfeatures.FeatureLimitation{
				{
					Code:    "trajectory_coverage_window_invalid",
					Message: "Trajectory end time is before start time, so coverage ratio cannot be calculated.",
				},
			}
	}

	windowStart := item.StartTime.UTC()
	windowEnd := item.EndTime.UTC()
	windowDuration := windowEnd.Sub(windowStart)

	if windowDuration == 0 {
		if len(item.CoverageGaps) == 0 {
			return ratioMetric{
				available: true,
				value:     1,
			}, nil
		}

		return ratioMetric{},
			[]flightfeatures.FeatureLimitation{
				{
					Code:    "trajectory_coverage_zero_duration_with_gaps",
					Message: "Coverage ratio is undefined for a zero-duration trajectory window containing coverage gaps.",
				},
			}
	}

	intervals := make(
		[]timeInterval,
		0,
		len(item.CoverageGaps),
	)
	limitations := make(
		[]flightfeatures.FeatureLimitation,
		0,
		2,
	)

	for _, gap := range item.CoverageGaps {
		if gap.StartTime.IsZero() ||
			gap.EndTime.IsZero() ||
			gap.EndTime.Before(gap.StartTime) {
			return ratioMetric{},
				append(
					limitations,
					flightfeatures.FeatureLimitation{
						Code: "trajectory_coverage_gap_window_invalid",
						Message: fmt.Sprintf(
							"Coverage gap %q has missing or reversed timestamps, so coverage ratio cannot be calculated reliably.",
							gap.ID,
						),
					},
				)
		}

		gapStart := gap.StartTime.UTC()
		gapEnd := gap.EndTime.UTC()
		actualDurationSeconds := int64(
			gapEnd.Sub(gapStart) / time.Second,
		)
		if gap.DurationSeconds != 0 &&
			gap.DurationSeconds != actualDurationSeconds {
			limitations = append(
				limitations,
				flightfeatures.FeatureLimitation{
					Code: "trajectory_coverage_gap_duration_metadata_mismatch",
					Message: fmt.Sprintf(
						"Coverage gap %q duration metadata does not match its timestamps.",
						gap.ID,
					),
				},
			)
		}

		if gapEnd.Before(windowStart) ||
			gapStart.After(windowEnd) ||
			gapEnd.Equal(windowStart) ||
			gapStart.Equal(windowEnd) {
			limitations = append(
				limitations,
				flightfeatures.FeatureLimitation{
					Code: "trajectory_coverage_gap_outside_window",
					Message: fmt.Sprintf(
						"Coverage gap %q lies outside the authoritative trajectory window and was excluded.",
						gap.ID,
					),
				},
			)
			continue
		}

		if gapStart.Before(windowStart) {
			gapStart = windowStart
		}
		if gapEnd.After(windowEnd) {
			gapEnd = windowEnd
		}
		if !gapEnd.After(gapStart) {
			continue
		}

		intervals = append(
			intervals,
			timeInterval{
				start: gapStart,
				end:   gapEnd,
			},
		)
	}

	uncoveredDuration := unionDuration(intervals)
	coverage := 1 -
		uncoveredDuration.Seconds()/
			windowDuration.Seconds()
	if coverage < 0 {
		coverage = 0
	}
	if coverage > 1 {
		coverage = 1
	}

	return ratioMetric{
		available: true,
		value:     coverage,
	}, limitations
}

func unionDuration(
	intervals []timeInterval,
) time.Duration {
	if len(intervals) == 0 {
		return 0
	}

	ordered := append(
		[]timeInterval(nil),
		intervals...,
	)
	sort.SliceStable(
		ordered,
		func(left int, right int) bool {
			if ordered[left].start.Equal(
				ordered[right].start,
			) {
				return ordered[left].end.Before(
					ordered[right].end,
				)
			}

			return ordered[left].start.Before(
				ordered[right].start,
			)
		},
	)

	currentStart := ordered[0].start
	currentEnd := ordered[0].end
	total := time.Duration(0)

	for _, interval := range ordered[1:] {
		if !interval.start.After(currentEnd) {
			if interval.end.After(currentEnd) {
				currentEnd = interval.end
			}
			continue
		}

		total += currentEnd.Sub(currentStart)
		currentStart = interval.start
		currentEnd = interval.end
	}

	return total + currentEnd.Sub(currentStart)
}
