package traffic

import (
	"context"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

type trafficServiceRepositoryStub struct{}

func (trafficServiceRepositoryStub) GetCurrent(context.Context) ([]CurrentTrafficItem, error) {
	return nil, nil
}
func (trafficServiceRepositoryStub) GetCurrentByBounds(context.Context, Bounds) ([]CurrentTrafficItem, error) {
	return nil, nil
}

type trafficRegionResolverStub struct{}

func (trafficRegionResolverStub) GetByCode(string) (region.Region, error) {
	return region.Region{}, nil
}

func TestTrafficServiceRejectsNilDependenciesAndNormalizesNilSlices(t *testing.T) {
	tests := []func(){
		func() { NewService(nil, trafficRegionResolverStub{}) },
		func() { NewService(trafficServiceRepositoryStub{}, nil) },
	}
	for _, run := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("constructor did not panic")
				}
			}()
			run()
		}()
	}

	service := NewService(trafficServiceRepositoryStub{}, trafficRegionResolverStub{})
	items, err := service.GetCurrent(context.Background())
	if err != nil || items == nil {
		t.Fatalf("GetCurrent() = %#v, %v", items, err)
	}
}
