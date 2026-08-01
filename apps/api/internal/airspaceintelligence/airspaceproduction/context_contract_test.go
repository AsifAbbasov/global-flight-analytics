package airspaceproduction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/jackc/pgx/v5"
)

type contextContractQueryer struct {
	calls int
}

func (queryer *contextContractQueryer) Query(
	context.Context,
	string,
	...any,
) (pgx.Rows, error) {
	queryer.calls++

	return nil, errors.New("unexpected PostgreSQL query")
}

type contextContractObservationReader struct {
	calls int
}

func (reader *contextContractObservationReader) ListAirspaceObservations(
	context.Context,
	ObservationQuery,
) ([]Observation, error) {
	reader.calls++

	return nil, nil
}

func TestPostgresObservationReaderRejectsNilContextBeforeQuery(
	t *testing.T,
) {
	queryer := &contextContractQueryer{}
	reader := &PostgresObservationReader{
		queryer: queryer,
	}

	_, err := reader.ListAirspaceObservations(
		nil,
		ObservationQuery{},
	)

	if !errors.Is(
		err,
		ErrObservationContextRequired,
	) {
		t.Fatalf(
			"error = %v, want observation context required",
			err,
		)
	}
	if queryer.calls != 0 {
		t.Fatalf(
			"PostgreSQL query calls = %d, want 0",
			queryer.calls,
		)
	}
}

func TestServiceRejectsNilContextBeforeObservationRead(
	t *testing.T,
) {
	reader := &contextContractObservationReader{}
	service, err := New(Config{
		ObservationReader: reader,
		RegionResolver:    region.NewService(),
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

	_, err = service.GetAirspaceRegionAnalytics(
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
	if reader.calls != 0 {
		t.Fatalf(
			"observation reader calls = %d, want 0",
			reader.calls,
		)
	}
}
