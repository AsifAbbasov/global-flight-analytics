package projectionarrival

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

var (
	ErrProjectionContractInvalid = errors.New(
		"projection contract is invalid",
	)
	ErrRouteContractInvalid = errors.New(
		"route contract is invalid",
	)
	ErrTrajectoryMismatch = errors.New(
		"projection, route, and current trajectory identifiers must match",
	)
	ErrFutureRouteEvidence = errors.New(
		"route evidence as-of time must not exceed projection as-of time",
	)
	ErrGeneratedAtInvalid = errors.New(
		"arrival generated-at time must not precede its inputs",
	)
	ErrArrivalContractInvalid = errors.New(
		"generated arrival projection contract is invalid",
	)
)

type Estimator struct {
	config Config
}

func New(
	config Config,
) (*Estimator, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate estimated arrival config: %w",
			err,
		)
	}

	return &Estimator{
		config: config,
	}, nil
}

type Request struct {
	Projection        projectioncontract.Result
	Route             routecontract.Result
	CurrentTrajectory trajectory.FlightTrajectory
	GeneratedAt       time.Time
}

func (
	estimator *Estimator,
) Estimate(
	request Request,
) (projectioncontract.Result, error) {
	if estimator == nil {
		return projectioncontract.Result{},
			ErrArrivalContractInvalid
	}

	projectionReport :=
		projectioncontract.Validate(
			request.Projection,
		)
	if projectionReport.Status !=
		projectioncontract.
			ValidationStatusValid {
		return projectioncontract.Result{},
			fmt.Errorf(
				"%w: %#v",
				ErrProjectionContractInvalid,
				projectionReport.Issues,
			)
	}

	routeReport :=
		routecontract.Validate(
			request.Route,
		)
	if routeReport.Status !=
		routecontract.
			ValidationStatusValid {
		return projectioncontract.Result{},
			fmt.Errorf(
				"%w: %#v",
				ErrRouteContractInvalid,
				routeReport.Issues,
			)
	}

	if request.Route.TrajectoryID !=
		request.Projection.TrajectoryID ||
		(strings.TrimSpace(
			request.CurrentTrajectory.ID,
		) != "" &&
			request.CurrentTrajectory.ID !=
				request.Projection.
					TrajectoryID) {
		return projectioncontract.Result{},
			ErrTrajectoryMismatch
	}

	projectionAsOf :=
		request.Projection.
			Horizon.AsOfTime.UTC()
	routeAsOf :=
		request.Route.Window.
			AsOfTime.UTC()
	if routeAsOf.After(projectionAsOf) {
		return projectioncontract.Result{},
			ErrFutureRouteEvidence
	}

	generatedAt :=
		request.GeneratedAt.UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(
			request.Projection.
				GeneratedAt.UTC(),
		) ||
		generatedAt.Before(
			request.Route.
				GeneratedAt.UTC(),
		) ||
		generatedAt.Before(
			projectionAsOf,
		) {
		return projectioncontract.Result{},
			ErrGeneratedAtInvalid
	}

	if request.Projection.Status ==
		projectioncontract.
			ResultStatusUnavailable {
		return estimator.withUnavailableArrival(
			request,
			"projection_unavailable",
			"Estimated arrival is unavailable because the position projection is unavailable.",
		)
	}

	if request.Route.Destination == nil {
		return estimator.withUnavailableArrival(
			request,
			"destination_unavailable",
			"Estimated arrival is unavailable because Route Intelligence did not resolve a destination airport.",
		)
	}

	destination :=
		request.Route.Destination
	if destination.Confidence.Score <
		estimator.config.
			MinimumDestinationConfidenceScore {
		return estimator.withUnavailableArrival(
			request,
			"destination_confidence_below_minimum",
			fmt.Sprintf(
				"Estimated arrival is withheld because destination confidence %.6f is below the configured minimum %.6f.",
				destination.Confidence.Score,
				estimator.config.
					MinimumDestinationConfidenceScore,
			),
		)
	}

	samples := buildPositionSamples(
		request.CurrentTrajectory,
		request.Projection,
	)
	if len(samples) == 0 {
		return estimator.withUnavailableArrival(
			request,
			"arrival_position_samples_unavailable",
			"Estimated arrival is unavailable because no usable current or projected position samples were available.",
		)
	}

	computation, exists :=
		estimator.computeArrival(
			samples,
			destination.Airport.Latitude,
			destination.Airport.Longitude,
			request.Projection,
		)
	if !exists {
		return estimator.withUnavailableArrival(
			request,
			"arrival_speed_or_duration_unavailable",
			"Estimated arrival is unavailable because the projected speed profile or bounded arrival duration was not usable.",
		)
	}

	result := request.Projection.Clone()
	arrivalConfidence :=
		estimator.arrivalConfidence(
			request.Projection,
			destination.Confidence.Score,
			computation,
		)

	result.Arrival =
		&projectioncontract.ArrivalEstimate{
			AirportICAOCode: strings.TrimSpace(
				destination.
					Airport.ICAOCode,
			),
			EarliestTime: computation.
				earliestTime.UTC(),
			EstimatedTime: computation.
				estimatedTime.UTC(),
			LatestTime: computation.
				latestTime.UTC(),
			Confidence: arrivalConfidence,
			Limitations: arrivalLimitations(
				computation.mode,
				request.Route.Status,
			),
		}

	if result.Status ==
		projectioncontract.
			ResultStatusComplete &&
		(computation.mode ==
			EstimateModeExtrapolated ||
			request.Route.Status !=
				routecontract.
					RouteStatusComplete) {
		result.Status =
			projectioncontract.
				ResultStatusLimited
	}

	result.Confidence =
		estimator.combinedConfidence(
			result.Confidence,
			arrivalConfidence,
		)
	result.Limitations =
		normalizeLimitations(
			append(
				result.Limitations,
				projectioncontract.Limitation{
					Code:    "estimated_arrival_boundary_attached",
					Message: "Projection includes an estimated airport-radius arrival interval.",
					Scope:   "arrival",
				},
			),
		)
	result.Explanations =
		normalizeExplanations(
			append(
				result.Explanations,
				projectioncontract.Explanation{
					Code:    MethodName,
					Message: "Estimated arrival is based on destination inference, projected position samples, and a bounded projected ground-speed profile.",
				},
			),
		)

	result.Provenance.Inputs = append(
		result.Provenance.Inputs,
		projectioncontract.InputReference{
			Name: "route_destination_inference",
			Classification: projectioncontract.
				InputClassificationDerived,
			SourceName:  "routeintelligence",
			ObservedAt:  routeAsOf,
			RetrievedAt: generatedAt,
		},
		projectioncontract.InputReference{
			Name: "projected_arrival_speed_profile",
			Classification: projectioncontract.
				InputClassificationEstimated,
			SourceName:  "projectionarrival",
			ObservedAt:  projectionAsOf,
			RetrievedAt: generatedAt,
			Limitation:  "Ground speed is derived from estimated projection points and is not an official flight-plan speed.",
		},
	)
	result.Provenance.Inputs =
		normalizeInputs(
			result.Provenance.Inputs,
		)
	result.Provenance.
		LatestInputObservedAt =
		latestInputObservedAt(
			result.Provenance.Inputs,
		)
	result.Provenance.InputFingerprint =
		arrivalFingerprint(
			request.Projection,
			request.Route,
			computation,
			estimator.config,
		)
	result.GeneratedAt = generatedAt

	return validateResult(result)
}

func (
	estimator *Estimator,
) computeArrival(
	samples []positionSample,
	destinationLatitude float64,
	destinationLongitude float64,
	projection projectioncontract.Result,
) (arrivalComputation, bool) {
	distances := make(
		[]float64,
		len(samples),
	)
	for index, sample := range samples {
		distanceM := greatCircleDistanceM(
			sample.latitude,
			sample.longitude,
			destinationLatitude,
			destinationLongitude,
		)
		if !nonNegativeFinite(distanceM) {
			return arrivalComputation{},
				false
		}
		distances[index] = distanceM
	}

	profile, profileAvailable :=
		calculateSpeedProfile(
			samples,
			estimator.config.
				MinimumGroundSpeedMPS,
			estimator.config.
				MaximumSpeedSampleCount,
		)

	for index, distanceM := range distances {
		if index > 0 &&
			distances[index-1] >
				estimator.config.
					ArrivalRadiusM &&
			distanceM <=
				estimator.config.
					ArrivalRadiusM &&
			distanceM <
				distances[index-1] {
			denominator :=
				distances[index-1] -
					distanceM
			if denominator > 0 {
				fraction :=
					(distances[index-1] -
						estimator.config.
							ArrivalRadiusM) /
						denominator
				fraction = math.Max(
					0,
					math.Min(1, fraction),
				)

				segmentDuration :=
					samples[index].
						timeValue.Sub(
						samples[index-1].
							timeValue,
					)
				if segmentDuration > 0 {
					estimatedTime :=
						samples[index-1].
							timeValue.Add(
							time.Duration(
								fraction *
									float64(
										segmentDuration,
									),
							),
						)

					segmentDistanceM :=
						greatCircleDistanceM(
							samples[index-1].
								latitude,
							samples[index-1].
								longitude,
							samples[index].
								latitude,
							samples[index].
								longitude,
						)
					segmentSpeedMPS :=
						segmentDistanceM /
							segmentDuration.Seconds()
					if positiveFinite(
						segmentSpeedMPS,
					) {
						uncertaintyM :=
							samples[index-1].
								horizontalUncertaintyM +
								fraction*
									(samples[index].
										horizontalUncertaintyM-
										samples[index-1].
											horizontalUncertaintyM)
						uncertaintyDuration :=
							time.Duration(
								uncertaintyM /
									segmentSpeedMPS *
									float64(time.Second),
							)
						earliestTime :=
							estimatedTime.Add(
								-uncertaintyDuration,
							)
						latestTime :=
							estimatedTime.Add(
								uncertaintyDuration,
							)
						earliestTime,
							estimatedTime,
							latestTime =
							enforceMinimumArrivalInterval(
								projection.Horizon.
									AsOfTime.UTC(),
								estimatedTime,
								earliestTime,
								latestTime,
								estimator.config.
									MinimumArrivalInterval,
							)

						speedStdDevMPS := 0.0
						speedSampleCount := 1
						if profileAvailable {
							speedStdDevMPS =
								profile.stdDevMPS
							speedSampleCount =
								profile.sampleCount
						}

						return arrivalComputation{
							mode:                     EstimateModeWithinProjection,
							earliestTime:             earliestTime,
							estimatedTime:            estimatedTime,
							latestTime:               latestTime,
							estimatedGroundSpeedMPS:  segmentSpeedMPS,
							groundSpeedStdDevMPS:     speedStdDevMPS,
							speedSampleCount:         speedSampleCount,
							remainingDistanceM:       0,
							lastPositionUncertaintyM: uncertaintyM,
						}, true
					}
				}
			}
		}

		if distanceM <=
			estimator.config.ArrivalRadiusM {
			estimatedTime :=
				samples[index].
					timeValue.UTC()
			if estimatedTime.Before(
				projection.Horizon.
					AsOfTime.UTC(),
			) {
				estimatedTime =
					projection.Horizon.
						AsOfTime.UTC()
			}

			speedMPS :=
				estimator.config.
					MinimumGroundSpeedMPS
			speedStdDevMPS := 0.0
			speedSampleCount := 0
			if profileAvailable {
				speedMPS =
					profile.meanMPS
				speedStdDevMPS =
					profile.stdDevMPS
				speedSampleCount =
					profile.sampleCount
			}

			uncertaintyDuration :=
				time.Duration(
					samples[index].
						horizontalUncertaintyM /
						speedMPS *
						float64(time.Second),
				)
			earliestTime :=
				estimatedTime.Add(
					-uncertaintyDuration,
				)
			latestTime :=
				estimatedTime.Add(
					uncertaintyDuration,
				)
			earliestTime,
				estimatedTime,
				latestTime =
				enforceMinimumArrivalInterval(
					projection.Horizon.
						AsOfTime.UTC(),
					estimatedTime,
					earliestTime,
					latestTime,
					estimator.config.
						MinimumArrivalInterval,
				)

			return arrivalComputation{
				mode:                    EstimateModeWithinProjection,
				earliestTime:            earliestTime,
				estimatedTime:           estimatedTime,
				latestTime:              latestTime,
				estimatedGroundSpeedMPS: speedMPS,
				groundSpeedStdDevMPS:    speedStdDevMPS,
				speedSampleCount:        speedSampleCount,
				remainingDistanceM:      0,
				lastPositionUncertaintyM: samples[index].
					horizontalUncertaintyM,
			}, true
		}

	}

	if !profileAvailable ||
		profile.sampleCount <
			estimator.config.
				MinimumSpeedSampleCount {
		return arrivalComputation{}, false
	}

	lastSample :=
		samples[len(samples)-1]
	lastDistanceM :=
		distances[len(distances)-1]
	remainingDistanceM := math.Max(
		0,
		lastDistanceM-
			estimator.config.
				ArrivalRadiusM,
	)
	estimatedDuration :=
		time.Duration(
			remainingDistanceM /
				profile.meanMPS *
				float64(time.Second),
		)
	if estimatedDuration >
		estimator.config.
			MaximumEstimatedArrivalDuration {
		return arrivalComputation{}, false
	}

	lowerSpeedMPS := math.Max(
		estimator.config.
			MinimumGroundSpeedMPS,
		profile.meanMPS-
			estimator.config.
				SpeedUncertaintyMultiplier*
				profile.stdDevMPS,
	)
	upperSpeedMPS :=
		profile.meanMPS +
			estimator.config.
				SpeedUncertaintyMultiplier*
				profile.stdDevMPS
	if !positiveFinite(lowerSpeedMPS) ||
		!positiveFinite(upperSpeedMPS) {
		return arrivalComputation{}, false
	}

	earliestDistanceM := math.Max(
		0,
		remainingDistanceM-
			lastSample.
				horizontalUncertaintyM,
	)
	latestDistanceM :=
		remainingDistanceM +
			lastSample.
				horizontalUncertaintyM

	earliestTime :=
		lastSample.timeValue.Add(
			time.Duration(
				earliestDistanceM /
					upperSpeedMPS *
					float64(time.Second),
			),
		)
	estimatedTime :=
		lastSample.timeValue.Add(
			estimatedDuration,
		)
	latestTime :=
		lastSample.timeValue.Add(
			time.Duration(
				latestDistanceM /
					lowerSpeedMPS *
					float64(time.Second),
			),
		)
	earliestTime,
		estimatedTime,
		latestTime =
		enforceMinimumArrivalInterval(
			projection.Horizon.
				AsOfTime.UTC(),
			estimatedTime,
			earliestTime,
			latestTime,
			estimator.config.
				MinimumArrivalInterval,
		)

	extrapolationDuration :=
		estimatedTime.Sub(
			projection.Horizon.
				EndTime.UTC(),
		)
	if extrapolationDuration < 0 {
		extrapolationDuration = 0
	}

	return arrivalComputation{
		mode:                    EstimateModeExtrapolated,
		earliestTime:            earliestTime,
		estimatedTime:           estimatedTime,
		latestTime:              latestTime,
		estimatedGroundSpeedMPS: profile.meanMPS,
		groundSpeedStdDevMPS:    profile.stdDevMPS,
		speedSampleCount:        profile.sampleCount,
		remainingDistanceM:      remainingDistanceM,
		lastPositionUncertaintyM: lastSample.
			horizontalUncertaintyM,
		extrapolationDuration: extrapolationDuration,
	}, true
}

func (
	estimator *Estimator,
) arrivalConfidence(
	projection projectioncontract.Result,
	destinationConfidenceScore float64,
	computation arrivalComputation,
) projectioncontract.Confidence {
	speedSupport := math.Min(
		1,
		float64(
			computation.
				speedSampleCount,
		)/
			float64(
				estimator.config.
					MinimumSpeedSampleCount,
			),
	)
	speedStability := 0.0
	if positiveFinite(
		computation.
			estimatedGroundSpeedMPS,
	) {
		speedStability =
			1 -
				math.Min(
					1,
					computation.
						groundSpeedStdDevMPS/
						computation.
							estimatedGroundSpeedMPS,
				)
		speedStability *=
			speedSupport
	}
	speedStability =
		clampUnit(speedStability)

	score :=
		estimator.config.
			ProjectionConfidenceWeight*
			projection.Confidence.Score +
			estimator.config.
				DestinationConfidenceWeight*
				destinationConfidenceScore +
			estimator.config.
				SpeedStabilityWeight*
				speedStability

	extrapolationRatio := 0.0
	if computation.
		extrapolationDuration > 0 {
		extrapolationRatio =
			math.Min(
				1,
				float64(
					computation.
						extrapolationDuration,
				)/
					float64(
						estimator.config.
							MaximumEstimatedArrivalDuration,
					),
			)
	}
	score *= 1 -
		estimator.config.
			MaximumExtrapolationConfidenceLoss*
			extrapolationRatio
	score = clampUnit(score)

	return projectioncontract.Confidence{
		Score: score,
		Level: estimator.confidenceLevel(
			score,
		),
		Reasons: []projectioncontract.ConfidenceReason{
			{
				Code:    "position_projection_confidence",
				Message: "Arrival confidence includes the position-projection confidence.",
				Contribution: estimator.config.
					ProjectionConfidenceWeight *
					projection.
						Confidence.Score,
			},
			{
				Code:    "destination_inference_confidence",
				Message: "Arrival confidence includes Route Intelligence destination confidence.",
				Contribution: estimator.config.
					DestinationConfidenceWeight *
					destinationConfidenceScore,
			},
			{
				Code:    "projected_speed_stability",
				Message: "Arrival confidence includes projected ground-speed stability and sample support.",
				Contribution: estimator.config.
					SpeedStabilityWeight *
					speedStability,
			},
			{
				Code:    "extrapolation_confidence_decay",
				Message: "Arrival confidence decreases when the estimate extends beyond the position-projection horizon.",
				Contribution: -estimator.config.
					MaximumExtrapolationConfidenceLoss *
					extrapolationRatio,
			},
		},
	}
}

func (
	estimator *Estimator,
) confidenceLevel(
	score float64,
) projectioncontract.ConfidenceLevel {
	switch {
	case score >= estimator.config.
		HighConfidenceMinimum:
		return projectioncontract.
			ConfidenceLevelHigh
	case score >= estimator.config.
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

func (
	estimator *Estimator,
) withUnavailableArrival(
	request Request,
	reason string,
	message string,
) (projectioncontract.Result, error) {
	result := request.Projection.Clone()
	result.Arrival = nil

	if result.Status ==
		projectioncontract.
			ResultStatusComplete {
		result.Status =
			projectioncontract.
				ResultStatusLimited
	}

	result.Limitations =
		normalizeLimitations(
			append(
				result.Limitations,
				projectioncontract.Limitation{
					Code: "estimated_arrival_unavailable",
					Message: strings.TrimSpace(
						message,
					),
					Scope: "arrival",
				},
				projectioncontract.Limitation{
					Code: "estimated_arrival_unavailable_reason",
					Message: "Estimated arrival reason: " +
						strings.TrimSpace(
							reason,
						) +
						".",
					Scope: "arrival",
				},
			),
		)

	if result.Status !=
		projectioncontract.
			ResultStatusUnavailable {
		result.Explanations =
			normalizeExplanations(
				append(
					result.Explanations,
					projectioncontract.Explanation{
						Code:    "estimated_arrival_withheld",
						Message: "Estimated arrival was withheld rather than publishing an unsupported interval.",
					},
				),
			)
	}

	routeFingerprint :=
		request.Route.Provenance.
			InputFingerprint
	result.Provenance.InputFingerprint =
		unavailableFingerprint(
			result.Provenance.
				InputFingerprint,
			routeFingerprint,
			reason,
			estimator.config,
		)

	if !request.Route.Window.
		AsOfTime.IsZero() &&
		!request.Route.Window.
			AsOfTime.After(
			result.Horizon.AsOfTime,
		) {
		result.Provenance.Inputs = append(
			result.Provenance.Inputs,
			projectioncontract.InputReference{
				Name: "route_destination_inference",
				Classification: projectioncontract.
					InputClassificationDerived,
				SourceName: "routeintelligence",
				ObservedAt: request.Route.Window.
					AsOfTime.UTC(),
				RetrievedAt: request.GeneratedAt.UTC(),
				Limitation:  reason,
			},
		)
		result.Provenance.Inputs =
			normalizeInputs(
				result.Provenance.Inputs,
			)
		result.Provenance.
			LatestInputObservedAt =
			latestInputObservedAt(
				result.Provenance.Inputs,
			)
	}

	result.GeneratedAt =
		request.GeneratedAt.UTC()

	return validateResult(result)
}

func arrivalLimitations(
	mode EstimateMode,
	routeStatus routecontract.RouteStatus,
) []projectioncontract.Limitation {
	result := []projectioncontract.Limitation{
		{
			Code:    "arrival_radius_not_touchdown",
			Message: "Estimated arrival represents entry into the configured airport radius, not runway touchdown or gate arrival.",
			Scope:   "arrival",
		},
		{
			Code:    "destination_is_inferred",
			Message: "Destination airport is inferred by Route Intelligence and is not an official flight-plan destination.",
			Scope:   "arrival",
		},
		{
			Code:    "no_operational_arrival_intent",
			Message: "Official flight plan, Air Traffic Control sequence, runway assignment, holding, diversion, and pilot intent are unavailable.",
			Scope:   "arrival",
		},
		{
			Code:    "no_weather_arrival_adjustment",
			Message: "Weather and wind are not applied to the estimated arrival interval.",
			Scope:   "arrival",
		},
		{
			Code:    "research_only_arrival",
			Message: "Estimated arrival is a research output and must not be used for operational aviation decisions.",
			Scope:   "arrival",
		},
	}

	if mode == EstimateModeExtrapolated {
		result = append(
			result,
			projectioncontract.Limitation{
				Code:    "arrival_extrapolated_beyond_projection_horizon",
				Message: "Estimated arrival extends beyond the position-projection horizon using a bounded projected ground-speed profile.",
				Scope:   "arrival",
			},
		)
	}
	if routeStatus !=
		routecontract.RouteStatusComplete {
		result = append(
			result,
			projectioncontract.Limitation{
				Code:    "route_intelligence_partial",
				Message: "Route Intelligence resolved a destination without a complete two-endpoint route.",
				Scope:   "arrival",
			},
		)
	}

	return normalizeLimitations(result)
}

func (
	estimator *Estimator,
) combinedConfidence(
	projectionConfidence projectioncontract.Confidence,
	arrivalConfidence projectioncontract.Confidence,
) projectioncontract.Confidence {
	score := math.Min(
		projectionConfidence.Score,
		arrivalConfidence.Score,
	)

	level :=
		estimator.confidenceLevel(
			score,
		)

	return projectioncontract.Confidence{
		Score: score,
		Level: level,
		Reasons: []projectioncontract.ConfidenceReason{
			{
				Code:         "combined_projection_and_arrival_confidence",
				Message:      "Overall result confidence equals the weaker confidence between position projection and estimated arrival.",
				Contribution: score,
			},
		},
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

func normalizeExplanations(
	items []projectioncontract.Explanation,
) []projectioncontract.Explanation {
	seen := make(
		map[string]projectioncontract.Explanation,
		len(items),
	)
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		message :=
			strings.TrimSpace(item.Message)
		if code == "" ||
			message == "" {
			continue
		}
		key := code + "\x00" +
			message
		seen[key] =
			projectioncontract.Explanation{
				Code:    code,
				Message: message,
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
		[]projectioncontract.Explanation,
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

func normalizeInputs(
	items []projectioncontract.InputReference,
) []projectioncontract.InputReference {
	type indexedInput struct {
		item  projectioncontract.InputReference
		index int
	}

	seen := make(
		map[string]indexedInput,
		len(items),
	)
	for index, item := range items {
		key :=
			strings.TrimSpace(item.Name) +
				"\x00" +
				string(item.Classification) +
				"\x00" +
				strings.TrimSpace(
					item.SourceName,
				) +
				"\x00" +
				item.ObservedAt.UTC().
					Format(time.RFC3339Nano)
		seen[key] = indexedInput{
			item:  item,
			index: index,
		}
	}

	values := make(
		[]indexedInput,
		0,
		len(seen),
	)
	for _, value := range seen {
		values = append(values, value)
	}
	sort.SliceStable(
		values,
		func(left int, right int) bool {
			return values[left].index <
				values[right].index
		},
	)

	result := make(
		[]projectioncontract.InputReference,
		0,
		len(values),
	)
	for _, value := range values {
		result = append(
			result,
			value.item,
		)
	}

	return result
}

func latestInputObservedAt(
	items []projectioncontract.InputReference,
) time.Time {
	var latest time.Time
	for _, item := range items {
		observedAt :=
			item.ObservedAt.UTC()
		if item.ObservedAt.IsZero() {
			continue
		}
		if latest.IsZero() ||
			observedAt.After(latest) {
			latest = observedAt
		}
	}

	return latest
}

func validateResult(
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
				ErrArrivalContractInvalid,
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
