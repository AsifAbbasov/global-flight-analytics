package metrics

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

func TestNewServiceReturnsDependencyErrors(t *testing.T) {
	service, err := NewService(nil, region.NewService())
	if service != nil || !errors.Is(err, dependency.ErrRequired) {
		t.Fatalf("NewService(nil, resolver) = %#v, %v", service, err)
	}

	service, err = NewService(&activeAircraftRepositoryStub{}, nil)
	if service != nil || !errors.Is(err, dependency.ErrRequired) {
		t.Fatalf("NewService(repository, nil) = %#v, %v", service, err)
	}
}
