package flightstate

import (
	"context"
	"errors"
	"testing"
)

type flightStateServiceRepositoryStub struct{ flightID, icao24 string }

func (s *flightStateServiceRepositoryStub) ListByFlightID(_ context.Context, value string) ([]FlightState, error) {
	s.flightID = value
	return nil, nil
}
func (s *flightStateServiceRepositoryStub) GetLatestByICAO24(_ context.Context, value string) (FlightState, error) {
	s.icao24 = value
	return FlightState{}, nil
}

func TestNewServiceRejectsNilRepository(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewService(nil) did not panic")
		}
	}()
	NewService(nil)
}

func TestServiceBoundaryContract(t *testing.T) {
	repository := &flightStateServiceRepositoryStub{}
	service := NewService(repository)
	items, err := service.ListByFlightID(context.Background(), " flight-id ")
	if err != nil || items == nil || repository.flightID != "flight-id" {
		t.Fatalf("items=%#v flightID=%q err=%v", items, repository.flightID, err)
	}
	_, err = service.GetLatestByICAO24(context.Background(), " ABC123 ")
	if err != nil || repository.icao24 != "abc123" {
		t.Fatalf("icao24=%q err=%v", repository.icao24, err)
	}
	_, err = service.ListByFlightID(context.Background(), " ")
	if !errors.Is(err, ErrServiceFlightIDRequired) {
		t.Fatalf("blank flight error=%v", err)
	}
}
