package historicalwindow

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func Build(
	ctx context.Context,
	request Request,
) (Plan, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Plan{}, err
	}

	normalized, err := normalizeRequest(request)
	if err != nil {
		return Plan{}, err
	}

	plan := Plan{
		Version: Version,

		RequestedStartTime: normalized.StartTime,
		RequestedEndTime:   normalized.EndTime,
		AsOfTime:           normalized.AsOfTime,

		Granularity: normalized.Granularity,

		TruncatedByAsOfTime: normalized.EndTime.After(
			normalized.AsOfTime,
		),
		MaximumBucketCount: normalized.MaximumBucketCount,
	}

	if normalized.Granularity ==
		historicalcontract.GranularityCustom {
		buildCustomPlan(&plan)
	} else {
		if err := buildClosedCalendarPlan(
			ctx,
			&plan,
		); err != nil {
			return Plan{}, err
		}
	}

	plan.Fingerprint = planFingerprint(plan)

	return plan.Clone(), nil
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
	plan *Plan,
) {
	cutoff := earlierTime(
		plan.RequestedEndTime,
		plan.AsOfTime,
	)

	if !plan.RequestedStartTime.Before(
		cutoff,
	) {
		plan.Exclusions = append(
			plan.Exclusions,
			Exclusion{
				Reason:    ExclusionReasonFutureAfterAsOfTime,
				StartTime: plan.RequestedStartTime,
				EndTime:   plan.RequestedEndTime,
			},
		)
		return
	}

	plan.Buckets = []Bucket{
		newBucket(
			1,
			plan.Granularity,
			plan.RequestedStartTime,
			cutoff,
		),
	}
	setEffectiveAndPreviousWindows(
		plan,
		plan.RequestedStartTime,
		cutoff,
	)

	if cutoff.Before(
		plan.RequestedEndTime,
	) {
		plan.Exclusions = append(
			plan.Exclusions,
			Exclusion{
				Reason:    ExclusionReasonFutureAfterAsOfTime,
				StartTime: cutoff,
				EndTime:   plan.RequestedEndTime,
			},
		)
	}
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
		plan.Exclusions = append(
			plan.Exclusions,
			Exclusion{
				Reason:    ExclusionReasonFutureAfterAsOfTime,
				StartTime: plan.RequestedStartTime,
				EndTime:   plan.RequestedEndTime,
			},
		)
		return nil
	}

	effectiveStart, err := CeilBoundary(
		plan.RequestedStartTime,
		plan.Granularity,
	)
	if err != nil {
		return err
	}
	effectiveEnd, err := FloorBoundary(
		cutoff,
		plan.Granularity,
	)
	if err != nil {
		return err
	}

	if !effectiveStart.Before(
		effectiveEnd,
	) {
		plan.Exclusions = append(
			plan.Exclusions,
			Exclusion{
				Reason:    ExclusionReasonNoCompleteBucket,
				StartTime: plan.RequestedStartTime,
				EndTime:   cutoff,
			},
		)
		appendFutureExclusion(
			plan,
			cutoff,
		)

		return nil
	}

	if plan.RequestedStartTime.Before(
		effectiveStart,
	) {
		plan.Exclusions = append(
			plan.Exclusions,
			Exclusion{
				Reason:    ExclusionReasonLeadingIncompleteBucket,
				StartTime: plan.RequestedStartTime,
				EndTime:   effectiveStart,
			},
		)
	}

	duration, err := boundaryDuration(
		plan.Granularity,
	)
	if err != nil {
		return err
	}
	bucketCount := int(
		effectiveEnd.Sub(effectiveStart) /
			duration,
	)
	if bucketCount >
		plan.MaximumBucketCount {
		return &BucketCountExceededError{
			Granularity: plan.Granularity,
			Count:       bucketCount,
			Maximum:     plan.MaximumBucketCount,
		}
	}

	plan.Buckets = make(
		[]Bucket,
		0,
		bucketCount,
	)

	current := effectiveStart
	for sequence := 1; current.Before(effectiveEnd); sequence++ {
		if sequence%1_024 == 0 {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		next, err := NextBoundary(
			current,
			plan.Granularity,
		)
		if err != nil {
			return err
		}
		if next.After(effectiveEnd) {
			return ErrBoundarySequenceInvalid
		}

		plan.Buckets = append(
			plan.Buckets,
			newBucket(
				sequence,
				plan.Granularity,
				current,
				next,
			),
		)
		current = next
	}

	setEffectiveAndPreviousWindows(
		plan,
		effectiveStart,
		effectiveEnd,
	)

	if effectiveEnd.Before(cutoff) {
		plan.Exclusions = append(
			plan.Exclusions,
			Exclusion{
				Reason:    ExclusionReasonTrailingIncompleteBucket,
				StartTime: effectiveEnd,
				EndTime:   cutoff,
			},
		)
	}

	appendFutureExclusion(plan, cutoff)

	return nil
}

func appendFutureExclusion(
	plan *Plan,
	cutoff time.Time,
) {
	if cutoff.Before(
		plan.RequestedEndTime,
	) {
		plan.Exclusions = append(
			plan.Exclusions,
			Exclusion{
				Reason:    ExclusionReasonFutureAfterAsOfTime,
				StartTime: cutoff,
				EndTime:   plan.RequestedEndTime,
			},
		)
	}
}

func setEffectiveAndPreviousWindows(
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
	plan.EffectiveWindow =
		&effectiveWindow

	duration := endTime.Sub(startTime)
	previousWindow :=
		historicalcontract.TimeWindow{
			StartTime: startTime.Add(-duration),
			EndTime:   startTime,
			AsOfTime:  plan.AsOfTime,
		}
	plan.PreviousWindow =
		&previousWindow
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
