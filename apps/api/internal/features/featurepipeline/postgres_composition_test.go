package featurepipeline

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractorcomposition"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewPostgresWrapsExtractorCompositionError(
	t *testing.T,
) {
	_, err := NewPostgres(PostgresConfig{})
	if !errors.Is(
		err,
		extractorcomposition.ErrAircraftLookupRequired,
	) {
		t.Fatalf(
			"NewPostgres() error = %v, want %v",
			err,
			extractorcomposition.ErrAircraftLookupRequired,
		)
	}

	assertConstructionComponent(
		t,
		err,
		ComponentExtractorComposition,
	)
}

func TestNewPostgresWrapsValidatorConfigurationError(
	t *testing.T,
) {
	policy := validator.DefaultPolicy()
	policy.NumericTolerance = 0

	_, err := NewPostgres(PostgresConfig{
		Extractor: extractorcomposition.Config{
			AircraftLookup: &postgresCompositionAircraftLookup{},
		},
		ValidatorPolicy: &policy,
		Pool:            &pgxpool.Pool{},
	})
	if !errors.Is(
		err,
		validator.ErrInvalidNumericTolerance,
	) {
		t.Fatalf(
			"NewPostgres() error = %v, want %v",
			err,
			validator.ErrInvalidNumericTolerance,
		)
	}

	assertConstructionComponent(
		t,
		err,
		ComponentValidator,
	)
}

func TestNewPostgresWrapsStoreConstructionError(
	t *testing.T,
) {
	_, err := NewPostgres(PostgresConfig{
		Extractor: extractorcomposition.Config{
			AircraftLookup: &postgresCompositionAircraftLookup{},
		},
	})
	if !errors.Is(
		err,
		featurestore.ErrPostgresPoolRequired,
	) {
		t.Fatalf(
			"NewPostgres() error = %v, want %v",
			err,
			featurestore.ErrPostgresPoolRequired,
		)
	}

	assertConstructionComponent(
		t,
		err,
		ComponentStore,
	)
}

func TestNewPostgresBuildsProductionComponents(
	t *testing.T,
) {
	composition, err := NewPostgres(
		PostgresConfig{
			Extractor: extractorcomposition.Config{
				AircraftLookup: &postgresCompositionAircraftLookup{
					result: aircraft.Aircraft{
						ICAO24: "ABC123",
					},
				},
			},
			Pool: &pgxpool.Pool{},
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
		composition.ExtractorComposition == nil ||
		composition.ExtractorComposition.Extractor ==
			nil ||
		composition.Validator == nil {
		t.Fatalf(
			"incomplete composition: %#v",
			composition,
		)
	}
	if !reflect.DeepEqual(
		composition.Versions,
		CurrentPostgresVersions(),
	) {
		t.Fatalf(
			"versions = %#v, want %#v",
			composition.Versions,
			CurrentPostgresVersions(),
		)
	}
}

func TestCurrentPostgresVersionsRemainStable(
	t *testing.T,
) {
	want := Versions{
		Pipeline:            "flight-feature-processing-pipeline-v1",
		ExtractorComponents: extractorcomposition.CurrentVersions(),
		Validator:           "flight-feature-validator-v1",
		Store:               "flight-feature-postgres-store-v1",
	}

	if got := CurrentPostgresVersions(); !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"CurrentPostgresVersions() = %#v, want %#v",
			got,
			want,
		)
	}
	if PostgresCompositionVersion !=
		"flight-feature-postgres-pipeline-composition-v1" {
		t.Fatalf(
			"PostgresCompositionVersion = %q",
			PostgresCompositionVersion,
		)
	}
	if want.Validator != validator.Version ||
		want.Store != featurestore.PostgresVersion {
		t.Fatal(
			"component versions diverged from the PostgreSQL pipeline manifest",
		)
	}
}

func assertConstructionComponent(
	t *testing.T,
	err error,
	component string,
) {
	t.Helper()

	var constructionErr *ConstructionError
	if !errors.As(err, &constructionErr) {
		t.Fatalf(
			"error = %T, want *ConstructionError",
			err,
		)
	}
	if constructionErr.Component != component {
		t.Fatalf(
			"component = %q, want %q",
			constructionErr.Component,
			component,
		)
	}
}

type postgresCompositionAircraftLookup struct {
	result aircraft.Aircraft
	err    error
}

func (
	lookup *postgresCompositionAircraftLookup,
) GetByICAO24(
	ctx context.Context,
	icao24 string,
) (aircraft.Aircraft, error) {
	if err := ctx.Err(); err != nil {
		return aircraft.Aircraft{}, err
	}
	if lookup.err != nil {
		return aircraft.Aircraft{}, lookup.err
	}

	result := lookup.result
	if result.ICAO24 == "" {
		result.ICAO24 = icao24
	}

	return result, nil
}
