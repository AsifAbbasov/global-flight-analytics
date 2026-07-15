package projectionbaseline

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

type eligibilityEvaluatorStub struct {
	evaluation trajectoryeligibility.Evaluation
	item       trajectory.FlightTrajectory
	now        time.Time
	calls      int
}

func (
	stub *eligibilityEvaluatorStub,
) Evaluate(
	item trajectory.FlightTrajectory,
	now time.Time,
) trajectoryeligibility.Evaluation {
	stub.calls++
	stub.item = item
	stub.now = now

	return stub.evaluation
}

func TestProjectBuildsDeterministicShortHorizonResult(
	t *testing.T,
) {
	config := validBaselineConfig()
	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := baselineTestRequest()
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

	if first.Status !=
		projectioncontract.ResultStatusComplete {
		t.Fatalf(
			"status = %q, want complete",
			first.Status,
		)
	}
	if len(first.Points) != 2 {
		t.Fatalf(
			"point count = %d, want 2",
			len(first.Points),
		)
	}
	if first.Provenance.InputFingerprint !=
		second.Provenance.InputFingerprint {
		t.Fatal(
			"deterministic request produced different fingerprints",
		)
	}
	if !equalProjectionPoints(
		first.Points,
		second.Points,
	) {
		t.Fatal(
			"deterministic request produced different points",
		)
	}

	const expectedFirstLongitude = 0.053959221
	if math.Abs(
		first.Points[0].Position.Longitude-
			expectedFirstLongitude,
	) > 1e-6 {
		t.Fatalf(
			"first longitude = %.12f, want %.12f",
			first.Points[0].Position.Longitude,
			expectedFirstLongitude,
		)
	}
	if first.Points[0].Position.AltitudeM == nil ||
		math.Abs(
			*first.Points[0].Position.
				AltitudeM-1120,
		) > 1e-9 {
		t.Fatalf(
			"first altitude = %#v, want 1120",
			first.Points[0].Position.AltitudeM,
		)
	}
	if math.Abs(
		first.Points[0].Uncertainty.
			HorizontalRadiusM-220,
	) > 1e-9 {
		t.Fatalf(
			"first horizontal uncertainty = %f, want 220",
			first.Points[0].Uncertainty.
				HorizontalRadiusM,
		)
	}
	if first.Points[0].Uncertainty.
		VerticalRadiusM == nil ||
		math.Abs(
			*first.Points[0].Uncertainty.
				VerticalRadiusM-50,
		) > 1e-9 {
		t.Fatalf(
			"first vertical uncertainty = %#v, want 50",
			first.Points[0].Uncertainty.
				VerticalRadiusM,
		)
	}
	if math.Abs(
		first.Points[0].Confidence.Score-
			0.675,
	) > 1e-9 ||
		first.Points[0].Confidence.Level !=
			projectioncontract.
				ConfidenceLevelMedium {
		t.Fatalf(
			"first confidence = %#v",
			first.Points[0].Confidence,
		)
	}
	if math.Abs(
		first.Confidence.Score-0.45,
	) > 1e-9 ||
		first.Confidence.Level !=
			projectioncontract.
				ConfidenceLevelLow {
		t.Fatalf(
			"result confidence = %#v",
			first.Confidence,
		)
	}
	if first.Method.DecisionClass !=
		projectioncontract.
			DecisionClassPhysicsDerived ||
		first.ScopeGuard !=
			projectioncontract.
				ScopeGuardResearchOnly {
		t.Fatalf(
			"unexpected method or scope guard: %#v %#v",
			first.Method,
			first.ScopeGuard,
		)
	}

	report := projectioncontract.Validate(first)
	if report.Status !=
		projectioncontract.
			ValidationStatusValid {
		t.Fatalf(
			"generated contract invalid: %#v",
			report.Issues,
		)
	}
}

func TestProjectIntegratesDefaultTrajectoryEligibility(
	t *testing.T,
) {
	config := validBaselineConfig()
	config.EligibilityEvaluator =
		trajectoryeligibility.NewDefault()

	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	result, err := baseline.Project(
		baselineTestRequest(),
	)
	if err != nil {
		t.Fatalf(
			"Project() error = %v",
			err,
		)
	}
	if result.Status !=
		projectioncontract.ResultStatusComplete {
		t.Fatalf(
			"status = %q, limitations = %#v",
			result.Status,
			result.Limitations,
		)
	}
}

func TestProjectReturnsUnavailableWhenEligibilityDenies(
	t *testing.T,
) {
	stub := allowedEligibilityStub()
	stub.evaluation.Decisions[0].Allowed =
		false
	stub.evaluation.Decisions[0].Reasons =
		[]trajectoryeligibility.ReasonCode{
			trajectoryeligibility.
				ReasonLowQualityScore,
		}

	config := validBaselineConfig()
	config.EligibilityEvaluator = stub
	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	result, err := baseline.Project(
		baselineTestRequest(),
	)
	if err != nil {
		t.Fatalf(
			"Project() error = %v",
			err,
		)
	}

	if result.Status !=
		projectioncontract.
			ResultStatusUnavailable ||
		len(result.Points) != 0 ||
		result.Confidence.Level !=
			projectioncontract.
				ConfidenceLevelNone ||
		!hasLimitation(
			result.Limitations,
			"projection_eligibility_low_quality_score",
		) {
		t.Fatalf(
			"unexpected unavailable result: %#v",
			result,
		)
	}
	if stub.calls != 1 {
		t.Fatalf(
			"eligibility calls = %d, want 1",
			stub.calls,
		)
	}
}

func TestProjectExcludesFutureTrajectoryPoints(
	t *testing.T,
) {
	config := validBaselineConfig()
	stub, ok := config.
		EligibilityEvaluator.(*eligibilityEvaluatorStub)
	if !ok {
		t.Fatal(
			"test eligibility evaluator type mismatch",
		)
	}
	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := baselineTestRequest()
	withoutFuture, err := baseline.Project(request)
	if err != nil {
		t.Fatalf(
			"Project() without future point error = %v",
			err,
		)
	}

	futurePoint := request.Trajectory.Points[len(request.Trajectory.Points)-1]
	futurePoint.ID = "point-future"
	futurePoint.ObservedAt =
		request.AsOfTime.Add(time.Minute)
	futurePoint.Latitude = 50
	futurePoint.Longitude = 60
	request.Trajectory.Points = append(
		request.Trajectory.Points,
		futurePoint,
	)
	request.Trajectory.PointCount =
		len(request.Trajectory.Points)
	request.Trajectory.EndTime =
		futurePoint.ObservedAt

	withFuture, err := baseline.Project(request)
	if err != nil {
		t.Fatalf(
			"Project() with future point error = %v",
			err,
		)
	}

	if withoutFuture.Provenance.
		InputFingerprint !=
		withFuture.Provenance.InputFingerprint {
		t.Fatal(
			"future point changed projection input fingerprint",
		)
	}
	if !equalProjectionPoints(
		withoutFuture.Points,
		withFuture.Points,
	) {
		t.Fatal(
			"future point changed projected points",
		)
	}
	if !hasLimitation(
		withFuture.Limitations,
		"future_observations_excluded",
	) {
		t.Fatalf(
			"future exclusion limitation missing: %#v",
			withFuture.Limitations,
		)
	}
	if stub.item.EndTime.After(
		request.AsOfTime,
	) {
		t.Fatalf(
			"eligibility received future end time: %s",
			stub.item.EndTime,
		)
	}
	for _, point := range stub.item.Points {
		if point.ObservedAt.After(
			request.AsOfTime,
		) {
			t.Fatalf(
				"eligibility received future point: %#v",
				point,
			)
		}
	}
}

func TestProjectHonorsOnGroundPolicy(
	t *testing.T,
) {
	config := validBaselineConfig()
	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := baselineTestRequest()
	request.Trajectory.Points[len(request.Trajectory.Points)-1].OnGround = true

	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf(
			"Project() error = %v",
			err,
		)
	}
	if result.Status !=
		projectioncontract.
			ResultStatusUnavailable ||
		!hasLimitation(
			result.Limitations,
			"projection_on_ground_not_allowed",
		) {
		t.Fatalf(
			"unexpected on-ground result: %#v",
			result,
		)
	}
}

func TestProjectMarksTruncatedHorizonLimited(
	t *testing.T,
) {
	config := validBaselineConfig()
	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := baselineTestRequest()
	request.RequestedDuration =
		10 * time.Minute

	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf(
			"Project() error = %v",
			err,
		)
	}

	if result.Status !=
		projectioncontract.ResultStatusLimited ||
		len(result.Points) != 5 ||
		!hasLimitation(
			result.Limitations,
			"projection_horizon_truncated",
		) {
		t.Fatalf(
			"unexpected truncated result: %#v",
			result,
		)
	}
}

func TestProjectCanReturnHorizontalOnlyLimitedResult(
	t *testing.T,
) {
	config := validBaselineConfig()
	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := baselineTestRequest()
	for index := range request.Trajectory.Points {
		request.Trajectory.Points[index].
			GeometricAltitudeM = 0
		request.Trajectory.Points[index].
			GeometricAltitudeStatus =
			flightstate.
				AltitudeStatusUnavailable
		request.Trajectory.Points[index].
			BarometricAltitudeM = 0
		request.Trajectory.Points[index].
			BarometricAltitudeStatus =
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
		result.Points[0].Position.AltitudeM != nil ||
		result.Points[0].Uncertainty.
			VerticalRadiusM != nil ||
		!hasLimitation(
			result.Limitations,
			"projection_altitude_unavailable",
		) {
		t.Fatalf(
			"unexpected horizontal-only result: %#v",
			result,
		)
	}
}

func TestProjectRejectsInvalidRequestMetadata(
	t *testing.T,
) {
	config := validBaselineConfig()
	baseline, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := baselineTestRequest()
	request.GeneratedAt =
		request.AsOfTime.Add(
			-time.Second,
		)
	_, err = baseline.Project(request)
	if !errors.Is(
		err,
		ErrGeneratedAtInvalid,
	) {
		t.Fatalf(
			"generated-at error = %v",
			err,
		)
	}

	request = baselineTestRequest()
	request.Trajectory.ID = ""
	_, err = baseline.Project(request)
	if !errors.Is(
		err,
		ErrTrajectoryIDRequired,
	) {
		t.Fatalf(
			"trajectory-id error = %v",
			err,
		)
	}
}

func validBaselineConfig() Config {
	policy, err := projectionhorizon.New(
		projectionhorizon.Config{
			Name:              "short-horizon-baseline-test",
			MinimumDuration:   time.Minute,
			DefaultDuration:   2 * time.Minute,
			MaximumDuration:   5 * time.Minute,
			Step:              time.Minute,
			MaximumPointCount: 5,
		},
	)
	if err != nil {
		panic(err)
	}

	return Config{
		HorizonPlanner:       policy,
		EligibilityEvaluator: allowedEligibilityStub(),

		InitialHorizontalUncertaintyM:  100,
		HorizontalUncertaintyGrowthMPS: 2,
		InitialVerticalUncertaintyM:    20,
		VerticalUncertaintyGrowthMPS:   0.5,

		MaximumConfidenceLoss: 0.5,

		MediumConfidenceMinimum: 0.6,
		HighConfidenceMinimum:   0.8,

		AllowOnGround: false,
	}
}

func allowedEligibilityStub() *eligibilityEvaluatorStub {
	return &eligibilityEvaluatorStub{
		evaluation: trajectoryeligibility.Evaluation{
			Decisions: []trajectoryeligibility.Decision{
				{
					Capability: trajectoryeligibility.
						CapabilityProjection,
					Allowed: true,
				},
			},
		},
	}
}

func baselineTestRequest() Request {
	asOfTime := baselineTestAsOfTime()

	return Request{
		Trajectory:        baselineTestTrajectory(),
		AsOfTime:          asOfTime,
		RequestedDuration: 2 * time.Minute,
		GeneratedAt: asOfTime.Add(
			time.Second,
		),
	}
}

func baselineTestTrajectory() trajectory.FlightTrajectory {
	asOfTime := baselineTestAsOfTime()
	points := make(
		[]trajectory.TrackPoint4D,
		0,
		5,
	)

	for index := 0; index < 5; index++ {
		observedAt := asOfTime.Add(
			time.Duration(index-4) *
				time.Minute,
		)
		points = append(
			points,
			trajectory.TrackPoint4D{
				ID:            fmtPointID(index),
				FlightStateID: fmtStateID(index),
				FlightID:      "123e4567-e89b-12d3-a456-426614174000",
				AircraftID:    "aircraft-001",
				ICAO24:        "4K1234",
				Callsign:      "AHY123",

				Latitude:  0,
				Longitude: 0,

				BarometricAltitudeM: 1000,
				BarometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,
				GeometricAltitudeM: 1000,
				GeometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,

				VelocityMPS:     100,
				HeadingDegrees:  90,
				VerticalRateMPS: 2,
				OnGround:        false,

				ObservedAt: observedAt,
				SourceName: "airplanes.live",
			},
		)
	}

	return trajectory.FlightTrajectory{
		ID: "trajectory-001",
		IdentityKey: "flight-identity-" +
			strings.Repeat("a", 64),
		IdentityBasis: trajectory.
			FlightIdentityBasisSourceFlightID,
		SplitReason: trajectory.
			FlightSplitReasonInitialObservation,
		FlightID:   "123e4567-e89b-12d3-a456-426614174000",
		AircraftID: "aircraft-001",
		ICAO24:     "4K1234",
		Callsign:   "AHY123",

		StartTime: points[0].ObservedAt,
		EndTime: points[len(points)-1].
			ObservedAt,
		DurationSeconds: int64(
			4 * time.Minute /
				time.Second,
		),
		PointCount:   len(points),
		SegmentCount: 1,

		CoverageGapCount: 0,
		QualityScore:     0.9,
		SourceName:       "airplanes.live",
		Points:           points,

		CreatedAt: points[0].ObservedAt,
		UpdatedAt: asOfTime,
	}
}

func baselineTestPlan() projectionhorizon.Plan {
	asOfTime := baselineTestAsOfTime()

	return projectionhorizon.Plan{
		Version:    projectionhorizon.Version,
		PolicyName: "short-horizon-baseline-test",
		AsOfTime:   asOfTime,
		EndTime: asOfTime.Add(
			2 * time.Minute,
		),
		Step:              time.Minute,
		RequestedDuration: 2 * time.Minute,
		EffectiveDuration: 2 * time.Minute,
		ForecastTimes: []time.Time{
			asOfTime.Add(time.Minute),
			asOfTime.Add(2 * time.Minute),
		},
	}
}

func baselineTestAsOfTime() time.Time {
	return time.Date(
		2026,
		time.July,
		15,
		17,
		0,
		0,
		0,
		time.UTC,
	)
}

func fmtPointID(
	index int,
) string {
	return "point-" +
		string(rune('0'+index))
}

func fmtStateID(
	index int,
) string {
	return "state-" +
		string(rune('0'+index))
}

func hasLimitation(
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

func equalProjectionPoints(
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
			left[index].Position.Latitude !=
				right[index].
					Position.Latitude ||
			left[index].Position.Longitude !=
				right[index].
					Position.Longitude ||
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
