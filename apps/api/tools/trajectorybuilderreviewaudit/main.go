package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type rule struct {
	path      string
	required  []string
	forbidden []string
}

func main() {
	root := flag.String("root", "", "repository root")
	strict := flag.Bool("strict", true, "fail on contract violation")
	flag.Parse()
	resolved, err := resolveRoot(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rules := []rule{
		{
			path: "apps/api/internal/features/trajectorybuilder/contracts.go",
			required: []string{
				`const Version = "trajectory-feature-builder-v2"`,
				"flightfeatures.TrajectoryRequiredFeatureFieldCount",
			},
		},
		{
			path: "apps/api/internal/features/trajectorybuilder/builder.go",
			required: []string{
				"return flightfeatures.TrajectoryFeatures{}, ErrContextRequired",
				"canonicalizeEvidence(ctx, item)",
				"calculateSamplingMetrics(ctx, evidence.points)",
				"calculateCoverageRatio(ctx, evidence, segmentSummary)",
				"calculatePathEfficiency(ctx, evidence)",
				"availableFieldCount := 0",
				"evidence.supportingPointCount",
			},
			forbidden: []string{
				"ctx = context.Background()",
				"availableFieldCount := 11",
				"pointCount := len(item.Points)",
			},
		},
		{
			path: "apps/api/internal/features/trajectorybuilder/evidence.go",
			required: []string{
				"func canonicalizeEvidence(",
				"resolveTrajectoryWindow(item)",
				"sort.SliceStable(observed",
				"TrajectoryLimitationDuplicateTimestampsCollapsed",
				"TrajectoryLimitationPointRecordsUnmaterialized",
				"result.supportingPointCount = item.PointCount",
			},
		},
		{
			path: "apps/api/internal/features/trajectorybuilder/sampling.go",
			required: []string{
				"points []canonicalPoint",
				"kahanAccumulator{}",
				"TrajectoryLimitationSamplingEvidenceInsufficient",
			},
			forbidden: []string{
				"sort.SliceStable",
				"intervalSeconds := timestamps[index]",
			},
		},
		{
			path: "apps/api/internal/features/trajectorybuilder/coverage.go",
			required: []string{
				"TrajectoryLimitationCoverageObservationEvidenceUnavailable",
				"flightfeatures.TemporalDurationSeconds(gapStart, gapEnd)",
				"unionDurationContext(ctx, intervals)",
				"windowDuration == 0",
			},
			forbidden: []string{
				"gap.DurationSeconds != 0 &&",
			},
		},
		{
			path: "apps/api/internal/features/trajectorybuilder/path.go",
			required: []string{
				"gapSeparatesInterval(",
				"TrajectoryLimitationPathDiscontinuityExcluded",
				"continuous := hasPrevious",
				"current.coordinates = append(current.coordinates, startCoordinate, endCoordinate)",
				"ratio > 1+pathRatioTolerance",
			},
			forbidden: []string{
				"segmentCoordinates = append",
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/trajectory_policy.go",
			required: []string{
				"canonical-unique-points-or-persisted-metadata-fallback-v1",
				"unique-chronological-observation-instants-kahan-v1",
				"non-invalid-segment-evidence-plus-clipped-gap-union-v1",
				"continuous-parts-no-gap-or-segment-discontinuity-bridging-v1",
			},
		},
		{
			path: "apps/api/internal/features/extractor/quality.go",
			required: []string{
				"supportingPointCount := 0",
				"evidence.SupportingPointCount > supportingPointCount",
			},
			forbidden: []string{
				"supportingPointCount := item.PointCount",
				"len(item.Points) > supportingPointCount",
			},
		},
		{
			path: "apps/api/internal/features/validator/rules.go",
			required: []string{
				"trajectoryPathEfficiencyComparable(item.Evidence)",
			},
		},
		{
			path: "apps/api/internal/features/validator/trajectory_policy.go",
			required: []string{
				"func trajectoryPathEfficiencyComparable(",
				"TrajectoryLimitationPathDiscontinuityExcluded",
				"TrajectoryLimitationPathSegmentFallback",
				"TrajectoryLimitationPathEvidenceInsufficient",
				"TrajectoryLimitationDuplicateTimestampsCollapsed",
			},
		},
		{
			path: "apps/api/internal/features/trajectorybuilder/trajectorybuilder_review_hardening_test.go",
			required: []string{
				"TestBuilderFiltersWindowAndCollapsesDuplicateTimestamps",
				"TestCanonicalMetricsArePermutationInvariant",
				"TestBuilderUsesPersistedPointCountWhenRecordsAreUnmaterialized",
				"TestBuilderObservesCancellationDuringLargeCanonicalization",
			},
		},
		{
			path:     "apps/api/internal/features/featurepipeline/contracts.go",
			required: []string{`flight-feature-processing-pipeline-v11`},
		},
		{
			path:     "apps/api/internal/features/validator/contracts.go",
			required: []string{`flight-feature-validator-v5`},
		},
		{
			path: "docs/122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md",
			required: []string{
				"TRAJECTORY_CANONICAL_EVIDENCE=ENFORCED",
				"TRAJECTORY_DUPLICATE_TIMESTAMP_POLICY=UNIQUE_DETERMINISTIC",
				"TRAJECTORY_PATH_GAP_BRIDGING=CLOSED",
				"TRAJECTORY_COVERAGE_REQUIRES_OBSERVATION_EVIDENCE=ENFORCED",
				"TRAJECTORY_BUILDER_PROCESSING_GENERATION=v11",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
				"TRAJECTORY_BUILDER_REVIEW_STATUS=CLOSED",
			},
		},
		{
			path: ".github/workflows/backend-ci.yml",
			required: []string{
				"Run trajectory builder review audit",
				"go run ./tools/trajectorybuilderreviewaudit -strict",
			},
		},
	}

	failures := make([]string, 0)
	for _, item := range rules {
		content, readErr := os.ReadFile(filepath.Join(resolved, item.path))
		if readErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.path, readErr))
			continue
		}
		text := string(content)
		for _, fragment := range item.required {
			if !strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s missing %q", item.path, fragment))
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s contains forbidden %q", item.path, fragment))
			}
		}
	}

	if len(failures) == 0 {
		fmt.Println("Trajectory builder review audit: PASS")
		return
	}
	fmt.Fprintln(os.Stderr, "Trajectory builder review audit: FAIL")
	for _, failure := range failures {
		fmt.Fprintln(os.Stderr, "-", failure)
	}
	if *strict {
		os.Exit(1)
	}
}

func resolveRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return filepath.Abs(strings.TrimSpace(explicit))
	}
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, statErr := os.Stat(filepath.Join(current, "apps/api/go.mod")); statErr == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root was not found")
		}
		current = parent
	}
}
