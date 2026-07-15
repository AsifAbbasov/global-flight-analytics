package dataqualityintegration

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestBuildTrajectoryReportAllowsHistoricalSimilarity(
	t *testing.T,
) {
	evaluatedAt := integrationTestTime()
	first := integrationTrajectory(
		"similarity-a",
		"ABC001",
		evaluatedAt,
	)
	second := integrationTrajectory(
		"similarity-b",
		"ABC002",
		evaluatedAt,
	)

	for index := range first.Points {
		first.Points[index].Latitude =
			40 + float64(index)*0.1
		first.Points[index].Longitude =
			49 + float64(index)*0.1
		second.Points[index].Latitude =
			first.Points[index].Latitude
		second.Points[index].Longitude =
			first.Points[index].Longitude
	}

	report, err := BuildTrajectoryReport(
		TrajectoryReportRequest{
			Trajectories: []trajectory.FlightTrajectory{
				first,
				second,
			},
			EvaluatedAt: evaluatedAt,
		},
	)
	if err != nil {
		t.Fatalf(
			"build similarity-enabled report: %v",
			err,
		)
	}

	if !report.Permissions.
		HistoricalSimilarity.Allowed {
		t.Fatalf(
			"expected historical similarity permission, got %#v",
			report.Permissions.
				HistoricalSimilarity.Reasons,
		)
	}
	if containsContractNotice(
		report.Limitations,
		"historical_similarity_not_implemented",
	) {
		t.Fatal(
			"historical similarity pending limitation must be removed",
		)
	}
}

func TestBuildTrajectoryReportDeniesHistoricalSimilarityForOneTrajectory(
	t *testing.T,
) {
	evaluatedAt := integrationTestTime()
	item := integrationTrajectory(
		"similarity-one",
		"ABC001",
		evaluatedAt,
	)

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
			"build single-trajectory report: %v",
			err,
		)
	}

	if report.Permissions.
		HistoricalSimilarity.Allowed {
		t.Fatal(
			"expected historical similarity to be denied for one trajectory",
		)
	}
	if len(
		report.Permissions.
			HistoricalSimilarity.Reasons,
	) != 1 ||
		report.Permissions.
			HistoricalSimilarity.Reasons[0] !=
			PermissionReasonHistoricalSimilarityRequiresTwoTrajectories {
		t.Fatalf(
			"unexpected historical similarity reasons: %#v",
			report.Permissions.
				HistoricalSimilarity.Reasons,
		)
	}
}
