package flight

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

type flightServiceRepositoryStub struct {
	list      []Flight
	item      Flight
	requested string
}

func (s *flightServiceRepositoryStub) List(context.Context) ([]Flight, error) {
	return s.list, nil
}
func (s *flightServiceRepositoryStub) GetByID(_ context.Context, value string) (Flight, error) {
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
	now := time.Now().UTC()
	valid := Flight{FirstSeenAt: now, LastSeenAt: now}
	repository := &flightServiceRepositoryStub{
		list: []Flight{valid},
		item: valid,
	}
	service := MustNewService(repository)

	items, err := service.List(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("List() = %#v, %v", items, err)
	}
	_, err = service.GetByID(context.Background(), " flight-id ")
	if err != nil || repository.requested != "flight-id" {
		t.Fatalf("requested = %q, err = %v", repository.requested, err)
	}

	repository.item = Flight{}
	_, err = service.GetByID(context.Background(), "flight-id")
	if !errors.Is(err, ErrServiceRepositoryResultInvalid) {
		t.Fatalf("invalid result error = %v", err)
	}
}
