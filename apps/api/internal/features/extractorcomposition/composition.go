package extractorcomposition

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/aircraftprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/geographicalbuilder"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/operationalbuilder"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/temporalbuilder"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/trajectorybuilder"
)

func New(config Config) (*Composition, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	processingIdentity, fingerprintIdentity, err :=
		resolveProcessingIdentity(config)
	if err != nil {
		return nil, &ComponentError{
			Component: ComponentAircraftProvider,
			Err:       err,
		}
	}

	geographicalBuilder, err := newGeographicalBuilder(config)
	if err != nil {
		return nil, err
	}
	aircraftFeatureProvider, metadataSourceName, providerVersion, err :=
		newAircraftFeatureProvider(config)
	if err != nil {
		return nil, err
	}

	featureExtractor, err := extractor.New(
		extractor.Config{
			TemporalBuilder:                 temporalbuilder.New(),
			GeographicalBuilder:             geographicalBuilder,
			OperationalBuilder:              operationalbuilder.New(),
			TrajectoryBuilder:               trajectorybuilder.New(),
			AircraftFeatureProvider:         aircraftFeatureProvider,
			AircraftMetadataSourceName:      metadataSourceName,
			AircraftMetadataProviderVersion: providerVersion,
			FingerprintIdentity:             fingerprintIdentity,
			ProcessingIdentity:              processingIdentity,
			Now:                             config.now,
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

func validateConfig(config Config) error {
	if config.geographicCellPrecision == 0 {
		return &ComponentError{
			Component: ComponentGeographicalBuilder,
			Err:       ErrGeographicCellPrecisionRequired,
		}
	}

	switch config.aircraftEnrichmentMode {
	case flightfeatures.AircraftEnrichmentModeEnabled:
		if dependencyMissing(config.aircraftLookup) {
			return ErrAircraftLookupRequired
		}
		if config.aircraftCacheMode == aircraftprovider.CacheModeEnabled {
			if config.aircraftPositiveCacheTTL == 0 {
				return &ComponentError{
					Component: ComponentAircraftProvider,
					Err:       ErrAircraftPositiveCacheDurationRequired,
				}
			}
			if config.aircraftNegativeCacheTTL == 0 {
				return &ComponentError{
					Component: ComponentAircraftProvider,
					Err:       ErrAircraftNegativeCacheDurationRequired,
				}
			}
		}
	case flightfeatures.AircraftEnrichmentModeDisabled:
		if !dependencyMissing(config.aircraftLookup) ||
			config.aircraftCacheMode != aircraftprovider.CacheModeDisabled ||
			config.aircraftPositiveCacheTTL != 0 ||
			config.aircraftNegativeCacheTTL != 0 ||
			config.isAircraftNotFound != nil ||
			config.aircraftNotFoundPolicyVersion !=
				disabledAircraftNotFoundPolicyVersion {
			return ErrAircraftEnrichmentConfigurationAmbiguous
		}
	default:
		return ErrAircraftEnrichmentModeRequired
	}

	return nil
}

func newGeographicalBuilder(config Config) (*geographicalbuilder.Builder, error) {
	builder, err := geographicalbuilder.New(
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
	return builder, nil
}

func newAircraftFeatureProvider(
	config Config,
) (extractor.AircraftFeatureProvider, string, string, error) {
	if config.aircraftEnrichmentMode ==
		flightfeatures.AircraftEnrichmentModeDisabled {
		return nil, "", "", nil
	}

	provider, err := aircraftprovider.New(
		aircraftprovider.Config{
			Lookup:           config.aircraftLookup,
			CacheMode:        config.aircraftCacheMode,
			PositiveCacheTTL: config.aircraftPositiveCacheTTL,
			NegativeCacheTTL: config.aircraftNegativeCacheTTL,
			Now:              config.now,
			IsNotFound:       config.isAircraftNotFound,
		},
	)
	if err != nil {
		return nil, "", "", &ComponentError{
			Component: ComponentAircraftProvider,
			Err:       err,
		}
	}

	return provider,
		aircraftprovider.MetadataSourceName,
		aircraftprovider.Version,
		nil
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
