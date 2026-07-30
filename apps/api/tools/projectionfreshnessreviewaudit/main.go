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
		"fail when a Projection Freshness review requirement is absent",
	)
	root := flag.String(
		"root",
		"",
		"optional path to the apps/api module root",
	)
	flag.Parse()

	apiRoot, err := resolveAPIRoot(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Projection freshness review audit: %v\n", err)
		if *strict {
			os.Exit(1)
		}
		return
	}

	failures := inspectRequirements(apiRoot, reviewRequirements())
	if len(failures) == 0 {
		fmt.Println("Projection freshness review audit: PASS")
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(os.Stderr, "Projection freshness review audit: %s\n", failure)
	}
	if *strict {
		os.Exit(1)
	}
}

func reviewRequirements() []requirement {
	return []requirement{
		{
			path: "internal/projectionintelligence/projectionfreshness/model.go",
			fragments: []string{
				`Version                  = "projection-pattern-freshness-guard-v3"`,
				`FingerprintVersion       = "projection-pattern-freshness-fingerprint-v3"`,
				`ComponentNewestAge`,
				`ComponentMeanAge`,
				`ComponentOldestAge`,
				`ComponentRecentSupport`,
				`var canonicalComponentNames = []ComponentName{`,
				`SelectionStatus projectionneighbors.Status`,
				`PatternStatus   projectionpatternconfidence.Status`,
				`PatternUsable   bool`,
				`Policy          Policy`,
				`SourceSelectionFingerprint string`,
				`SourcePatternFingerprint   string`,
				`result.Policy.Validate()`,
				`validateUpstreamSnapshot(result)`,
				`validateResultComponents(result)`,
				`validateSelectedTrajectoryIDs(result.SelectedTrajectoryIDs)`,
				`validateLimitations(result.Limitations)`,
				`validateDecisionSemantics(result)`,
				`pattern freshness component %q does not match aggregate evidence`,
				`pattern freshness component %q weight does not match policy`,
				`component.Weight <= 0`,
				`pattern freshness score does not match weighted components`,
				`pattern freshness limitations do not match policy evidence`,
				`selected trajectory identifiers must be sorted`,
				`pattern freshness limitations must be sorted and unique`,
			},
			forbidden: []string{
				`Version                  = "projection-pattern-freshness-guard-v1"`,
				`Version                  = "projection-pattern-freshness-guard-v2"`,
				`FingerprintVersion       = "projection-pattern-freshness-fingerprint-v1"`,
				`FingerprintVersion       = "projection-pattern-freshness-fingerprint-v2"`,
				`component.Weight < 0`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionfreshness/config.go",
			fragments: []string{
				`freshness age thresholds must satisfy newest <= mean <= oldest`,
				`freshness score thresholds must satisfy zero < minimum <= complete <= one`,
				`freshness component weights must be finite, positive, and sum to one`,
				`config.MaximumNewestNeighborAge > config.MaximumMeanNeighborAge`,
				`config.MaximumMeanNeighborAge > config.MaximumOldestNeighborAge`,
				`config.RecentNeighborAgeLimit > config.MaximumOldestNeighborAge`,
				`positiveUnitInterval(config.MinimumUsableScore)`,
				`positiveUnitInterval(config.CompleteScoreMinimum)`,
				`if !finite(weight) || weight <= 0 {`,
				`math.Abs(total-1) > scoreComparisonTolerance`,
			},
			forbidden: []string{
				`freshness component weights must be finite, non-negative, and sum to one`,
				`if !finite(weight) || weight < 0 {`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionfreshness/config_test.go",
			fragments: []string{
				`TestConfigValidateRejectsZeroComponentWeights`,
				`name: "newest age weight"`,
				`name: "mean age weight"`,
				`name: "oldest age weight"`,
				`name: "recent support weight"`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionfreshness/policy.go",
			fragments: []string{
				`type Policy struct {`,
				`MaximumNewestNeighborAge time.Duration`,
				`MinimumRecentNeighborCount int`,
				`MinimumUsableScore   float64`,
				`NewestAgeWeight     float64`,
				`func (config Config) policySnapshot() Policy`,
				`func (policy Policy) config() Config`,
				`func (policy Policy) Validate() error`,
				`return policy.config().Validate()`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionfreshness/evaluator.go",
			fragments: []string{
				`if err := selection.Validate(); err != nil {`,
				`if err := pattern.Validate(); err != nil {`,
				`if err := validateLineage(selection, pattern); err != nil {`,
				`pattern.SourceSelectionFingerprint != selection.InputFingerprint`,
				`pattern.SelectionStatus != selection.Status`,
				`sameSelectedTrajectoryIDs(selection, pattern)`,
				`SelectionStatus: selection.Status`,
				`PatternStatus:   pattern.Status`,
				`PatternUsable:   pattern.Usable`,
				`Policy:          config.policySnapshot()`,
				`SourceSelectionFingerprint: selection.InputFingerprint`,
				`SourcePatternFingerprint:   pattern.InputFingerprint`,
				`if err := result.Validate(); err != nil {`,
			},
			forbidden: []string{
				`pattern.Status != projectionpatternconfidence.StatusComplete && usable`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionfreshness/measurement.go",
			fragments: []string{
				`age := asOfTime.Sub(neighbor.CandidateEndTime.UTC())`,
				`if age <= config.RecentNeighborAgeLimit {`,
				`sort.Slice(metrics.ages`,
				`metrics.meanAge = meanDuration(metrics.ages)`,
				`count := int64(len(values))`,
				`quotient += nanoseconds / count`,
				`remainder += nanoseconds % count`,
				`return time.Duration(quotient + remainder/count)`,
			},
			forbidden: []string{
				`total += age.Nanoseconds()`,
				`total += value.Nanoseconds()`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionfreshness/components.go",
			fragments: []string{
				`Name: ComponentNewestAge`,
				`Name: ComponentMeanAge`,
				`Name: ComponentOldestAge`,
				`Name: ComponentRecentSupport`,
				`score += component.Score * component.Weight`,
				`return clampUnit(score)`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionfreshness/decision.go",
			fragments: []string{
				`selectionStatus projectionneighbors.Status`,
				`patternStatus projectionpatternconfidence.Status`,
				`patternUsable bool`,
				`historical_neighbors_unavailable`,
				`pattern_confidence_unusable`,
				`newest_historical_neighbor_too_old`,
				`mean_historical_neighbor_age_too_old`,
				`oldest_historical_neighbor_too_old`,
				`recent_historical_neighbor_support_insufficient`,
				`pattern_freshness_score_below_minimum`,
				`pattern_freshness_limited`,
				`pattern_confidence_not_complete`,
				`neighbor_selection_not_complete`,
				`limitations = normalizeNotices(limitations)`,
				`return freshnessDecision{`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionfreshness/fingerprint.go",
			fragments: []string{
				`writeFingerprintString(digest, FingerprintVersion)`,
				`writeFingerprintString(digest, selection.InputFingerprint)`,
				`writeFingerprintString(digest, string(selection.Status))`,
				`writeFingerprintString(digest, pattern.InputFingerprint)`,
				`writeFingerprintString(digest, pattern.SourceSelectionFingerprint)`,
				`writeFingerprintString(digest, string(pattern.SelectionStatus))`,
				`writeFingerprintString(digest, string(pattern.Status))`,
				`writeFingerprintBool(digest, pattern.Usable)`,
				`writeFingerprintTime(digest, selection.AsOfTime)`,
				`writeFingerprintFloat(digest, config.MinimumUsableScore)`,
				`writeFingerprintFloat(digest, config.CompleteScoreMinimum)`,
				`selection.AsOfTime.UTC().Sub(neighbor.CandidateEndTime.UTC())`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionfreshness/hardening_test.go",
			fragments: []string{
				`TestEvaluateBlocksUnusablePatternConfidence`,
				`TestEvaluateRejectsSourceSelectionFingerprintMismatch`,
				`TestEvaluateFingerprintIncludesUpstreamStatus`,
				`TestMeanDurationAvoidsInt64Overflow`,
				`TestEvaluateRejectsInconsistentCandidateAgeEvidence`,
				`TestEvaluateReportsAllHardViolations`,
				`TestConfigValidateRejectsThresholdOrderAndZeroScores`,
				`TestResultValidateRejectsCrossFieldMutations`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionfreshness/policy_integrity_test.go",
			fragments: []string{
				`TestEvaluatePublishesPolicyAndUpstreamSnapshot`,
				`TestEvaluateRejectsPatternSelectionStatusMismatch`,
				`TestResultValidateReconstructsPolicyDecision`,
				`TestResultValidateAcceptsReconstructedLimitedDecision`,
				`TestResultValidateRejectsZeroComponentWeight`,
				`Validate() accepted a zero component weight`,
				`coordinated component and score`,
				`decision and limitation`,
				`source pattern fingerprint`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/fixtures_test.go",
			fragments: []string{
				`validProductionBlockedFreshness`,
				`validProductionLimitedFreshness`,
				`evaluateProductionFreshnessFixture`,
				`projectionfreshness.New(config)`,
				`evaluator.Evaluate(selection, pattern)`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/freshness_fixture_contract_test.go",
			fragments: []string{
				`TestProductionFreshnessFixtureMatchesHardenedContracts`,
				`freshness source selection fingerprint`,
				`freshness source pattern fingerprint`,
				`freshness policy fixture validation error`,
				`freshness fixture validation error`,
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			fragments: []string{
				"## 31. Projection Freshness Review Hardening",
				"PROJECTION_FRESHNESS_VERSION=projection-pattern-freshness-guard-v3",
				"PROJECTION_FRESHNESS_FINGERPRINT_VERSION=projection-pattern-freshness-fingerprint-v3",
				"PROJECTION_FRESHNESS_PATTERN_USABILITY_GUARD=ENFORCED",
				"PROJECTION_FRESHNESS_EXACT_SELECTION_LINEAGE=ENFORCED",
				"PROJECTION_FRESHNESS_TIMESTAMP_DERIVED_AGE=ENFORCED",
				"PROJECTION_FRESHNESS_OVERFLOW_SAFE_MEAN=ENFORCED",
				"PROJECTION_FRESHNESS_POSITIVE_COMPONENT_WEIGHTS=ENFORCED",
				"PROJECTION_FRESHNESS_POLICY_SNAPSHOT=ENFORCED",
				"PROJECTION_FRESHNESS_DECISION_RECONSTRUCTION=ENFORCED",
				"PROJECTION_FRESHNESS_PERMANENT_AUDIT_COMMIT=619e24878a5025decf6fe21abddba537ce195560",
				"PROJECTION_FRESHNESS_PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30523502590",
				"PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_COMMIT=e3e99758d6f654db12ccce32ec55ad1339fb518f",
				"PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_GITHUB_ACTIONS_RUN=30527541240",
				"PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_BACKEND_QUALITY_JOB=90821894564",
				"PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_BACKEND_RACE_SAFETY_JOB=90821894465",
				"PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_POSTGRESQL_16_INTEGRATION_JOB=90821894536",
				"PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_BACKEND_CONTAINER_JOB=90822090361",
				"PROJECTION_FRESHNESS_WEIGHT_POLICY_CONSISTENCY=CLOSED",
				"PROJECTION_FRESHNESS_REVIEW_STATUS=CLOSED",
			},
			forbidden: []string{
				"PROJECTION_FRESHNESS_POSITIVE_COMPONENT_WEIGHTS=IMPLEMENTED_PENDING_EXACT_CI",
				"PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_COMMIT=PENDING",
				"PROJECTION_FRESHNESS_WEIGHT_POLICY_CONSISTENCY=IMPLEMENTED_PENDING_EXACT_CI",
				"PROJECTION_FRESHNESS_REVIEW_STATUS=OPEN",
			},
		},
		{
			path: "../../docs/139_PROJECTION_FRESHNESS_REVIEW_HARDENING.md",
			fragments: []string{
				"# Projection Freshness Review Hardening",
				"Status: closed",
				"LINEAGE_AGE_INTEGRITY_COMMIT=0b47aa3231c93d573a6026651a4085d376a40583",
				"POLICY_DECISION_INTEGRITY_COMMIT=072d0eb349fcd0e42c1d3c0bcf54c51cefb08a19",
				"PERMANENT_AUDIT_COMMIT=619e24878a5025decf6fe21abddba537ce195560",
				"PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30523502590",
				"WEIGHT_POLICY_CORRECTION_COMMIT=e3e99758d6f654db12ccce32ec55ad1339fb518f",
				"WEIGHT_POLICY_CORRECTION_GITHUB_ACTIONS_RUN=30527541240",
				"WEIGHT_POLICY_CORRECTION_BACKEND_QUALITY_JOB=90821894564",
				"WEIGHT_POLICY_CORRECTION_BACKEND_RACE_SAFETY_JOB=90821894465",
				"WEIGHT_POLICY_CORRECTION_POSTGRESQL_16_INTEGRATION_JOB=90821894536",
				"WEIGHT_POLICY_CORRECTION_BACKEND_CONTAINER_JOB=90822090361",
				"WEIGHT_POLICY_CONSISTENCY=CLOSED",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
				"PROJECTION_FRESHNESS_ENGINEERING_IMPLEMENTATION=COMPLETE",
				"PROJECTION_FRESHNESS_ENGINEERING_DEBT=CLOSED",
				"PROJECTION_FRESHNESS_ADDITIONAL_CODE_FIXES_REQUIRED=NO",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO",
				"PROJECTION_FRESHNESS_REVIEW_STATUS=CLOSED",
				"Strictly positive component-weight correction",
				"four fixed components as a versioned domain schema",
				"mandatory integer basis points",
				"Run projection freshness review audit",
			},
			forbidden: []string{
				"Status: open",
				"WEIGHT_POLICY_CORRECTION_COMMIT=PENDING",
				"WEIGHT_POLICY_CONSISTENCY=IMPLEMENTED_PENDING_EXACT_CI",
				"OPEN_CONFIRMED_FINDINGS=1",
				"PROJECTION_FRESHNESS_ENGINEERING_DEBT=OPEN",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES",
				"PROJECTION_FRESHNESS_REVIEW_STATUS=OPEN",
			},
		},
		{
			path: "../../docs/DOCUMENT_INDEX.md",
			fragments: []string{
				"## Document 139 — Projection Freshness Review Hardening",
				"139_PROJECTION_FRESHNESS_REVIEW_HARDENING.md",
				"timestamp-derived selected-neighbor age evidence",
				"strictly positive component weights",
				"policy and upstream-state snapshots",
				"e3e99758d6f654db12ccce32ec55ad1339fb518f",
				"30527541240",
				"formal reclosure with zero open, unclassified, or deferred",
			},
			forbidden: []string{
				"component-weight correction awaits exact Continuous Integration",
				"temporarily reopened",
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				`- name: Run projection freshness review audit`,
				`run: go run ./tools/projectionfreshnessreviewaudit -strict`,
			},
		},
	}
}

func inspectRequirements(root string, requirements []requirement) []string {
	failures := make([]string, 0)
	for _, item := range requirements {
		path := filepath.Clean(filepath.Join(root, filepath.FromSlash(item.path)))
		content, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("read %s: %v", item.path, err))
			continue
		}
		text := string(content)
		for _, fragment := range item.fragments {
			if !strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s is missing required fragment %q", item.path, fragment),
				)
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s contains forbidden fragment %q", item.path, fragment),
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
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for current := workingDirectory; ; current = filepath.Dir(current) {
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
	}
	return "", fmt.Errorf("apps/api module root was not found from %s", workingDirectory)
}

func validateAPIRoot(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve API root %q: %w", path, err)
	}
	info, err := os.Stat(filepath.Join(absolute, "go.mod"))
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("%s is not an apps/api module root", absolute)
	}
	return absolute, nil
}
