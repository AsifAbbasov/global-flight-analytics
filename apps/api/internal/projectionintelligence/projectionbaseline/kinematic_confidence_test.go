package projectionbaseline

import (
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

func TestProjectConfidenceDecreasesWithObservationAge(t *testing.T) {
	config := validBaselineConfig()
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	freshRequest := baselineTestRequest()
	fresh, err := baseline.Project(freshRequest)
	if err != nil {
		t.Fatalf("fresh Project() error = %v", err)
	}

	staleRequest := baselineTestRequest()
	shift := 90 * time.Second
	for index := range staleRequest.Trajectory.Points {
		staleRequest.Trajectory.Points[index].ObservedAt =
			staleRequest.Trajectory.Points[index].ObservedAt.Add(-shift)
	}
	for index := range staleRequest.Trajectory.Segments {
		staleRequest.Trajectory.Segments[index].StartTime =
			staleRequest.Trajectory.Segments[index].StartTime.Add(-shift)
		staleRequest.Trajectory.Segments[index].EndTime =
			staleRequest.Trajectory.Segments[index].EndTime.Add(-shift)
	}
	staleRequest.Trajectory.StartTime =
		staleRequest.Trajectory.StartTime.Add(-shift)
	staleRequest.Trajectory.EndTime =
		staleRequest.Trajectory.EndTime.Add(-shift)

	stale, err := baseline.Project(staleRequest)
	if err != nil {
		t.Fatalf("stale Project() error = %v", err)
	}

	if stale.Points[0].Confidence.Score >= fresh.Points[0].Confidence.Score {
		t.Fatalf(
			"stale confidence = %f, fresh confidence = %f",
			stale.Points[0].Confidence.Score,
			fresh.Points[0].Confidence.Score,
		)
	}
	if stale.Provenance.InputFingerprint == fresh.Provenance.InputFingerprint {
		t.Fatal("observation age evidence did not change the fingerprint")
	}
}

func TestProjectRejectsOutOfBoundsKinematics(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		mutate func(*Request)
	}{
		{
			name: "ground speed",
			code: "projection_velocity_out_of_bounds",
			mutate: func(request *Request) {
				request.Trajectory.Points[len(request.Trajectory.Points)-1].VelocityMPS =
					maximumSupportedGroundSpeedMPS + 1
			},
		},
		{
			name: "heading",
			code: "projection_heading_out_of_bounds",
			mutate: func(request *Request) {
				request.Trajectory.Points[len(request.Trajectory.Points)-1].HeadingDegrees = 360
			},
		},
		{
			name: "vertical rate",
			code: "projection_vertical_rate_out_of_bounds",
			mutate: func(request *Request) {
				request.Trajectory.Points[len(request.Trajectory.Points)-1].VerticalRateMPS =
					maximumSupportedAbsoluteVerticalRateMPS + 1
			},
		},
		{
			name: "altitude",
			code: "projection_altitude_out_of_bounds",
			mutate: func(request *Request) {
				latest := &request.Trajectory.Points[len(request.Trajectory.Points)-1]
				latest.GeometricAltitudeM = maximumSupportedAltitudeM + 1
				latest.GeometricAltitudeStatus = flightstate.AltitudeStatusObserved
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := validBaselineConfig()
			baseline, err := New(config)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			request := baselineTestRequest()
			test.mutate(&request)

			result, err := baseline.Project(request)
			if err != nil {
				t.Fatalf("Project() error = %v", err)
			}
			if result.Status != projectioncontract.ResultStatusUnavailable ||
				!hasLimitation(result.Limitations, test.code) {
				t.Fatalf("unexpected result: %#v", result)
			}
		})
	}
}

func TestProjectReportsSelectedAltitudeReference(t *testing.T) {
	config := validBaselineConfig()
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	geometric, err := baseline.Project(baselineTestRequest())
	if err != nil {
		t.Fatalf("geometric Project() error = %v", err)
	}
	if !hasAltitudeReference(geometric, "geometric") {
		t.Fatalf("geometric altitude provenance missing: %#v", geometric.Provenance.Inputs)
	}

	barometricRequest := baselineTestRequest()
	for index := range barometricRequest.Trajectory.Points {
		barometricRequest.Trajectory.Points[index].GeometricAltitudeM = 0
		barometricRequest.Trajectory.Points[index].GeometricAltitudeStatus =
			flightstate.AltitudeStatusUnavailable
	}
	barometric, err := baseline.Project(barometricRequest)
	if err != nil {
		t.Fatalf("barometric Project() error = %v", err)
	}
	if !hasAltitudeReference(barometric, "barometric") {
		t.Fatalf("barometric altitude provenance missing: %#v", barometric.Provenance.Inputs)
	}
}

func TestProjectRejectsUnsafeOnGroundMotionWhenAllowed(t *testing.T) {
	config := validBaselineConfig()
	config.AllowOnGround = true
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := baselineTestRequest()
	latest := &request.Trajectory.Points[len(request.Trajectory.Points)-1]
	latest.OnGround = true
	latest.VelocityMPS = maximumSupportedOnGroundSpeedMPS + 1

	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if result.Status != projectioncontract.ResultStatusUnavailable ||
		!hasLimitation(result.Limitations, "projection_on_ground_motion_out_of_bounds") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestNilBaselineReturnsLifecycleError(t *testing.T) {
	var baseline *Baseline
	_, err := baseline.Project(baselineTestRequest())
	if err != ErrBaselineUnavailable {
		t.Fatalf("error = %v, want %v", err, ErrBaselineUnavailable)
	}
}

func TestConfigRejectsNegativeMaximumObservationAge(t *testing.T) {
	config := validBaselineConfig()
	config.MaximumObservationAge = -time.Second
	if err := config.Validate(); err != ErrMaximumObservationAgeInvalid {
		t.Fatalf("error = %v, want %v", err, ErrMaximumObservationAgeInvalid)
	}
}

func TestInputFingerprintCoversObservationAgePolicy(t *testing.T) {
	config := validBaselineConfig()
	request := baselineTestRequest()
	plan, err := config.HorizonPlanner.Build(baselineTestHorizonRequest(request))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	point := request.Trajectory.Points[len(request.Trajectory.Points)-1]
	original := inputFingerprint(request.Trajectory, point, plan, config)

	config.MaximumObservationAge = 3 * time.Minute
	changed := inputFingerprint(request.Trajectory, point, plan, config)
	if original == changed {
		t.Fatal("maximum observation age change did not change fingerprint")
	}
}

func baselineTestHorizonRequest(request Request) projectionhorizon.Request {
	return projectionhorizon.Request{
		AsOfTime:          request.AsOfTime,
		RequestedDuration: request.RequestedDuration,
	}
}

func hasAltitudeReference(result projectioncontract.Result, reference string) bool {
	for _, input := range result.Provenance.Inputs {
		if input.Name == "altitude" &&
			strings.Contains(strings.ToLower(input.Limitation), reference) {
			return true
		}
	}
	return false
}
