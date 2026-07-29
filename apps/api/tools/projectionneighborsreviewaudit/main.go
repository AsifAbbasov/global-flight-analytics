package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type requirement struct {
	path      string
	fragments []string
	forbidden []string
}

func main() {
	strict := flag.Bool(
		"strict",
		false,
		"fail when a Projection Neighbors review requirement is absent",
	)
	root := flag.String(
		"root",
		"",
		"optional path to the apps/api module root",
	)
	flag.Parse()

	apiRoot, err := resolveAPIRoot(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Projection neighbors review audit: %v\n", err)
		if *strict {
			os.Exit(1)
		}
		return
	}

	failures := inspectRequirements(apiRoot, reviewRequirements())
	if len(failures) == 0 {
		fmt.Println("Projection neighbors review audit: PASS")
		return
	}

	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Projection neighbors review audit: %s\n",
			failure,
		)
	}
	if *strict {
		os.Exit(1)
	}
}

func reviewRequirements() []requirement {
	return []requirement{
		{
			path: "internal/projectionintelligence/projectionneighbors/model.go",
			fragments: []string{
				`Version            = "projection-historical-neighbor-selection-v5"`,
				`FingerprintVersion = "projection-historical-neighbor-selection-fingerprint-v5"`,
				"RejectionDuplicateCandidate",
				"RejectionContinuationDiscontinuous",
				"RejectionCrossRoute",
				"CandidateEvaluationTruncated bool",
				"QualifiedSelectionLimited    bool",
				"Deprecated: use CandidateEvaluationTruncated.",
				"candidate evaluation truncation does not match checked and input counts",
				"deprecated truncation alias does not match candidate evaluation truncation",
				"qualified selection limiting does not match qualified and selected counts",
				"qualified selection limiting must fill the configured selection limit",
			},
			forbidden: []string{
				`Version            = "projection-historical-neighbor-selection-v4"`,
				`FingerprintVersion = "projection-historical-neighbor-selection-fingerprint-v4"`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionneighbors/selector.go",
			fragments: []string{
				"context, err := selector.prepareSelectionContext(",
				"evaluation, err := selector.evaluateCandidatePool(",
				"return selector.assembleSelectionResult(",
				"canonicalPointLess(",
			},
			forbidden: []string{
				"for _, prepared := range candidates {",
				"qualifiedCandidateCount := len(neighbors)",
			},
		},
		{
			path: "internal/projectionintelligence/projectionneighbors/selection_context.go",
			fragments: []string{
				"type selectionContext struct",
				"func (selector *Selector) prepareSelectionContext(",
				"prepareRouteScope(",
				"snapshotAt(",
				"prepareCandidatePool(",
				"ErrCurrentTrajectoryNotComparable",
			},
		},
		{
			path: "internal/projectionintelligence/projectionneighbors/candidate_pool.go",
			fragments: []string{
				"candidateIDCounts := make(",
				"candidateIDCounts[strings.TrimSpace(candidate.ID)]++",
				"candidateIDCounts[candidateID] > 1",
				"routeEvidence := routeScope.evidenceByCandidate[candidateID]",
				"RejectionCrossRoute",
				"pool.CandidateEvaluationTruncated = true",
				"pool.Truncated = true",
				"pool.Candidates[:config.MaximumCandidateCount]",
			},
		},
		{
			path: "internal/projectionintelligence/projectionneighbors/candidate_evaluation.go",
			fragments: []string{
				"type candidateEvaluation struct",
				"func (selector *Selector) evaluateCandidatePool(",
				"func (selector *Selector) evaluateCandidate(",
				"anchorSearch := findAnchor(",
				"anchorSearchFailureDiscontinuous",
				"candidateSimilarityFailure(err)",
				"ErrSimilarityEngineFailed",
				"ErrSimilarityEvidenceInvalid",
				"RejectionSimilarityBelowMinimum",
			},
		},
		{
			path: "internal/projectionintelligence/projectionneighbors/anchor.go",
			fragments: []string{
				"type AnchorEvidence struct",
				"MaximumObservedGap",
				"anchorSearchFailureDiscontinuous",
				"for segmentStart := 0; segmentStart < len(candidatePoints); {",
				"continuationEndIndex := anchorStart + 1",
				"gap > maximumContinuationGap",
				"maximumObservedGap(",
			},
		},
		{
			path: "internal/projectionintelligence/projectionneighbors/fingerprint.go",
			fragments: []string{
				"routeScope RouteScope",
				"routeScope.InputFingerprint",
				"config.effectiveMaximumContinuationGap()",
				"canonicalPointLess(",
				"trajectoryFingerprintSortKey(",
			},
		},
		{
			path: "internal/projectionintelligence/projectionneighbors/route_scope.go",
			fragments: []string{
				`RouteScopeVersion = "projection-neighbor-route-scope-v1"`,
				`RouteScopeUniform  RouteScopeMode = "uniform_route_scope"`,
				`RouteScopeExplicit RouteScopeMode = "explicit_candidate_routes"`,
				"func UniformRouteScope(",
				"func ExplicitRouteScope(",
				"func (scope RouteScope) ValidateForCandidates(",
				"duplicate route evidence for candidate",
				"route evidence is missing for candidate",
				"route scope input fingerprint is invalid",
				"candidate route evidence fingerprint is invalid",
			},
		},
		{
			path: "internal/projectionintelligence/projectionneighbors/result_assembly.go",
			fragments: []string{
				"func (selector *Selector) assembleSelectionResult(",
				"func rankAndLimitNeighbors(",
				"qualifiedSelectionLimited := qualifiedCandidateCount > selectionLimit",
				"CandidateEvaluationTruncated: candidateEvaluationTruncated",
				"QualifiedSelectionLimited:    ranked.qualifiedSelectionLimited",
				`Code: "qualified_neighbors_limited"`,
				`Code:    "candidate_evaluation_truncated"`,
				"func selectionStatus(",
			},
		},
		{
			path: "internal/projectionintelligence/projectionneighbors/continuation_integrity_test.go",
			fragments: []string{
				"TestConfigValidateRejectsNegativeContinuationGap",
				"TestSelectionFingerprintIncludesContinuationGapPolicy",
				"TestSelectRejectsDiscontinuousContinuation",
				"TestFindAnchorPublishesContinuousEvidence",
				"TestFindAnchorHandlesLargeTrajectory",
			},
		},
		{
			path: "internal/projectionintelligence/projectionneighbors/route_scope_test.go",
			fragments: []string{
				"TestSelectRejectsMissingRouteScope",
				"TestSelectionFingerprintIncludesRouteScope",
				"TestSelectRejectsCrossRouteCandidateBeforeSimilarity",
				"TestSelectRejectsIncompleteExplicitRouteScope",
				"TestSelectRejectsTamperedRouteScopeFingerprint",
			},
		},
		{
			path: "internal/projectionintelligence/projectionneighbors/selection_pipeline_test.go",
			fragments: []string{
				"TestSelectSeparatesEvaluationTruncationFromSelectionLimiting",
				"TestSelectReportsCandidateEvaluationTruncationIndependently",
				"TestResultValidateRejectsFalseExplicitLimitFlags",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/route_scope.go",
			fragments: []string{
				`routeScopedCandidateEvidenceVersion = "projection-read-route-scoped-candidates-v1"`,
				"func routeScopedHistoricalCandidateEvidence(",
				"projectionneighbors.UniformRouteScope(",
				"scope.ValidateForCandidates(candidates)",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/postgres_snapshot_source.go",
			fragments: []string{
				"source.LoadHistoricalCandidates(",
				"routeScopedHistoricalCandidateEvidence(",
				"snapshot.HistoricalCandidateRouteScope = routeScopePointer(routeScope)",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/route_scope.go",
			fragments: []string{
				"func validateHistoricalCandidateRouteScope(",
				"scope.ValidateForCandidates(candidates)",
				"historical candidate route scope does not match the current route",
				"func routeScopeFromState(",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/route_scope_test.go",
			fragments: []string{
				"TestProjectForwardsRouteScopeToNeighborSelector",
				"selector.request.RouteScope.InputFingerprint",
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			fragments: []string{
				"## 29. Projection Neighbors Review Hardening",
				"PROJECTION_NEIGHBORS_SELECTION_VERSION=projection-historical-neighbor-selection-v5",
				"PROJECTION_NEIGHBORS_FINGERPRINT_VERSION=projection-historical-neighbor-selection-fingerprint-v5",
				"PROJECTION_NEIGHBORS_ROUTE_SCOPE=SOURCE_ATTESTED",
				"PROJECTION_NEIGHBORS_CONTINUATION_GAP_POLICY=ENFORCED",
				"PROJECTION_NEIGHBORS_SELECTOR_PIPELINE=DECOMPOSED",
				"PROJECTION_NEIGHBORS_PERMANENT_AUDIT_COMMIT=PENDING",
				"PROJECTION_NEIGHBORS_REVIEW_STATUS=IMPLEMENTED_PENDING_PERMANENT_AUDIT_CI",
				"apps/api/tools/projectionneighborsreviewaudit",
			},
		},
		{
			path: "../../docs/137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md",
			fragments: []string{
				"# Projection Neighbors Review Hardening",
				"Status: implemented pending permanent audit Continuous Integration",
				"AUTHORITATIVE_BASELINE_COMMIT=e13a117f969e2922d09a7804fe50005d01bc2ecf",
				"CANDIDATE_INTEGRITY_COMMIT=e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff",
				"CONTINUATION_INTEGRITY_COMMIT=911a1b102c68af2746a13bfca48b008cf7225ff8",
				"ROUTE_SCOPE_INTEGRITY_COMMIT=3eee05fb44484aa6e389af66520aba23d4ae277e",
				"SELECTOR_PIPELINE_COMMIT=353d19bc97f561e1897ece1967e7304c0e10b5fb",
				"PERMANENT_AUDIT_COMMIT=PENDING",
				"PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=PENDING",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
				"PROJECTION_NEIGHBORS_ENGINEERING_IMPLEMENTATION=COMPLETE",
				"PROJECTION_NEIGHBORS_ADDITIONAL_CODE_FIXES_REQUIRED=NO",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES",
				"PROJECTION_NEIGHBORS_REVIEW_STATUS=IMPLEMENTED_PENDING_PERMANENT_AUDIT_CI",
			},
		},
		{
			path: "../../docs/DOCUMENT_INDEX.md",
			fragments: []string{
				"## Document 137 — Projection Neighbors Review Hardening",
				"137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md",
				"source-attested origin-destination route scope",
				"candidate-evaluation truncation",
				"353d19bc97f561e1897ece1967e7304c0e10b5fb",
				"pending permanent audit Continuous Integration",
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				"Run projection neighbors review audit",
				"go run ./tools/projectionneighborsreviewaudit -strict",
			},
		},
	}
}

func inspectRequirements(
	apiRoot string,
	requirements []requirement,
) []string {
	failures := make([]string, 0)
	for _, item := range requirements {
		path := filepath.Clean(filepath.Join(apiRoot, item.path))
		content, err := os.ReadFile(path)
		if err != nil {
			failures = append(
				failures,
				fmt.Sprintf("read %s: %v", item.path, err),
			)
			continue
		}

		text := string(content)
		normalizedText := normalizeRequirementText(text)
		for _, fragment := range item.fragments {
			if !strings.Contains(
				normalizedText,
				normalizeRequirementText(fragment),
			) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s misses %q",
						item.path,
						fragment,
					),
				)
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(
				normalizedText,
				normalizeRequirementText(fragment),
			) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s retains forbidden %q",
						item.path,
						fragment,
					),
				)
			}
		}
	}

	sort.Strings(failures)
	return failures
}

func normalizeRequirementText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func resolveAPIRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return validateAPIRoot(explicit)
	}

	current, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("read working directory: %w", err)
	}

	for candidate := filepath.Clean(current); ; candidate = filepath.Dir(candidate) {
		if root, err := validateAPIRoot(candidate); err == nil {
			return root, nil
		}
		nested := filepath.Join(candidate, "apps", "api")
		if root, err := validateAPIRoot(nested); err == nil {
			return root, nil
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			break
		}
	}

	return "", fmt.Errorf("could not locate apps/api module root from %s", current)
}

func validateAPIRoot(candidate string) (string, error) {
	absolute, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve API root %q: %w", candidate, err)
	}
	moduleFile := filepath.Join(absolute, "go.mod")
	content, err := os.ReadFile(moduleFile)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", moduleFile, err)
	}
	if !strings.Contains(
		string(content),
		"module github.com/AsifAbbasov/global-flight-analytics/apps/api",
	) {
		return "", fmt.Errorf("%s is not the Global Flight Analytics API module", absolute)
	}
	return absolute, nil
}
