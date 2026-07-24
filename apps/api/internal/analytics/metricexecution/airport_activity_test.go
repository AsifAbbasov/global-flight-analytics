package metricexecution

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestAirportActivityClassifiesEligibleAirportCrossings(
	t *testing.T,
) {
	service := metricTestService(t, allowUnlessDeniedICAO)
	selectedAirport := airport.Airport{
		ICAOCode:  "UBBB",
		Latitude:  40.4675,
		Longitude: 50.0467,
	}

	arrival := airportMovementTrajectory(
		"a", "ARRIVAL",
		40.90, 50.05,
		40.47, 50.05,
	)
	departure := airportMovementTrajectory(
		"b", "DEPARTURE",
		40.47, 50.05,
		40.90, 50.05,
	)

	execution, err := service.AirportActivity(
		context.Background(),
		AirportActivityRequest{
			Airport: selectedAirport,
			Trajectories: []trajectory.FlightTrajectory{
				arrival,
				departure,
			},
			RadiusKilometers: 15,
		},
	)
	if err != nil {
		t.Fatalf("expected airport activity execution, got %v", err)
	}
	if execution.Result.Value != 2 ||
		execution.Result.Status != analyticalresult.StatusComplete {
		t.Fatalf("expected complete airport activity two, got %#v", execution.Result)
	}
}

func TestAirportActivityExcludesUnrelatedAndAmbiguousTrajectories(
	t *testing.T,
) {
	service := metricTestService(t, allowUnlessDeniedICAO)
	selectedAirport := airport.Airport{
		ICAOCode:  "UBBB",
		Latitude:  40.4675,
		Longitude: 50.0467,
	}

	unrelated := airportMovementTrajectory(
		"a", "UNRELATED",
		41.20, 50.50,
		41.30, 50.60,
	)
	ambiguous := airportMovementTrajectory(
		"b", "AMBIGUOUS",
		40.47, 50.05,
		40.48, 50.06,
	)

	execution, err := service.AirportActivity(
		context.Background(),
		AirportActivityRequest{
			Airport: selectedAirport,
			Trajectories: []trajectory.FlightTrajectory{
				unrelated,
				ambiguous,
			},
			RadiusKilometers: 15,
		},
	)
	if err != nil {
		t.Fatalf("expected airport activity execution, got %v", err)
	}
	if execution.Result.Value != 0 ||
		execution.Result.Status != analyticalresult.StatusLimited {
		t.Fatalf("expected limited zero activity, got %#v", execution.Result)
	}
	if !containsNotice(
		nil,
		execution.Result.Limitations,
		NoticeCodeUnrelatedAirportTrajectoriesExcluded,
	) {
		t.Fatal("expected unrelated trajectory limitation")
	}
	if !containsNotice(
		nil,
		execution.Result.Limitations,
		NoticeCodeAmbiguousAirportMovementsExcluded,
	) {
		t.Fatal("expected ambiguous movement limitation")
	}
}

func TestAirportActivityRejectsInvalidRadius(
	t *testing.T,
) {
	service := metricTestService(t, allowUnlessDeniedICAO)
	_, err := service.AirportActivity(
		context.Background(),
		AirportActivityRequest{
			Airport: airport.Airport{
				ICAOCode:  "UBBB",
				Latitude:  40.4675,
				Longitude: 50.0467,
			},
			RadiusKilometers: 101,
		},
	)
	if !errors.Is(err, ErrAirportActivityRadiusInvalid) {
		t.Fatalf("expected radius error, got %v", err)
	}
}

func airportMovementTrajectory(
	identityCharacter string,
	icao24 string,
	startLatitude float64,
	startLongitude float64,
	endLatitude float64,
	endLongitude float64,
) trajectory.FlightTrajectory {
	item := healthyMetricTrajectory(
		identityCharacter,
		icao24,
	)
	item.Points = []trajectory.TrackPoint4D{
		{
			Latitude:   startLatitude,
			Longitude:  startLongitude,
			ObservedAt: metricTestTime().Add(-4 * time.Minute),
			SourceName: "airplanes.live",
			ICAO24:     icao24,
		},
		{
			Latitude:   endLatitude,
			Longitude:  endLongitude,
			ObservedAt: metricTestTime().Add(-30 * time.Second),
			SourceName: "airplanes.live",
			ICAO24:     icao24,
		},
	}
	item.PointCount = len(item.Points)
	return item
}
