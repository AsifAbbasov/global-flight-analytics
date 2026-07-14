package featurepipeline

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractorcomposition"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

func TestNewInMemoryWrapsExtractorCompositionError(
	t *testing.T,
) {
	_, err := NewInMemory(InMemoryConfig{})
	if !errors.Is(
		err,
		extractorcomposition.ErrAircraftLookupRequired,
	) {
		t.Fatalf(
			"NewInMemory() error = %v, want %v",
			err,
			extractorcomposition.ErrAircraftLookupRequired,
		)
	}

	var constructionErr *ConstructionError
	if !errors.As(err, &constructionErr) {
		t.Fatalf(
			"NewInMemory() error = %T, want *ConstructionError",
			err,
		)
	}
	if constructionErr.Component !=
		ComponentExtractorComposition {
		t.Fatalf(
			"component = %q, want %q",
			constructionErr.Component,
			ComponentExtractorComposition,
		)
	}
}

func TestNewInMemoryWrapsValidatorConfigurationError(
	t *testing.T,
) {
	policy := validator.DefaultPolicy()
	policy.NumericTolerance = 0

	_, err := NewInMemory(InMemoryConfig{
		Extractor: extractorcomposition.Config{
			AircraftLookup: &compositionAircraftLookup{},
		},
		ValidatorPolicy: &policy,
	})
	if !errors.Is(
		err,
		validator.ErrInvalidNumericTolerance,
	) {
		t.Fatalf(
			"NewInMemory() error = %v, want %v",
			err,
			validator.ErrInvalidNumericTolerance,
		)
	}

	var constructionErr *ConstructionError
	if !errors.As(err, &constructionErr) {
		t.Fatalf(
			"NewInMemory() error = %T, want *ConstructionError",
			err,
		)
	}
	if constructionErr.Component !=
		ComponentValidator {
		t.Fatalf(
			"component = %q, want %q",
			constructionErr.Component,
			ComponentValidator,
		)
	}
}

func TestNewInMemoryBuildsProductionComponentsWithSharedClock(
	t *testing.T,
) {
	fixedNow := time.Date(
		2026,
		time.July,
		14,
		18,
		0,
		0,
		0,
		time.FixedZone("test", 4*60*60),
	)
	composition, err := NewInMemory(
		InMemoryConfig{
			Extractor: extractorcomposition.Config{
				AircraftLookup: &compositionAircraftLookup{
					result: aircraft.Aircraft{
						ICAO24: "ABC123",
					},
				},
			},
			Now: func() time.Time {
				return fixedNow
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"NewInMemory() error = %v",
			err,
		)
	}
	if composition.Pipeline == nil ||
		composition.Store == nil ||
		composition.ExtractorComposition == nil ||
		composition.ExtractorComposition.Extractor == nil ||
		composition.Validator == nil {
		t.Fatalf(
			"incomplete composition: %#v",
			composition,
		)
	}
	if !reflect.DeepEqual(
		composition.Versions,
		CurrentVersions(),
	) {
		t.Fatalf(
			"versions = %#v, want %#v",
			composition.Versions,
			CurrentVersions(),
		)
	}

	asOfTime := fixedNow.Add(-time.Minute)
	record, err := composition.Store.Put(
		context.Background(),
		storableFeatures(
			flightfeatures.ValidationStatusValid,
			"fingerprint-clock",
			asOfTime,
		),
	)
	if err != nil {
		t.Fatalf("Store.Put() error = %v", err)
	}
	if !record.StoredAt.Equal(fixedNow.UTC()) {
		t.Fatalf(
			"StoredAt = %v, want %v",
			record.StoredAt,
			fixedNow.UTC(),
		)
	}

	_, report, err := composition.Validator.Validate(
		context.Background(),
		flightfeatures.FlightFeatures{},
	)
	if err != nil {
		t.Fatalf(
			"Validator.Validate() error = %v",
			err,
		)
	}
	if !report.ValidatedAt.Equal(fixedNow.UTC()) {
		t.Fatalf(
			"ValidatedAt = %v, want %v",
			report.ValidatedAt,
			fixedNow.UTC(),
		)
	}
}

func TestCurrentVersionsRemainStable(t *testing.T) {
	want := Versions{
		Pipeline:            "flight-feature-processing-pipeline-v1",
		ExtractorComponents: extractorcomposition.CurrentVersions(),
		Validator:           "flight-feature-validator-v1",
		Store:               "flight-feature-store-v1",
	}

	if got := CurrentVersions(); !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"CurrentVersions() = %#v, want %#v",
			got,
			want,
		)
	}
	if want.Validator != validator.Version ||
		want.Store != featurestore.Version {
		t.Fatal(
			"component version constants diverged from the pipeline manifest",
		)
	}
}

type compositionAircraftLookup struct {
	result aircraft.Aircraft
	err    error
}

func (lookup *compositionAircraftLookup) GetByICAO24(
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
