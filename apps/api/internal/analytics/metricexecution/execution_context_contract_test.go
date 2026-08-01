package metricexecution

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/confidencereport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/scopeguard"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type contextContractExecutor struct {
	filterCalls     int
	confidenceCalls int
}

func (executor *contextContractExecutor) FilterTrajectories(
	[]trajectory.FlightTrajectory,
	trajectoryeligibility.Capability,
) (scopeguard.FilterResult, error) {
	executor.filterCalls++

	return scopeguard.FilterResult{}, nil
}

func (executor *contextContractExecutor) EvaluateConfidence(
	confidencereport.Request,
) (confidencereport.Report, error) {
	executor.confidenceCalls++

	return confidencereport.Report{}, nil
}

func TestExecuteTrajectoryMetricRejectsNilContextBeforeExecution(
	t *testing.T,
) {
	analyticsExecutor := &contextContractExecutor{}
	service := &Service{
		executor: analyticsExecutor,
	}
	operationCalls := 0

	_, err := executeTrajectoryMetric[int](
		nil,
		service,
		"active_aircraft",
		trajectoryeligibility.CapabilityTrafficMetrics,
		nil,
		PublicationMetadata{},
		nil,
		func(
			context.Context,
			[]trajectory.FlightTrajectory,
			time.Time,
		) (metricCalculation[int], error) {
			operationCalls++

			return metricCalculation[int]{}, nil
		},
	)

	if !errors.Is(
		err,
		ErrMetricExecutionContextRequired,
	) {
		t.Fatalf(
			"error = %v, want context required",
			err,
		)
	}
	if analyticsExecutor.filterCalls != 0 {
		t.Fatalf(
			"filter calls = %d, want 0",
			analyticsExecutor.filterCalls,
		)
	}
	if operationCalls != 0 {
		t.Fatalf(
			"operation calls = %d, want 0",
			operationCalls,
		)
	}
	if analyticsExecutor.confidenceCalls != 0 {
		t.Fatalf(
			"confidence calls = %d, want 0",
			analyticsExecutor.confidenceCalls,
		)
	}
}

func TestExecuteSnapshotMetricRejectsNilContextBeforeExecution(
	t *testing.T,
) {
	analyticsExecutor := &contextContractExecutor{}
	service := &Service{
		executor: analyticsExecutor,
	}
	operationCalls := 0

	_, err := executeSnapshotMetric[int](
		nil,
		service,
		"data_freshness",
		trajectoryeligibility.CapabilityTrafficMetrics,
		PublicationMetadata{},
		func(
			context.Context,
			time.Time,
		) (metricCalculation[int], error) {
			operationCalls++

			return metricCalculation[int]{}, nil
		},
	)

	if !errors.Is(
		err,
		ErrMetricExecutionContextRequired,
	) {
		t.Fatalf(
			"error = %v, want context required",
			err,
		)
	}
	if analyticsExecutor.filterCalls != 0 {
		t.Fatalf(
			"filter calls = %d, want 0",
			analyticsExecutor.filterCalls,
		)
	}
	if operationCalls != 0 {
		t.Fatalf(
			"operation calls = %d, want 0",
			operationCalls,
		)
	}
	if analyticsExecutor.confidenceCalls != 0 {
		t.Fatalf(
			"confidence calls = %d, want 0",
			analyticsExecutor.confidenceCalls,
		)
	}
}
