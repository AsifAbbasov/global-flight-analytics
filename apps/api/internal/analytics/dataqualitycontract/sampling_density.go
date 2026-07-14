package dataqualitycontract

import (
	"fmt"
	"math"
	"sort"
	"time"
)

func EvaluateSamplingDensity(input SamplingDensityInput) (SamplingDensity, error) {
	if input.WindowStart.IsZero() {
		return SamplingDensity{}, ErrWindowStartRequired
	}
	if input.WindowEnd.IsZero() {
		return SamplingDensity{}, ErrWindowEndRequired
	}
	if !input.WindowEnd.After(input.WindowStart) {
		return SamplingDensity{}, ErrWindowRangeInvalid
	}
	if input.ExpectedInterval <= 0 {
		return SamplingDensity{}, ErrExpectedIntervalInvalid
	}

	totalIntervals := intervalCount(
		input.WindowEnd.Sub(input.WindowStart),
		input.ExpectedInterval,
	)
	covered := make(map[int]struct{}, totalIntervals)
	observations := append([]time.Time(nil), input.ObservationTimes...)
	sort.Slice(observations, func(i, j int) bool {
		return observations[i].Before(observations[j])
	})

	for _, observedAt := range observations {
		if observedAt.Before(input.WindowStart) || !observedAt.Before(input.WindowEnd) {
			return SamplingDensity{}, fmt.Errorf(
				"%w: observed_at=%s window=[%s,%s)",
				ErrObservationOutsideWindow,
				observedAt.UTC().Format(timeFormat),
				input.WindowStart.UTC().Format(timeFormat),
				input.WindowEnd.UTC().Format(timeFormat),
			)
		}
		index := int(observedAt.Sub(input.WindowStart) / input.ExpectedInterval)
		covered[index] = struct{}{}
	}

	coveredCount := len(covered)
	score := 0.0
	if totalIntervals > 0 {
		score = float64(coveredCount) / float64(totalIntervals)
	}

	result := SamplingDensity{
		Score:                clampUnit(score),
		ObservedSampleCount:  len(observations),
		ExpectedSampleCount:  totalIntervals,
		CoveredIntervalCount: coveredCount,
		TotalIntervalCount:   totalIntervals,
		DuplicateSampleCount: len(observations) - coveredCount,
		WindowStart:          input.WindowStart.UTC(),
		WindowEnd:            input.WindowEnd.UTC(),
		ExpectedInterval:     input.ExpectedInterval,
		Explanation:          "Sampling density measures covered expected intervals, not raw point volume.",
	}
	if err := result.Validate(); err != nil {
		return SamplingDensity{}, err
	}
	return result, nil
}

func intervalCount(duration, interval time.Duration) int {
	count := int(duration / interval)
	if duration%interval != 0 {
		count++
	}
	return count
}

func (value SamplingDensity) Validate() error {
	if math.IsNaN(value.Score) || math.IsInf(value.Score, 0) ||
		value.Score < 0 || value.Score > 1 {
		return fmt.Errorf("%w: %f", ErrSamplingDensityScoreInvalid, value.Score)
	}
	if value.WindowStart.IsZero() {
		return ErrWindowStartRequired
	}
	if value.WindowEnd.IsZero() {
		return ErrWindowEndRequired
	}
	if !value.WindowEnd.After(value.WindowStart) {
		return ErrWindowRangeInvalid
	}
	if value.ExpectedInterval <= 0 {
		return ErrExpectedIntervalInvalid
	}
	if value.ObservedSampleCount < 0 || value.ExpectedSampleCount <= 0 ||
		value.CoveredIntervalCount < 0 || value.TotalIntervalCount <= 0 ||
		value.CoveredIntervalCount > value.TotalIntervalCount ||
		value.DuplicateSampleCount < 0 ||
		value.CoveredIntervalCount+value.DuplicateSampleCount != value.ObservedSampleCount ||
		value.ExpectedSampleCount != value.TotalIntervalCount {
		return ErrSamplingCountsInvalid
	}
	return nil
}
