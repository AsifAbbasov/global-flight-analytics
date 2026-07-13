package metricquery

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type Repository interface {
	ListTrajectoriesByEndTime(
		ctx context.Context,
		observedFrom time.Time,
		observedTo time.Time,
		limit int,
	) ([]trajectory.FlightTrajectory, error)

	ListTrajectoriesByIDs(
		ctx context.Context,
		trajectoryIDs []string,
	) ([]trajectory.FlightTrajectory, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func New(
	repository Repository,
) (*Service, error) {
	return NewWithClock(
		repository,
		time.Now,
	)
}

func NewWithClock(
	repository Repository,
	now func() time.Time,
) (*Service, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}
	if now == nil {
		now = time.Now
	}

	return &Service{
		repository: repository,
		now:        now,
	}, nil
}

func (
	service *Service,
) Recent(
	ctx context.Context,
	request RecentRequest,
) ([]trajectory.FlightTrajectory, error) {
	window, err := request.Normalize(
		service.now(),
	)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	items, err :=
		service.repository.
			ListTrajectoriesByEndTime(
				ctx,
				window.ObservedFrom,
				window.ObservedTo,
				window.Limit,
			)
	if err != nil {
		return nil, fmt.Errorf(
			"list recent analytical trajectories: %w",
			err,
		)
	}

	return append(
		[]trajectory.FlightTrajectory(nil),
		items...,
	), nil
}

func (
	service *Service,
) ByIDs(
	ctx context.Context,
	trajectoryIDs []string,
) ([]trajectory.FlightTrajectory, error) {
	normalizedIDs, err :=
		NormalizeTrajectoryIDs(
			trajectoryIDs,
		)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	items, err :=
		service.repository.
			ListTrajectoriesByIDs(
				ctx,
				normalizedIDs,
			)
	if err != nil {
		return nil, fmt.Errorf(
			"list analytical trajectories by ids: %w",
			err,
		)
	}

	return append(
		[]trajectory.FlightTrajectory(nil),
		items...,
	), nil
}
