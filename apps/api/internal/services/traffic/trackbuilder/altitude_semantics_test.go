package trackbuilder

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestBuildManyPropagatesAltitudeStatusesToTrackPoints(
	t *testing.T,
) {
	builder := mustNewBuilder(
		t,
		Config{},
	)

	observedAt := altitudeSemanticTrackBuilderTestTime()

	state := flightstate.FlightState{
		ID:                       "state-1",
		ICAO24:                   "ABC123",
		Callsign:                 "AHY101",
		Latitude:                 40.4093,
		Longitude:                49.8671,
		BarometricAltitudeM:      0,
		BarometricAltitudeStatus: flightstate.AltitudeStatusGround,
		GeometricAltitudeM:       0,
		GeometricAltitudeStatus:  flightstate.AltitudeStatusObserved,
		VelocityMPS:              0,
		HeadingDegrees:           90,
		VerticalRateMPS:          0,
		OnGround:                 true,
		ObservedAt:               observedAt,
		SourceName:               "airplanes.live",
	}

	result := builder.BuildMany(
		[]InputState{
			{
				State:        state,
				QualityScore: 1,
			},
		},
	)

	builtTrajectory := requireTrajectory(
		t,
		result,
		"ABC123",
	)

	if len(builtTrajectory.Points) != 1 {
		t.Fatalf(
			"expected 1 trajectory point, got %d",
			len(builtTrajectory.Points),
		)
	}

	point := builtTrajectory.Points[0]

	if point.BarometricAltitudeStatus != flightstate.AltitudeStatusGround {
		t.Fatalf(
			"expected ground barometric altitude status, got %q",
			point.BarometricAltitudeStatus,
		)
	}

	if point.GeometricAltitudeStatus != flightstate.AltitudeStatusObserved {
		t.Fatalf(
			"expected observed geometric altitude status, got %q",
			point.GeometricAltitudeStatus,
		)
	}
}

func TestBuildManyCanonicalizesLegacyAltitudeStatuses(
	t *testing.T,
) {
	builder := mustNewBuilder(
		t,
		Config{},
	)

	observedAt := altitudeSemanticTrackBuilderTestTime()

	state := flightstate.FlightState{
		ID:                  "state-1",
		ICAO24:              "ABC123",
		Latitude:            40.4093,
		Longitude:           49.8671,
		BarometricAltitudeM: 1000,
		GeometricAltitudeM:  0,
		VelocityMPS:         220,
		HeadingDegrees:      90,
		ObservedAt:          observedAt,
		SourceName:          "legacy-source",
	}

	result := builder.BuildMany(
		[]InputState{
			{
				State:        state,
				QualityScore: 1,
			},
		},
	)

	point := requireTrajectory(
		t,
		result,
		"ABC123",
	).Points[0]

	if point.BarometricAltitudeStatus != flightstate.AltitudeStatusObserved {
		t.Fatalf(
			"expected non-zero legacy altitude to resolve to observed, got %q",
			point.BarometricAltitudeStatus,
		)
	}

	if point.GeometricAltitudeStatus != flightstate.AltitudeStatusUnavailable {
		t.Fatalf(
			"expected zero legacy altitude to resolve to unavailable, got %q",
			point.GeometricAltitudeStatus,
		)
	}
}

func altitudeSemanticTrackBuilderTestTime() time.Time {
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
