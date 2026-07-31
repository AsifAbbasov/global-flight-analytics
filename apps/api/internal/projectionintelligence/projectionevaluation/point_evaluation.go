package projectionevaluation

import (
	"math"
	"sort"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func evaluateForecastPoints(
	projection projectioncontract.Result,
	truth normalizedTruth,
	config Config,
) ([]PointEvaluation, int) {
	points := make([]PointEvaluation, 0, len(projection.Points))
	rejectedImplausible := 0
	for _, forecast := range projection.Points {
		actual, status := truthAt(truth.points, forecast.ForecastTime, config)
		if status == truthMatchImplausibleMovement {
			rejectedImplausible++
		}
		if status != truthMatchAvailable {
			continue
		}
		horizontalErrorM := greatCircleDistanceM(
			forecast.Position.Latitude,
			forecast.Position.Longitude,
			actual.latitude,
			actual.longitude,
		)
		if !nonNegativeFinite(horizontalErrorM) {
			continue
		}
		ratio := horizontalErrorM / config.MaximumHorizontalErrorM
		accuracy := clampUnit(1 - ratio)
		point := PointEvaluation{
			Sequence:                       forecast.Sequence,
			ForecastTime:                   forecast.ForecastTime.UTC(),
			LeadTime:                       forecast.ForecastTime.UTC().Sub(projection.Horizon.AsOfTime.UTC()),
			ActualSource:                   actual.source,
			ActualTime:                     actual.timeValue.UTC(),
			ForecastLatitude:               forecast.Position.Latitude,
			ForecastLongitude:              forecast.Position.Longitude,
			ActualLatitude:                 actual.latitude,
			ActualLongitude:                actual.longitude,
			ForecastHorizontalUncertaintyM: forecast.Uncertainty.HorizontalRadiusM,
			HorizontalErrorM:               horizontalErrorM,
			HorizontalErrorRatio:           ratio,
			WithinHorizontalUncertainty:    horizontalErrorM <= forecast.Uncertainty.HorizontalRadiusM,
			ForecastConfidence:             cloneConfidence(forecast.Confidence),
			NormalizedHorizontalAccuracy:   accuracy,
			AbsoluteConfidenceGap:          math.Abs(forecast.Confidence.Score - accuracy),
		}
		point.ForecastAltitudeM = cloneFloat(forecast.Position.AltitudeM)
		point.ActualAltitudeM = cloneFloat(actual.altitudeM)
		point.ForecastVerticalUncertaintyM = cloneFloat(forecast.Uncertainty.VerticalRadiusM)
		if forecast.Position.AltitudeM != nil && actual.altitudeM != nil {
			altitudeErrorM := math.Abs(*forecast.Position.AltitudeM - *actual.altitudeM)
			point.AltitudeAbsoluteErrorM = float64Pointer(altitudeErrorM)
			point.AltitudeErrorRatio = float64Pointer(altitudeErrorM / config.MaximumAltitudeErrorM)
			if forecast.Uncertainty.VerticalRadiusM != nil {
				point.WithinVerticalUncertainty = boolPointer(altitudeErrorM <= *forecast.Uncertainty.VerticalRadiusM)
			}
		}
		points = append(points, point)
	}
	sort.Slice(points, func(left, right int) bool {
		return points[left].ForecastTime.Before(points[right].ForecastTime)
	})
	return points, rejectedImplausible
}
