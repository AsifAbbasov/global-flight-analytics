package metricexecution

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/confidencereport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/executor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type metricCalculation[T any] struct {
	Value       T
	Factors     []confidencereport.Factor
	Warnings    []analyticalresult.Notice
	Limitations []analyticalresult.Notice
}

type trajectoryMetricOperation[T any] func(
	ctx context.Context,
	allowed []trajectory.FlightTrajectory,
	evaluatedAt time.Time,
) (metricCalculation[T], error)

type trajectoryMetricPreparation func(
	allowed []trajectory.FlightTrajectory,
) (
	[]trajectory.FlightTrajectory,
	[]analyticalresult.Notice,
)

type snapshotMetricOperation[T any] func(
	ctx context.Context,
	evaluatedAt time.Time,
) (metricCalculation[T], error)

func executeTrajectoryMetric[T any](
	ctx context.Context,
	service *Service,
	metricID string,
	capability trajectoryeligibility.Capability,
	trajectories []trajectory.FlightTrajectory,
	metadata PublicationMetadata,
	preparation trajectoryMetricPreparation,
	operation trajectoryMetricOperation[T],
) (Execution[T], error) {
	if service == nil ||
		service.executor == nil {
		return Execution[T]{},
			ErrExecutorRequired
	}

	if operation == nil {
		return Execution[T]{},
			ErrMetricOperationRequired
	}

	filtered, err :=
		service.executor.FilterTrajectories(
			trajectories,
			capability,
		)
	if err != nil {
		return Execution[T]{},
			fmt.Errorf(
				"filter metric trajectories: %w",
				err,
			)
	}

	if preparation != nil {
		prepared, preparationWarnings :=
			preparation(filtered.Allowed)
		filtered.Allowed = prepared
		metadata.Warnings = mergeNotices(
			metadata.Warnings,
			preparationWarnings,
		)
	}

	contributorCount :=
		filtered.AllowedCount() +
			filtered.DeniedCount()

	scope := buildScopeSummary(
		filtered,
		capability,
		contributorCount,
	)

	if contributorCount > 0 &&
		filtered.AllowedCount() == 0 {
		decision, decisionErr :=
			aggregateDeniedDecision(
				filtered,
				capability,
			)
		if decisionErr != nil {
			return Execution[T]{},
				decisionErr
		}

		result, resultErr :=
			analyticalresult.NewDenied[T](
				decision,
				metadata.Sources,
			)
		if resultErr != nil {
			return Execution[T]{},
				fmt.Errorf(
					"build denied metric result: %w",
					resultErr,
				)
		}

		result, resultErr = attachDataQuality(
			result,
			metadata.DataQuality,
		)
		if resultErr != nil {
			return Execution[T]{}, resultErr
		}

		return Execution[T]{
			MetricID: metricID,
			Result:   result,
			Scope:    scope,
		}, nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return buildFailedExecution[T](
			metricID,
			scope,
			nil,
			metadata,
			err,
		)
	}

	calculation, operationErr :=
		operation(
			ctx,
			filtered.Allowed,
			filtered.EvaluatedAt,
		)
	if operationErr != nil {
		return buildFailedExecution[T](
			metricID,
			scope,
			allowedEligibility(
				scope,
			),
			metadata,
			operationErr,
		)
	}

	factors, automaticLimitations :=
		buildTrajectoryConfidence(
			filtered.Allowed,
			contributorCount,
			filtered.EvaluatedAt,
			capability,
		)

	calculation.Factors = append(
		factors,
		calculation.Factors...,
	)

	if filtered.DeniedCount() > 0 {
		automaticLimitations = append(
			automaticLimitations,
			analyticalresult.Notice{
				Code: NoticeCodeIneligibleTrajectoriesExcluded,
				Message: fmt.Sprintf(
					"%d ineligible trajectory contributors were excluded from the metric.",
					filtered.DeniedCount(),
				),
			},
		)
	}

	calculation.Warnings = mergeNotices(
		metadata.Warnings,
		calculation.Warnings,
	)
	calculation.Limitations = mergeNotices(
		metadata.Limitations,
		automaticLimitations,
		calculation.Limitations,
	)

	return publishExecution(
		service,
		metricID,
		scope,
		allowedEligibility(
			scope,
		),
		metadata.Sources,
		metadata.DataQuality,
		calculation,
	)
}

func executeSnapshotMetric[T any](
	ctx context.Context,
	service *Service,
	metricID string,
	capability trajectoryeligibility.Capability,
	metadata PublicationMetadata,
	operation snapshotMetricOperation[T],
) (Execution[T], error) {
	if service == nil ||
		service.executor == nil {
		return Execution[T]{},
			ErrExecutorRequired
	}

	if operation == nil {
		return Execution[T]{},
			ErrMetricOperationRequired
	}

	filtered, err :=
		service.executor.FilterTrajectories(
			nil,
			capability,
		)
	if err != nil {
		return Execution[T]{},
			fmt.Errorf(
				"obtain metric evaluation time: %w",
				err,
			)
	}

	scope := buildScopeSummary(
		filtered,
		capability,
		0,
	)

	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return buildFailedExecution[T](
			metricID,
			scope,
			nil,
			metadata,
			err,
		)
	}

	calculation, operationErr :=
		operation(
			ctx,
			filtered.EvaluatedAt,
		)
	if operationErr != nil {
		return buildFailedExecution[T](
			metricID,
			scope,
			nil,
			metadata,
			operationErr,
		)
	}

	calculation.Warnings = mergeNotices(
		metadata.Warnings,
		calculation.Warnings,
	)
	calculation.Limitations = mergeNotices(
		metadata.Limitations,
		calculation.Limitations,
	)

	return publishExecution(
		service,
		metricID,
		scope,
		nil,
		metadata.Sources,
		metadata.DataQuality,
		calculation,
	)
}

func publishExecution[T any](
	service *Service,
	metricID string,
	scope ScopeSummary,
	eligibility *analyticalresult.Eligibility,
	sources []analyticalresult.Source,
	dataQuality *dataqualitycontract.Report,
	calculation metricCalculation[T],
) (Execution[T], error) {
	report, err :=
		service.executor.
			ConfidenceEvaluator().
			Evaluate(
				confidencereport.Request{
					Factors:  calculation.Factors,
					Warnings: calculation.Warnings,
					Limitations: calculation.
						Limitations,
					EvaluatedAt: scope.
						EvaluatedAt,
				},
			)
	if err != nil {
		return Execution[T]{},
			fmt.Errorf(
				"evaluate metric confidence: %w",
				err,
			)
	}

	reportCopy := report.Clone()

	if report.Level ==
		analyticalresult.
			ConfidenceLevelNone {
		failure := analyticalresult.Failure{
			Code: executor.
				FailureCodeConfidenceUnavailable,
			Message:   "Metric confidence is zero; the calculated value cannot be published.",
			Retriable: false,
		}

		result, resultErr :=
			analyticalresult.NewFailed[T](
				failure,
				eligibility,
				sources,
				report.Warnings,
				report.Limitations,
				scope.EvaluatedAt,
			)
		if resultErr != nil {
			return Execution[T]{},
				fmt.Errorf(
					"build confidence-unavailable metric result: %w",
					resultErr,
				)
		}

		result, resultErr = attachDataQuality(
			result,
			dataQuality,
		)
		if resultErr != nil {
			return Execution[T]{}, resultErr
		}

		return Execution[T]{
			MetricID:         metricID,
			Result:           result,
			Scope:            scope,
			ConfidenceReport: &reportCopy,
		}, nil
	}

	warnings := append(
		[]analyticalresult.Notice(nil),
		report.Warnings...,
	)
	limitations := append(
		[]analyticalresult.Notice(nil),
		report.Limitations...,
	)

	if report.Level ==
		analyticalresult.
			ConfidenceLevelLow &&
		!containsNotice(
			warnings,
			limitations,
			executor.NoticeCodeLowConfidence,
		) {
		limitations = append(
			limitations,
			analyticalresult.Notice{
				Code: executor.
					NoticeCodeLowConfidence,
				Message: "Confidence is low; use this metric result with caution.",
			},
		)

		sort.SliceStable(
			limitations,
			func(
				left int,
				right int,
			) bool {
				return limitations[left].Code <
					limitations[right].Code
			},
		)

		reportCopy.Limitations =
			append(
				[]analyticalresult.Notice(nil),
				limitations...,
			)
	}

	confidence :=
		report.AnalyticalConfidence()

	if len(warnings) > 0 ||
		len(limitations) > 0 {
		result, resultErr :=
			analyticalresult.NewLimited(
				calculation.Value,
				confidence,
				eligibility,
				sources,
				warnings,
				limitations,
				scope.EvaluatedAt,
			)
		if resultErr != nil {
			return Execution[T]{},
				fmt.Errorf(
					"build limited metric result: %w",
					resultErr,
				)
		}

		result, resultErr = attachDataQuality(
			result,
			dataQuality,
		)
		if resultErr != nil {
			return Execution[T]{}, resultErr
		}

		return Execution[T]{
			MetricID:         metricID,
			Result:           result,
			Scope:            scope,
			ConfidenceReport: &reportCopy,
		}, nil
	}

	result, resultErr :=
		analyticalresult.NewComplete(
			calculation.Value,
			confidence,
			eligibility,
			sources,
			scope.EvaluatedAt,
		)
	if resultErr != nil {
		return Execution[T]{},
			fmt.Errorf(
				"build complete metric result: %w",
				resultErr,
			)
	}

	result, resultErr = attachDataQuality(
		result,
		dataQuality,
	)
	if resultErr != nil {
		return Execution[T]{}, resultErr
	}

	return Execution[T]{
		MetricID:         metricID,
		Result:           result,
		Scope:            scope,
		ConfidenceReport: &reportCopy,
	}, nil
}

func buildFailedExecution[T any](
	metricID string,
	scope ScopeSummary,
	eligibility *analyticalresult.Eligibility,
	metadata PublicationMetadata,
	operationErr error,
) (Execution[T], error) {
	mapper := metadata.FailureMapper
	if mapper == nil {
		mapper =
			executor.DefaultFailureMapper
	}

	failure := mapper(
		operationErr,
	)

	result, err :=
		analyticalresult.NewFailed[T](
			failure,
			eligibility,
			metadata.Sources,
			metadata.Warnings,
			metadata.Limitations,
			scope.EvaluatedAt,
		)
	if err != nil {
		return Execution[T]{},
			fmt.Errorf(
				"build failed metric result: %w",
				err,
			)
	}

	result, err = attachDataQuality(
		result,
		metadata.DataQuality,
	)
	if err != nil {
		return Execution[T]{}, err
	}

	return Execution[T]{
		MetricID: metricID,
		Result:   result,
		Scope:    scope,
	}, nil
}

func allowedEligibility(
	scope ScopeSummary,
) *analyticalresult.Eligibility {
	if scope.InputCount == 0 {
		return nil
	}

	return &analyticalresult.Eligibility{
		Capability:  scope.Capability,
		Allowed:     true,
		EvaluatedAt: scope.EvaluatedAt,
	}
}

func containsNotice(
	warnings []analyticalresult.Notice,
	limitations []analyticalresult.Notice,
	code string,
) bool {
	for _, collection := range [][]analyticalresult.Notice{
		warnings,
		limitations,
	} {
		for _, notice := range collection {
			if notice.Code == code {
				return true
			}
		}
	}

	return false
}
