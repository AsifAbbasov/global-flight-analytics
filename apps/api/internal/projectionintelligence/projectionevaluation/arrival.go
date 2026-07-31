package projectionevaluation

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func normalizeActualArrival(actual *ActualArrival) *ActualArrival {
	if actual == nil {
		return nil
	}
	cloned := *actual
	cloned.AirportICAOCode = strings.ToUpper(strings.TrimSpace(cloned.AirportICAOCode))
	cloned.SourceName = strings.TrimSpace(cloned.SourceName)
	cloned.BoundaryTime = cloned.BoundaryTime.UTC()
	cloned.ObservedAt = cloned.ObservedAt.UTC()
	cloned.AvailableAt = cloned.AvailableAt.UTC()
	return &cloned
}

func validateActualArrival(
	actual *ActualArrival,
	projection projectioncontract.Result,
	evaluatedAt time.Time,
) error {
	if actual == nil {
		return nil
	}
	if !airportICAOPattern.MatchString(actual.AirportICAOCode) ||
		actual.BoundaryTime.IsZero() || actual.ObservedAt.IsZero() || actual.AvailableAt.IsZero() ||
		actual.BoundaryTime.Before(projection.Horizon.AsOfTime.UTC()) ||
		actual.ObservedAt.Before(actual.BoundaryTime) ||
		actual.AvailableAt.Before(actual.ObservedAt) ||
		actual.AvailableAt.After(evaluatedAt.UTC()) || actual.SourceName == "" {
		return ErrActualArrivalInvalid
	}
	return nil
}

func evaluateArrival(
	predicted *projectioncontract.ArrivalEstimate,
	actual *ActualArrival,
) (ArrivalMetrics, *Notice) {
	metrics := ArrivalMetrics{
		ActualTruthAvailable: actual != nil,
		PredictionAvailable:  predicted != nil,
	}
	if predicted != nil {
		metrics.PredictedAirportICAOCode = strings.ToUpper(strings.TrimSpace(predicted.AirportICAOCode))
		metrics.EarliestTime = predicted.EarliestTime.UTC()
		metrics.EstimatedTime = predicted.EstimatedTime.UTC()
		metrics.LatestTime = predicted.LatestTime.UTC()
	}
	if actual != nil {
		metrics.ActualAirportICAOCode = actual.AirportICAOCode
		metrics.ActualBoundaryTime = actual.BoundaryTime.UTC()
	}
	if predicted == nil {
		if actual == nil {
			return metrics, nil
		}
		return metrics, &Notice{
			Code:    "arrival_prediction_unavailable_with_actual_truth",
			Message: "Actual arrival truth was available, but the projection did not publish an arrival estimate.",
		}
	}
	if actual == nil {
		return metrics, &Notice{
			Code:    "actual_arrival_truth_unavailable",
			Message: "The projection published an arrival estimate, but no independent actual arrival truth was supplied.",
		}
	}
	if metrics.PredictedAirportICAOCode != metrics.ActualAirportICAOCode {
		return metrics, &Notice{
			Code:    "arrival_airport_mismatch",
			Message: fmt.Sprintf("Predicted arrival airport %s does not match actual arrival airport %s.", metrics.PredictedAirportICAOCode, metrics.ActualAirportICAOCode),
		}
	}
	metrics.AirportMatched = true
	metrics.Available = true
	metrics.SignedErrorSeconds = metrics.EstimatedTime.Sub(metrics.ActualBoundaryTime).Seconds()
	metrics.EstimatedAbsoluteErrorSeconds = math.Abs(metrics.SignedErrorSeconds)
	metrics.IntervalWidthSeconds = metrics.LatestTime.Sub(metrics.EarliestTime).Seconds()
	metrics.IntervalCoveredActual = !metrics.ActualBoundaryTime.Before(metrics.EarliestTime) && !metrics.ActualBoundaryTime.After(metrics.LatestTime)
	return metrics, nil
}

func arrivalEvaluationComplete(
	predicted *projectioncontract.ArrivalEstimate,
	actual *ActualArrival,
	metrics ArrivalMetrics,
) bool {
	switch {
	case predicted == nil && actual == nil:
		return true
	case predicted != nil && actual != nil && metrics.Available:
		return true
	default:
		return false
	}
}
