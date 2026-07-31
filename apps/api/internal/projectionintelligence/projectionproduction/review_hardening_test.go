package projectionproduction

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionarrival"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func TestComposeUsesOneAuthorizedHorizonPlan(t *testing.T) {
	fixture := newProductionFixture()
	composer, err := New(fixture.config)
	if err != nil {
		t.Fatal(err)
	}
	result, err := composer.Compose(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	planner := fixture.config.HorizonPlanner.(*fakeHorizonPlanner)
	if planner.calls != 1 {
		t.Fatalf("horizon planner calls = %d, want 1", planner.calls)
	}
	if fixture.historical.plan.Fingerprint != result.HorizonPlan.Fingerprint ||
		fixture.historical.plan.Fingerprint != planner.plan.Fingerprint {
		t.Fatal("historical projector did not receive the authorized plan snapshot")
	}
}

func TestComposeRejectsRouteFromAnotherTrajectory(t *testing.T) {
	fixture := newProductionFixture()
	fixture.request.Route.TrajectoryID = "other-trajectory"
	composer, _ := New(fixture.config)
	result, err := composer.Compose(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Strategy != StrategyKinematic || result.FallbackReason != "route_contract_invalid" ||
		fixture.historical.calls != 0 || fixture.arrival.calls != 0 {
		t.Fatalf("route mismatch was not isolated: %#v", result)
	}
}

func TestComposeRejectsFutureRouteEvidence(t *testing.T) {
	fixture := newProductionFixture()
	fixture.request.Route.Window.AsOfTime = fixture.request.AsOfTime.Add(time.Minute)
	fixture.request.Route.Window.EndTime = fixture.request.Route.Window.AsOfTime
	fixture.request.Route.GeneratedAt = fixture.request.GeneratedAt.Add(time.Minute)
	composer, _ := New(fixture.config)
	result, err := composer.Compose(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	if result.FallbackReason != "route_contract_invalid" || fixture.historical.calls != 0 {
		t.Fatalf("future route evidence was accepted: %#v", result)
	}
}

func TestComposeBindsSelectionPatternAndFreshness(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*productionFixture)
	}{
		{
			name: "selection request mismatch",
			mutate: func(f *productionFixture) {
				f.selector.result.CurrentTrajectoryID = "other-trajectory"
			},
		},
		{
			name: "pattern selection mismatch",
			mutate: func(f *productionFixture) {
				f.pattern.result.SelectedTrajectoryIDs = []string{"historical-a", "historical-z"}
			},
		},
		{
			name: "freshness pattern mismatch",
			mutate: func(f *productionFixture) {
				f.freshness.result.SourcePatternFingerprint = "sha256:" + strings.Repeat("a", 64)
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newProductionFixture()
			testCase.mutate(&fixture)
			fixture.config.DependencyFailurePolicy = DependencyFailureReturnError
			composer, _ := New(fixture.config)
			_, err := composer.Compose(fixture.request)
			if !errors.Is(err, ErrEvidenceBindingInvalid) {
				t.Fatalf("error = %v, want evidence binding sentinel", err)
			}
		})
	}
}

func TestComposeBindsRouteHistoryAndFrequencyToPlan(t *testing.T) {
	fixture := newProductionFixture()
	fixture.request.RouteHistory.AsOfTime = fixture.request.AsOfTime.Add(time.Minute)
	fixture.request.RouteHistory.WindowEnd = fixture.request.RouteHistory.AsOfTime
	fixture.config.DependencyFailurePolicy = DependencyFailureReturnError
	composer, _ := New(fixture.config)
	_, err := composer.Compose(fixture.request)
	if !errors.Is(err, ErrEvidenceBindingInvalid) {
		t.Fatalf("error = %v, want evidence binding sentinel", err)
	}

	fixture = newProductionFixture()
	fixture.frequency.result.HistoryInputFingerprint = "sha256:" + strings.Repeat("9", 64)
	fixture.config.DependencyFailurePolicy = DependencyFailureReturnError
	composer, _ = New(fixture.config)
	_, err = composer.Compose(fixture.request)
	if !errors.Is(err, ErrEvidenceBindingInvalid) {
		t.Fatalf("frequency error = %v, want evidence binding sentinel", err)
	}
}

func TestComposeRejectsHistoricalProjectionPostconditionDrift(t *testing.T) {
	fixture := newProductionFixture()
	fixture.historical.result.TrajectoryID = "other-trajectory"
	composer, _ := New(fixture.config)
	result, err := composer.Compose(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Strategy != StrategyKinematic ||
		result.FallbackReason != "historical_projection_postcondition_failed" ||
		fixture.kinematic.calls != 1 {
		t.Fatalf("historical identity drift was not rejected: %#v", result)
	}
}

func TestComposeRejectsUnavailableHistoricalProjection(t *testing.T) {
	fixture := newProductionFixture()
	fixture.historical.result.Status = projectioncontract.ResultStatusUnavailable
	fixture.historical.result.Points = nil
	fixture.historical.result.Confidence = projectioncontract.Confidence{
		Score: 0,
		Level: projectioncontract.ConfidenceLevelNone,
	}
	fixture.historical.result.Limitations = []projectioncontract.Limitation{{
		Code: "unavailable", Message: "Historical projection unavailable.", Scope: "result",
	}}
	composer, _ := New(fixture.config)
	result, err := composer.Compose(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Strategy != StrategyKinematic || fixture.kinematic.calls != 1 {
		t.Fatalf("unavailable historical result was selected: %#v", result)
	}
}

func TestComposeDefensivelyClonesDependencyInputs(t *testing.T) {
	fixture := newProductionFixture()
	fixture.selector.mutate = true
	originalLatitude := fixture.request.CurrentTrajectory.Points[0].Latitude
	originalCandidateID := fixture.request.HistoricalCandidates[0].ID
	composer, _ := New(fixture.config)
	_, err := composer.Compose(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	if fixture.request.CurrentTrajectory.Points[0].Latitude != originalLatitude ||
		fixture.request.HistoricalCandidates[0].ID != originalCandidateID {
		t.Fatal("dependency mutated caller-owned request evidence")
	}
	if fixture.pattern.candidates[0].ID != originalCandidateID {
		t.Fatal("selector mutation leaked into the next production stage")
	}
}

func TestComposePreservesUnderlyingDependencyError(t *testing.T) {
	fixture := newProductionFixture()
	cause := errors.New("selector root cause")
	fixture.selector.err = cause
	fixture.config.DependencyFailurePolicy = DependencyFailureReturnError
	composer, _ := New(fixture.config)
	_, err := composer.Compose(fixture.request)
	if !errors.Is(err, ErrNeighborSelectionFailed) || !errors.Is(err, cause) {
		t.Fatalf("dependency error chain = %v", err)
	}
}

func TestComposePublishesLimitedEvidenceNotice(t *testing.T) {
	fixture := newProductionFixture()
	fixture.freshness.result = validProductionLimitedFreshness(productionTestAsOfTime())
	fixture.config.FreshnessLimitedPolicy = LimitedEvidenceAllow
	composer, _ := New(fixture.config)
	result, err := composer.Compose(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	if !hasProductionNotice(result.Notices, "historical_projection_authorized_with_limited_freshness") {
		t.Fatalf("limited evidence notice missing: %#v", result.Notices)
	}
}

func TestRequestFingerprintCoversCandidatesOnFallback(t *testing.T) {
	first := newProductionFixture()
	first.selector.err = errProductionDependency
	composer, _ := New(first.config)
	one, err := composer.Compose(first.request)
	if err != nil {
		t.Fatal(err)
	}

	second := newProductionFixture()
	second.selector.err = errProductionDependency
	second.request.HistoricalCandidates[0].ID = "different-candidate"
	composer, _ = New(second.config)
	two, err := composer.Compose(second.request)
	if err != nil {
		t.Fatal(err)
	}
	if one.InputFingerprint == two.InputFingerprint {
		t.Fatal("different candidate pools produced the same request fingerprint")
	}
}

func TestCompositionFingerprintBindsPublishedProjection(t *testing.T) {
	fixture := newProductionFixture()
	composer, _ := New(fixture.config)
	result, err := composer.Compose(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	result.Projection.Points[0].Position.Latitude += 0.1
	if err := result.Validate(); err == nil {
		t.Fatal("tampered projection passed composition fingerprint validation")
	}
}

type mutatingArrivalProjectionEstimator struct{}

func (mutatingArrivalProjectionEstimator) Estimate(
	request projectionarrival.Request,
) (projectioncontract.Result, error) {
	result := request.Projection.Clone()
	result.Points[0].Position.Latitude += 1
	return result, nil
}

func TestArrivalAdapterRejectsPositionProjectionMutation(t *testing.T) {
	adapter, err := NewArrivalAdapter(mutatingArrivalProjectionEstimator{})
	if err != nil {
		t.Fatal(err)
	}
	fixture := newProductionFixture()
	_, err = adapter.EstimateArrival(projectionarrival.Request{
		Projection:        fixture.historical.result.Clone(),
		Route:             fixture.request.Route.Clone(),
		CurrentTrajectory: fixture.request.CurrentTrajectory.Clone(),
		GeneratedAt:       fixture.request.GeneratedAt,
	})
	if !errors.Is(err, ErrArrivalProjectionMutation) {
		t.Fatalf("arrival mutation error = %v", err)
	}
}

func hasProductionNotice(items []Notice, code string) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}
	return false
}
