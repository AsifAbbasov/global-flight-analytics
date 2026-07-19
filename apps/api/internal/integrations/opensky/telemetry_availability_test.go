package opensky

import (
	"math"
	"testing"
	"time"
)

func TestMapStateVectorPreservesMissingKinematicTelemetry(
	t *testing.T,
) {
	snapshot := time.Date(
		2026,
		time.July,
		20,
		12,
		0,
		10,
		0,
		time.UTC,
	)
	positionTime := snapshot.Add(
		-5 * time.Second,
	)
	latitude := 40.4093
	longitude := 49.8671

	mapped, usable, err := MapStateVector(
		StateVector{
			SnapshotTime: snapshot,
			ICAO24:       "abc123",
			TimePosition: &positionTime,
			LastContact: snapshot.Add(
				-time.Second,
			),
			Latitude:  &latitude,
			Longitude: &longitude,
		},
	)
	if err != nil {
		t.Fatalf(
			"map state vector: %v",
			err,
		)
	}
	if !usable {
		t.Fatal(
			"fresh position must remain usable even when kinematics are unavailable",
		)
	}
	if !mapped.TelemetryAvailabilityKnown {
		t.Fatal(
			"OpenSky mapping did not establish explicit availability",
		)
	}
	if mapped.HasVelocity() ||
		mapped.HasHeading() ||
		mapped.HasVerticalRate() {
		t.Fatalf(
			"missing OpenSky kinematics became available: %#v",
			mapped,
		)
	}
	if !mapped.HasOnGroundState() {
		t.Fatal(
			"OpenSky on_ground field must remain explicitly available",
		)
	}
}

func TestMapStateVectorPreservesRealZeroKinematicTelemetry(
	t *testing.T,
) {
	snapshot := time.Date(
		2026,
		time.July,
		20,
		12,
		0,
		10,
		0,
		time.UTC,
	)
	positionTime := snapshot.Add(
		-5 * time.Second,
	)
	latitude := 0.0
	longitude := 0.0
	velocity := 0.0
	heading := 0.0
	verticalRate := 0.0

	mapped, usable, err := MapStateVector(
		StateVector{
			SnapshotTime:    snapshot,
			ICAO24:          "abc123",
			TimePosition:    &positionTime,
			LastContact:     snapshot.Add(-time.Second),
			Latitude:        &latitude,
			Longitude:       &longitude,
			VelocityMPS:     &velocity,
			TrueTrack:       &heading,
			VerticalRateMPS: &verticalRate,
		},
	)
	if err != nil {
		t.Fatalf(
			"map state vector: %v",
			err,
		)
	}
	if !usable {
		t.Fatal(
			"valid zero-valued telemetry was rejected",
		)
	}
	if !mapped.HasVelocity() ||
		!mapped.HasHeading() ||
		!mapped.HasVerticalRate() {
		t.Fatalf(
			"real zero kinematics were marked unavailable: %#v",
			mapped,
		)
	}
	if mapped.VelocityMPS != 0 ||
		mapped.HeadingDegrees != 0 ||
		mapped.VerticalRateMPS != 0 {
		t.Fatalf(
			"real zero kinematics changed value: %#v",
			mapped,
		)
	}
}

func TestOptionalFiniteFloat64RejectsNonFiniteValues(
	t *testing.T,
) {
	values := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}
	for _, value := range values {
		mapped, available :=
			optionalFiniteFloat64(
				&value,
			)
		if available || mapped != 0 {
			t.Fatalf(
				"non-finite value became available: value=%v mapped=%v available=%t",
				value,
				mapped,
				available,
			)
		}
	}
}
