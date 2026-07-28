package historicalsimilarity

import "math"

type comparisonMetrics struct {
	meanDistanceKM    float64
	maximumDistanceKM float64

	startEndpointDistanceKM float64
	endEndpointDistanceKM   float64
	endpointObservedKM      float64

	referencePathLengthKM float64
	candidatePathLengthKM float64
	pathLengthDifference  float64

	referenceDurationSeconds float64
	candidateDurationSeconds float64
	durationDifference       float64
}

func buildResult(
	reference preparedTrajectory,
	candidate preparedTrajectory,
	config Config,
) Result {
	metrics := calculateMetrics(
		reference,
		candidate,
	)
	components := buildComponents(
		metrics,
		config,
	)

	score := 0.0
	compensation := 0.0
	for _, component := range components {
		contribution :=
			component.Score *
				component.Weight
		corrected := contribution -
			compensation
		next := score + corrected
		compensation = (next - score) -
			corrected
		score = next
	}
	score = clampRatio(score)

	limitations := append(
		roleNotices(
			"reference",
			reference.limitations,
		),
		roleNotices(
			"candidate",
			candidate.limitations,
		)...,
	)
	limitations = append(
		limitations,
		Notice{
			Code:    "spherical_earth_similarity_model",
			Message: "Distances and path resampling use a deterministic spherical-Earth great-circle model; ellipsoidal geodesic corrections are outside this heuristic contract.",
		},
	)

	result := Result{
		Version: Version,

		ReferenceTrajectoryID: reference.id,
		CandidateTrajectoryID: candidate.id,

		Score: score,
		Level: LevelForScore(score),

		Confidence: comparisonConfidence(
			reference.quality,
			candidate.quality,
		),
		Policy: config.scoringPolicy(),

		ReferencePointCount: len(reference.points),
		CandidatePointCount: len(candidate.points),
		SampleCount:         config.SampleCount,

		MeanDistanceKM: metrics.
			meanDistanceKM,
		MaximumDistanceKM: metrics.
			maximumDistanceKM,
		StartEndpointDistanceKM: metrics.
			startEndpointDistanceKM,
		EndEndpointDistanceKM: metrics.
			endEndpointDistanceKM,
		ReferencePathLengthKM: metrics.
			referencePathLengthKM,
		CandidatePathLengthKM: metrics.
			candidatePathLengthKM,
		ReferenceDurationSeconds: metrics.
			referenceDurationSeconds,
		CandidateDurationSeconds: metrics.
			candidateDurationSeconds,

		Components: components,
		Reasons: []string{
			"geometry_shape_similarity",
			"worst_endpoint_proximity",
			"path_length_similarity",
			"duration_similarity",
		},
		Limitations: normalizeNotices(
			limitations,
		),
	}
	result.InputFingerprint =
		comparisonFingerprint(
			reference,
			candidate,
			config,
		)
	return result
}

func calculateMetrics(
	reference preparedTrajectory,
	candidate preparedTrajectory,
) comparisonMetrics {
	meanDistance, maximumDistance :=
		sampleDistances(
			reference.samples,
			candidate.samples,
		)
	startDistance := haversineKM(
		reference.samples[0],
		candidate.samples[0],
	)
	endDistance := haversineKM(
		reference.samples[len(reference.samples)-1],
		candidate.samples[len(candidate.samples)-1],
	)

	return comparisonMetrics{
		meanDistanceKM:    meanDistance,
		maximumDistanceKM: maximumDistance,

		startEndpointDistanceKM: startDistance,
		endEndpointDistanceKM:   endDistance,
		endpointObservedKM: math.Max(
			startDistance,
			endDistance,
		),

		referencePathLengthKM: reference.pathLengthKM,
		candidatePathLengthKM: candidate.pathLengthKM,
		pathLengthDifference: relativeDifference(
			reference.pathLengthKM,
			candidate.pathLengthKM,
		),

		referenceDurationSeconds: reference.durationSeconds,
		candidateDurationSeconds: candidate.durationSeconds,
		durationDifference: relativeDifference(
			reference.durationSeconds,
			candidate.durationSeconds,
		),
	}
}

func buildComponents(
	metrics comparisonMetrics,
	config Config,
) []Component {
	return []Component{
		{
			Name: ComponentGeometry,
			Score: inverseScaleScore(
				metrics.meanDistanceKM,
				config.GeometryScoreScaleKM,
			),
			Weight: config.GeometryWeight,
			ObservedValue: metrics.
				meanDistanceKM,
			Unit: "kilometres",
		},
		{
			Name: ComponentEndpoints,
			Score: inverseScaleScore(
				metrics.endpointObservedKM,
				config.EndpointScoreScaleKM,
			),
			Weight: config.EndpointsWeight,
			ObservedValue: metrics.
				endpointObservedKM,
			Unit: "kilometres",
		},
		{
			Name: ComponentPathLength,
			Score: 1 - metrics.
				pathLengthDifference,
			Weight: config.PathLengthWeight,
			ObservedValue: metrics.
				pathLengthDifference,
			Unit: "ratio",
		},
		{
			Name: ComponentDuration,
			Score: 1 - metrics.
				durationDifference,
			Weight: config.DurationWeight,
			ObservedValue: metrics.
				durationDifference,
			Unit: "ratio",
		},
	}
}

func inverseScaleScore(
	value float64,
	scale float64,
) float64 {
	return clampRatio(1 - value/scale)
}

func relativeDifference(
	left float64,
	right float64,
) float64 {
	if left == 0 && right == 0 {
		return 0
	}
	scale := math.Max(
		math.Abs(left),
		math.Abs(right),
	)
	if scale == 0 {
		return 0
	}
	return clampRatio(
		math.Abs(left-right) / scale,
	)
}
