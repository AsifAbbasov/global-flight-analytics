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
		"apps/api/internal/features/flightfeatures/processing_identity.go",
		"type ProcessingIdentity struct",
		"AircraftEnrichmentMode",
		"AircraftCacheMode",
		"ProcessingComponentVersions",
	)
	require(
		"apps/api/internal/features/flightfeatures/model.go",
		"ProcessingIdentityFingerprint",
		"ProcessingIdentity              ProcessingIdentity",
		`flight-feature-processing-pipeline-v11`,
	)
	require(
		"apps/api/internal/features/extractorcomposition/contracts.go",
		"DefaultConfigWithoutAircraftEnrichment",
		"WithoutAircraftCache",
		"type FeatureExtractor interface",
		`flight-feature-extractor-composition-v6`,
	)
	require(
		"apps/api/internal/features/extractorcomposition/composition.go",
		"validateConfig",
		"newGeographicalBuilder",
		"newAircraftFeatureProvider",
		"ProcessingIdentity:              processingIdentity",
	)
	require(
		"apps/api/internal/features/extractorcomposition/processing_identity_test.go",
		"TestProcessingManifestIsPersistedInFeatureProvenance",
		"TestCompositionSupportsExplicitlyDisabledAircraftEnrichment",
		"TestAircraftCacheModeChangesInputFingerprint",
	)
	require(
		"apps/api/internal/features/aircraftprovider/contracts.go",
		"CacheModeEnabled",
		"CacheModeDisabled",
		`aircraft-feature-provider-v4`,
	)
	require(
		"apps/api/internal/features/aircraftprovider/provider.go",
		"func (provider *Provider) acquire(",
		"provider.cacheMode == CacheModeEnabled",
		"provider.storeCacheLocked(",
	)
	require(
		"apps/api/internal/features/extractor/extractor.go",
		"ProcessingIdentityFingerprint: extractor.fingerprintIdentity",
		"ProcessingIdentity:            extractor.processingIdentity",
	)
	require(
		"apps/api/cmd/materialize-flight-features/main.go",
		"featurepipeline.NewPostgres(",
		"extractorcomposition.DefaultConfig(",
	)
	reject(
		"apps/api/internal/features/extractorcomposition/composition.go",
		"func NewExtractor(",
	)
	require(
		"docs/115_EXTRACTOR_COMPOSITION_REVIEW_HARDENING.md",
		"PROCESSING_MANIFEST_PERSISTENCE=ENFORCED",
		"OPTIONAL_AIRCRAFT_ENRICHMENT=EXPLICIT",
		"AIRCRAFT_CACHE_DISABLE_MODE=EXPLICIT",
		"STALE_REVIEW_FINDINGS=CLASSIFIED",
	)

	if len(failures) == 0 {
		fmt.Println("Extractor composition review audit: PASS")
		return
	}

	fmt.Fprintln(os.Stderr, "Extractor composition review audit: FAIL")
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
