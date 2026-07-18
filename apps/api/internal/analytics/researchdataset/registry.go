package researchdataset

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/sourceconstraints"
)

var (
	ErrDatasetUnknown = errors.New(
		"research dataset is unknown",
	)
	ErrDatasetNotAdopted = errors.New(
		"research dataset is not adopted for bounded offline research",
	)
	ErrManifestInvalid = errors.New(
		"research dataset manifest is invalid",
	)
	ErrBlockedTable = errors.New(
		"research dataset manifest selects a blocked table",
	)
)

const (
	megabyte = int64(1024 * 1024)
	gigabyte = int64(1024 * 1024 * 1024)
)

func DefaultRegistry() map[ID]Profile {
	openSkyStatic := func(
		id ID,
		satelliteDerived bool,
	) sourceconstraints.SourceProfile {
		return sourceconstraints.SourceProfile{
			ID:                        string(id),
			Class:                     sourceconstraints.SourceClassPublicStaticDataset,
			FreeAccess:                true,
			ExternallyCollected:       true,
			SatelliteDerived:          satelliteDerived,
			SupportsHistoricalFlights: true,
			AttributionRequired:       true,
			AttributionText:           sourceconstraints.OpenSkyAttribution,
		}
	}

	return map[ID]Profile{
		IDEmergencyReference: {
			ID:        IDEmergencyReference,
			Name:      "Reference Datasets for In-Flight Emergency Situations",
			Selection: SelectionAdopted,
			Source:    openSkyStatic(IDEmergencyReference, false),
			Purposes: []string{
				"validate retention and classification of observed 7500, 7600, and 7700 transponder codes",
				"benchmark trajectory evidence around observed special transponder codes",
			},
			RequiredLabels: []string{
				"observed transponder code only",
				"not a confirmed incident",
				"research use only",
			},
			BlockedPurposes: []string{
				"infer incident cause",
				"confirm an emergency",
				"import third-party incident narratives as surveillance fact",
			},
			MaximumDownloadBytes:        512 * megabyte,
			MaximumRecords:              2_000_000,
			RequiresLicenseReview:       true,
			ProductionDependencyAllowed: false,
		},
		IDClimbingAircraft: {
			ID:        IDClimbingAircraft,
			Name:      "Climbing Aircraft Dataset",
			Selection: SelectionAdopted,
			Source:    openSkyStatic(IDClimbingAircraft, false),
			Purposes: []string{
				"offline climb prediction benchmark",
				"validate altitude and speed projection error by aircraft type",
			},
			RequiredLabels: []string{
				"historical benchmark",
				"2017 distribution",
				"not guaranteed representative of current regional traffic",
			},
			BlockedPurposes: []string{
				"production training dependency",
				"operational trajectory prediction certification",
			},
			MaximumDownloadBytes:        2 * gigabyte,
			MaximumRecords:              5_000_000,
			RequiresRegionFilter:        false,
			RequiresLicenseReview:       true,
			ProductionDependencyAllowed: false,
		},
		IDTrinoSnapshot2026: {
			ID:        IDTrinoSnapshot2026,
			Name:      "Complete One Day Trino Tables Snapshot, March 2026",
			Selection: SelectionAdopted,
			Source:    openSkyStatic(IDTrinoSnapshot2026, false),
			Purposes: []string{
				"offline OpenSky schema compatibility verification",
				"bounded state-vector, flight, ADS-B, and MLAT replay tests",
			},
			RequiredLabels: []string{
				"static one-day snapshot",
				"bounded selected tables only",
				"research use only",
			},
			BlockedPurposes: []string{
				"automatic full-snapshot import",
				"production data dependency",
				"satellite ADS-C analysis",
			},
			AllowedTables: []string{
				"flights_data4",
				"flights_data5",
				"identification_data4",
				"operational_status_data4",
				"position_data4",
				"readsb_mlat_sv",
				"state_vectors_data4",
				"velocity_data4",
			},
			BlockedTables: []string{
				"readsb_adsc_sv",
			},
			MaximumDownloadBytes:        4 * gigabyte,
			MaximumRecords:              10_000_000,
			RequiresRegionFilter:        true,
			RequiresLicenseReview:       true,
			ProductionDependencyAllowed: false,
		},
		IDWeeklyStateVectors: {
			ID:        IDWeeklyStateVectors,
			Name:      "Weekly 24 Hours of State Vector Data, 2017-2022",
			Selection: SelectionAdopted,
			Source:    openSkyStatic(IDWeeklyStateVectors, false),
			Purposes: []string{
				"bounded external Historical Replay benchmark",
				"gap detection and trajectory reconstruction verification",
			},
			RequiredLabels: []string{
				"Monday-only sample",
				"not representative of full weekly seasonality",
				"bounded regional sample",
			},
			BlockedPurposes: []string{
				"global bulk ingestion",
				"general weekly seasonality claims",
				"production data dependency",
			},
			MaximumDownloadBytes:        2 * gigabyte,
			MaximumRecords:              5_000_000,
			RequiresRegionFilter:        true,
			RequiresLicenseReview:       true,
			ProductionDependencyAllowed: false,
		},
		IDPRCTakeoffWeight: {
			ID:        IDPRCTakeoffWeight,
			Name:      "OpenSky and EUROCONTROL PRC Data Challenge 2024 Dataset",
			Selection: SelectionAdopted,
			Source:    openSkyStatic(IDPRCTakeoffWeight, false),
			Purposes: []string{
				"offline take-off weight estimation benchmark",
				"build bounded aircraft-type and route-distance calibration experiments",
			},
			RequiredLabels: []string{
				"estimated take-off weight",
				"not live actual take-off weight",
				"model applicability required",
			},
			BlockedPurposes: []string{
				"claim live measured aircraft mass",
				"import closed EUROCONTROL operational feeds",
				"store the complete trajectory archive in project PostgreSQL",
			},
			MaximumDownloadBytes:        2 * gigabyte,
			MaximumRecords:              2_000_000,
			RequiresLicenseReview:       true,
			ProductionDependencyAllowed: false,
		},
		IDRawPhysicalLayer: deferredProfile(
			IDRawPhysicalLayer,
			"OpenSky Raw Physical Layer Data",
			"raw Mode S physical-layer decoding is outside the current canonical State Vector architecture",
		),
		IDLocaRDS: deferredProfile(
			IDLocaRDS,
			"LocaRDS",
			"crowdsourced sensor localization research is outside the current product scope",
		),
		IDCOVID19: deferredProfile(
			IDCOVID19,
			"COVID-19 Flight Dataset",
			"the dataset ended in 2022 and does not strengthen current regional production analytics",
		),
		IDAircraftMetadata: deferredProfile(
			IDAircraftMetadata,
			"OpenSky Aircraft Metadata Database",
			"mixed-source metadata requires record-level licence and provenance review",
		),
		IDGICB: deferredProfile(
			IDGICB,
			"World Aircraft Common GICB Capabilities",
			"Comm-B capability analysis requires a separate raw Mode S decoding scope",
		),
		IDADSC: {
			ID:        IDADSC,
			Name:      "OpenSky ADS-C Dataset",
			Selection: SelectionBlocked,
			Source:    openSkyStatic(IDADSC, true),
			RequiredLabels: []string{
				"blocked by fixed project source constraints",
			},
			BlockedPurposes: []string{
				"satellite-derived aviation surveillance",
				"navigation intent ingestion",
				"oceanic coverage claims",
			},
			BlockedTables: []string{
				"readsb_adsc_sv",
			},
			ProductionDependencyAllowed: false,
		},
	}
}

func deferredProfile(
	id ID,
	name string,
	reason string,
) Profile {
	return Profile{
		ID:        id,
		Name:      name,
		Selection: SelectionDeferred,
		Source: sourceconstraints.SourceProfile{
			ID:                        string(id),
			Class:                     sourceconstraints.SourceClassPublicStaticDataset,
			FreeAccess:                true,
			ExternallyCollected:       true,
			SupportsHistoricalFlights: true,
			AttributionRequired:       true,
			AttributionText:           sourceconstraints.OpenSkyAttribution,
		},
		BlockedPurposes: []string{
			reason,
			"production data dependency",
		},
		RequiresLicenseReview:       true,
		ProductionDependencyAllowed: false,
	}
}

func ProfileByID(
	id ID,
) (Profile, error) {
	profile, exists := DefaultRegistry()[id]
	if !exists {
		return Profile{}, fmt.Errorf(
			"%w: %s",
			ErrDatasetUnknown,
			id,
		)
	}
	profile.Purposes = append(
		[]string(nil),
		profile.Purposes...,
	)
	profile.RequiredLabels = append(
		[]string(nil),
		profile.RequiredLabels...,
	)
	profile.BlockedPurposes = append(
		[]string(nil),
		profile.BlockedPurposes...,
	)
	profile.AllowedTables = append(
		[]string(nil),
		profile.AllowedTables...,
	)
	profile.BlockedTables = append(
		[]string(nil),
		profile.BlockedTables...,
	)
	return profile, nil
}

func AdoptedIDs() []ID {
	registry := DefaultRegistry()
	result := make([]ID, 0)
	for id, profile := range registry {
		if profile.Selection == SelectionAdopted {
			result = append(result, id)
		}
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left] < result[right]
	})
	return result
}

func EvaluateSourceBoundary(
	id ID,
) (sourceconstraints.Decision, error) {
	profile, err := ProfileByID(id)
	if err != nil {
		return sourceconstraints.Decision{}, err
	}
	return sourceconstraints.Evaluate(
		sourceconstraints.Request{
			Constraints: sourceconstraints.FixedProjectConstraints(),
			Source:      profile.Source,
			Capability:  sourceconstraints.CapabilityHistoricalFlightObservation,
		},
	)
}

func contains(
	values []string,
	candidate string,
) bool {
	candidate = strings.TrimSpace(candidate)
	for _, value := range values {
		if strings.EqualFold(
			strings.TrimSpace(value),
			candidate,
		) {
			return true
		}
	}
	return false
}
