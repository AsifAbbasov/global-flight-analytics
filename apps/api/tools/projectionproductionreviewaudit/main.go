package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type requirement struct {
	path      string
	fragments []string
	forbidden []string
}

func main() {
	strict := flag.Bool("strict", false, "fail when a Projection Production review requirement is absent")
	root := flag.String("root", "", "optional path to the apps/api module root")
	flag.Parse()

	apiRoot, err := resolveAPIRoot(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Projection production review audit: %v\n", err)
		if *strict {
			os.Exit(1)
		}
		return
	}

	failures := inspectRequirements(apiRoot, reviewRequirements())
	failures = append(failures, inspectExactClosureMarkers(apiRoot)...)
	if len(failures) == 0 {
		fmt.Println("Projection production review audit: PASS")
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(os.Stderr, "Projection production review audit: %s\n", failure)
	}
	if *strict {
		os.Exit(1)
	}
}

func reviewRequirements() []requirement {
	return []requirement{
		{
			path: "internal/projectionintelligence/projectionbaseline/baseline.go",
			fragments: []string{
				"ProjectWithPlan(",
				"request does not match authorized plan",
				"if err := plan.Validate(); err != nil",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/continuation_core.go",
			fragments: []string{
				"ProjectApprovedWithPlan(",
				"projectWithPlan(",
				"request does not match authorized plan",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/approved_projection_lineage.go",
			fragments: []string{
				`ApprovedProjectionLineageVersion = "historical-approved-projection-lineage-v1"`,
				"type ApprovedProjectionLineage struct",
				"ValidateApprovedProjectionLineage(",
				"expectedFingerprint := continuationFingerprint(",
				"expectedInputs := continuationInputs(",
				"projection input fingerprint does not match approved evidence",
				"projection provenance inputs do not match approved neighbors",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/approved_projection_lineage_test.go",
			fragments: []string{
				"TestValidateApprovedProjectionLineageReconstructsFingerprintAndInputs",
				"TestValidateApprovedProjectionLineageRejectsForeignFingerprint",
				"TestValidateApprovedProjectionLineageRejectsForeignNeighborProvenance",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/continuation_fallback.go",
			fragments: []string{
				"FallbackProjector.ProjectWithPlan(",
				`"%w: %w"`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/config.go",
			fragments: []string{
				"ProjectWithPlan(",
				"ProjectApprovedWithPlan(",
				") (HistoricalProjectionOutcome, error)",
				"historicalProjectionAdapter()",
				"EstimateArrival(projectionarrival.Request)",
				"only an Estimated Arrival delta",
			},
			forbidden: []string{
				") (projectioncontract.Result, error)\n}\n\ntype NeighborSelector",
				"Estimate(\n\t\tprojectionarrival.Request",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/historical_adapter.go",
			fragments: []string{
				"type HistoricalProjectionSource interface",
				"ValidateApprovedProjectionLineage(",
				"type HistoricalProjectionOutcome struct",
				"type HistoricalProjectionAdapter struct",
				"func (*HistoricalProjectionAdapter) historicalProjectionAdapter()",
				"NewHistoricalProjectionAdapter(",
				"ValidateAgainst(",
				"historicalProvenanceNamesMatch(",
				`"historical_neighbor:"+trajectoryID`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/request_snapshot.go",
			fragments: []string{
				"func (request Request) Clone() Request",
				"request.CurrentTrajectory.Clone()",
				"cloneTrajectories(request.HistoricalCandidates)",
				"request.Route.Clone()",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/evidence.go",
			fragments: []string{
				"type AuthorizedHistoricalEvidence struct",
				"selection does not match production request and plan",
				"authorized pattern does not match selection",
				"authorized freshness does not match selection, pattern, or plan",
				"authorized route-frequency evidence does not match route history and plan",
				"route identity does not match current trajectory",
				"route evidence exceeds production time boundary",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/composer.go",
			fragments: []string{
				"snapshot := request.Clone()",
				"ProjectApprovedWithPlan(",
				"outcome.ValidateAgainst(plan, approvedEvidence)",
				`"historical_projection_lineage_mismatch"`,
				"KinematicProjector.ProjectWithPlan(",
				"historical_projection_authorized_with_limited_freshness",
				"historical_projection_authorized_with_limited_route_frequency",
				"requestInputFingerprint(",
				"Finalize(result)",
				`fmt.Errorf("%w: %w", sentinel, cause)`,
			},
			forbidden: []string{
				"HistoricalProjector.ProjectApproved(",
				"KinematicProjector.Project(",
				"ArrivalEstimator.Estimate(",
				`"%w: %v"`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/arrival_adapter.go",
			fragments: []string{
				"type ArrivalOutcome struct",
				"ErrArrivalProjectionMutation",
				"sameProjectionExceptArrival(before, result)",
				"return ArrivalOutcome{}, ErrArrivalProjectionMutation",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/projection_postconditions.go",
			fragments: []string{
				"projection identity does not match current trajectory",
				"projection time boundary does not match authorized plan",
				"historical projector postconditions failed",
				"ResultStatusUnavailable",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/model.go",
			fragments: []string{
				`Version                       = "projection-production-composition-v2"`,
				`RequestFingerprintVersion     = "projection-production-request-fingerprint-v2"`,
				`CompositionFingerprintVersion = "projection-production-composition-fingerprint-v2"`,
				"HorizonPlan projectionhorizon.Plan",
				"CompositionFingerprint string",
				"result.CompositionFingerprint != compositionFingerprint(result)",
				"result.Projection.Status == projectioncontract.ResultStatusUnavailable",
			},
			forbidden: []string{
				"projection-production-composition-v1",
				"projection-production-composition-fingerprint-v1",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/fingerprint.go",
			fragments: []string{
				"func requestInputFingerprint(",
				"Request:                     request.Clone()",
				"func compositionFingerprint(result Result)",
				"cloned.CompositionFingerprint = \"\"",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/integration_contract_test.go",
			fragments: []string{
				"KinematicProjector         = (*projectionbaseline.Baseline)(nil)",
				"HistoricalProjectionSource = (*projectioncontinuation.Baseline)(nil)",
				"HistoricalProjector        = (*HistoricalProjectionAdapter)(nil)",
				"ArrivalProjectionEstimator = (*projectionarrival.Estimator)(nil)",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/review_hardening_test.go",
			fragments: []string{
				"TestComposeUsesOneAuthorizedHorizonPlan",
				"TestComposeRejectsHistoricalProjectionPostconditionDrift",
				"TestComposeRejectsHistoricalProjectionLineageDrift",
				"TestComposeReturnsHistoricalProjectionLineageErrorWhenConfigured",
				"TestComposeRejectsUnavailableHistoricalProjection",
				"TestComposeDefensivelyClonesDependencyInputs",
				"TestComposePreservesUnderlyingDependencyError",
				"TestCompositionFingerprintBindsPublishedProjection",
				"TestArrivalAdapterRejectsPositionProjectionMutation",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/composition.go",
			fragments: []string{
				"projectionproduction.NewHistoricalProjectionAdapter(historical)",
				"HistoricalProjector:        historicalAdapter",
				"projectionproduction.NewArrivalAdapter(arrival)",
				"ArrivalEstimator:           arrivalAdapter",
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				"Run projection production review audit",
				"go run ./tools/projectionproductionreviewaudit -strict",
			},
		},
		{
			path: "../../docs/144_PROJECTION_PRODUCTION_REVIEW_HARDENING.md",
			fragments: []string{
				"Status: reopened — Historical Projector output-lineage correction implemented, exact Continuous Integration and formal reclosure pending",
				"FORMAL_CLOSURE_REOPENED_BASELINE=0f1a31f56f4baf232e978d240216068a001a184e",
				"REOPEN_CONFIRMED_FINDING=HISTORICAL_PROJECTOR_OUTPUT_LINEAGE_BINDING",
				"HISTORICAL_PROJECTOR_OUTPUT_LINEAGE_BINDING=IMPLEMENTED_PENDING_EXACT_CI",
				"HISTORICAL_PROJECTION_PROVENANCE_RECONSTRUCTION=IMPLEMENTED_PENDING_EXACT_CI",
				"HISTORICAL_SELECTED_NEIGHBOR_PROVENANCE=IMPLEMENTED_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_REVIEW_STATUS=OPEN_PENDING_EXACT_CI_AND_FORMAL_RECLOSURE",
				"TestComposeRejectsHistoricalProjectionLineageDrift",
			},
			forbidden: []string{
				"Status: closed",
				"PROJECTION_PRODUCTION_REVIEW_STATUS=CLOSED",
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			fragments: []string{
				"## 38. Projection Production Review Reopen and Output-Lineage Correction",
				"PROJECTION_PRODUCTION_FORMAL_CLOSURE_REOPENED_BASELINE=0f1a31f56f4baf232e978d240216068a001a184e",
				"PROJECTION_PRODUCTION_APPROVED_LINEAGE_VERSION=historical-approved-projection-lineage-v1",
				"PROJECTION_PRODUCTION_HISTORICAL_PROJECTOR_OUTPUT_LINEAGE_BINDING=IMPLEMENTED_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_REVIEW_STATUS=OPEN_PENDING_EXACT_CI_AND_FORMAL_RECLOSURE",
			},
			forbidden: []string{
				"PROJECTION_PRODUCTION_REVIEW_STATUS=CLOSED",
			},
		},
		{
			path: "../../docs/DOCUMENT_INDEX.md",
			fragments: []string{
				"## Document 144 — Projection Production Review Hardening",
				"formal reopen for one Historical Projector output-lineage finding",
				"independent continuation fingerprint and canonical provenance",
				"Exact Continuous Integration and formal",
				"reclosure for the corrective increment are pending",
			},
			forbidden: []string{
				"is the formally closed review record\nfor one immutable production Horizon Plan",
			},
		},
	}
}

type exactMarkerRequirement struct {
	path    string
	markers map[string]string
}

func inspectExactClosureMarkers(apiRoot string) []string {
	requirements := []exactMarkerRequirement{
		{
			path: "../../docs/144_PROJECTION_PRODUCTION_REVIEW_HARDENING.md",
			markers: map[string]string{
				"ENGINEERING_CLOSURE_COMMIT":                         "c01b6ee0affff185adeda8e7fb0e1c39681cbe8c",
				"ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN":             "30624533886",
				"ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB":  "91136606689",
				"ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB":        "91136606649",
				"ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB":            "91136606715",
				"ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB":          "91136827987",
				"PRIOR_FORMAL_CLOSURE_COMMIT":                        "0f1a31f56f4baf232e978d240216068a001a184e",
				"PRIOR_FORMAL_CLOSURE_GITHUB_ACTIONS_RUN":            "30626948379",
				"PRIOR_FORMAL_CLOSURE_BACKEND_RACE_SAFETY_JOB":       "91144310170",
				"PRIOR_FORMAL_CLOSURE_BACKEND_QUALITY_JOB":           "91144310191",
				"PRIOR_FORMAL_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB": "91144310201",
				"PRIOR_FORMAL_CLOSURE_BACKEND_CONTAINER_JOB":         "91144541785",
				"FORMAL_CLOSURE_REOPENED_BASELINE":                   "0f1a31f56f4baf232e978d240216068a001a184e",
				"REOPEN_CONFIRMED_FINDING":                           "HISTORICAL_PROJECTOR_OUTPUT_LINEAGE_BINDING",
				"HISTORICAL_PROJECTOR_OUTPUT_LINEAGE_BINDING":        "IMPLEMENTED_PENDING_EXACT_CI",
				"HISTORICAL_PROJECTION_PROVENANCE_RECONSTRUCTION":    "IMPLEMENTED_PENDING_EXACT_CI",
				"HISTORICAL_SELECTED_NEIGHBOR_PROVENANCE":            "IMPLEMENTED_PENDING_EXACT_CI",
				"PERMANENT_REVIEW_AUDIT":                             "UPDATED_PENDING_EXACT_CI",
				"ENGINEERING_IMPLEMENTATION":                         "CORRECTIVE_IMPLEMENTATION_COMPLETE_PENDING_EXACT_CI",
				"ENGINEERING_DEBT":                                   "OPEN_PENDING_EXACT_CI_AND_FORMAL_RECLOSURE",
				"OPEN_CONFIRMED_FINDINGS":                            "0_PENDING_EXACT_CI",
				"UNCLASSIFIED_FINDINGS":                              "0_PENDING_EXACT_CI",
				"DEFERRED_FINDINGS":                                  "0_PENDING_EXACT_CI",
				"ADDITIONAL_CODE_FIXES_REQUIRED":                     "NO_PENDING_EXACT_CI",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED":              "YES",
				"PROJECTION_PRODUCTION_FORMAL_CLOSURE":               "REOPENED_PENDING_EXACT_CI_AND_FORMAL_RECLOSURE",
				"PROJECTION_PRODUCTION_REVIEW_STATUS":                "OPEN_PENDING_EXACT_CI_AND_FORMAL_RECLOSURE",
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			markers: map[string]string{
				"PROJECTION_PRODUCTION_ENGINEERING_CLOSURE_COMMIT":                      "c01b6ee0affff185adeda8e7fb0e1c39681cbe8c",
				"PROJECTION_PRODUCTION_ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN":          "30624533886",
				"PROJECTION_PRODUCTION_PRIOR_FORMAL_CLOSURE_COMMIT":                     "0f1a31f56f4baf232e978d240216068a001a184e",
				"PROJECTION_PRODUCTION_PRIOR_FORMAL_CLOSURE_GITHUB_ACTIONS_RUN":         "30626948379",
				"PROJECTION_PRODUCTION_FORMAL_CLOSURE_REOPENED_BASELINE":                "0f1a31f56f4baf232e978d240216068a001a184e",
				"PROJECTION_PRODUCTION_APPROVED_LINEAGE_VERSION":                        "historical-approved-projection-lineage-v1",
				"PROJECTION_PRODUCTION_HISTORICAL_PROJECTOR_OUTPUT_LINEAGE_BINDING":     "IMPLEMENTED_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_HISTORICAL_PROJECTION_PROVENANCE_RECONSTRUCTION": "IMPLEMENTED_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_HISTORICAL_SELECTED_NEIGHBOR_PROVENANCE":         "IMPLEMENTED_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_PERMANENT_REVIEW_AUDIT":                          "UPDATED_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_ENGINEERING_IMPLEMENTATION":                      "CORRECTIVE_IMPLEMENTATION_COMPLETE_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_ENGINEERING_DEBT":                                "OPEN_PENDING_EXACT_CI_AND_FORMAL_RECLOSURE",
				"PROJECTION_PRODUCTION_OPEN_CONFIRMED_FINDINGS":                         "0_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_UNCLASSIFIED_FINDINGS":                           "0_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_DEFERRED_FINDINGS":                               "0_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_ADDITIONAL_CODE_FIXES_REQUIRED":                  "NO_PENDING_EXACT_CI",
				"PROJECTION_PRODUCTION_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED":           "YES",
				"PROJECTION_PRODUCTION_FORMAL_CLOSURE":                                  "REOPENED_PENDING_EXACT_CI_AND_FORMAL_RECLOSURE",
				"PROJECTION_PRODUCTION_REVIEW_STATUS":                                   "OPEN_PENDING_EXACT_CI_AND_FORMAL_RECLOSURE",
			},
		},
	}

	failures := make([]string, 0)
	for _, requirement := range requirements {
		path := filepath.Clean(filepath.Join(apiRoot, requirement.path))
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("read exact markers from %s: %v", requirement.path, err))
			continue
		}
		valuesByKey := make(map[string][]string)
		for _, line := range strings.Split(string(contentBytes), "\n") {
			trimmed := strings.TrimSpace(line)
			key, value, found := strings.Cut(trimmed, "=")
			if !found || strings.TrimSpace(key) == "" {
				continue
			}
			key = strings.TrimSpace(key)
			valuesByKey[key] = append(valuesByKey[key], strings.TrimSpace(value))
		}
		for key, expected := range requirement.markers {
			values := valuesByKey[key]
			if len(values) == 0 {
				failures = append(failures, fmt.Sprintf("%s is missing exact marker %s=%s", requirement.path, key, expected))
				continue
			}
			for _, actual := range values {
				if actual != expected {
					failures = append(failures, fmt.Sprintf("%s marker %s has value %q, expected %q", requirement.path, key, actual, expected))
				}
			}
		}
	}
	return failures
}

func inspectRequirements(apiRoot string, requirements []requirement) []string {
	failures := make([]string, 0)
	for _, requirement := range requirements {
		path := filepath.Clean(filepath.Join(apiRoot, requirement.path))
		content, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("read %s: %v", requirement.path, err))
			continue
		}
		text := string(content)
		for _, fragment := range requirement.fragments {
			if !strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s is missing %q", requirement.path, fragment))
			}
		}
		for _, forbidden := range requirement.forbidden {
			if strings.Contains(text, forbidden) {
				failures = append(failures, fmt.Sprintf("%s contains forbidden %q", requirement.path, forbidden))
			}
		}
	}
	return failures
}

func resolveAPIRoot(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Abs(explicit)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	current := workingDirectory
	for {
		if fileExists(filepath.Join(current, "go.mod")) &&
			dirExists(filepath.Join(current, "internal", "projectionintelligence")) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("apps/api root not found from %s", workingDirectory)
		}
		current = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
