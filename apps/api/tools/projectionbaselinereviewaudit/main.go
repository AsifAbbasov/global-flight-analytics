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
		"fail when a Projection Baseline review requirement is absent",
	)
	root := flag.String(
		"root",
		"",
		"optional path to the apps/api module root",
	)
	flag.Parse()

	apiRoot, err := resolveAPIRoot(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Projection baseline review audit: %v\n", err)
		if *strict {
			os.Exit(1)
		}
		return
	}

	failures := inspectRequirements(apiRoot, reviewRequirements())
	if len(failures) == 0 {
		fmt.Println("Projection baseline review audit: PASS")
		return
	}

	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Projection baseline review audit: %s\n",
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
			path: "internal/projectionintelligence/projectionbaseline/baseline.go",
			fragments: []string{
				`Version    = "short-horizon-kinematic-baseline-v3"`,
				"ErrBaselineUnavailable",
				"ErrHorizonPlanInvalid",
				"if err := plan.Validate(); err != nil",
				"buildCutoffSnapshot(",
				"validateProjectionEligibilityEvaluation(",
				"selectLatestProjectionPoint(",
				"effectiveHorizontalFallbackPolicy()",
				`Code:    "projection_eligibility_altitude_fallback"`,
				`Code:    "projection_on_ground_stationary_model"`,
				"eligibilityPolicyInputReference(",
				"baselineExplanations(",
				"if latestPoint.OnGround {",
			},
			forbidden: []string{
				"trajectorySnapshotAt(",
				"func clampUnit(",
				`Version    = "short-horizon-kinematic-baseline-v1"`,
				`Version    = "short-horizon-kinematic-baseline-v2"`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/cutoff_snapshot.go",
			fragments: []string{
				"func buildCutoffSnapshot(",
				"completedSegmentsAt(",
				"completedCoverageGapsAt(",
				"trajectoryquality.TrajectoryScore(",
				"qualityEvidenceCoversLatestPoint(",
				"item.EndTime.UTC().After(asOfTime)",
				"QualityEvidenceAvailable",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/postgres_source.go",
			fragments: []string{
				"item.Segments = filterSegmentsAt(",
				"item.QualityScore = trajectoryquality.TrajectoryScore(",
				"filterCoverageGapsAt(",
				"item.EndTime.After(cutoff)",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/fingerprint.go",
			fragments: []string{
				"plan.Fingerprint",
				"writeEligibilityPolicyFingerprintEvidence(",
				"altitude.normalizedReference()",
				"config.effectiveMaximumObservationAge().Nanoseconds()",
				"maximumObservationAgeConfidenceLoss",
				"maximumSupportedGroundSpeedMPS",
				"maximumSupportedAbsoluteVerticalRateMPS",
				"minimumSupportedAltitudeM",
				"maximumSupportedAltitudeM",
				"config.effectiveHorizontalFallbackPolicy()",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/fingerprint_evidence.go",
			fragments: []string{
				`FingerprintVersion = "projection-baseline-input-fingerprint-v4"`,
				"point.SourceName",
				"point.OnGround",
				"point.TelemetryAvailabilityKnown",
				"segment.QualityScore",
				"gap.EndTime",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/provenance.go",
			fragments: []string{
				"selectLatestProjectionPoint(",
				"eligibilityPolicyInputReference(",
				"InputFingerprint: inputFingerprint(",
				"LatestInputObservedAt",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/confidence.go",
			fragments: []string{
				"maximumObservationAgeConfidenceLoss",
				"LatestObservedAt",
				"MaximumObservationAge",
				"MaximumHorizonLoss",
				"ageProgress",
				"horizonProgress",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/kinematics.go",
			fragments: []string{
				"maximumSupportedGroundSpeedMPS",
				"maximumSupportedAbsoluteVerticalRateMPS",
				"minimumSupportedAltitudeM",
				"maximumSupportedAltitudeM",
				"maximumSupportedOnGroundSpeedMPS",
				"maximumSupportedOnGroundVerticalRateMPS",
				`Code:    "projection_heading_out_of_bounds"`,
				`Code:    "projection_on_ground_motion_out_of_bounds"`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/altitude.go",
			fragments: []string{
				"type altitudeReference string",
				`altitudeReferenceGeometric   altitudeReference = "geometric"`,
				`altitudeReferenceBarometric  altitudeReference = "barometric"`,
				"type altitudeSelection struct",
				"func selectAltitude(",
				"provenanceLimitation() string",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/latest_observation.go",
			fragments: []string{
				"func selectLatestProjectionPoint(",
				"equivalentProjectionEvidence(",
				`Code:    "projection_latest_observation_ambiguous"`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/eligibility_policy.go",
			fragments: []string{
				"func validateProjectionEligibilityEvaluation(",
				"projection eligibility must return exactly one projection decision",
				"reason.IsKnown()",
				"ProjectionPolicyIdentity()",
				"writeEligibilityPolicyFingerprintEvidence(",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/horizontal_fallback.go",
			fragments: []string{
				"type HorizontalFallbackPolicy string",
				`HorizontalFallbackAllowLimited HorizontalFallbackPolicy = "allow_limited"`,
				`HorizontalFallbackReject       HorizontalFallbackPolicy = "reject"`,
				"ReasonMissingAltitude",
			},
		},
		{
			path: "internal/analytics/trajectoryeligibility/policy_identity.go",
			fragments: []string{
				`ProjectionPolicyName    = "trajectory_projection_eligibility"`,
				`ProjectionPolicyVersion = "trajectory-projection-eligibility-v1"`,
				"func (evaluator *Evaluator) ProjectionPolicyIdentity() PolicyIdentity",
				"policy.RequireAltitude",
			},
		},
		{
			path: "internal/analytics/trajectoryeligibility/reason_validation.go",
			fragments: []string{
				"func (reason ReasonCode) IsKnown() bool",
				"ReasonMissingAltitude",
				"ReasonStaleObservations",
				"ReasonInsufficientRecentContinuity",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/cutoff_integrity_test.go",
			fragments: []string{
				"TestProjectExcludesFutureAggregateQualityEvidence",
				"TestProjectExcludesCoverageGapNotCompletedAtCutoff",
				"TestUnavailableResultPreservesCutoffProvenance",
				"TestInputFingerprintCoversSourceAndGroundState",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/kinematic_confidence_test.go",
			fragments: []string{
				"TestProjectConfidenceDecreasesWithObservationAge",
				"TestProjectRejectsOutOfBoundsKinematics",
				"TestProjectReportsSelectedAltitudeReference",
				"TestProjectRejectsUnsafeOnGroundMotionWhenAllowed",
				"TestNilBaselineReturnsLifecycleError",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/collaboration_integrity_test.go",
			fragments: []string{
				"TestProjectRejectsDuplicateProjectionEligibilityDecision",
				"TestProjectRejectsUnknownEligibilityReason",
				"TestProjectAllowsExplicitHorizontalFallbackForMissingAltitude",
				"TestProjectRejectsConflictingLatestObservations",
				"TestProjectUsesStationaryOnGroundModel",
				"TestDefaultEligibilityPolicyIsPublishedInProvenance",
				"TestHorizontalFallbackCanBeRejected",
			},
		},
		{
			path: "internal/projectionintelligence/projectionbaseline/horizon_plan_validation_test.go",
			fragments: []string{
				"TestProjectRejectsInvalidHorizonPlannerResult",
			},
		},
		{
			path: "internal/projectionintelligence/projectionread/postgres_source_cutoff_test.go",
			fragments: []string{
				"TestFilterSegmentsAtExcludesIncompleteSegment",
				"TestFilterCoverageGapsAtExcludesGapEndingAfterCutoff",
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			fragments: []string{
				"## 28. Projection Baseline Review Hardening",
				"PROJECTION_BASELINE_METHOD_VERSION=short-horizon-kinematic-baseline-v3",
				"PROJECTION_BASELINE_INPUT_FINGERPRINT_VERSION=projection-baseline-input-fingerprint-v4",
				"PROJECTION_BASELINE_CUTOFF_ISOLATION=ENFORCED",
				"PROJECTION_BASELINE_OBSERVATION_AGE_CONFIDENCE=ENFORCED",
				"PROJECTION_BASELINE_ELIGIBILITY_POLICY_PROVENANCE=ENFORCED",
				"apps/api/tools/projectionbaselinereviewaudit",
			},
		},
		{
			path: "../../docs/136_PROJECTION_BASELINE_REVIEW_HARDENING.md",
			fragments: []string{
				"# Projection Baseline Review Hardening",
				"AUTHORITATIVE_BASELINE_COMMIT=b4da27772fad838bf2a237ff9989621bfae6d5f2",
				"CUTOFF_INTEGRITY_COMMIT=0f2c1b2c6f91f104b8e0880e85dc8144fed6a910",
				"KINEMATIC_CONFIDENCE_COMMIT=af9c377193c21c048721e9cc28bf885d6ad276ec",
				"COLLABORATION_INTEGRITY_COMMIT=560e4ed15cabbf0042110e00363a3a7c4d0c0d2e",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"PROJECTION_BASELINE_ENGINEERING_IMPLEMENTATION=COMPLETE",
				"PERMANENT_AUDIT_COMMIT=",
				"PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=",
			},
		},
		{
			path: "../../docs/DOCUMENT_INDEX.md",
			fragments: []string{
				"## Document 136 — Projection Baseline Review Hardening",
				"136_PROJECTION_BASELINE_REVIEW_HARDENING.md",
				"cutoff-safe quality recomputation",
				"permanent Projection Baseline",
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				"Run projection baseline review audit",
				"go run ./tools/projectionbaselinereviewaudit -strict",
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
		for _, fragment := range item.fragments {
			if !strings.Contains(text, fragment) {
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
			if strings.Contains(text, fragment) {
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
