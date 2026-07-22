package aircraft

import (
	"context"
	"errors"
	"testing"
)

type aircraftServiceRepositoryStub struct {
	list      []Aircraft
	requested string
}

func (s *aircraftServiceRepositoryStub) List(context.Context) ([]Aircraft, error) { return s.list, nil }
func (s *aircraftServiceRepositoryStub) GetByICAO24(_ context.Context, value string) (Aircraft, error) {
	s.requested = value
	return Aircraft{}, nil
}

func TestServiceRejectsNilRepositoryAndNormalizesBoundary(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewService(nil) did not panic")
		}
	}()
	NewService(nil)
}

func TestServiceNormalizesICAO24AndNilList(t *testing.T) {
	repository := &aircraftServiceRepositoryStub{}
	service := NewService(repository)
	items, err := service.List(context.Background())
	if err != nil || items == nil {
		t.Fatalf("List() = %#v, %v", items, err)
	}
	_, err = service.GetByICAO24(context.Background(), " ABC123 ")
	if err != nil || repository.requested != "abc123" {
		t.Fatalf("requested = %q, err = %v", repository.requested, err)
	}
	_, err = service.GetByICAO24(context.Background(), " ")
	if !errors.Is(err, ErrServiceICAO24Required) {
		t.Fatalf("blank error = %v", err)
	}
}
