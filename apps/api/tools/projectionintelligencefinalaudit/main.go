package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type moduleReview struct {
	name         string
	workflowName string
	command      string
	document     string
	auditSource  string
}

type fileRequirement struct {
	path      string
	required  []string
	forbidden []string
}

func main() {
	strict := flag.Bool(
		"strict",
		false,
		"fail when a Projection Intelligence reconciliation requirement is absent",
	)
	root := flag.String(
		"root",
		"",
		"optional path to the apps/api module root",
	)
	flag.Parse()

	apiRoot, err := resolveAPIRoot(*root)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Projection Intelligence final reconciliation audit: %v\n",
			err,
		)
		if *strict {
			os.Exit(1)
		}
		return
	}

	reviews := projectionReviews()
	failures := inspectWorkflow(apiRoot, reviews)
	failures = append(failures, inspectModuleAuditDelegation(apiRoot, reviews)...)
	failures = append(failures, inspectModuleDocuments(apiRoot, reviews)...)
	failures = append(failures, inspectFiles(apiRoot, crossModuleRequirements())...)

	if len(failures) == 0 {
		fmt.Println("Projection Intelligence final reconciliation audit: PASS")
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Projection Intelligence final reconciliation audit: %s\n",
			failure,
		)
	}
	if *strict {
		os.Exit(1)
	}
}

func projectionReviews() []moduleReview {
	return []moduleReview{
		{
			name:         "Projection Contract",
			workflowName: "Run projection contract review audit",
			command:      "go run ./tools/projectioncontractreviewaudit -strict",
			document:     "../../docs/134_PROJECTION_CONTRACT_REVIEW_HARDENING.md",
			auditSource:  "tools/projectioncontractreviewaudit/main.go",
		},
		{
			name:         "Projection Horizon",
			workflowName: "Run projection horizon review audit",
			command:      "go run ./tools/projectionhorizonreviewaudit -strict",
			document:     "../../docs/135_PROJECTION_HORIZON_REVIEW_HARDENING.md",
			auditSource:  "tools/projectionhorizonreviewaudit/main.go",
		},
		{
			name:         "Projection Baseline",
			workflowName: "Run projection baseline review audit",
			command:      "go run ./tools/projectionbaselinereviewaudit -strict",
			document:     "../../docs/136_PROJECTION_BASELINE_REVIEW_HARDENING.md",
			auditSource:  "tools/projectionbaselinereviewaudit/main.go",
		},
		{
			name:         "Projection Neighbors",
			workflowName: "Run projection neighbors review audit",
			command:      "go run ./tools/projectionneighborsreviewaudit -strict",
			document:     "../../docs/137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md",
			auditSource:  "tools/projectionneighborsreviewaudit/main.go",
		},
		{
			name:         "Projection Pattern Confidence",
			workflowName: "Run projection pattern confidence review audit",
			command:      "go run ./tools/projectionpatternconfidencereviewaudit -strict",
			document:     "../../docs/138_PROJECTION_PATTERN_CONFIDENCE_REVIEW_HARDENING.md",
			auditSource:  "tools/projectionpatternconfidencereviewaudit/main.go",
		},
		{
			name:         "Projection Freshness",
			workflowName: "Run projection freshness review audit",
			command:      "go run ./tools/projectionfreshnessreviewaudit -strict",
			document:     "../../docs/139_PROJECTION_FRESHNESS_REVIEW_HARDENING.md",
			auditSource:  "tools/projectionfreshnessreviewaudit/main.go",
		},
		{
			name:         "Projection Route Frequency",
			workflowName: "Run projection route frequency review audit",
			command:      "go run ./tools/projectionroutefrequencyreviewaudit -strict",
			document:     "../../docs/140_PROJECTION_ROUTE_FREQUENCY_REVIEW_HARDENING.md",
			auditSource:  "tools/projectionroutefrequencyreviewaudit/main.go",
		},
		{
			name:         "Projection Continuation",
			workflowName: "Run projection continuation review audit",
			command:      "go run ./tools/projectioncontinuationreviewaudit -strict",
			document:     "../../docs/141_PROJECTION_CONTINUATION_REVIEW_HARDENING.md",
			auditSource:  "tools/projectioncontinuationreviewaudit/main.go",
		},
		{
			name:         "Projection Arrival",
			workflowName: "Run projection arrival review audit",
			command:      "go run ./tools/projectionarrivalreviewaudit -strict",
			document:     "../../docs/142_PROJECTION_ARRIVAL_REVIEW_HARDENING.md",
			auditSource:  "tools/projectionarrivalreviewaudit/main.go",
		},
		{
			name:         "Projection Evaluation",
			workflowName: "Run projection evaluation review audit",
			command:      "go run ./tools/projectionevaluationreviewaudit -strict",
			document:     "../../docs/143_PROJECTION_EVALUATION_REVIEW_HARDENING.md",
			auditSource:  "tools/projectionevaluationreviewaudit/main.go",
		},
		{
			name:         "Projection Production",
			workflowName: "Run projection production review audit",
			command:      "go run ./tools/projectionproductionreviewaudit -strict",
			document:     "../../docs/144_PROJECTION_PRODUCTION_REVIEW_HARDENING.md",
			auditSource:  "tools/projectionproductionreviewaudit/main.go",
		},
		{
			name:         "Projection Read",
			workflowName: "Run projection read review audit",
			command:      "go run ./tools/projectionreadreviewaudit -strict",
			document:     "../../docs/145_PROJECTION_READ_REVIEW_HARDENING.md",
			auditSource:  "tools/projectionreadreviewaudit/main.go",
		},
	}
}

func crossModuleRequirements() []fileRequirement {
	return []fileRequirement{
		{
			path: "internal/projectionintelligence/projectionread/postgres_snapshot.go",
			required: []string{
				"IsoLevel:   pgx.RepeatableRead",
				"AccessMode: pgx.ReadOnly",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/composer.go",
			required: []string{
				"authorized := AuthorizedHistoricalEvidence{",
				"approvedEvidence := projectioncontinuation.ApprovedEvidence{",
				"ProjectApprovedWithPlan(",
				"EstimateArrival(projectionarrival.Request{",
			},
		},
		{
			path: "internal/projectionintelligence/projectionarrival/computation.go",
			required: []string{
				"radialClosingSpeedMPS :=",
				"maximumArrivalTime :=",
				"latestTime.After(maximumArrivalTime)",
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/truth.go",
			required: []string{
				"ErrTruthAvailabilityEvidenceMissing",
				"evidence.AvailableAt.After(evaluatedAt)",
				"truthMatchImplausibleMovement",
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/fingerprint.go",
			required: []string{
				"func projectionSnapshotFingerprint(",
				"func truthSnapshotFingerprint(",
				"func aggregateFingerprint(results []Result)",
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			required: []string{
				"## 40. Projection Intelligence Final Cross-Module Reconciliation",
				"PROJECTION_INTELLIGENCE_RECONCILIATION_BASELINE_COMMIT=a917741a1c3e7e6621ec2767bd9484ae8ffa21a8",
				"PROJECTION_INTELLIGENCE_MODULE_REVIEW_AUDITS=12",
				"PROJECTION_INTELLIGENCE_MODULE_FORMAL_CLOSURES=12",
				"PROJECTION_INTELLIGENCE_CROSS_MODULE_AUDIT=IMPLEMENTED_PENDING_EXACT_CI",
				"PROJECTION_INTELLIGENCE_OPEN_CONFIRMED_CROSS_MODULE_FINDINGS=0",
				"PROJECTION_INTELLIGENCE_FINAL_RECONCILIATION=OPEN_PENDING_EXACT_CI",
			},
			forbidden: []string{
				"PROJECTION_INTELLIGENCE_FINAL_RECONCILIATION=COMPLETE",
			},
		},
		{
			path: "../../docs/146_PROJECTION_INTELLIGENCE_FINAL_RECONCILIATION.md",
			required: []string{
				"# Projection Intelligence Final Cross-Module Reconciliation",
				"Status: engineering implementation complete; exact Continuous Integration and formal closure pending",
				"RECONCILIATION_BASELINE_COMMIT=a917741a1c3e7e6621ec2767bd9484ae8ffa21a8",
				"RECONCILIATION_BASELINE_GITHUB_ACTIONS_RUN=30653437694",
				"EXTERNAL_REVIEW_DECLARED_COMMIT=a1689dc",
				"EXTERNAL_REVIEW_REPORTED_P0_FINDINGS=7",
				"EXTERNAL_REVIEW_OPEN_CONFIRMED_FINDINGS=0",
				"MODULE_REVIEW_AUDITS=12",
				"MODULE_FORMAL_CLOSURES=12",
				"CROSS_MODULE_AUDIT=IMPLEMENTED_PENDING_EXACT_CI",
				"OPEN_CONFIRMED_CROSS_MODULE_FINDINGS=0",
				"FINAL_RECONCILIATION=OPEN_PENDING_EXACT_CI",
			},
			forbidden: []string{
				"FINAL_RECONCILIATION=COMPLETE",
				"REVIEW_STATUS=CLOSED",
			},
		},
		{
			path: "../../docs/DOCUMENT_INDEX.md",
			required: []string{
				"## Document 146 — Projection Intelligence Final Cross-Module Reconciliation",
				"twelve formally closed Projection Intelligence module reviews",
				"external static review declared at `a1689dc`",
				"exact Continuous Integration pending",
			},
		},
	}
}

func inspectWorkflow(
	apiRoot string,
	reviews []moduleReview,
) []string {
	path := filepath.Clean(filepath.Join(
		apiRoot,
		"../../.github/workflows/backend-ci.yml",
	))
	content, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("read Backend Continuous Integration workflow: %v", err)}
	}
	workflow := string(content)
	failures := make([]string, 0)
	previousIndex := -1
	for _, review := range reviews {
		for _, fragment := range []string{review.workflowName, review.command} {
			if count := strings.Count(workflow, fragment); count != 1 {
				failures = append(
					failures,
					fmt.Sprintf("%s workflow fragment %q count is %d, expected 1", review.name, fragment, count),
				)
			}
		}
		index := strings.Index(workflow, review.command)
		if index >= 0 && index <= previousIndex {
			failures = append(
				failures,
				fmt.Sprintf("%s audit is not ordered after the preceding Projection audit", review.name),
			)
		}
		if index >= 0 {
			previousIndex = index
		}
	}

	finalName := "Run Projection Intelligence final reconciliation audit"
	finalCommand := "go run ./tools/projectionintelligencefinalaudit -strict"
	for _, fragment := range []string{finalName, finalCommand} {
		if count := strings.Count(workflow, fragment); count != 1 {
			failures = append(
				failures,
				fmt.Sprintf("final reconciliation workflow fragment %q count is %d, expected 1", fragment, count),
			)
		}
	}
	finalIndex := strings.Index(workflow, finalCommand)
	if finalIndex >= 0 && finalIndex <= previousIndex {
		failures = append(
			failures,
			"final reconciliation audit is not ordered after all module review audits",
		)
	}
	return failures
}

func inspectModuleAuditDelegation(
	apiRoot string,
	reviews []moduleReview,
) []string {
	failures := make([]string, 0)
	for _, review := range reviews {
		path := filepath.Clean(filepath.Join(apiRoot, review.auditSource))
		content, err := os.ReadFile(path)
		if err != nil {
			failures = append(
				failures,
				fmt.Sprintf("read %s permanent audit source: %v", review.name, err),
			)
			continue
		}
		text := string(content)
		documentName := filepath.Base(review.document)
		for _, fragment := range []string{documentName, "Status: closed"} {
			if strings.Contains(text, fragment) {
				continue
			}
			failures = append(
				failures,
				fmt.Sprintf(
					"%s permanent audit does not delegate closure enforcement to %q",
					review.name,
					fragment,
				),
			)
		}
	}
	return failures
}

func inspectModuleDocuments(
	apiRoot string,
	reviews []moduleReview,
) []string {
	failures := make([]string, 0)
	for _, review := range reviews {
		path := filepath.Clean(filepath.Join(apiRoot, review.document))
		content, err := os.ReadFile(path)
		if err != nil {
			failures = append(
				failures,
				fmt.Sprintf("read %s review document: %v", review.name, err),
			)
			continue
		}
		if !strings.Contains(string(content), "Status: closed") {
			failures = append(
				failures,
				fmt.Sprintf("%s review document is not explicitly closed", review.name),
			)
		}
	}
	return failures
}

func inspectFiles(
	apiRoot string,
	requirements []fileRequirement,
) []string {
	failures := make([]string, 0)
	for _, requirement := range requirements {
		path := filepath.Clean(filepath.Join(apiRoot, requirement.path))
		content, err := os.ReadFile(path)
		if err != nil {
			failures = append(
				failures,
				fmt.Sprintf("read %s: %v", requirement.path, err),
			)
			continue
		}
		text := string(content)
		for _, fragment := range requirement.required {
			if !strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s is missing %q", requirement.path, fragment),
				)
			}
		}
		for _, fragment := range requirement.forbidden {
			if strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s contains forbidden %q", requirement.path, fragment),
				)
			}
		}
	}
	return failures
}

func resolveAPIRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return validateAPIRoot(explicit)
	}

	current, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve current directory: %w", err)
	}
	for {
		if root, err := validateAPIRoot(current); err == nil {
			return root, nil
		}
		candidate := filepath.Join(current, "apps", "api")
		if root, err := validateAPIRoot(candidate); err == nil {
			return root, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", fmt.Errorf("apps/api module root was not found")
}

func validateAPIRoot(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve API root: %w", err)
	}
	info, err := os.Stat(filepath.Join(absolute, "go.mod"))
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("%s is not the apps/api module root", absolute)
	}
	return absolute, nil
}
