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
		"fail when a Projection Pattern Confidence review requirement is absent",
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
			"Projection pattern confidence review audit: %v\n",
			err,
		)
		if *strict {
			os.Exit(1)
		}
		return
	}

	failures := inspectRequirements(apiRoot, reviewRequirements())
	if len(failures) == 0 {
		fmt.Println("Projection pattern confidence review audit: PASS")
		return
	}

	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Projection pattern confidence review audit: %s\n",
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
			path: "internal/projectionintelligence/projectionpatternconfidence/model.go",
			fragments: []string{
				`Version            = "projection-pattern-confidence-v4"`,
				`FingerprintVersion = "projection-pattern-confidence-fingerprint-v4"`,
				`ComponentSimilarityStrength`,
				`ComponentSupport`,
				`ComponentSimilarityConsistency`,
				`ComponentAnchorProximity`,
				`ComponentContinuationAgreement`,
				`var canonicalComponentNames = []ComponentName{`,
				`ContinuationAgreementKnown       bool`,
				`MeanCandidateAgeSeconds float64`,
				`validateContinuationAgreement(result)`,
				`validateDecisionSemantics(result)`,
			},
			forbidden: []string{
				`Version            = "projection-pattern-confidence-v1"`,
				`Version            = "projection-pattern-confidence-v2"`,
				`Version            = "projection-pattern-confidence-v3"`,
				`FingerprintVersion = "projection-pattern-confidence-fingerprint-v1"`,
				`FingerprintVersion = "projection-pattern-confidence-fingerprint-v2"`,
				`FingerprintVersion = "projection-pattern-confidence-fingerprint-v3"`,
				`ComponentFreshness ComponentName = "freshness"`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionpatternconfidence/config.go",
			fragments: []string{
				`minimum neighbor count must be at least two`,
				`minimum usable score must be finite, greater than zero, and at most one`,
				`pattern confidence component weights must be finite, positive, and sum to one`,
				`Deprecated: ignored. Freshness belongs exclusively to projectionfreshness.`,
				`if normalized.MinimumNeighborCount < 2 {`,
				`if !positiveUnitInterval(normalized.MinimumUsableScore) {`,
				`normalized.ContinuationAgreementSampleCount < 1`,
				`normalized.MaximumContinuationDivergenceMPS <`,
				`normalized.ContinuationDivergenceNormalizationMPS`,
				`normalized.SimilarityStrengthWeight,`,
				`normalized.SupportWeight,`,
				`normalized.SimilarityConsistencyWeight,`,
				`normalized.AnchorProximityWeight,`,
				`normalized.ContinuationAgreementWeight,`,
				`if !positiveFinite(weight) {`,
				`if math.Abs(total-1) > scoreComparisonTolerance {`,
			},
			forbidden: []string{
				`normalized.MaximumCandidateAge`,
				`MaximumCandidateAge.Seconds()`,
				`freshnessScore`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionpatternconfidence/evidence.go",
			fragments: []string{
				`similarityInputFingerprint string`,
				`minimumSimilarityScore`,
				`similarityStandardDeviation`,
				`continuationAgreementScore`,
				`math.Sqrt(variance / divisor)`,
				`1 - evidence.similarityStandardDeviation/maximumUnitIntervalStandardDeviation`,
				`if continuation.known {`,
			},
			forbidden: []string{
				`candidateAge`,
				`freshnessScore`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionpatternconfidence/components.go",
			fragments: []string{
				`Name:   ComponentSimilarityStrength`,
				`Name:   ComponentSupport`,
				`Name:   ComponentSimilarityConsistency`,
				`Name:   ComponentAnchorProximity`,
				`Name:   ComponentContinuationAgreement`,
				`score += component.Score * component.Weight`,
				`status == StatusLimited`,
				`return projectioncontract.ConfidenceLevelMedium`,
			},
			forbidden: []string{
				`Name:   ComponentFreshness`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionpatternconfidence/continuation_agreement.go",
			fragments: []string{
				`if len(selection.Neighbors) < 2 {`,
				`indexSelectedCandidates(selection, candidates)`,
				`config.ContinuationAgreementSampleCount`,
				`pairCount := len(trajectoryIDs) * (len(trajectoryIDs) - 1) / 2`,
				`comparisonCount := pairCount * config.ContinuationAgreementSampleCount`,
				`divergence := spread / leftVector.elapsedS`,
				`1 - meanDivergence/config.ContinuationDivergenceNormalizationMPS`,
				`anchorLatitude`,
				`endpointLatitude`,
				`eastM`,
				`northM`,
				`func interpolatePosition(`,
				`normalizeLongitudeDegrees`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionpatternconfidence/evaluator.go",
			fragments: []string{
				`Evaluate preserves the legacy interface but deliberately does not authorize`,
				`continuationAgreementEvidence{}`,
				`func (evaluator *Evaluator) EvaluateWithContinuations(`,
				`extractContinuationAgreement(`,
				`MeanCandidateAgeSeconds:     0,`,
				`if err := result.Validate(); err != nil {`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionpatternconfidence/decision.go",
			fragments: []string{
				`agreementAvailable := evidence.continuation.known`,
				`agreementAcceptable := agreementAvailable &&`,
				`config.MaximumContinuationDivergenceMPS`,
				`pattern_continuation_agreement_unavailable`,
				`pattern_continuation_divergence_above_maximum`,
				`supportSufficient &&`,
				`similarityFloorSufficient &&`,
				`dispersionAcceptable &&`,
				`agreementAvailable &&`,
				`agreementAcceptable &&`,
				`scoreSufficient`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionpatternconfidence/fingerprint.go",
			fragments: []string{
				`writeFingerprintString(digest, FingerprintVersion)`,
				`writeFingerprintString(digest, neighbor.trajectoryID)`,
				`writeFingerprintFloat(digest, neighbor.similarityScore)`,
				`writeFingerprintString(digest, neighbor.similarityInputFingerprint)`,
				`writeFingerprintFloat(digest, neighbor.anchorDistanceKM)`,
				`writeFingerprintBool(digest, evidence.continuation.known)`,
				`writeFingerprintFloat(digest, vector.anchorLatitude)`,
				`writeFingerprintFloat(digest, vector.endpointLatitude)`,
				`writeFingerprintFloat(digest, vector.eastM)`,
				`writeFingerprintFloat(digest, vector.northM)`,
				`writeFingerprintFloat(digest, config.ContinuationAgreementWeight)`,
			},
			forbidden: []string{
				`CandidateAge`,
				`MaximumCandidateAge`,
				`MeanCandidateAgeSeconds`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionpatternconfidence/validation.go",
			fragments: []string{
				`func validatePolicy(policy Policy) error`,
				`if !positiveFinite(weight) {`,
				`math.Abs(total-1) > scoreComparisonTolerance`,
				`func validateContinuationAgreement(result Result) error`,
				`expectedPairs := result.NeighborCount * (result.NeighborCount - 1) / 2`,
				`continuation spread and divergence measurements are inconsistent`,
				`pattern confidence requires the canonical component catalog`,
				`expectedScores := []float64{`,
				`expectedWeights := []float64{`,
				`pattern confidence score does not match weighted components`,
				`func validateDecisionSemantics(result Result) error`,
				`pattern confidence usable decision does not match policy evidence`,
				`pattern confidence status does not match policy evidence`,
				`pattern_continuation_agreement_unavailable`,
				`pattern_continuation_divergence_above_maximum`,
				`selected trajectory identifiers must be sorted`,
				`pattern confidence limitations must be sorted and unique`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionpatternconfidence/evaluator_test.go",
			fragments: []string{
				`TestEvaluateWithContinuationsProducesCompleteConfidence`,
				`TestEvaluateWithoutContinuationsCannotAuthorize`,
				`TestEvaluateWithContinuationsRejectsOpposingRoutes`,
				`TestEvaluateWithContinuationsRejectsIntermediateRouteConflict`,
				`TestEvaluateWithContinuationsChangesFingerprintForRouteMutation`,
				`TestEvaluateWithContinuationsIgnoresCandidateOrder`,
				`TestEvaluateWithContinuationsRejectsMissingCandidate`,
				`TestEvaluateWithContinuationsUsesInterpolatedSamples`,
				`TestEvaluateWithContinuationsRejectsWeakSimilarityFloor`,
				`TestEvaluateWithContinuationsRejectsSimilarityDispersion`,
				`TestEvaluateWithContinuationsIgnoresFreshnessEvidence`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionpatternconfidence/integrity_test.go",
			fragments: []string{
				`TestResultValidateRejectsUnknownComponentName`,
				`TestResultValidateRejectsComponentScoreMismatch`,
				`TestResultValidateRejectsPolicyMutation`,
				`TestResultValidateRejectsComponentWeightMutation`,
				`TestResultValidateRejectsUnknownAgreementForUsableResult`,
				`TestResultValidateRejectsInvalidPairCount`,
				`TestResultValidateRejectsDivergenceDecisionMismatch`,
				`TestResultValidateRejectsInconsistentSpreadEvidence`,
				`TestResultValidateRejectsMissingDecisionLimitation`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionpatternconfidence/config_test.go",
			fragments: []string{
				`TestConfigValidateAcceptsContinuationPolicy`,
				`TestConfigValidateMigratesVersionThreeWeights`,
				`TestConfigValidateRejectsInvalidValues`,
				`config.MinimumNeighborCount = 1`,
				`config.MinimumUsableScore = 0`,
				`config.ContinuationAgreementSampleCount = 33`,
				`config.SimilarityStrengthWeight = 0`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/config.go",
			fragments: []string{
				`type PatternConfidenceEvaluator interface {`,
				`EvaluateWithContinuations(`,
				`[]trajectory.FlightTrajectory,`,
			},
			forbidden: []string{
				"type PatternConfidenceEvaluator interface {\n\tEvaluate(",
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/pattern_evaluation.go",
			fragments: []string{
				`return evaluator.EvaluateWithContinuations(`,
				`selection,`,
				`candidates,`,
			},
			forbidden: []string{
				`continuationAwarePatternConfidenceEvaluator`,
				`evaluator.(`,
				`return evaluator.Evaluate(selection)`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionproduction/pattern_evaluation_test.go",
			fragments: []string{
				`var _ PatternConfidenceEvaluator = (*continuationEvaluatorProbe)(nil)`,
				`var _ PatternConfidenceEvaluator = (*projectionpatternconfidence.Evaluator)(nil)`,
				`TestEvaluatePatternConfidenceRequiresContinuationEvidence`,
				`probe.candidateCount != 1`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/config.go",
			fragments: []string{
				`type PatternConfidenceEvaluator interface {`,
				`EvaluateWithContinuations(`,
				`[]trajectory.FlightTrajectory,`,
			},
			forbidden: []string{
				"type PatternConfidenceEvaluator interface {\n\tEvaluate(",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/pattern_evaluation.go",
			fragments: []string{
				`return evaluator.EvaluateWithContinuations(`,
				`selection,`,
				`candidates,`,
			},
			forbidden: []string{
				`continuationAwarePatternConfidenceEvaluator`,
				`evaluator.(`,
				`return evaluator.Evaluate(selection)`,
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontinuation/pattern_evaluation_test.go",
			fragments: []string{
				`var _ PatternConfidenceEvaluator = (*continuationEvaluatorProbe)(nil)`,
				`var _ PatternConfidenceEvaluator = (*projectionpatternconfidence.Evaluator)(nil)`,
				`TestEvaluatePatternConfidenceRequiresContinuationEvidence`,
				`probe.candidateCount != 1`,
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				`- name: Run projection pattern confidence review audit`,
				`run: go run ./tools/projectionpatternconfidencereviewaudit -strict`,
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
