package airport

import (
	"context"
	"errors"
	"testing"
)

type airportServiceRepositoryStub struct{ requested string }

func (s *airportServiceRepositoryStub) List(context.Context) ([]Airport, error) { return nil, nil }
func (s *airportServiceRepositoryStub) GetByICAO(_ context.Context, value string) (Airport, error) {
	s.requested = value
	return Airport{}, nil
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
	repository := &airportServiceRepositoryStub{}
	service := NewService(repository)
	items, err := service.List(context.Background())
	if err != nil || items == nil {
		t.Fatalf("List() = %#v, %v", items, err)
	}
	_, err = service.GetByICAO(context.Background(), " ubbb ")
	if err != nil || repository.requested != "UBBB" {
		t.Fatalf("requested = %q, err = %v", repository.requested, err)
	}
	_, err = service.GetByICAO(context.Background(), " ")
	if !errors.Is(err, ErrServiceICAORequired) {
		t.Fatalf("blank error = %v", err)
	}
}
