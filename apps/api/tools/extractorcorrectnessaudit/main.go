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
		"apps/api/internal/features/extractor/extractor.go",
		"dependencyMissing(config.TemporalBuilder)",
		"ErrContextRequired",
		"validateSnapshotEvidence(item, request.AsOfTime)",
		"quality, err := buildInitialQuality(",
	)
	reject(
		"apps/api/internal/features/extractor/extractor.go",
		"ctx = context.Background()",
		"func clamp01(",
	)
	require(
		"apps/api/internal/features/extractor/snapshot_validation.go",
		"point.ObservedAt.After(cutoff)",
		"segment.StartTime.After(cutoff)",
		"segment.EndTime.After(cutoff)",
		"gap.StartTime.After(cutoff)",
		"gap.EndTime.After(cutoff)",
	)
	require(
		"apps/api/internal/features/extractor/quality.go",
		"evidence.AvailableFieldCount > evidence.TotalFieldCount",
		"math.IsNaN(inputQualityScore)",
		"math.IsInf(inputQualityScore, 0)",
		"ErrInvalidInputQualityScore",
	)
	require(
		"apps/api/internal/features/extractor/fingerprint.go",
		"aircraft.CanonicalICAO24(item.ICAO24)",
		"strings.TrimSpace(item.Callsign)",
	)
	require(
		"apps/api/internal/features/extractor/correctness_hardening_test.go",
		"TestExtractorRejectsNestedEvidenceAfterAsOfTime",
		"TestFingerprintNormalizesSemanticAircraftIdentity",
		"TestBuildInitialQualityRejectsInvalidEvidenceCounts",
		"TestBuildInitialQualityRejectsNonFiniteInputQuality",
		"TestNewRejectsTypedNilBuilder",
		"TestExtractorStopsAfterAircraftProviderCancellation",
	)
	require(
		"apps/api/internal/features/extractor/contracts.go",
		`const Version = "flight-feature-extractor-v6"`,
	)
	require(
		"apps/api/internal/features/extractorcomposition/contracts.go",
		`const Version = "flight-feature-extractor-composition-v6"`,
	)
	require(
		"apps/api/internal/features/featurepipeline/contracts.go",
		`const Version = "flight-feature-processing-pipeline-v8"`,
	)
	require(
		"apps/api/internal/features/flightfeatures/model.go",
		`CurrentProcessingVersion ProcessingVersion = "flight-feature-processing-pipeline-v8"`,
	)
	require(
		"docs/112_EXTRACTOR_INPUT_CORRECTNESS_HARDENING.md",
		"EXTRACTOR_NESTED_TEMPORAL_GUARD=ENFORCED",
		"EXTRACTOR_INVALID_MATH_MASKING=CLOSED",
		"EXTRACTOR_PROCESSING_GENERATION=v4",
	)

	if len(failures) == 0 {
		fmt.Println("Extractor correctness audit: PASS")
		return
	}

	fmt.Fprintln(os.Stderr, "Extractor correctness audit: FAIL")
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
