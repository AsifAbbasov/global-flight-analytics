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
			path: "apps/api/internal/features/temporalbuilder/contracts.go",
			required: []string{
				`const Version = "temporal-feature-builder-v2"`,
				"flightfeatures.TemporalRequiredFeatureFieldCount",
			},
		},
		{
			path: "apps/api/internal/features/temporalbuilder/errors.go",
			required: []string{
				"ErrContextRequired",
			},
		},
		{
			path: "apps/api/internal/features/temporalbuilder/builder.go",
			required: []string{
				"return flightfeatures.TemporalFeatures{}, ErrContextRequired",
				"flightfeatures.TemporalDurationSeconds(",
				"item.DurationSeconds != durationSeconds",
				"evaluateSegmentEvidence(",
				"temporal_segment_boundary_fallback",
				"temporal_invalid_segment_evidence",
				"trajectory_point_count_metadata_mismatch",
				"index%contextCheckInterval == 0",
			},
			forbidden: []string{
				"ctx = context.Background()",
				"endTime.Sub(startTime) / time.Second",
				"item.DurationSeconds != 0 &&",
				"One or more trajectory points",
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/temporal_policy.go",
			required: []string{
				"TemporalDurationRoundingPolicyTruncateFractionalSeconds",
				"CurrentTemporalDurationRoundingPolicy",
				"func TemporalDurationSeconds(",
				"endTime.Sub(startTime) / time.Second",
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/schema.go",
			required: []string{
				"fractional seconds truncated toward zero",
			},
		},
		{
			path: "apps/api/internal/features/validator/rules.go",
			required: []string{
				"expectedDuration := flightfeatures.TemporalDurationSeconds(",
			},
			forbidden: []string{
				"expectedDuration := int64(end.Sub(start) / time.Second)",
			},
		},
		{
			path: "apps/api/internal/features/extractor/snapshot_validation.go",
			required: []string{
				"point.ObservedAt.After(cutoff)",
				"segment.StartTime.After(cutoff)",
				"segment.EndTime.After(cutoff)",
				"gap.StartTime.After(cutoff)",
				"gap.EndTime.After(cutoff)",
			},
		},
		{
			path: "apps/api/internal/features/validator/quality_provenance_test.go",
			required: []string{
				"TestValidatorAcceptsSystemUpdateAfterAsOfTime",
			},
		},
		{
			path: "apps/api/cmd/materialize-flight-features/operation.go",
			required: []string{
				"asOfTime = item.EndTime.UTC()",
				"asOfTime.Before(item.EndTime.UTC())",
			},
		},
		{
			path: "apps/api/internal/features/temporalbuilder/temporalbuilder_review_hardening_test.go",
			required: []string{
				"TestBuilderRejectsNilContext",
				"TestBuilderUsesPersistedSegmentBoundariesWhenPointsAreAbsent",
				"TestBuilderReportsZeroDurationMetadataForNonZeroWindow",
				"TestBuilderUsesCentralFractionalSecondPolicy",
				"TestBuilderReportsPointCountMetadataMismatch",
				"TestBuilderObservesCancellationDuringPointScan",
				"TestBuilderLimitationMessagesContainExactCounts",
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/model.go",
			required: []string{
				`flight-feature-processing-pipeline-v11`,
			},
		},
		{
			path: "apps/api/internal/features/featurepipeline/contracts.go",
			required: []string{
				`flight-feature-processing-pipeline-v11`,
			},
		},
		{
			path: "docs/119_TEMPORAL_BUILDER_REVIEW_HARDENING.md",
			required: []string{
				"TEMPORAL_SEGMENT_BOUNDARY_FALLBACK=ENFORCED",
				"TEMPORAL_DURATION_METADATA_ZERO_SENTINEL=CLOSED",
				"TEMPORAL_DURATION_ROUNDING_POLICY=CENTRALIZED",
				"TEMPORAL_BUILDER_NIL_CONTEXT=REJECTED",
				"TEMPORAL_BUILDER_PROCESSING_GENERATION=v8",
				"TEMPORAL_BUILDER_REVIEW_STATUS=CLOSED",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
			},
		},
		{
			path: ".github/workflows/backend-ci.yml",
			required: []string{
				"Run temporal builder review audit",
				"go run ./tools/temporalbuilderreviewaudit -strict",
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
				failures = append(
					failures,
					fmt.Sprintf("%s missing %q", item.path, fragment),
				)
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s contains forbidden %q", item.path, fragment),
				)
			}
		}
	}

	if len(failures) == 0 {
		fmt.Println("Temporal builder review audit: PASS")
		return
	}

	fmt.Fprintln(os.Stderr, "Temporal builder review audit: FAIL")
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
		if _, statErr := os.Stat(
			filepath.Join(current, "apps/api/go.mod"),
		); statErr == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root was not found")
		}
		current = parent
	}
}
