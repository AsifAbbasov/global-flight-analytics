package metricexecution

import (
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/confidencereport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/scopeguard"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
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
	ErrAirportActivityAirportInvalid = errors.New(
		"airport activity airport is invalid",
	)
	ErrAirportActivityRadiusInvalid = errors.New(
		"airport activity radius must be finite, positive and at most one hundred kilometers",
	)
)

type analyticalExecutor interface {
	FilterTrajectories(
		items []trajectory.FlightTrajectory,
		capability trajectoryeligibility.Capability,
	) (scopeguard.FilterResult, error)

	EvaluateConfidence(
		request confidencereport.Request,
	) (confidencereport.Report, error)
}

type Service struct {
	executor analyticalExecutor
}

func New(
	analyticsExecutor analyticalExecutor,
) (*Service, error) {
	if analyticsExecutor == nil {
		return nil, ErrExecutorRequired
	}

	return &Service{
		executor: analyticsExecutor,
	}, nil
}
