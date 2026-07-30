package projectionroutefrequency

import (
	"errors"
	"testing"
	"time"
)

func TestConfigValidateRejectsIncoherentPolicyTargets(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*Config)
		wantError error
	}{
		{
			name: "history window",
			mutate: func(config *Config) {
				config.HistoryWindow = 0
			},
			wantError: ErrHistoryWindowInvalid,
		},
		{
			name: "recent window exceeds history",
			mutate: func(config *Config) {
				config.RecentWindow = config.HistoryWindow + time.Hour
			},
			wantError: ErrRecentWindowInvalid,
		},
		{
			name: "distinct-day target exceeds observations",
			mutate: func(config *Config) {
				config.TargetDistinctDayCount = config.TargetObservationCount + 1
			},
			wantError: ErrDistinctDayCountInvalid,
		},
		{
			name: "recent target exceeds observations",
			mutate: func(config *Config) {
				config.TargetRecentObservationCount = config.TargetObservationCount + 1
			},
			wantError: ErrRecentObservationCountInvalid,
		},
		{
			name: "zero route confidence",
			mutate: func(config *Config) {
				config.MinimumRouteConfidenceScore = 0
			},
			wantError: ErrMinimumRouteConfidenceInvalid,
		},
		{
			name: "zero usable score",
			mutate: func(config *Config) {
				config.MinimumUsableScore = 0
			},
			wantError: ErrScoreThresholdInvalid,
		},
		{
			name: "zero component weight",
			mutate: func(config *Config) {
				config.DistinctDayWeight += config.ObservationCountWeight
				config.ObservationCountWeight = 0
			},
			wantError: ErrComponentWeightInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := validRouteFrequencyConfig()
			test.mutate(&config)
			if err := config.Validate(); !errors.Is(err, test.wantError) {
				t.Fatalf("Validate() error = %v, want %v", err, test.wantError)
			}
		})
	}
}

func TestEvaluateRejectsHistoryWindowMismatch(t *testing.T) {
	evaluator := newRouteFrequencyEvaluator(t)
	history := validRouteHistory()
	history.WindowStart = history.WindowStart.Add(time.Hour)

	_, err := evaluator.Evaluate(validRouteFrequencyRoute(), history)
	if !errors.Is(err, ErrRouteHistoryWindowMismatch) {
		t.Fatalf("Evaluate() error = %v", err)
	}
}

func TestEvaluateReportsAllBlockingReasons(t *testing.T) {
	evaluator := newRouteFrequencyEvaluator(t)
	route := validRouteFrequencyRoute()
	route.Confidence = validRouteFrequencyConfidence(0.5, "route", 2)
	history := validRouteHistory()
	history.ObservationCount = 1
	history.DistinctFlightCount = 1
	history.DistinctDayCount = 1
	history.RecentObservationCount = 0
	history.LastObservedAt = history.AsOfTime.Add(-8 * 24 * time.Hour)

	result, err := evaluator.Evaluate(route, history)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Decision != DecisionBlocked || result.Usable {
		t.Fatalf("unexpected decision: %#v", result)
	}
	for _, code := range []string{
		"route_confidence_below_minimum",
		"route_observation_count_below_minimum",
		"route_distinct_day_count_below_minimum",
		"recent_route_observation_count_below_minimum",
		"latest_route_observation_too_old",
		"route_frequency_score_below_minimum",
	} {
		if !hasRouteFrequencyNotice(result.Limitations, code) {
			t.Fatalf("missing limitation %q in %#v", code, result.Limitations)
		}
	}
}

func TestResultValidateRejectsWeightedScoreMismatch(t *testing.T) {
	result, err := newRouteFrequencyEvaluator(t).Evaluate(
		validRouteFrequencyRoute(),
		validRouteHistory(),
	)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	result.Score -= 0.01
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted a weighted-score mismatch")
	}
}

func TestResultValidateRejectsZeroComponentWeight(t *testing.T) {
	result, err := newRouteFrequencyEvaluator(t).Evaluate(
		validRouteFrequencyRoute(),
		validRouteHistory(),
	)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	transferred := result.Components[0].Weight
	result.Components[0].Weight = 0
	result.Components[1].Weight += transferred
	result.Score = 0
	for _, component := range result.Components {
		result.Score += component.Score * component.Weight
	}
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted a zero component weight")
	}
}
