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
				failures = append(failures, fmt.Sprintf("%s missing %q", path, fragment))
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
				failures = append(failures, fmt.Sprintf("%s contains forbidden %q", path, fragment))
			}
		}
	}

	require(
		"apps/api/internal/features/flightfeatures/model.go",
		"OptionalCoverageScore float64",
		"TrajectoryCreatedAt",
		"AircraftMetadataSourceName",
		"AircraftMetadataProviderVersion",
		"AircraftMetadataRetrievedAt",
		`flight-feature-processing-pipeline-v7`,
	)
	require(
		"apps/api/internal/features/flightfeatures/requirements.go",
		"CurrentGroupRequirementCounts",
		"CurrentGroupFieldCount",
		"Required int",
		"Optional int",
	)
	require(
		"apps/api/internal/features/extractor/quality.go",
		"requiredAvailable",
		"optionalAvailable",
		"OptionalCoverageScore",
		"ErrMixedRequirementGroupEvidence",
	)
	require(
		"apps/api/internal/features/extractor/extractor.go",
		"normalizedTrajectoryCreatedAt",
		"AircraftMetadataRetrievedAt",
		"aircraftMetadataProviderVersion",
	)
	require(
		"apps/api/internal/features/extractor/fingerprint.go",
		"AircraftMetadataSourceName",
		"AircraftMetadataProviderVersion",
	)
	reject(
		"apps/api/internal/features/extractor/extractor.go",
		"return item.EndTime.UTC()",
	)
	require(
		"apps/api/internal/features/extractorcomposition/identity.go",
		"aircraftprovider.MetadataSourceName",
	)
	require(
		"apps/api/internal/features/extractorcomposition/metadata_source_identity_test.go",
		"TestProcessingIdentityIncludesAircraftMetadataSource",
	)
	require(
		"apps/api/internal/features/validator/rules.go",
		"availabilityRequired bool",
		"quality.optional_coverage_score",
		"aircraft_metadata_provenance_required",
		"trajectory_record_timestamps_unavailable",
	)
	require(
		"apps/api/internal/features/extractor/quality_provenance_test.go",
		"TestBuildInitialQualitySeparatesOptionalCoverage",
		"TestExtractorDoesNotInventTrajectoryUpdateTimestamp",
		"TestExtractorRecordsExplicitAircraftMetadataProvenance",
		"TestFingerprintIncludesAircraftMetadataIdentity",
	)
	require(
		"apps/api/internal/features/validator/quality_provenance_test.go",
		"TestValidatorAcceptsSystemUpdateAfterAsOfTime",
		"TestValidatorRequiresAircraftProvenanceWhenMetadataIsAvailable",
	)
	require(
		"docs/113_EXTRACTOR_QUALITY_AND_PROVENANCE_SEMANTICS.md",
		"REQUIRED_COMPLETENESS_OPTIONAL_COVERAGE=SEPARATED",
		"TRAJECTORY_ENDTIME_PROVENANCE_FALLBACK=REMOVED",
		"AIRCRAFT_METADATA_PROVENANCE=EXPLICIT",
		"EXTRACTOR_PROCESSING_GENERATION=v5",
	)

	if len(failures) == 0 {
		fmt.Println("Feature quality and provenance audit: PASS")
		return
	}

	fmt.Fprintln(os.Stderr, "Feature quality and provenance audit: FAIL")
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
		if _, statErr := os.Stat(filepath.Join(current, "apps/api/go.mod")); statErr == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root was not found")
		}
		current = parent
	}
}
