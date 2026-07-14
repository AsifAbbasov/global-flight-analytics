package dataqualitycontract

import (
	"fmt"
	"math"
	"time"
)

const timeFormat = time.RFC3339Nano

func EvaluateFreshness(input FreshnessInput) (Freshness, error) {
	if input.ObservedAt.IsZero() {
		return Freshness{}, ErrObservedAtRequired
	}
	if input.EvaluatedAt.IsZero() {
		return Freshness{}, ErrEvaluatedAtRequired
	}
	if input.ObservedAt.After(input.EvaluatedAt) {
		return Freshness{}, fmt.Errorf(
			"%w: observed_at=%s evaluated_at=%s",
			ErrObservationInFuture,
			input.ObservedAt.UTC().Format(timeFormat),
			input.EvaluatedAt.UTC().Format(timeFormat),
		)
	}
	if input.ExpectedInterval <= 0 {
		return Freshness{}, fmt.Errorf(
			"%w: %s",
			ErrExpectedIntervalInvalid,
			input.ExpectedInterval,
		)
	}
	if input.StaleAfter < input.ExpectedInterval {
		return Freshness{}, fmt.Errorf(
			"%w: expected_interval=%s stale_after=%s",
			ErrStaleAfterInvalid,
			input.ExpectedInterval,
			input.StaleAfter,
		)
	}

	age := input.EvaluatedAt.Sub(input.ObservedAt)
	score, status, explanation := calculateFreshness(
		age,
		input.ExpectedInterval,
		input.StaleAfter,
	)

	result := Freshness{
		Score:                   score,
		Status:                  status,
		AgeSeconds:              age.Seconds(),
		ExpectedIntervalSeconds: input.ExpectedInterval.Seconds(),
		StaleAfterSeconds:       input.StaleAfter.Seconds(),
		ObservedAt:              input.ObservedAt.UTC(),
		EvaluatedAt:             input.EvaluatedAt.UTC(),
		Explanation:             explanation,
	}
	if err := result.Validate(); err != nil {
		return Freshness{}, err
	}
	return result, nil
}

func calculateFreshness(
	age time.Duration,
	expectedInterval time.Duration,
	staleAfter time.Duration,
) (float64, FreshnessStatus, string) {
	if age <= expectedInterval {
		return 1, FreshnessStatusFresh,
			"The newest observation is within the expected publication interval."
	}
	if age >= staleAfter {
		return 0, FreshnessStatusStale,
			"The newest observation is older than the configured stale threshold."
	}

	remaining := float64(staleAfter - age)
	decayWindow := float64(staleAfter - expectedInterval)
	score := remaining / decayWindow
	return clampUnit(score), FreshnessStatusAging,
		"The newest observation is older than the expected interval but has not reached the stale threshold."
}

func (value Freshness) Validate() error {
	if math.IsNaN(value.Score) || math.IsInf(value.Score, 0) ||
		value.Score < 0 || value.Score > 1 {
		return fmt.Errorf("%w: %f", ErrFreshnessScoreInvalid, value.Score)
	}
	if !value.Status.IsKnown() {
		return fmt.Errorf("%w: %q", ErrFreshnessStatusInvalid, value.Status)
	}
	if value.ObservedAt.IsZero() {
		return ErrObservedAtRequired
	}
	if value.EvaluatedAt.IsZero() {
		return ErrEvaluatedAtRequired
	}
	if value.ObservedAt.After(value.EvaluatedAt) {
		return ErrObservationInFuture
	}
	if value.ExpectedIntervalSeconds <= 0 {
		return ErrExpectedIntervalInvalid
	}
	if value.StaleAfterSeconds < value.ExpectedIntervalSeconds {
		return ErrStaleAfterInvalid
	}
	return nil
}

func (status FreshnessStatus) IsKnown() bool {
	switch status {
	case FreshnessStatusFresh,
		FreshnessStatusAging,
		FreshnessStatusStale,
		FreshnessStatusUnknown:
		return true
	default:
		return false
	}
}

func clampUnit(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
