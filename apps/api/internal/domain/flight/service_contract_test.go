package flight

import (
	"context"
	"errors"
	"testing"
)

type flightServiceRepositoryStub struct{ requested string }

func (s *flightServiceRepositoryStub) List(context.Context) ([]Flight, error) { return nil, nil }
func (s *flightServiceRepositoryStub) GetByID(_ context.Context, value string) (Flight, error) {
	s.requested = value
	return Flight{}, nil
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
	repository := &flightServiceRepositoryStub{}
	service := NewService(repository)
	items, err := service.List(context.Background())
	if err != nil || items == nil {
		t.Fatalf("List() = %#v, %v", items, err)
	}
	_, err = service.GetByID(context.Background(), " id ")
	if err != nil || repository.requested != "id" {
		t.Fatalf("requested = %q, err = %v", repository.requested, err)
	}
	_, err = service.GetByID(context.Background(), " ")
	if !errors.Is(err, ErrServiceFlightIDRequired) {
		t.Fatalf("blank error = %v", err)
	}
}
