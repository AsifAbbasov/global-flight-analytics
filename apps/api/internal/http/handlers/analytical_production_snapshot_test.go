package handlers

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricquery"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestBuildProductionQualitySnapshotUsesCoveredIntervals(
	t *testing.T,
) {
	window := metricquery.Window{
		ObservedFrom: time.Date(
			2026,
			time.July,
			24,
			12,
			0,
			0,
			0,
			time.UTC,
		),
		ObservedTo: time.Date(
			2026,
			time.July,
			24,
			12,
			1,
			0,
			0,
			time.UTC,
		),
		Limit: productionQualityResultLimit,
	}

	firstObservation := window.ObservedFrom.
		Add(5 * time.Second)
	secondObservation := window.ObservedFrom.
		Add(15 * time.Second)
	fallbackObservation := window.ObservedFrom.
		Add(45 * time.Second)

	metricSnapshot, err :=
		buildProductionQualitySnapshot(
			[]trajectory.FlightTrajectory{
				{
					Points: []trajectory.TrackPoint4D{
						{
							ObservedAt: firstObservation,
						},
						{
							ObservedAt: firstObservation.
								Add(time.Second),
						},
						{
							ObservedAt: secondObservation,
						},
					},
				},
				{
					EndTime: fallbackObservation,
				},
			},
			window,
		)
	if err != nil {
		t.Fatalf(
			"expected server-owned snapshot, got %v",
			err,
		)
	}

	if metricSnapshot.ObservedSamples != 3 ||
		metricSnapshot.ExpectedSamples != 6 {
		t.Fatalf(
			"expected three covered intervals out of six, got %#v",
			metricSnapshot,
		)
	}

	if !metricSnapshot.Time.Equal(
		fallbackObservation,
	) {
		t.Fatalf(
			"expected latest observation %s, got %s",
			fallbackObservation,
			metricSnapshot.Time,
		)
	}
}

func TestBuildProductionQualitySnapshotPreservesNoObservationState(
	t *testing.T,
) {
	window := metricquery.Window{
		ObservedFrom: time.Date(
			2026,
			time.July,
			24,
			12,
			0,
			0,
			0,
			time.UTC,
		),
		ObservedTo: time.Date(
			2026,
			time.July,
			24,
			12,
			1,
			0,
			0,
			time.UTC,
		),
		Limit: productionQualityResultLimit,
	}

	metricSnapshot, err :=
		buildProductionQualitySnapshot(
			nil,
			window,
		)
	if err != nil {
		t.Fatalf(
			"expected empty server snapshot, got %v",
			err,
		)
	}

	if metricSnapshot.ObservedSamples != 0 ||
		metricSnapshot.ExpectedSamples != 6 ||
		!metricSnapshot.Time.IsZero() {
		t.Fatalf(
			"unexpected empty snapshot: %#v",
			metricSnapshot,
		)
	}
}
