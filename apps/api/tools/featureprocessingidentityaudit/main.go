package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := flag.String("root", "", "repository root")
	strict := flag.Bool("strict", true, "fail on contract violation")
	flag.Parse()

	resolved, err := resolveRoot(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	failures := make([]string, 0)
	require := func(path string, fragments ...string) {
		content, readErr := os.ReadFile(filepath.Join(resolved, path))
		if readErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, readErr))
			return
		}
		text := string(content)
		for _, fragment := range fragments {
			if !strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s missing %q", path, fragment),
				)
			}
		}
	}
	reject := func(path string, fragments ...string) {
		content, readErr := os.ReadFile(filepath.Join(resolved, path))
		if readErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, readErr))
			return
		}
		text := string(content)
		for _, fragment := range fragments {
			if strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s contains forbidden %q", path, fragment),
				)
			}
		}
	}

	require(
		"apps/api/internal/features/flightfeatures/model.go",
		"type ProcessingVersion string",
		"CurrentProcessingVersion",
		"LegacyProcessingVersion",
		"ProcessingVersion   ProcessingVersion",
	)
	require(
		"apps/api/internal/features/extractor/contracts.go",
		"FingerprintIdentity     string",
		"AsOfTime   time.Time",
	)
	require(
		"apps/api/internal/domain/aircraft/model.go",
		"MetadataUpdatedAt time.Time",
	)
	require(
		"apps/api/internal/repository/postgres/aircraft_repository.go",
		"GREATEST(",
		"&item.MetadataUpdatedAt",
	)
	require(
		"apps/api/internal/features/aircraftprovider/provider.go",
		"applyTemporalPolicy",
		"aircraft_metadata_newer_than_feature_as_of",
		"normalized.MetadataUpdatedAt.After(asOfTime.UTC())",
	)
	require(
		"apps/api/internal/features/aircraftprovider/provider_test.go",
		"TestProviderAppliesTemporalPolicyAfterCacheLookup",
	)
	require(
		"apps/api/internal/features/extractor/fingerprint.go",
		"type canonicalFingerprintInput struct",
		"ProcessingIdentity string",
		"Aircraft           flightfeatures.AircraftFeatures",
		"fingerprintExtractionInput",
	)
	require(
		"apps/api/internal/features/extractorcomposition/contracts.go",
		"type ProcessingIdentity struct",
		"func DefaultConfig(",
		"WithGeographicCellPrecision",
		"WithAircraftCacheDurations",
		"WithAircraftNotFoundPolicy",
		"WithClock",
		"geographicCellPrecision int",
		"aircraftPositiveCacheTTL",
		"aircraftNegativeCacheTTL",
	)
	reject(
		"apps/api/internal/features/extractorcomposition/contracts.go",
		"type Config struct {\n\tAircraftLookup aircraftprovider.AircraftLookup",
		"\tIsAircraftNotFound            func(error) bool",
		"\tNow func() time.Time\n}",
	)
	require(
		"apps/api/internal/features/extractorcomposition/identity.go",
		"CurrentVersions()",
		"config.geographicCellPrecision",
		"config.aircraftPositiveCacheTTL",
		"config.aircraftNegativeCacheTTL",
		"ErrAircraftNotFoundPolicyVersionRequired",
	)
	reject(
		"apps/api/internal/features/extractorcomposition/identity.go",
		"DefaultGeographicCellPrecision",
		"DefaultPositiveCacheTTL",
		"DefaultNegativeCacheTTL",
		"== 0",
	)
	require(
		"apps/api/internal/features/extractorcomposition/processing_identity_test.go",
		"TestDefaultConfigUsesExplicitDefaults",
		"TestNewRejectsZeroExplicitConfigurationValues",
		"TestConfigurationMethodsDoNotMutateSource",
		"TestGeographicPrecisionChangesInputFingerprint",
		"TestAircraftMetadataChangesInputFingerprint",
		"TestNewRejectsTypedNilAircraftLookup",
	)
	require(
		"apps/api/cmd/materialize-flight-features/main.go",
		"extractorcomposition.DefaultConfig(",
		"WithAircraftNotFoundPolicy(",
	)
	reject(
		"apps/api/cmd/materialize-flight-features/main.go",
		"extractorcomposition."+"Config{",
	)
	require(
		"apps/api/cmd/verify-postgres-feature-pipeline/main.go",
		"extractorcomposition.DefaultConfig(",
	)
	reject(
		"apps/api/cmd/verify-postgres-feature-pipeline/main.go",
		"extractorcomposition."+"Config{",
	)
	reject(
		"apps/api/internal/features/extractorcomposition/composition.go",
		"func NewExtractor(",
	)
	require(
		"apps/api/internal/features/featurestore/contracts.go",
		"ProcessingVersion flightfeatures.ProcessingVersion",
		"processingVersions ...flightfeatures.ProcessingVersion",
	)
	require(
		"apps/api/internal/features/featurestore/memory.go",
		"key.ProcessingVersion",
		"makeLegacyRecordID",
		"normalizeRequestedProcessingVersion",
	)
	require(
		"apps/api/internal/features/featurestore/postgres.go",
		"processing_version",
		"string(key.ProcessingVersion)",
		"AND processing_version = $3",
		"LIMIT $4;",
		`Field: "processing_version"`,
	)
	require(
		"apps/api/internal/features/featurestore/timestamp_consistency_integration_test.go",
		"featureTimestampSnapshotTableDDL",
		"processing_version text NOT NULL",
		"processing_version,",
	)
	require(
		"apps/api/cmd/verify-postgres-feature-pipeline/main.go",
		"Processing version isolation: PASS",
		"alternateFeatures.Provenance.ProcessingVersion",
	)
	require(
		"docs/105_FEATURE_SNAPSHOT_PROCESSING_IDENTITY.md",
		"FP-02_PROCESSING_IDENTITY_STATUS=CLOSED",
		"FEATURE_SNAPSHOT_PROCESSING_IDENTITY=ENFORCED",
	)
	require(
		"docs/109_EXTRACTOR_COMPOSITION_PROCESSING_IDENTITY.md",
		"EXTRACTOR_COMPOSITION_PROCESSING_IDENTITY=ENFORCED",
		"GEOGRAPHIC_PRECISION_FINGERPRINT_COLLISION=CLOSED",
		"AIRCRAFT_METADATA_FINGERPRINT_INPUT=ENFORCED",
	)
	require(
		"docs/110_AIRCRAFT_METADATA_TEMPORAL_SAFETY.md",
		"AIRCRAFT_METADATA_TEMPORAL_GATE=ENFORCED",
		"AIRCRAFT_CACHE_AS_OF_ISOLATION=ENFORCED",
		"FUTURE_AIRCRAFT_METADATA_LEAKAGE=CLOSED",
	)
	require(
		"docs/111_EXTRACTOR_COMPOSITION_EXPLICIT_CONFIG.md",
		"EXTRACTOR_COMPOSITION_EXPLICIT_CONFIG=ENFORCED",
		"ZERO_VALUE_CONFIG_SENTINELS=CLOSED",
		"PRODUCTION_CONFIG_LITERAL_BYPASS=REJECTED",
	)

	migrations, globErr := filepath.Glob(
		filepath.Join(
			resolved,
			"database/migrations/*_flight_feature_processing_identity.sql",
		),
	)
	if globErr != nil || len(migrations) != 1 {
		failures = append(
			failures,
			fmt.Sprintf("processing identity migration count = %d", len(migrations)),
		)
	} else {
		content, readErr := os.ReadFile(migrations[0])
		if readErr != nil {
			failures = append(failures, readErr.Error())
		} else {
			text := string(content)
			for _, fragment := range []string{
				"ADD COLUMN processing_version text",
				"jsonb_set(",
				"flight-feature-processing-legacy-v1",
				"flight_feature_snapshots_processing_identity_uq",
			} {
				if !strings.Contains(text, fragment) {
					failures = append(
						failures,
						fmt.Sprintf("migration missing %q", fragment),
					)
				}
			}
		}
	}

	if len(failures) == 0 {
		fmt.Println("Feature processing identity audit: PASS")
		return
	}

	fmt.Fprintln(os.Stderr, "Feature processing identity audit: FAIL")
	for _, failure := range failures {
		fmt.Fprintln(os.Stderr, "-", failure)
	}
	if *strict {
		os.Exit(1)
	}
}

func resolveRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return filepath.Abs(strings.TrimSpace(explicit))
	}
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, statErr := os.Stat(
			filepath.Join(current, "apps/api/go.mod"),
		); statErr == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root was not found")
		}
		current = parent
	}
}
