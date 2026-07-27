package historicalwindow

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestBuildRejectsNilContext(
	t *testing.T,
) {
	_, err := Build(
		nil,
		validReviewRequest(),
	)
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf(
			"Build() error = %v, want %v",
			err,
			ErrContextRequired,
		)
	}
}

func TestBuildPreservesMidGenerationCancellation(
	t *testing.T,
) {
	ctx := &cancelAfterErrorChecks{
		cancelAt: 4,
	}
	request := validReviewRequest()
	request.EndTime = request.StartTime.Add(
		10 * time.Hour,
	)
	request.AsOfTime = request.EndTime

	_, err := Build(ctx, request)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Build() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestNextBoundaryRejectsUnalignedInput(
	t *testing.T,
) {
	_, err := NextBoundary(
		time.Date(
			2026,
			time.July,
			1,
			10,
			15,
			0,
			0,
			time.UTC,
		),
		historicalcontract.GranularityHour,
	)
	if !errors.Is(
		err,
		ErrBoundarySequenceInvalid,
	) {
		t.Fatalf(
			"NextBoundary() error = %v, want %v",
			err,
			ErrBoundarySequenceInvalid,
		)
	}
}

func TestBuildEnforcesWeeklyLimitBeyondDurationRange(
	t *testing.T,
) {
	startTime := alignedReviewWeek(
		t,
		time.Date(
			1800,
			time.January,
			1,
			0,
			0,
			0,
			0,
			time.UTC,
		),
	)
	endTime := alignedReviewWeek(
		t,
		startTime.AddDate(400, 0, 0),
	)

	_, err := Build(
		context.Background(),
		Request{
			StartTime:          startTime,
			EndTime:            endTime,
			AsOfTime:           endTime,
			Granularity:        historicalcontract.GranularityWeek,
			MaximumBucketCount: 16_000,
		},
	)

	var countErr *BucketCountExceededError
	if !errors.As(err, &countErr) ||
		countErr.Count != 16_001 ||
		countErr.Maximum != 16_000 {
		t.Fatalf(
			"unexpected bucket count error: %#v",
			err,
		)
	}
}

func TestBuildPreviousWindowUsesCalendarBucketsBeyondDurationRange(
	t *testing.T,
) {
	startTime := alignedReviewWeek(
		t,
		time.Date(
			1800,
			time.January,
			1,
			0,
			0,
			0,
			0,
			time.UTC,
		),
	)
	endTime := alignedReviewWeek(
		t,
		startTime.AddDate(400, 0, 0),
	)

	plan, err := Build(
		context.Background(),
		Request{
			StartTime:          startTime,
			EndTime:            endTime,
			AsOfTime:           endTime,
			Granularity:        historicalcontract.GranularityWeek,
			MaximumBucketCount: 25_000,
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	expectedPreviousStart := plan.EffectiveWindow.
		StartTime.AddDate(
		0,
		0,
		-7*len(plan.Buckets),
	)
	if plan.PreviousWindow == nil ||
		!plan.PreviousWindow.StartTime.Equal(
			expectedPreviousStart,
		) ||
		!plan.PreviousWindow.EndTime.Equal(
			plan.EffectiveWindow.StartTime,
		) {
		t.Fatalf(
			"unexpected previous window: %#v",
			plan.PreviousWindow,
		)
	}
}

func TestPlanFingerprintIgnoresExecutionLimit(
	t *testing.T,
) {
	request := validReviewRequest()
	request.MaximumBucketCount = 2
	first, err := Build(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf("first Build() error = %v", err)
	}

	request.MaximumBucketCount = 100
	second, err := Build(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf("second Build() error = %v", err)
	}

	if first.Fingerprint != second.Fingerprint {
		t.Fatalf(
			"execution limit changed semantic fingerprint: %q %q",
			first.Fingerprint,
			second.Fingerprint,
		)
	}
}

func TestValidatePlanRejectsDerivedFieldTampering(
	t *testing.T,
) {
	plan, err := Build(
		context.Background(),
		validReviewRequest(),
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	plan.Buckets[0].Key = "tampered"
	plan.Buckets[0].Sequence = 99
	if err := ValidatePlan(plan); !errors.Is(
		err,
		ErrPlanIntegrityInvalid,
	) {
		t.Fatalf(
			"ValidatePlan() error = %v, want %v",
			err,
			ErrPlanIntegrityInvalid,
		)
	}
}

func TestCanonicalizePlanRebuildsDerivedEvidence(
	t *testing.T,
) {
	plan, err := Build(
		context.Background(),
		validReviewRequest(),
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	original := plan.Clone()

	plan.Fingerprint = "sha256:" +
		strings.Repeat("0", 64)
	plan.Buckets[0].Key = "tampered"
	plan.Buckets[0].Sequence = 99
	plan.PreviousWindow.StartTime =
		plan.PreviousWindow.StartTime.Add(time.Hour)

	canonical, err := CanonicalizePlan(plan)
	if err != nil {
		t.Fatalf("CanonicalizePlan() error = %v", err)
	}
	if err := compareCanonicalPlan(
		canonical,
		original,
	); err != nil {
		t.Fatalf(
			"canonical plan differs from original: %v",
			err,
		)
	}
}

func TestDurationRejectsReversedIntervals(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		1,
		1,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(-time.Hour)

	if (Bucket{
		StartTime: startTime,
		EndTime:   endTime,
	}).Duration() != 0 {
		t.Fatal("reversed bucket duration must be zero")
	}
	if (Exclusion{
		StartTime: startTime,
		EndTime:   endTime,
	}).Duration() != 0 {
		t.Fatal("reversed exclusion duration must be zero")
	}
}

func TestBuildEntireFutureExclusionStartsAtRequestedStart(
	t *testing.T,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		1,
		10,
		0,
		0,
		0,
		time.UTC,
	)
	startTime := asOfTime.Add(time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	plan, err := Build(
		context.Background(),
		Request{
			StartTime:   startTime,
			EndTime:     endTime,
			AsOfTime:    asOfTime,
			Granularity: historicalcontract.GranularityHour,
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(plan.Exclusions) != 1 ||
		!plan.Exclusions[0].StartTime.Equal(startTime) ||
		!plan.Exclusions[0].EndTime.Equal(endTime) {
		t.Fatalf("unexpected exclusions: %#v", plan.Exclusions)
	}
}

func TestCustomGranularityRemainsSupported(
	t *testing.T,
) {
	request := validReviewRequest()
	request.Granularity =
		historicalcontract.GranularityCustom

	plan, err := Build(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(plan.Buckets) != 1 {
		t.Fatalf(
			"custom bucket count = %d, want 1",
			len(plan.Buckets),
		)
	}
}

func validReviewRequest() Request {
	startTime := time.Date(
		2026,
		time.July,
		1,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	return Request{
		StartTime:   startTime,
		EndTime:     startTime.Add(2 * time.Hour),
		AsOfTime:    startTime.Add(3 * time.Hour),
		Granularity: historicalcontract.GranularityHour,
	}
}

func alignedReviewWeek(
	t *testing.T,
	value time.Time,
) time.Time {
	t.Helper()

	aligned, err := FloorBoundary(
		value,
		historicalcontract.GranularityWeek,
	)
	if err != nil {
		t.Fatalf("FloorBoundary() error = %v", err)
	}
	return aligned
}

type cancelAfterErrorChecks struct {
	calls    int
	cancelAt int
}

func (ctx *cancelAfterErrorChecks) Deadline() (
	time.Time,
	bool,
) {
	return time.Time{}, false
}

func (ctx *cancelAfterErrorChecks) Done() <-chan struct{} {
	return nil
}

func (ctx *cancelAfterErrorChecks) Err() error {
	ctx.calls++
	if ctx.calls >= ctx.cancelAt {
		return context.Canceled
	}
	return nil
}

func (ctx *cancelAfterErrorChecks) Value(any) any {
	return nil
}
