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

	require(
		"apps/api/internal/features/flightfeatures/model.go",
		"type ProcessingVersion string",
		"CurrentProcessingVersion",
		"LegacyProcessingVersion",
		"ProcessingVersion   ProcessingVersion",
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
		"apps/api/cmd/verify-postgres-feature-pipeline/main.go",
		"Processing version isolation: PASS",
		"alternateFeatures.Provenance.ProcessingVersion",
	)
	require(
		"docs/105_FEATURE_SNAPSHOT_PROCESSING_IDENTITY.md",
		"FP-02_PROCESSING_IDENTITY_STATUS=CLOSED",
		"FEATURE_SNAPSHOT_PROCESSING_IDENTITY=ENFORCED",
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
