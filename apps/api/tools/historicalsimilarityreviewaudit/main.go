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
		"fail when a Historical Similarity review contract is absent",
	)
	flag.Parse()

	requirements := []requirement{
		{
			path: "internal/historicalintelligence/historicalsimilarity/model.go",
			fragments: []string{
				"historical-trajectory-similarity-v2",
				"type EvidenceConfidence struct",
				"type EvidenceQuality struct",
				"Score and Level represent route-shape similarity only",
				"MaximumSampleCount",
				"MaximumInputPointCount",
			},
		},
		{
			path: "internal/historicalintelligence/historicalsimilarity/config.go",
			fragments: []string{
				"GeometryScoreScaleKM",
				"EndpointScoreScaleKM",
				"config.SampleCount > MaximumSampleCount",
				"No panic path is necessary",
			},
			forbidden: []string{
				"MaximumMeanDistanceKM",
				"MaximumEndpointDistanceKM",
				"panic(",
			},
		},
		{
			path: "internal/historicalintelligence/historicalsimilarity/engine.go",
			fragments: []string{
				"engine.prepare(reference)",
				"engine.prepare(candidate)",
				"buildResult",
				"result.Validate",
			},
			forbidden: []string{
				"func (engine *Engine) Rank",
				"continue",
			},
		},
		{
			path: "internal/historicalintelligence/historicalsimilarity/prepare.go",
			fragments: []string{
				"canonicalPoints",
				"equal_timestamp_points_canonicalized",
				"near_antipodal_interpolation_fallback",
				"MaximumInputPointCount",
				"assessEvidenceQuality",
			},
		},
		{
			path: "internal/historicalintelligence/historicalsimilarity/quality.go",
			fragments: []string{
				"DeclaredQualityScore",
				"SegmentQualityScore",
				"CoverageContinuityScore",
				"ObservationCadenceScore",
				"PointRetentionScore",
				"trajectory_coverage_gaps_present",
				"trajectory_non_observed_segments_present",
				"comparison_confidence_uses_weaker_trajectory_evidence",
			},
		},
		{
			path: "internal/historicalintelligence/historicalsimilarity/scoring.go",
			fragments: []string{
				"math.Max(",
				"endpointObservedKM",
				"relativeDifference",
				"worst_endpoint_proximity",
			},
			forbidden: []string{
				"(startDistance +",
				"math.Max(left, right),\n\t\t1",
			},
		},
		{
			path: "internal/historicalintelligence/historicalsimilarity/geodesy.go",
			fragments: []string{
				"interpolateGreatCircle",
				"usedAntipodalFallback",
				"sampleDistances",
				"compensation",
			},
		},
		{
			path: "internal/historicalintelligence/historicalsimilarity/fingerprint.go",
			fragments: []string{
				"FingerprintVersion",
				"math.Float64bits",
				"appendPreparedFingerprint",
				"reference",
				"candidate",
			},
			forbidden: []string{
				"%.12f",
				"sort.Strings(records)",
				"trajectory.FlightTrajectory",
			},
		},
		{
			path: "internal/historicalintelligence/historicalsimilarity/validation.go",
			fragments: []string{
				"validateResultComponents",
				"expectedComponents",
				"weighted similarity score",
				"mean distance exceeds maximum distance",
				"component mathematics",
				"validateResultConfidence",
				"trajectory evidence quality mathematics",
			},
		},
		{
			path: "internal/historicalintelligence/historicalsimilarity/engine_test.go",
			fragments: []string{
				"TestCompareSeparatesSimilarityFromEvidenceConfidence",
				"TestConfigRejectsExcessiveSampleCount",
				"TestCompareRejectsExcessiveInputPoints",
				"TestCompareCanonicalizesEqualTimestamps",
				"TestFingerprintUsesExactFloatBits",
				"TestEndpointComponentUsesWorstEndpoint",
				"TestRelativeDifferenceUsesExactZeroScalePolicy",
				"TestCompareHandlesDateLineAndPolarRoutes",
				"TestResultValidateRejectsUnknownComponent",
				"TestResultValidateRejectsWeightedScoreMismatch",
				"TestResultValidateRejectsConfidenceMismatch",
			},
		},
		{
			path: "../../docs/130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md",
			fragments: []string{
				"Status: implemented review remediation",
				"SIMILARITY_CONFIDENCE_SEPARATED=YES",
				"TRAJECTORY_QUALITY_EVIDENCE=BOUND",
				"SAMPLE_COUNT_MAXIMUM=ENFORCED",
				"PUBLIC_RANK_API=REMOVED",
				"FINGERPRINT_PREPARED_EXACT=ENFORCED",
				"EQUAL_TIMESTAMP_CANONICALIZATION=ENFORCED",
				"RESULT_MATHEMATICAL_VALIDATION=ENFORCED",
				"ENDPOINT_SCORE_USES_WORST_ENDPOINT=YES",
				"RELATIVE_DIFFERENCE_ZERO_SCALE=EXACT",
				"GREAT_CIRCLE_RESAMPLING=ENFORCED",
				"HISTORICAL_SIMILARITY_ENGINEERING_REMEDIATION=IMPLEMENTED",
			},
		},
	}

	failures := make([]string, 0)
	for _, item := range requirements {
		content, err := os.ReadFile(
			filepath.Clean(item.path),
		)
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

	if len(failures) == 0 {
		fmt.Println(
			"Historical similarity review audit: PASS",
		)
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Historical similarity review audit: %s\n",
			failure,
		)
	}
	if *strict {
		os.Exit(1)
	}
}
