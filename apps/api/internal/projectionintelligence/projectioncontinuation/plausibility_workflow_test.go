package projectioncontinuation

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func TestProjectFallsBackWhenPlausibilityRemovesRequiredSupport(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	fallback := config.
		FallbackProjector.(*fallbackProjectorStub)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := continuationTestRequest()
	for candidateIndex := range request.Candidates {
		request.Candidates[candidateIndex].
			Points[5].Longitude += 10
		request.Candidates[candidateIndex].
			Points[6].Longitude += 10
	}

	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if fallback.calls != 1 ||
		result.Method.Name !=
			projectionbaseline.MethodName ||
		!hasFallbackReason(
			result.Limitations,
			"historical_continuation_plausibility_support_insufficient",
		) {
		t.Fatalf(
			"unexpected plausibility fallback: %#v",
			result,
		)
	}
}

func TestProjectMarksLimitedWhenPlausibilityFiltersPartialSupport(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	config.MinimumPointSupport = 1
	config.MinimumAltitudeSupport = 1
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := continuationTestRequest()
	request.Candidates[0].
		Points[5].Longitude += 10
	request.Candidates[0].
		Points[6].Longitude += 10

	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if result.Status !=
		projectioncontract.ResultStatusLimited ||
		!hasProjectionLimitationWithScope(
			result.Limitations,
			"historical_continuation_plausibility_filtered",
			"support",
		) {
		t.Fatalf(
			"unexpected partially filtered result: %#v",
			result,
		)
	}
}

func hasProjectionLimitationWithScope(
	items []projectioncontract.Limitation,
	code string,
	scope string,
) bool {
	for _, item := range items {
		if item.Code == code &&
			item.Scope == scope {
			return true
		}
	}

	return false
}

func TestPlausibilityPolicyChangesInputFingerprint(
	t *testing.T,
) {
	firstConfig := validContinuationConfig(t)
	first, err := New(firstConfig)
	if err != nil {
		t.Fatalf("first New() error = %v", err)
	}

	secondConfig := validContinuationConfig(t)
	secondConfig.PlausibilityPolicy =
		DefaultPlausibilityPolicy()
	secondConfig.PlausibilityPolicy.
		MaximumHorizontalSpeedMPS = 350
	second, err := New(secondConfig)
	if err != nil {
		t.Fatalf("second New() error = %v", err)
	}

	request := continuationTestRequest()
	firstResult, err := first.Project(request)
	if err != nil {
		t.Fatalf("first Project() error = %v", err)
	}
	secondResult, err := second.Project(request)
	if err != nil {
		t.Fatalf("second Project() error = %v", err)
	}

	if firstResult.Provenance.InputFingerprint ==
		secondResult.Provenance.InputFingerprint {
		t.Fatal(
			"plausibility policy did not change input fingerprint",
		)
	}
}
