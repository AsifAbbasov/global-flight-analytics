package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestSaveFlightStatesAllowsEmptyBatchWithoutPool(t *testing.T) {
	repository := NewFlightStateRepository(nil)

	err := repository.SaveFlightStates(
		context.Background(),
		nil,
	)
	if err != nil {
		t.Fatalf(
			"expected empty batch to succeed without database pool, got %v",
			err,
		)
	}
}

func TestSaveFlightStatesRequiresPoolForNonEmptyBatch(t *testing.T) {
	repository := NewFlightStateRepository(nil)

	err := repository.SaveFlightStates(
		context.Background(),
		[]flightstate.FlightState{
			{
				ICAO24:              "ABC123",
				Callsign:            "AHY101",
				Latitude:            40.4093,
				Longitude:           49.8671,
				BarometricAltitudeM: 9753.6,
				GeometricAltitudeM:  9906,
				VelocityMPS:         231.5,
				HeadingDegrees:      92,
				VerticalRateMPS:     2.54,
				OnGround:            false,
				ObservedAt: time.Date(
					2026,
					time.July,
					4,
					12,
					0,
					0,
					0,
					time.UTC,
				),
				SourceName: "airplanes.live",
			},
		},
	)

	if !errors.Is(
		err,
		ErrFlightStateRepositoryPoolRequired,
	) {
		t.Fatalf(
			"expected ErrFlightStateRepositoryPoolRequired, got %v",
			err,
		)
	}
}

func TestSaveFlightStatesHandlesNilRepository(t *testing.T) {
	var repository *FlightStateRepository

	err := repository.SaveFlightStates(
		context.Background(),
		[]flightstate.FlightState{
			{
				ICAO24: "ABC123",
			},
		},
	)

	if !errors.Is(
		err,
		ErrFlightStateRepositoryPoolRequired,
	) {
		t.Fatalf(
			"expected ErrFlightStateRepositoryPoolRequired, got %v",
			err,
		)
	}
}
