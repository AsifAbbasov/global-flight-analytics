package metricexecution

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestActiveAircraftCountsOneAircraftAcrossMultipleFlightIdentities(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	older := healthyMetricTrajectory(
		"a",
		"ABC123",
	)
	newer := healthyMetricTrajectory(
		"b",
		"ABC123",
	)
	newer.EndTime = older.EndTime.Add(
		time.Minute,
	)

	execution, err := service.ActiveAircraft(
		context.Background(),
		ActiveAircraftRequest{
			Trajectories: []trajectory.FlightTrajectory{
				older,
				newer,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"expected active aircraft execution, got %v",
			err,
		)
	}

	if execution.Result.Value != 1 {
		t.Fatalf(
			"expected one aircraft, got %d",
			execution.Result.Value,
		)
	}

	if execution.Result.Status !=
		analyticalresult.StatusLimited {
		t.Fatalf(
			"expected deduplication warning to limit result, got %s",
			execution.Result.Status,
		)
	}

	if !containsNotice(
		execution.Result.Warnings,
		nil,
		NoticeCodeDuplicateTrajectoriesRemoved,
	) {
		t.Fatalf(
			"expected aircraft deduplication warning, got %#v",
			execution.Result.Warnings,
		)
	}
}

func TestUniqueAircraftTrajectoriesKeepsNewestTrajectory(
	t *testing.T,
) {
	older := trajectory.FlightTrajectory{
		ID:           "older",
		ICAO24:       "ABC123",
		EndTime:      metricTestTime().Add(-time.Minute),
		QualityScore: 0.90,
	}
	newer := trajectory.FlightTrajectory{
		ID:           "newer",
		ICAO24:       "abc123",
		EndTime:      metricTestTime(),
		QualityScore: 0.80,
	}

	items, removed := uniqueAircraftTrajectories(
		[]trajectory.FlightTrajectory{
			older,
			newer,
		},
	)

	if removed != 1 ||
		len(items) != 1 ||
		items[0].ID != "newer" {
		t.Fatalf(
			"expected newest trajectory, got removed=%d items=%#v",
			removed,
			items,
		)
	}
}
