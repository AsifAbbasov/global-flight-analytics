package projectionarrival

import (
	"math"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func (
	estimator *Estimator,
) computeArrival(
	samples []positionSample,
	destinationLatitude float64,
	destinationLongitude float64,
	projection projectioncontract.Result,
) (arrivalComputation, bool) {
	distances, valid := arrivalDistances(
		samples,
		destinationLatitude,
		destinationLongitude,
	)
	if !valid {
		return arrivalComputation{}, false
	}

	profile, profileAvailable :=
		calculateSpeedProfile(
			samples,
			estimator.config.
				MinimumGroundSpeedMPS,
			estimator.config.
				MaximumSpeedSampleCount,
		)

	if computation, exists :=
		estimator.arrivalWithinProjection(
			samples,
			distances,
			profile,
			profileAvailable,
			projection,
		); exists {
		return computation, true
	}

	return estimator.extrapolatedArrival(
		samples,
		distances,
		profile,
		profileAvailable,
		projection,
	)
}

func arrivalDistances(
	samples []positionSample,
	destinationLatitude float64,
	destinationLongitude float64,
) ([]float64, bool) {
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
			return nil, false
		}
		distances[index] = distanceM
	}

	return distances, true
}

func (
	estimator *Estimator,
) arrivalWithinProjection(
	samples []positionSample,
	distances []float64,
	profile speedProfile,
	profileAvailable bool,
	projection projectioncontract.Result,
) (arrivalComputation, bool) {
	for index, distanceM := range distances {
		if index > 0 {
			if computation, exists :=
				estimator.arrivalAtRadiusCrossing(
					samples[index-1],
					samples[index],
					distances[index-1],
					distanceM,
					profile,
					profileAvailable,
					projection,
				); exists {
				return computation, true
			}
		}

		if computation, exists :=
			estimator.arrivalInsideRadius(
				samples[index],
				distanceM,
				profile,
				profileAvailable,
				projection,
			); exists {
			return computation, true
		}
	}

	return arrivalComputation{}, false
}

func (
	estimator *Estimator,
) arrivalAtRadiusCrossing(
	previous positionSample,
	current positionSample,
	previousDistanceM float64,
	currentDistanceM float64,
	profile speedProfile,
	profileAvailable bool,
	projection projectioncontract.Result,
) (arrivalComputation, bool) {
	if previousDistanceM <=
		estimator.config.ArrivalRadiusM ||
		currentDistanceM >
			estimator.config.ArrivalRadiusM ||
		currentDistanceM >= previousDistanceM {
		return arrivalComputation{}, false
	}

	denominator :=
		previousDistanceM - currentDistanceM
	if denominator <= 0 {
		return arrivalComputation{}, false
	}
	fraction :=
		(previousDistanceM -
			estimator.config.ArrivalRadiusM) /
			denominator
	fraction = math.Max(
		0,
		math.Min(1, fraction),
	)

	segmentDuration :=
		current.timeValue.Sub(
			previous.timeValue,
		)
	if segmentDuration <= 0 {
		return arrivalComputation{}, false
	}

	estimatedTime :=
		previous.timeValue.Add(
			time.Duration(
				fraction *
					float64(segmentDuration),
			),
		)
	segmentDistanceM :=
		greatCircleDistanceM(
			previous.latitude,
			previous.longitude,
			current.latitude,
			current.longitude,
		)
	segmentSpeedMPS :=
		segmentDistanceM /
			segmentDuration.Seconds()
	if !positiveFinite(segmentSpeedMPS) {
		return arrivalComputation{}, false
	}

	uncertaintyM :=
		previous.horizontalUncertaintyM +
			fraction*
				(current.horizontalUncertaintyM-
					previous.horizontalUncertaintyM)
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
		speedStdDevMPS = profile.stdDevMPS
		speedSampleCount = profile.sampleCount
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

func (
	estimator *Estimator,
) arrivalInsideRadius(
	sample positionSample,
	distanceM float64,
	profile speedProfile,
	profileAvailable bool,
	projection projectioncontract.Result,
) (arrivalComputation, bool) {
	if distanceM >
		estimator.config.ArrivalRadiusM {
		return arrivalComputation{}, false
	}

	estimatedTime := sample.timeValue.UTC()
	if estimatedTime.Before(
		projection.Horizon.AsOfTime.UTC(),
	) {
		estimatedTime =
			projection.Horizon.AsOfTime.UTC()
	}

	speedMPS :=
		estimator.config.MinimumGroundSpeedMPS
	speedStdDevMPS := 0.0
	speedSampleCount := 0
	if profileAvailable {
		speedMPS = profile.meanMPS
		speedStdDevMPS = profile.stdDevMPS
		speedSampleCount = profile.sampleCount
	}

	uncertaintyDuration :=
		time.Duration(
			sample.horizontalUncertaintyM /
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
		lastPositionUncertaintyM: sample.
			horizontalUncertaintyM,
	}, true
}

func (
	estimator *Estimator,
) extrapolatedArrival(
	samples []positionSample,
	distances []float64,
	profile speedProfile,
	profileAvailable bool,
	projection projectioncontract.Result,
) (arrivalComputation, bool) {
	if !profileAvailable ||
		profile.sampleCount <
			estimator.config.
				MinimumSpeedSampleCount {
		return arrivalComputation{}, false
	}

	lastSample := samples[len(samples)-1]
	lastDistanceM :=
		distances[len(distances)-1]
	remainingDistanceM := math.Max(
		0,
		lastDistanceM-
			estimator.config.ArrivalRadiusM,
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
		estimator.config.MinimumGroundSpeedMPS,
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
			lastSample.horizontalUncertaintyM,
	)
	latestDistanceM :=
		remainingDistanceM +
			lastSample.horizontalUncertaintyM

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
			projection.Horizon.EndTime.UTC(),
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
