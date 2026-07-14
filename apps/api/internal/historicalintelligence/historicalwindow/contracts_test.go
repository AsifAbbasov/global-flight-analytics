package historicalwindow

import (
	"regexp"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestPlanCloneDoesNotShareMutableState(
	t *testing.T,
) {
	plan := mustBuild(
		t,
		Request{
			StartTime: time.Date(
				2026,
				time.July,
				1,
				0,
				0,
				0,
				0,
				time.UTC,
			),
			EndTime: time.Date(
				2026,
				time.July,
				1,
				3,
				0,
				0,
				0,
				time.UTC,
			),
			AsOfTime: time.Date(
				2026,
				time.July,
				1,
				4,
				0,
				0,
				0,
				time.UTC,
			),
			Granularity: historicalcontract.
				GranularityHour,
		},
	)

	cloned := plan.Clone()
	cloned.Buckets[0].Key = "changed"
	cloned.EffectiveWindow.StartTime =
		time.Time{}
	cloned.PreviousWindow.StartTime =
		time.Time{}
	cloned.Exclusions = append(
		cloned.Exclusions,
		Exclusion{
			Reason: ExclusionReasonFutureAfterAsOfTime,
		},
	)

	if plan.Buckets[0].Key == "changed" ||
		plan.EffectiveWindow.StartTime.IsZero() ||
		plan.PreviousWindow.StartTime.IsZero() ||
		len(plan.Exclusions) != 0 {
		t.Fatal(
			"Plan.Clone() shared mutable state",
		)
	}
}

func TestBucketContainsUsesHalfOpenInterval(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	bucket := Bucket{
		StartTime: startTime,
		EndTime:   startTime.Add(time.Hour),
	}

	if !bucket.Contains(startTime) {
		t.Fatal(
			"bucket must include its start time",
		)
	}
	if !bucket.Contains(
		startTime.Add(59 * time.Minute),
	) {
		t.Fatal(
			"bucket must include time before its end",
		)
	}
	if bucket.Contains(bucket.EndTime) {
		t.Fatal(
			"bucket must exclude its end time",
		)
	}
}

func TestBucketAndExclusionDuration(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	bucket := Bucket{
		StartTime: startTime,
		EndTime:   startTime.Add(time.Hour),
	}
	exclusion := Exclusion{
		StartTime: startTime,
		EndTime:   startTime.Add(30 * time.Minute),
	}

	if bucket.Duration() != time.Hour {
		t.Fatalf(
			"bucket duration = %s",
			bucket.Duration(),
		)
	}
	if exclusion.Duration() !=
		30*time.Minute {
		t.Fatalf(
			"exclusion duration = %s",
			exclusion.Duration(),
		)
	}
}

func TestBucketKeyFormat(
	t *testing.T,
) {
	plan := mustBuild(
		t,
		Request{
			StartTime: time.Date(
				2026,
				time.July,
				1,
				0,
				0,
				0,
				0,
				time.UTC,
			),
			EndTime: time.Date(
				2026,
				time.July,
				1,
				1,
				0,
				0,
				0,
				time.UTC,
			),
			AsOfTime: time.Date(
				2026,
				time.July,
				1,
				2,
				0,
				0,
				0,
				time.UTC,
			),
			Granularity: historicalcontract.
				GranularityHour,
		},
	)

	pattern := regexp.MustCompile(
		`^historical-bucket-[0-9a-f]{64}$`,
	)
	if len(plan.Buckets) != 1 ||
		!pattern.MatchString(
			plan.Buckets[0].Key,
		) {
		t.Fatalf(
			"unexpected bucket key: %#v",
			plan.Buckets,
		)
	}
}
