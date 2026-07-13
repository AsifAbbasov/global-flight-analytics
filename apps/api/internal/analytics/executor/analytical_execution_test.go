package executor

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/confidencereport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/scopeguard"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type resultExecutionEvaluatorFunction func(
	item trajectory.FlightTrajectory,
	now time.Time,
) trajectoryeligibility.Evaluation

func (
	function resultExecutionEvaluatorFunction,
) Evaluate(
	item trajectory.FlightTrajectory,
	now time.Time,
) trajectoryeligibility.Evaluation {
	return function(
		item,
		now,
	)
}

func TestExecuteTrajectoryResultCreatesCompleteResult(
	t *testing.T,
) {
	executor := resultTestExecutor(
		t,
		true,
		nil,
	)
	request := completeExecutionRequest[int]()

	called := 0
	request.Operation = func(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (TrajectoryCalculation[int], error) {
		called++

		if ctx == nil {
			t.Fatal("expected non-nil operation context")
		}

		return TrajectoryCalculation[int]{
			Value:   42,
			Factors: highConfidenceFactors(),
		}, nil
	}

	execution, err := ExecuteTrajectoryResult(
		nil,
		executor,
		request,
	)
	if err != nil {
		t.Fatalf(
			"expected complete analytical execution, got %v",
			err,
		)
	}

	if called != 1 {
		t.Fatalf(
			"expected one operation call, got %d",
			called,
		)
	}

	if execution.Result.Status !=
		analyticalresult.StatusComplete ||
		!execution.Result.IsUsable() {
		t.Fatalf(
			"expected usable complete result, got %#v",
			execution.Result,
		)
	}

	if execution.Result.Value != 42 ||
		execution.Result.Confidence.Level !=
			analyticalresult.ConfidenceLevelHigh {
		t.Fatalf(
			"unexpected complete result: %#v",
			execution.Result,
		)
	}

	if execution.ConfidenceReport == nil ||
		execution.ConfidenceReport.Score != 0.90 {
		t.Fatalf(
			"expected confidence report score 0.90, got %#v",
			execution.ConfidenceReport,
		)
	}

	if execution.Result.CalculatedAt !=
		resultExecutionTestTime() ||
		execution.ScopeDecision.EvaluatedAt !=
			resultExecutionTestTime() {
		t.Fatal("expected one guard timestamp across the execution")
	}
}

func TestExecuteTrajectoryResultCreatesLimitedResultFromWarnings(
	t *testing.T,
) {
	executor := resultTestExecutor(
		t,
		true,
		nil,
	)
	request := completeExecutionRequest[string]()
	request.Operation = func(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (TrajectoryCalculation[string], error) {
		return TrajectoryCalculation[string]{
			Value:   "limited-value",
			Factors: highConfidenceFactors(),
			Warnings: []analyticalresult.Notice{
				{
					Code:    "coverage_partial",
					Message: "Observation coverage is partial.",
				},
			},
		}, nil
	}

	execution, err := ExecuteTrajectoryResult(
		context.Background(),
		executor,
		request,
	)
	if err != nil {
		t.Fatalf(
			"expected limited analytical execution, got %v",
			err,
		)
	}

	if execution.Result.Status !=
		analyticalresult.StatusLimited ||
		!execution.Result.IsUsable() {
		t.Fatalf(
			"expected usable limited result, got %#v",
			execution.Result,
		)
	}

	if len(execution.Result.Warnings) != 1 ||
		execution.Result.Warnings[0].Code !=
			"coverage_partial" {
		t.Fatalf(
			"expected warning preservation, got %#v",
			execution.Result.Warnings,
		)
	}
}

func TestExecuteTrajectoryResultAutomaticallyLimitsLowConfidence(
	t *testing.T,
) {
	executor := resultTestExecutor(
		t,
		true,
		nil,
	)
	request := completeExecutionRequest[int]()
	request.Operation = func(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (TrajectoryCalculation[int], error) {
		return TrajectoryCalculation[int]{
			Value: 7,
			Factors: []confidencereport.Factor{
				confidencereport.Evidence(
					confidencereport.
						FactorCodeTrajectoryQuality,
					1,
					0.40,
					"Trajectory quality is limited.",
				),
			},
		}, nil
	}

	execution, err := ExecuteTrajectoryResult(
		context.Background(),
		executor,
		request,
	)
	if err != nil {
		t.Fatalf(
			"expected low-confidence analytical execution, got %v",
			err,
		)
	}

	if execution.Result.Status !=
		analyticalresult.StatusLimited {
		t.Fatalf(
			"expected limited status, got %s",
			execution.Result.Status,
		)
	}

	if execution.Result.Confidence.Level !=
		analyticalresult.ConfidenceLevelLow {
		t.Fatalf(
			"expected low confidence, got %s",
			execution.Result.Confidence.Level,
		)
	}

	if !hasAnalyticalNotice(
		execution.Result.Limitations,
		NoticeCodeLowConfidence,
	) {
		t.Fatalf(
			"expected automatic low-confidence limitation, got %#v",
			execution.Result.Limitations,
		)
	}

	if execution.ConfidenceReport == nil ||
		!hasAnalyticalNotice(
			execution.ConfidenceReport.Limitations,
			NoticeCodeLowConfidence,
		) {
		t.Fatal(
			"expected confidence report to match the published limitation",
		)
	}
}

func TestExecuteTrajectoryResultCreatesDeniedResultWithoutOperation(
	t *testing.T,
) {
	executor := resultTestExecutor(
		t,
		false,
		[]trajectoryeligibility.ReasonCode{
			trajectoryeligibility.
				ReasonMissingIdentity,
		},
	)
	request := completeExecutionRequest[int]()

	called := false
	request.Operation = func(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (TrajectoryCalculation[int], error) {
		called = true

		return TrajectoryCalculation[int]{
			Value:   42,
			Factors: highConfidenceFactors(),
		}, nil
	}

	execution, err := ExecuteTrajectoryResult(
		context.Background(),
		executor,
		request,
	)
	if err != nil {
		t.Fatalf(
			"expected typed denial result, got %v",
			err,
		)
	}

	if called {
		t.Fatal("expected denied operation not to execute")
	}

	if !execution.IsDenied() ||
		execution.Result.Status !=
			analyticalresult.StatusDenied {
		t.Fatalf(
			"expected denied result, got %#v",
			execution.Result,
		)
	}

	if execution.ConfidenceReport != nil {
		t.Fatal("expected no confidence report for denied operation")
	}

	if execution.Result.Eligibility == nil ||
		execution.Result.Eligibility.Allowed ||
		len(execution.Result.Eligibility.Reasons) != 1 {
		t.Fatalf(
			"expected denial reasons, got %#v",
			execution.Result.Eligibility,
		)
	}
}

func TestExecuteTrajectoryResultMapsOperationFailure(
	t *testing.T,
) {
	executor := resultTestExecutor(
		t,
		true,
		nil,
	)
	request := completeExecutionRequest[int]()
	expectedError := errors.New(
		"provider unavailable",
	)
	request.Operation = func(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (TrajectoryCalculation[int], error) {
		return TrajectoryCalculation[int]{},
			expectedError
	}

	execution, err := ExecuteTrajectoryResult(
		context.Background(),
		executor,
		request,
	)
	if err != nil {
		t.Fatalf(
			"expected failed analytical result, got %v",
			err,
		)
	}

	if !execution.IsFailed() ||
		execution.Result.Failure == nil {
		t.Fatalf(
			"expected failed result metadata, got %#v",
			execution.Result,
		)
	}

	if execution.Result.Failure.Code !=
		FailureCodeAnalyticalOperationFailed ||
		execution.Result.Failure.Message !=
			expectedError.Error() {
		t.Fatalf(
			"unexpected default failure mapping: %#v",
			execution.Result.Failure,
		)
	}

	if execution.ConfidenceReport != nil {
		t.Fatal("expected no confidence report after operation failure")
	}
}

func TestExecuteTrajectoryResultUsesCustomFailureMapper(
	t *testing.T,
) {
	executor := resultTestExecutor(
		t,
		true,
		nil,
	)
	request := completeExecutionRequest[int]()
	request.Operation = func(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (TrajectoryCalculation[int], error) {
		return TrajectoryCalculation[int]{},
			errors.New("temporary outage")
	}
	request.FailureMapper = func(
		err error,
	) analyticalresult.Failure {
		return analyticalresult.Failure{
			Code:      "provider_temporarily_unavailable",
			Message:   "Provider is temporarily unavailable.",
			Retriable: true,
		}
	}

	execution, err := ExecuteTrajectoryResult(
		context.Background(),
		executor,
		request,
	)
	if err != nil {
		t.Fatalf(
			"expected custom failed result, got %v",
			err,
		)
	}

	if execution.Result.Failure == nil ||
		!execution.Result.Failure.Retriable ||
		execution.Result.Failure.Code !=
			"provider_temporarily_unavailable" {
		t.Fatalf(
			"unexpected custom failure metadata: %#v",
			execution.Result.Failure,
		)
	}
}

func TestExecuteTrajectoryResultCreatesFailedResultForZeroConfidence(
	t *testing.T,
) {
	executor := resultTestExecutor(
		t,
		true,
		nil,
	)
	request := completeExecutionRequest[int]()
	request.Operation = func(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (TrajectoryCalculation[int], error) {
		return TrajectoryCalculation[int]{
			Value: 42,
			Factors: []confidencereport.Factor{
				confidencereport.Evidence(
					confidencereport.
						FactorCodeTrajectoryQuality,
					1,
					0,
					"Trajectory quality provides no confidence.",
				),
			},
		}, nil
	}

	execution, err := ExecuteTrajectoryResult(
		context.Background(),
		executor,
		request,
	)
	if err != nil {
		t.Fatalf(
			"expected confidence-unavailable result, got %v",
			err,
		)
	}

	if execution.Result.Status !=
		analyticalresult.StatusFailed ||
		execution.Result.Failure == nil ||
		execution.Result.Failure.Code !=
			FailureCodeConfidenceUnavailable {
		t.Fatalf(
			"expected confidence-unavailable failure, got %#v",
			execution.Result,
		)
	}

	if execution.ConfidenceReport == nil ||
		execution.ConfidenceReport.Score != 0 {
		t.Fatalf(
			"expected zero-confidence report, got %#v",
			execution.ConfidenceReport,
		)
	}
}

func TestExecuteTrajectoryResultReturnsConfidenceContractError(
	t *testing.T,
) {
	executor := resultTestExecutor(
		t,
		true,
		nil,
	)
	request := completeExecutionRequest[int]()
	request.Operation = func(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (TrajectoryCalculation[int], error) {
		return TrajectoryCalculation[int]{
			Value: 42,
		}, nil
	}

	execution, err := ExecuteTrajectoryResult(
		context.Background(),
		executor,
		request,
	)
	if err == nil {
		t.Fatal("expected missing confidence factor error")
	}

	if execution.Result.Status != "" {
		t.Fatalf(
			"expected empty execution on contract error, got %#v",
			execution,
		)
	}

	if !strings.Contains(
		err.Error(),
		"confidence",
	) {
		t.Fatalf(
			"expected confidence error context, got %v",
			err,
		)
	}
}

func TestExecuteTrajectoryResultValidatesInfrastructure(
	t *testing.T,
) {
	request := completeExecutionRequest[int]()

	_, err := ExecuteTrajectoryResult(
		context.Background(),
		(*Executor)(nil),
		request,
	)
	if !errors.Is(
		err,
		ErrExecutorRequired,
	) {
		t.Fatalf(
			"expected executor requirement, got %v",
			err,
		)
	}

	executor := resultTestExecutor(
		t,
		true,
		nil,
	)
	request.Operation = nil

	_, err = ExecuteTrajectoryResult(
		context.Background(),
		executor,
		request,
	)
	if !errors.Is(
		err,
		ErrResultOperationRequired,
	) {
		t.Fatalf(
			"expected operation requirement, got %v",
			err,
		)
	}
}

func TestExecuteTrajectoryResultDoesNotMutateInputs(
	t *testing.T,
) {
	executor := resultTestExecutor(
		t,
		true,
		nil,
	)
	request := completeExecutionRequest[int]()
	request.Sources = []analyticalresult.Source{
		{
			Name: "airplanes.live",
			Role: analyticalresult.
				SourceRoleObservation,
		},
	}

	originalTrajectory := request.Trajectory
	originalSources := append(
		[]analyticalresult.Source(nil),
		request.Sources...,
	)

	request.Operation = func(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (TrajectoryCalculation[int], error) {
		return TrajectoryCalculation[int]{
			Value:   42,
			Factors: highConfidenceFactors(),
		}, nil
	}

	execution, err := ExecuteTrajectoryResult(
		context.Background(),
		executor,
		request,
	)
	if err != nil {
		t.Fatalf(
			"expected analytical result, got %v",
			err,
		)
	}

	request.Sources[0].Name =
		"mutated-source"

	if !reflect.DeepEqual(
		request.Trajectory,
		originalTrajectory,
	) {
		t.Fatal("expected trajectory input not to be mutated")
	}

	if originalSources[0].Name !=
		"airplanes.live" ||
		execution.Result.Sources[0].Name !=
			"airplanes.live" {
		t.Fatal("expected source metadata to be defensively copied")
	}
}

func TestDefaultFailureMapperRecognizesContextErrors(
	t *testing.T,
) {
	canceled := DefaultFailureMapper(
		context.Canceled,
	)
	if canceled.Code !=
		FailureCodeOperationCanceled ||
		!canceled.Retriable {
		t.Fatalf(
			"unexpected cancellation mapping: %#v",
			canceled,
		)
	}

	deadline := DefaultFailureMapper(
		context.DeadlineExceeded,
	)
	if deadline.Code !=
		FailureCodeOperationDeadlineExceeded ||
		!deadline.Retriable {
		t.Fatalf(
			"unexpected deadline mapping: %#v",
			deadline,
		)
	}
}

func completeExecutionRequest[T any]() TrajectoryResultRequest[T] {
	return TrajectoryResultRequest[T]{
		Trajectory: trajectory.FlightTrajectory{
			IdentityKey: "flight-identity-" +
				strings.Repeat(
					"a",
					64,
				),
			ICAO24: "ABC123",
		},
		Capability: trajectoryeligibility.
			CapabilityTrafficMetrics,
	}
}

func highConfidenceFactors() []confidencereport.Factor {
	return []confidencereport.Factor{
		confidencereport.Evidence(
			confidencereport.
				FactorCodeTrajectoryQuality,
			0.50,
			0.90,
			"Trajectory quality strongly supports the result.",
		),
		confidencereport.Evidence(
			confidencereport.
				FactorCodeIdentityReliability,
			0.50,
			0.90,
			"Flight identity strongly supports the result.",
		),
	}
}

func resultTestExecutor(
	t *testing.T,
	allowed bool,
	reasons []trajectoryeligibility.ReasonCode,
) *Executor {
	t.Helper()

	guard, err := scopeguard.New(
		scopeguard.Config{
			Evaluator: resultExecutionEvaluatorFunction(
				func(
					item trajectory.FlightTrajectory,
					now time.Time,
				) trajectoryeligibility.Evaluation {
					return trajectoryeligibility.Evaluation{
						Decisions: []trajectoryeligibility.Decision{
							{
								Capability: trajectoryeligibility.
									CapabilityTrafficMetrics,
								Allowed: allowed,
								Reasons: append(
									[]trajectoryeligibility.ReasonCode(nil),
									reasons...,
								),
							},
						},
					}
				},
			),
			Now: func() time.Time {
				return resultExecutionTestTime()
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"expected scope guard, got %v",
			err,
		)
	}

	return NewWithDependencies(
		nil,
		guard,
		confidencereport.NewDefault(),
	)
}

func resultExecutionTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		13,
		18,
		0,
		0,
		0,
		time.UTC,
	)
}

func hasAnalyticalNotice(
	notices []analyticalresult.Notice,
	code string,
) bool {
	for _, notice := range notices {
		if notice.Code == code {
			return true
		}
	}

	return false
}
