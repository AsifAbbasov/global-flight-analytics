package featurepipeline

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractorcomposition"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PostgresCompositionVersion = "flight-feature-postgres-pipeline-composition-v1"
	ComponentStore             = "store"
)

type PostgresConfig struct {
	Extractor       extractorcomposition.Config
	ValidatorPolicy *validator.Policy
	Pool            *pgxpool.Pool
	Executor        featurestore.PostgresExecutor
	Now             func() time.Time
}

type PostgresComposition struct {
	Pipeline             *Pipeline
	Store                *featurestore.PostgresStore
	ExtractorComposition *extractorcomposition.Composition
	Validator            *validator.Validator
	Versions             Versions
}

func NewPostgres(
	config PostgresConfig,
) (*PostgresComposition, error) {
	now := config.Now
	if now == nil {
		now = time.Now
	}

	extractorConfig := config.Extractor
	extractorConfig.Now = now

	extractorComposition, err :=
		extractorcomposition.New(extractorConfig)
	if err != nil {
		return nil, &ConstructionError{
			Component: ComponentExtractorComposition,
			Err:       err,
		}
	}

	featureValidator, err := validator.New(
		validator.Config{
			Policy: config.ValidatorPolicy,
			Now:    now,
		},
	)
	if err != nil {
		return nil, &ConstructionError{
			Component: ComponentValidator,
			Err:       err,
		}
	}

	var store *featurestore.PostgresStore
	if config.Executor != nil {
		store, err = featurestore.NewPostgresWithExecutor(
			config.Executor,
			now,
		)
	} else {
		store, err = featurestore.NewPostgres(
			featurestore.PostgresConfig{
				Pool: config.Pool,
				Now:  now,
			},
		)
	}
	if err != nil {
		return nil, &ConstructionError{
			Component: ComponentStore,
			Err:       err,
		}
	}

	pipeline, err := New(Config{
		Extractor: extractorComposition.Extractor,
		Validator: featureValidator,
		Store:     store,
	})
	if err != nil {
		return nil, &ConstructionError{
			Component: ComponentPipeline,
			Err:       err,
		}
	}

	return &PostgresComposition{
		Pipeline:             pipeline,
		Store:                store,
		ExtractorComposition: extractorComposition,
		Validator:            featureValidator,
		Versions:             CurrentPostgresVersions(),
	}, nil
}

func CurrentPostgresVersions() Versions {
	return Versions{
		Pipeline:            Version,
		ExtractorComponents: extractorcomposition.CurrentVersions(),
		Validator:           validator.Version,
		Store:               featurestore.PostgresVersion,
	}
}
