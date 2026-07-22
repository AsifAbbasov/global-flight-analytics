package aircraft

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

type aircraftServiceRepositoryStub struct {
	list      []Aircraft
	item      Aircraft
	requested string
}

func (s *aircraftServiceRepositoryStub) List(context.Context) ([]Aircraft, error) {
	return s.list, nil
}
func (s *aircraftServiceRepositoryStub) GetByICAO24(_ context.Context, value string) (Aircraft, error) {
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
	valid := Aircraft{ICAO24: "abc123"}
	repository := &aircraftServiceRepositoryStub{
		list: []Aircraft{valid},
		item: valid,
	}
	service := MustNewService(repository)

	items, err := service.List(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("List() = %#v, %v", items, err)
	}
	_, err = service.GetByICAO24(context.Background(), " ABC123 ")
	if err != nil || repository.requested != "abc123" {
		t.Fatalf("requested = %q, err = %v", repository.requested, err)
	}

	repository.item = Aircraft{}
	_, err = service.GetByICAO24(context.Background(), "abc123")
	if !errors.Is(err, ErrServiceRepositoryResultInvalid) {
		t.Fatalf("invalid result error = %v", err)
	}
}
