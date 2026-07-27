package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type rule struct {
	path      string
	required  []string
	forbidden []string
}

func main() {
	root := flag.String("root", "", "repository root")
	strict := flag.Bool("strict", true, "fail on contract violation")
	flag.Parse()

	resolved, err := resolveRoot(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rules := []rule{
		{
			path: "apps/api/internal/features/flightfeatures/schema.go",
			required: []string{
				`Name:        "geographical.minimum_latitude"`,
				`Name:        "geographical.maximum_latitude"`,
				`Name:        "geographical.minimum_longitude"`,
				`Name:        "geographical.maximum_longitude"`,
				"SchemaForVersion(SchemaVersionV1)",
				"DefinitionByNameForVersion(",
			},
			forbidden: []string{
				`Name:        "geographical.geographic_cell_precision"`,
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/requirements.go",
			required: []string{
				"GeographicalRequiredFeatureFieldCount = 15",
				"GroupRequirementCountsForVersion(",
				"GroupFieldCountForVersion(",
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/schema_registry.go",
			required: []string{
				"SupportedSchemaVersions",
				"SchemaForVersion",
				"CompatibilityBetween",
				"func (value AvailabilityStatus) IsValid() bool",
				"func (value FeatureGroup) IsValid() bool",
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/model.go",
			required: []string{
				`flight-feature-processing-pipeline-v7`,
				"processing-configuration mirror",
				"ProcessingIdentityFingerprint",
				"AircraftMetadataProviderVersion",
				"AircraftMetadataRetrievedAt",
			},
		},
		{
			path: "apps/api/internal/features/geographicalbuilder/contracts.go",
			required: []string{
				`geographical-feature-builder-v2`,
				"flightfeatures.GeographicalRequiredFeatureFieldCount",
			},
			forbidden: []string{
				"GeographicalFeatureFieldCount  = 11",
			},
		},
		{
			path: "apps/api/internal/features/temporalbuilder/contracts.go",
			required: []string{
				"flightfeatures.TemporalRequiredFeatureFieldCount",
			},
		},
		{
			path: "apps/api/internal/features/operationalbuilder/contracts.go",
			required: []string{
				"flightfeatures.OperationalRequiredFeatureFieldCount",
			},
		},
		{
			path: "apps/api/internal/features/trajectorybuilder/contracts.go",
			required: []string{
				"flightfeatures.TrajectoryRequiredFeatureFieldCount",
			},
		},
		{
			path: "apps/api/internal/features/validator/contracts.go",
			required: []string{
				`flight-feature-validator-v3`,
			},
		},
		{
			path: "apps/api/internal/features/featurepipeline/contracts.go",
			required: []string{
				`flight-feature-processing-pipeline-v7`,
			},
		},
		{
			path: "apps/api/internal/features/featurestore/contracts.go",
			required: []string{
				"OutputFingerprint string",
			},
		},
		{
			path: "apps/api/internal/features/featurestore/snapshot_payload_v1.go",
			required: []string{
				"snapshotPayloadVersionV1",
				"OutputFingerprint",
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/schema_review_hardening_test.go",
			required: []string{
				"TestCurrentSchemaMatchesExactGroupFieldContract",
				"TestGeographicalSchemaContainsEveryAnalyticalField",
				"TestSchemaRegistrySupportsExplicitVersionLookup",
			},
		},
		{
			path: "docs/118_FLIGHT_FEATURES_SCHEMA_REVIEW_HARDENING.md",
			required: []string{
				"FLIGHT_FEATURES_SCHEMA_MODEL_ALIGNMENT=CLOSED",
				"GEOGRAPHICAL_COMPLETENESS_DENOMINATOR=15",
				"FLIGHT_FEATURES_PROCESSING_GENERATION=v7",
				"FLIGHT_FEATURES_REVIEW_STATUS=CLOSED",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
			},
		},
		{
			path: ".github/workflows/backend-ci.yml",
			required: []string{
				"Run flight features review audit",
				"go run ./tools/flightfeaturesreviewaudit -strict",
			},
		},
	}

	failures := make([]string, 0)
	for _, item := range rules {
		content, readErr := os.ReadFile(filepath.Join(resolved, item.path))
		if readErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.path, readErr))
			continue
		}
		text := string(content)
		for _, fragment := range item.required {
			if !strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s missing %q", item.path, fragment))
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s contains forbidden %q", item.path, fragment))
			}
		}
	}

	if len(failures) == 0 {
		fmt.Println("Flight features review audit: PASS")
		return
	}
	fmt.Fprintln(os.Stderr, "Flight features review audit: FAIL")
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
