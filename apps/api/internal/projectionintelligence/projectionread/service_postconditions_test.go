package projectionread

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
)

func TestServiceRejectsNilContext(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	service, err := NewService(
		ServiceConfig{
			DataSource: &dataSourceStub{},
			Composer:   &composerStub{},
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
		nil,
		Request{
			TrajectoryID:      "73aa02ab-7061-4e9e-a238-d32710371ee3",
			AsOfTime:          asOfTime,
			RequestedDuration: time.Minute,
		},
	)
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf(
			"error = %v, want ErrContextRequired",
			err,
		)
	}
}

func TestServiceRejectsSnapshotIdentityMismatch(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"83aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	source := &dataSourceStub{
		snapshot: Snapshot{
			CurrentTrajectory: current,
		},
	}
	composer := &composerStub{
		buildValidResult: true,
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
			TrajectoryID:      "73aa02ab-7061-4e9e-a238-d32710371ee3",
			AsOfTime:          asOfTime,
			RequestedDuration: time.Minute,
		},
	)
	if !errors.Is(
		err,
		ErrSnapshotIdentityMismatch,
	) {
		t.Fatalf(
			"error = %v, want snapshot identity sentinel",
			err,
		)
	}
	if composer.calls != 0 {
		t.Fatalf(
			"composer calls = %d, want 0",
			composer.calls,
		)
	}
}

func TestServiceRejectsInvalidComposerOutput(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	service, err := NewService(
		ServiceConfig{
			DataSource: &dataSourceStub{
				snapshot: Snapshot{
					CurrentTrajectory: current,
				},
			},
			Composer: &composerStub{
				result: projectionproduction.Result{
					Version: projectionproduction.Version,
				},
				preserveConfiguredResult: true,
			},
			Policy: DefaultPolicy(),
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
			TrajectoryID:      current.ID,
			AsOfTime:          asOfTime,
			RequestedDuration: time.Minute,
		},
	)
	if !errors.Is(err, ErrComposedResultInvalid) {
		t.Fatalf(
			"error = %v, want composed result sentinel",
			err,
		)
	}
}

func TestServiceRejectsValidResultForAnotherTrajectory(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	composer := &composerStub{
		buildValidResult: true,
		transform: func(
			result projectionproduction.Result,
		) (projectionproduction.Result, error) {
			result.Projection.TrajectoryID =
				"83aa02ab-7061-4e9e-a238-d32710371ee3"
			return projectionproduction.Finalize(result)
		},
	}
	service, err := NewService(
		ServiceConfig{
			DataSource: &dataSourceStub{
				snapshot: Snapshot{
					CurrentTrajectory: current,
				},
			},
			Composer: composer,
			Policy:   DefaultPolicy(),
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
			TrajectoryID:      current.ID,
			AsOfTime:          asOfTime,
			RequestedDuration: time.Minute,
		},
	)
	if !errors.Is(
		err,
		ErrSnapshotIdentityMismatch,
	) {
		t.Fatalf(
			"error = %v, want identity sentinel",
			err,
		)
	}
}

func TestServiceGeneratesAfterSnapshotAcquisition(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	times := []time.Time{
		asOfTime.Add(time.Second),
		asOfTime.Add(3 * time.Second),
	}
	clockCalls := 0
	composer := &composerStub{
		buildValidResult: true,
	}
	service, err := NewService(
		ServiceConfig{
			DataSource: &dataSourceStub{
				snapshot: Snapshot{
					CurrentTrajectory: current,
				},
			},
			Composer: composer,
			Policy:   DefaultPolicy(),
			Now: func() time.Time {
				value := times[clockCalls]
				clockCalls++
				return value
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.Get(
		context.Background(),
		Request{
			TrajectoryID:      current.ID,
			AsOfTime:          asOfTime,
			RequestedDuration: time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if clockCalls != 2 ||
		!composer.request.GeneratedAt.Equal(
			times[1],
		) {
		t.Fatalf(
			"clock calls=%d generatedAt=%s, want %s",
			clockCalls,
			composer.request.GeneratedAt,
			times[1],
		)
	}
}
