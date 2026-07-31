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
	strict := flag.Bool("strict", false, "fail when a Projection Evaluation review requirement is absent")
	root := flag.String("root", "", "optional path to the apps/api module root")
	flag.Parse()

	apiRoot, err := resolveAPIRoot(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Projection evaluation review audit: %v\n", err)
		if *strict {
			os.Exit(1)
		}
		return
	}
	failures := inspectRequirements(apiRoot, reviewRequirements())
	if len(failures) == 0 {
		fmt.Println("Projection evaluation review audit: PASS")
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(os.Stderr, "Projection evaluation review audit: %s\n", failure)
	}
	if *strict {
		os.Exit(1)
	}
}

func reviewRequirements() []requirement {
	return []requirement{
		{
			path: "internal/projectionintelligence/projectionevaluation/model.go",
			fragments: []string{
				`Version                     = "projection-replay-evaluation-v2"`,
				`FingerprintVersion          = "projection-replay-evaluation-fingerprint-v2"`,
				`ProjectionSnapshotVersion   = "projection-replay-projection-snapshot-v2"`,
				`TruthSnapshotVersion        = "projection-replay-truth-snapshot-v2"`,
				`AggregateVersion            = "projection-replay-aggregate-v2"`,
				`TruthKnowledgeCutoffMode    = "point_availability_evidence"`,
				`type TruthAvailability struct`,
				`type EvaluationPolicy struct`,
				`type EndpointMetrics struct`,
				`type ConfidenceMetrics struct`,
				`type LeadTimeMetrics struct`,
				`ArrivalPredictionRecall`,
				`TrajectoryMacroMeanHorizontalErrorM`,
			},
			forbidden: []string{
				`projection-replay-evaluation-v1`,
				`projection-replay-aggregate-v1`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/config.go",
			fragments: []string{
				`EvaluationPolicyVersion = "projection-replay-evaluation-policy-v2"`,
				`MaximumTruthGroundSpeedMPS`,
				`MaximumTruthVerticalRateMPS`,
				`LeadTimeBucketSize`,
				`evaluationPolicyFingerprint(policy)`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/truth.go",
			fragments: []string{
				`ErrTruthAvailabilityEvidenceMissing`,
				`ErrAmbiguousTruthTimestamp`,
				`evidence.AvailableAt.After(evaluatedAt)`,
				`sameTruthContent(`,
				`truthMatchImplausibleMovement`,
				`distanceM/seconds > config.MaximumTruthGroundSpeedMPS`,
				`verticalRate > config.MaximumTruthVerticalRateMPS`,
			},
			forbidden: []string{
				`indexed[left].index <`,
				`result[len(result)-1] = indexedPoint.point`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/fingerprint.go",
			fragments: []string{
				`func projectionSnapshotFingerprint(`,
				`string(projection.Method.DecisionClass)`,
				`point.Position.Latitude`,
				`point.Uncertainty.HorizontalRadiusM`,
				`writeConfidence(digest, point.Confidence)`,
				`func truthSnapshotFingerprint(`,
				`string(point.GeometricAltitudeStatus)`,
				`string(point.BarometricAltitudeStatus)`,
				`item.availableAt`,
				`func aggregateFingerprint(results []Result)`,
			},
			forbidden: []string{
				`generatedAt time.Time`,
				`writeFingerprintTime(\n\t\tdigest,\n\t\tgeneratedAt`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/evaluator.go",
			fragments: []string{
				`TruthAvailability []TruthAvailability`,
				`normalizeTruthPoints(`,
				`buildEndpointMetrics(`,
				`buildConfidenceMetrics(`,
				`buildLeadTimeMetrics(`,
				`projectionSnapshotFingerprint(`,
				`truthSnapshotFingerprint(`,
				`truth_after_availability_cutoff_excluded`,
				`implausible_truth_interpolation_rejected`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/validation.go",
			fragments: []string{
				`expectedError := greatCircleDistanceM(`,
				`expectedRatio := expectedError / result.Policy.MaximumHorizontalErrorM`,
				`samePositionMetrics(`,
				`sameEndpointMetrics(`,
				`sameConfidenceMetrics(`,
				`sameLeadTimeMetrics(`,
				`arrival derived metrics are inconsistent`,
				`airportICAOPattern.MatchString`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/metrics.go",
			fragments: []string{
				`func median(values []float64) float64`,
				`return (sortedValues[middle-1] + sortedValues[middle]) / 2`,
				`func buildEndpointMetrics(`,
				`func buildConfidenceMetrics(`,
				`func buildLeadTimeMetrics(`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/aggregate.go",
			fragments: []string{
				`func Aggregate(results []Result, generatedAt time.Time)`,
				`evaluationGroupKey(evaluation)`,
				`aggregateFingerprint(results)`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/aggregate_accumulator.go",
			fragments: []string{
				`if evaluation.Status == StatusUnavailable {`,
				`TrajectoryMacroMeanHorizontalErrorM`,
				`ArrivalPredictionRecall`,
				`ArrivalAirportAccuracy`,
				`leadTimeSummaries()`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/aggregate_identity.go",
			fragments: []string{
				`string(result.ProjectionMethod.DecisionClass)`,
				`result.ProjectionHorizonEndTime.Sub(result.ProjectionAsOfTime)`,
				`fmt.Sprintf("%d", result.ForecastStep)`,
				`result.Policy.InputFingerprint`,
			},
			forbidden: []string{
				`func methodKey(`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/review_hardening_test.go",
			fragments: []string{
				`TestEvaluationFingerprintBindsProjectionOutput`,
				`TestEvaluationFingerprintBindsAltitudeStatuses`,
				`TestEvaluationIsOrderIndependentForIdenticalTruthDuplicates`,
				`TestEvaluateRejectsImplausibleTruthInterpolation`,
				`TestResultValidationRejectsTamperedDerivedMetrics`,
				`TestMedianUsesAverageOfEvenMiddleValues`,
				`TestAggregateSeparatesDecisionClassAndPolicy`,
				`TestAggregateExcludesUnavailableEvaluationFromAccuracyMetrics`,
				`TestAggregatePublishesArrivalSelectivePredictionAccounting`,
				`TestAggregateFingerprintExcludesGeneratedAt`,
			},
		},
		{
			path: "internal/projectionintelligence/projectionevaluation/truth_test.go",
			fragments: []string{
				`TestNormalizeTruthPointsRejectsConflictingDuplicateTimestamps`,
				`TestNormalizeTruthPointsIsOrderIndependentForIdenticalDuplicates`,
				`TestNormalizeTruthPointsRequiresAvailabilityEvidence`,
				`TestTruthAtRejectsImplausibleMovement`,
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				`- name: Run projection evaluation review audit`,
				`go run ./tools/projectionevaluationreviewaudit -strict`,
			},
		},
		{
			path: "../../docs/143_PROJECTION_EVALUATION_REVIEW_HARDENING.md",
			fragments: []string{
				`Status: engineering implementation complete, exact Continuous Integration and formal closure pending`,
				`REVIEW_BASELINE_COMMIT=61e1696b16e39f49a3850530312555c3593acfc5`,
				`REPLAY_KNOWLEDGE_CUTOFF=IMPLEMENTED_PENDING_EXACT_CI`,
				`AGGREGATION_IDENTITY=IMPLEMENTED_PENDING_EXACT_CI`,
				`UNAVAILABLE_ACCURACY_ISOLATION=IMPLEMENTED_PENDING_EXACT_CI`,
				`ARRIVAL_SELECTIVE_PREDICTION_ACCOUNTING=IMPLEMENTED_PENDING_EXACT_CI`,
				`OPEN_CONFIRMED_FINDINGS=0_PENDING_EXACT_CI`,
				`PROJECTION_EVALUATION_REVIEW_STATUS=OPEN_PENDING_EXACT_CI_AND_FORMAL_CLOSURE`,
			},
		},
		{
			path: "../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md",
			fragments: []string{
				`## 37. Projection Evaluation Replay and Aggregation Integrity`,
				`PROJECTION_EVALUATION_VERSION=projection-replay-evaluation-v2`,
				`PROJECTION_EVALUATION_TRUTH_KNOWLEDGE_CUTOFF=POINT_AVAILABILITY_EVIDENCE`,
				`PROJECTION_EVALUATION_AGGREGATION_IDENTITY=METHOD_VERSION_CLASS_HORIZON_STEP_POLICY`,
				`PROJECTION_EVALUATION_REVIEW_STATUS=OPEN_PENDING_EXACT_CI_AND_FORMAL_CLOSURE`,
			},
		},
		{
			path: "../../docs/DOCUMENT_INDEX.md",
			fragments: []string{
				`## Document 143 — Projection Evaluation Review Hardening`,
				`strict event-time and system-availability replay cutoffs`,
				`unavailable-result accuracy isolation`,
			},
		},
	}
}

func inspectRequirements(apiRoot string, requirements []requirement) []string {
	failures := make([]string, 0)
	for _, requirement := range requirements {
		path := filepath.Clean(filepath.Join(apiRoot, requirement.path))
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("read %s: %v", requirement.path, err))
			continue
		}
		content := string(contentBytes)
		for _, fragment := range requirement.fragments {
			if !strings.Contains(content, fragment) {
				failures = append(failures, fmt.Sprintf("%s is missing %q", requirement.path, fragment))
			}
		}
		for _, fragment := range requirement.forbidden {
			if strings.Contains(content, fragment) {
				failures = append(failures, fmt.Sprintf("%s contains forbidden %q", requirement.path, fragment))
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
		if root, err := validateAPIRoot(current); err == nil {
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

func validateAPIRoot(candidate string) (string, error) {
	absolute, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(absolute, "go.mod")); err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(absolute, "internal", "projectionintelligence", "projectionevaluation")); err != nil {
		return "", err
	}
	return absolute, nil
}
