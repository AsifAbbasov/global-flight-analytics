package metricexecution

import (
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/executor"
)

var (
	ErrExecutorRequired = errors.New(
		"analytics executor is required",
	)
	ErrMetricOperationRequired = errors.New(
		"metric operation is required",
	)
	ErrAggregateDenialReasonsMissing = errors.New(
		"aggregate metric denial reasons are missing",
	)
	ErrAirportMovementConflict = errors.New(
		"trajectory cannot be classified as both arrival and departure",
	)
	ErrAirportMovementRoleMissing = errors.New(
		"eligible trajectory airport movement role is missing",
	)
)

type Service struct {
	executor *executor.Executor
}

func New(
	analyticsExecutor *executor.Executor,
) (*Service, error) {
	if analyticsExecutor == nil {
		return nil, ErrExecutorRequired
	}

	return &Service{
		executor: analyticsExecutor,
	}, nil
}

func (
	service *Service,
) Executor() *executor.Executor {
	if service == nil {
		return nil
	}

	return service.executor
}
