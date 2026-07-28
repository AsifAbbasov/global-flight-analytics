package projectionbaseline

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

func TestProjectExcludesFutureAggregateQualityEvidence(t *testing.T) {
	config := validBaselineConfig()
	stub := config.EligibilityEvaluator.(*eligibilityEvaluatorStub)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := baselineTestRequest()
	withoutFuture, err := baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() without future evidence error = %v", err)
	}

	futurePoint := request.Trajectory.Points[len(request.Trajectory.Points)-1]
	futurePoint.ID = "point-future-quality"
	futurePoint.FlightStateID = "state-future-quality"
	futurePoint.ObservedAt = request.AsOfTime.Add(time.Minute)
	futurePoint.SourceName = "future-source"
	request.Trajectory.Points = append(request.Trajectory.Points, futurePoint)
	request.Trajectory.PointCount = len(request.Trajectory.Points)
	request.Trajectory.EndTime = futurePoint.ObservedAt
	request.Trajectory.Segments = append(
		request.Trajectory.Segments,
		trajectory.TrajectorySegment{
			ID:              "segment-future",
			SequenceNumber:  2,
			Status:          trajectory.SegmentStatusObserved,
			QualityScore:    0.1,
			StartTime:       futurePoint.ObservedAt,
			EndTime:         futurePoint.ObservedAt,
			DurationSeconds: 0,
			PointCount:      1,
			SourceName:      futurePoint.SourceName,
		},
	)
	request.Trajectory.SegmentCount = len(request.Trajectory.Segments)
	request.Trajectory.QualityScore = 0.5

	withFuture, err := baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() with future evidence error = %v", err)
	}

	if withoutFuture.Provenance.InputFingerprint !=
		withFuture.Provenance.InputFingerprint {
		t.Fatal("future aggregate evidence changed the cutoff fingerprint")
	}
	if !equalProjectionPoints(withoutFuture.Points, withFuture.Points) {
		t.Fatal("future aggregate evidence changed projected points")
	}
	if stub.item.QualityScore != 0.9 {
		t.Fatalf("cutoff quality = %f, want 0.9", stub.item.QualityScore)
	}
	if stub.item.SegmentCount != 1 || len(stub.item.Segments) != 1 {
		t.Fatalf("cutoff segments = %#v", stub.item.Segments)
	}
}

func TestProjectExcludesCoverageGapNotCompletedAtCutoff(t *testing.T) {
	config := validBaselineConfig()
	stub := config.EligibilityEvaluator.(*eligibilityEvaluatorStub)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := baselineTestRequest()
	request.Trajectory.CoverageGaps = []trajectory.CoverageGap{
		{
			ID:        "future-gap",
			StartTime: request.AsOfTime.Add(-30 * time.Second),
			EndTime:   request.AsOfTime.Add(30 * time.Second),
			Reason:    trajectory.CoverageGapReasonTimeGap,
		},
	}
	request.Trajectory.CoverageGapCount = 1

	_, err = baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if stub.item.CoverageGapCount != 0 || len(stub.item.CoverageGaps) != 0 {
		t.Fatalf("future-created gap reached eligibility: %#v", stub.item.CoverageGaps)
	}
}

func TestUnavailableResultPreservesCutoffProvenance(t *testing.T) {
	stub := allowedEligibilityStub()
	stub.evaluation.Decisions[0].Allowed = false
	stub.evaluation.Decisions[0].Reasons = []trajectoryeligibility.ReasonCode{
		trajectoryeligibility.ReasonLowQualityScore,
	}

	config := validBaselineConfig()
	config.EligibilityEvaluator = stub
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := baselineTestRequest()
	result, err := baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}

	if result.Provenance.InputFingerprint == "" {
		t.Fatal("unavailable result fingerprint is empty")
	}
	if len(result.Provenance.Inputs) == 0 {
		t.Fatal("unavailable result inputs are empty")
	}
	if !result.Provenance.LatestInputObservedAt.Equal(request.AsOfTime) {
		t.Fatalf(
			"latest input observed at = %s, want %s",
			result.Provenance.LatestInputObservedAt,
			request.AsOfTime,
		)
	}
}

func TestInputFingerprintCoversSourceAndGroundState(t *testing.T) {
	config := validBaselineConfig()
	request := baselineTestRequest()
	plan, err := config.HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	item := request.Trajectory
	point := item.Points[len(item.Points)-1]
	original := inputFingerprint(item, point, plan, config)

	item.Points[len(item.Points)-1].SourceName = "changed-source"
	point = item.Points[len(item.Points)-1]
	withSourceChange := inputFingerprint(item, point, plan, config)
	if original == withSourceChange {
		t.Fatal("source-name change did not change fingerprint")
	}

	item = request.Trajectory
	item.Points[len(item.Points)-1].OnGround = true
	point = item.Points[len(item.Points)-1]
	withGroundChange := inputFingerprint(item, point, plan, config)
	if original == withGroundChange {
		t.Fatal("on-ground change did not change fingerprint")
	}
}
