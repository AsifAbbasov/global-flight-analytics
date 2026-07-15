package projectionevaluation

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

var (
	ErrAggregateGeneratedAtInvalid = errors.New(
		"aggregate generated-at time is required",
	)
	ErrAggregateInputInvalid = errors.New(
		"projection evaluation aggregate input is invalid",
	)
	ErrAggregateResultInvalid = errors.New(
		"projection evaluation aggregate result is invalid",
	)
)

type methodAccumulator struct {
	method projectioncontract.Method

	evaluationCount  int
	completeCount    int
	partialCount     int
	unavailableCount int

	forecastPointCount  int
	evaluatedPointCount int

	horizontalErrors  []float64
	horizontalCovered int

	altitudeErrors            []float64
	verticalCoverageEvaluated int
	verticalCovered           int

	arrivalErrors  []float64
	arrivalCovered int
}

func Aggregate(
	results []Result,
	generatedAt time.Time,
) (AggregateResult, error) {
	generatedAt = generatedAt.UTC()
	if generatedAt.IsZero() {
		return AggregateResult{},
			ErrAggregateGeneratedAtInvalid
	}

	if len(results) == 0 {
		result := AggregateResult{
			Version:         AggregateVersion,
			Status:          StatusUnavailable,
			EvaluationCount: 0,
			MethodCount:     0,
			Limitations: []Notice{
				{
					Code:    "projection_evaluations_unavailable",
					Message: "No projection evaluations were supplied for aggregation.",
				},
			},
			InputFingerprint: aggregateFingerprint(
				nil,
				generatedAt,
			),
			GeneratedAt: generatedAt,
		}
		if err := result.Validate(); err != nil {
			return AggregateResult{},
				fmt.Errorf(
					"%w: %v",
					ErrAggregateResultInvalid,
					err,
				)
		}
		return result, nil
	}

	accumulators := make(
		map[string]*methodAccumulator,
	)
	aggregateStatus := StatusComplete
	limitations := make(
		[]Notice,
		0,
		2,
	)

	for index, evaluation := range results {
		if err := evaluation.Validate(); err != nil {
			return AggregateResult{},
				fmt.Errorf(
					"%w at index %d: %v",
					ErrAggregateInputInvalid,
					index,
					err,
				)
		}
		if evaluation.EvaluatedAt.After(
			generatedAt,
		) {
			return AggregateResult{},
				fmt.Errorf(
					"%w at index %d: evaluation time exceeds aggregate generation time",
					ErrAggregateInputInvalid,
					index,
				)
		}

		key := methodKey(
			evaluation.ProjectionMethod,
		)
		accumulator := accumulators[key]
		if accumulator == nil {
			accumulator =
				&methodAccumulator{
					method: evaluation.
						ProjectionMethod,
				}
			accumulators[key] = accumulator
		}

		accumulator.evaluationCount++
		switch evaluation.Status {
		case StatusComplete:
			accumulator.completeCount++
		case StatusPartial:
			accumulator.partialCount++
			aggregateStatus = StatusPartial
		case StatusUnavailable:
			accumulator.unavailableCount++
			aggregateStatus = StatusPartial
		}

		accumulator.forecastPointCount +=
			evaluation.Position.
				ForecastPointCount
		accumulator.evaluatedPointCount +=
			evaluation.Position.
				EvaluatedPointCount

		for _, point := range evaluation.Points {
			accumulator.horizontalErrors =
				append(
					accumulator.
						horizontalErrors,
					point.HorizontalErrorM,
				)
			if point.
				WithinHorizontalUncertainty {
				accumulator.
					horizontalCovered++
			}

			if point.
				AltitudeAbsoluteErrorM != nil {
				accumulator.altitudeErrors =
					append(
						accumulator.
							altitudeErrors,
						*point.
							AltitudeAbsoluteErrorM,
					)
			}
			if point.
				WithinVerticalUncertainty != nil {
				accumulator.
					verticalCoverageEvaluated++
				if *point.
					WithinVerticalUncertainty {
					accumulator.
						verticalCovered++
				}
			}
		}

		if evaluation.Arrival.Available {
			accumulator.arrivalErrors =
				append(
					accumulator.
						arrivalErrors,
					evaluation.Arrival.
						EstimatedAbsoluteErrorSeconds,
				)
			if evaluation.Arrival.
				IntervalCoveredActual {
				accumulator.
					arrivalCovered++
			}
		}
	}

	keys := make(
		[]string,
		0,
		len(accumulators),
	)
	for key := range accumulators {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	methods := make(
		[]MethodSummary,
		0,
		len(keys),
	)
	for _, key := range keys {
		accumulator :=
			accumulators[key]
		summary := MethodSummary{
			MethodName:    accumulator.method.Name,
			MethodVersion: accumulator.method.Version,
			DecisionClass: accumulator.method.
				DecisionClass,

			EvaluationCount: accumulator.
				evaluationCount,
			CompleteEvaluationCount: accumulator.completeCount,
			PartialEvaluationCount:  accumulator.partialCount,
			UnavailableEvaluationCount: accumulator.
				unavailableCount,

			ForecastPointCount: accumulator.
				forecastPointCount,
			EvaluatedPointCount: accumulator.
				evaluatedPointCount,

			AltitudeEvaluatedPointCount: len(
				accumulator.
					altitudeErrors,
			),
			ArrivalEvaluationCount: len(
				accumulator.
					arrivalErrors,
			),
		}

		if summary.ForecastPointCount > 0 {
			summary.PointCoverageRatio =
				float64(
					summary.
						EvaluatedPointCount,
				) /
					float64(
						summary.
							ForecastPointCount,
					)
		}
		if len(
			accumulator.horizontalErrors,
		) > 0 {
			summary.MeanHorizontalErrorM =
				mean(
					accumulator.
						horizontalErrors,
				)
			summary.MedianHorizontalErrorM =
				percentileNearestRank(
					accumulator.
						horizontalErrors,
					0.50,
				)
			summary.P95HorizontalErrorM =
				percentileNearestRank(
					accumulator.
						horizontalErrors,
					0.95,
				)
			summary.HorizontalRMSEM =
				rootMeanSquare(
					accumulator.
						horizontalErrors,
				)
			summary.
				HorizontalUncertaintyCoverageRatio =
				float64(
					accumulator.
						horizontalCovered,
				) /
					float64(
						len(
							accumulator.
								horizontalErrors,
						),
					)
		}
		if len(
			accumulator.altitudeErrors,
		) > 0 {
			summary.
				MeanAltitudeAbsoluteErrorM =
				mean(
					accumulator.
						altitudeErrors,
				)
			summary.AltitudeRMSEM =
				rootMeanSquare(
					accumulator.
						altitudeErrors,
				)
		}
		if accumulator.
			verticalCoverageEvaluated > 0 {
			summary.
				VerticalUncertaintyCoverageRatio =
				float64(
					accumulator.
						verticalCovered,
				) /
					float64(
						accumulator.
							verticalCoverageEvaluated,
					)
		}
		if len(
			accumulator.arrivalErrors,
		) > 0 {
			summary.
				MeanArrivalAbsoluteErrorSeconds =
				mean(
					accumulator.
						arrivalErrors,
				)
			summary.
				ArrivalIntervalCoverageRatio =
				float64(
					accumulator.
						arrivalCovered,
				) /
					float64(
						len(
							accumulator.
								arrivalErrors,
						),
					)
		}

		methods = append(
			methods,
			summary,
		)
	}

	if aggregateStatus == StatusPartial {
		limitations = append(
			limitations,
			Notice{
				Code:    "aggregate_contains_partial_or_unavailable_evaluations",
				Message: "At least one projection evaluation was partial or unavailable.",
			},
		)
	}

	result := AggregateResult{
		Version:         AggregateVersion,
		Status:          aggregateStatus,
		EvaluationCount: len(results),
		MethodCount:     len(methods),
		Methods:         methods,
		Limitations: normalizeNotices(
			limitations,
		),
		InputFingerprint: aggregateFingerprint(
			results,
			generatedAt,
		),
		GeneratedAt: generatedAt,
	}

	if err := result.Validate(); err != nil {
		return AggregateResult{},
			fmt.Errorf(
				"%w: %v",
				ErrAggregateResultInvalid,
				err,
			)
	}

	return result.Clone(), nil
}

func methodKey(
	method projectioncontract.Method,
) string {
	return strings.TrimSpace(
		method.Name,
	) + "\x00" +
		strings.TrimSpace(
			method.Version,
		)
}
