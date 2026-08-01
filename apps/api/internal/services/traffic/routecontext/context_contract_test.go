package routecontext

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type routeContextContractTrajectoryReader struct {
	calls int
}

func (
	reader *routeContextContractTrajectoryReader,
) GetLatestTrajectoryByICAO24(
	context.Context,
	string,
) (trajectory.FlightTrajectory, error) {
	reader.calls++

	return trajectory.FlightTrajectory{}, nil
}

type routeContextContractAirportLister struct {
	calls int
}

func (
	lister *routeContextContractAirportLister,
) List(
	context.Context,
) ([]airport.Airport, error) {
	lister.calls++

	return nil, nil
}

func TestGetByICAO24RejectsNilContextBeforeDependencyReads(
	t *testing.T,
) {
	trajectoryReader := &routeContextContractTrajectoryReader{}
	airportLister := &routeContextContractAirportLister{}
	service := New(Config{
		TrajectoryReader: trajectoryReader,
		AirportLister:    airportLister,
	})

	result, err := service.GetByICAO24(
		nil,
		"ABC123",
	)

	if !errors.Is(
		err,
		ErrRouteContextRequired,
	) {
		t.Fatalf(
			"error = %v, want route context required",
			err,
		)
	}
	if trajectoryReader.calls != 0 {
		t.Fatalf(
			"trajectory reader calls = %d, want 0",
			trajectoryReader.calls,
		)
	}
	if airportLister.calls != 0 {
		t.Fatalf(
			"airport lister calls = %d, want 0",
			airportLister.calls,
		)
	}
	if result.ICAO24 != "" ||
		result.TrajectoryID != "" ||
		result.Origin != nil ||
		result.Destination != nil {
		t.Fatalf(
			"result = %#v, want empty context",
			result,
		)
	}
}
