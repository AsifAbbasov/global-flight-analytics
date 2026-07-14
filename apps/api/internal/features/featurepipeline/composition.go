package featurepipeline

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractorcomposition"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

func NewInMemory(
	config InMemoryConfig,
) (*InMemoryComposition, error) {
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

	store := featurestore.NewMemory(
		featurestore.MemoryConfig{
			Now: now,
		},
	)

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

	return &InMemoryComposition{
		Pipeline:             pipeline,
		Store:                store,
		ExtractorComposition: extractorComposition,
		Validator:            featureValidator,
		Versions:             CurrentVersions(),
	}, nil
}

func CurrentVersions() Versions {
	return Versions{
		Pipeline:            Version,
		ExtractorComponents: extractorcomposition.CurrentVersions(),
		Validator:           validator.Version,
		Store:               featurestore.Version,
	}
}
