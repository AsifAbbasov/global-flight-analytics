package extractorcomposition

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/aircraftprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/geographicalbuilder"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/operationalbuilder"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/temporalbuilder"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/trajectorybuilder"
)

func New(config Config) (*Composition, error) {
	if config.AircraftLookup == nil {
		return nil, ErrAircraftLookupRequired
	}

	geographicalBuilder, err :=
		geographicalbuilder.New(
			geographicalbuilder.Config{
				GeographicCellPrecision: config.GeographicCellPrecision,
			},
		)
	if err != nil {
		return nil, &ComponentError{
			Component: ComponentGeographicalBuilder,
			Err:       err,
		}
	}

	aircraftProvider, err :=
		aircraftprovider.New(
			aircraftprovider.Config{
				Lookup:           config.AircraftLookup,
				PositiveCacheTTL: config.AircraftPositiveCacheTTL,
				NegativeCacheTTL: config.AircraftNegativeCacheTTL,
				Now:              config.Now,
				IsNotFound:       config.IsAircraftNotFound,
			},
		)
	if err != nil {
		return nil, &ComponentError{
			Component: ComponentAircraftProvider,
			Err:       err,
		}
	}

	featureExtractor, err := extractor.New(
		extractor.Config{
			TemporalBuilder:         temporalbuilder.New(),
			GeographicalBuilder:     geographicalBuilder,
			OperationalBuilder:      operationalbuilder.New(),
			TrajectoryBuilder:       trajectorybuilder.New(),
			AircraftFeatureProvider: aircraftProvider,
			Now:                     config.Now,
		},
	)
	if err != nil {
		return nil, &ComponentError{
			Component: ComponentExtractor,
			Err:       err,
		}
	}

	return &Composition{
		Extractor: featureExtractor,
		Versions:  CurrentVersions(),
	}, nil
}

func NewExtractor(config Config) (
	*extractor.Extractor,
	error,
) {
	composition, err := New(config)
	if err != nil {
		return nil, err
	}

	return composition.Extractor, nil
}

func CurrentVersions() Versions {
	return Versions{
		Composition:         Version,
		Extractor:           extractor.Version,
		AircraftProvider:    aircraftprovider.Version,
		TemporalBuilder:     temporalbuilder.Version,
		GeographicalBuilder: geographicalbuilder.Version,
		OperationalBuilder:  operationalbuilder.Version,
		TrajectoryBuilder:   trajectorybuilder.Version,
	}
}
