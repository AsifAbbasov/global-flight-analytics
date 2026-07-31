package projectionread

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
)

func TestServiceAllowsDefaultProjectionDuration(t *testing.T) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	source := &dataSourceStub{
		snapshot: Snapshot{
			CurrentTrajectory: current,
		},
	}
	composer := &composerStub{
		result: projectionproduction.Result{
			Version: projectionproduction.Version,
		},
	}
	service, err := NewService(
		ServiceConfig{
			DataSource: source,
			Composer:   composer,
			Policy:     DefaultPolicy(),
			Now: func() time.Time {
				return asOfTime.Add(time.Second)
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.Get(
		context.Background(),
		Request{
			TrajectoryID: current.ID,
			AsOfTime:     asOfTime,
		},
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if source.calls != 1 || composer.calls != 1 {
		t.Fatalf(
			"unexpected dependency calls: source=%d composer=%d",
			source.calls,
			composer.calls,
		)
	}
	expectedDuration := DefaultPolicy().Horizon.DefaultDuration
	if composer.request.RequestedDuration != expectedDuration {
		t.Fatalf(
			"composer requested duration = %s, want canonical default %s",
			composer.request.RequestedDuration,
			expectedDuration,
		)
	}
}

func TestServiceRejectsInvalidHorizonDurationBeforeLoadingSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{
			name:     "negative",
			duration: -time.Second,
		},
		{
			name:     "below minimum",
			duration: 30 * time.Second,
		},
		{
			name:     "off fixed grid",
			duration: 61 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			asOfTime := projectionReadTestAsOfTime()
			source := &dataSourceStub{}
			composer := &composerStub{}
			service, err := NewService(
				ServiceConfig{
					DataSource: source,
					Composer:   composer,
					Policy:     DefaultPolicy(),
					Now: func() time.Time {
						return asOfTime
					},
				},
			)
			if err != nil {
				t.Fatalf("NewService() error = %v", err)
			}

			_, err = service.Get(
				context.Background(),
				Request{
					TrajectoryID: "73aa02ab-7061-4e9e-a238-d32710371ee3",
					AsOfTime:     asOfTime,
					RequestedDuration: test.
						duration,
				},
			)
			if !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("error = %v, want %v", err, ErrInvalidRequest)
			}
			if source.calls != 0 || composer.calls != 0 {
				t.Fatalf(
					"invalid duration reached dependencies: source=%d composer=%d",
					source.calls,
					composer.calls,
				)
			}
		})
	}
}
