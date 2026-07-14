package historicalwindow

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

const (
	Version = "historical-time-window-v1"

	FingerprintVersion = "historical-time-window-fingerprint-v1"

	BucketKeyPrefix = "historical-bucket-"

	DefaultMaximumBucketCount = 10_000
	MaximumBucketCount        = 100_000
)

type ExclusionReason string

const (
	ExclusionReasonLeadingIncompleteBucket  ExclusionReason = "leading_incomplete_bucket"
	ExclusionReasonTrailingIncompleteBucket ExclusionReason = "trailing_incomplete_bucket"
	ExclusionReasonNoCompleteBucket         ExclusionReason = "no_complete_bucket"
	ExclusionReasonFutureAfterAsOfTime      ExclusionReason = "future_after_as_of_time"
)

type Request struct {
	StartTime time.Time
	EndTime   time.Time
	AsOfTime  time.Time

	Granularity historicalcontract.Granularity

	MaximumBucketCount int
}

type Bucket struct {
	Key      string
	Sequence int

	StartTime time.Time
	EndTime   time.Time
}

func (bucket Bucket) Duration() time.Duration {
	if bucket.StartTime.IsZero() ||
		bucket.EndTime.IsZero() {
		return 0
	}

	return bucket.EndTime.Sub(bucket.StartTime)
}

func (bucket Bucket) Contains(
	value time.Time,
) bool {
	if bucket.StartTime.IsZero() ||
		bucket.EndTime.IsZero() {
		return false
	}

	normalized := value.UTC()

	return !normalized.Before(bucket.StartTime) &&
		normalized.Before(bucket.EndTime)
}

type Exclusion struct {
	Reason ExclusionReason

	StartTime time.Time
	EndTime   time.Time
}

func (exclusion Exclusion) Duration() time.Duration {
	if exclusion.StartTime.IsZero() ||
		exclusion.EndTime.IsZero() {
		return 0
	}

	return exclusion.EndTime.Sub(
		exclusion.StartTime,
	)
}

type Plan struct {
	Version     string
	Fingerprint string

	RequestedStartTime time.Time
	RequestedEndTime   time.Time
	AsOfTime           time.Time

	Granularity historicalcontract.Granularity

	EffectiveWindow *historicalcontract.TimeWindow
	PreviousWindow  *historicalcontract.TimeWindow

	Buckets    []Bucket
	Exclusions []Exclusion

	TruncatedByAsOfTime bool
	MaximumBucketCount  int
}

func (plan Plan) Clone() Plan {
	cloned := plan
	cloned.EffectiveWindow = cloneWindow(
		plan.EffectiveWindow,
	)
	cloned.PreviousWindow = cloneWindow(
		plan.PreviousWindow,
	)
	cloned.Buckets = append(
		[]Bucket(nil),
		plan.Buckets...,
	)
	cloned.Exclusions = append(
		[]Exclusion(nil),
		plan.Exclusions...,
	)

	return cloned
}

func (plan Plan) HasBuckets() bool {
	return len(plan.Buckets) > 0
}

func cloneWindow(
	window *historicalcontract.TimeWindow,
) *historicalcontract.TimeWindow {
	if window == nil {
		return nil
	}

	cloned := *window

	return &cloned
}
