package featurepipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractorcomposition"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewPostgresRejectsAmbiguousStorageSource(t *testing.T) {
	_, err := NewPostgres(PostgresConfig{
		Extractor: extractorcomposition.DefaultConfig(
			&postgresCompositionAircraftLookup{},
		),
		Pool:     &pgxpool.Pool{},
		Executor: &fakePostgresExecutor{},
	})
	if !errors.Is(err, ErrPostgresSourceAmbiguous) {
		t.Fatalf(
			"NewPostgres() error = %v, want %v",
			err,
			ErrPostgresSourceAmbiguous,
		)
	}
}

func TestNewPostgresRejectsTypedNilExecutor(t *testing.T) {
	var executor *fakePostgresExecutor

	_, err := NewPostgres(PostgresConfig{
		Extractor: extractorcomposition.DefaultConfig(
			&postgresCompositionAircraftLookup{},
		),
		Executor: executor,
	})
	if !errors.Is(err, ErrPostgresSourceRequired) {
		t.Fatalf(
			"NewPostgres() error = %v, want %v",
			err,
			ErrPostgresSourceRequired,
		)
	}
}

type fakePostgresExecutor struct{}

func (*fakePostgresExecutor) QueryRow(
	context.Context,
	string,
	...any,
) pgx.Row {
	return nil
}

func (*fakePostgresExecutor) Query(
	context.Context,
	string,
	...any,
) (pgx.Rows, error) {
	return nil, featurestore.ErrSnapshotNotFound
}
