package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type auditFailure struct {
	Check  string
	Detail string
}

type fileRule struct {
	Name            string
	Path            string
	Required        []string
	RequiredCompact []string
	Forbidden       []string
}

type findingDisposition string

const (
	dispositionFixed    findingDisposition = "FIXED"
	dispositionRetained findingDisposition = "DELIBERATELY_RETAINED"
	dispositionRejected findingDisposition = "REJECTED_NON_BLOCKING"
)

type finding struct {
	ID          string
	Name        string
	Disposition findingDisposition
	Evidence    []string
}

var analyticalCoreFindings = []finding{
	{
		ID:          "AC-01",
		Name:        "Eligibility must precede aircraft deduplication",
		Disposition: dispositionFixed,
		Evidence: []string{
			"97_ANALYTICAL_CONTRIBUTOR_SEMANTICS_HARDENING.md",
		},
	},
	{
		ID:          "AC-02",
		Name:        "Materially future observations must not contribute",
		Disposition: dispositionFixed,
		Evidence: []string{
			"97_ANALYTICAL_CONTRIBUTOR_SEMANTICS_HARDENING.md",
		},
	},
	{
		ID:          "AC-03",
		Name:        "Airport Activity must belong to a concrete airport",
		Disposition: dispositionFixed,
		Evidence: []string{
			"98_AIRPORT_AND_GEOGRAPHIC_METRIC_INTEGRITY.md",
		},
	},
	{
		ID:          "AC-04",
		Name:        "Traffic Density numerator and area must share one region",
		Disposition: dispositionFixed,
		Evidence: []string{
			"98_AIRPORT_AND_GEOGRAPHIC_METRIC_INTEGRITY.md",
		},
	},
	{
		ID:          "AC-05",
		Name:        "Parallel calculator and registry foundation is not runtime architecture",
		Disposition: dispositionRetained,
		Evidence: []string{
			"100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md",
		},
	},
	{
		ID:          "AC-06",
		Name:        "Metric execution must depend on narrow behavior",
		Disposition: dispositionFixed,
		Evidence: []string{
			"100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md",
		},
	},
	{
		ID:          "AC-07",
		Name:        "Traffic Density must reject non-finite arithmetic",
		Disposition: dispositionFixed,
		Evidence: []string{
			"97_ANALYTICAL_CONTRIBUTOR_SEMANTICS_HARDENING.md",
		},
	},
	{
		ID:          "AC-08",
		Name:        "Coverage and freshness trust must not come from caller snapshots",
		Disposition: dispositionFixed,
		Evidence: []string{
			"99_PROVENANCE_AND_ANALYTICAL_TRUST_HARDENING.md",
			"101_SERVER_OWNED_QUALITY_METRICS.md",
		},
	},
	{
		ID:          "AC-09",
		Name:        "Published analytical provenance must be complete",
		Disposition: dispositionFixed,
		Evidence: []string{
			"99_PROVENANCE_AND_ANALYTICAL_TRUST_HARDENING.md",
			"101_SERVER_OWNED_QUALITY_METRICS.md",
		},
	},
	{
		ID:          "AC-10",
		Name:        "A zero analytical reference time must be rejected",
		Disposition: dispositionFixed,
		Evidence: []string{
			"100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md",
		},
	},
	{
		ID:          "AC-11",
		Name:        "Accepted UUID identifiers must be canonicalized",
		Disposition: dispositionFixed,
		Evidence: []string{
			"100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md",
		},
	},
	{
		ID:          "AC-12",
		Name:        "Public numeric precision must have an explicit contract",
		Disposition: dispositionFixed,
		Evidence: []string{
			"102_ANALYTICAL_CORE_REVIEW_CLOSURE.md",
		},
	},
	{
		ID:          "AC-13",
		Name:        "Mechanical function-length thresholds are not correctness rules",
		Disposition: dispositionRejected,
		Evidence: []string{
			"102_ANALYTICAL_CORE_REVIEW_CLOSURE.md",
		},
	},
	{
		ID:          "AC-14",
		Name:        "Value plus HasValue is an intentional presence contract",
		Disposition: dispositionRetained,
		Evidence: []string{
			"102_ANALYTICAL_CORE_REVIEW_CLOSURE.md",
		},
	},
	{
		ID:          "AC-15",
		Name:        "Nullable result sections are status-discriminated evidence",
		Disposition: dispositionRetained,
		Evidence: []string{
			"102_ANALYTICAL_CORE_REVIEW_CLOSURE.md",
		},
	},
	{
		ID:          "AC-16",
		Name:        "And and With naming bans are non-blocking style preferences",
		Disposition: dispositionRejected,
		Evidence: []string{
			"102_ANALYTICAL_CORE_REVIEW_CLOSURE.md",
		},
	},
	{
		ID:          "AC-17",
		Name:        "Metric identifiers must use one canonical namespace",
		Disposition: dispositionFixed,
		Evidence: []string{
			"100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md",
		},
	},
	{
		ID:          "AC-18",
		Name:        "Public failures must not expose raw operation errors",
		Disposition: dispositionFixed,
		Evidence: []string{
			"99_PROVENANCE_AND_ANALYTICAL_TRUST_HARDENING.md",
		},
	},
	{
		ID:          "AC-19",
		Name:        "Source ordering must not use a manual quadratic sort",
		Disposition: dispositionFixed,
		Evidence: []string{
			"102_ANALYTICAL_CORE_REVIEW_CLOSURE.md",
		},
	},
}

func main() {
	os.Exit(
		run(
			os.Args[1:],
			os.Stdout,
			os.Stderr,
		),
	)
}

func run(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) int {
	flags := flag.NewFlagSet(
		"analyticalcorefinalaudit",
		flag.ContinueOnError,
	)
	flags.SetOutput(stderr)

	rootValue := flags.String(
		"root",
		"",
		"repository root; auto-detected when omitted",
	)
	strict := flags.Bool(
		"strict",
		true,
		"return a non-zero exit code when an Analytical Core closure invariant fails",
	)

	if err := flags.Parse(args); err != nil {
		return 1
	}

	root, err := resolveRepositoryRoot(
		*rootValue,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"locate repository root: %v\n",
			err,
		)
		return 1
	}

	failures := auditRepository(
		root,
		stdout,
	)
	if len(failures) == 0 {
		fmt.Fprintln(
			stdout,
			"Analytical Core final source audit: PASS",
		)
		return 0
	}

	fmt.Fprintln(
		stderr,
		"Analytical Core final source audit: FAIL",
	)
	for _, failure := range failures {
		fmt.Fprintf(
			stderr,
			"- %s: %s\n",
			failure.Check,
			failure.Detail,
		)
	}

	if *strict {
		return 1
	}
	return 0
}

func resolveRepositoryRoot(
	explicit string,
) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		root, err := filepath.Abs(
			strings.TrimSpace(explicit),
		)
		if err != nil {
			return "", err
		}
		if err := validateRepositoryRoot(
			root,
		); err != nil {
			return "", err
		}
		return root, nil
	}

	current, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if validateRepositoryRoot(current) == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New(
				"repository root containing apps/api/go.mod and apps/web/package.json was not found",
			)
		}
		current = parent
	}
}

func validateRepositoryRoot(
	root string,
) error {
	for _, relativePath := range []string{
		"apps/api/go.mod",
		"apps/web/package.json",
	} {
		path := filepath.Join(
			root,
			filepath.FromSlash(relativePath),
		)
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf(
				"required repository file %s: %w",
				path,
				err,
			)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf(
				"required repository path is not a file: %s",
				path,
			)
		}
	}

	return nil
}

func auditRepository(
	root string,
	output io.Writer,
) []auditFailure {
	groups := []struct {
		name  string
		check func(string) []auditFailure
	}{
		{
			name:  "Analytical finding register",
			check: auditFindingRegister,
		},
		{
			name:  "Analytical documentation surface",
			check: auditDocumentationSurface,
		},
		{
			name:  "Analytical correctness contracts",
			check: auditCorrectnessContracts,
		},
		{
			name:  "Analytical contributor ordering",
			check: auditContributorOrdering,
		},
		{
			name:  "Analytical presence and precision",
			check: auditPresencePrecisionContracts,
		},
		{
			name:  "Analytical compatibility boundary",
			check: auditCompatibilityBoundary,
		},
		{
			name:  "Analytical frontend contracts",
			check: auditFrontendContracts,
		},
		{
			name:  "Analytical standard source sorting",
			check: auditStandardSourceSorting,
		},
		{
			name:  "Analytical Continuous Integration coverage",
			check: auditContinuousIntegration,
		},
	}

	failures := make(
		[]auditFailure,
		0,
	)

	for _, group := range groups {
		groupFailures := group.check(root)
		if len(groupFailures) == 0 {
			fmt.Fprintf(
				output,
				"%s: PASS\n",
				group.name,
			)
			continue
		}

		failures = append(
			failures,
			groupFailures...,
		)
	}

	sort.Slice(
		failures,
		func(left int, right int) bool {
			if failures[left].Check ==
				failures[right].Check {
				return failures[left].Detail <
					failures[right].Detail
			}
			return failures[left].Check <
				failures[right].Check
		},
	)

	return failures
}

func auditFindingRegister(
	root string,
) []auditFailure {
	failures := make(
		[]auditFailure,
		0,
	)

	if len(analyticalCoreFindings) != 19 {
		failures = append(
			failures,
			auditFailure{
				Check: "Analytical finding register count",
				Detail: fmt.Sprintf(
					"finding count is %d; expected 19",
					len(analyticalCoreFindings),
				),
			},
		)
	}

	documentPath := filepath.Join(
		root,
		"docs",
		"102_ANALYTICAL_CORE_REVIEW_CLOSURE.md",
	)
	document, err := os.ReadFile(documentPath)
	if err != nil {
		return append(
			failures,
			auditFailure{
				Check: "Analytical finding register document",
				Detail: fmt.Sprintf(
					"read %s: %v",
					documentPath,
					err,
				),
			},
		)
	}
	documentText := string(document)

	seen := make(
		map[string]struct{},
		len(analyticalCoreFindings),
	)
	counts := map[findingDisposition]int{}

	for index, item := range analyticalCoreFindings {
		expectedID := fmt.Sprintf(
			"AC-%02d",
			index+1,
		)
		if item.ID != expectedID {
			failures = append(
				failures,
				auditFailure{
					Check: "Analytical finding register sequence",
					Detail: fmt.Sprintf(
						"finding index %d has id %s; expected %s",
						index,
						item.ID,
						expectedID,
					),
				},
			)
		}

		if _, exists := seen[item.ID]; exists {
			failures = append(
				failures,
				auditFailure{
					Check: "Analytical finding register uniqueness",
					Detail: fmt.Sprintf(
						"duplicate finding id %s",
						item.ID,
					),
				},
			)
		}
		seen[item.ID] = struct{}{}
		counts[item.Disposition]++

		line := findingTableLine(
			documentText,
			item.ID,
		)
		if line == "" {
			failures = append(
				failures,
				auditFailure{
					Check: "Analytical finding register coverage",
					Detail: fmt.Sprintf(
						"%s is missing from Document 102",
						item.ID,
					),
				},
			)
		} else if !strings.Contains(
			line,
			string(item.Disposition),
		) {
			failures = append(
				failures,
				auditFailure{
					Check: "Analytical finding disposition",
					Detail: fmt.Sprintf(
						"%s does not contain disposition %s",
						item.ID,
						item.Disposition,
					),
				},
			)
		}

		for _, evidence := range item.Evidence {
			path := filepath.Join(
				root,
				"docs",
				evidence,
			)
			info, statErr := os.Stat(path)
			if statErr != nil ||
				!info.Mode().IsRegular() {
				failures = append(
					failures,
					auditFailure{
						Check: "Analytical finding evidence",
						Detail: fmt.Sprintf(
							"%s evidence file is unavailable: docs/%s",
							item.ID,
							evidence,
						),
					},
				)
			}
		}
	}

	expectedCounts := map[findingDisposition]int{
		dispositionFixed:    14,
		dispositionRetained: 3,
		dispositionRejected: 2,
	}
	for disposition, expected := range expectedCounts {
		if counts[disposition] != expected {
			failures = append(
				failures,
				auditFailure{
					Check: "Analytical finding disposition totals",
					Detail: fmt.Sprintf(
						"%s count is %d; expected %d",
						disposition,
						counts[disposition],
						expected,
					),
				},
			)
		}
	}

	return failures
}

func findingTableLine(
	document string,
	id string,
) string {
	for _, line := range strings.Split(
		document,
		"\n",
	) {
		if strings.Contains(
			line,
			"| "+id+" |",
		) {
			return line
		}
	}
	return ""
}

func auditDocumentationSurface(
	root string,
) []auditFailure {
	failures := make(
		[]auditFailure,
		0,
	)

	for _, document := range []string{
		"97_ANALYTICAL_CONTRIBUTOR_SEMANTICS_HARDENING.md",
		"98_AIRPORT_AND_GEOGRAPHIC_METRIC_INTEGRITY.md",
		"99_PROVENANCE_AND_ANALYTICAL_TRUST_HARDENING.md",
		"100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md",
		"101_SERVER_OWNED_QUALITY_METRICS.md",
		"102_ANALYTICAL_CORE_REVIEW_CLOSURE.md",
	} {
		path := filepath.Join(
			root,
			"docs",
			document,
		)
		info, err := os.Stat(path)
		if err != nil ||
			!info.Mode().IsRegular() {
			failures = append(
				failures,
				auditFailure{
					Check: "Analytical document register",
					Detail: fmt.Sprintf(
						"missing docs/%s",
						document,
					),
				},
			)
		}
	}

	failures = append(
		failures,
		auditRules(
			root,
			[]fileRule{
				{
					Name: "Analytical closure document index",
					Path: "docs/DOCUMENT_INDEX.md",
					Required: []string{
						"<!-- ANALYTICAL-CORE-REVIEW-CLOSURE:DOCUMENT-INDEX -->",
						"102_ANALYTICAL_CORE_REVIEW_CLOSURE.md",
					},
				},
				{
					Name: "Analytical closure repository status",
					Path: "README.md",
					Required: []string{
						"<!-- ANALYTICAL-CORE-REVIEW-CLOSURE:README -->",
						"ANALYTICAL_CORE_REVIEW_STATUS=CLOSED",
						"Open required changes: 0",
					},
				},
				{
					Name: "Analytical closure implementation status",
					Path: "docs/25_IMPLEMENTATION_SEQUENCE.md",
					Required: []string{
						"<!-- ANALYTICAL-CORE-REVIEW-CLOSURE:IMPLEMENTATION -->",
						"Analytical Core Review Closure",
						"Status: COMPLETED.",
					},
				},
				{
					Name: "Query architecture post-closure reference",
					Path: "docs/100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md",
					Required: []string{
						"<!-- ANALYTICAL-CORE-REVIEW-CLOSURE:QUERY-ARCHITECTURE -->",
						"Document 102",
						"ANALYTICAL_CORE_REVIEW_STATUS=CLOSED",
					},
				},
				{
					Name: "Server-owned metrics post-closure reference",
					Path: "docs/101_SERVER_OWNED_QUALITY_METRICS.md",
					Required: []string{
						"<!-- ANALYTICAL-CORE-REVIEW-CLOSURE:SERVER-OWNED-QUALITY -->",
						"Document 102",
						"ANALYTICAL_CORE_REVIEW_STATUS=CLOSED",
					},
				},
				{
					Name: "Analytical closure totals",
					Path: "docs/102_ANALYTICAL_CORE_REVIEW_CLOSURE.md",
					Required: []string{
						"Original findings: 19",
						"Fixed: 14",
						"Deliberately retained: 3",
						"Rejected or non-blocking: 2",
						"Deferred: 0",
						"Unclassified: 0",
						"ANALYTICAL_CORE_REVIEW_STATUS=CLOSED",
						"Release decision: ACCEPTABLE",
					},
				},
			},
		)...,
	)

	return failures
}

func auditCorrectnessContracts(
	root string,
) []auditFailure {
	return auditRules(
		root,
		[]fileRule{
			{
				Name: "Eligible contributor preparation",
				Path: "apps/api/internal/analytics/metricexecution/active_aircraft.go",
				Required: []string{
					"prepareUniqueAircraftContributors(",
					"additional eligible trajectories",
				},
			},
			{
				Name: "Traffic Density contributor preparation",
				Path: "apps/api/internal/analytics/metricexecution/traffic_density.go",
				Required: []string{
					"prepareUniqueAircraftContributors(",
					"len(allowed)",
					"request.AreaSquareKilometers",
				},
			},
			{
				Name: "Future observation exclusion",
				Path: "apps/api/internal/analytics/trajectoryeligibility/evaluator.go",
				Required: []string{
					"policy.MaximumFutureObservationSkew",
					"ReasonFutureObservation",
				},
			},
			{
				Name: "Finite Traffic Density arithmetic",
				Path: "apps/api/internal/analytics/metrics/traffic_density.go",
				Required: []string{
					"math.IsNaN(data.AreaSquareKilometers)",
					"math.IsInf(data.AreaSquareKilometers, 0)",
					"area must be finite and greater than zero",
				},
			},
			{
				Name: "Airport and region ownership",
				Path: "apps/api/internal/http/handlers/analytical_metrics.go",
				Required: []string{
					`ctx.Query("airport_icao")`,
					"RecentWithinBounds(",
					"trafficDensityAreaSquareKilometers(",
					"area_square_kilometers is derived from region",
				},
			},
			{
				Name: "Server-owned quality snapshot",
				Path: "apps/api/internal/http/handlers/analytical_production_snapshot.go",
				Required: []string{
					"productionQualityResultLimit",
					"metricquery.MaximumResultLimit",
					"serverTrajectoryQuerySource",
					"EvaluateSamplingDensity(",
				},
				RequiredCompact: []string{
					"dataqualityintegration.DefaultExpectedObservationInterval",
				},
			},
			{
				Name: "Server-owned freshness threshold",
				Path: "apps/api/internal/http/handlers/analytical_metrics.go",
				RequiredCompact: []string{
					"dataqualityintegration.DefaultStaleAfter",
				},
			},
			{
				Name: "Reference time and UUID normalization",
				Path: "apps/api/internal/analytics/metricquery/model.go",
				Required: []string{
					"ErrReferenceTimeRequired",
					"if now.IsZero()",
					"canonical := parsed.String()",
					"seen[canonical]",
				},
			},
			{
				Name: "Strict analytical provenance",
				Path: "apps/api/internal/analytics/analyticalresult/validation.go",
				Required: []string{
					"ErrSourceNamePlaceholder",
					"ErrSourceRetrievedAtMissing",
					"ErrSourceRetrievedAfterCalculation",
					`strings.EqualFold(`,
					`"unknown"`,
				},
			},
			{
				Name: "Sanitized analytical failure",
				Path: "apps/api/internal/analytics/metricexecution/execution.go",
				Required: []string{
					`failure.Message =`,
					`"Analytical metric calculation failed."`,
				},
				Forbidden: []string{
					"failure.Message = operationErr.Error()",
				},
			},
			{
				Name: "Canonical analytical Metric IDs",
				Path: "apps/api/internal/analytics/metricexecution/model.go",
				Required: []string{
					"MetricIDActiveAircraft  = metrics.ActiveAircraftMetricID",
					"MetricIDTrafficDensity  = metrics.TrafficDensityMetricID",
					"MetricIDAirportActivity = metrics.AirportActivityMetricID",
					"MetricIDCoverageScore   = metrics.CoverageScoreMetricID",
					"MetricIDDataFreshness   = metrics.DataFreshnessMetricID",
				},
			},
		},
	)
}

func auditContributorOrdering(
	root string,
) []auditFailure {
	path := filepath.Join(
		root,
		filepath.FromSlash(
			"apps/api/internal/analytics/metricexecution/execution.go",
		),
	)
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return []auditFailure{
			{
				Check: "Analytical contributor ordering",
				Detail: fmt.Sprintf(
					"read %s: %v",
					path,
					err,
				),
			},
		}
	}

	content := string(contentBytes)
	filterIndex := strings.Index(
		content,
		"FilterTrajectories(",
	)
	preparationIndex := strings.Index(
		content,
		"if preparation != nil",
	)
	allowedPreparationIndex := strings.Index(
		content,
		"preparation(filtered.Allowed)",
	)

	failures := make(
		[]auditFailure,
		0,
	)
	if filterIndex < 0 {
		failures = append(
			failures,
			auditFailure{
				Check:  "Analytical contributor ordering",
				Detail: "FilterTrajectories call is missing",
			},
		)
	}
	if preparationIndex < 0 ||
		allowedPreparationIndex < 0 {
		failures = append(
			failures,
			auditFailure{
				Check:  "Analytical contributor ordering",
				Detail: "eligible contributor preparation is missing",
			},
		)
	}
	if filterIndex >= 0 &&
		preparationIndex >= 0 &&
		filterIndex > preparationIndex {
		failures = append(
			failures,
			auditFailure{
				Check:  "Analytical contributor ordering",
				Detail: "preparation appears before capability filtering",
			},
		)
	}
	if preparationIndex >= 0 &&
		allowedPreparationIndex >= 0 &&
		preparationIndex > allowedPreparationIndex {
		failures = append(
			failures,
			auditFailure{
				Check:  "Analytical contributor ordering",
				Detail: "allowed contributor preparation is outside its guarded block",
			},
		)
	}

	return failures
}

func auditPresencePrecisionContracts(
	root string,
) []auditFailure {
	return auditRules(
		root,
		[]fileRule{
			{
				Name: "Analytical result presence model",
				Path: "apps/api/internal/analytics/analyticalresult/model.go",
				Required: []string{
					"Value        T",
					"HasValue     bool",
					"func (result Result[T]) ValueOrZero() T",
				},
			},
			{
				Name: "Analytical status presence invariants",
				Path: "apps/api/internal/analytics/analyticalresult/validation.go",
				Required: []string{
					"case StatusComplete:",
					"if !result.HasValue",
					"case StatusDenied:",
					"if result.HasValue",
					"case StatusFailed:",
				},
			},
			{
				Name: "Public value presence and precision",
				Path: "apps/api/internal/http/handlers/analytical_metrics_response.go",
				Required: []string{
					"HasValue:     result.HasValue",
					"if result.HasValue",
					"response.Value = result.Value",
				},
				Forbidden: []string{
					"math.Round(",
					"strconv.FormatFloat(",
					"fmt.Sprintf(",
					"ValueOrZero(",
				},
			},
			{
				Name: "Frontend result presence model",
				Path: "apps/web/types/analytics.ts",
				Required: []string{
					"value?: TValue",
					"has_value: boolean",
					"eligibility?: AnalyticalEligibility",
					"failure?: AnalyticalFailure",
					"confidence_report?: AnalyticalConfidenceReport",
				},
			},
			{
				Name: "Frontend display-only precision",
				Path: "apps/web/components/analytics/analytics-overview.tsx",
				Required: []string{
					"function formatRatio(value: number): string",
					"maximumFractionDigits: 1",
					"function formatDensity(value: number): string",
					"maximumSignificantDigits: 4",
				},
			},
		},
	)
}

func auditCompatibilityBoundary(
	root string,
) []auditFailure {
	failures := auditRules(
		root,
		[]fileRule{
			{
				Name: "Executor compatibility boundary",
				Path: "apps/api/internal/analytics/executor/executor.go",
				Required: []string{
					"type Executor struct",
					"scopeGuard",
					"confidenceEvaluator",
					"EvaluateConfidence(",
					"FilterTrajectories(",
				},
				Forbidden: []string{
					"calculator *calculator.Calculator",
					"func (executor *Executor) Calculator(",
					"func (executor *Executor) ScopeGuard(",
					"func (executor *Executor) ConfidenceEvaluator(",
				},
			},
			{
				Name: "Metric service dependency inversion",
				Path: "apps/api/internal/analytics/metricexecution/service.go",
				Required: []string{
					"type analyticalExecutor interface",
					"FilterTrajectories(",
					"EvaluateConfidence(",
					"executor analyticalExecutor",
				},
				Forbidden: []string{
					"*executor.Executor",
					"func (service *Service) Executor(",
				},
			},
		},
	)

	for _, relativeRoot := range []string{
		"apps/api/cmd/server",
		"apps/api/internal/server",
	} {
		rootPath := filepath.Join(
			root,
			filepath.FromSlash(relativeRoot),
		)
		walkErr := filepath.Walk(
			rootPath,
			func(
				path string,
				info os.FileInfo,
				err error,
			) error {
				if err != nil {
					return err
				}
				if info.IsDir() ||
					!strings.HasSuffix(
						info.Name(),
						".go",
					) ||
					strings.HasSuffix(
						info.Name(),
						"_test.go",
					) {
					return nil
				}

				contentBytes, readErr :=
					os.ReadFile(path)
				if readErr != nil {
					return readErr
				}
				content := string(contentBytes)

				for _, forbiddenImport := range []string{
					"/internal/analytics/calculator",
					"/internal/analytics/registry",
				} {
					if strings.Contains(
						content,
						forbiddenImport,
					) {
						failures = append(
							failures,
							auditFailure{
								Check: "Analytical runtime compatibility boundary",
								Detail: fmt.Sprintf(
									"%s imports compatibility package %s",
									path,
									forbiddenImport,
								),
							},
						)
					}
				}

				return nil
			},
		)
		if walkErr != nil {
			failures = append(
				failures,
				auditFailure{
					Check: "Analytical runtime compatibility boundary",
					Detail: fmt.Sprintf(
						"walk %s: %v",
						rootPath,
						walkErr,
					),
				},
			)
		}
	}

	return failures
}

func auditFrontendContracts(
	root string,
) []auditFailure {
	return auditRules(
		root,
		[]fileRule{
			{
				Name: "Analytical frontend API ownership",
				Path: "apps/web/lib/api/analytics.ts",
				Required: []string{
					"'traffic.traffic_density'",
					"'traffic.coverage_score'",
					"'traffic.data_freshness'",
					"buildProductionQualitySearchParameters(",
				},
				RequiredCompact: []string{
					"searchParameters.set('airport_icao'",
					"searchParameters.set('radius_kilometers'",
				},
				Forbidden: []string{
					"observed_samples",
					"expected_samples",
					"max_age_seconds",
					"area_square_kilometers",
					"arrival_trajectory_ids",
					"departure_trajectory_ids",
				},
			},
			{
				Name: "Analytical React Query ownership",
				Path: "apps/web/lib/queries/analytics.ts",
				Required: []string{
					"parameters.airportICAO",
					"parameters.radiusKilometers",
					"parameters.windowMinutes",
					"normalizeRegionCode(parameters.regionCode)",
				},
				Forbidden: []string{
					"parameters === null",
					"enabled: parameters !== null",
					"observedSamples",
					"expectedSamples",
					"maximumAgeSeconds",
					"areaSquareKilometers",
					"arrivalTrajectoryIDs",
					"departureTrajectoryIDs",
				},
			},
			{
				Name: "Analytical overview server scope",
				Path: "apps/web/components/analytics/analytics-overview.tsx",
				Required: []string{
					"const productionQualityParameters",
					"useAnalyticalCoverageScore(",
					"useAnalyticalDataFreshness(",
					"Observation Coverage",
				},
				Forbidden: []string{
					"buildCoverageParameters",
					"buildFreshnessParameters",
					"observedSamples",
					"maximumAgeSeconds",
				},
			},
		},
	)
}

func auditStandardSourceSorting(
	root string,
) []auditFailure {
	return auditRules(
		root,
		[]fileRule{
			{
				Name: "Analytical source standard sorting",
				Path: "apps/api/internal/http/handlers/analytical_metrics.go",
				Required: []string{
					`"sort"`,
					"sort.Strings(names)",
				},
				Forbidden: []string{
					"func sortStrings(",
					"for left := 0; left < len(values); left++",
					"for right := left + 1; right < len(values); right++",
				},
			},
		},
	)
}

func auditContinuousIntegration(
	root string,
) []auditFailure {
	return auditRules(
		root,
		[]fileRule{
			{
				Name: "Backend Analytical Core closure gate",
				Path: ".github/workflows/backend-ci.yml",
				Required: []string{
					"apps/web/components/analytics/**",
					"apps/web/lib/api/analytics.ts",
					"apps/web/lib/queries/analytics.ts",
					"Run Analytical Core closure audit",
					"go run ./tools/analyticalcorefinalaudit -strict",
					"go run ./tools/projectaudit -mode all -strict",
					"go run ./tools/codereviewaudit -strict",
					"go run ./tools/stage14finalaudit -strict",
					"Run analytical core race tests",
					"./internal/analytics/...",
				},
			},
			{
				Name: "Frontend Analytical Core quality gate",
				Path: ".github/workflows/frontend-ci.yml",
				Required: []string{
					"- 'apps/web/**'",
					"Run ESLint",
					"Run TypeScript validation",
					"Build production frontend",
				},
			},
		},
	)
}

func auditRules(
	root string,
	rules []fileRule,
) []auditFailure {
	failures := make(
		[]auditFailure,
		0,
	)

	for _, rule := range rules {
		path := filepath.Join(
			root,
			filepath.FromSlash(rule.Path),
		)
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			failures = append(
				failures,
				auditFailure{
					Check: rule.Name,
					Detail: fmt.Sprintf(
						"read %s: %v",
						rule.Path,
						err,
					),
				},
			)
			continue
		}

		content := string(contentBytes)
		compactContent := compactSource(
			content,
		)

		for _, fragment := range rule.Required {
			if !strings.Contains(
				content,
				fragment,
			) {
				failures = append(
					failures,
					auditFailure{
						Check: rule.Name,
						Detail: fmt.Sprintf(
							"%s is missing required fragment %q",
							rule.Path,
							fragment,
						),
					},
				)
			}
		}
		for _, fragment := range rule.RequiredCompact {
			if !strings.Contains(
				compactContent,
				compactSource(fragment),
			) {
				failures = append(
					failures,
					auditFailure{
						Check: rule.Name,
						Detail: fmt.Sprintf(
							"%s is missing required compact fragment %q",
							rule.Path,
							fragment,
						),
					},
				)
			}
		}

		for _, fragment := range rule.Forbidden {
			if strings.Contains(
				content,
				fragment,
			) {
				failures = append(
					failures,
					auditFailure{
						Check: rule.Name,
						Detail: fmt.Sprintf(
							"%s contains forbidden fragment %q",
							rule.Path,
							fragment,
						),
					},
				)
			}
		}
	}

	return failures
}

func compactSource(
	value string,
) string {
	return strings.Map(
		func(character rune) rune {
			if unicode.IsSpace(character) {
				return -1
			}
			return character
		},
		value,
	)
}
