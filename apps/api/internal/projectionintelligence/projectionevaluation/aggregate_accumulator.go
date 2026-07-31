package projectionevaluation

import (
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

type leadTimeAccumulator struct {
	start       time.Duration
	end         time.Duration
	errors      []float64
	covered     int
	confidences []float64
	accuracies  []float64
	gaps        []float64
}

type methodAccumulator struct {
	method          projectioncontract.Method
	horizonDuration time.Duration
	forecastStep    time.Duration
	policy          EvaluationPolicy

	evaluationCount       int
	completeCount         int
	partialCount          int
	unavailableCount      int
	accuracyEligibleCount int

	forecastPointCount   int
	evaluatedPointCount  int
	horizontalErrors     []float64
	horizontalCovered    int
	trajectoryMeanErrors []float64
	trajectoryRMSEs      []float64

	altitudeErrors            []float64
	verticalCoverageEvaluated int
	verticalCovered           int

	endpointErrors         []float64
	endpointAltitudeErrors []float64

	confidenceScores     []float64
	normalizedAccuracies []float64
	confidenceGaps       []float64

	leadTimes map[time.Duration]*leadTimeAccumulator

	actualArrivalTruthCount          int
	arrivalPredictionCount           int
	arrivalPredictionWithTruthCount  int
	arrivalBothAvailableCount        int
	matchedArrivalCount              int
	missingArrivalPredictionCount    int
	unexpectedArrivalPredictionCount int
	arrivalAirportMismatchCount      int
	arrivalErrors                    []float64
	arrivalCovered                   int
}

func (accumulator *methodAccumulator) addEvaluation(evaluation Result) {
	accumulator.evaluationCount++
	switch evaluation.Status {
	case StatusComplete:
		accumulator.completeCount++
	case StatusPartial:
		accumulator.partialCount++
	case StatusUnavailable:
		accumulator.unavailableCount++
	}
	accumulator.addArrivalAvailability(evaluation.Arrival)
	if evaluation.Status == StatusUnavailable {
		return
	}

	accumulator.accuracyEligibleCount++
	accumulator.forecastPointCount += evaluation.Position.ForecastPointCount
	accumulator.evaluatedPointCount += evaluation.Position.EvaluatedPointCount
	if evaluation.Position.EvaluatedPointCount > 0 {
		accumulator.trajectoryMeanErrors = append(accumulator.trajectoryMeanErrors, evaluation.Position.MeanHorizontalErrorM)
		accumulator.trajectoryRMSEs = append(accumulator.trajectoryRMSEs, evaluation.Position.HorizontalRMSEM)
	}
	for _, point := range evaluation.Points {
		accumulator.horizontalErrors = append(accumulator.horizontalErrors, point.HorizontalErrorM)
		if point.WithinHorizontalUncertainty {
			accumulator.horizontalCovered++
		}
		if point.AltitudeAbsoluteErrorM != nil {
			accumulator.altitudeErrors = append(accumulator.altitudeErrors, *point.AltitudeAbsoluteErrorM)
		}
		if point.WithinVerticalUncertainty != nil {
			accumulator.verticalCoverageEvaluated++
			if *point.WithinVerticalUncertainty {
				accumulator.verticalCovered++
			}
		}
		accumulator.confidenceScores = append(accumulator.confidenceScores, point.ForecastConfidence.Score)
		accumulator.normalizedAccuracies = append(accumulator.normalizedAccuracies, point.NormalizedHorizontalAccuracy)
		accumulator.confidenceGaps = append(accumulator.confidenceGaps, point.AbsoluteConfidenceGap)
		accumulator.addLeadTimePoint(point)
	}
	if evaluation.Endpoint.Available {
		accumulator.endpointErrors = append(accumulator.endpointErrors, evaluation.Endpoint.HorizontalErrorM)
		if evaluation.Endpoint.AltitudeAvailable {
			accumulator.endpointAltitudeErrors = append(accumulator.endpointAltitudeErrors, evaluation.Endpoint.AltitudeAbsoluteErrorM)
		}
	}
	if evaluation.Arrival.Available {
		accumulator.arrivalErrors = append(accumulator.arrivalErrors, evaluation.Arrival.EstimatedAbsoluteErrorSeconds)
		if evaluation.Arrival.IntervalCoveredActual {
			accumulator.arrivalCovered++
		}
	}
}

func (accumulator *methodAccumulator) addArrivalAvailability(metrics ArrivalMetrics) {
	if metrics.ActualTruthAvailable {
		accumulator.actualArrivalTruthCount++
	}
	if metrics.PredictionAvailable {
		accumulator.arrivalPredictionCount++
	}
	if metrics.ActualTruthAvailable && metrics.PredictionAvailable {
		accumulator.arrivalPredictionWithTruthCount++
		accumulator.arrivalBothAvailableCount++
	}
	if metrics.AirportMatched {
		accumulator.matchedArrivalCount++
	}
	if metrics.ActualTruthAvailable && !metrics.PredictionAvailable {
		accumulator.missingArrivalPredictionCount++
	}
	if !metrics.ActualTruthAvailable && metrics.PredictionAvailable {
		accumulator.unexpectedArrivalPredictionCount++
	}
	if metrics.ActualTruthAvailable && metrics.PredictionAvailable && !metrics.AirportMatched {
		accumulator.arrivalAirportMismatchCount++
	}
}

func (accumulator *methodAccumulator) addLeadTimePoint(point PointEvaluation) {
	start := (point.LeadTime / accumulator.policy.LeadTimeBucketSize) * accumulator.policy.LeadTimeBucketSize
	item := accumulator.leadTimes[start]
	if item == nil {
		item = &leadTimeAccumulator{start: start, end: start + accumulator.policy.LeadTimeBucketSize}
		accumulator.leadTimes[start] = item
	}
	item.errors = append(item.errors, point.HorizontalErrorM)
	if point.WithinHorizontalUncertainty {
		item.covered++
	}
	item.confidences = append(item.confidences, point.ForecastConfidence.Score)
	item.accuracies = append(item.accuracies, point.NormalizedHorizontalAccuracy)
	item.gaps = append(item.gaps, point.AbsoluteConfidenceGap)
}

func (accumulator *methodAccumulator) summary() MethodSummary {
	summary := MethodSummary{
		MethodName:                       accumulator.method.Name,
		MethodVersion:                    accumulator.method.Version,
		DecisionClass:                    accumulator.method.DecisionClass,
		ProjectionHorizonDuration:        accumulator.horizonDuration,
		ForecastStep:                     accumulator.forecastStep,
		EvaluationPolicyVersion:          accumulator.policy.Version,
		EvaluationPolicyFingerprint:      accumulator.policy.InputFingerprint,
		LeadTimeBucketSize:               accumulator.policy.LeadTimeBucketSize,
		EvaluationCount:                  accumulator.evaluationCount,
		CompleteEvaluationCount:          accumulator.completeCount,
		PartialEvaluationCount:           accumulator.partialCount,
		UnavailableEvaluationCount:       accumulator.unavailableCount,
		AccuracyEligibleEvaluationCount:  accumulator.accuracyEligibleCount,
		ForecastPointCount:               accumulator.forecastPointCount,
		EvaluatedPointCount:              accumulator.evaluatedPointCount,
		AltitudeEvaluatedPointCount:      len(accumulator.altitudeErrors),
		EndpointEvaluationCount:          len(accumulator.endpointErrors),
		EndpointAltitudeEvaluationCount:  len(accumulator.endpointAltitudeErrors),
		ConfidenceEvaluationPointCount:   len(accumulator.confidenceScores),
		ActualArrivalTruthCount:          accumulator.actualArrivalTruthCount,
		ArrivalPredictionCount:           accumulator.arrivalPredictionCount,
		MatchedArrivalCount:              accumulator.matchedArrivalCount,
		MissingArrivalPredictionCount:    accumulator.missingArrivalPredictionCount,
		UnexpectedArrivalPredictionCount: accumulator.unexpectedArrivalPredictionCount,
		ArrivalAirportMismatchCount:      accumulator.arrivalAirportMismatchCount,
		ArrivalEvaluationCount:           len(accumulator.arrivalErrors),
	}
	if summary.ForecastPointCount > 0 {
		summary.PointCoverageRatio = float64(summary.EvaluatedPointCount) / float64(summary.ForecastPointCount)
	}
	if len(accumulator.horizontalErrors) > 0 {
		summary.MeanHorizontalErrorM = mean(accumulator.horizontalErrors)
		summary.MedianHorizontalErrorM = median(accumulator.horizontalErrors)
		summary.P95HorizontalErrorM = percentileNearestRank(accumulator.horizontalErrors, 0.95)
		summary.HorizontalRMSEM = rootMeanSquare(accumulator.horizontalErrors)
		summary.HorizontalUncertaintyCoverageRatio = float64(accumulator.horizontalCovered) / float64(len(accumulator.horizontalErrors))
	}
	summary.TrajectoryMacroMeanHorizontalErrorM = mean(accumulator.trajectoryMeanErrors)
	summary.TrajectoryMacroMeanHorizontalRMSEM = mean(accumulator.trajectoryRMSEs)
	if len(accumulator.altitudeErrors) > 0 {
		summary.MeanAltitudeAbsoluteErrorM = mean(accumulator.altitudeErrors)
		summary.AltitudeRMSEM = rootMeanSquare(accumulator.altitudeErrors)
	}
	if accumulator.verticalCoverageEvaluated > 0 {
		summary.VerticalUncertaintyCoverageRatio = float64(accumulator.verticalCovered) / float64(accumulator.verticalCoverageEvaluated)
	}
	if len(accumulator.endpointErrors) > 0 {
		summary.MeanEndpointHorizontalErrorM = mean(accumulator.endpointErrors)
		summary.EndpointHorizontalRMSEM = rootMeanSquare(accumulator.endpointErrors)
	}
	if len(accumulator.endpointAltitudeErrors) > 0 {
		summary.MeanEndpointAltitudeErrorM = mean(accumulator.endpointAltitudeErrors)
	}
	if len(accumulator.confidenceScores) > 0 {
		summary.MeanForecastConfidence = mean(accumulator.confidenceScores)
		summary.MeanNormalizedHorizontalAccuracy = mean(accumulator.normalizedAccuracies)
		summary.MeanAbsoluteConfidenceGap = mean(accumulator.confidenceGaps)
		summary.ConfidenceCalibrationRMSE = rootMeanSquare(accumulator.confidenceGaps)
	}
	summary.LeadTimes = accumulator.leadTimeSummaries()
	if accumulator.actualArrivalTruthCount > 0 {
		summary.ArrivalPredictionRecall = float64(accumulator.arrivalPredictionWithTruthCount) / float64(accumulator.actualArrivalTruthCount)
	}
	if accumulator.arrivalBothAvailableCount > 0 {
		summary.ArrivalAirportAccuracy = float64(accumulator.matchedArrivalCount) / float64(accumulator.arrivalBothAvailableCount)
	}
	if len(accumulator.arrivalErrors) > 0 {
		summary.MeanArrivalAbsoluteErrorSeconds = mean(accumulator.arrivalErrors)
		summary.ArrivalIntervalCoverageRatio = float64(accumulator.arrivalCovered) / float64(len(accumulator.arrivalErrors))
	}
	return summary
}

func (accumulator *methodAccumulator) leadTimeSummaries() []LeadTimeMetrics {
	starts := make([]time.Duration, 0, len(accumulator.leadTimes))
	for start := range accumulator.leadTimes {
		starts = append(starts, start)
	}
	sort.Slice(starts, func(left, right int) bool { return starts[left] < starts[right] })
	result := make([]LeadTimeMetrics, 0, len(starts))
	for _, start := range starts {
		item := accumulator.leadTimes[start]
		result = append(result, LeadTimeMetrics{
			BucketStart:                        item.start,
			BucketEnd:                          item.end,
			EvaluatedPointCount:                len(item.errors),
			MeanHorizontalErrorM:               mean(item.errors),
			MedianHorizontalErrorM:             median(item.errors),
			P95HorizontalErrorM:                percentileNearestRank(item.errors, 0.95),
			HorizontalRMSEM:                    rootMeanSquare(item.errors),
			HorizontalUncertaintyCoverageRatio: float64(item.covered) / float64(len(item.errors)),
			MeanForecastConfidence:             mean(item.confidences),
			MeanNormalizedHorizontalAccuracy:   mean(item.accuracies),
			MeanAbsoluteConfidenceGap:          mean(item.gaps),
		})
	}
	return result
}
