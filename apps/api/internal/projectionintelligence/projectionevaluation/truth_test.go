package projectionevaluation

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestTruthAtInterpolatesPositionAndAltitude(t *testing.T) {
	start := evaluationTestAsOfTime()
	points := []canonicalTruthPoint{
		{point: trajectory.TrackPoint4D{ID: "left", Latitude: 0, Longitude: 0, GeometricAltitudeM: 1000, GeometricAltitudeStatus: flightstate.AltitudeStatusObserved, ObservedAt: start}},
		{point: trajectory.TrackPoint4D{ID: "right", Latitude: 0, Longitude: 0.02, GeometricAltitudeM: 1200, GeometricAltitudeStatus: flightstate.AltitudeStatusObserved, ObservedAt: start.Add(2 * time.Minute)}},
	}
	actual, status := truthAt(points, start.Add(time.Minute), validEvaluationConfig())
	if status != truthMatchAvailable {
		t.Fatalf("truthAt() status = %q", status)
	}
	if actual.source != ActualPointSourceInterpolated {
		t.Fatalf("source = %q", actual.source)
	}
	if math.Abs(actual.longitude-0.01) > 1e-6 {
		t.Fatalf("longitude = %f", actual.longitude)
	}
	if actual.altitudeM == nil || math.Abs(*actual.altitudeM-1100) > 1e-9 {
		t.Fatalf("altitude = %#v", actual.altitudeM)
	}
}

func TestTruthAtRejectsInterpolationAcrossLargeGap(t *testing.T) {
	start := evaluationTestAsOfTime()
	points := []canonicalTruthPoint{
		{point: trajectory.TrackPoint4D{ID: "left", Latitude: 0, Longitude: 0, ObservedAt: start}},
		{point: trajectory.TrackPoint4D{ID: "right", Latitude: 0, Longitude: 0.05, ObservedAt: start.Add(5 * time.Minute)}},
	}
	_, status := truthAt(points, start.Add(time.Minute), validEvaluationConfig())
	if status != truthMatchGapExceeded {
		t.Fatalf("status = %q, want gap_exceeded", status)
	}
}

func TestTruthAtRejectsImplausibleMovement(t *testing.T) {
	start := evaluationTestAsOfTime()
	config := validEvaluationConfig()
	points := []canonicalTruthPoint{
		{point: trajectory.TrackPoint4D{ID: "left", Latitude: 0, Longitude: 0, ObservedAt: start}},
		{point: trajectory.TrackPoint4D{ID: "right", Latitude: 0, Longitude: 10, ObservedAt: start.Add(time.Minute)}},
	}
	_, status := truthAt(points, start.Add(30*time.Second), config)
	if status != truthMatchImplausibleMovement {
		t.Fatalf("status = %q, want implausible_movement", status)
	}
}

func TestNormalizeTruthPointsAppliesObservationAndAvailabilityCutoffs(t *testing.T) {
	asOfTime := evaluationTestAsOfTime()
	evaluatedAt := asOfTime.Add(2 * time.Minute)
	item := trajectory.FlightTrajectory{ID: "trajectory", Points: []trajectory.TrackPoint4D{
		{ID: "future-event", Latitude: 0, Longitude: 3, ObservedAt: evaluatedAt.Add(time.Minute)},
		{ID: "future-availability", Latitude: 0, Longitude: 2, ObservedAt: asOfTime.Add(2 * time.Minute)},
		{ID: "first", Latitude: 0, Longitude: 1, ObservedAt: asOfTime.Add(time.Minute)},
		{ID: "past", Latitude: 0, Longitude: 0, ObservedAt: asOfTime.Add(-time.Minute)},
	}}
	truth, err := normalizeTruthPoints(item, []TruthAvailability{
		{PointID: "future-availability", SourceName: "ingest", AvailableAt: evaluatedAt.Add(time.Minute)},
		{PointID: "first", SourceName: "ingest", AvailableAt: asOfTime.Add(time.Minute + time.Second)},
	}, asOfTime, evaluatedAt)
	if err != nil {
		t.Fatalf("normalizeTruthPoints() error = %v", err)
	}
	if len(truth.points) != 1 || truth.points[0].point.ID != "first" ||
		truth.excludedAfterObservationCutoff != 1 || truth.excludedAfterAvailabilityCutoff != 1 {
		t.Fatalf("unexpected truth normalization: %#v", truth)
	}
}

func TestNormalizeTruthPointsRejectsConflictingDuplicateTimestamps(t *testing.T) {
	asOfTime := evaluationTestAsOfTime()
	observedAt := asOfTime.Add(time.Minute)
	item := trajectory.FlightTrajectory{ID: "trajectory", Points: []trajectory.TrackPoint4D{
		{ID: "a", Latitude: 0, Longitude: 1, ObservedAt: observedAt, SourceName: "source"},
		{ID: "b", Latitude: 0, Longitude: 2, ObservedAt: observedAt, SourceName: "source"},
	}}
	_, err := normalizeTruthPoints(item, []TruthAvailability{
		{PointID: "a", SourceName: "ingest", AvailableAt: observedAt},
		{PointID: "b", SourceName: "ingest", AvailableAt: observedAt},
	}, asOfTime, observedAt.Add(time.Minute))
	if !errors.Is(err, ErrAmbiguousTruthTimestamp) {
		t.Fatalf("error = %v", err)
	}
}

func TestNormalizeTruthPointsIsOrderIndependentForIdenticalDuplicates(t *testing.T) {
	asOfTime := evaluationTestAsOfTime()
	observedAt := asOfTime.Add(time.Minute)
	first := trajectory.TrackPoint4D{ID: "a", Latitude: 0, Longitude: 1, ObservedAt: observedAt, SourceName: "source"}
	second := first
	second.ID = "b"
	availability := []TruthAvailability{
		{PointID: "a", SourceName: "ingest", AvailableAt: observedAt},
		{PointID: "b", SourceName: "ingest", AvailableAt: observedAt},
	}
	left, err := normalizeTruthPoints(trajectory.FlightTrajectory{ID: "trajectory", Points: []trajectory.TrackPoint4D{first, second}}, availability, asOfTime, observedAt.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	right, err := normalizeTruthPoints(trajectory.FlightTrajectory{ID: "trajectory", Points: []trajectory.TrackPoint4D{second, first}}, availability, asOfTime, observedAt.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(left.points) != 1 || len(right.points) != 1 || left.points[0].point.ID != "a" || right.points[0].point.ID != "a" {
		t.Fatalf("normalization depends on order: left=%#v right=%#v", left, right)
	}
}

func TestNormalizeTruthPointsRequiresAvailabilityEvidence(t *testing.T) {
	asOfTime := evaluationTestAsOfTime()
	item := trajectory.FlightTrajectory{ID: "trajectory", Points: []trajectory.TrackPoint4D{{ID: "point", Latitude: 0, Longitude: 1, ObservedAt: asOfTime.Add(time.Minute)}}}
	_, err := normalizeTruthPoints(item, nil, asOfTime, asOfTime.Add(2*time.Minute))
	if !errors.Is(err, ErrTruthAvailabilityEvidenceMissing) {
		t.Fatalf("error = %v", err)
	}
}

func TestGreatCircleDistanceAcrossDateline(t *testing.T) {
	distanceM := greatCircleDistanceM(0, 179.9, 0, -179.9)
	if math.Abs(distanceM-22239) > 100 {
		t.Fatalf("distance = %f", distanceM)
	}
}
