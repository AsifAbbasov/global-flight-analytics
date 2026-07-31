package projectionarrival

import (
	"math"

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
		calculateClosingSpeedProfile(
			samples,
			distances,
			estimator.config.MaximumGroundSpeedMPS,
			estimator.config.MaximumSpeedSampleCount,
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
	if len(samples) == 0 ||
		!validLatitude(destinationLatitude) ||
		!validLongitude(destinationLongitude) {
		return nil, false
	}

	distances := make([]float64, len(samples))
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
	if previousDistanceM <= estimator.config.ArrivalRadiusM ||
		currentDistanceM > estimator.config.ArrivalRadiusM ||
		currentDistanceM >= previousDistanceM {
		return arrivalComputation{}, false
	}

	segmentDuration := current.timeValue.Sub(previous.timeValue)
	if segmentDuration <= 0 {
		return arrivalComputation{}, false
	}
	durationSeconds := segmentDuration.Seconds()
	if !positiveFinite(durationSeconds) {
		return arrivalComputation{}, false
	}

	groundDistanceM := greatCircleDistanceM(
		previous.latitude,
		previous.longitude,
		current.latitude,
		current.longitude,
	)
	groundSpeedMPS := groundDistanceM / durationSeconds
	if !nonNegativeFinite(groundSpeedMPS) ||
		groundSpeedMPS > estimator.config.MaximumGroundSpeedMPS {
		return arrivalComputation{}, false
	}

	distanceClosedM := previousDistanceM - currentDistanceM
	radialClosingSpeedMPS := distanceClosedM / durationSeconds
	if !positiveFinite(radialClosingSpeedMPS) {
		return arrivalComputation{}, false
	}

	fraction :=
		(previousDistanceM - estimator.config.ArrivalRadiusM) /
			distanceClosedM
	fraction = math.Max(0, math.Min(1, fraction))
	crossingOffset, valid := durationCeilFraction(
		segmentDuration,
		fraction,
	)
	if !valid {
		return arrivalComputation{}, false
	}
	estimatedTime := previous.timeValue.Add(crossingOffset)

	uncertaintyM :=
		previous.horizontalUncertaintyM +
			fraction*(current.horizontalUncertaintyM-
				previous.horizontalUncertaintyM)
	uncertaintyDuration, valid := durationCeilSeconds(
		uncertaintyM / radialClosingSpeedMPS,
	)
	if !valid {
		return arrivalComputation{}, false
	}

	earliestTime := estimatedTime.Add(-uncertaintyDuration)
	latestTime := estimatedTime.Add(uncertaintyDuration)
	earliestTime, estimatedTime, latestTime =
		enforceMinimumArrivalInterval(
			projection.Horizon.AsOfTime.UTC(),
			estimatedTime,
			earliestTime,
			latestTime,
			estimator.config.MinimumArrivalInterval,
		)

	closingStdDevMPS := 0.0
	speedSampleCount := 1
	if profileAvailable {
		closingStdDevMPS = profile.closingSpeedStdDevMPS
		speedSampleCount = profile.sampleCount
	}

	return arrivalComputation{
		mode:                     EstimateModeWithinProjection,
		earliestTime:             earliestTime,
		estimatedTime:            estimatedTime,
		latestTime:               latestTime,
		estimatedClosingSpeedMPS: radialClosingSpeedMPS,
		closingSpeedStdDevMPS:    closingStdDevMPS,
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
	if distanceM > estimator.config.ArrivalRadiusM {
		return arrivalComputation{}, false
	}

	estimatedTime := sample.timeValue.UTC()
	if estimatedTime.Before(projection.Horizon.AsOfTime.UTC()) {
		estimatedTime = projection.Horizon.AsOfTime.UTC()
	}

	closingSpeedMPS := estimator.config.MinimumGroundSpeedMPS
	closingStdDevMPS := 0.0
	speedSampleCount := 0
	if profileAvailable &&
		positiveFinite(profile.meanClosingSpeedMPS) {
		closingSpeedMPS = profile.meanClosingSpeedMPS
		closingStdDevMPS = profile.closingSpeedStdDevMPS
		speedSampleCount = profile.sampleCount
	}

	uncertaintyDuration, valid := durationCeilSeconds(
		sample.horizontalUncertaintyM / closingSpeedMPS,
	)
	if !valid {
		return arrivalComputation{}, false
	}
	earliestTime := estimatedTime.Add(-uncertaintyDuration)
	latestTime := estimatedTime.Add(uncertaintyDuration)
	earliestTime, estimatedTime, latestTime =
		enforceMinimumArrivalInterval(
			projection.Horizon.AsOfTime.UTC(),
			estimatedTime,
			earliestTime,
			latestTime,
			estimator.config.MinimumArrivalInterval,
		)

	return arrivalComputation{
		mode:                     EstimateModeWithinProjection,
		earliestTime:             earliestTime,
		estimatedTime:            estimatedTime,
		latestTime:               latestTime,
		estimatedClosingSpeedMPS: closingSpeedMPS,
		closingSpeedStdDevMPS:    closingStdDevMPS,
		speedSampleCount:         speedSampleCount,
		remainingDistanceM:       0,
		lastPositionUncertaintyM: sample.horizontalUncertaintyM,
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
		profile.sampleCount < estimator.config.MinimumSpeedSampleCount ||
		profile.meanClosingSpeedMPS <
			estimator.config.MinimumGroundSpeedMPS {
		return arrivalComputation{}, false
	}

	conservativeClosingSpeedMPS :=
		profile.meanClosingSpeedMPS -
			estimator.config.SpeedUncertaintyMultiplier*
				profile.closingSpeedStdDevMPS
	if conservativeClosingSpeedMPS <
		estimator.config.MinimumGroundSpeedMPS {
		return arrivalComputation{}, false
	}
	optimisticClosingSpeedMPS := math.Min(
		estimator.config.MaximumGroundSpeedMPS,
		profile.meanClosingSpeedMPS+
			estimator.config.SpeedUncertaintyMultiplier*
				profile.closingSpeedStdDevMPS,
	)
	if !positiveFinite(optimisticClosingSpeedMPS) {
		return arrivalComputation{}, false
	}

	lastSample := samples[len(samples)-1]
	remainingDistanceM := math.Max(
		0,
		distances[len(distances)-1]-
			estimator.config.ArrivalRadiusM,
	)
	earliestDistanceM := math.Max(
		0,
		remainingDistanceM-lastSample.horizontalUncertaintyM,
	)
	latestDistanceM :=
		remainingDistanceM + lastSample.horizontalUncertaintyM

	earliestDuration, earliestValid := durationCeilSeconds(
		earliestDistanceM / optimisticClosingSpeedMPS,
	)
	estimatedDuration, estimatedValid := durationCeilSeconds(
		remainingDistanceM / profile.meanClosingSpeedMPS,
	)
	latestDuration, latestValid := durationCeilSeconds(
		latestDistanceM / conservativeClosingSpeedMPS,
	)
	if !earliestValid || !estimatedValid || !latestValid ||
		earliestDuration >
			estimator.config.MaximumEstimatedArrivalDuration ||
		estimatedDuration >
			estimator.config.MaximumEstimatedArrivalDuration ||
		latestDuration >
			estimator.config.MaximumEstimatedArrivalDuration {
		return arrivalComputation{}, false
	}

	earliestTime := lastSample.timeValue.Add(earliestDuration)
	estimatedTime := lastSample.timeValue.Add(estimatedDuration)
	latestTime := lastSample.timeValue.Add(latestDuration)
	earliestTime, estimatedTime, latestTime =
		enforceMinimumArrivalInterval(
			projection.Horizon.AsOfTime.UTC(),
			estimatedTime,
			earliestTime,
			latestTime,
			estimator.config.MinimumArrivalInterval,
		)

	maximumArrivalTime := lastSample.timeValue.Add(
		estimator.config.MaximumEstimatedArrivalDuration,
	)
	if earliestTime.After(maximumArrivalTime) ||
		estimatedTime.After(maximumArrivalTime) ||
		latestTime.After(maximumArrivalTime) {
		return arrivalComputation{}, false
	}

	extrapolationDuration := estimatedTime.Sub(
		projection.Horizon.EndTime.UTC(),
	)
	if extrapolationDuration < 0 {
		extrapolationDuration = 0
	}

	return arrivalComputation{
		mode:                     EstimateModeExtrapolated,
		earliestTime:             earliestTime,
		estimatedTime:            estimatedTime,
		latestTime:               latestTime,
		estimatedClosingSpeedMPS: profile.meanClosingSpeedMPS,
		closingSpeedStdDevMPS:    profile.closingSpeedStdDevMPS,
		speedSampleCount:         profile.sampleCount,
		remainingDistanceM:       remainingDistanceM,
		lastPositionUncertaintyM: lastSample.horizontalUncertaintyM,
		extrapolationDuration:    extrapolationDuration,
	}, true
}
