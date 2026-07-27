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
			path: "apps/api/internal/features/geographicalbuilder/contracts.go",
			required: []string{
				`const Version = "geographical-feature-builder-v3"`,
				"MinimumGeographicCellPrecision = 1",
				"MaximumGeographicCellPrecision = 6",
				"flightfeatures.GeographicalRequiredFeatureFieldCount",
			},
		},
		{
			path: "apps/api/internal/features/geographicalbuilder/errors.go",
			required: []string{
				"ErrContextRequired",
				"zero selects the default",
			},
		},
		{
			path: "apps/api/internal/features/geographicalbuilder/builder.go",
			required: []string{
				"return flightfeatures.GeographicalFeatures{}, ErrContextRequired",
				"point.ObservedAt.IsZero()",
				"observedAt.Before(startTime)",
				"observedAt.After(endTime)",
				"sort.SliceStable(observed",
				"len(points.coordinates) >= 2",
				"collectSegmentCoordinates(",
				"pathEdges:",
				"segments.pathEdges",
				"GeographicalLimitationSegmentDiscontinuityExcluded",
				"LimitationTrajectoryPointCountMetadataMismatch",
			},
			forbidden: []string{
				"ctx = context.Background()",
				"observedPathDistanceKM(coordinates)",
			},
		},
		{
			path: "apps/api/internal/features/geographicalbuilder/geometry.go",
			required: []string{
				"type coordinateEdge struct",
				"observedEdgeDistanceKMContext(",
				"Kahan compensated summation",
				"edgeSetCrossesAntimeridianContext(",
				"uniqueGeographicCellCountContext(",
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/geographical_policy.go",
			required: []string{
				"mean-earth-sphere-haversine-v1",
				"decimal-degree-round-half-away-from-zero-v1",
				"CircularLongitudeSpanDegrees",
				"GeographicalLimitationSegmentEndpointFallback",
				"HasLimitationCode",
			},
		},
		{
			path: "apps/api/internal/features/validator/geographical_policy.go",
			required: []string{
				"validateGeographicalLongitudeEnvelope",
				"CircularLongitudeSpanDegrees",
				"validateGeographicalDistanceRelationships",
				"GeographicalLimitationSegmentEndpointFallback",
			},
		},
		{
			path: "apps/api/internal/features/validator/rules.go",
			required: []string{
				"validateGeographicalLongitudeEnvelope(",
				"validateGeographicalDistanceRelationships(",
				"!flightfeatures.HasLimitationCode(",
			},
			forbidden: []string{
				"item.MinimumLongitude > item.MaximumLongitude &&\n\t\t!item.CrossesAntimeridian",
			},
		},
		{
			path: "apps/api/internal/features/geographicalbuilder/geographicalbuilder_review_hardening_test.go",
			required: []string{
				"TestBuilderRejectsNilContext",
				"TestBuilderOrdersEligiblePointsByObservationTime",
				"TestBuilderTieBreaksEqualObservationTimesDeterministically",
				"TestBuilderExcludesMissingAndOutsideWindowPointTimestamps",
				"TestBuilderUsesSegmentsWhenOnlyOnePointIsEligible",
				"TestSegmentFallbackExcludesDiscontinuityFromObservedPath",
				"TestCircularEnvelopeMayWrapWithoutPathCrossing",
				"TestGeometryPassesObserveContextCancellation",
			},
		},
		{
			path: "apps/api/internal/features/validator/geographical_review_hardening_test.go",
			required: []string{
				"TestValidatorAcceptsWrappedEnvelopeWithoutPathCrossing",
				"TestValidatorAllowsDisconnectedSegmentFallbackDistances",
			},
		},
		{
			path: "apps/api/internal/features/extractor/snapshot_validation.go",
			required: []string{
				"point.ObservedAt.After(cutoff)",
				"segment.StartTime.After(cutoff)",
				"segment.EndTime.After(cutoff)",
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/schema.go",
			required: []string{
				`Name:        "geographical.minimum_latitude"`,
				`Name:        "geographical.maximum_latitude"`,
				`Name:        "geographical.minimum_longitude"`,
				`Name:        "geographical.maximum_longitude"`,
				"excludes unobserved discontinuities",
			},
			forbidden: []string{
				`Name:        "geographical.geographic_cell_precision"`,
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/model.go",
			required: []string{
				`flight-feature-processing-pipeline-v10`,
				"western circular envelope bound",
				"eastern circular envelope bound",
			},
		},
		{
			path: "apps/api/internal/features/validator/contracts.go",
			required: []string{
				`flight-feature-validator-v4`,
			},
		},
		{
			path: "apps/api/internal/features/featurepipeline/contracts.go",
			required: []string{
				`flight-feature-processing-pipeline-v10`,
			},
		},
		{
			path: "docs/120_GEOGRAPHICAL_BUILDER_REVIEW_HARDENING.md",
			required: []string{
				"GEOGRAPHICAL_POINT_WINDOW_FILTERING=ENFORCED",
				"GEOGRAPHICAL_POINT_CHRONOLOGICAL_ORDER=ENFORCED",
				"GEOGRAPHICAL_SEGMENT_GAP_BRIDGING=CLOSED",
				"GEOGRAPHICAL_CIRCULAR_ENVELOPE_VALIDATION=ENFORCED",
				"GEOGRAPHICAL_BUILDER_PROCESSING_GENERATION=v9",
				"GEOGRAPHICAL_BUILDER_REVIEW_STATUS=CLOSED",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
			},
		},
		{
			path: ".github/workflows/backend-ci.yml",
			required: []string{
				"Run geographical builder review audit",
				"go run ./tools/geographicalbuilderreviewaudit -strict",
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
		fmt.Println("Geographical builder review audit: PASS")
		return
	}
	fmt.Fprintln(os.Stderr, "Geographical builder review audit: FAIL")
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
