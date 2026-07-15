package projectioncontinuation

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

type neighborSelectorStub struct {
	result projectionneighbors.Result
	err    error
	calls  int
}

func (
	stub *neighborSelectorStub,
) Select(
	projectionneighbors.Request,
) (projectionneighbors.Result, error) {
	stub.calls++
	return stub.result.Clone(),
		stub.err
}

type patternEvaluatorStub struct {
	result projectionpatternconfidence.Result
	err    error
	calls  int
}

func (
	stub *patternEvaluatorStub,
) Evaluate(
	projectionneighbors.Result,
) (projectionpatternconfidence.Result, error) {
	stub.calls++
	return stub.result.Clone(),
		stub.err
}

type fallbackProjectorStub struct {
	result projectioncontract.Result
	err    error
	calls  int
}

func (
	stub *fallbackProjectorStub,
) Project(
	projectionbaseline.Request,
) (projectioncontract.Result, error) {
	stub.calls++
	return stub.result.Clone(),
		stub.err
}

func TestProjectBuildsHistoricalNeighborContinuation(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	fallback := config.FallbackProjector.(*fallbackProjectorStub)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := continuationTestRequest()
	first, err := baseline.Project(request)
	if err != nil {
		t.Fatalf(
			"Project() error = %v",
			err,
		)
	}
	second, err := baseline.Project(request)
	if err != nil {
		t.Fatalf(
			"second Project() error = %v",
			err,
		)
	}

	if fallback.calls != 0 {
		t.Fatalf(
			"fallback calls = %d, want 0",
			fallback.calls,
		)
	}
	if first.Method.Name != MethodName ||
		first.Method.Version != Version ||
		first.Method.DecisionClass !=
			projectioncontract.
				DecisionClassExperimental {
		t.Fatalf(
			"unexpected method: %#v",
			first.Method,
		)
	}
	if first.Status !=
		projectioncontract.ResultStatusComplete ||
		len(first.Points) != 2 {
		t.Fatalf(
			"unexpected projection result: %#v",
			first,
		)
	}
	if first.Points[0].Position.Latitude <=
		request.CurrentTrajectory.Points[len(request.CurrentTrajectory.Points)-1].Latitude ||
		first.Points[0].Position.Longitude <=
			request.CurrentTrajectory.Points[len(request.CurrentTrajectory.Points)-1].Longitude {
		t.Fatalf(
			"historical continuation did not translate movement: %#v",
			first.Points[0].Position,
		)
	}
	if first.Points[0].Position.AltitudeM == nil ||
		first.Points[0].Uncertainty.
			VerticalRadiusM == nil ||
		first.Points[0].Uncertainty.
			HorizontalRadiusM <= 0 {
		t.Fatalf(
			"expected explicit position and uncertainty: %#v",
			first.Points[0],
		)
	}
	if first.Provenance.InputFingerprint !=
		second.Provenance.InputFingerprint {
		t.Fatal(
			"deterministic input produced different fingerprints",
		)
	}
	if !equalContinuationPoints(
		first.Points,
		second.Points,
	) {
		t.Fatal(
			"deterministic input produced different points",
		)
	}

	report := projectioncontract.Validate(first)
	if report.Status !=
		projectioncontract.ValidationStatusValid {
		t.Fatalf(
			"generated contract invalid: %#v",
			report.Issues,
		)
	}
}

func TestProjectFallsBackWhenPatternIsNotUsable(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	pattern := config.
		PatternConfidenceEvaluator.(*patternEvaluatorStub)
	pattern.result.Status =
		projectionpatternconfidence.
			StatusUnavailable
	pattern.result.Usable = false
	pattern.result.Score = 0.4
	pattern.result.Level =
		projectioncontract.
			ConfidenceLevelLow
	pattern.result.Limitations =
		[]projectionpatternconfidence.Notice{
			{
				Code:    "pattern_confidence_below_minimum",
				Message: "Pattern confidence is below the configured minimum.",
			},
		}

	fallback := config.FallbackProjector.(*fallbackProjectorStub)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	result, err := baseline.Project(
		continuationTestRequest(),
	)
	if err != nil {
		t.Fatalf(
			"Project() error = %v",
			err,
		)
	}

	if fallback.calls != 1 {
		t.Fatalf(
			"fallback calls = %d, want 1",
			fallback.calls,
		)
	}
	if result.Method.Name !=
		projectionbaseline.MethodName ||
		!hasProjectionLimitation(
			result.Limitations,
			"historical_neighbor_strategy_fallback",
		) ||
		!hasProjectionExplanation(
			result.Explanations,
			"kinematic_fallback_selected",
		) {
		t.Fatalf(
			"fallback result lacks strategy evidence: %#v",
			result,
		)
	}
	if result.Provenance.InputFingerprint ==
		fallback.result.Provenance.
			InputFingerprint {
		t.Fatal(
			"fallback strategy did not update the input fingerprint",
		)
	}

	report := projectioncontract.Validate(result)
	if report.Status !=
		projectioncontract.ValidationStatusValid {
		t.Fatalf(
			"fallback contract invalid: %#v",
			report.Issues,
		)
	}
}

func TestProjectFallsBackWhenPointSupportIsInsufficient(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	fallback := config.FallbackProjector.(*fallbackProjectorStub)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := continuationTestRequest()
	request.Candidates =
		request.Candidates[:1]

	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf(
			"Project() error = %v",
			err,
		)
	}
	if fallback.calls != 1 ||
		result.Method.Name !=
			projectionbaseline.MethodName ||
		!hasFallbackReason(
			result.Limitations,
			"historical_continuation_point_support_insufficient",
		) {
		t.Fatalf(
			"unexpected insufficient-support fallback: %#v",
			result,
		)
	}
}

func TestProjectMarksLimitedWhenAltitudeSupportIsPartial(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := continuationTestRequest()
	for pointIndex := range request.Candidates[1].Points {
		request.Candidates[1].Points[pointIndex].GeometricAltitudeStatus =
			flightstate.
				AltitudeStatusUnavailable
		request.Candidates[1].Points[pointIndex].BarometricAltitudeStatus =
			flightstate.
				AltitudeStatusUnavailable
	}

	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf(
			"Project() error = %v",
			err,
		)
	}
	if result.Status !=
		projectioncontract.ResultStatusLimited ||
		result.Points[0].Position.AltitudeM != nil {
		t.Fatalf(
			"insufficient altitude support should produce horizontal-only limited output: %#v",
			result,
		)
	}

	for pointIndex := range request.Candidates[0].Points {
		request.Candidates[0].Points[pointIndex].GeometricAltitudeStatus =
			flightstate.
				AltitudeStatusUnavailable
		request.Candidates[0].Points[pointIndex].BarometricAltitudeStatus =
			flightstate.
				AltitudeStatusUnavailable
	}

	result, err = baseline.Project(request)
	if err != nil {
		t.Fatalf(
			"Project() without historical altitude error = %v",
			err,
		)
	}
	if result.Status !=
		projectioncontract.ResultStatusLimited ||
		result.Points[0].Position.AltitudeM != nil ||
		!hasProjectionLimitation(
			result.Limitations,
			"historical_continuation_altitude_partial",
		) {
		t.Fatalf(
			"unexpected horizontal-only continuation: %#v",
			result,
		)
	}
}

func validContinuationConfig(
	t *testing.T,
) Config {
	t.Helper()

	policy, err := projectionhorizon.New(
		projectionhorizon.Config{
			Name:              "historical-continuation-test",
			MinimumDuration:   time.Minute,
			DefaultDuration:   2 * time.Minute,
			MaximumDuration:   5 * time.Minute,
			Step:              time.Minute,
			MaximumPointCount: 5,
		},
	)
	if err != nil {
		t.Fatalf(
			"projectionhorizon.New() error = %v",
			err,
		)
	}

	request := continuationTestRequest()
	selection :=
		continuationTestSelection(
			request,
		)
	pattern :=
		continuationTestPattern(
			selection,
		)

	return Config{
		HorizonPlanner: policy,
		NeighborSelector: &neighborSelectorStub{
			result: selection,
		},
		PatternConfidenceEvaluator: &patternEvaluatorStub{
			result: pattern,
		},
		FallbackProjector: &fallbackProjectorStub{
			result: validFallbackResult(
				request,
			),
		},

		MinimumPointSupport:    2,
		MinimumAltitudeSupport: 2,

		InitialHorizontalUncertaintyM:  100,
		HorizontalUncertaintyGrowthMPS: 1,
		InitialVerticalUncertaintyM:    20,
		VerticalUncertaintyGrowthMPS:   0.2,
		NeighborSpreadMultiplier:       1.5,

		MaximumConfidenceLoss: 0.4,

		MediumConfidenceMinimum: 0.6,
		HighConfidenceMinimum:   0.8,
	}
}

func continuationTestRequest() Request {
	asOfTime := continuationTestAsOfTime()
	current :=
		continuationCurrentTrajectory(
			asOfTime,
		)

	return Request{
		CurrentTrajectory: current,
		Candidates: []trajectory.FlightTrajectory{
			continuationCandidate(
				"historical-a",
				asOfTime.Add(
					-24*time.Hour,
				),
				0.01,
				0,
			),
			continuationCandidate(
				"historical-b",
				asOfTime.Add(
					-25*time.Hour,
				),
				0,
				0.01,
			),
		},
		AsOfTime:          asOfTime,
		RequestedDuration: 2 * time.Minute,
		GeneratedAt:       asOfTime.Add(time.Second),
	}
}

func continuationCurrentTrajectory(
	asOfTime time.Time,
) trajectory.FlightTrajectory {
	points := make(
		[]trajectory.TrackPoint4D,
		0,
		5,
	)
	for index := 0; index < 5; index++ {
		points = append(
			points,
			trajectory.TrackPoint4D{
				ID: "current-point-" +
					string(
						rune('0'+index),
					),
				FlightID:   "flight-current",
				AircraftID: "aircraft-current",
				ICAO24:     "4K0001",
				Callsign:   "AHY001",
				Latitude: 40.96 +
					float64(index)*0.01,
				Longitude: 49.96 +
					float64(index)*0.01,
				GeometricAltitudeM: 1000,
				GeometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,
				BarometricAltitudeM: 1000,
				BarometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,
				ObservedAt: asOfTime.Add(
					time.Duration(
						index-4,
					) * time.Minute,
				),
				SourceName: "airplanes.live",
			},
		)
	}

	return trajectory.FlightTrajectory{
		ID:           "current",
		FlightID:     "flight-current",
		AircraftID:   "aircraft-current",
		ICAO24:       "4K0001",
		Callsign:     "AHY001",
		StartTime:    points[0].ObservedAt,
		EndTime:      points[len(points)-1].ObservedAt,
		PointCount:   len(points),
		QualityScore: 0.9,
		SourceName:   "airplanes.live",
		Points:       points,
	}
}

func continuationCandidate(
	id string,
	endTime time.Time,
	latitudeStep float64,
	longitudeStep float64,
) trajectory.FlightTrajectory {
	startTime := endTime.Add(
		-6 * time.Minute,
	)
	points := make(
		[]trajectory.TrackPoint4D,
		0,
		7,
	)

	for index := 0; index < 7; index++ {
		latitude := 39.96 +
			float64(index)*0.01
		longitude := 48.96 +
			float64(index)*0.01
		if index >= 4 {
			latitude = 40 +
				float64(index-4)*
					latitudeStep
			longitude = 49 +
				float64(index-4)*
					longitudeStep
		}

		points = append(
			points,
			trajectory.TrackPoint4D{
				ID: id + "-point-" +
					string(
						rune('0'+index),
					),
				Latitude:  latitude,
				Longitude: longitude,
				GeometricAltitudeM: 900 +
					float64(index-4)*
						100,
				GeometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,
				BarometricAltitudeM: 900 +
					float64(index-4)*
						100,
				BarometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,
				ObservedAt: startTime.Add(
					time.Duration(index) *
						time.Minute,
				),
				SourceName: "historical-store",
			},
		)
	}

	return trajectory.FlightTrajectory{
		ID:           id,
		StartTime:    points[0].ObservedAt,
		EndTime:      points[len(points)-1].ObservedAt,
		PointCount:   len(points),
		QualityScore: 0.85,
		SourceName:   "historical-store",
		Points:       points,
	}
}

func continuationTestSelection(
	request Request,
) projectionneighbors.Result {
	neighbors := make(
		[]projectionneighbors.Neighbor,
		0,
		len(request.Candidates),
	)
	for index, candidate := range request.Candidates {
		score := 0.9 -
			float64(index)*0.1
		anchor :=
			candidate.Points[4]
		neighbors = append(
			neighbors,
			projectionneighbors.Neighbor{
				TrajectoryID:    candidate.ID,
				SimilarityScore: score,
				SimilarityLevel: historicalsimilarity.
					LevelForScore(score),
				SimilarityInputFingerprint: "sha256:" +
					strings.Repeat(
						string(
							rune('a'+index),
						),
						64,
					),
				AnchorPointIndex:   4,
				AnchorObservedAt:   anchor.ObservedAt,
				AnchorDistanceKM:   float64(index + 1),
				CandidateStartTime: candidate.StartTime,
				CandidateEndTime:   candidate.EndTime,
				CandidateAge: request.AsOfTime.Sub(
					candidate.EndTime,
				),
				PrefixPointCount:       5,
				ContinuationPointCount: 2,
				ContinuationEndTime: anchor.ObservedAt.Add(
					2 * time.Minute,
				),
			},
		)
	}

	return projectionneighbors.Result{
		Version:                      projectionneighbors.Version,
		Status:                       projectionneighbors.StatusComplete,
		CurrentTrajectoryID:          request.CurrentTrajectory.ID,
		AsOfTime:                     request.AsOfTime,
		RequiredContinuationDuration: 2 * time.Minute,

		InputCandidateCount:     2,
		CheckedCandidateCount:   2,
		QualifiedCandidateCount: 2,
		RejectedCandidateCount:  0,

		SelectionLimit: 2,
		Neighbors:      neighbors,
		InputFingerprint: "sha256:" +
			strings.Repeat("e", 64),
	}
}

func continuationTestPattern(
	selection projectionneighbors.Result,
) projectionpatternconfidence.Result {
	return projectionpatternconfidence.Result{
		Version: projectionpatternconfidence.Version,
		Status: projectionpatternconfidence.
			StatusComplete,
		Usable: true,

		NeighborCount:       2,
		TargetNeighborCount: 2,

		MeanSimilarityScore:     0.85,
		MeanCandidateAgeSeconds: 24.5 * 60 * 60,
		MeanAnchorDistanceKM:    1.5,

		Score: 0.85,
		Level: projectioncontract.
			ConfidenceLevelHigh,

		Components: []projectionpatternconfidence.Component{
			{
				Name: projectionpatternconfidence.
					ComponentSimilarity,
				Score:  0.85,
				Weight: 0.25,
			},
			{
				Name: projectionpatternconfidence.
					ComponentSupport,
				Score:  1,
				Weight: 0.25,
			},
			{
				Name: projectionpatternconfidence.
					ComponentFreshness,
				Score:  0.8,
				Weight: 0.25,
			},
			{
				Name: projectionpatternconfidence.
					ComponentAnchorProximity,
				Score:  0.9,
				Weight: 0.25,
			},
		},
		SelectedTrajectoryIDs: []string{
			"historical-a",
			"historical-b",
		},
		InputFingerprint: "sha256:" +
			strings.Repeat("f", 64),
	}
}

func validFallbackResult(
	request Request,
) projectioncontract.Result {
	asOfTime := request.AsOfTime
	altitude := 1000.0
	verticalUncertainty := 50.0

	return projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        projectioncontract.ResultStatusComplete,
		TrajectoryID:  request.CurrentTrajectory.ID,
		Method: projectioncontract.Method{
			Name:    projectionbaseline.MethodName,
			Version: projectionbaseline.Version,
			DecisionClass: projectioncontract.
				DecisionClassPhysicsDerived,
		},
		Horizon: projectioncontract.Horizon{
			AsOfTime: asOfTime,
			EndTime: asOfTime.Add(
				2 * time.Minute,
			),
			Step: time.Minute,
		},
		Points: []projectioncontract.ProjectionPoint{
			{
				Sequence: 0,
				ForecastTime: asOfTime.Add(
					time.Minute,
				),
				Position: projectioncontract.Position{
					Latitude:  41.01,
					Longitude: 50.01,
					AltitudeM: &altitude,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 500,
					VerticalRadiusM:   &verticalUncertainty,
				},
				Confidence: projectioncontract.Confidence{
					Score: 0.7,
					Level: projectioncontract.
						ConfidenceLevelMedium,
				},
			},
			{
				Sequence: 1,
				ForecastTime: asOfTime.Add(
					2 * time.Minute,
				),
				Position: projectioncontract.Position{
					Latitude:  41.02,
					Longitude: 50.02,
					AltitudeM: &altitude,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 700,
					VerticalRadiusM:   &verticalUncertainty,
				},
				Confidence: projectioncontract.Confidence{
					Score: 0.6,
					Level: projectioncontract.
						ConfidenceLevelMedium,
				},
			},
		},
		Confidence: projectioncontract.Confidence{
			Score: 0.6,
			Level: projectioncontract.
				ConfidenceLevelMedium,
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
				Code:    "kinematic_baseline",
				Message: "Fallback uses the kinematic baseline.",
			},
		},
		ScopeGuard: projectioncontract.
			ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: "sha256:" +
				strings.Repeat("a", 64),
			Inputs: []projectioncontract.InputReference{
				{
					Name: "latest_trajectory_point",
					Classification: projectioncontract.
						InputClassificationObserved,
					ObservedAt: asOfTime,
				},
			},
			LatestInputObservedAt: asOfTime,
		},
		GeneratedAt: request.GeneratedAt,
	}
}

func continuationTestAsOfTime() time.Time {
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

func equalContinuationPoints(
	left []projectioncontract.ProjectionPoint,
	right []projectioncontract.ProjectionPoint,
) bool {
	if len(left) != len(right) {
		return false
	}

	for index := range left {
		if left[index].Sequence !=
			right[index].Sequence ||
			!left[index].ForecastTime.Equal(
				right[index].
					ForecastTime,
			) ||
			math.Abs(
				left[index].
					Position.Latitude-
					right[index].
						Position.Latitude,
			) > 1e-12 ||
			math.Abs(
				left[index].
					Position.Longitude-
					right[index].
						Position.Longitude,
			) > 1e-12 ||
			left[index].Uncertainty.
				HorizontalRadiusM !=
				right[index].Uncertainty.
					HorizontalRadiusM ||
			left[index].Confidence.Score !=
				right[index].
					Confidence.Score {
			return false
		}
	}

	return true
}

func hasProjectionLimitation(
	items []projectioncontract.Limitation,
	code string,
) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}

	return false
}

func hasFallbackReason(
	items []projectioncontract.Limitation,
	reason string,
) bool {
	for _, item := range items {
		if item.Code ==
			"historical_neighbor_fallback_reason" &&
			strings.Contains(
				item.Message,
				reason,
			) {
			return true
		}
	}

	return false
}

func hasProjectionExplanation(
	items []projectioncontract.Explanation,
	code string,
) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}

	return false
}
