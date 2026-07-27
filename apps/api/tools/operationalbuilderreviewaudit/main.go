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
			path: "apps/api/internal/domain/trajectory/model.go",
			required: []string{
				"VelocityAvailable",
				"HeadingAvailable",
				"VerticalRateAvailable",
				"OnGroundAvailable",
				"TelemetryAvailabilityKnown",
			},
		},
		{
			path: "apps/api/internal/domain/trajectory/track_point.go",
			required: []string{
				"func TrackPoint4DFromFlightState(",
				"state.VelocityAvailable",
				"state.HeadingAvailable",
				"state.VerticalRateAvailable",
				"state.OnGroundAvailable",
				"state.TelemetryAvailabilityKnown",
				"func (point TrackPoint4D) HasVelocity() bool",
				"func (point TrackPoint4D) HasHeading() bool",
				"func (point TrackPoint4D) HasVerticalRate() bool",
				"func (point TrackPoint4D) HasOnGroundState() bool",
			},
		},
		{
			path: "apps/api/internal/services/traffic/trackbuilder/trackbuilder.go",
			required: []string{
				"return trajectory.TrackPoint4DFromFlightState(state)",
			},
			forbidden: []string{
				"VelocityMPS:     state.VelocityMPS",
				"HeadingDegrees:  state.HeadingDegrees",
			},
		},
		{
			path: "apps/api/internal/features/extractor/fingerprint.go",
			required: []string{
				"VelocityAvailable",
				"HeadingAvailable",
				"VerticalRateAvailable",
				"OnGroundAvailable",
				"TelemetryAvailabilityKnown",
			},
		},
		{
			path: "apps/api/internal/repository/postgres/trajectory_feature_read.go",
			required: []string{
				"type FeatureTrajectoryReader struct",
				"NewFeatureTrajectoryReader(",
				"repository.withTrajectoryReadSnapshot(",
				"snapshot.getTrajectoryByID(",
				"snapshot.getLatestTrajectoryByICAO24(",
				"repository.listTrajectoryPoints(",
				"item.Points = points",
			},
		},
		{
			path: "apps/api/internal/repository/postgres/trajectory_child_read.go",
			forbidden: []string{
				"listTrajectoryPoints",
			},
		},
		{
			path: "apps/api/cmd/materialize-flight-features/main.go",
			required: []string{
				"postgres.NewFeatureTrajectoryReader(pool)",
			},
			forbidden: []string{
				"trajectoryRepository := postgres.NewTrajectoryRepository(pool)",
			},
		},
		{
			path: "apps/api/internal/repository/postgres/trajectory_read_queries.go",
			required: []string{
				"const trajectoryPointSelectColumns",
				"const trajectoryPointsByFlightIDAndWindowQuery",
				"const trajectoryPointsByICAO24AndWindowQuery",
				"state.observed_at >= $2",
				"state.observed_at <= $3",
				"state.latitude IS NOT NULL",
				"state.longitude IS NOT NULL",
				"state.observed_at ASC",
				"state.id ASC",
			},
			forbidden: []string{
				"COALESCE(state.velocity_mps, 0)",
				"COALESCE(state.heading_degrees, 0)",
				"COALESCE(state.vertical_rate_mps, 0)",
				"COALESCE(state.on_ground, false)",
			},
		},
		{
			path: "apps/api/internal/repository/postgres/trajectory_point_row_scan.go",
			required: []string{
				"var velocity pgtype.Float8",
				"var heading pgtype.Float8",
				"var verticalRate pgtype.Float8",
				"var onGround pgtype.Bool",
				"applyTelemetryDatabaseValues(",
				"trajectory.TrackPoint4DFromFlightState(state)",
			},
		},
		{
			path: "apps/api/internal/features/operationalbuilder/contracts.go",
			required: []string{
				`const Version = "operational-feature-builder-v2"`,
				"flightfeatures.OperationalRequiredFeatureFieldCount",
			},
		},
		{
			path: "apps/api/internal/features/operationalbuilder/errors.go",
			required: []string{
				"ErrContextRequired",
				"ErrTrajectoryStartTimeRequired",
				"ErrTrajectoryEndTimeRequired",
				"ErrInvalidTrajectoryWindow",
			},
		},
		{
			path: "apps/api/internal/features/operationalbuilder/builder.go",
			required: []string{
				"return flightfeatures.OperationalFeatures{}, ErrContextRequired",
				"collectSamples(ctx, item)",
				"item.PointCount != len(item.Points)",
				"samples.supportingPointCount",
				"samples.groundStateCount",
				"summarizeContext(",
				"sumContext(",
				"OperationalLimitationAggregateNonFinite",
			},
			forbidden: []string{
				"ctx = context.Background()",
				"SupportingPointCount: len(item.Points)",
				"float64(len(item.Points))",
			},
		},
		{
			path: "apps/api/internal/features/operationalbuilder/samples.go",
			required: []string{
				"resolveOperationalWindow(item)",
				"sort.SliceStable(ordered",
				"point.HasVelocity()",
				"point.HasHeading()",
				"point.HasVerticalRate()",
				"point.HasOnGroundState()",
				"point.HeadingDegrees >= 0",
				"point.HeadingDegrees < 360",
				"previousHeadingUsable = false",
				"selectedAltitude := barometric",
				"selectedAltitude = geometric",
				"shortestHeadingChange(",
				"collection.supportingPointCount++",
			},
			forbidden: []string{
				"normalizeHeading(point.HeadingDegrees)",
				"operational_heading_normalized",
				"operational_point_order_nonmonotonic",
			},
		},
		{
			path: "apps/api/internal/features/flightfeatures/operational_policy.go",
			required: []string{
				"observation-weighted-kahan-v1",
				"single-source-prefer-barometric-v1",
				"chronological-shortest-arc-contiguous-valid-runs-v1",
				"OperationalLimitationHeadingSequenceGap",
				"OperationalLimitationOnGroundMeasurementUnavailable",
			},
		},
		{
			path: "apps/api/internal/features/operationalbuilder/operationalbuilder_review_hardening_test.go",
			required: []string{
				"TestBuilderRejectsNilContext",
				"TestBuilderFiltersWindowAndIsPermutationInvariant",
				"TestBuilderPreservesUnavailableZeroTelemetry",
				"TestBuilderGroundSharesUseOnlyAvailableStates",
				"TestBuilderExcludesInvalidHeadingAndDoesNotBridgeGap",
				"TestBuilderUsesSingleAltitudeSource",
				"TestBuilderRejectsConflictingGroundAltitude",
				"TestBuilderRejectsNonFiniteAggregate",
				"TestBuilderObservesCancellationDuringPointScan",
			},
		},
		{
			path: "apps/api/internal/repository/postgres/trajectory_point_read_integration_test.go",
			required: []string{
				"TestTrajectoryPointReadPreservesNullableOperationalTelemetry",
				"TEST_DATABASE_URL",
				"first.HasVelocity()",
				"second.HasVelocity()",
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
			path: "docs/121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md",
			required: []string{
				"OPERATIONAL_PRODUCTION_POINT_HYDRATION=ENFORCED",
				"OPERATIONAL_NULL_TELEMETRY_SEMANTICS=PRESERVED",
				"OPERATIONAL_INVALID_HEADING_NORMALIZATION=CLOSED",
				"OPERATIONAL_ALTITUDE_SOURCE_MIXING=CLOSED",
				"OPERATIONAL_BUILDER_PROCESSING_GENERATION=v10",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
				"OPERATIONAL_BUILDER_REVIEW_STATUS=CLOSED",
			},
		},
		{
			path: ".github/workflows/backend-ci.yml",
			required: []string{
				"Run operational builder review audit",
				"go run ./tools/operationalbuilderreviewaudit -strict",
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
		fmt.Println("Operational builder review audit: PASS")
		return
	}
	fmt.Fprintln(os.Stderr, "Operational builder review audit: FAIL")
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
