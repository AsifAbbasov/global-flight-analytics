package traffic

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

type trafficServiceRepositoryStub struct{}

func (trafficServiceRepositoryStub) GetCurrent(context.Context) ([]CurrentTrafficItem, error) {
	return nil, nil
}
func (trafficServiceRepositoryStub) GetCurrentByBounds(context.Context, Bounds) ([]CurrentTrafficItem, error) {
	return nil, nil
}

func TestNewServiceReturnsDependencyErrors(t *testing.T) {
	service, err := NewService(nil, region.NewService())
	if service != nil || !errors.Is(err, dependency.ErrRequired) {
		t.Fatalf("NewService(nil, resolver) = %#v, %v", service, err)
	}

	service, err = NewService(trafficServiceRepositoryStub{}, nil)
	if service != nil || !errors.Is(err, dependency.ErrRequired) {
		t.Fatalf("NewService(repository, nil) = %#v, %v", service, err)
	}
}

func TestServiceKeepsCollectionsNonNil(t *testing.T) {
	service := MustNewService(trafficServiceRepositoryStub{}, region.NewService())
	items, err := service.GetCurrent(context.Background())
	if err != nil || items == nil {
		t.Fatalf("GetCurrent() = %#v, %v", items, err)
	}
}
