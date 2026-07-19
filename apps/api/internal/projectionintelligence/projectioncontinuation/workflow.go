package projectioncontinuation

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

type continuationPreparation struct {
	selection projectionneighbors.Result
	pattern   projectionpatternconfidence.Result

	current         trajectory.FlightTrajectory
	currentEndpoint trajectory.TrackPoint4D

	currentAltitudeM         float64
	currentAltitudeAvailable bool

	candidateByID map[string]trajectory.FlightTrajectory

	fallbackReason       string
	selectionFingerprint string
	patternFingerprint   string
}

func (
	baseline *Baseline,
) prepareContinuation(
	request Request,
	plan projectionhorizon.Plan,
) continuationPreparation {
	selection, err := baseline.config.
		NeighborSelector.Select(
		projectionneighbors.Request{
			CurrentTrajectory: request.CurrentTrajectory,
			Candidates:        request.Candidates,
			AsOfTime:          plan.AsOfTime,
			RequiredContinuationDuration: plan.
				EffectiveDuration,
		},
	)
	if err != nil {
		return continuationPreparation{
			fallbackReason: "historical_neighbor_selection_failed",
		}
	}
	if err := selection.Validate(); err != nil {
		return continuationPreparation{
			fallbackReason: "historical_neighbor_selection_invalid",
			selectionFingerprint: selection.
				InputFingerprint,
		}
	}

	pattern, err := baseline.config.
		PatternConfidenceEvaluator.
		Evaluate(selection)
	if err != nil {
		return continuationPreparation{
			fallbackReason: "historical_pattern_confidence_failed",
			selectionFingerprint: selection.
				InputFingerprint,
		}
	}
	if err := pattern.Validate(); err != nil {
		return continuationPreparation{
			fallbackReason: "historical_pattern_confidence_invalid",
			selectionFingerprint: selection.
				InputFingerprint,
			patternFingerprint: pattern.
				InputFingerprint,
		}
	}
	if !patternMatchesSelection(
		pattern,
		selection,
	) {
		return continuationPreparation{
			fallbackReason: "historical_pattern_selection_mismatch",
			selectionFingerprint: selection.
				InputFingerprint,
			patternFingerprint: pattern.
				InputFingerprint,
		}
	}
	if !pattern.Usable {
		return continuationPreparation{
			fallbackReason: "historical_pattern_not_usable",
			selectionFingerprint: selection.
				InputFingerprint,
			patternFingerprint: pattern.
				InputFingerprint,
		}
	}

	current := trajectorySnapshotAt(
		request.CurrentTrajectory,
		plan.AsOfTime,
	)
	if len(current.Points) == 0 {
		return continuationPreparation{
			fallbackReason: "current_as_of_endpoint_unavailable",
			selectionFingerprint: selection.
				InputFingerprint,
			patternFingerprint: pattern.
				InputFingerprint,
		}
	}

	currentEndpoint :=
		current.Points[len(current.Points)-1]
	currentAltitudeM,
		currentAltitudeAvailable :=
		usableAltitude(currentEndpoint)

	return continuationPreparation{
		selection:                selection,
		pattern:                  pattern,
		current:                  current,
		currentEndpoint:          currentEndpoint,
		currentAltitudeM:         currentAltitudeM,
		currentAltitudeAvailable: currentAltitudeAvailable,
		candidateByID: buildCandidateIndex(
			request.Candidates,
			plan.AsOfTime,
		),
		selectionFingerprint: selection.
			InputFingerprint,
		patternFingerprint: pattern.
			InputFingerprint,
	}
}

func (
	preparation continuationPreparation,
) requiresFallback() bool {
	return preparation.fallbackReason != ""
}

type continuationPointResult struct {
	points           []projectioncontract.ProjectionPoint
	altitudeComplete bool
	fallbackReason   string
}

func (
	baseline *Baseline,
) projectForecastPoints(
	preparation continuationPreparation,
	plan projectionhorizon.Plan,
) continuationPointResult {
	points := make(
		[]projectioncontract.ProjectionPoint,
		0,
		len(plan.ForecastTimes),
	)
	altitudeComplete := true

	for index, forecastTime := range plan.ForecastTimes {
		samples := baseline.projectSamplesAt(
			preparation,
			forecastTime.Sub(plan.AsOfTime),
		)
		if len(samples) <
			baseline.config.MinimumPointSupport {
			return continuationPointResult{
				fallbackReason: "historical_continuation_point_support_insufficient",
			}
		}

		point, altitudeAvailable, err :=
			baseline.combineSamples(
				samples,
				preparation.pattern,
				plan,
				index,
				forecastTime,
			)
		if err != nil {
			return continuationPointResult{
				fallbackReason: "historical_continuation_combination_failed",
			}
		}
		if !altitudeAvailable {
			altitudeComplete = false
		}
		points = append(points, point)
	}

	return continuationPointResult{
		points:           points,
		altitudeComplete: altitudeComplete,
	}
}

func (
	baseline *Baseline,
) projectSamplesAt(
	preparation continuationPreparation,
	offset time.Duration,
) []projectedSample {
	samples := make(
		[]projectedSample,
		0,
		len(preparation.selection.Neighbors),
	)

	for _, neighbor := range preparation.selection.Neighbors {
		sample, usable :=
			baseline.projectNeighborSample(
				preparation,
				neighbor,
				offset,
			)
		if usable {
			samples = append(samples, sample)
		}
	}

	return samples
}

func (
	baseline *Baseline,
) projectNeighborSample(
	preparation continuationPreparation,
	neighbor projectionneighbors.Neighbor,
	offset time.Duration,
) (projectedSample, bool) {
	candidate, exists :=
		preparation.candidateByID[neighbor.TrajectoryID]
	if !exists ||
		neighbor.AnchorPointIndex < 0 ||
		neighbor.AnchorPointIndex >=
			len(candidate.Points) {
		return projectedSample{}, false
	}

	anchor :=
		candidate.Points[neighbor.AnchorPointIndex]
	targetTime :=
		anchor.ObservedAt.UTC().Add(offset)
	future, exists := interpolateTrajectoryPoint(
		candidate.Points,
		targetTime,
	)
	if !exists {
		return projectedSample{}, false
	}

	distanceM := greatCircleDistanceM(
		anchor.Latitude,
		anchor.Longitude,
		future.latitude,
		future.longitude,
	)
	bearing := initialBearingDegrees(
		anchor.Latitude,
		anchor.Longitude,
		future.latitude,
		future.longitude,
	)
	latitude, longitude, valid :=
		destinationPoint(
			preparation.currentEndpoint.Latitude,
			preparation.currentEndpoint.Longitude,
			bearing,
			distanceM,
		)
	if !valid ||
		!positiveFinite(
			neighbor.SimilarityScore,
		) {
		return projectedSample{}, false
	}

	sample := projectedSample{
		trajectoryID: neighbor.TrajectoryID,
		weight:       neighbor.SimilarityScore,
		latitude:     latitude,
		longitude:    longitude,
	}

	anchorAltitudeM,
		anchorAltitudeAvailable :=
		usableAltitude(anchor)
	if preparation.currentAltitudeAvailable &&
		anchorAltitudeAvailable &&
		future.altitudeM != nil {
		projectedAltitude :=
			preparation.currentAltitudeM +
				(*future.altitudeM -
					anchorAltitudeM)
		if finite(projectedAltitude) {
			sample.altitudeM =
				float64Pointer(
					projectedAltitude,
				)
		}
	}

	return sample, true
}
