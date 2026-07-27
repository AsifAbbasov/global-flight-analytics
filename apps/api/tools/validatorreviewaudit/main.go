package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	strict := flag.Bool("strict", false, "fail on any validator review contract violation")
	flag.Parse()
	root, err := repositoryRoot()
	if err != nil {
		fail(err)
	}
	checks := []struct {
		path     string
		contains []string
		excludes []string
	}{
		{
			path:     "apps/api/internal/features/validator/contracts.go",
			contains: []string{"flight-feature-validator-v6", "dimensionless relative tolerance"},
		},
		{
			path:     "apps/api/internal/features/featurepipeline/contracts.go",
			contains: []string{"flight-feature-processing-pipeline-v12"},
		},
		{
			path:     "apps/api/internal/features/validator/validator.go",
			contains: []string{"ErrContextRequired", "collectCurrentGroupLimitations"},
			excludes: []string{"ctx = context.Background()"},
		},
		{
			path:     "apps/api/internal/features/validator/rules.go",
			contains: []string{"IssueSeverityError", "evidence_limitation_required"},
			excludes: []string{"relationshipSeverity(", "markValidAsWarning bool", "+collector.tolerance", "-collector.tolerance"},
		},
		{
			path:     "apps/api/internal/features/validator/unavailable_payload.go",
			contains: []string{"unavailable_group_payload_not_zero"},
		},
		{
			path: "apps/api/internal/features/validator/validator_review_hardening_test.go",
			contains: []string{
				"TestValidatorRejectsNonFiniteValueInPartialGroup",
				"TestValidatorRebuildsQualityLimitationsFromCurrentEvidence",
				"TestValidatorRejectsResidualPayloadInUnavailableGroup",
				"TestValidatorRejectsAvailableOperationalGroupWithoutSupport",
			},
		},
		{
			path:     ".github/workflows/backend-ci.yml",
			contains: []string{"go run ./tools/validatorreviewaudit -strict"},
		},
		{
			path:     "docs/123_VALIDATOR_REVIEW_HARDENING.md",
			contains: []string{"OPEN_CONFIRMED_FINDINGS=0", "UNCLASSIFIED_FINDINGS=0", "DEFERRED_FINDINGS=0"},
		},
	}
	violations := 0
	for _, check := range checks {
		data, readErr := os.ReadFile(filepath.Join(root, check.path))
		if readErr != nil {
			fmt.Printf("FAIL %s: %v\n", check.path, readErr)
			violations++
			continue
		}
		value := string(data)
		for _, expected := range check.contains {
			if !strings.Contains(value, expected) {
				fmt.Printf("FAIL %s: missing %q\n", check.path, expected)
				violations++
			}
		}
		for _, forbidden := range check.excludes {
			if strings.Contains(value, forbidden) {
				fmt.Printf("FAIL %s: contains forbidden %q\n", check.path, forbidden)
				violations++
			}
		}
	}
	if violations > 0 {
		fmt.Printf("Validator review audit: FAIL (%d violations)\n", violations)
		if *strict {
			os.Exit(1)
		}
		return
	}
	fmt.Println("Validator review audit: PASS")
}

func repositoryRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "apps", "api", "go.mod")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root not found")
		}
		current = parent
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
