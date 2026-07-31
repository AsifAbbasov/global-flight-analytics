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
	strict := flag.Bool(
		"strict",
		false,
		"fail when a Projection Continuation review requirement is absent",
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
			"Projection continuation review audit: %v\n",
			err,
		)
		if *strict {
			os.Exit(1)
		}
		return
	}

	failures := inspectRequirements(
		apiRoot,
		reviewRequirements(),
	)
	failures = append(
		failures,
		inspectExactClosureMarkers(apiRoot)...,
	)
	if len(failures) == 0 {
		fmt.Println(
			"Projection continuation review audit: PASS",
		)
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Projection continuation review audit: %s\n",
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
			path: "internal/projectionintelligence/projectioncontinuation/continuation_core.go",
			fragments: []string{
				`Version    = "local-historical-neighbor-continuation-v3"`,
				`FingerprintVersion         = "local-historical-neighbor-continuation-fingerprint-v3"`,
				`FallbackFingerprintVersion = "local-historical-neighbor-fallback-fingerprint-v3"`,
			},
			forbidden: []string{
				`local-historical-neighbor-continuation-v2`,
				`ErrCurrentTrajectoryUnavailable`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/config.go",
			fragments: []string{
				`config.NeighborSpreadMultiplier < 1`,
				`config.MaximumConfidenceLoss >= 1`,
				`neighbor spread multiplier must be finite and at least one`,
				`maximum confidence loss must be finite, non-negative, and less than one`,
			},
			forbidden: []string{
				`neighbor spread multiplier must be finite and greater than zero`,
				`maximum confidence loss must be finite and between zero and one`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/continuation_combination.go",
			fragments: []string{
				`type sampleCombination struct`,
				`composeUncertainty(`,
				`combined := configured + disagreement`,
				`independence is not established`,
				`effectiveWeightedSupportRatio(`,
				`totalWeight * totalWeight / squaredWeightSum`,
				`uncertaintyAgreementFactor(`,
				`agreementFactor *`,
				`"effective_weighted_neighbor_support"`,
				`"neighbor_agreement"`,
			},
			forbidden: []string{
				`math.Max(`,
				`float64(len(samples)) /`,
				`pattern_confidence_support_and_horizon_decay`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/workflow.go",
			fragments: []string{
				`if !selectedCandidateEvidenceMatches(`,
				`historical_selected_candidate_evidence_mismatch`,
				`confidenceComplete   bool`,
				`if !combination.confidenceAvailable {`,
			},
			forbidden: []string{
				`if approvedEvidence != nil &&`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/result_builder.go",
			fragments: []string{
				`historical_continuation_confidence_none`,
				`Scope:   "confidence"`,
				`!confidenceComplete ||`,
				`neighbor_disagreement_uncertainty_and_confidence`,
				`effective_weighted_support`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/interpolation.go",
			fragments: []string{
				`continuationPointLess(`,
				`math.Float64bits(leftValues[index])`,
				`strings.TrimSpace(left.ID)`,
			},
			forbidden: []string{
				`indexed[left].index <`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/geodesic.go",
			fragments: []string{
				`weightedMeanVectorNormEpsilon = 1e-12`,
				`vectorNorm := math.Sqrt(`,
				`vectorNorm <= weightedMeanVectorNormEpsilon`,
			},
			forbidden: []string{
				`horizontal == 0 && z == 0`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/continuation_fallback.go",
			fragments: []string{
				`"%w: %w"`,
			},
			forbidden: []string{
				`"%w: %v"`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/approved_evidence_test.go",
			fragments: []string{
				`TestProjectApprovedUsesAuthorizedEvidenceWithoutReevaluation`,
				`TestProjectApprovedRejectsPatternFromAnotherSelection`,
				`TestProjectApprovedRejectsAnchorTimestampMismatch`,
				`TestProjectApprovedPreservesObservedCandidateProvenance`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/plausibility_interpolation_test.go",
			fragments: []string{
				`TestInterpolationPlausibilityRejectsUnsafeSegments`,
				`exact endpoint still validates segment`,
				`TestInterpolationPlausibilityAcceptsValidSegment`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/plausibility_workflow_test.go",
			fragments: []string{
				`TestProjectFallsBackWhenPlausibilityRemovesRequiredSupport`,
				`TestProjectMarksLimitedWhenPlausibilityFiltersPartialSupport`,
				`TestPlausibilityPolicyChangesInputFingerprint`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/config_test.go",
			fragments: []string{
				`spread multiplier below one`,
				`config.NeighborSpreadMultiplier = 0.99`,
				`confidence loss reaches one`,
				`config.MaximumConfidenceLoss = 1`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/review_closure_test.go",
			fragments: []string{
				`TestComposeUncertaintyPreservesBothComponents`,
				`TestEffectiveWeightedSupportPenalizesConcentration`,
				`TestNeighborDisagreementRaisesUncertaintyAndLowersConfidence`,
				`TestZeroPointConfidenceMarksResultLimited`,
				`TestWeightedMeanRejectsNearAntipodalSamples`,
				`TestFallbackPreservesUnderlyingCause`,
				`TestStandaloneProjectRejectsCandidateEvidenceDrift`,
				`TestContinuationFingerprintDistinguishesTruncatedPlan`,
				`TestProjectRejectsIrregularForecastTimes`,
				`TestTrajectorySnapshotOrdersEqualTimestampsCanonically`,
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				`Run projection continuation review audit`,
				`go run ./tools/projectioncontinuationreviewaudit -strict`,
			},
		},
		{
			path: "../../docs/141_PROJECTION_CONTINUATION_REVIEW_HARDENING.md",
			fragments: []string{
				`Status: closed`,
				`INTERPOLATION_PLAUSIBILITY=CI_CONFIRMED_COMMIT_739073de31e4c1da2aa105d495bc789a294cb3c9_RUN_30576928637`,
				`ENGINEERING_CLOSURE_COMMIT=13838c4273a3be6bde63835e1d8f51af6f6daa21`,
				`ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30593549087`,
				`ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91040848886`,
				`ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91040848927`,
				`ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91040848967`,
				`ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91041042383`,
				`UNCERTAINTY_CONFIDENCE_CONSISTENCY=CI_CONFIRMED`,
				`GEODESIC_NUMERICAL_STABILITY=CI_CONFIRMED`,
				`EFFECTIVE_WEIGHTED_SUPPORT=CI_CONFIRMED`,
				`FALLBACK_ERROR_CAUSE_PRESERVATION=CI_CONFIRMED`,
				`STANDALONE_CANDIDATE_EVIDENCE_BINDING=CI_CONFIRMED`,
				`PERMANENT_REVIEW_AUDIT=CI_CONFIRMED`,
				`ENGINEERING_IMPLEMENTATION=COMPLETE`,
				`ENGINEERING_DEBT=CLOSED`,
				`OPEN_CONFIRMED_FINDINGS=0`,
				`UNCLASSIFIED_FINDINGS=0`,
				`DEFERRED_FINDINGS=0`,
				`ADDITIONAL_CODE_FIXES_REQUIRED=NO`,
				`FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO`,
				`PROJECTION_CONTINUATION_FORMAL_CLOSURE=COMPLETE`,
				`PROJECTION_CONTINUATION_REVIEW_STATUS=CLOSED`,
			},
			forbidden: []string{
				`Status: engineering implementation complete, exact Continuous Integration and formal closure pending`,
				`IMPLEMENTED_PENDING_EXACT_CI`,
				`ENFORCED_PENDING_EXACT_CI`,
				`COMPLETE_PENDING_EXACT_CI`,
				`0_PENDING_EXACT_CI`,
				`FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES`,
				`PROJECTION_CONTINUATION_REVIEW_STATUS=OPEN`,
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			fragments: []string{
				`## 35. Projection Continuation Engineering and Formal Review Closure`,
				`PROJECTION_CONTINUATION_VERSION=local-historical-neighbor-continuation-v3`,
				`PROJECTION_CONTINUATION_ADDITIVE_UNCERTAINTY_COMPOSITION=ENFORCED`,
				`PROJECTION_CONTINUATION_EFFECTIVE_WEIGHTED_SUPPORT=ENFORCED`,
				`PROJECTION_CONTINUATION_DISAGREEMENT_CONFIDENCE_PENALTY=ENFORCED`,
				`PROJECTION_CONTINUATION_NEAR_ANTIPODAL_GUARD=ENFORCED`,
				`PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_COMMIT=13838c4273a3be6bde63835e1d8f51af6f6daa21`,
				`PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30593549087`,
				`PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91040848886`,
				`PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91040848927`,
				`PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91040848967`,
				`PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91041042383`,
				`PROJECTION_CONTINUATION_PERMANENT_REVIEW_AUDIT=CI_CONFIRMED`,
				`PROJECTION_CONTINUATION_ENGINEERING_IMPLEMENTATION=COMPLETE`,
				`PROJECTION_CONTINUATION_ENGINEERING_DEBT=CLOSED`,
				`PROJECTION_CONTINUATION_OPEN_CONFIRMED_FINDINGS=0`,
				`PROJECTION_CONTINUATION_UNCLASSIFIED_FINDINGS=0`,
				`PROJECTION_CONTINUATION_DEFERRED_FINDINGS=0`,
				`PROJECTION_CONTINUATION_ADDITIONAL_CODE_FIXES_REQUIRED=NO`,
				`PROJECTION_CONTINUATION_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO`,
				`PROJECTION_CONTINUATION_FORMAL_CLOSURE=COMPLETE`,
				`PROJECTION_CONTINUATION_REVIEW_STATUS=CLOSED`,
			},
			forbidden: []string{
				`PROJECTION_CONTINUATION_PERMANENT_REVIEW_AUDIT=ENFORCED_PENDING_EXACT_CI`,
				`PROJECTION_CONTINUATION_ENGINEERING_IMPLEMENTATION=COMPLETE_PENDING_EXACT_CI`,
				`PROJECTION_CONTINUATION_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES`,
				`PROJECTION_CONTINUATION_REVIEW_STATUS=OPEN`,
			},
		},
		{
			path: "../../docs/DOCUMENT_INDEX.md",
			fragments: []string{
				`is the formally closed review`,
				`zero open, unclassified or deferred findings`,
			},
		},
	}
}

type exactMarkerRequirement struct {
	path    string
	markers map[string]string
}

func inspectExactClosureMarkers(
	apiRoot string,
) []string {
	requirements := []exactMarkerRequirement{
		{
			path: "../../docs/141_PROJECTION_CONTINUATION_REVIEW_HARDENING.md",
			markers: map[string]string{
				"ENGINEERING_CLOSURE_COMMIT":                        "13838c4273a3be6bde63835e1d8f51af6f6daa21",
				"ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN":            "30593549087",
				"ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB": "91040848886",
				"ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB":       "91040848927",
				"ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB":           "91040848967",
				"ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB":         "91041042383",
				"PERMANENT_REVIEW_AUDIT":                            "CI_CONFIRMED",
				"ENGINEERING_IMPLEMENTATION":                        "COMPLETE",
				"ENGINEERING_DEBT":                                  "CLOSED",
				"OPEN_CONFIRMED_FINDINGS":                           "0",
				"UNCLASSIFIED_FINDINGS":                             "0",
				"DEFERRED_FINDINGS":                                 "0",
				"ADDITIONAL_CODE_FIXES_REQUIRED":                    "NO",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED":             "NO",
				"PROJECTION_CONTINUATION_FORMAL_CLOSURE":            "COMPLETE",
				"PROJECTION_CONTINUATION_REVIEW_STATUS":             "CLOSED",
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			markers: map[string]string{
				"PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_COMMIT":                        "13838c4273a3be6bde63835e1d8f51af6f6daa21",
				"PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN":            "30593549087",
				"PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB": "91040848886",
				"PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB":       "91040848927",
				"PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB":           "91040848967",
				"PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB":         "91041042383",
				"PROJECTION_CONTINUATION_PERMANENT_REVIEW_AUDIT":                            "CI_CONFIRMED",
				"PROJECTION_CONTINUATION_ENGINEERING_IMPLEMENTATION":                        "COMPLETE",
				"PROJECTION_CONTINUATION_ENGINEERING_DEBT":                                  "CLOSED",
				"PROJECTION_CONTINUATION_OPEN_CONFIRMED_FINDINGS":                           "0",
				"PROJECTION_CONTINUATION_UNCLASSIFIED_FINDINGS":                             "0",
				"PROJECTION_CONTINUATION_DEFERRED_FINDINGS":                                 "0",
				"PROJECTION_CONTINUATION_ADDITIONAL_CODE_FIXES_REQUIRED":                    "NO",
				"PROJECTION_CONTINUATION_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED":             "NO",
				"PROJECTION_CONTINUATION_FORMAL_CLOSURE":                                    "COMPLETE",
				"PROJECTION_CONTINUATION_REVIEW_STATUS":                                     "CLOSED",
			},
		},
	}

	failures := make([]string, 0)
	for _, requirement := range requirements {
		path := filepath.Clean(
			filepath.Join(apiRoot, requirement.path),
		)
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			failures = append(
				failures,
				fmt.Sprintf(
					"read exact markers from %s: %v",
					requirement.path,
					err,
				),
			)
			continue
		}

		valuesByKey := make(map[string][]string)
		for _, line := range strings.Split(
			string(contentBytes),
			"\n",
		) {
			trimmed := strings.TrimSpace(line)
			key, value, found := strings.Cut(
				trimmed,
				"=",
			)
			if !found || strings.TrimSpace(key) == "" {
				continue
			}
			key = strings.TrimSpace(key)
			valuesByKey[key] = append(
				valuesByKey[key],
				strings.TrimSpace(value),
			)
		}

		for key, expected := range requirement.markers {
			values := valuesByKey[key]
			if len(values) == 0 {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s is missing exact marker %s=%s",
						requirement.path,
						key,
						expected,
					),
				)
				continue
			}
			for _, actual := range values {
				if actual == expected {
					continue
				}
				failures = append(
					failures,
					fmt.Sprintf(
						"%s marker %s has value %q, expected %q",
						requirement.path,
						key,
						actual,
						expected,
					),
				)
			}
		}
	}

	return failures
}

func inspectRequirements(
	apiRoot string,
	requirements []requirement,
) []string {
	failures := make([]string, 0)
	for _, item := range requirements {
		path := filepath.Clean(
			filepath.Join(apiRoot, item.path),
		)
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			failures = append(
				failures,
				fmt.Sprintf(
					"read %s: %v",
					item.path,
					err,
				),
			)
			continue
		}
		content := string(contentBytes)
		for _, fragment := range item.fragments {
			if !strings.Contains(content, fragment) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s is missing %q",
						item.path,
						fragment,
					),
				)
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(content, fragment) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s contains forbidden %q",
						item.path,
						fragment,
					),
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
		return "", err
	}
	for {
		if root, rootErr := validateAPIRoot(current); rootErr == nil {
			return root, nil
		}
		candidate := filepath.Join(current, "apps", "api")
		if root, rootErr := validateAPIRoot(candidate); rootErr == nil {
			return root, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf(
		"apps/api root containing go.mod was not found",
	)
}

func validateAPIRoot(candidate string) (string, error) {
	absolute, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(filepath.Join(absolute, "go.mod"))
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("go.mod is unavailable")
	}
	return absolute, nil
}
