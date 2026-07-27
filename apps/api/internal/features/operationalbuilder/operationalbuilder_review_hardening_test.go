package operationalbuilder

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestBuilderRejectsNilContext(t *testing.T) {
	_, err := New().Build(nil, trajectory.FlightTrajectory{})
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf("Build() error = %v, want %v", err, ErrContextRequired)
	}
}

func TestBuilderFiltersWindowAndIsPermutationInvariant(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	points := []trajectory.TrackPoint4D{
		operationalPoint("middle", start.Add(time.Minute), 1000, 100, 10, 1, false),
		operationalPoint("first", start, 500, 50, 350, 0, true),
		operationalPoint("last", start.Add(2*time.Minute), 1500, 150, 20, -2, false),
		operationalPoint("before", start.Add(-time.Second), 9000, 900, 180, 9, false),
		operationalPoint("after", start.Add(3*time.Minute), 9000, 900, 180, 9, false),
		operationalPoint("missing-time", time.Time{}, 9000, 900, 180, 9, false),
	}
	first := trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    start.Add(2 * time.Minute),
		PointCount: len(points),
		Points:     append([]trajectory.TrackPoint4D(nil), points...),
	}
	second := first.Clone()
	second.Points[0], second.Points[2] = second.Points[2], second.Points[0]
	second.Points[1], second.Points[5] = second.Points[5], second.Points[1]

	left, err := New().Build(context.Background(), first)
	if err != nil {
		t.Fatalf("first Build() error = %v", err)
	}
	right, err := New().Build(context.Background(), second)
	if err != nil {
		t.Fatalf("second Build() error = %v", err)
	}
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("permutation changed features\nleft=%#v\nright=%#v", left, right)
	}
	if left.MeanVelocityMPS != 100 || left.HeadingChangeDegrees != 30 {
		t.Fatalf("unexpected chronological metrics: %#v", left)
	}
	if left.Evidence.SupportingPointCount != 3 {
		t.Fatalf("supporting count = %d, want 3", left.Evidence.SupportingPointCount)
	}
	for _, code := range []string{
		flightfeatures.OperationalLimitationPointTimestampMissing,
		flightfeatures.OperationalLimitationPointOutsideWindow,
	} {
		if !flightfeatures.HasLimitationCode(left.Evidence.Limitations, code) {
			t.Fatalf("missing limitation %q: %#v", code, left.Evidence.Limitations)
		}
	}
}

func TestBuilderPreservesUnavailableZeroTelemetry(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	point := trajectory.TrackPoint4D{
		ID:                         "nullable",
		ObservedAt:                 start,
		TelemetryAvailabilityKnown: true,
		VelocityMPS:                0,
		VelocityAvailable:          false,
		HeadingDegrees:             0,
		HeadingAvailable:           false,
		VerticalRateMPS:            0,
		VerticalRateAvailable:      false,
		OnGround:                   false,
		OnGroundAvailable:          false,
		BarometricAltitudeStatus:   flightstate.AltitudeStatusUnavailable,
		GeometricAltitudeStatus:    flightstate.AltitudeStatusUnavailable,
	}
	features, err := New().Build(context.Background(), trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    start,
		PointCount: 1,
		Points:     []trajectory.TrackPoint4D{point},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.Evidence.Status != flightfeatures.AvailabilityStatusUnavailable ||
		features.Evidence.AvailableFieldCount != 0 ||
		features.Evidence.SupportingPointCount != 0 {
		t.Fatalf("nullable zeros became evidence: %#v", features)
	}
}

func TestBuilderGroundSharesUseOnlyAvailableStates(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	points := make([]trajectory.TrackPoint4D, 3)
	for index := range points {
		points[index] = trajectory.TrackPoint4D{
			ID:                         string(rune('a' + index)),
			ObservedAt:                 start.Add(time.Duration(index) * time.Second),
			TelemetryAvailabilityKnown: true,
			BarometricAltitudeStatus:   flightstate.AltitudeStatusUnavailable,
			GeometricAltitudeStatus:    flightstate.AltitudeStatusUnavailable,
		}
	}
	points[0].OnGroundAvailable = true
	points[0].OnGround = true
	points[1].OnGroundAvailable = true
	points[1].OnGround = false

	features, err := New().Build(context.Background(), trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    start.Add(2 * time.Second),
		PointCount: 3,
		Points:     points,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.GroundObservationShare != 0.5 || features.AirborneObservationShare != 0.5 {
		t.Fatalf("shares used unavailable states: %#v", features)
	}
	if features.Evidence.SupportingPointCount != 2 {
		t.Fatalf("supporting count = %d, want 2", features.Evidence.SupportingPointCount)
	}
}

func TestBuilderExcludesInvalidHeadingAndDoesNotBridgeGap(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	points := []trajectory.TrackPoint4D{
		headingOnlyPoint("one", start, 350, true),
		headingOnlyPoint("invalid", start.Add(time.Second), -10, true),
		headingOnlyPoint("two", start.Add(2*time.Second), 10, true),
		headingOnlyPoint("unavailable", start.Add(3*time.Second), 0, false),
		headingOnlyPoint("three", start.Add(4*time.Second), 20, true),
	}
	features, err := New().Build(context.Background(), trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    start.Add(4 * time.Second),
		PointCount: len(points),
		Points:     points,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.HeadingChangeDegrees != 0 {
		t.Fatalf("heading gaps were bridged: %#v", features)
	}
	for _, code := range []string{
		flightfeatures.OperationalLimitationInvalidHeadingObservations,
		flightfeatures.OperationalLimitationHeadingMeasurementUnavailable,
		flightfeatures.OperationalLimitationHeadingSequenceGap,
	} {
		if !flightfeatures.HasLimitationCode(features.Evidence.Limitations, code) {
			t.Fatalf("missing limitation %q: %#v", code, features.Evidence.Limitations)
		}
	}
}

func TestBuilderUsesSingleAltitudeSource(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	points := []trajectory.TrackPoint4D{
		{
			ID:                         "barometric",
			ObservedAt:                 start,
			BarometricAltitudeM:        1000,
			BarometricAltitudeStatus:   flightstate.AltitudeStatusObserved,
			GeometricAltitudeM:         900,
			GeometricAltitudeStatus:    flightstate.AltitudeStatusObserved,
			TelemetryAvailabilityKnown: true,
		},
		{
			ID:                         "geometric-only",
			ObservedAt:                 start.Add(time.Second),
			BarometricAltitudeStatus:   flightstate.AltitudeStatusUnavailable,
			GeometricAltitudeM:         1100,
			GeometricAltitudeStatus:    flightstate.AltitudeStatusObserved,
			TelemetryAvailabilityKnown: true,
		},
	}
	features, err := New().Build(context.Background(), trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    start.Add(time.Second),
		PointCount: 2,
		Points:     points,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.MinimumAltitudeM != 1000 || features.MaximumAltitudeM != 1000 ||
		features.MeanAltitudeM != 1000 || features.AltitudeRangeM != 0 {
		t.Fatalf("altitude sources were mixed: %#v", features)
	}
	if !flightfeatures.HasLimitationCode(
		features.Evidence.Limitations,
		flightfeatures.OperationalLimitationMixedAltitudeSourceExcluded,
	) {
		t.Fatalf("missing source limitation: %#v", features.Evidence.Limitations)
	}
}

func TestBuilderRejectsConflictingGroundAltitude(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	features, err := New().Build(context.Background(), trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    start,
		PointCount: 1,
		Points: []trajectory.TrackPoint4D{{
			ObservedAt:                 start,
			BarometricAltitudeM:        999,
			BarometricAltitudeStatus:   flightstate.AltitudeStatusGround,
			OnGround:                   true,
			OnGroundAvailable:          true,
			TelemetryAvailabilityKnown: true,
		}},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.Evidence.AvailableFieldCount != 2 {
		t.Fatalf("conflicting ground altitude became evidence: %#v", features)
	}
	if !flightfeatures.HasLimitationCode(
		features.Evidence.Limitations,
		flightfeatures.OperationalLimitationInvalidAltitudeObservations,
	) {
		t.Fatalf("missing invalid-altitude limitation: %#v", features.Evidence.Limitations)
	}
}

func TestBuilderRejectsNonFiniteAggregate(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	points := []trajectory.TrackPoint4D{
		velocityOnlyPoint("one", start, math.MaxFloat64),
		velocityOnlyPoint("two", start.Add(time.Second), math.MaxFloat64),
	}
	features, err := New().Build(context.Background(), trajectory.FlightTrajectory{
		StartTime:  start,
		EndTime:    start.Add(time.Second),
		PointCount: 2,
		Points:     points,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if features.MeanVelocityMPS != 0 || features.MaximumVelocityMPS != 0 {
		t.Fatalf("non-finite aggregate was published: %#v", features)
	}
	if !flightfeatures.HasLimitationCode(
		features.Evidence.Limitations,
		flightfeatures.OperationalLimitationAggregateNonFinite,
	) {
		t.Fatalf("missing aggregate limitation: %#v", features.Evidence.Limitations)
	}
}

func TestBuilderObservesCancellationDuringPointScan(t *testing.T) {
	points := make([]trajectory.TrackPoint4D, contextCheckInterval*3)
	ctx := &operationalCountingContext{cancelAfter: 4}
	_, err := New().Build(ctx, trajectory.FlightTrajectory{Points: points})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Build() error = %v, want context.Canceled", err)
	}
}

func operationalPoint(
	id string,
	observedAt time.Time,
	altitude float64,
	velocity float64,
	heading float64,
	verticalRate float64,
	onGround bool,
) trajectory.TrackPoint4D {
	return trajectory.TrackPoint4D{
		ID:                         id,
		ObservedAt:                 observedAt,
		BarometricAltitudeM:        altitude,
		BarometricAltitudeStatus:   flightstate.AltitudeStatusObserved,
		VelocityMPS:                velocity,
		VelocityAvailable:          true,
		HeadingDegrees:             heading,
		HeadingAvailable:           true,
		VerticalRateMPS:            verticalRate,
		VerticalRateAvailable:      true,
		OnGround:                   onGround,
		OnGroundAvailable:          true,
		TelemetryAvailabilityKnown: true,
	}
}

func headingOnlyPoint(
	id string,
	observedAt time.Time,
	heading float64,
	available bool,
) trajectory.TrackPoint4D {
	return trajectory.TrackPoint4D{
		ID:                         id,
		ObservedAt:                 observedAt,
		BarometricAltitudeStatus:   flightstate.AltitudeStatusUnavailable,
		GeometricAltitudeStatus:    flightstate.AltitudeStatusUnavailable,
		HeadingDegrees:             heading,
		HeadingAvailable:           available,
		TelemetryAvailabilityKnown: true,
	}
}

func velocityOnlyPoint(
	id string,
	observedAt time.Time,
	velocity float64,
) trajectory.TrackPoint4D {
	return trajectory.TrackPoint4D{
		ID:                         id,
		ObservedAt:                 observedAt,
		BarometricAltitudeStatus:   flightstate.AltitudeStatusUnavailable,
		GeometricAltitudeStatus:    flightstate.AltitudeStatusUnavailable,
		VelocityMPS:                velocity,
		VelocityAvailable:          true,
		TelemetryAvailabilityKnown: true,
	}
}

type operationalCountingContext struct {
	calls       int
	cancelAfter int
}

func (ctx *operationalCountingContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (ctx *operationalCountingContext) Done() <-chan struct{}       { return nil }
func (ctx *operationalCountingContext) Value(any) any               { return nil }
func (ctx *operationalCountingContext) Err() error {
	ctx.calls++
	if ctx.calls >= ctx.cancelAfter {
		return context.Canceled
	}
	return nil
}
