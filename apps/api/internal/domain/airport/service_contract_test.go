package airport

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

type airportServiceRepositoryStub struct {
	list      []Airport
	item      Airport
	requested string
}

func (s *airportServiceRepositoryStub) List(context.Context) ([]Airport, error) {
	return s.list, nil
}
func (s *airportServiceRepositoryStub) GetByICAO(_ context.Context, value string) (Airport, error) {
	s.requested = value
	return s.item, nil
}

func TestNewServiceReturnsDependencyError(t *testing.T) {
	service, err := NewService(nil)
	if service != nil || !errors.Is(err, dependency.ErrRequired) {
		t.Fatalf("NewService(nil) = %#v, %v", service, err)
	}
}

func TestServiceNormalizesAndValidatesRepositoryResults(t *testing.T) {
	valid := Airport{ICAOCode: "UBBB", Latitude: 40.4675, Longitude: 50.0467}
	repository := &airportServiceRepositoryStub{
		list: []Airport{valid},
		item: valid,
	}
	service := MustNewService(repository)

	items, err := service.List(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("List() = %#v, %v", items, err)
	}
	_, err = service.GetByICAO(context.Background(), " ubbb ")
	if err != nil || repository.requested != "UBBB" {
		t.Fatalf("requested = %q, err = %v", repository.requested, err)
	}

	repository.item = Airport{}
	_, err = service.GetByICAO(context.Background(), "UBBB")
	if !errors.Is(err, ErrServiceRepositoryResultInvalid) {
		t.Fatalf("invalid result error = %v", err)
	}
}
