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
			path: "apps/api/internal/features/featurepipeline/contracts.go",
			required: []string{
				"type FeatureWriter interface",
				"Writer    FeatureWriter",
				"func (result Result) Features()",
			},
			forbidden: []string{
				"Features         flightfeatures.FlightFeatures",
				"Store     featurestore.Store",
			},
		},
		{
			path: "apps/api/internal/features/featurepipeline/pipeline.go",
			required: []string{
				"validator.ValidateReport(",
				"ErrContextRequired",
				"pipeline.writer.Put(",
			},
			forbidden: []string{
				"ctx = context.Background()",
				"pipeline.store.Put(",
			},
		},
		{
			path: "apps/api/internal/features/validator/report_validation.go",
			required: []string{
				"ErrInvalidReport",
				"report.ValidatorVersion",
				"report.ValidatedAt.IsZero()",
				"entries must be unique",
				"does not match issue severities",
			},
		},
		{
			path: "apps/api/internal/features/featurepipeline/postgres_composition.go",
			required: []string{
				"ErrPostgresSourceRequired",
				"ErrPostgresSourceAmbiguous",
				"dependencyMissing(config.Executor)",
			},
		},
		{
			path: ".github/workflows/backend-ci.yml",
			required: []string{
				"Run feature pipeline contract audit",
				"go run ./tools/featurepipelinecontractaudit -strict",
				"Verify PostgreSQL feature pipeline",
				"go run ./cmd/verify-postgres-feature-pipeline",
			},
		},
		{
			path: "apps/api/cmd/materialize-flight-features/main.go",
			required: []string{
				"featurepipeline.NewPostgres(",
				"newMaterializationOperation(",
			},
		},
		{
			path: "apps/api/cmd/materialize-flight-features/command_test.go",
			forbidden: []string{
				"Features:        features,",
			},
		},
		{
			path: "apps/api/cmd/verify-postgres-feature-pipeline/main.go",
			required: []string{
				"result.Features().Quality.Status",
			},
			forbidden: []string{
				"result.Features.Quality.Status",
			},
		},
		{
			path: "apps/api/tools/projectaudit/reachability_classification.go",
			forbidden: []string{
				`modulePath + "/internal/features/featurepipeline"`,
			},
		},
	}

	failures := make([]string, 0)
	for _, item := range rules {
		content, readErr := os.ReadFile(filepath.Join(resolved, item.path))
		if readErr != nil {
			failures = append(
				failures,
				fmt.Sprintf("%s: %v", item.path, readErr),
			)
			continue
		}
		text := string(content)
		for _, fragment := range item.required {
			if !strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s missing %q", item.path, fragment),
				)
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s contains forbidden %q", item.path, fragment),
				)
			}
		}
	}

	if len(failures) == 0 {
		fmt.Println("Feature pipeline contract audit: PASS")
		return
	}

	fmt.Fprintln(os.Stderr, "Feature pipeline contract audit: FAIL")
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
