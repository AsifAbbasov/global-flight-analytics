package projectionevaluation

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

var (
	ErrProjectionInvalid = errors.New(
		"projection contract is invalid",
	)
	ErrTrajectoryIdentifierMismatch = errors.New(
		"projection and actual trajectory identifiers must match",
	)
	ErrEvaluatedAtInvalid = errors.New(
		"evaluation time must not precede projection generation",
	)
	ErrActualArrivalInvalid = errors.New(
		"actual arrival evidence is invalid",
	)
	ErrEvaluationResultInvalid = errors.New(
		"projection evaluation result is invalid",
	)
)

type Evaluator struct {
	config Config
}

func New(
	config Config,
) (*Evaluator, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate projection evaluation config: %w",
			err,
		)
	}

	return &Evaluator{
		config: config,
	}, nil
}

type Request struct {
	Projection       projectioncontract.Result
	ActualTrajectory trajectory.FlightTrajectory
	ActualArrival    *ActualArrival
	EvaluatedAt      time.Time
}

func (
	evaluator *Evaluator,
) Evaluate(
	request Request,
) (Result, error) {
	if evaluator == nil {
		return Result{},
			ErrEvaluationResultInvalid
	}

	projectionReport :=
		projectioncontract.Validate(
			request.Projection,
		)
	if projectionReport.Status !=
		projectioncontract.
			ValidationStatusValid {
		return Result{},
			fmt.Errorf(
				"%w: %#v",
				ErrProjectionInvalid,
				projectionReport.Issues,
			)
	}

	if strings.TrimSpace(
		request.ActualTrajectory.ID,
	) == "" ||
		request.ActualTrajectory.ID !=
			request.Projection.TrajectoryID {
		return Result{},
			ErrTrajectoryIdentifierMismatch
	}

	evaluatedAt :=
		request.EvaluatedAt.UTC()
	if evaluatedAt.IsZero() ||
		evaluatedAt.Before(
			request.Projection.
				GeneratedAt.UTC(),
		) {
		return Result{},
			ErrEvaluatedAtInvalid
	}

	if err := validateActualArrival(
		request.ActualArrival,
		request.Projection,
		evaluatedAt,
	); err != nil {
		return Result{}, err
	}

	asOfTime :=
		request.Projection.
			Horizon.AsOfTime.UTC()
	truthPoints,
		excludedAfterEvaluation :=
		normalizeTruthPoints(
			request.ActualTrajectory,
			asOfTime,
			evaluatedAt,
		)

	evaluations := make(
		[]PointEvaluation,
		0,
		len(request.Projection.Points),
	)
	limitations := make(
		[]Notice,
		0,
		5,
	)

	if excludedAfterEvaluation > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "truth_after_evaluation_cutoff_excluded",
				Message: fmt.Sprintf(
					"%d actual trajectory points after the evaluation cutoff were excluded.",
					excludedAfterEvaluation,
				),
			},
		)
	}

	for _, forecast := range request.Projection.Points {
		actual, available := truthAt(
			truthPoints,
			forecast.ForecastTime,
			evaluator.config.
				MaximumInterpolationGap,
		)
		if !available {
			continue
		}

		horizontalErrorM :=
			greatCircleDistanceM(
				forecast.Position.Latitude,
				forecast.Position.Longitude,
				actual.latitude,
				actual.longitude,
			)
		if !nonNegativeFinite(
			horizontalErrorM,
		) {
			continue
		}

		point := PointEvaluation{
			Sequence:     forecast.Sequence,
			ForecastTime: forecast.ForecastTime.UTC(),
			ActualSource: actual.source,
			ActualTime:   actual.timeValue.UTC(),

			ForecastLatitude:  forecast.Position.Latitude,
			ForecastLongitude: forecast.Position.Longitude,
			ActualLatitude:    actual.latitude,
			ActualLongitude:   actual.longitude,

			HorizontalErrorM: horizontalErrorM,
			HorizontalErrorRatio: horizontalErrorM /
				evaluator.config.
					MaximumHorizontalErrorM,
			WithinHorizontalUncertainty: horizontalErrorM <=
				forecast.Uncertainty.
					HorizontalRadiusM,

			ForecastConfidence: cloneConfidence(forecast.Confidence),
		}

		if forecast.Position.AltitudeM != nil {
			point.ForecastAltitudeM =
				cloneFloat(
					forecast.Position.
						AltitudeM,
				)
		}
		if actual.altitudeM != nil {
			point.ActualAltitudeM =
				cloneFloat(
					actual.altitudeM,
				)
		}
		if forecast.Position.AltitudeM != nil &&
			actual.altitudeM != nil {
			altitudeErrorM := math.Abs(
				*forecast.Position.AltitudeM -
					*actual.altitudeM,
			)
			point.AltitudeAbsoluteErrorM =
				float64Pointer(
					altitudeErrorM,
				)
			point.AltitudeErrorRatio =
				float64Pointer(
					altitudeErrorM /
						evaluator.config.
							MaximumAltitudeErrorM,
				)

			if forecast.Uncertainty.
				VerticalRadiusM != nil {
				within :=
					altitudeErrorM <=
						*forecast.Uncertainty.
							VerticalRadiusM
				point.
					WithinVerticalUncertainty =
					boolPointer(within)
			}
		}

		evaluations = append(
			evaluations,
			point,
		)
	}

	sort.SliceStable(
		evaluations,
		func(left int, right int) bool {
			return evaluations[left].
				ForecastTime.Before(
				evaluations[right].
					ForecastTime,
			)
		},
	)

	positionMetrics :=
		buildPositionMetrics(
			len(request.Projection.Points),
			evaluations,
		)
	arrivalMetrics,
		arrivalLimitation :=
		evaluateArrival(
			request.Projection.Arrival,
			request.ActualArrival,
		)
	if arrivalLimitation != nil {
		limitations = append(
			limitations,
			*arrivalLimitation,
		)
	}

	status := StatusUnavailable
	switch {
	case len(evaluations) <
		evaluator.config.
			MinimumEvaluatedPointCount:
		limitations = append(
			limitations,
			Notice{
				Code: "insufficient_evaluated_projection_points",
				Message: fmt.Sprintf(
					"Evaluation requires at least %d forecast points with actual truth, but %d were available.",
					evaluator.config.
						MinimumEvaluatedPointCount,
					len(evaluations),
				),
			},
		)
	case positionMetrics.
		EvaluatedPointCount ==
		positionMetrics.
			ForecastPointCount &&
		arrivalEvaluationComplete(
			request.Projection.Arrival,
			request.ActualArrival,
			arrivalMetrics,
		):
		status = StatusComplete
	default:
		status = StatusPartial
		if positionMetrics.
			MissingActualPointCount > 0 {
			limitations = append(
				limitations,
				Notice{
					Code: "actual_trajectory_coverage_partial",
					Message: fmt.Sprintf(
						"%d forecast points could not be matched to actual trajectory evidence.",
						positionMetrics.
							MissingActualPointCount,
					),
				},
			)
		}
	}

	result := Result{
		Version: Version,
		Status:  status,

		TrajectoryID:       request.Projection.TrajectoryID,
		ProjectionMethod:   request.Projection.Method,
		ProjectionAsOfTime: asOfTime,
		ProjectionGeneratedAt: request.Projection.
			GeneratedAt.UTC(),
		EvaluatedAt: evaluatedAt,

		ProjectionInputFingerprint: request.Projection.
			Provenance.
			InputFingerprint,
		EvaluationInputFingerprint: evaluationFingerprint(
			request.Projection,
			request.ActualTrajectory,
			request.ActualArrival,
			evaluatedAt,
			evaluator.config,
		),

		Points: append(
			[]PointEvaluation(nil),
			evaluations...,
		),
		Position: positionMetrics,
		Arrival:  arrivalMetrics,

		Limitations: normalizeNotices(
			limitations,
		),
	}

	if err := result.Validate(); err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %v",
				ErrEvaluationResultInvalid,
				err,
			)
	}

	return result.Clone(), nil
}

func validateActualArrival(
	actualArrival *ActualArrival,
	projection projectioncontract.Result,
	evaluatedAt time.Time,
) error {
	if actualArrival == nil {
		return nil
	}

	airportICAOCode := strings.ToUpper(
		strings.TrimSpace(
			actualArrival.AirportICAOCode,
		),
	)
	if len(airportICAOCode) != 4 ||
		actualArrival.BoundaryTime.IsZero() ||
		actualArrival.ObservedAt.IsZero() ||
		actualArrival.BoundaryTime.UTC().Before(
			projection.Horizon.AsOfTime.UTC(),
		) ||
		actualArrival.ObservedAt.UTC().Before(
			actualArrival.BoundaryTime.UTC(),
		) ||
		actualArrival.ObservedAt.UTC().After(
			evaluatedAt,
		) ||
		strings.TrimSpace(
			actualArrival.SourceName,
		) == "" {
		return ErrActualArrivalInvalid
	}

	return nil
}

func evaluateArrival(
	predicted *projectioncontract.ArrivalEstimate,
	actual *ActualArrival,
) (ArrivalMetrics, *Notice) {
	if predicted == nil {
		if actual == nil {
			return ArrivalMetrics{}, nil
		}

		return ArrivalMetrics{},
			&Notice{
				Code:    "arrival_prediction_unavailable_with_actual_truth",
				Message: "Actual arrival truth was available, but the projection did not publish an arrival estimate.",
			}
	}
	if actual == nil {
		return ArrivalMetrics{},
			&Notice{
				Code:    "actual_arrival_truth_unavailable",
				Message: "The projection published an arrival estimate, but no independent actual arrival truth was supplied.",
			}
	}

	predictedICAO := strings.ToUpper(
		strings.TrimSpace(
			predicted.AirportICAOCode,
		),
	)
	actualICAO := strings.ToUpper(
		strings.TrimSpace(
			actual.AirportICAOCode,
		),
	)
	if predictedICAO != actualICAO {
		return ArrivalMetrics{},
			&Notice{
				Code: "arrival_airport_mismatch",
				Message: fmt.Sprintf(
					"Predicted arrival airport %s does not match actual arrival airport %s.",
					predictedICAO,
					actualICAO,
				),
			}
	}

	signedErrorSeconds :=
		predicted.EstimatedTime.UTC().
			Sub(
				actual.BoundaryTime.UTC(),
			).Seconds()

	return ArrivalMetrics{
		Available: true,

		AirportICAOCode:    predictedICAO,
		ActualBoundaryTime: actual.BoundaryTime.UTC(),

		EarliestTime:  predicted.EarliestTime.UTC(),
		EstimatedTime: predicted.EstimatedTime.UTC(),
		LatestTime:    predicted.LatestTime.UTC(),

		EstimatedAbsoluteErrorSeconds: math.Abs(
			signedErrorSeconds,
		),
		SignedErrorSeconds: signedErrorSeconds,
		IntervalWidthSeconds: predicted.LatestTime.UTC().
			Sub(
				predicted.EarliestTime.UTC(),
			).Seconds(),
		IntervalCoveredActual: !actual.BoundaryTime.UTC().
			Before(
				predicted.EarliestTime.UTC(),
			) &&
			!actual.BoundaryTime.UTC().
				After(
					predicted.LatestTime.UTC(),
				),
	}, nil
}

func arrivalEvaluationComplete(
	predicted *projectioncontract.ArrivalEstimate,
	actual *ActualArrival,
	metrics ArrivalMetrics,
) bool {
	switch {
	case predicted == nil &&
		actual == nil:
		return true
	case predicted != nil &&
		actual != nil &&
		metrics.Available:
		return true
	default:
		return false
	}
}
