package flightphase

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestDetectRecognizesBasicFlightProfile(
	t *testing.T,
) {
	base := flightPhaseTestTime()
	points := []trajectory.TrackPoint4D{
		phasePoint(
			"ground-start",
			base,
			0,
			10,
			0,
			true,
		),
		phasePoint(
			"takeoff",
			base.Add(10*time.Second),
			400,
			80,
			4,
			false,
		),
		phasePoint(
			"climb",
			base.Add(20*time.Second),
			2500,
			180,
			5,
			false,
		),
		phasePoint(
			"cruise",
			base.Add(30*time.Second),
			9000,
			230,
			0.2,
			false,
		),
		phasePoint(
			"descent",
			base.Add(40*time.Second),
			3000,
			190,
			-4,
			false,
		),
		phasePoint(
			"landing",
			base.Add(50*time.Second),
			600,
			90,
			-3,
			false,
		),
		phasePoint(
			"ground-end",
			base.Add(60*time.Second),
			0,
			8,
			0,
			true,
		),
	}

	result, err := NewDefault().Detect(
		trajectory.FlightTrajectory{
			ID:     "trajectory-1",
			Points: points,
		},
	)
	if err != nil {
		t.Fatalf("detect flight phases: %v", err)
	}

	expected := []Phase{
		PhaseGround,
		PhaseTakeoff,
		PhaseClimb,
		PhaseCruise,
		PhaseDescent,
		PhaseLanding,
		PhaseGround,
	}
	if len(result.Points) != len(expected) {
		t.Fatalf(
			"expected %d classified points, got %d",
			len(expected),
			len(result.Points),
		)
	}

	for index, phase := range expected {
		if result.Points[index].Phase != phase {
			t.Fatalf(
				"expected phase %s at index %d, got %s",
				phase,
				index,
				result.Points[index].Phase,
			)
		}
	}

	if result.CurrentPhase != PhaseGround {
		t.Fatalf(
			"expected current ground phase, got %s",
			result.CurrentPhase,
		)
	}
	if len(result.Segments) != len(expected) {
		t.Fatalf(
			"expected %d segments, got %d",
			len(expected),
			len(result.Segments),
		)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("validate result: %v", err)
	}
}

func TestDetectReordersPointsAndExcludesMissingTime(
	t *testing.T,
) {
	base := flightPhaseTestTime()
	result, err := NewDefault().Detect(
		trajectory.FlightTrajectory{
			Points: []trajectory.TrackPoint4D{
				phasePoint(
					"later",
					base.Add(time.Minute),
					2000,
					180,
					2,
					false,
				),
				{
					ID: "missing-time",
				},
				phasePoint(
					"earlier",
					base,
					0,
					10,
					0,
					true,
				),
			},
		},
	)
	if err != nil {
		t.Fatalf("detect reordered phases: %v", err)
	}

	if result.ExcludedPointCount != 1 ||
		result.ClassifiedPointCount != 2 {
		t.Fatalf(
			"expected one exclusion and two classifications, got %#v",
			result,
		)
	}
	if result.Points[0].PointID != "earlier" ||
		result.Points[1].PointID != "later" {
		t.Fatalf(
			"expected chronological order, got %#v",
			result.Points,
		)
	}
	if !containsPhaseLimitation(
		result.Limitations,
		LimitationCodeInputReordered,
	) {
		t.Fatal("expected input reorder limitation")
	}
	if !containsPhaseLimitation(
		result.Limitations,
		LimitationCodeZeroTimeExcluded,
	) {
		t.Fatal("expected zero-time exclusion limitation")
	}
}

func TestDetectUsesGeometricAltitudeFallback(
	t *testing.T,
) {
	base := flightPhaseTestTime()
	point := phasePoint(
		"geometric",
		base,
		2500,
		180,
		3,
		false,
	)
	point.BarometricAltitudeStatus =
		flightstate.AltitudeStatusUnavailable
	point.GeometricAltitudeM = 2500
	point.GeometricAltitudeStatus =
		flightstate.AltitudeStatusObserved

	result, err := NewDefault().Detect(
		trajectory.FlightTrajectory{
			Points: []trajectory.TrackPoint4D{
				point,
			},
		},
	)
	if err != nil {
		t.Fatalf("detect geometric fallback: %v", err)
	}

	if result.Points[0].Phase != PhaseClimb {
		t.Fatalf(
			"expected climb from geometric altitude, got %s",
			result.Points[0].Phase,
		)
	}
	if !containsPhaseLimitation(
		result.Limitations,
		LimitationCodeGeometricFallback,
	) {
		t.Fatal("expected geometric fallback limitation")
	}
}

func TestDetectReturnsUnknownWithoutRequiredSignals(
	t *testing.T,
) {
	base := flightPhaseTestTime()
	point := trajectory.TrackPoint4D{
		ID:                       "unknown",
		ObservedAt:               base,
		VelocityMPS:              math.NaN(),
		VerticalRateMPS:          math.NaN(),
		BarometricAltitudeStatus: flightstate.AltitudeStatusUnavailable,
		GeometricAltitudeStatus:  flightstate.AltitudeStatusUnavailable,
	}

	result, err := NewDefault().Detect(
		trajectory.FlightTrajectory{
			Points: []trajectory.TrackPoint4D{
				point,
			},
		},
	)
	if err != nil {
		t.Fatalf("detect unknown phase: %v", err)
	}

	if result.CurrentPhase != PhaseUnknown {
		t.Fatalf(
			"expected unknown phase, got %s",
			result.CurrentPhase,
		)
	}
	if !containsReason(
		result.Points[0].Reasons,
		ReasonAltitudeUnavailable,
	) ||
		!containsReason(
			result.Points[0].Reasons,
			ReasonVerticalRateUnavailable,
		) {
		t.Fatalf(
			"expected missing signal reasons, got %#v",
			result.Points[0].Reasons,
		)
	}
}

func TestDetectEmptyTrajectoryIsExplainable(
	t *testing.T,
) {
	result, err := NewDefault().Detect(
		trajectory.FlightTrajectory{},
	)
	if err != nil {
		t.Fatalf("detect empty trajectory: %v", err)
	}
	if result.CurrentPhase != PhaseUnknown ||
		result.ClassifiedPointCount != 0 {
		t.Fatalf(
			"expected empty unknown result, got %#v",
			result,
		)
	}
	if !containsPhaseLimitation(
		result.Limitations,
		LimitationCodeNoTrajectoryPoints,
	) {
		t.Fatal("expected no-points limitation")
	}
}

func TestConfigValidationRejectsInvalidThresholds(
	t *testing.T,
) {
	config := DefaultConfig()
	config.CruiseMinimumAltitudeM =
		config.TakeoffMaximumAltitudeM

	_, err := New(config)
	if !errors.Is(
		err,
		ErrAltitudeThresholdOrderInvalid,
	) {
		t.Fatalf(
			"expected altitude threshold order error, got %v",
			err,
		)
	}

	config = DefaultConfig()
	config.CruiseMaximumAbsoluteVerticalRateMPS =
		config.ClimbMinimumVerticalRateMPS
	_, err = New(config)
	if !errors.Is(
		err,
		ErrVerticalRateThresholdOrderInvalid,
	) {
		t.Fatalf(
			"expected vertical-rate threshold order error, got %v",
			err,
		)
	}
}

func TestResultCloneCopiesOwnedSlices(
	t *testing.T,
) {
	base := flightPhaseTestTime()
	result, err := NewDefault().Detect(
		trajectory.FlightTrajectory{
			Points: []trajectory.TrackPoint4D{
				phasePoint(
					"ground",
					base,
					0,
					10,
					0,
					true,
				),
			},
		},
	)
	if err != nil {
		t.Fatalf("detect phase: %v", err)
	}

	clone := result.Clone()
	clone.Points[0].Reasons[0] =
		ReasonPhaseThresholdsUnresolved
	clone.Segments[0].Reasons[0] =
		ReasonPhaseThresholdsUnresolved

	if result.Points[0].Reasons[0] !=
		ReasonOnGroundFlag {
		t.Fatal("expected point reasons to be copied")
	}
	if result.Segments[0].Reasons[0] !=
		ReasonOnGroundFlag {
		t.Fatal("expected segment reasons to be copied")
	}
}

func phasePoint(
	id string,
	observedAt time.Time,
	altitudeM float64,
	velocityMPS float64,
	verticalRateMPS float64,
	onGround bool,
) trajectory.TrackPoint4D {
	status := flightstate.AltitudeStatusObserved
	if onGround {
		status = flightstate.AltitudeStatusGround
	}

	return trajectory.TrackPoint4D{
		ID:                       id,
		ObservedAt:               observedAt,
		BarometricAltitudeM:      altitudeM,
		BarometricAltitudeStatus: status,
		VelocityMPS:              velocityMPS,
		VerticalRateMPS:          verticalRateMPS,
		OnGround:                 onGround,
	}
}

func flightPhaseTestTime() time.Time {
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

func containsPhaseLimitation(
	values []Notice,
	code string,
) bool {
	for _, value := range values {
		if value.Code == code {
			return true
		}
	}
	return false
}

func containsReason(
	values []ReasonCode,
	reason ReasonCode,
) bool {
	for _, value := range values {
		if value == reason {
			return true
		}
	}
	return false
}
