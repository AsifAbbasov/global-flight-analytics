package dataqualityintegration

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestBuildTrajectoryReportProducesVersionedEvidence(
	t *testing.T,
) {
	evaluatedAt := integrationTestTime()
	items := []trajectory.FlightTrajectory{
		integrationTrajectory("trajectory-b", "ABC002", evaluatedAt),
		integrationTrajectory("trajectory-a", "ABC001", evaluatedAt),
	}
	for itemIndex := range items {
		for pointIndex := range items[itemIndex].Points {
			items[itemIndex].Points[pointIndex].OnGround = true
		}
	}

	report, err := BuildTrajectoryReport(
		TrajectoryReportRequest{
			Trajectories: items,
			EvaluatedAt:  evaluatedAt,
		},
	)
	if err != nil {
		t.Fatalf("build trajectory report: %v", err)
	}
	if report == nil {
		t.Fatal("expected data-quality report")
	}
	if report.Freshness.Score != 1 {
		t.Fatalf(
			"expected fresh score 1, got %f",
			report.Freshness.Score,
		)
	}
	if report.SamplingDensity.Score <= 0 {
		t.Fatalf(
			"expected positive sampling density, got %f",
			report.SamplingDensity.Score,
		)
	}
	if report.Provenance.InputFingerprint == "" {
		t.Fatal("expected deterministic input fingerprint")
	}
	if !report.Permissions.PhaseDetection.Allowed {
		t.Fatalf(
			"expected implemented phase detection to be allowed, got %#v",
			report.Permissions.PhaseDetection.Reasons,
		)
	}
	if containsContractNotice(
		report.Limitations,
		"phase_detection_not_implemented",
	) {
		t.Fatal(
			"phase-detection pending limitation must be removed after integration",
		)
	}
	if report.Permissions.HistoricalSimilarity.Allowed {
		t.Fatal("expected historical similarity to remain denied")
	}
	if err := report.Validate(); err != nil {
		t.Fatalf("validate report: %v", err)
	}
}

func TestBuildTrajectoryReportFingerprintIgnoresInputOrder(
	t *testing.T,
) {
	evaluatedAt := integrationTestTime()
	first := integrationTrajectory("trajectory-a", "ABC001", evaluatedAt)
	second := integrationTrajectory("trajectory-b", "ABC002", evaluatedAt)

	left, err := BuildTrajectoryReport(
		TrajectoryReportRequest{
			Trajectories: []trajectory.FlightTrajectory{first, second},
			EvaluatedAt:  evaluatedAt,
		},
	)
	if err != nil {
		t.Fatalf("build left report: %v", err)
	}

	right, err := BuildTrajectoryReport(
		TrajectoryReportRequest{
			Trajectories: []trajectory.FlightTrajectory{second, first},
			EvaluatedAt:  evaluatedAt,
		},
	)
	if err != nil {
		t.Fatalf("build right report: %v", err)
	}

	if left.Provenance.InputFingerprint !=
		right.Provenance.InputFingerprint {
		t.Fatalf(
			"expected stable fingerprint, got %q and %q",
			left.Provenance.InputFingerprint,
			right.Provenance.InputFingerprint,
		)
	}
}

func TestBuildTrajectoryReportExplainsFutureObservationExclusion(
	t *testing.T,
) {
	evaluatedAt := integrationTestTime()
	item := integrationTrajectory(
		"trajectory-a",
		"ABC001",
		evaluatedAt,
	)
	item.Points = append(
		item.Points,
		trajectory.TrackPoint4D{
			ID:         "future-point",
			ObservedAt: evaluatedAt.Add(time.Minute),
			SourceName: "airplanes.live",
		},
	)

	report, err := BuildTrajectoryReport(
		TrajectoryReportRequest{
			Trajectories: []trajectory.FlightTrajectory{item},
			EvaluatedAt:  evaluatedAt,
		},
	)
	if err != nil {
		t.Fatalf("build report: %v", err)
	}

	if !containsContractNotice(
		report.Warnings,
		NoticeCodeFutureObservationsExcluded,
	) {
		t.Fatalf(
			"expected future observation warning, got %#v",
			report.Warnings,
		)
	}
}

func TestBuildTrajectoryReportHandlesEmptyAndInvalidInputs(
	t *testing.T,
) {
	report, err := BuildTrajectoryReport(
		TrajectoryReportRequest{},
	)
	if err != nil || report != nil {
		t.Fatalf(
			"expected empty input to produce no report, got report=%#v err=%v",
			report,
			err,
		)
	}

	_, err = BuildTrajectoryReport(
		TrajectoryReportRequest{
			Trajectories: []trajectory.FlightTrajectory{
				{ID: "trajectory-a"},
			},
			EvaluatedAt: integrationTestTime(),
		},
	)
	if !errors.Is(err, ErrNoUsableObservationTimes) {
		t.Fatalf(
			"expected no usable observation error, got %v",
			err,
		)
	}
}

func integrationTrajectory(
	id string,
	icao24 string,
	evaluatedAt time.Time,
) trajectory.FlightTrajectory {
	start := evaluatedAt.Add(-40 * time.Second)
	points := make([]trajectory.TrackPoint4D, 0, 4)
	for index := 0; index < 4; index++ {
		points = append(
			points,
			trajectory.TrackPoint4D{
				ID:            id + "-point",
				FlightStateID: id + "-state",
				ICAO24:        icao24,
				Callsign:      "TEST01",
				ObservedAt: start.Add(
					time.Duration(index) *
						10 *
						time.Second,
				),
				SourceName: "airplanes.live",
			},
		)
	}

	return trajectory.FlightTrajectory{
		ID:               id,
		IdentityKey:      "flight-identity-" + id,
		IdentityBasis:    trajectory.FlightIdentityBasisSourceFlightID,
		FlightID:         "flight-" + id,
		ICAO24:           icao24,
		Callsign:         "TEST01",
		StartTime:        start,
		EndTime:          evaluatedAt.Add(-10 * time.Second),
		DurationSeconds:  30,
		SegmentCount:     1,
		PointCount:       len(points),
		CoverageGapCount: 0,
		QualityScore:     0.95,
		SourceName:       "airplanes.live",
		Points:           points,
		CreatedAt:        evaluatedAt.Add(-5 * time.Second),
		UpdatedAt:        evaluatedAt.Add(-5 * time.Second),
	}
}

func integrationTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		15,
		0,
		0,
		0,
		0,
		time.UTC,
	)
}

func containsContractNotice(
	values []dataqualitycontract.Notice,
	code string,
) bool {
	for _, value := range values {
		if value.Code == code {
			return true
		}
	}
	return false
}
