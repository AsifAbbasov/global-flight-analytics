package projectioncontinuation

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

func TestComposeUncertaintyPreservesBothComponents(
	t *testing.T,
) {
	combined, valid := composeUncertainty(100, 40)
	if !valid || combined != 140 {
		t.Fatalf(
			"composeUncertainty() = %f, %t; want 140, true",
			combined,
			valid,
		)
	}
}

func TestEffectiveWeightedSupportPenalizesConcentration(
	t *testing.T,
) {
	balanced := effectiveWeightedSupportRatio(
		[]projectedSample{
			{weight: 1},
			{weight: 1},
		},
		2,
	)
	concentrated := effectiveWeightedSupportRatio(
		[]projectedSample{
			{weight: 1},
			{weight: 0.01},
		},
		2,
	)

	if math.Abs(balanced-1) > 1e-12 ||
		concentrated <= 0 ||
		concentrated >= balanced {
		t.Fatalf(
			"balanced=%f concentrated=%f",
			balanced,
			concentrated,
		)
	}
}

func TestNeighborDisagreementRaisesUncertaintyAndLowersConfidence(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	plan, err := config.HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime:          continuationTestAsOfTime(),
			RequestedDuration: 2 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	pattern := config.PatternConfidenceEvaluator.(*patternEvaluatorStub).result

	closeSamples := []projectedSample{
		{
			trajectoryID: "a",
			weight:       1,
			latitude:     41,
			longitude:    50,
		},
		{
			trajectoryID: "b",
			weight:       1,
			latitude:     41.0001,
			longitude:    50.0001,
		},
	}
	divergentSamples := append(
		[]projectedSample(nil),
		closeSamples...,
	)
	divergentSamples[1].longitude = 51

	closeResult, err := baseline.combineSamples(
		closeSamples,
		pattern,
		plan,
		0,
		plan.ForecastTimes[0],
	)
	if err != nil {
		t.Fatalf("close combineSamples() error = %v", err)
	}
	divergentResult, err := baseline.combineSamples(
		divergentSamples,
		pattern,
		plan,
		0,
		plan.ForecastTimes[0],
	)
	if err != nil {
		t.Fatalf("divergent combineSamples() error = %v", err)
	}

	if divergentResult.point.Uncertainty.HorizontalRadiusM <=
		closeResult.point.Uncertainty.HorizontalRadiusM {
		t.Fatalf(
			"divergent uncertainty=%f close=%f",
			divergentResult.point.Uncertainty.HorizontalRadiusM,
			closeResult.point.Uncertainty.HorizontalRadiusM,
		)
	}
	if divergentResult.point.Confidence.Score >=
		closeResult.point.Confidence.Score {
		t.Fatalf(
			"divergent confidence=%f close=%f",
			divergentResult.point.Confidence.Score,
			closeResult.point.Confidence.Score,
		)
	}
}

func TestZeroPointConfidenceMarksResultLimited(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := continuationTestRequest()
	plan, err := config.HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	selection := continuationTestSelection(request)
	pattern := continuationTestPattern(selection)
	preparation := baseline.prepareContinuation(
		request,
		plan,
		&ApprovedEvidence{
			Selection: selection,
			Pattern:   pattern,
		},
	)
	if preparation.requiresFallback() {
		t.Fatalf(
			"prepareContinuation() fallback = %s",
			preparation.fallbackReason,
		)
	}

	altitude := 1000.0
	vertical := 50.0
	points := make(
		[]projectioncontract.ProjectionPoint,
		0,
		len(plan.ForecastTimes),
	)
	for index, forecastTime := range plan.ForecastTimes {
		points = append(
			points,
			projectioncontract.ProjectionPoint{
				Sequence:     index,
				ForecastTime: forecastTime,
				Position: projectioncontract.Position{
					Latitude:  41 + float64(index)*0.01,
					Longitude: 50 + float64(index)*0.01,
					AltitudeM: &altitude,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 500,
					VerticalRadiusM:   &vertical,
				},
				Confidence: projectioncontract.Confidence{
					Score: 0,
					Level: projectioncontract.
						ConfidenceLevelNone,
				},
			},
		)
	}

	result := baseline.buildContinuationResult(
		preparation,
		plan,
		continuationPointResult{
			points:             points,
			altitudeComplete:   true,
			confidenceComplete: false,
		},
		request.GeneratedAt,
	)
	if result.Status != projectioncontract.ResultStatusLimited ||
		result.Confidence.Score != 0 ||
		result.Confidence.Level !=
			projectioncontract.ConfidenceLevelNone ||
		!hasProjectionLimitation(
			result.Limitations,
			"historical_continuation_confidence_none",
		) {
		t.Fatalf("unexpected zero-confidence result: %#v", result)
	}

	report := projectioncontract.Validate(result)
	if report.Status != projectioncontract.ValidationStatusValid {
		t.Fatalf("result invalid: %#v", report.Issues)
	}
}

func TestWeightedMeanRejectsNearAntipodalSamples(
	t *testing.T,
) {
	_, _, valid := weightedMeanGeoPoint(
		[]weightedGeoPoint{
			{latitude: 0, longitude: 0, weight: 1},
			{latitude: 0, longitude: 180, weight: 1},
		},
	)
	if valid {
		t.Fatal("near-antipodal weighted mean must be undefined")
	}
}

func TestFallbackPreservesUnderlyingCause(
	t *testing.T,
) {
	cause := errors.New("fallback cause")
	config := validContinuationConfig(t)
	config.FallbackProjector.(*fallbackProjectorStub).err = cause
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := continuationTestRequest()
	plan, planErr := config.HorizonPlanner.Build(projectionhorizon.Request{
		AsOfTime: request.AsOfTime, RequestedDuration: request.RequestedDuration,
	})
	if planErr != nil {
		t.Fatalf("Build() error = %v", planErr)
	}
	_, err = baseline.fallback(
		request,
		plan,
		"test_reason",
		"",
		"",
	)
	if !errors.Is(err, ErrFallbackProjectionFailed) ||
		!errors.Is(err, cause) {
		t.Fatalf("fallback error = %v", err)
	}
}

func TestStandaloneProjectRejectsCandidateEvidenceDrift(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	fallback := config.FallbackProjector.(*fallbackProjectorStub)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := continuationTestRequest()
	request.Candidates[0].Points[4].ObservedAt =
		request.Candidates[0].Points[4].ObservedAt.
			Add(time.Second)

	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if fallback.calls != 1 ||
		!hasFallbackReason(
			result.Limitations,
			"historical_selected_candidate_evidence_mismatch",
		) {
		t.Fatalf("unexpected mismatch result: %#v", result)
	}
}

func TestContinuationFingerprintDistinguishesTruncatedPlan(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	request := continuationTestRequest()
	selection := continuationTestSelection(request)
	pattern := continuationTestPattern(selection)

	exactPlan, err := config.HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime:          request.AsOfTime,
			RequestedDuration: 5 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("exact Build() error = %v", err)
	}
	truncatedPlan, err := config.HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime:          request.AsOfTime,
			RequestedDuration: 6 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("truncated Build() error = %v", err)
	}
	if !truncatedPlan.Truncated ||
		exactPlan.EffectiveDuration !=
			truncatedPlan.EffectiveDuration {
		t.Fatalf(
			"unexpected plans: exact=%#v truncated=%#v",
			exactPlan,
			truncatedPlan,
		)
	}

	exactFingerprint := continuationFingerprint(
		request.CurrentTrajectory,
		selection,
		pattern,
		exactPlan,
		config,
	)
	truncatedFingerprint := continuationFingerprint(
		request.CurrentTrajectory,
		selection,
		pattern,
		truncatedPlan,
		config,
	)
	if exactFingerprint == truncatedFingerprint {
		t.Fatal("truncated horizon reused exact-plan fingerprint")
	}
}

type irregularContinuationHorizonPlanner struct {
	plan projectionhorizon.Plan
}

func (planner irregularContinuationHorizonPlanner) Build(
	projectionhorizon.Request,
) (projectionhorizon.Plan, error) {
	return planner.plan.Clone(), nil
}

func TestProjectRejectsIrregularForecastTimes(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	request := continuationTestRequest()
	plan, err := config.HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	plan.ForecastTimes[0] = plan.ForecastTimes[0].Add(time.Second)
	config.HorizonPlanner = irregularContinuationHorizonPlanner{
		plan: plan,
	}
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = baseline.Project(request)
	if !errors.Is(err, ErrHorizonPlanInvalid) {
		t.Fatalf("Project() error = %v, want %v", err, ErrHorizonPlanInvalid)
	}
}

func TestTrajectorySnapshotOrdersEqualTimestampsCanonically(
	t *testing.T,
) {
	observedAt := continuationTestAsOfTime()
	first := trajectory.TrackPoint4D{
		ID:         "a",
		Latitude:   41,
		Longitude:  50,
		ObservedAt: observedAt,
	}
	second := trajectory.TrackPoint4D{
		ID:         "b",
		Latitude:   42,
		Longitude:  51,
		ObservedAt: observedAt,
	}

	left := trajectorySnapshotAt(
		trajectory.FlightTrajectory{
			ID:     "trajectory",
			Points: []trajectory.TrackPoint4D{second, first},
		},
		observedAt,
	)
	right := trajectorySnapshotAt(
		trajectory.FlightTrajectory{
			ID:     "trajectory",
			Points: []trajectory.TrackPoint4D{first, second},
		},
		observedAt,
	)
	if len(left.Points) != 2 || len(right.Points) != 2 ||
		left.Points[0].ID != "a" || right.Points[0].ID != "a" ||
		left.Points[1].ID != "b" || right.Points[1].ID != "b" {
		t.Fatalf("left=%#v right=%#v", left.Points, right.Points)
	}
}
