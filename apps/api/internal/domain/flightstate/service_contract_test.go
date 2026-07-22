package flightstate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

type flightStateServiceRepositoryStub struct {
	list      []FlightState
	item      FlightState
	requested string
}

func (s *flightStateServiceRepositoryStub) ListByFlightID(_ context.Context, value string) ([]FlightState, error) {
	s.requested = value
	return s.list, nil
}
func (s *flightStateServiceRepositoryStub) GetLatestByICAO24(_ context.Context, value string) (FlightState, error) {
	s.requested = value
	return s.item, nil
}

func validServiceFlightState() FlightState {
	return FlightState{
		ICAO24:                   "abc123",
		Latitude:                 40.4,
		Longitude:                49.8,
		BarometricAltitudeStatus: AltitudeStatusUnavailable,
		GeometricAltitudeStatus:  AltitudeStatusUnavailable,
		ObservedAt:               time.Now().UTC(),
		SourceName:               "opensky",
	}
}

func TestNewServiceReturnsDependencyError(t *testing.T) {
	service, err := NewService(nil)
	if service != nil || !errors.Is(err, dependency.ErrRequired) {
		t.Fatalf("NewService(nil) = %#v, %v", service, err)
	}
}

func TestServiceNormalizesAndValidatesRepositoryResults(t *testing.T) {
	valid := validServiceFlightState()
	repository := &flightStateServiceRepositoryStub{
		list: []FlightState{valid},
		item: valid,
	}
	service := MustNewService(repository)

	items, err := service.ListByFlightID(context.Background(), " flight-id ")
	if err != nil || len(items) != 1 || repository.requested != "flight-id" {
		t.Fatalf("ListByFlightID() = %#v, requested=%q, err=%v", items, repository.requested, err)
	}
	_, err = service.GetLatestByICAO24(context.Background(), " ABC123 ")
	if err != nil || repository.requested != "abc123" {
		t.Fatalf("requested = %q, err = %v", repository.requested, err)
	}

	repository.item = FlightState{}
	_, err = service.GetLatestByICAO24(context.Background(), "abc123")
	if !errors.Is(err, ErrServiceRepositoryResultInvalid) {
		t.Fatalf("invalid result error = %v", err)
	}
}
