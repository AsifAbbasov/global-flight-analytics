package deduplicator

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestRemoveExactDuplicatesKeepsKnownZeroAndGroundDistinct(
	t *testing.T,
) {
	observedAt := altitudeSemanticDedupTestTime()

	knownZero := altitudeSemanticDedupState(
		observedAt,
	)
	knownZero.BarometricAltitudeM = 0
	knownZero.BarometricAltitudeStatus = flightstate.AltitudeStatusObserved

	ground := knownZero
	ground.BarometricAltitudeStatus = flightstate.AltitudeStatusGround
	ground.OnGround = true

	result := RemoveExactDuplicates(
		[]flightstate.FlightState{
			knownZero,
			ground,
		},
	)

	if result.DuplicateCount != 0 {
		t.Fatalf(
			"expected known zero and ground to remain distinct, got %d duplicates",
			result.DuplicateCount,
		)
	}

	if len(result.UniqueStates) != 2 {
		t.Fatalf(
			"expected 2 unique states, got %d",
			len(result.UniqueStates),
		)
	}
}

func TestRemoveExactDuplicatesKeepsUnknownAndUnavailableDistinct(
	t *testing.T,
) {
	observedAt := altitudeSemanticDedupTestTime()

	unknown := altitudeSemanticDedupState(
		observedAt,
	)
	unknown.BarometricAltitudeM = 0
	unknown.BarometricAltitudeStatus = flightstate.AltitudeStatusUnknown

	unavailable := unknown
	unavailable.BarometricAltitudeStatus = flightstate.AltitudeStatusUnavailable

	result := RemoveExactDuplicates(
		[]flightstate.FlightState{
			unknown,
			unavailable,
		},
	)

	if result.DuplicateCount != 0 {
		t.Fatalf(
			"expected unknown and unavailable to remain distinct, got %d duplicates",
			result.DuplicateCount,
		)
	}
}

func TestRemoveExactDuplicatesKeepsGeometricStatusesDistinct(
	t *testing.T,
) {
	observedAt := altitudeSemanticDedupTestTime()

	observed := altitudeSemanticDedupState(
		observedAt,
	)
	observed.GeometricAltitudeM = 0
	observed.GeometricAltitudeStatus = flightstate.AltitudeStatusObserved

	unavailable := observed
	unavailable.GeometricAltitudeStatus = flightstate.AltitudeStatusUnavailable

	result := RemoveExactDuplicates(
		[]flightstate.FlightState{
			observed,
			unavailable,
		},
	)

	if result.DuplicateCount != 0 {
		t.Fatalf(
			"expected geometric altitude statuses to remain distinct, got %d duplicates",
			result.DuplicateCount,
		)
	}
}

func TestRemoveExactDuplicatesStillRemovesSameAltitudeSemantics(
	t *testing.T,
) {
	observedAt := altitudeSemanticDedupTestTime()

	first := altitudeSemanticDedupState(
		observedAt,
	)

	duplicate := first
	duplicate.ID = "state-2"

	result := RemoveExactDuplicates(
		[]flightstate.FlightState{
			first,
			duplicate,
		},
	)

	if result.DuplicateCount != 1 {
		t.Fatalf(
			"expected same altitude semantics to deduplicate, got %d duplicates",
			result.DuplicateCount,
		)
	}
}

func altitudeSemanticDedupTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		9,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}

func altitudeSemanticDedupState(
	observedAt time.Time,
) flightstate.FlightState {
	return flightstate.FlightState{
		ID:                       "state-1",
		ICAO24:                   "ABC123",
		Latitude:                 40.4093,
		Longitude:                49.8671,
		BarometricAltitudeM:      1000,
		BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
		GeometricAltitudeM:       1050,
		GeometricAltitudeStatus:  flightstate.AltitudeStatusObserved,
		VelocityMPS:              220,
		HeadingDegrees:           90,
		ObservedAt:               observedAt,
	}
}
