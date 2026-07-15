package projectionroutefrequency

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestEvaluateAllowsWellSupportedRoute(
	t *testing.T,
) {
	evaluator := newRouteFrequencyEvaluator(t)
	route := validRouteFrequencyRoute()
	history := validRouteHistory()

	first, err := evaluator.Evaluate(
		route,
		history,
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}
	second, err := evaluator.Evaluate(
		route,
		history,
	)
	if err != nil {
		t.Fatalf(
			"second Evaluate() error = %v",
			err,
		)
	}

	if first.Decision != DecisionAllowed ||
		!first.Usable ||
		first.Score <
			validRouteFrequencyConfig().
				CompleteScoreMinimum {
		t.Fatalf(
			"unexpected allowed result: %#v",
			first,
		)
	}
	if first.RouteKey != "UBBB>LTBA" ||
		first.InputFingerprint !=
			second.InputFingerprint {
		t.Fatalf(
			"route key or deterministic fingerprint invalid: %#v",
			first,
		)
	}
	if err := first.Validate(); err != nil {
		t.Fatalf(
			"result validation error = %v",
			err,
		)
	}
}

func TestEvaluateBlocksLowObservationCount(
	t *testing.T,
) {
	evaluator := newRouteFrequencyEvaluator(t)
	history := validRouteHistory()
	history.ObservationCount = 3
	history.DistinctFlightCount = 3
	history.RecentObservationCount = 2

	result, err := evaluator.Evaluate(
		validRouteFrequencyRoute(),
		history,
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Decision != DecisionBlocked ||
		result.Usable ||
		!hasRouteFrequencyNotice(
			result.Limitations,
			"route_observation_count_below_minimum",
		) {
		t.Fatalf(
			"low observation count did not block: %#v",
			result,
		)
	}
}

func TestEvaluateBlocksStaleLatestObservation(
	t *testing.T,
) {
	evaluator := newRouteFrequencyEvaluator(t)
	history := validRouteHistory()
	history.LastObservedAt =
		history.AsOfTime.Add(
			-8 * 24 * time.Hour,
		)

	result, err := evaluator.Evaluate(
		validRouteFrequencyRoute(),
		history,
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Decision != DecisionBlocked ||
		!hasRouteFrequencyNotice(
			result.Limitations,
			"latest_route_observation_too_old",
		) {
		t.Fatalf(
			"stale route observation did not block: %#v",
			result,
		)
	}
}

func TestEvaluateReturnsLimitedForModerateSupport(
	t *testing.T,
) {
	config := validRouteFrequencyConfig()
	config.CompleteScoreMinimum = 0.99
	evaluator, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	result, err := evaluator.Evaluate(
		validRouteFrequencyRoute(),
		validRouteHistory(),
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Decision != DecisionLimited ||
		!result.Usable ||
		!hasRouteFrequencyNotice(
			result.Limitations,
			"route_frequency_support_limited",
		) {
		t.Fatalf(
			"moderate support did not produce limited result: %#v",
			result,
		)
	}
}

func TestEvaluateBlocksIncompleteRoute(
	t *testing.T,
) {
	evaluator := newRouteFrequencyEvaluator(t)
	route := validRouteFrequencyRoute()
	route.Status =
		routecontract.RouteStatusPartial
	route.Origin = nil
	route.Confidence.EvidenceCount = 1

	result, err := evaluator.Evaluate(
		route,
		validRouteHistory(),
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Decision != DecisionBlocked ||
		!hasRouteFrequencyNotice(
			result.Limitations,
			"complete_route_unavailable",
		) {
		t.Fatalf(
			"incomplete route did not block: %#v",
			result,
		)
	}
}

func TestEvaluateRejectsHistoryKeyAndAsOfMismatch(
	t *testing.T,
) {
	evaluator := newRouteFrequencyEvaluator(t)
	route := validRouteFrequencyRoute()

	keyMismatch := validRouteHistory()
	keyMismatch.RouteKey = "UBBB>UGTB"
	_, err := evaluator.Evaluate(
		route,
		keyMismatch,
	)
	if !errors.Is(
		err,
		ErrRouteHistoryKeyMismatch,
	) {
		t.Fatalf(
			"route key mismatch error = %v",
			err,
		)
	}

	asOfMismatch := validRouteHistory()
	asOfMismatch.AsOfTime =
		asOfMismatch.AsOfTime.Add(time.Minute)
	asOfMismatch.WindowEnd =
		asOfMismatch.AsOfTime
	_, err = evaluator.Evaluate(
		route,
		asOfMismatch,
	)
	if !errors.Is(
		err,
		ErrRouteHistoryAsOfMismatch,
	) {
		t.Fatalf(
			"route as-of mismatch error = %v",
			err,
		)
	}
}

func newRouteFrequencyEvaluator(
	t *testing.T,
) *Evaluator {
	t.Helper()

	evaluator, err := New(
		validRouteFrequencyConfig(),
	)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	return evaluator
}

func validRouteHistory() HistorySummary {
	asOfTime := routeFrequencyAsOfTime()

	return HistorySummary{
		RouteKey: "UBBB>LTBA",

		WindowStart: asOfTime.Add(
			-60 * 24 * time.Hour,
		),
		WindowEnd: asOfTime,
		AsOfTime:  asOfTime,

		ObservationCount:       12,
		DistinctFlightCount:    10,
		DistinctDayCount:       7,
		RecentObservationCount: 5,
		LastObservedAt: asOfTime.Add(
			-24 * time.Hour,
		),

		SourceNames: []string{
			"historical-route-store",
		},
		InputFingerprint: "sha256:" +
			strings.Repeat("f", 64),
	}
}

func validRouteFrequencyRoute() routecontract.Result {
	asOfTime := routeFrequencyAsOfTime()

	return routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        routecontract.RouteStatusComplete,
		TrajectoryID:  "trajectory-001",
		FlightID:      "flight-001",
		AircraftID:    "aircraft-001",
		ICAO24:        "4A1234",
		Callsign:      "AHY123",
		Window: routecontract.RouteWindow{
			StartTime: asOfTime.Add(
				-30 * time.Minute,
			),
			EndTime:  asOfTime,
			AsOfTime: asOfTime,
		},
		Origin: &routecontract.EndpointInference{
			Role: routecontract.
				EndpointRoleOrigin,
			Airport: routecontract.AirportReference{
				ICAOCode:  "UBBB",
				Name:      "Heydar Aliyev International Airport",
				Latitude:  40.4675,
				Longitude: 50.0467,
			},
			DistanceKM: 5,
			Confidence: validRouteFrequencyConfidence(
				0.9,
				"origin",
				1,
			),
			Evidence: []routecontract.Evidence{
				validRouteFrequencyEvidence(
					asOfTime,
					"origin",
				),
			},
		},
		Destination: &routecontract.EndpointInference{
			Role: routecontract.
				EndpointRoleDestination,
			Airport: routecontract.AirportReference{
				ICAOCode:  "LTBA",
				Name:      "Istanbul Airport",
				Latitude:  40.9769,
				Longitude: 28.8146,
			},
			DistanceKM: 6,
			Confidence: validRouteFrequencyConfidence(
				0.9,
				"destination",
				1,
			),
			Evidence: []routecontract.Evidence{
				validRouteFrequencyEvidence(
					asOfTime,
					"destination",
				),
			},
		},
		Summary: routecontract.RouteSummary{
			GreatCircleDistanceKM: 1760,
			SameAirport:           false,
		},
		Confidence: validRouteFrequencyConfidence(
			0.9,
			"route",
			2,
		),
		Provenance: routecontract.Provenance{
			ResolverVersion: "route-resolver-test-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("a", 64),
			TrajectoryUpdatedAt: asOfTime,
			SourceNames: []string{
				"test-source",
			},
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}
}

func validRouteFrequencyConfidence(
	score float64,
	code string,
	evidenceCount int,
) routecontract.Confidence {
	return routecontract.Confidence{
		Score: score,
		Level: routecontract.
			ConfidenceLevelForScore(
				score,
			),
		EvidenceCount: evidenceCount,
		Reasons: []routecontract.ConfidenceReason{
			{
				Code:         code,
				Message:      "Route confidence reason.",
				Contribution: score,
			},
		},
	}
}

func validRouteFrequencyEvidence(
	asOfTime time.Time,
	summary string,
) routecontract.Evidence {
	return routecontract.Evidence{
		Type: routecontract.
			EvidenceTypeTrajectoryEndpointProximity,
		SourceName:    "test-source",
		SourceVersion: "test-source-v1",
		Score:         0.9,
		Weight:        1,
		ObservedAt:    asOfTime,
		Summary:       summary,
	}
}

func routeFrequencyAsOfTime() time.Time {
	return time.Date(
		2026,
		time.July,
		16,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}

func hasRouteFrequencyNotice(
	items []Notice,
	code string,
) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}

	return false
}
