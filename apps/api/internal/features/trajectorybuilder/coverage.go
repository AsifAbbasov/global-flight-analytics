package trajectorybuilder

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func calculateCoverageRatio(
	ctx context.Context,
	evidence canonicalEvidence,
	segmentSummary segmentStatusSummary,
) (ratioMetric, []flightfeatures.FeatureLimitation, error) {
	if !evidence.windowAvailable {
		return ratioMetric{}, []flightfeatures.FeatureLimitation{{
			Code:    flightfeatures.TrajectoryLimitationCoverageWindowUnavailable,
			Message: "Trajectory start and end timestamps are required for coverage ratio calculation.",
		}}, nil
	}

	windowDuration := evidence.windowEnd.Sub(evidence.windowStart)
	usableSegmentCount := 0
	if segmentSummary.available {
		usableSegmentCount = segmentSummary.observedCount +
			segmentSummary.interpolatedCount +
			segmentSummary.estimatedCount
	}
	hasObservationEvidence := usableSegmentCount > 0 ||
		(windowDuration == 0 && len(evidence.points) > 0)
	if !hasObservationEvidence {
		return ratioMetric{}, []flightfeatures.FeatureLimitation{{
			Code:    flightfeatures.TrajectoryLimitationCoverageObservationEvidenceUnavailable,
			Message: "Positive-duration coverage requires at least one materialized non-invalid trajectory segment; a zero-duration instant requires one canonical point. Absence of coverage gaps alone does not prove observation coverage.",
		}}, nil
	}

	intervals := make([]timeInterval, 0, len(evidence.gaps))
	limitations := make([]flightfeatures.FeatureLimitation, 0, 4)
	outsideCount := 0
	durationMismatchCount := 0

	for index, gap := range evidence.gaps {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return ratioMetric{}, nil, err
			}
		}
		if gap.StartTime.IsZero() || gap.EndTime.IsZero() || gap.EndTime.Before(gap.StartTime) {
			return ratioMetric{}, append(limitations, flightfeatures.FeatureLimitation{
				Code: flightfeatures.TrajectoryLimitationCoverageGapWindowInvalid,
				Message: fmt.Sprintf(
					"Coverage gap %q has missing or reversed timestamps, so coverage ratio cannot be calculated reliably.",
					gap.ID,
				),
			}), nil
		}

		gapStart := gap.StartTime.UTC()
		gapEnd := gap.EndTime.UTC()
		actualDurationSeconds := flightfeatures.TemporalDurationSeconds(gapStart, gapEnd)
		if gap.DurationSeconds != actualDurationSeconds {
			durationMismatchCount++
		}

		if !gapEnd.After(evidence.windowStart) || !gapStart.Before(evidence.windowEnd) {
			outsideCount++
			continue
		}
		if gapStart.Before(evidence.windowStart) {
			gapStart = evidence.windowStart
		}
		if gapEnd.After(evidence.windowEnd) {
			gapEnd = evidence.windowEnd
		}
		if gapEnd.After(gapStart) {
			intervals = append(intervals, timeInterval{start: gapStart, end: gapEnd})
		}
	}

	if outsideCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationCoverageGapOutsideWindow,
			Message: fmt.Sprintf(
				"%d coverage gaps lie outside the authoritative trajectory window and were excluded.",
				outsideCount,
			),
		})
	}
	if durationMismatchCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationCoverageGapDurationMismatch,
			Message: fmt.Sprintf(
				"%d coverage-gap duration metadata values do not match timestamp durations under the shared truncate-fractional-seconds policy.",
				durationMismatchCount,
			),
		})
	}

	if windowDuration == 0 {
		return ratioMetric{available: true, value: 1}, limitations, nil
	}
	uncoveredDuration, err := unionDurationContext(ctx, intervals)
	if err != nil {
		return ratioMetric{}, nil, err
	}
	coverage := 1 - uncoveredDuration.Seconds()/windowDuration.Seconds()
	if math.IsNaN(coverage) || math.IsInf(coverage, 0) {
		return ratioMetric{}, append(limitations, flightfeatures.FeatureLimitation{
			Code:    flightfeatures.TrajectoryLimitationCoverageAggregateNonFinite,
			Message: "Coverage aggregation produced a non-finite ratio.",
		}), nil
	}
	if coverage < -pathRatioTolerance || coverage > 1+pathRatioTolerance {
		return ratioMetric{}, append(limitations, flightfeatures.FeatureLimitation{
			Code:    flightfeatures.TrajectoryLimitationCoverageRatioOutOfRange,
			Message: "Coverage ratio is outside the inclusive zero-to-one range beyond numerical tolerance.",
		}), nil
	}
	if coverage < 0 {
		coverage = 0
	}
	if coverage > 1 {
		coverage = 1
	}
	return ratioMetric{available: true, value: coverage}, limitations, nil
}

func unionDurationContext(
	ctx context.Context,
	intervals []timeInterval,
) (time.Duration, error) {
	if len(intervals) == 0 {
		return 0, nil
	}
	ordered := append([]timeInterval(nil), intervals...)
	// canonicalEvidence already sorts gaps, but clipping can equalize starts.
	for index := 1; index < len(ordered); index++ {
		if ordered[index].start.Before(ordered[index-1].start) {
			// Stable insertion keeps this helper allocation-free beyond the copy.
			value := ordered[index]
			position := index
			for position > 0 && value.start.Before(ordered[position-1].start) {
				ordered[position] = ordered[position-1]
				position--
			}
			ordered[position] = value
		}
	}
	currentStart := ordered[0].start
	currentEnd := ordered[0].end
	total := time.Duration(0)
	for index, interval := range ordered[1:] {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
		}
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
	return total + currentEnd.Sub(currentStart), nil
}
