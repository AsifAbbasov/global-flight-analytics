package historicalwindow

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func CanonicalizePlan(
	plan Plan,
) (Plan, error) {
	if plan.Version != Version {
		return Plan{}, ErrPlanVersionInvalid
	}

	normalized, err := normalizeRequest(
		Request{
			StartTime:          plan.RequestedStartTime,
			EndTime:            plan.RequestedEndTime,
			AsOfTime:           plan.AsOfTime,
			Granularity:        plan.Granularity,
			MaximumBucketCount: plan.MaximumBucketCount,
		},
	)
	if err != nil {
		return Plan{}, fmt.Errorf(
			"%w: request: %v",
			ErrPlanIntegrityInvalid,
			err,
		)
	}

	canonical, err := buildNormalizedPlan(
		context.Background(),
		normalized,
	)
	if err != nil {
		return Plan{}, fmt.Errorf(
			"%w: rebuild: %v",
			ErrPlanIntegrityInvalid,
			err,
		)
	}

	return canonical.Clone(), nil
}

func ValidatePlan(
	plan Plan,
) error {
	canonical, err := CanonicalizePlan(plan)
	if err != nil {
		return err
	}

	return compareCanonicalPlan(plan, canonical)
}

func compareCanonicalPlan(
	actual Plan,
	expected Plan,
) error {
	if actual.Version != expected.Version {
		return planFieldError("version")
	}
	if actual.Fingerprint != expected.Fingerprint {
		return planFieldError("fingerprint")
	}
	if !canonicalTimeEqual(
		actual.RequestedStartTime,
		expected.RequestedStartTime,
	) {
		return planFieldError("requested_start_time")
	}
	if !canonicalTimeEqual(
		actual.RequestedEndTime,
		expected.RequestedEndTime,
	) {
		return planFieldError("requested_end_time")
	}
	if !canonicalTimeEqual(
		actual.AsOfTime,
		expected.AsOfTime,
	) {
		return planFieldError("as_of_time")
	}
	if actual.Granularity != expected.Granularity {
		return planFieldError("granularity")
	}
	if actual.TruncatedByAsOfTime !=
		expected.TruncatedByAsOfTime {
		return planFieldError("truncated_by_as_of_time")
	}
	if actual.MaximumBucketCount !=
		expected.MaximumBucketCount {
		return planFieldError("maximum_bucket_count")
	}
	if !windowsEqual(
		actual.EffectiveWindow,
		expected.EffectiveWindow,
	) {
		return planFieldError("effective_window")
	}
	if !windowsEqual(
		actual.PreviousWindow,
		expected.PreviousWindow,
	) {
		return planFieldError("previous_window")
	}
	if len(actual.Buckets) != len(expected.Buckets) {
		return planFieldError("buckets")
	}
	for index := range expected.Buckets {
		if !bucketsEqual(
			actual.Buckets[index],
			expected.Buckets[index],
		) {
			return planFieldError(
				fmt.Sprintf("buckets[%d]", index),
			)
		}
	}
	if len(actual.Exclusions) !=
		len(expected.Exclusions) {
		return planFieldError("exclusions")
	}
	for index := range expected.Exclusions {
		if !exclusionsEqual(
			actual.Exclusions[index],
			expected.Exclusions[index],
		) {
			return planFieldError(
				fmt.Sprintf("exclusions[%d]", index),
			)
		}
	}

	return nil
}

func planFieldError(field string) error {
	return fmt.Errorf(
		"%w: %s",
		ErrPlanIntegrityInvalid,
		field,
	)
}

func canonicalTimeEqual(
	actual time.Time,
	expected time.Time,
) bool {
	return actual.Location() == time.UTC &&
		actual.Equal(expected)
}

func windowsEqual(
	actual *historicalcontract.TimeWindow,
	expected *historicalcontract.TimeWindow,
) bool {
	if actual == nil || expected == nil {
		return actual == nil && expected == nil
	}

	return canonicalTimeEqual(
		actual.StartTime,
		expected.StartTime,
	) && canonicalTimeEqual(
		actual.EndTime,
		expected.EndTime,
	) && canonicalTimeEqual(
		actual.AsOfTime,
		expected.AsOfTime,
	)
}

func bucketsEqual(
	actual Bucket,
	expected Bucket,
) bool {
	return actual.Key == expected.Key &&
		actual.Sequence == expected.Sequence &&
		canonicalTimeEqual(
			actual.StartTime,
			expected.StartTime,
		) && canonicalTimeEqual(
		actual.EndTime,
		expected.EndTime,
	)
}

func exclusionsEqual(
	actual Exclusion,
	expected Exclusion,
) bool {
	return actual.Reason == expected.Reason &&
		canonicalTimeEqual(
			actual.StartTime,
			expected.StartTime,
		) && canonicalTimeEqual(
		actual.EndTime,
		expected.EndTime,
	)
}
