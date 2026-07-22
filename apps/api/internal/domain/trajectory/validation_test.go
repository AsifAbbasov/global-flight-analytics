package trajectory

import (
	"errors"
	"testing"
	"time"
)

func TestFlightTrajectoryValidateRejectsCountAndDurationConflicts(t *testing.T) {
	start := time.Now().UTC()
	value := FlightTrajectory{StartTime: start, EndTime: start.Add(time.Minute), DurationSeconds: 60, PointCount: 2, Points: []TrackPoint4D{{Latitude: 40, Longitude: 49, ObservedAt: start}}, QualityScore: 1}
	if err := value.Validate(); !errors.Is(err, ErrTrajectoryCountInvalid) {
		t.Fatalf("count error = %v", err)
	}
	value.PointCount = 1
	value.DurationSeconds = 59
	if err := value.Validate(); !errors.Is(err, ErrTrajectoryDurationInvalid) {
		t.Fatalf("duration error = %v", err)
	}
}
