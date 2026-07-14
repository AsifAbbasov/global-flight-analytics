package dataqualitycontract

import (
	"errors"
	"testing"
	"time"
)

func TestEvaluateFreshnessBoundaries(t *testing.T) {
	evaluatedAt := time.Date(2026, time.July, 15, 1, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		age    time.Duration
		status FreshnessStatus
		score  float64
	}{
		{"within expected interval", 30 * time.Second, FreshnessStatusFresh, 1},
		{"between expected and stale", 3 * time.Minute, FreshnessStatusAging, 0.5},
		{"at stale boundary", 5 * time.Minute, FreshnessStatusStale, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := EvaluateFreshness(FreshnessInput{
				ObservedAt:       evaluatedAt.Add(-test.age),
				EvaluatedAt:      evaluatedAt,
				ExpectedInterval: time.Minute,
				StaleAfter:       5 * time.Minute,
			})
			if err != nil {
				t.Fatalf("evaluate freshness: %v", err)
			}
			if result.Status != test.status || result.Score != test.score {
				t.Fatalf("expected %s %.2f, got %s %.2f", test.status, test.score, result.Status, result.Score)
			}
		})
	}
}

func TestEvaluateFreshnessRejectsFutureObservation(t *testing.T) {
	now := time.Now().UTC()
	_, err := EvaluateFreshness(FreshnessInput{
		ObservedAt:       now.Add(time.Second),
		EvaluatedAt:      now,
		ExpectedInterval: time.Minute,
		StaleAfter:       5 * time.Minute,
	})
	if !errors.Is(err, ErrObservationInFuture) {
		t.Fatalf("expected future observation error, got %v", err)
	}
}
