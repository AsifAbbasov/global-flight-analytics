package weathercontext

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheruncertainty"
	"github.com/jackc/pgx/v5"
)

type contextContractSnapshotQueryer struct {
	calls int
}

func (
	queryer *contextContractSnapshotQueryer,
) QueryRow(
	context.Context,
	string,
	...any,
) pgx.Row {
	queryer.calls++

	return nil
}

type contextContractTrajectoryReader struct {
	calls int
}

func (
	reader *contextContractTrajectoryReader,
) GetTrajectory(
	context.Context,
	TrajectoryRequest,
) (trajectory.FlightTrajectory, error) {
	reader.calls++

	return trajectory.FlightTrajectory{}, nil
}

type contextContractWeatherReader struct {
	calls int
}

func (
	reader *contextContractWeatherReader,
) GetLatestSnapshot(
	context.Context,
	WeatherSnapshotRequest,
) (domainweather.CurrentSnapshot, error) {
	reader.calls++

	return domainweather.CurrentSnapshot{}, nil
}

type contextContractProjectionReader struct {
	calls int
}

func (
	reader *contextContractProjectionReader,
) GetProjection(
	context.Context,
	ProjectionRequest,
) (projectionproduction.Result, error) {
	reader.calls++

	return projectionproduction.Result{}, nil
}

func TestPostgresSnapshotReaderRejectsNilContextBeforeQuery(
	t *testing.T,
) {
	queryer := &contextContractSnapshotQueryer{}
	reader, err := newPostgresSnapshotReader(
		queryer,
		DefaultPostgresSnapshotPolicy(),
	)
	if err != nil {
		t.Fatalf(
			"create snapshot reader: %v",
			err,
		)
	}

	result, err := reader.GetLatestSnapshot(
		nil,
		WeatherSnapshotRequest{},
	)

	if !errors.Is(
		err,
		ErrSnapshotContextRequired,
	) {
		t.Fatalf(
			"error = %v, want snapshot context required",
			err,
		)
	}
	if queryer.calls != 0 {
		t.Fatalf(
			"PostgreSQL query calls = %d, want 0",
			queryer.calls,
		)
	}
	if result.Provider != "" ||
		!result.ObservedAt.IsZero() ||
		!result.RetrievedAt.IsZero() {
		t.Fatalf(
			"snapshot = %#v, want empty snapshot",
			result,
		)
	}
}

func TestServiceRejectsNilContextBeforeDependencyReads(
	t *testing.T,
) {
	trajectoryReader := &contextContractTrajectoryReader{}
	weatherReader := &contextContractWeatherReader{}
	projectionReader := &contextContractProjectionReader{}
	service, err := NewService(Config{
		TrajectoryReader:      trajectoryReader,
		WeatherSnapshotReader: weatherReader,
		ProjectionReader:      projectionReader,
		TrustPolicy:           weathertrust.DefaultPolicy(),
		AlignmentPolicy:       weatheralignment.DefaultPolicy(),
		EncounterPolicy:       weatherencounter.DefaultPolicy(),
		UncertaintyPolicy:     weatheruncertainty.DefaultPolicy(),
		Now: func() time.Time {
			return time.Date(
				2026,
				time.August,
				1,
				12,
				0,
				0,
				0,
				time.UTC,
			)
		},
	})
	if err != nil {
		t.Fatalf(
			"create service: %v",
			err,
		)
	}

	result, err := service.Get(
		nil,
		Request{},
	)

	if !errors.Is(
		err,
		ErrProductionContextRequired,
	) {
		t.Fatalf(
			"error = %v, want production context required",
			err,
		)
	}
	if trajectoryReader.calls != 0 {
		t.Fatalf(
			"trajectory reader calls = %d, want 0",
			trajectoryReader.calls,
		)
	}
	if weatherReader.calls != 0 {
		t.Fatalf(
			"weather reader calls = %d, want 0",
			weatherReader.calls,
		)
	}
	if projectionReader.calls != 0 {
		t.Fatalf(
			"projection reader calls = %d, want 0",
			projectionReader.calls,
		)
	}
	if result.Version != "" ||
		result.InputFingerprint != "" ||
		!result.GeneratedAt.IsZero() {
		t.Fatalf(
			"result = %#v, want empty result",
			result,
		)
	}
}
