package featurestore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewPostgresWithExecutorRequiresExecutor(
	t *testing.T,
) {
	_, err := NewPostgresWithExecutor(nil, time.Now)
	if !errors.Is(
		err,
		ErrPostgresExecutorRequired,
	) {
		t.Fatalf(
			"NewPostgresWithExecutor() error = %v, want %v",
			err,
			ErrPostgresExecutorRequired,
		)
	}
}

func TestNewPostgresWithExecutorBuildsStore(
	t *testing.T,
) {
	store, err := NewPostgresWithExecutor(
		executorStub{},
		time.Now,
	)
	if err != nil {
		t.Fatalf(
			"NewPostgresWithExecutor() error = %v",
			err,
		)
	}
	if store == nil {
		t.Fatal(
			"NewPostgresWithExecutor() returned nil store",
		)
	}
}

func TestPostgresExecutorVersionRemainsStable(
	t *testing.T,
) {
	if PostgresExecutorVersion !=
		"flight-feature-postgres-executor-v1" {
		t.Fatalf(
			"PostgresExecutorVersion = %q",
			PostgresExecutorVersion,
		)
	}
}

func TestProductionPostgresTypesImplementExecutor(
	t *testing.T,
) {
	var _ PostgresExecutor = (*pgxpool.Pool)(nil)
	var _ PostgresExecutor = (pgx.Tx)(nil)
}

type executorStub struct{}

func (executorStub) QueryRow(
	context.Context,
	string,
	...any,
) pgx.Row {
	return nil
}

func (executorStub) Query(
	context.Context,
	string,
	...any,
) (pgx.Rows, error) {
	return nil, nil
}
