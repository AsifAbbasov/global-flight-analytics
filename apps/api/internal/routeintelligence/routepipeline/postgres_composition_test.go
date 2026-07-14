package routepipeline

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewPostgresRequiresPool(
	t *testing.T,
) {
	composition, err := NewPostgres(
		PostgresConfig{},
	)
	if !errors.Is(
		err,
		ErrPostgresPoolRequired,
	) {
		t.Fatalf(
			"NewPostgres() error = %v, want %v",
			err,
			ErrPostgresPoolRequired,
		)
	}
	if composition != nil {
		t.Fatalf(
			"composition = %#v, want nil",
			composition,
		)
	}
}

func TestNewPostgresComposesProductionComponents(
	t *testing.T,
) {
	composition, err := NewPostgres(
		PostgresConfig{
			Pool: new(pgxpool.Pool),
		},
	)
	if err != nil {
		t.Fatalf(
			"NewPostgres() error = %v",
			err,
		)
	}

	if composition.Pipeline == nil ||
		composition.Store == nil ||
		composition.TrajectoryRepository == nil ||
		composition.TrajectoryService == nil ||
		composition.AirportRepository == nil ||
		composition.AirportService == nil {
		t.Fatalf(
			"incomplete composition: %#v",
			composition,
		)
	}
	if composition.Versions.Composition !=
		PostgresCompositionVersion ||
		composition.Versions.Pipeline !=
			CurrentVersions() ||
		composition.Versions.Store !=
			"route-postgres-store-v1" {
		t.Fatalf(
			"unexpected versions: %#v",
			composition.Versions,
		)
	}
}
