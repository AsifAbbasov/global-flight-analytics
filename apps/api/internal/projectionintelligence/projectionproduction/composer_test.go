package projectionproduction

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionfreshness"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
)

func TestComposeSelectsHistoricalStrategyAndAttachesArrival(
	t *testing.T,
) {
	fixture := newProductionFixture()
	composer, err := New(fixture.config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	first, err := composer.Compose(
		fixture.request,
	)
	if err != nil {
		t.Fatalf(
			"Compose() error = %v",
			err,
		)
	}
	second, err := composer.Compose(
		fixture.request,
	)
	if err != nil {
		t.Fatalf(
			"second Compose() error = %v",
			err,
		)
	}

	if first.Strategy !=
		StrategyHistoricalNeighbor ||
		first.FallbackReason != "" ||
		first.Projection.Method.Name !=
			"local_historical_neighbor_continuation" ||
		first.ArrivalStatus !=
			ArrivalStatusAttached ||
		first.Projection.Arrival == nil {
		t.Fatalf(
			"unexpected historical composition: %#v",
			first,
		)
	}
	if first.NeighborSelection == nil ||
		first.PatternConfidence == nil ||
		first.Freshness == nil ||
		first.RouteFrequency == nil {
		t.Fatalf(
			"historical evidence trace is incomplete: %#v",
			first,
		)
	}
	if fixture.historical.calls != 2 ||
		fixture.kinematic.calls != 0 ||
		fixture.arrival.calls != 2 {
		t.Fatalf(
			"unexpected dependency calls: historical=%d kinematic=%d arrival=%d",
			fixture.historical.calls,
			fixture.kinematic.calls,
			fixture.arrival.calls,
		)
	}
	if first.InputFingerprint !=
		second.InputFingerprint {
		t.Fatal(
			"deterministic production input produced different fingerprints",
		)
	}
	if err := first.Validate(); err != nil {
		t.Fatalf(
			"result validation error = %v",
			err,
		)
	}
}

func TestComposeFallsBackWhenFreshnessBlocks(
	t *testing.T,
) {
	fixture := newProductionFixture()
	fixture.freshness.result.Decision =
		projectionfreshness.DecisionBlocked
	fixture.freshness.result.Usable = false
	fixture.freshness.result.Limitations =
		[]projectionfreshness.Notice{
			{
				Code:    "stale",
				Message: "Historical pattern is stale.",
			},
		}

	composer, err := New(fixture.config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}
	result, err := composer.Compose(
		fixture.request,
	)
	if err != nil {
		t.Fatalf(
			"Compose() error = %v",
			err,
		)
	}

	if result.Strategy != StrategyKinematic ||
		result.FallbackReason !=
			"pattern_freshness_guard_blocked" ||
		result.Projection.Method.Name !=
			projectionbaseline.MethodName ||
		fixture.historical.calls != 0 ||
		fixture.kinematic.calls != 1 {
		t.Fatalf(
			"freshness fallback failed: result=%#v historical=%d kinematic=%d",
			result,
			fixture.historical.calls,
			fixture.kinematic.calls,
		)
	}
}

func TestComposeFallsBackWhenRouteHistoryIsMissing(
	t *testing.T,
) {
	fixture := newProductionFixture()
	fixture.request.RouteHistory = nil

	composer, err := New(fixture.config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}
	result, err := composer.Compose(
		fixture.request,
	)
	if err != nil {
		t.Fatalf(
			"Compose() error = %v",
			err,
		)
	}

	if result.Strategy != StrategyKinematic ||
		result.FallbackReason !=
			"route_history_unavailable" ||
		result.RouteFrequency != nil ||
		fixture.frequency.calls != 0 {
		t.Fatalf(
			"missing route history did not fall back safely: %#v",
			result,
		)
	}
}

func TestComposeLimitedGuardPolicyIsExplicit(
	t *testing.T,
) {
	fixture := newProductionFixture()
	fixture.freshness.result.Decision =
		projectionfreshness.DecisionLimited
	fixture.freshness.result.Limitations =
		[]projectionfreshness.Notice{
			{
				Code:    "freshness_limited",
				Message: "Freshness is usable but limited.",
			},
		}

	rejectComposer, err := New(fixture.config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}
	rejected, err := rejectComposer.Compose(
		fixture.request,
	)
	if err != nil {
		t.Fatalf(
			"reject Compose() error = %v",
			err,
		)
	}
	if rejected.Strategy != StrategyKinematic {
		t.Fatalf(
			"reject policy selected %q",
			rejected.Strategy,
		)
	}

	allowFixture := newProductionFixture()
	allowFixture.freshness.result.Decision =
		projectionfreshness.DecisionLimited
	allowFixture.freshness.result.Limitations =
		[]projectionfreshness.Notice{
			{
				Code:    "freshness_limited",
				Message: "Freshness is usable but limited.",
			},
		}
	allowFixture.config.FreshnessLimitedPolicy =
		LimitedEvidenceAllow
	allowComposer, err := New(
		allowFixture.config,
	)
	if err != nil {
		t.Fatalf(
			"allow New() error = %v",
			err,
		)
	}
	allowed, err := allowComposer.Compose(
		allowFixture.request,
	)
	if err != nil {
		t.Fatalf(
			"allow Compose() error = %v",
			err,
		)
	}
	if allowed.Strategy !=
		StrategyHistoricalNeighbor {
		t.Fatalf(
			"allow policy selected %q",
			allowed.Strategy,
		)
	}
}

func TestComposeRouteFrequencyBlockedUsesKinematicFallback(
	t *testing.T,
) {
	fixture := newProductionFixture()
	fixture.frequency.result.Decision =
		projectionroutefrequency.DecisionBlocked
	fixture.frequency.result.Usable = false
	fixture.frequency.result.Limitations =
		[]projectionroutefrequency.Notice{
			{
				Code:    "low_frequency",
				Message: "Route support is insufficient.",
			},
		}

	composer, err := New(fixture.config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}
	result, err := composer.Compose(
		fixture.request,
	)
	if err != nil {
		t.Fatalf(
			"Compose() error = %v",
			err,
		)
	}

	if result.Strategy != StrategyKinematic ||
		result.FallbackReason !=
			"route_frequency_guard_blocked" ||
		fixture.historical.calls != 0 {
		t.Fatalf(
			"route-frequency fallback failed: %#v",
			result,
		)
	}
}

func TestComposeDependencyFailurePolicyControlsFailure(
	t *testing.T,
) {
	fallbackFixture := newProductionFixture()
	fallbackFixture.selector.err =
		errProductionDependency
	fallbackComposer, err := New(
		fallbackFixture.config,
	)
	if err != nil {
		t.Fatalf(
			"fallback New() error = %v",
			err,
		)
	}
	fallbackResult, err := fallbackComposer.Compose(
		fallbackFixture.request,
	)
	if err != nil {
		t.Fatalf(
			"fallback Compose() error = %v",
			err,
		)
	}
	if fallbackResult.Strategy !=
		StrategyKinematic ||
		fallbackResult.FallbackReason !=
			"historical_neighbor_selection_failed" {
		t.Fatalf(
			"dependency fallback result = %#v",
			fallbackResult,
		)
	}

	errorFixture := newProductionFixture()
	errorFixture.selector.err =
		errProductionDependency
	errorFixture.config.DependencyFailurePolicy =
		DependencyFailureReturnError
	errorComposer, err := New(errorFixture.config)
	if err != nil {
		t.Fatalf(
			"error New() error = %v",
			err,
		)
	}
	_, err = errorComposer.Compose(
		errorFixture.request,
	)
	if !errors.Is(
		err,
		ErrNeighborSelectionFailed,
	) {
		t.Fatalf(
			"error policy result = %v",
			err,
		)
	}
}

func TestComposeRecognizesHistoricalProjectorInternalFallback(
	t *testing.T,
) {
	fixture := newProductionFixture()
	fixture.historical.result =
		fixture.kinematic.result.Clone()

	composer, err := New(fixture.config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}
	result, err := composer.Compose(
		fixture.request,
	)
	if err != nil {
		t.Fatalf(
			"Compose() error = %v",
			err,
		)
	}

	if result.Strategy != StrategyKinematic ||
		result.FallbackReason !=
			"historical_projector_internal_fallback" ||
		fixture.kinematic.calls != 0 {
		t.Fatalf(
			"internal fallback was not preserved: %#v",
			result,
		)
	}
}

func TestComposeArrivalFailurePolicyPreservesOrReturnsError(
	t *testing.T,
) {
	preserveFixture := newProductionFixture()
	preserveFixture.arrival.err =
		errProductionDependency
	preserveComposer, err := New(
		preserveFixture.config,
	)
	if err != nil {
		t.Fatalf(
			"preserve New() error = %v",
			err,
		)
	}
	preserved, err := preserveComposer.Compose(
		preserveFixture.request,
	)
	if err != nil {
		t.Fatalf(
			"preserve Compose() error = %v",
			err,
		)
	}
	if preserved.ArrivalStatus !=
		ArrivalStatusFailed ||
		preserved.Projection.Arrival != nil {
		t.Fatalf(
			"arrival failure did not preserve projection: %#v",
			preserved,
		)
	}

	errorFixture := newProductionFixture()
	errorFixture.arrival.err =
		errProductionDependency
	errorFixture.config.ArrivalFailurePolicy =
		ArrivalFailureReturnError
	errorComposer, err := New(errorFixture.config)
	if err != nil {
		t.Fatalf(
			"error New() error = %v",
			err,
		)
	}
	_, err = errorComposer.Compose(
		errorFixture.request,
	)
	if !errors.Is(
		err,
		ErrArrivalEstimationFailed,
	) {
		t.Fatalf(
			"arrival error policy result = %v",
			err,
		)
	}
}

func TestComposeInvalidRouteFallsBackAndSkipsArrival(
	t *testing.T,
) {
	fixture := newProductionFixture()
	fixture.request.Route.TrajectoryID = ""

	composer, err := New(fixture.config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}
	result, err := composer.Compose(
		fixture.request,
	)
	if err != nil {
		t.Fatalf(
			"Compose() error = %v",
			err,
		)
	}

	if result.Strategy != StrategyKinematic ||
		result.FallbackReason !=
			"route_contract_invalid" ||
		result.ArrivalStatus !=
			ArrivalStatusSkipped ||
		fixture.arrival.calls != 0 {
		t.Fatalf(
			"invalid route fallback failed: %#v",
			result,
		)
	}
}
