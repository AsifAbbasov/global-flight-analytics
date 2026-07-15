package historicalread

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/jackc/pgx/v5"
)

type failingExecutor struct {
	err error
}

func (executor failingExecutor) Query(
	context.Context,
	string,
	...any,
) (pgx.Rows, error) {
	return nil, executor.err
}

func TestNewPostgresWithExecutorRejectsNil(
	t *testing.T,
) {
	_, err := NewPostgresWithExecutor(nil)
	if !errors.Is(
		err,
		ErrPostgresExecutorRequired,
	) {
		t.Fatalf(
			"error = %v, want %v",
			err,
			ErrPostgresExecutorRequired,
		)
	}
}

func TestNewPostgresWithExecutorUsesProvidedExecutor(
	t *testing.T,
) {
	sentinel := errors.New(
		"executor query failed",
	)
	repository, err := NewPostgresWithExecutor(
		failingExecutor{err: sentinel},
	)
	if err != nil {
		t.Fatalf(
			"compose repository: %v",
			err,
		)
	}

	startTime := time.Date(
		2026,
		time.July,
		15,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	_, err = repository.Read(
		context.Background(),
		Query{
			Window: historicalcontract.TimeWindow{
				StartTime: startTime,
				EndTime: startTime.Add(
					time.Hour,
				),
				AsOfTime: startTime.Add(
					2 * time.Hour,
				),
			},
			Limit: 10,
		},
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf(
			"read error = %v, want wrapped sentinel",
			err,
		)
	}

	var databaseErr *DatabaseError
	if !errors.As(err, &databaseErr) {
		t.Fatalf(
			"read error = %T, want *DatabaseError",
			err,
		)
	}
	if databaseErr.Operation != "read flights" {
		t.Fatalf(
			"operation = %q, want read flights",
			databaseErr.Operation,
		)
	}
}
