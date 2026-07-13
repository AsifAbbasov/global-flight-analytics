package executor

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/confidencereport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/scopeguard"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

var (
	ErrExecutorRequired = errors.New(
		"analytics executor is required",
	)
	ErrResultOperationRequired = errors.New(
		"analytical result operation is required",
	)
)

const (
	FailureCodeAnalyticalOperationFailed = "analytical_operation_failed"
	FailureCodeOperationCanceled         = "analytical_operation_canceled"
	FailureCodeOperationDeadlineExceeded = "analytical_operation_deadline_exceeded"
	FailureCodeConfidenceUnavailable     = "confidence_unavailable"
	NoticeCodeLowConfidence              = "confidence_low"
)

type TrajectoryCalculation[T any] struct {
	Value       T
	Factors     []confidencereport.Factor
	Warnings    []analyticalresult.Notice
	Limitations []analyticalresult.Notice
}

type TrajectoryResultOperation[T any] func(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (TrajectoryCalculation[T], error)

type FailureMapper func(
	err error,
) analyticalresult.Failure

type TrajectoryResultRequest[T any] struct {
	Trajectory    trajectory.FlightTrajectory
	Capability    trajectoryeligibility.Capability
	Sources       []analyticalresult.Source
	Operation     TrajectoryResultOperation[T]
	FailureMapper FailureMapper
}

type TrajectoryExecution[T any] struct {
	Result           analyticalresult.Result[T]
	ScopeDecision    scopeguard.Decision
	ConfidenceReport *confidencereport.Report
}

func (
	execution TrajectoryExecution[T],
) IsDenied() bool {
	return execution.Result.Status ==
		analyticalresult.StatusDenied
}

func (
	execution TrajectoryExecution[T],
) IsFailed() bool {
	return execution.Result.Status ==
		analyticalresult.StatusFailed
}

func ExecuteTrajectoryResult[T any](
	ctx context.Context,
	executor *Executor,
	request TrajectoryResultRequest[T],
) (TrajectoryExecution[T], error) {
	if executor == nil {
		return TrajectoryExecution[T]{},
			ErrExecutorRequired
	}

	if request.Operation == nil {
		return TrajectoryExecution[T]{},
			ErrResultOperationRequired
	}

	decision, err := executor.scopeGuard.Require(
		request.Trajectory,
		request.Capability,
	)
	if err != nil {
		if errors.Is(
			err,
			scopeguard.ErrDenied,
		) {
			result, resultErr :=
				analyticalresult.NewDenied[T](
					decision,
					request.Sources,
				)
			if resultErr != nil {
				return TrajectoryExecution[T]{},
					fmt.Errorf(
						"build denied analytical result: %w",
						resultErr,
					)
			}

			return TrajectoryExecution[T]{
				Result:        result,
				ScopeDecision: decision,
			}, nil
		}

		return TrajectoryExecution[T]{},
			fmt.Errorf(
				"require analytical scope: %w",
				err,
			)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	calculation, operationErr :=
		request.Operation(
			ctx,
			request.Trajectory,
		)
	if operationErr != nil {
		failureMapper := request.FailureMapper
		if failureMapper == nil {
			failureMapper =
				DefaultFailureMapper
		}

		failure := failureMapper(
			operationErr,
		)
		eligibility :=
			analyticalresult.
				EligibilityFromScopeDecision(
					decision,
				)

		result, resultErr :=
			analyticalresult.NewFailed[T](
				failure,
				&eligibility,
				request.Sources,
				nil,
				nil,
				decision.EvaluatedAt,
			)
		if resultErr != nil {
			return TrajectoryExecution[T]{},
				fmt.Errorf(
					"build failed analytical result: %w",
					resultErr,
				)
		}

		return TrajectoryExecution[T]{
			Result:        result,
			ScopeDecision: decision,
		}, nil
	}

	report, err :=
		executor.confidenceEvaluator.
			Evaluate(
				confidencereport.Request{
					Factors:  calculation.Factors,
					Warnings: calculation.Warnings,
					Limitations: calculation.
						Limitations,
					EvaluatedAt: decision.
						EvaluatedAt,
				},
			)
	if err != nil {
		return TrajectoryExecution[T]{},
			fmt.Errorf(
				"evaluate analytical confidence: %w",
				err,
			)
	}

	reportCopy := report.Clone()
	eligibility :=
		analyticalresult.
			EligibilityFromScopeDecision(
				decision,
			)

	if report.Level ==
		analyticalresult.
			ConfidenceLevelNone {
		failure := analyticalresult.Failure{
			Code:      FailureCodeConfidenceUnavailable,
			Message:   "Analytical confidence is zero; the calculated value cannot be published.",
			Retriable: false,
		}

		result, resultErr :=
			analyticalresult.NewFailed[T](
				failure,
				&eligibility,
				request.Sources,
				report.Warnings,
				append(
					[]analyticalresult.Notice(nil),
					report.Limitations...,
				),
				decision.EvaluatedAt,
			)
		if resultErr != nil {
			return TrajectoryExecution[T]{},
				fmt.Errorf(
					"build confidence-unavailable analytical result: %w",
					resultErr,
				)
		}

		return TrajectoryExecution[T]{
			Result:           result,
			ScopeDecision:    decision,
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
		!containsNoticeCode(
			warnings,
			limitations,
			NoticeCodeLowConfidence,
		) {
		limitations = append(
			limitations,
			analyticalresult.Notice{
				Code:    NoticeCodeLowConfidence,
				Message: "Confidence is low; use this analytical result with caution.",
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

		reportCopy.Limitations = append(
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
				&eligibility,
				request.Sources,
				warnings,
				limitations,
				decision.EvaluatedAt,
			)
		if resultErr != nil {
			return TrajectoryExecution[T]{},
				fmt.Errorf(
					"build limited analytical result: %w",
					resultErr,
				)
		}

		return TrajectoryExecution[T]{
			Result:           result,
			ScopeDecision:    decision,
			ConfidenceReport: &reportCopy,
		}, nil
	}

	result, resultErr :=
		analyticalresult.NewComplete(
			calculation.Value,
			confidence,
			&eligibility,
			request.Sources,
			decision.EvaluatedAt,
		)
	if resultErr != nil {
		return TrajectoryExecution[T]{},
			fmt.Errorf(
				"build complete analytical result: %w",
				resultErr,
			)
	}

	return TrajectoryExecution[T]{
		Result:           result,
		ScopeDecision:    decision,
		ConfidenceReport: &reportCopy,
	}, nil
}

func DefaultFailureMapper(
	err error,
) analyticalresult.Failure {
	code := FailureCodeAnalyticalOperationFailed
	retriable := false

	switch {
	case errors.Is(
		err,
		context.Canceled,
	):
		code =
			FailureCodeOperationCanceled
		retriable = true

	case errors.Is(
		err,
		context.DeadlineExceeded,
	):
		code =
			FailureCodeOperationDeadlineExceeded
		retriable = true
	}

	message := strings.TrimSpace(
		err.Error(),
	)
	if message == "" {
		message =
			"Analytical operation failed."
	}

	return analyticalresult.Failure{
		Code:      code,
		Message:   message,
		Retriable: retriable,
	}
}

func containsNoticeCode(
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
