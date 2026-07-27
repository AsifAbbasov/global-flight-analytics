package historicalwindow

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

const (
	maximumInt64 = int64(1<<63 - 1)
	minimumInt64 = int64(-1 << 63)
)

func Build(
	ctx context.Context,
	request Request,
) (Plan, error) {
	if ctx == nil {
		return Plan{}, ErrContextRequired
	}
	if err := ctx.Err(); err != nil {
		return Plan{}, err
	}

	normalized, err := normalizeRequest(request)
	if err != nil {
		return Plan{}, err
	}

	plan, err := buildNormalizedPlan(ctx, normalized)
	if err != nil {
		return Plan{}, err
	}

	return plan.Clone(), nil
}

func buildNormalizedPlan(
	ctx context.Context,
	request Request,
) (Plan, error) {
	plan := Plan{
		Version: Version,

		RequestedStartTime: request.StartTime,
		RequestedEndTime:   request.EndTime,
		AsOfTime:           request.AsOfTime,

		Granularity: request.Granularity,

		TruncatedByAsOfTime: request.EndTime.After(
			request.AsOfTime,
		),
		MaximumBucketCount: request.MaximumBucketCount,
	}

	var err error
	if request.Granularity ==
		historicalcontract.GranularityCustom {
		err = buildCustomPlan(ctx, &plan)
	} else {
		err = buildClosedCalendarPlan(ctx, &plan)
	}
	if err != nil {
		return Plan{}, err
	}
	if err := ctx.Err(); err != nil {
		return Plan{}, err
	}

	plan.Fingerprint = planFingerprint(plan)

	return plan, nil
}

func normalizeRequest(
	request Request,
) (Request, error) {
	if request.StartTime.IsZero() {
		return Request{}, ErrStartTimeRequired
	}
	if request.EndTime.IsZero() {
		return Request{}, ErrEndTimeRequired
	}
	if request.AsOfTime.IsZero() {
		return Request{}, ErrAsOfTimeRequired
	}
	if !isSupportedGranularity(
		request.Granularity,
	) {
		return Request{},
			ErrUnsupportedGranularity
	}

	startTime := request.StartTime.UTC()
	endTime := request.EndTime.UTC()
	asOfTime := request.AsOfTime.UTC()

	if !startTime.Before(endTime) {
		return Request{}, ErrWindowNotPositive
	}

	maximumBucketCount :=
		request.MaximumBucketCount
	if maximumBucketCount == 0 {
		maximumBucketCount =
			DefaultMaximumBucketCount
	}
	if maximumBucketCount < 1 ||
		maximumBucketCount >
			MaximumBucketCount {
		return Request{},
			ErrInvalidMaximumBucketCount
	}

	return Request{
		StartTime: startTime,
		EndTime:   endTime,
		AsOfTime:  asOfTime,

		Granularity: request.Granularity,

		MaximumBucketCount: maximumBucketCount,
	}, nil
}

func buildCustomPlan(
	ctx context.Context,
	plan *Plan,
) error {
	cutoff := earlierTime(
		plan.RequestedEndTime,
		plan.AsOfTime,
	)

	if !plan.RequestedStartTime.Before(
		cutoff,
	) {
		appendEntireFutureExclusion(plan)
		return ctx.Err()
	}

	plan.Buckets = []Bucket{
		newBucket(
			1,
			plan.Granularity,
			plan.RequestedStartTime,
			cutoff,
		),
	}
	assignEffectiveWindow(
		plan,
		plan.RequestedStartTime,
		cutoff,
	)
	if err := assignPreviousCustomWindow(
		plan,
		plan.RequestedStartTime,
		cutoff,
	); err != nil {
		return err
	}

	appendFutureExclusion(plan, cutoff)

	return ctx.Err()
}

func buildClosedCalendarPlan(
	ctx context.Context,
	plan *Plan,
) error {
	cutoff := earlierTime(
		plan.RequestedEndTime,
		plan.AsOfTime,
	)

	if !plan.RequestedStartTime.Before(
		cutoff,
	) {
		appendEntireFutureExclusion(plan)
		return ctx.Err()
	}

	effectiveStart, effectiveEnd, err :=
		closedCalendarBounds(plan, cutoff)
	if err != nil {
		return err
	}
	if !effectiveStart.Before(effectiveEnd) {
		appendNoCompleteBucketExclusion(
			plan,
			cutoff,
		)
		appendFutureExclusion(plan, cutoff)
		return ctx.Err()
	}

	appendLeadingExclusion(
		plan,
		effectiveStart,
	)

	plan.Buckets, err = generateBuckets(
		ctx,
		plan.Granularity,
		effectiveStart,
		effectiveEnd,
		plan.MaximumBucketCount,
	)
	if err != nil {
		return err
	}

	assignEffectiveWindow(
		plan,
		effectiveStart,
		effectiveEnd,
	)
	if err := assignPreviousCalendarWindow(
		plan,
		effectiveStart,
	); err != nil {
		return err
	}

	appendTrailingExclusion(
		plan,
		effectiveEnd,
		cutoff,
	)
	appendFutureExclusion(plan, cutoff)

	return ctx.Err()
}

func closedCalendarBounds(
	plan *Plan,
	cutoff time.Time,
) (time.Time, time.Time, error) {
	effectiveStart, err := CeilBoundary(
		plan.RequestedStartTime,
		plan.Granularity,
	)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	effectiveEnd, err := FloorBoundary(
		cutoff,
		plan.Granularity,
	)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return effectiveStart, effectiveEnd, nil
}

func generateBuckets(
	ctx context.Context,
	granularity historicalcontract.Granularity,
	startTime time.Time,
	endTime time.Time,
	maximum int,
) ([]Bucket, error) {
	buckets := make([]Bucket, 0)
	current := startTime

	for current.Before(endTime) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(buckets) >= maximum {
			return nil, &BucketCountExceededError{
				Granularity: granularity,
				Count:       len(buckets) + 1,
				Maximum:     maximum,
			}
		}

		next, err := NextBoundary(
			current,
			granularity,
		)
		if err != nil {
			return nil, err
		}
		if !next.After(current) ||
			next.After(endTime) {
			return nil, ErrBoundarySequenceInvalid
		}

		buckets = append(
			buckets,
			newBucket(
				len(buckets)+1,
				granularity,
				current,
				next,
			),
		)
		current = next
	}

	if !current.Equal(endTime) {
		return nil, ErrBoundarySequenceInvalid
	}

	return buckets, nil
}

func assignEffectiveWindow(
	plan *Plan,
	startTime time.Time,
	endTime time.Time,
) {
	effectiveWindow :=
		historicalcontract.TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  plan.AsOfTime,
		}
	plan.EffectiveWindow = &effectiveWindow
}

func assignPreviousCalendarWindow(
	plan *Plan,
	startTime time.Time,
) error {
	previousStart, err := shiftBoundary(
		startTime,
		plan.Granularity,
		-len(plan.Buckets),
	)
	if err != nil {
		return err
	}

	assignPreviousWindow(
		plan,
		previousStart,
		startTime,
	)
	return nil
}

func assignPreviousCustomWindow(
	plan *Plan,
	startTime time.Time,
	endTime time.Time,
) error {
	previousStart, err := previousStartForSpan(
		startTime,
		endTime,
	)
	if err != nil {
		return err
	}

	assignPreviousWindow(
		plan,
		previousStart,
		startTime,
	)
	return nil
}

func assignPreviousWindow(
	plan *Plan,
	startTime time.Time,
	endTime time.Time,
) {
	previousWindow :=
		historicalcontract.TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  plan.AsOfTime,
		}
	plan.PreviousWindow = &previousWindow
}

func previousStartForSpan(
	startTime time.Time,
	endTime time.Time,
) (time.Time, error) {
	seconds, nanoseconds, err := exactSpan(
		startTime,
		endTime,
	)
	if err != nil {
		return time.Time{}, err
	}

	previousSeconds, ok := checkedSubtractInt64(
		startTime.Unix(),
		seconds,
	)
	if !ok {
		return time.Time{},
			ErrPreviousWindowOutOfRange
	}
	previousNanoseconds := int64(
		startTime.Nanosecond(),
	) - nanoseconds
	if previousNanoseconds < 0 {
		previousSeconds, ok = checkedSubtractInt64(
			previousSeconds,
			1,
		)
		if !ok {
			return time.Time{},
				ErrPreviousWindowOutOfRange
		}
		previousNanoseconds += int64(time.Second)
	}

	previousStart := time.Unix(
		previousSeconds,
		previousNanoseconds,
	).UTC()
	if !previousStart.Before(startTime) {
		return time.Time{},
			ErrPreviousWindowOutOfRange
	}

	return previousStart, nil
}

func exactSpan(
	startTime time.Time,
	endTime time.Time,
) (int64, int64, error) {
	seconds, ok := checkedSubtractInt64(
		endTime.Unix(),
		startTime.Unix(),
	)
	if !ok {
		return 0, 0,
			ErrPreviousWindowOutOfRange
	}
	nanoseconds := int64(
		endTime.Nanosecond(),
	) - int64(startTime.Nanosecond())
	if nanoseconds < 0 {
		seconds, ok = checkedSubtractInt64(
			seconds,
			1,
		)
		if !ok {
			return 0, 0,
				ErrPreviousWindowOutOfRange
		}
		nanoseconds += int64(time.Second)
	}
	if seconds < 0 ||
		(seconds == 0 && nanoseconds <= 0) {
		return 0, 0, ErrWindowNotPositive
	}

	return seconds, nanoseconds, nil
}

func checkedSubtractInt64(
	left int64,
	right int64,
) (int64, bool) {
	if right > 0 &&
		left < minimumInt64+right {
		return 0, false
	}
	if right < 0 &&
		left > maximumInt64+right {
		return 0, false
	}

	return left - right, true
}

func appendLeadingExclusion(
	plan *Plan,
	effectiveStart time.Time,
) {
	if !plan.RequestedStartTime.Before(
		effectiveStart,
	) {
		return
	}

	plan.Exclusions = append(
		plan.Exclusions,
		Exclusion{
			Reason:    ExclusionReasonLeadingIncompleteBucket,
			StartTime: plan.RequestedStartTime,
			EndTime:   effectiveStart,
		},
	)
}

func appendTrailingExclusion(
	plan *Plan,
	effectiveEnd time.Time,
	cutoff time.Time,
) {
	if !effectiveEnd.Before(cutoff) {
		return
	}

	plan.Exclusions = append(
		plan.Exclusions,
		Exclusion{
			Reason:    ExclusionReasonTrailingIncompleteBucket,
			StartTime: effectiveEnd,
			EndTime:   cutoff,
		},
	)
}

func appendNoCompleteBucketExclusion(
	plan *Plan,
	cutoff time.Time,
) {
	plan.Exclusions = append(
		plan.Exclusions,
		Exclusion{
			Reason:    ExclusionReasonNoCompleteBucket,
			StartTime: plan.RequestedStartTime,
			EndTime:   cutoff,
		},
	)
}

func appendEntireFutureExclusion(
	plan *Plan,
) {
	plan.Exclusions = append(
		plan.Exclusions,
		Exclusion{
			Reason:    ExclusionReasonFutureAfterAsOfTime,
			StartTime: plan.RequestedStartTime,
			EndTime:   plan.RequestedEndTime,
		},
	)
}

func appendFutureExclusion(
	plan *Plan,
	cutoff time.Time,
) {
	if !cutoff.Before(
		plan.RequestedEndTime,
	) {
		return
	}

	plan.Exclusions = append(
		plan.Exclusions,
		Exclusion{
			Reason:    ExclusionReasonFutureAfterAsOfTime,
			StartTime: cutoff,
			EndTime:   plan.RequestedEndTime,
		},
	)
}

func newBucket(
	sequence int,
	granularity historicalcontract.Granularity,
	startTime time.Time,
	endTime time.Time,
) Bucket {
	return Bucket{
		Key: bucketKey(
			granularity,
			startTime,
			endTime,
		),
		Sequence:  sequence,
		StartTime: startTime.UTC(),
		EndTime:   endTime.UTC(),
	}
}

func earlierTime(
	left time.Time,
	right time.Time,
) time.Time {
	if left.Before(right) {
		return left
	}

	return right
}
