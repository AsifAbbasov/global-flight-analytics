package dataqualitycontract

import (
	"errors"
	"testing"
	"time"
)

func TestSamplingDensityCountsCoveredIntervals(t *testing.T) {
	start := time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC)
	result, err := EvaluateSamplingDensity(SamplingDensityInput{
		WindowStart:      start,
		WindowEnd:        start.Add(4 * time.Minute),
		ExpectedInterval: time.Minute,
		ObservationTimes: []time.Time{
			start.Add(10 * time.Second),
			start.Add(20 * time.Second),
			start.Add(2*time.Minute + 10*time.Second),
		},
	})
	if err != nil {
		t.Fatalf("evaluate sampling density: %v", err)
	}
	if result.ObservedSampleCount != 3 || result.CoveredIntervalCount != 2 ||
		result.DuplicateSampleCount != 1 || result.TotalIntervalCount != 4 ||
		result.Score != 0.5 {
		t.Fatalf("unexpected sampling density: %#v", result)
	}
}

func TestSamplingDensityRejectsWindowEndObservation(t *testing.T) {
	start := time.Now().UTC().Truncate(time.Second)
	_, err := EvaluateSamplingDensity(SamplingDensityInput{
		WindowStart:      start,
		WindowEnd:        start.Add(time.Minute),
		ExpectedInterval: time.Second,
		ObservationTimes: []time.Time{start.Add(time.Minute)},
	})
	if !errors.Is(err, ErrObservationOutsideWindow) {
		t.Fatalf("expected outside-window error, got %v", err)
	}
}
