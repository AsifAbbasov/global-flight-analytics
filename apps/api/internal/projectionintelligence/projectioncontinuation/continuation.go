package projectioncontinuation

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

const (
	Version    = "local-historical-neighbor-continuation-v1"
	MethodName = "local_historical_neighbor_continuation"

	FingerprintVersion         = "local-historical-neighbor-continuation-fingerprint-v1"
	FallbackFingerprintVersion = "local-historical-neighbor-fallback-fingerprint-v1"
)

var (
	ErrTrajectoryIDRequired = errors.New(
		"projection trajectory identifier is required",
	)
	ErrGeneratedAtInvalid = errors.New(
		"projection generated-at time must not be before the as-of time",
	)
	ErrCurrentTrajectoryUnavailable = errors.New(
		"current trajectory does not contain a usable as-of endpoint",
	)
	ErrContinuationContractInvalid = errors.New(
		"generated historical continuation contract is invalid",
	)
	ErrFallbackProjectionFailed = errors.New(
		"kinematic fallback projection failed",
	)
)

type Baseline struct {
	config Config
}

func New(
	config Config,
) (*Baseline, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate local historical continuation config: %w",
			err,
		)
	}

	return &Baseline{
		config: config,
	}, nil
}

type Request struct {
	CurrentTrajectory trajectory.FlightTrajectory
	Candidates        []trajectory.FlightTrajectory

	AsOfTime          time.Time
	RequestedDuration time.Duration
	GeneratedAt       time.Time
}

type projectedSample struct {
	trajectoryID string
	weight       float64

	latitude  float64
	longitude float64
	altitudeM *float64
}

func (
	baseline *Baseline,
) Project(
	request Request,
) (projectioncontract.Result, error) {
	if baseline == nil {
		return projectioncontract.Result{},
			ErrHorizonPlannerRequired
	}
	if strings.TrimSpace(
		request.CurrentTrajectory.ID,
	) == "" {
		return projectioncontract.Result{},
			ErrTrajectoryIDRequired
	}

	plan, err := baseline.config.
		HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime: request.AsOfTime,
			RequestedDuration: request.
				RequestedDuration,
		},
	)
	if err != nil {
		return projectioncontract.Result{},
			fmt.Errorf(
				"build historical continuation horizon: %w",
				err,
			)
	}

	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(
			plan.AsOfTime,
		) {
		return projectioncontract.Result{},
			ErrGeneratedAtInvalid
	}

	selection, err := baseline.config.
		NeighborSelector.Select(
		projectionneighbors.Request{
			CurrentTrajectory:            request.CurrentTrajectory,
			Candidates:                   request.Candidates,
			AsOfTime:                     plan.AsOfTime,
			RequiredContinuationDuration: plan.EffectiveDuration,
		},
	)
	if err != nil {
		return baseline.fallback(
			request,
			"historical_neighbor_selection_failed",
			"",
			"",
		)
	}
	if err := selection.Validate(); err != nil {
		return baseline.fallback(
			request,
			"historical_neighbor_selection_invalid",
			selection.InputFingerprint,
			"",
		)
	}

	pattern, err := baseline.config.
		PatternConfidenceEvaluator.
		Evaluate(selection)
	if err != nil {
		return baseline.fallback(
			request,
			"historical_pattern_confidence_failed",
			selection.InputFingerprint,
			"",
		)
	}
	if err := pattern.Validate(); err != nil {
		return baseline.fallback(
			request,
			"historical_pattern_confidence_invalid",
			selection.InputFingerprint,
			pattern.InputFingerprint,
		)
	}
	if !patternMatchesSelection(
		pattern,
		selection,
	) {
		return baseline.fallback(
			request,
			"historical_pattern_selection_mismatch",
			selection.InputFingerprint,
			pattern.InputFingerprint,
		)
	}
	if !pattern.Usable {
		return baseline.fallback(
			request,
			"historical_pattern_not_usable",
			selection.InputFingerprint,
			pattern.InputFingerprint,
		)
	}

	current := trajectorySnapshotAt(
		request.CurrentTrajectory,
		plan.AsOfTime,
	)
	if len(current.Points) == 0 {
		return baseline.fallback(
			request,
			"current_as_of_endpoint_unavailable",
			selection.InputFingerprint,
			pattern.InputFingerprint,
		)
	}

	currentEndpoint :=
		current.Points[len(current.Points)-1]
	currentAltitudeM,
		currentAltitudeAvailable :=
		usableAltitude(currentEndpoint)

	candidateByID :=
		buildCandidateIndex(
			request.Candidates,
			plan.AsOfTime,
		)

	points := make(
		[]projectioncontract.ProjectionPoint,
		0,
		len(plan.ForecastTimes),
	)
	altitudeComplete := true

	for index, forecastTime := range plan.ForecastTimes {
		offset := forecastTime.Sub(
			plan.AsOfTime,
		)
		samples := make(
			[]projectedSample,
			0,
			len(selection.Neighbors),
		)

		for _, neighbor := range selection.Neighbors {
			candidate, exists :=
				candidateByID[neighbor.TrajectoryID]
			if !exists ||
				neighbor.AnchorPointIndex < 0 ||
				neighbor.AnchorPointIndex >=
					len(candidate.Points) {
				continue
			}

			anchor := candidate.Points[neighbor.AnchorPointIndex]
			targetTime :=
				anchor.ObservedAt.UTC().
					Add(offset)
			future, exists :=
				interpolateTrajectoryPoint(
					candidate.Points,
					targetTime,
				)
			if !exists {
				continue
			}

			distanceM :=
				greatCircleDistanceM(
					anchor.Latitude,
					anchor.Longitude,
					future.latitude,
					future.longitude,
				)
			bearing :=
				initialBearingDegrees(
					anchor.Latitude,
					anchor.Longitude,
					future.latitude,
					future.longitude,
				)
			latitude, longitude, valid :=
				destinationPoint(
					currentEndpoint.Latitude,
					currentEndpoint.Longitude,
					bearing,
					distanceM,
				)
			if !valid ||
				!positiveFinite(
					neighbor.
						SimilarityScore,
				) {
				continue
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
			if currentAltitudeAvailable &&
				anchorAltitudeAvailable &&
				future.altitudeM != nil {
				projectedAltitude :=
					currentAltitudeM +
						(*future.altitudeM -
							anchorAltitudeM)
				if finite(
					projectedAltitude,
				) {
					sample.altitudeM =
						float64Pointer(
							projectedAltitude,
						)
				}
			}

			samples = append(
				samples,
				sample,
			)
		}

		if len(samples) <
			baseline.config.
				MinimumPointSupport {
			return baseline.fallback(
				request,
				"historical_continuation_point_support_insufficient",
				selection.InputFingerprint,
				pattern.InputFingerprint,
			)
		}

		point, altitudeAvailable,
			err := baseline.combineSamples(
			samples,
			pattern,
			plan,
			index,
			forecastTime,
		)
		if err != nil {
			return baseline.fallback(
				request,
				"historical_continuation_combination_failed",
				selection.InputFingerprint,
				pattern.InputFingerprint,
			)
		}
		if !altitudeAvailable {
			altitudeComplete = false
		}

		points = append(
			points,
			point,
		)
	}

	status :=
		projectioncontract.
			ResultStatusComplete
	limitations :=
		historicalContinuationLimitations(
			selection,
			pattern,
		)
	if plan.Truncated ||
		selection.Status !=
			projectionneighbors.StatusComplete ||
		pattern.Status !=
			projectionpatternconfidence.
				StatusComplete ||
		!altitudeComplete {
		status =
			projectioncontract.
				ResultStatusLimited
	}
	if plan.Truncated {
		limitations = append(
			limitations,
			projectioncontract.Limitation{
				Code:    "projection_horizon_truncated",
				Message: "Requested duration exceeded the configured maximum and was truncated.",
				Scope:   "horizon",
			},
		)
	}
	if !altitudeComplete {
		limitations = append(
			limitations,
			projectioncontract.Limitation{
				Code:    "historical_continuation_altitude_partial",
				Message: "At least one forecast point lacked sufficient historical altitude support, so only horizontal position was published for that point.",
				Scope:   "position",
			},
		)
	}

	result := projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        status,

		TrajectoryID: current.ID,
		FlightID:     current.FlightID,
		AircraftID:   current.AircraftID,
		ICAO24:       current.ICAO24,
		Callsign:     current.Callsign,

		Method: projectioncontract.Method{
			Name:    MethodName,
			Version: Version,
			DecisionClass: projectioncontract.
				DecisionClassExperimental,
		},
		Horizon: plan.ContractHorizon(),
		Points:  points,

		Confidence: minimumPointConfidence(points),
		Limitations: normalizeLimitations(
			limitations,
		),
		Explanations: []projectioncontract.Explanation{
			{
				Code:    "translated_historical_continuations",
				Message: "Observed movements after each selected historical anchor were translated onto the current trajectory endpoint.",
			},
			{
				Code:    "similarity_weighted_consensus",
				Message: "Forecast coordinates combine usable historical continuations using normalized similarity weights.",
			},
			{
				Code:    "neighbor_disagreement_uncertainty",
				Message: "Published uncertainty includes configured growth and weighted disagreement between historical continuation samples.",
			},
		},
		ScopeGuard: projectioncontract.
			ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: continuationFingerprint(
				current,
				selection,
				pattern,
				plan,
				baseline.config,
			),
			Inputs: continuationInputs(
				currentEndpoint,
				selection,
			),
			LatestInputObservedAt: currentEndpoint.
				ObservedAt.UTC(),
		},
		GeneratedAt: generatedAt,
	}

	return validateProjectionResult(
		result,
	)
}

func (
	baseline *Baseline,
) combineSamples(
	samples []projectedSample,
	pattern projectionpatternconfidence.Result,
	plan projectionhorizon.Plan,
	sequence int,
	forecastTime time.Time,
) (
	projectioncontract.ProjectionPoint,
	bool,
	error,
) {
	geoPoints := make(
		[]weightedGeoPoint,
		0,
		len(samples),
	)
	totalWeight := 0.0
	for _, sample := range samples {
		geoPoints = append(
			geoPoints,
			weightedGeoPoint{
				latitude:  sample.latitude,
				longitude: sample.longitude,
				weight:    sample.weight,
			},
		)
		totalWeight += sample.weight
	}

	latitude, longitude, valid :=
		weightedMeanGeoPoint(
			geoPoints,
		)
	if !valid ||
		!positiveFinite(totalWeight) {
		return projectioncontract.
				ProjectionPoint{},
			false,
			ErrContinuationContractInvalid
	}

	offsetSeconds := forecastTime.Sub(
		plan.AsOfTime,
	).Seconds()
	horizonSeconds :=
		plan.EffectiveDuration.Seconds()
	if offsetSeconds <= 0 ||
		horizonSeconds <= 0 {
		return projectioncontract.
				ProjectionPoint{},
			false,
			ErrContinuationContractInvalid
	}

	horizontalSpreadSquared := 0.0
	altitudeWeight := 0.0
	weightedAltitude := 0.0
	for _, sample := range samples {
		distanceM :=
			greatCircleDistanceM(
				latitude,
				longitude,
				sample.latitude,
				sample.longitude,
			)
		horizontalSpreadSquared +=
			sample.weight *
				distanceM *
				distanceM

		if sample.altitudeM != nil {
			weightedAltitude +=
				sample.weight *
					*sample.altitudeM
			altitudeWeight +=
				sample.weight
		}
	}
	horizontalSpreadM := math.Sqrt(
		horizontalSpreadSquared /
			totalWeight,
	)
	configuredHorizontal :=
		baseline.config.
			InitialHorizontalUncertaintyM +
			baseline.config.
				HorizontalUncertaintyGrowthMPS*
				offsetSeconds
	horizontalUncertaintyM := math.Max(
		configuredHorizontal,
		horizontalSpreadM*
			baseline.config.
				NeighborSpreadMultiplier,
	)
	if !positiveFinite(
		horizontalUncertaintyM,
	) {
		return projectioncontract.
				ProjectionPoint{},
			false,
			ErrContinuationContractInvalid
	}

	position := projectioncontract.Position{
		Latitude:  latitude,
		Longitude: longitude,
	}
	uncertainty :=
		projectioncontract.Uncertainty{
			HorizontalRadiusM: horizontalUncertaintyM,
		}

	altitudeSampleCount := 0
	for _, sample := range samples {
		if sample.altitudeM != nil {
			altitudeSampleCount++
		}
	}
	altitudeAvailable :=
		altitudeSampleCount >=
			baseline.config.
				MinimumAltitudeSupport &&
			altitudeWeight > 0
	if altitudeAvailable {
		altitudeM :=
			weightedAltitude /
				altitudeWeight
		verticalSpreadSquared := 0.0
		for _, sample := range samples {
			if sample.altitudeM == nil {
				continue
			}
			delta :=
				*sample.altitudeM -
					altitudeM
			verticalSpreadSquared +=
				sample.weight *
					delta *
					delta
		}
		verticalSpreadM := math.Sqrt(
			verticalSpreadSquared /
				altitudeWeight,
		)
		configuredVertical :=
			baseline.config.
				InitialVerticalUncertaintyM +
				baseline.config.
					VerticalUncertaintyGrowthMPS*
					offsetSeconds
		verticalUncertaintyM := math.Max(
			configuredVertical,
			verticalSpreadM*
				baseline.config.
					NeighborSpreadMultiplier,
		)
		if finite(altitudeM) &&
			positiveFinite(
				verticalUncertaintyM,
			) {
			position.AltitudeM =
				float64Pointer(altitudeM)
			uncertainty.VerticalRadiusM =
				float64Pointer(
					verticalUncertaintyM,
				)
		} else {
			altitudeAvailable = false
		}
	}

	supportRatio := clampUnit(
		float64(len(samples)) /
			float64(
				pattern.NeighborCount,
			),
	)
	progress :=
		offsetSeconds /
			horizonSeconds
	score := pattern.Score *
		supportRatio *
		(1 -
			baseline.config.
				MaximumConfidenceLoss*
				progress)
	score = clampUnit(score)

	return projectioncontract.ProjectionPoint{
		Sequence:     sequence,
		ForecastTime: forecastTime.UTC(),
		Position:     position,
		Uncertainty:  uncertainty,
		Confidence: projectioncontract.Confidence{
			Score: score,
			Level: baseline.
				confidenceLevel(score),
			Reasons: []projectioncontract.
				ConfidenceReason{
				{
					Code:         "pattern_confidence_support_and_horizon_decay",
					Message:      "Point confidence combines pattern confidence, usable neighbor support, and configured horizon decay.",
					Contribution: score,
				},
			},
		},
	}, altitudeAvailable, nil
}

func (
	baseline *Baseline,
) fallback(
	request Request,
	reason string,
	selectionFingerprint string,
	patternFingerprint string,
) (projectioncontract.Result, error) {
	result, err := baseline.config.
		FallbackProjector.Project(
		projectionbaseline.Request{
			Trajectory:        request.CurrentTrajectory,
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
			GeneratedAt:       request.GeneratedAt,
		},
	)
	if err != nil {
		return projectioncontract.Result{},
			fmt.Errorf(
				"%w: %v",
				ErrFallbackProjectionFailed,
				err,
			)
	}

	result.Limitations = append(
		result.Limitations,
		projectioncontract.Limitation{
			Code:    "historical_neighbor_strategy_fallback",
			Message: "Historical-neighbor continuation was not usable; the result was produced by the conservative kinematic baseline.",
			Scope:   "method",
		},
		projectioncontract.Limitation{
			Code: "historical_neighbor_fallback_reason",
			Message: "Fallback reason: " +
				strings.TrimSpace(
					reason,
				) +
				".",
			Scope: "method",
		},
	)
	if result.Status !=
		projectioncontract.
			ResultStatusUnavailable {
		result.Explanations = append(
			result.Explanations,
			projectioncontract.Explanation{
				Code:    "kinematic_fallback_selected",
				Message: "Historical pattern evidence was unavailable or insufficient, so the deterministic kinematic baseline was selected.",
			},
		)
	}

	latestObservedAt :=
		result.Provenance.
			LatestInputObservedAt
	result.Provenance.Inputs = append(
		result.Provenance.Inputs,
		projectioncontract.InputReference{
			Name: "historical_neighbor_strategy_decision",
			Classification: projectioncontract.
				InputClassificationDerived,
			SourceName: "projectioncontinuation",
			ObservedAt: latestObservedAt,
			Limitation: reason,
		},
	)
	result.Provenance.InputFingerprint =
		fallbackFingerprint(
			result.Provenance.
				InputFingerprint,
			reason,
			selectionFingerprint,
			patternFingerprint,
		)
	result.Limitations =
		normalizeLimitations(
			result.Limitations,
		)

	return validateProjectionResult(
		result,
	)
}

func buildCandidateIndex(
	items []trajectory.FlightTrajectory,
	asOfTime time.Time,
) map[string]trajectory.FlightTrajectory {
	result := make(
		map[string]trajectory.FlightTrajectory,
		len(items),
	)
	duplicateIDs := make(map[string]bool)

	for _, item := range items {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		if _, exists := result[id]; exists {
			duplicateIDs[id] = true
			delete(result, id)
			continue
		}
		if duplicateIDs[id] {
			continue
		}

		result[id] =
			trajectorySnapshotAt(
				item,
				asOfTime,
			)
	}

	return result
}

func historicalContinuationLimitations(
	selection projectionneighbors.Result,
	pattern projectionpatternconfidence.Result,
) []projectioncontract.Limitation {
	result := []projectioncontract.Limitation{
		{
			Code:    "historical_neighbor_continuation_experimental",
			Message: "Historical-neighbor continuation is project-derived and experimental until calibrated by replay.",
			Scope:   "method",
		},
		{
			Code:    "historical_behavior_not_intent",
			Message: "Historical continuation patterns do not represent official flight plans, Air Traffic Control instructions, pilot intent, or guaranteed future maneuvers.",
			Scope:   "method",
		},
		{
			Code:    "no_weather_adjustment",
			Message: "Weather and wind are not applied by this continuation baseline.",
			Scope:   "method",
		},
		{
			Code:    "research_only",
			Message: "Projection is a research estimate and must not be used for operational aviation decisions.",
			Scope:   "result",
		},
	}

	for _, limitation := range selection.Limitations {
		result = append(
			result,
			projectioncontract.Limitation{
				Code: "neighbor_selection_" +
					limitation.Code,
				Message: limitation.Message,
				Scope:   "selection",
			},
		)
	}
	for _, limitation := range pattern.Limitations {
		result = append(
			result,
			projectioncontract.Limitation{
				Code: "pattern_confidence_" +
					limitation.Code,
				Message: limitation.Message,
				Scope:   "confidence",
			},
		)
	}

	return result
}

func continuationInputs(
	currentEndpoint trajectory.TrackPoint4D,
	selection projectionneighbors.Result,
) []projectioncontract.InputReference {
	result := []projectioncontract.InputReference{
		{
			Name: "current_trajectory_endpoint",
			Classification: projectioncontract.
				InputClassificationObserved,
			SourceName: currentEndpoint.SourceName,
			ObservedAt: currentEndpoint.
				ObservedAt.UTC(),
		},
		{
			Name: "historical_neighbor_selection",
			Classification: projectioncontract.
				InputClassificationDerived,
			SourceName: "projectionneighbors",
			ObservedAt: currentEndpoint.
				ObservedAt.UTC(),
		},
		{
			Name: "historical_pattern_confidence",
			Classification: projectioncontract.
				InputClassificationDerived,
			SourceName: "projectionpatternconfidence",
			ObservedAt: currentEndpoint.
				ObservedAt.UTC(),
		},
	}

	for _, neighbor := range selection.Neighbors {
		result = append(
			result,
			projectioncontract.InputReference{
				Name: "historical_neighbor:" +
					neighbor.
						TrajectoryID,
				Classification: projectioncontract.
					InputClassificationDerived,
				SourceName: "historical_trajectory",
				ObservedAt: neighbor.
					CandidateEndTime.UTC(),
			},
		)
	}

	return result
}

func patternMatchesSelection(
	pattern projectionpatternconfidence.Result,
	selection projectionneighbors.Result,
) bool {
	if pattern.NeighborCount !=
		len(selection.Neighbors) {
		return false
	}

	selected := make(
		[]string,
		0,
		len(selection.Neighbors),
	)
	for _, neighbor := range selection.Neighbors {
		selected = append(
			selected,
			strings.TrimSpace(
				neighbor.TrajectoryID,
			),
		)
	}
	sort.Strings(selected)

	if len(selected) !=
		len(pattern.SelectedTrajectoryIDs) {
		return false
	}
	for index := range selected {
		if selected[index] !=
			pattern.SelectedTrajectoryIDs[index] {
			return false
		}
	}

	return true
}

func minimumPointConfidence(
	points []projectioncontract.ProjectionPoint,
) projectioncontract.Confidence {
	if len(points) == 0 {
		return projectioncontract.Confidence{
			Score: 0,
			Level: projectioncontract.
				ConfidenceLevelNone,
		}
	}

	minimum := points[0].Confidence
	for _, point := range points[1:] {
		if point.Confidence.Score <
			minimum.Score {
			minimum = point.Confidence
		}
	}

	minimum.Reasons =
		[]projectioncontract.ConfidenceReason{
			{
				Code:         "minimum_historical_continuation_point_confidence",
				Message:      "Result confidence equals the lowest confidence across historical-continuation forecast points.",
				Contribution: minimum.Score,
			},
		}

	return minimum
}

func (
	baseline *Baseline,
) confidenceLevel(
	score float64,
) projectioncontract.ConfidenceLevel {
	switch {
	case score >= baseline.config.
		HighConfidenceMinimum:
		return projectioncontract.
			ConfidenceLevelHigh
	case score >= baseline.config.
		MediumConfidenceMinimum:
		return projectioncontract.
			ConfidenceLevelMedium
	case score > 0:
		return projectioncontract.
			ConfidenceLevelLow
	default:
		return projectioncontract.
			ConfidenceLevelNone
	}
}

func normalizeLimitations(
	items []projectioncontract.Limitation,
) []projectioncontract.Limitation {
	seen := make(
		map[string]projectioncontract.Limitation,
		len(items),
	)
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		message :=
			strings.TrimSpace(item.Message)
		scope := strings.TrimSpace(item.Scope)
		if code == "" ||
			message == "" ||
			scope == "" {
			continue
		}

		key := code + "\x00" +
			message + "\x00" +
			scope
		seen[key] =
			projectioncontract.Limitation{
				Code:    code,
				Message: message,
				Scope:   scope,
			}
	}

	keys := make(
		[]string,
		0,
		len(seen),
	)
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(
		[]projectioncontract.Limitation,
		0,
		len(keys),
	)
	for _, key := range keys {
		result = append(
			result,
			seen[key],
		)
	}

	return result
}

func validateProjectionResult(
	result projectioncontract.Result,
) (projectioncontract.Result, error) {
	report := projectioncontract.Validate(
		result,
	)
	if report.Status !=
		projectioncontract.
			ValidationStatusValid {
		return projectioncontract.Result{},
			fmt.Errorf(
				"%w: %#v",
				ErrContinuationContractInvalid,
				report.Issues,
			)
	}

	return result.Clone(), nil
}

func clampUnit(value float64) float64 {
	if !finite(value) ||
		value <= 0 {
		return 0
	}
	if value >= 1 {
		return 1
	}

	return value
}
