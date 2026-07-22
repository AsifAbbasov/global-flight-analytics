package metricquery

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const meanEarthRadiusKilometers = 6371.0088

var (
	ErrRegionalRepositoryUnsupported = errors.New(
		"analytical trajectory repository does not support regional queries",
	)
	ErrBoundsInvalid = errors.New(
		"analytical geographic bounds are invalid",
	)
)

type Bounds struct {
	MinLatitude  float64
	MaxLatitude  float64
	MinLongitude float64
	MaxLongitude float64
}

type RegionalRepository interface {
	ListTrajectoriesWithinBounds(
		ctx context.Context,
		observedFrom time.Time,
		observedTo time.Time,
		bounds Bounds,
		limit int,
	) ([]trajectory.FlightTrajectory, error)
}

func (bounds Bounds) Validate() error {
	values := []struct {
		name  string
		value float64
	}{
		{name: "minimum latitude", value: bounds.MinLatitude},
		{name: "maximum latitude", value: bounds.MaxLatitude},
		{name: "minimum longitude", value: bounds.MinLongitude},
		{name: "maximum longitude", value: bounds.MaxLongitude},
	}

	for _, item := range values {
		if math.IsNaN(item.value) || math.IsInf(item.value, 0) {
			return fmt.Errorf(
				"%w: %s must be finite",
				ErrBoundsInvalid,
				item.name,
			)
		}
	}

	if bounds.MinLatitude < -90 || bounds.MaxLatitude > 90 {
		return fmt.Errorf(
			"%w: latitude must be between minus ninety and ninety",
			ErrBoundsInvalid,
		)
	}
	if bounds.MinLongitude < -180 || bounds.MaxLongitude > 180 {
		return fmt.Errorf(
			"%w: longitude must be between minus one hundred eighty and one hundred eighty",
			ErrBoundsInvalid,
		)
	}
	if bounds.MinLatitude >= bounds.MaxLatitude {
		return fmt.Errorf(
			"%w: minimum latitude must be less than maximum latitude",
			ErrBoundsInvalid,
		)
	}
	if bounds.MinLongitude >= bounds.MaxLongitude {
		return fmt.Errorf(
			"%w: minimum longitude must be less than maximum longitude",
			ErrBoundsInvalid,
		)
	}

	return nil
}

func (bounds Bounds) AreaSquareKilometers() (float64, error) {
	if err := bounds.Validate(); err != nil {
		return 0, err
	}

	minimumLatitudeRadians := degreesToRadians(bounds.MinLatitude)
	maximumLatitudeRadians := degreesToRadians(bounds.MaxLatitude)
	longitudeWidthRadians := degreesToRadians(
		bounds.MaxLongitude - bounds.MinLongitude,
	)

	area := meanEarthRadiusKilometers *
		meanEarthRadiusKilometers *
		math.Abs(
			math.Sin(maximumLatitudeRadians)-
				math.Sin(minimumLatitudeRadians),
		) *
		longitudeWidthRadians

	if area <= 0 || math.IsNaN(area) || math.IsInf(area, 0) {
		return 0, fmt.Errorf(
			"%w: calculated area must be positive and finite",
			ErrBoundsInvalid,
		)
	}

	return area, nil
}

func (service *Service) RecentWithinBounds(
	ctx context.Context,
	request RecentRequest,
	bounds Bounds,
) ([]trajectory.FlightTrajectory, error) {
	window, err := request.Normalize(service.now())
	if err != nil {
		return nil, err
	}
	if err := bounds.Validate(); err != nil {
		return nil, err
	}

	repository, ok := service.repository.(RegionalRepository)
	if !ok {
		return nil, ErrRegionalRepositoryUnsupported
	}

	if ctx == nil {
		ctx = context.Background()
	}

	items, err := repository.ListTrajectoriesWithinBounds(
		ctx,
		window.ObservedFrom,
		window.ObservedTo,
		bounds,
		window.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list recent analytical trajectories within bounds: %w",
			err,
		)
	}

	return append(
		[]trajectory.FlightTrajectory(nil),
		items...,
	), nil
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}
