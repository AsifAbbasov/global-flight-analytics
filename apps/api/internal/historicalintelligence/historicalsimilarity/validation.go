package historicalsimilarity

import (
	"fmt"
	"math"
	"strings"
)

var expectedReasons = []string{
	"geometry_shape_similarity",
	"worst_endpoint_proximity",
	"path_length_similarity",
	"duration_similarity",
}

func (result Result) Validate() error {
	validators := []func(Result) error{
		validateResultIdentity,
		validateResultCounts,
		validateResultMeasurements,
		validateResultPolicy,
		validateResultComponents,
		validateResultConfidence,
		validateResultExplanations,
		validateResultFingerprint,
	}
	for _, validate := range validators {
		if err := validate(result); err != nil {
			return err
		}
	}
	return nil
}

func validateResultIdentity(
	result Result,
) error {
	if result.Version != Version {
		return invalidResult("version")
	}
	if strings.TrimSpace(
		result.ReferenceTrajectoryID,
	) == "" ||
		strings.TrimSpace(
			result.CandidateTrajectoryID,
		) == "" ||
		result.ReferenceTrajectoryID ==
			result.CandidateTrajectoryID {
		return invalidResult(
			"trajectory identifiers",
		)
	}
	if !ratio(result.Score) ||
		result.Level !=
			LevelForScore(result.Score) {
		return invalidResult(
			"similarity score or level",
		)
	}
	return nil
}

func validateResultCounts(
	result Result,
) error {
	if result.ReferencePointCount < 2 ||
		result.ReferencePointCount >
			MaximumInputPointCount ||
		result.CandidatePointCount < 2 ||
		result.CandidatePointCount >
			MaximumInputPointCount ||
		result.SampleCount < 2 ||
		result.SampleCount >
			MaximumSampleCount {
		return invalidResult("point counts")
	}
	return nil
}

func validateResultMeasurements(
	result Result,
) error {
	values := []float64{
		result.MeanDistanceKM,
		result.MaximumDistanceKM,
		result.StartEndpointDistanceKM,
		result.EndEndpointDistanceKM,
		result.ReferencePathLengthKM,
		result.CandidatePathLengthKM,
		result.ReferenceDurationSeconds,
		result.CandidateDurationSeconds,
	}
	for _, value := range values {
		if !finite(value) || value < 0 {
			return invalidResult(
				"non-negative measurement",
			)
		}
	}
	if result.MeanDistanceKM >
		result.MaximumDistanceKM+
			scaledTolerance(
				result.MaximumDistanceKM,
			) {
		return invalidResult(
			"mean distance exceeds maximum distance",
		)
	}
	if result.ReferenceDurationSeconds <= 0 ||
		result.CandidateDurationSeconds <= 0 {
		return invalidResult(
			"trajectory duration",
		)
	}
	return nil
}

func validateResultPolicy(
	result Result,
) error {
	config := Config{
		MinimumPointCount: 2,
		SampleCount:       result.SampleCount,
		GeometryScoreScaleKM: result.
			Policy.GeometryScoreScaleKM,
		EndpointScoreScaleKM: result.
			Policy.EndpointScoreScaleKM,
		GeometryWeight: result.
			Policy.GeometryWeight,
		EndpointsWeight: result.
			Policy.EndpointsWeight,
		PathLengthWeight: result.
			Policy.PathLengthWeight,
		DurationWeight: result.
			Policy.DurationWeight,
	}
	if err := config.Validate(); err != nil {
		return invalidResult("scoring policy")
	}
	return nil
}

func validateResultComponents(
	result Result,
) error {
	if len(result.Components) !=
		componentCount {
		return invalidResult("component count")
	}

	expected := expectedComponents(result)
	weightedScore := 0.0
	compensation := 0.0
	for index, component := range result.Components {
		if index >= len(expected) ||
			component.Name !=
				expected[index].Name {
			return invalidResult(
				"component order or name",
			)
		}
		want := expected[index]
		if !nearlyEqual(
			component.Score,
			want.Score,
		) ||
			!nearlyEqual(
				component.Weight,
				want.Weight,
			) ||
			!nearlyEqual(
				component.ObservedValue,
				want.ObservedValue,
			) ||
			component.Unit != want.Unit {
			return invalidResult(
				"component mathematics",
			)
		}

		contribution :=
			component.Score *
				component.Weight
		corrected := contribution -
			compensation
		next := weightedScore +
			corrected
		compensation =
			(next - weightedScore) -
				corrected
		weightedScore = next
	}
	weightedScore = clampRatio(
		weightedScore,
	)
	if !nearlyEqual(
		result.Score,
		weightedScore,
	) {
		return invalidResult(
			"weighted similarity score",
		)
	}
	return nil
}

func expectedComponents(
	result Result,
) []Component {
	pathDifference := relativeDifference(
		result.ReferencePathLengthKM,
		result.CandidatePathLengthKM,
	)
	durationDifference := relativeDifference(
		result.ReferenceDurationSeconds,
		result.CandidateDurationSeconds,
	)
	endpointObserved := math.Max(
		result.StartEndpointDistanceKM,
		result.EndEndpointDistanceKM,
	)

	return []Component{
		{
			Name: ComponentGeometry,
			Score: inverseScaleScore(
				result.MeanDistanceKM,
				result.Policy.
					GeometryScoreScaleKM,
			),
			Weight: result.Policy.
				GeometryWeight,
			ObservedValue: result.
				MeanDistanceKM,
			Unit: "kilometres",
		},
		{
			Name: ComponentEndpoints,
			Score: inverseScaleScore(
				endpointObserved,
				result.Policy.
					EndpointScoreScaleKM,
			),
			Weight: result.Policy.
				EndpointsWeight,
			ObservedValue: endpointObserved,
			Unit:          "kilometres",
		},
		{
			Name:  ComponentPathLength,
			Score: 1 - pathDifference,
			Weight: result.Policy.
				PathLengthWeight,
			ObservedValue: pathDifference,
			Unit:          "ratio",
		},
		{
			Name:  ComponentDuration,
			Score: 1 - durationDifference,
			Weight: result.Policy.
				DurationWeight,
			ObservedValue: durationDifference,
			Unit:          "ratio",
		},
	}
}

func validateResultConfidence(
	result Result,
) error {
	if err := validateEvidenceQuality(
		result.Confidence.Reference,
	); err != nil {
		return err
	}
	if err := validateEvidenceQuality(
		result.Confidence.Candidate,
	); err != nil {
		return err
	}

	expectedScore := math.Min(
		result.Confidence.Reference.Score,
		result.Confidence.Candidate.Score,
	)
	if !ratio(result.Confidence.Score) ||
		!nearlyEqual(
			result.Confidence.Score,
			expectedScore,
		) ||
		result.Confidence.Level !=
			ConfidenceLevelForScore(
				result.Confidence.Score,
			) {
		return invalidResult(
			"comparison confidence",
		)
	}
	if len(result.Confidence.Reasons) == 0 {
		return invalidResult(
			"comparison confidence reasons",
		)
	}
	for _, reason := range result.Confidence.Reasons {
		if strings.TrimSpace(reason.Code) == "" ||
			strings.TrimSpace(reason.Message) == "" {
			return invalidResult(
				"comparison confidence reason",
			)
		}
	}
	return nil
}

func validateEvidenceQuality(
	quality EvidenceQuality,
) error {
	ratios := []float64{
		quality.Score,
		quality.DeclaredQualityScore,
		quality.SegmentQualityScore,
		quality.CoverageContinuityScore,
		quality.ObservationCadenceScore,
		quality.PointRetentionScore,
	}
	for _, value := range ratios {
		if !ratio(value) {
			return invalidResult(
				"trajectory evidence quality ratio",
			)
		}
	}
	if quality.InputPointCount < 1 ||
		quality.UsablePointCount < 1 ||
		quality.ExcludedPointCount < 0 ||
		quality.UsablePointCount+
			quality.ExcludedPointCount !=
			quality.InputPointCount ||
		quality.EqualTimestampPointCount < 0 ||
		quality.EqualTimestampPointCount >
			quality.UsablePointCount-1 ||
		quality.CoverageGapCount < 0 ||
		quality.RelevantSegmentCount < 0 ||
		quality.NonObservedSegmentCount < 0 ||
		quality.InvalidSegmentCount < 0 ||
		quality.InvalidSegmentCount >
			quality.NonObservedSegmentCount ||
		quality.NonObservedSegmentCount >
			quality.RelevantSegmentCount {
		return invalidResult(
			"trajectory evidence quality counts",
		)
	}

	expectedScore :=
		quality.DeclaredQualityScore*
			declaredQualityWeight +
			quality.SegmentQualityScore*
				segmentQualityWeight +
			quality.CoverageContinuityScore*
				coverageContinuityWeight +
			quality.ObservationCadenceScore*
				observationCadenceWeight +
			quality.PointRetentionScore*
				pointRetentionWeight
	expectedScore = clampRatio(expectedScore)
	if !nearlyEqual(
		quality.Score,
		expectedScore,
	) {
		return invalidResult(
			"trajectory evidence quality mathematics",
		)
	}
	return nil
}

func validateResultExplanations(
	result Result,
) error {
	if len(result.Reasons) !=
		len(expectedReasons) {
		return invalidResult("reasons")
	}
	for index, reason := range result.Reasons {
		if reason != expectedReasons[index] {
			return invalidResult(
				"reason order or value",
			)
		}
	}

	seen := make(map[string]struct{})
	for _, limitation := range result.Limitations {
		if strings.TrimSpace(
			limitation.Code,
		) == "" ||
			strings.TrimSpace(
				limitation.Message,
			) == "" {
			return invalidResult("limitation")
		}
		key := limitation.Code + "\x00" +
			limitation.Message
		if _, exists := seen[key]; exists {
			return invalidResult(
				"duplicate limitation",
			)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateResultFingerprint(
	result Result,
) error {
	if !fingerprintPattern.MatchString(
		result.InputFingerprint,
	) {
		return invalidResult(
			"input fingerprint",
		)
	}
	return nil
}

func invalidResult(
	field string,
) error {
	return fmt.Errorf(
		"%w: %s",
		ErrResultInvalid,
		field,
	)
}

func nearlyEqual(
	left float64,
	right float64,
) bool {
	if !finite(left) ||
		!finite(right) {
		return false
	}
	return math.Abs(left-right) <=
		scaledTolerance(
			math.Max(
				math.Abs(left),
				math.Abs(right),
			),
		)
}

func scaledTolerance(
	scale float64,
) float64 {
	return numericTolerance *
		math.Max(1, math.Abs(scale))
}
