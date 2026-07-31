package projectionevaluation

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

var (
	ErrAggregateGeneratedAtInvalid = errors.New("aggregate generated-at time is required")
	ErrAggregateInputInvalid       = errors.New("projection evaluation aggregate input is invalid")
	ErrAggregateResultInvalid      = errors.New("projection evaluation aggregate result is invalid")
)

func Aggregate(results []Result, generatedAt time.Time) (AggregateResult, error) {
	generatedAt = generatedAt.UTC()
	if generatedAt.IsZero() {
		return AggregateResult{}, ErrAggregateGeneratedAtInvalid
	}
	if len(results) == 0 {
		result := AggregateResult{
			Version:          AggregateVersion,
			Status:           StatusUnavailable,
			Limitations:      []Notice{{Code: "projection_evaluations_unavailable", Message: "No projection evaluations were supplied for aggregation."}},
			InputFingerprint: aggregateFingerprint(nil),
			GeneratedAt:      generatedAt,
		}
		if err := result.Validate(); err != nil {
			return AggregateResult{}, fmt.Errorf("%w: %v", ErrAggregateResultInvalid, err)
		}
		return result, nil
	}

	accumulators := make(map[string]*methodAccumulator)
	aggregateStatus := StatusComplete
	for index, evaluation := range results {
		if err := evaluation.Validate(); err != nil {
			return AggregateResult{}, fmt.Errorf("%w at index %d: %v", ErrAggregateInputInvalid, index, err)
		}
		if evaluation.EvaluatedAt.After(generatedAt) {
			return AggregateResult{}, fmt.Errorf("%w at index %d: evaluation time exceeds aggregate generation time", ErrAggregateInputInvalid, index)
		}
		key := evaluationGroupKey(evaluation)
		accumulator := accumulators[key]
		if accumulator == nil {
			accumulator = &methodAccumulator{
				method:          evaluation.ProjectionMethod,
				horizonDuration: evaluation.ProjectionHorizonEndTime.Sub(evaluation.ProjectionAsOfTime),
				forecastStep:    evaluation.ForecastStep,
				policy:          evaluation.Policy,
				leadTimes:       make(map[time.Duration]*leadTimeAccumulator),
			}
			accumulators[key] = accumulator
		}
		accumulator.addEvaluation(evaluation)
		if evaluation.Status != StatusComplete {
			aggregateStatus = StatusPartial
		}
	}

	keys := make([]string, 0, len(accumulators))
	for key := range accumulators {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	methods := make([]MethodSummary, 0, len(keys))
	for _, key := range keys {
		methods = append(methods, accumulators[key].summary())
	}

	limitations := []Notice(nil)
	if aggregateStatus == StatusPartial {
		limitations = append(limitations, Notice{
			Code:    "aggregate_contains_partial_or_unavailable_evaluations",
			Message: "At least one projection evaluation was partial or unavailable; unavailable evaluations are excluded from accuracy metrics.",
		})
	}
	result := AggregateResult{
		Version:          AggregateVersion,
		Status:           aggregateStatus,
		EvaluationCount:  len(results),
		MethodCount:      len(methods),
		Methods:          methods,
		Limitations:      normalizeNotices(limitations),
		InputFingerprint: aggregateFingerprint(results),
		GeneratedAt:      generatedAt,
	}
	if err := result.Validate(); err != nil {
		return AggregateResult{}, fmt.Errorf("%w: %v", ErrAggregateResultInvalid, err)
	}
	return result.Clone(), nil
}
