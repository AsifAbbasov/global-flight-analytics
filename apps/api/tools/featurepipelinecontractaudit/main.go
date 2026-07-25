package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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
				"validator.NormalizeReport(",
				"validator.ValidateReport(",
				"validated.ValidationReport =",
				"result.Record.Features.ValidationReport.Clone()",
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
				"NormalizeStoredReport",
				"ValidateStoredReport",
				"AuditStateLegacyUnavailable",
				"report.ValidatorVersion",
				"report.ValidatedAt.IsZero()",
				"entries must be unique",
				"does not match issue severities",
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/model.go",
			required: []string{
				"type ValidationReport struct",
				"ValidationAuditStateComplete",
				"ValidationAuditStateLegacyUnavailable",
				"ValidationReport ValidationReport",
				"features.ValidationReport.Clone()",
			},
		},
		{
			path: "apps/api/internal/features/featurestore/validation_audit.go",
			required: []string{
				"normalizeValidationReport",
				"validator.NormalizeStoredReport",
				"validator.ValidateStoredReport",
			},
		},
		{
			path: "database/migrations/027_flight_feature_validation_audit.sql",
			required: []string{
				"ValidationReport",
				"legacy_unavailable",
				"flight_feature_snapshots_validation_report_present",
				"flight_feature_snapshots_validation_report_status_match",
			},
		},
		{
			path: "docs/104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md",
			required: []string{
				"FP-07_COMPOSITION_PLACEMENT=DELIBERATELY_RETAINED_NON_BLOCKING",
				"FP-09_COMPOSITION_HANDLE=DELIBERATELY_RETAINED_NON_BLOCKING",
				"FEATURE_PIPELINE_REVIEW_STATUS=CLOSED",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
			},
		},
		{
			path: "docs/108_FEATURE_PIPELINE_VALIDATION_AUDIT_AND_FINAL_CLOSURE.md",
			required: []string{
				"FEATURE_PIPELINE_VALIDATION_AUDIT_TRAIL=CLOSED",
				"FEATURE_PIPELINE_REVIEW_STATUS=CLOSED",
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
				"Validation audit trail: PASS",
				"validator.AuditStateComplete",
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

	failures = append(
		failures,
		validateFeaturePipelineConfig(
			filepath.Join(
				resolved,
				"apps/api/internal/features/featurepipeline/contracts.go",
			),
		)...,
	)

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

func validateFeaturePipelineConfig(path string) []string {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		return []string{
			fmt.Sprintf(
				"%s: parse feature pipeline contracts: %v",
				path,
				err,
			),
		}
	}

	var config *ast.StructType
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, specification := range general.Specs {
			typeSpec, ok := specification.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "Config" {
				continue
			}
			config, _ = typeSpec.Type.(*ast.StructType)
		}
	}
	if config == nil {
		return []string{
			"feature pipeline Config struct was not found",
		}
	}

	writerValid := false
	processingVersionValid := false
	for _, field := range config.Fields.List {
		if len(field.Names) != 1 {
			continue
		}
		switch field.Names[0].Name {
		case "Writer":
			identifier, ok := field.Type.(*ast.Ident)
			writerValid = ok &&
				identifier.Name == "FeatureWriter"
		case "ProcessingVersion":
			selector, ok := field.Type.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			packageName, ok := selector.X.(*ast.Ident)
			processingVersionValid = ok &&
				packageName.Name == "flightfeatures" &&
				selector.Sel.Name == "ProcessingVersion"
		}
	}

	failures := make([]string, 0, 2)
	if !writerValid {
		failures = append(
			failures,
			"feature pipeline Config.Writer must use FeatureWriter",
		)
	}
	if !processingVersionValid {
		failures = append(
			failures,
			"feature pipeline Config.ProcessingVersion must use flightfeatures.ProcessingVersion",
		)
	}

	return failures
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
