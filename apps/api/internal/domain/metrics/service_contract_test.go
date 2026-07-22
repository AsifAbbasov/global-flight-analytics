package metrics

import (
	"context"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

type metricsServiceRepositoryStub struct{}

func (metricsServiceRepositoryStub) CountActiveAircraft(
	context.Context,
	ActiveAircraftQuery,
) (ActiveAircraftObservationSummary, error) {
	return ActiveAircraftObservationSummary{}, nil
}

func TestMetricsServiceRejectsNilDependencies(t *testing.T) {
	tests := []struct {
		name string
		run  func()
	}{
		{name: "repository", run: func() { NewService(nil, region.NewService()) }},
		{name: "region resolver", run: func() { NewService(metricsServiceRepositoryStub{}, nil) }},
		{name: "clock", run: func() { newServiceWithClock(metricsServiceRepositoryStub{}, region.NewService(), nil) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("constructor did not panic")
				}
			}()
			test.run()
		})
	}
}
