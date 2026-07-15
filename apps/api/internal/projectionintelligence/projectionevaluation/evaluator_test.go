package projectionevaluation

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func TestEvaluateProducesCompleteReplayMetrics(
	t *testing.T,
) {
	evaluator := newEvaluationEvaluator(t)
	request := evaluationTestRequest(true)

	first, err := evaluator.Evaluate(request)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}
	second, err := evaluator.Evaluate(request)
	if err != nil {
		t.Fatalf(
			"second Evaluate() error = %v",
			err,
		)
	}

	if first.Status != StatusComplete {
		t.Fatalf(
			"status = %q, want complete",
			first.Status,
		)
	}
	if first.Position.ForecastPointCount != 3 ||
		first.Position.EvaluatedPointCount != 3 ||
		first.Position.CoverageRatio != 1 {
		t.Fatalf(
			"unexpected point coverage: %#v",
			first.Position,
		)
	}
	if first.Position.MeanHorizontalErrorM <= 0 ||
		first.Position.HorizontalRMSEM <= 0 ||
		first.Position.
			HorizontalUncertaintyCoverageRatio != 1 {
		t.Fatalf(
			"unexpected horizontal metrics: %#v",
			first.Position,
		)
	}
	if first.Position.
		AltitudeEvaluatedPointCount != 3 ||
		first.Position.
			MeanAltitudeAbsoluteErrorM != 50 ||
		first.Position.
			VerticalUncertaintyCoverageRatio != 1 {
		t.Fatalf(
			"unexpected altitude metrics: %#v",
			first.Position,
		)
	}
	if !first.Arrival.Available ||
		first.Arrival.
			EstimatedAbsoluteErrorSeconds != 30 ||
		!first.Arrival.
			IntervalCoveredActual {
		t.Fatalf(
			"unexpected arrival metrics: %#v",
			first.Arrival,
		)
	}
	if first.EvaluationInputFingerprint !=
		second.EvaluationInputFingerprint {
		t.Fatal(
			"deterministic replay input produced different fingerprints",
		)
	}
	if err := first.Validate(); err != nil {
		t.Fatalf(
			"result validation error = %v",
			err,
		)
	}
}

func TestEvaluateProducesPartialMetricsWhenTruthCoverageIsMissing(
	t *testing.T,
) {
	config := validEvaluationConfig()
	config.MaximumInterpolationGap =
		time.Minute
	evaluator, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := evaluationTestRequest(false)
	request.ActualTrajectory.Points =
		[]trajectory.TrackPoint4D{
			request.ActualTrajectory.Points[0],
			request.ActualTrajectory.Points[2],
		}

	result, err := evaluator.Evaluate(request)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Status != StatusPartial ||
		result.Position.EvaluatedPointCount != 2 ||
		result.Position.MissingActualPointCount != 1 ||
		!hasEvaluationNotice(
			result.Limitations,
			"actual_trajectory_coverage_partial",
		) {
		t.Fatalf(
			"unexpected partial evaluation: %#v",
			result,
		)
	}
}

func TestEvaluateReturnsUnavailableWhenMinimumTruthIsNotMet(
	t *testing.T,
) {
	config := validEvaluationConfig()
	config.MinimumEvaluatedPointCount = 2
	evaluator, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := evaluationTestRequest(false)
	request.ActualTrajectory.Points =
		request.ActualTrajectory.Points[:1]

	result, err := evaluator.Evaluate(request)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Status != StatusUnavailable ||
		len(result.Points) != 1 ||
		!hasEvaluationNotice(
			result.Limitations,
			"insufficient_evaluated_projection_points",
		) {
		t.Fatalf(
			"unexpected unavailable evaluation: %#v",
			result,
		)
	}
}

func TestEvaluateExcludesTruthAfterEvaluationCutoff(
	t *testing.T,
) {
	evaluator := newEvaluationEvaluator(t)
	request := evaluationTestRequest(false)
	request.EvaluatedAt =
		request.Projection.
			Horizon.AsOfTime.Add(
			2 * time.Minute,
		)
	request.ActualTrajectory.Points =
		append(
			request.ActualTrajectory.Points,
			trajectory.TrackPoint4D{
				ID:        "truth-after-cutoff",
				Latitude:  0,
				Longitude: 100,
				ObservedAt: request.EvaluatedAt.Add(
					time.Minute,
				),
			},
		)

	result, err := evaluator.Evaluate(request)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Position.EvaluatedPointCount != 2 ||
		!hasEvaluationNotice(
			result.Limitations,
			"truth_after_evaluation_cutoff_excluded",
		) {
		t.Fatalf(
			"evaluation cutoff was not enforced: %#v",
			result,
		)
	}
}

func TestEvaluateReportsArrivalAirportMismatch(
	t *testing.T,
) {
	evaluator := newEvaluationEvaluator(t)
	request := evaluationTestRequest(true)
	request.ActualArrival.AirportICAOCode =
		"CCCC"

	result, err := evaluator.Evaluate(request)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Status != StatusPartial ||
		result.Arrival.Available ||
		!hasEvaluationNotice(
			result.Limitations,
			"arrival_airport_mismatch",
		) {
		t.Fatalf(
			"arrival mismatch was not reported: %#v",
			result,
		)
	}
}

func TestEvaluateRejectsIdentifierMismatchAndInvalidTime(
	t *testing.T,
) {
	evaluator := newEvaluationEvaluator(t)

	mismatch := evaluationTestRequest(false)
	mismatch.ActualTrajectory.ID =
		"other-trajectory"
	_, err := evaluator.Evaluate(mismatch)
	if !errors.Is(
		err,
		ErrTrajectoryIdentifierMismatch,
	) {
		t.Fatalf(
			"identifier mismatch error = %v",
			err,
		)
	}

	invalidTime := evaluationTestRequest(false)
	invalidTime.EvaluatedAt =
		invalidTime.Projection.
			GeneratedAt.Add(-time.Second)
	_, err = evaluator.Evaluate(invalidTime)
	if !errors.Is(
		err,
		ErrEvaluatedAtInvalid,
	) {
		t.Fatalf(
			"evaluated-at error = %v",
			err,
		)
	}
}

func newEvaluationEvaluator(
	t *testing.T,
) *Evaluator {
	t.Helper()

	evaluator, err := New(
		validEvaluationConfig(),
	)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	return evaluator
}

func evaluationTestRequest(
	withArrival bool,
) Request {
	asOfTime := evaluationTestAsOfTime()
	projection :=
		validEvaluationProjection(
			asOfTime,
			withArrival,
		)
	actualTrajectory :=
		validActualTrajectory(
			asOfTime,
		)

	request := Request{
		Projection:       projection,
		ActualTrajectory: actualTrajectory,
		EvaluatedAt: asOfTime.Add(
			5 * time.Minute,
		),
	}
	if withArrival {
		request.ActualArrival =
			&ActualArrival{
				AirportICAOCode: "BBBB",
				BoundaryTime: asOfTime.Add(
					150 * time.Second,
				),
				SourceName: "actual-arrival-truth",
				ObservedAt: asOfTime.Add(
					4 * time.Minute,
				),
			}
	}

	return request
}

func validEvaluationProjection(
	asOfTime time.Time,
	withArrival bool,
) projectioncontract.Result {
	points := make(
		[]projectioncontract.ProjectionPoint,
		0,
		3,
	)

	for index := 0; index < 3; index++ {
		altitudeM :=
			1000 +
				float64(index)*100
		verticalRadiusM := 100.0
		points = append(
			points,
			projectioncontract.ProjectionPoint{
				Sequence: index,
				ForecastTime: asOfTime.Add(
					time.Duration(
						index+1,
					) * time.Minute,
				),
				Position: projectioncontract.Position{
					Latitude: 0,
					Longitude: 0.01 *
						float64(
							index+1,
						),
					AltitudeM: &altitudeM,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 1000,
					VerticalRadiusM:   &verticalRadiusM,
				},
				Confidence: projectioncontract.Confidence{
					Score: 0.8 -
						float64(index)*0.05,
					Level: projectioncontract.
						ConfidenceLevelHigh,
					Reasons: []projectioncontract.
						ConfidenceReason{
						{
							Code:    "point_confidence",
							Message: "Projection point confidence.",
							Contribution: 0.8 -
								float64(index)*0.05,
						},
					},
				},
			},
		)
	}

	result := projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status: projectioncontract.
			ResultStatusComplete,
		TrajectoryID: "trajectory-001",
		FlightID:     "flight-001",
		AircraftID:   "aircraft-001",
		ICAO24:       "4A1234",
		Callsign:     "AHY123",
		Method: projectioncontract.Method{
			Name:    "evaluation_test_method",
			Version: "evaluation-test-method-v1",
			DecisionClass: projectioncontract.
				DecisionClassExperimental,
		},
		Horizon: projectioncontract.Horizon{
			AsOfTime: asOfTime,
			EndTime: asOfTime.Add(
				3 * time.Minute,
			),
			Step: time.Minute,
		},
		Points: points,
		Confidence: projectioncontract.Confidence{
			Score: 0.7,
			Level: projectioncontract.
				ConfidenceLevelMedium,
			Reasons: []projectioncontract.
				ConfidenceReason{
				{
					Code:         "result_confidence",
					Message:      "Projection result confidence.",
					Contribution: 0.7,
				},
			},
		},
		Limitations: []projectioncontract.Limitation{
			{
				Code:    "research_only",
				Message: "Projection is a research estimate.",
				Scope:   "result",
			},
		},
		Explanations: []projectioncontract.Explanation{
			{
				Code:    "evaluation_test_method",
				Message: "Projection fixture for replay evaluation.",
			},
		},
		ScopeGuard: projectioncontract.
			ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: "sha256:" +
				strings.Repeat(
					"a",
					64,
				),
			Inputs: []projectioncontract.InputReference{
				{
					Name: "current_position",
					Classification: projectioncontract.
						InputClassificationObserved,
					SourceName: "test-source",
					ObservedAt: asOfTime,
					RetrievedAt: asOfTime.Add(
						time.Second,
					),
				},
			},
			LatestInputObservedAt: asOfTime,
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}

	if withArrival {
		result.Arrival =
			&projectioncontract.ArrivalEstimate{
				AirportICAOCode: "BBBB",
				EarliestTime: asOfTime.Add(
					2 * time.Minute,
				),
				EstimatedTime: asOfTime.Add(
					3 * time.Minute,
				),
				LatestTime: asOfTime.Add(
					4 * time.Minute,
				),
				Confidence: projectioncontract.Confidence{
					Score: 0.7,
					Level: projectioncontract.
						ConfidenceLevelMedium,
					Reasons: []projectioncontract.
						ConfidenceReason{
						{
							Code:         "arrival_confidence",
							Message:      "Arrival confidence.",
							Contribution: 0.7,
						},
					},
				},
				Limitations: []projectioncontract.Limitation{
					{
						Code:    "arrival_radius",
						Message: "Arrival represents airport-radius entry.",
						Scope:   "arrival",
					},
				},
			}
	}

	return result
}

func validActualTrajectory(
	asOfTime time.Time,
) trajectory.FlightTrajectory {
	points := make(
		[]trajectory.TrackPoint4D,
		0,
		3,
	)

	for index := 0; index < 3; index++ {
		points = append(
			points,
			trajectory.TrackPoint4D{
				ID: "actual-point-" +
					string(
						rune('0'+index),
					),
				FlightID:   "flight-001",
				AircraftID: "aircraft-001",
				ICAO24:     "4A1234",
				Callsign:   "AHY123",
				Latitude:   0,
				Longitude: 0.0105 *
					float64(
						index+1,
					),
				GeometricAltitudeM: 1050 +
					float64(index)*100,
				GeometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,
				BarometricAltitudeM: 1050 +
					float64(index)*100,
				BarometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,
				ObservedAt: asOfTime.Add(
					time.Duration(
						index+1,
					) * time.Minute,
				),
				SourceName: "actual-truth-source",
			},
		)
	}

	return trajectory.FlightTrajectory{
		ID:         "trajectory-001",
		FlightID:   "flight-001",
		AircraftID: "aircraft-001",
		ICAO24:     "4A1234",
		Callsign:   "AHY123",
		StartTime:  points[0].ObservedAt,
		EndTime: points[len(points)-1].
			ObservedAt,
		PointCount:   len(points),
		QualityScore: 1,
		SourceName:   "actual-truth-source",
		Points:       points,
		UpdatedAt: points[len(points)-1].
			ObservedAt,
	}
}

func evaluationTestAsOfTime() time.Time {
	return time.Date(
		2026,
		time.July,
		16,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}

func hasEvaluationNotice(
	items []Notice,
	code string,
) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}

	return false
}

func TestEvaluationMetricsAreFinite(
	t *testing.T,
) {
	evaluator := newEvaluationEvaluator(t)
	result, err := evaluator.Evaluate(
		evaluationTestRequest(true),
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	values := []float64{
		result.Position.
			MeanHorizontalErrorM,
		result.Position.
			HorizontalRMSEM,
		result.Position.
			MeanAltitudeAbsoluteErrorM,
		result.Arrival.
			EstimatedAbsoluteErrorSeconds,
	}
	for _, value := range values {
		if math.IsNaN(value) ||
			math.IsInf(value, 0) {
			t.Fatalf(
				"metric is not finite: %f",
				value,
			)
		}
	}
}
