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
	if dependencyMissing(config.aircraftLookup) {
		return nil, ErrAircraftLookupRequired
	}
	if config.geographicCellPrecision == 0 {
		return nil, &ComponentError{
			Component: ComponentGeographicalBuilder,
			Err:       ErrGeographicCellPrecisionRequired,
		}
	}
	if config.aircraftPositiveCacheTTL == 0 {
		return nil, &ComponentError{
			Component: ComponentAircraftProvider,
			Err:       ErrAircraftPositiveCacheDurationRequired,
		}
	}
	if config.aircraftNegativeCacheTTL == 0 {
		return nil, &ComponentError{
			Component: ComponentAircraftProvider,
			Err:       ErrAircraftNegativeCacheDurationRequired,
		}
	}

	processingIdentity, fingerprintIdentity, err :=
		resolveProcessingIdentity(config)
	if err != nil {
		return nil, &ComponentError{
			Component: ComponentAircraftProvider,
			Err:       err,
		}
	}

	geographicalBuilder, err :=
		geographicalbuilder.New(
			geographicalbuilder.Config{
				GeographicCellPrecision: config.geographicCellPrecision,
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
				Lookup:           config.aircraftLookup,
				PositiveCacheTTL: config.aircraftPositiveCacheTTL,
				NegativeCacheTTL: config.aircraftNegativeCacheTTL,
				Now:              config.now,
				IsNotFound:       config.isAircraftNotFound,
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
			FingerprintIdentity:     fingerprintIdentity,
			Now:                     config.now,
		},
	)
	if err != nil {
		return nil, &ComponentError{
			Component: ComponentExtractor,
			Err:       err,
		}
	}

	return &Composition{
		Extractor:           featureExtractor,
		Versions:            processingIdentity.Versions,
		ProcessingIdentity:  processingIdentity,
		FingerprintIdentity: fingerprintIdentity,
	}, nil
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
