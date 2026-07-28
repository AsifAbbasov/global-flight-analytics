package projectionbaseline

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func TestProjectRejectsDuplicateProjectionEligibilityDecision(t *testing.T) {
	config := validBaselineConfig()
	stub := config.EligibilityEvaluator.(*eligibilityEvaluatorStub)
	stub.evaluation.Decisions = append(
		stub.evaluation.Decisions,
		stub.evaluation.Decisions[0],
	)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = baseline.Project(baselineTestRequest())
	if !errors.Is(err, ErrEligibilityEvaluationInvalid) {
		t.Fatalf("Project() error = %v, want %v", err, ErrEligibilityEvaluationInvalid)
	}
}

func TestProjectRejectsUnknownEligibilityReason(t *testing.T) {
	config := validBaselineConfig()
	stub := config.EligibilityEvaluator.(*eligibilityEvaluatorStub)
	stub.evaluation.Decisions[0].Allowed = false
	stub.evaluation.Decisions[0].Reasons = []trajectoryeligibility.ReasonCode{
		"unknown_projection_reason",
	}
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = baseline.Project(baselineTestRequest())
	if !errors.Is(err, ErrEligibilityEvaluationInvalid) {
		t.Fatalf("Project() error = %v, want %v", err, ErrEligibilityEvaluationInvalid)
	}
}

func TestProjectAllowsExplicitHorizontalFallbackForMissingAltitude(t *testing.T) {
	config := validBaselineConfig()
	stub := config.EligibilityEvaluator.(*eligibilityEvaluatorStub)
	stub.evaluation.Decisions[0].Allowed = false
	stub.evaluation.Decisions[0].Reasons = []trajectoryeligibility.ReasonCode{
		trajectoryeligibility.ReasonMissingAltitude,
	}
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := baselineTestRequest()
	for index := range request.Trajectory.Points {
		request.Trajectory.Points[index].BarometricAltitudeM = 0
		request.Trajectory.Points[index].BarometricAltitudeStatus =
			flightstate.AltitudeStatusUnavailable
		request.Trajectory.Points[index].GeometricAltitudeM = 0
		request.Trajectory.Points[index].GeometricAltitudeStatus =
			flightstate.AltitudeStatusUnavailable
	}

	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if result.Status != projectioncontract.ResultStatusLimited ||
		!hasLimitation(result.Limitations, "projection_eligibility_altitude_fallback") ||
		!hasLimitation(result.Limitations, "projection_altitude_unavailable") ||
		len(result.Points) == 0 ||
		result.Points[0].Position.AltitudeM != nil {
		t.Fatalf("unexpected horizontal fallback result: %#v", result)
	}
}

func TestProjectRejectsConflictingLatestObservations(t *testing.T) {
	config := validBaselineConfig()
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := baselineTestRequest()
	conflict := request.Trajectory.Points[len(request.Trajectory.Points)-1]
	conflict.ID = "point-latest-conflict"
	conflict.FlightStateID = "state-latest-conflict"
	conflict.SourceName = "conflicting-source"
	conflict.Latitude += 0.5
	request.Trajectory.Points = append(request.Trajectory.Points, conflict)
	request.Trajectory.PointCount = len(request.Trajectory.Points)

	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if result.Status != projectioncontract.ResultStatusUnavailable ||
		!hasLimitation(result.Limitations, "projection_latest_observation_ambiguous") {
		t.Fatalf("unexpected ambiguous observation result: %#v", result)
	}
}

func TestProjectUsesStationaryOnGroundModel(t *testing.T) {
	config := validBaselineConfig()
	config.AllowOnGround = true
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := baselineTestRequest()
	latest := &request.Trajectory.Points[len(request.Trajectory.Points)-1]
	latest.OnGround = true
	latest.VelocityMPS = 20
	latest.VerticalRateMPS = 0
	latest.BarometricAltitudeM = 0
	latest.BarometricAltitudeStatus = flightstate.AltitudeStatusGround
	latest.GeometricAltitudeM = 0
	latest.GeometricAltitudeStatus = flightstate.AltitudeStatusGround

	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if result.Status != projectioncontract.ResultStatusLimited ||
		!hasLimitation(result.Limitations, "projection_on_ground_stationary_model") {
		t.Fatalf("unexpected on-ground status: %#v", result)
	}
	for _, point := range result.Points {
		if point.Position.Latitude != latest.Latitude ||
			point.Position.Longitude != latest.Longitude ||
			point.Position.AltitudeM == nil ||
			*point.Position.AltitudeM != 0 {
			t.Fatalf("on-ground point moved: %#v", point)
		}
	}
}

func TestDefaultEligibilityPolicyIsPublishedInProvenance(t *testing.T) {
	config := validBaselineConfig()
	config.EligibilityEvaluator = trajectoryeligibility.NewDefault()
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := baseline.Project(baselineTestRequest())
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}

	for _, input := range result.Provenance.Inputs {
		if input.Name == "eligibility_policy" {
			if !strings.Contains(
				input.SourceName,
				trajectoryeligibility.ProjectionPolicyVersion,
			) || !strings.Contains(input.Limitation, "sha256:") {
				t.Fatalf("unexpected eligibility provenance: %#v", input)
			}
			return
		}
	}
	t.Fatal("eligibility policy provenance is missing")
}

func TestHorizontalFallbackCanBeRejected(t *testing.T) {
	config := validBaselineConfig()
	config.HorizontalFallbackPolicy = HorizontalFallbackReject
	stub := config.EligibilityEvaluator.(*eligibilityEvaluatorStub)
	stub.evaluation.Decisions[0].Allowed = false
	stub.evaluation.Decisions[0].Reasons = []trajectoryeligibility.ReasonCode{
		trajectoryeligibility.ReasonMissingAltitude,
	}
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := baselineTestRequest()
	request.GeneratedAt = request.AsOfTime.Add(time.Second)
	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if result.Status != projectioncontract.ResultStatusUnavailable ||
		!hasLimitation(result.Limitations, "projection_eligibility_missing_altitude") {
		t.Fatalf("unexpected rejected fallback result: %#v", result)
	}
}
