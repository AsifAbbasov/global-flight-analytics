package dataqualityintegration

import (
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestBuildTrajectoryReportDeniesPhaseDetectionForUnknownSignals(
	t *testing.T,
) {
	evaluatedAt := integrationTestTime()
	item := integrationTrajectory(
		"trajectory-unknown",
		"ABC999",
		evaluatedAt,
	)

	for index := range item.Points {
		item.Points[index].OnGround = false
		item.Points[index].VerticalRateMPS = math.NaN()
	}

	report, err := BuildTrajectoryReport(
		TrajectoryReportRequest{
			Trajectories: []trajectory.FlightTrajectory{
				item,
			},
			EvaluatedAt: evaluatedAt,
		},
	)
	if err != nil {
		t.Fatalf(
			"build unknown-signal data-quality report: %v",
			err,
		)
	}

	if report.Permissions.PhaseDetection.Allowed {
		t.Fatal(
			"expected phase detection to be denied for unknown-only evidence",
		)
	}
	if !containsPermissionReason(
		report.Permissions.PhaseDetection,
		PermissionReasonPhaseDetectionOnlyUnknownPoints,
	) {
		t.Fatalf(
			"expected unknown-only denial reason, got %#v",
			report.Permissions.PhaseDetection.Reasons,
		)
	}
	if !containsContractNotice(
		report.Limitations,
		LimitationCodePhaseDetectionUnavailable,
	) {
		t.Fatalf(
			"expected unavailable phase-detection limitation, got %#v",
			report.Limitations,
		)
	}
}

func TestBuildTrajectoryReportAllowsPartialPhaseDetection(
	t *testing.T,
) {
	evaluatedAt := integrationTestTime()
	item := integrationTrajectory(
		"trajectory-partial",
		"ABC998",
		evaluatedAt,
	)

	item.Points[0].OnGround = true
	for index := 1; index < len(item.Points); index++ {
		item.Points[index].OnGround = false
		item.Points[index].VerticalRateMPS = math.NaN()
	}

	report, err := BuildTrajectoryReport(
		TrajectoryReportRequest{
			Trajectories: []trajectory.FlightTrajectory{
				item,
			},
			EvaluatedAt: evaluatedAt,
		},
	)
	if err != nil {
		t.Fatalf(
			"build partial phase-detection report: %v",
			err,
		)
	}

	if !report.Permissions.PhaseDetection.Allowed {
		t.Fatalf(
			"expected phase detection to be allowed with partial evidence, got %#v",
			report.Permissions.PhaseDetection.Reasons,
		)
	}
	if !containsContractNotice(
		report.Limitations,
		LimitationCodePhaseDetectionPartial,
	) {
		t.Fatalf(
			"expected partial phase-detection limitation, got %#v",
			report.Limitations,
		)
	}
}

func containsPermissionReason(
	permission dataqualitycontract.Permission,
	reason string,
) bool {
	for _, current := range permission.Reasons {
		if current == reason {
			return true
		}
	}

	return false
}
