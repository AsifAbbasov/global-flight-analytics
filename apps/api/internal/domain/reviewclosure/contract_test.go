package reviewclosure

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flight"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerhealth"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
)

var (
	_ func(aircraft.Repository) (*aircraft.Service, error)                       = aircraft.NewService
	_ func(airport.Repository) (*airport.Service, error)                         = airport.NewService
	_ func(flight.Repository) (*flight.Service, error)                           = flight.NewService
	_ func(flightstate.Repository) (*flightstate.Service, error)                 = flightstate.NewService
	_ func(metrics.Repository, metrics.RegionResolver) (*metrics.Service, error) = metrics.NewService
	_ func(traffic.Repository, traffic.RegionResolver) (*traffic.Service, error) = traffic.NewService
)

func TestReviewClosureUsesTypedPolicyAndValueObjects(t *testing.T) {
	policy := providerhealth.Policy{}
	var _ providerhealth.BasisPoints = policy.MinimumHealthySuccessRatio
	var _ providerhealth.BasisPoints = policy.MaximumHealthyRejectionRatio

	altitude, err := flightstate.NewAltitude(
		0,
		flightstate.AltitudeStatusObserved,
	)
	if err != nil || altitude.Status() != flightstate.AltitudeStatusObserved {
		t.Fatalf("altitude value object = %+v, %v", altitude, err)
	}

	category, err := flightstate.NewAircraftCategory(0)
	if err != nil || !category.Available() {
		t.Fatalf("aircraft category value object = %+v, %v", category, err)
	}
}
