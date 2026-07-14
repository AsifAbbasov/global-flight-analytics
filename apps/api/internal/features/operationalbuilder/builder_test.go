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

func TestBuilderBuildsCompleteOperationalFeatures(t *testing.T) {
	startTime := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	item := trajectory.FlightTrajectory{
		PointCount: 3,
		Points: []trajectory.TrackPoint4D{
			{
				BarometricAltitudeM:      999,
				BarometricAltitudeStatus: flightstate.AltitudeStatusGround,
				VelocityMPS:              0,
				VerticalRateMPS:          0,
				HeadingDegrees:           350,
				OnGround:                 true,
				ObservedAt:               startTime,
			},
			{
				BarometricAltitudeM:      1000,
				BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
				VelocityMPS:              100,
				VerticalRateMPS:          -5,
				HeadingDegrees:           10,
				ObservedAt:               startTime.Add(time.Minute),
			},
			{
				BarometricAltitudeM:      2000,
				BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
				VelocityMPS:              200,
				VerticalRateMPS:          10,
				HeadingDegrees:           20,
				ObservedAt:               startTime.Add(2 * time.Minute),
			},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	want := flightfeatures.OperationalFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Status:               flightfeatures.AvailabilityStatusAvailable,
			AvailableFieldCount:  OperationalFeatureFieldCount,
			TotalFieldCount:      OperationalFeatureFieldCount,
			SupportingPointCount: 3,
		},
		MinimumAltitudeM:               0,
		MaximumAltitudeM:               2000,
		MeanAltitudeM:                  1000,
		AltitudeRangeM:                 2000,
		MeanVelocityMPS:                100,
		MaximumVelocityMPS:             200,
		MeanAbsoluteVerticalRateMPS:    5,
		MaximumAbsoluteVerticalRateMPS: 10,
		HeadingChangeDegrees:           30,
		GroundObservationShare:         1.0 / 3.0,
		AirborneObservationShare:       2.0 / 3.0,
	}

	if !reflect.DeepEqual(features, want) {
		t.Fatalf(
			"features = %#v, want %#v",
			features,
			want,
		)
	}
}

func TestBuilderReturnsPartialEvidenceForMissingSignals(
	t *testing.T,
) {
	item := trajectory.FlightTrajectory{
		PointCount: 2,
		Points: []trajectory.TrackPoint4D{
			{
				BarometricAltitudeStatus: flightstate.AltitudeStatusUnavailable,
				GeometricAltitudeStatus:  flightstate.AltitudeStatusUnavailable,
				VelocityMPS:              -1,
				VerticalRateMPS:          math.NaN(),
				HeadingDegrees:           math.NaN(),
				OnGround:                 true,
			},
			{
				BarometricAltitudeStatus: flightstate.AltitudeStatusUnknown,
				GeometricAltitudeStatus:  flightstate.AltitudeStatusUnknown,
				VelocityMPS:              math.NaN(),
				VerticalRateMPS:          math.Inf(1),
				HeadingDegrees:           math.Inf(-1),
			},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.Evidence.Status !=
		flightfeatures.AvailabilityStatusPartial ||
		features.Evidence.AvailableFieldCount != 2 ||
		features.Evidence.TotalFieldCount !=
			OperationalFeatureFieldCount ||
		features.Evidence.SupportingPointCount != 2 {
		t.Fatalf(
			"unexpected partial evidence: %#v",
			features.Evidence,
		)
	}
	if features.GroundObservationShare != 0.5 ||
		features.AirborneObservationShare != 0.5 {
		t.Fatalf(
			"unexpected observation shares: %#v",
			features,
		)
	}

	for _, code := range []string{
		"operational_altitude_unavailable",
		"operational_velocity_unavailable",
		"operational_invalid_velocity_observations",
		"operational_vertical_rate_unavailable",
		"operational_invalid_vertical_rate_observations",
		"operational_heading_unavailable",
		"operational_invalid_heading_observations",
	} {
		if !hasLimitation(
			features.Evidence.Limitations,
			code,
		) {
			t.Fatalf(
				"missing limitation %q in %#v",
				code,
				features.Evidence.Limitations,
			)
		}
	}
}

func TestBuilderReturnsUnavailableWithoutPoints(t *testing.T) {
	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.Evidence.Status !=
		flightfeatures.AvailabilityStatusUnavailable ||
		features.Evidence.AvailableFieldCount != 0 ||
		features.Evidence.TotalFieldCount !=
			OperationalFeatureFieldCount ||
		features.Evidence.SupportingPointCount != 0 {
		t.Fatalf(
			"unexpected unavailable evidence: %#v",
			features.Evidence,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"operational_point_evidence_unavailable",
	) {
		t.Fatalf(
			"missing unavailable limitation: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderUsesGeometricAltitudeFallback(t *testing.T) {
	item := trajectory.FlightTrajectory{
		PointCount: 2,
		Points: []trajectory.TrackPoint4D{
			{
				BarometricAltitudeM:      500,
				BarometricAltitudeStatus: flightstate.AltitudeStatusInvalid,
				GeometricAltitudeM:       600,
				GeometricAltitudeStatus:  flightstate.AltitudeStatusObserved,
			},
			{
				BarometricAltitudeStatus: flightstate.AltitudeStatusUnavailable,
				GeometricAltitudeM:       700,
				GeometricAltitudeStatus:  flightstate.AltitudeStatusObserved,
			},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.MinimumAltitudeM != 600 ||
		features.MaximumAltitudeM != 700 ||
		features.MeanAltitudeM != 650 ||
		features.AltitudeRangeM != 100 {
		t.Fatalf(
			"unexpected fallback altitude features: %#v",
			features,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"operational_geometric_altitude_fallback",
	) || !hasLimitation(
		features.Evidence.Limitations,
		"operational_invalid_altitude_observations",
	) {
		t.Fatalf(
			"missing altitude fallback limitations: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderNormalizesHeadingAndUsesShortestArc(
	t *testing.T,
) {
	item := trajectory.FlightTrajectory{
		PointCount: 3,
		Points: []trajectory.TrackPoint4D{
			{HeadingDegrees: -10},
			{HeadingDegrees: 370},
			{HeadingDegrees: 180},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.HeadingChangeDegrees != 190 {
		t.Fatalf(
			"heading change = %v, want 190",
			features.HeadingChangeDegrees,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"operational_heading_normalized",
	) {
		t.Fatalf(
			"missing heading normalization limitation: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderSupportsSingleHeading(t *testing.T) {
	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{
			Points: []trajectory.TrackPoint4D{
				{HeadingDegrees: 123},
			},
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.HeadingChangeDegrees != 0 {
		t.Fatalf(
			"heading change = %v, want 0",
			features.HeadingChangeDegrees,
		)
	}
}

func TestBuilderReportsNonMonotonicPointOrder(t *testing.T) {
	base := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	item := trajectory.FlightTrajectory{
		Points: []trajectory.TrackPoint4D{
			{
				ObservedAt:     base.Add(time.Minute),
				HeadingDegrees: 10,
			},
			{
				ObservedAt:     base,
				HeadingDegrees: 20,
			},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"operational_point_order_nonmonotonic",
	) {
		t.Fatalf(
			"missing point-order limitation: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderReportsPointCountMetadataMismatch(
	t *testing.T,
) {
	item := trajectory.FlightTrajectory{
		PointCount: 5,
		Points: []trajectory.TrackPoint4D{
			{},
			{},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"trajectory_point_count_metadata_mismatch",
	) {
		t.Fatalf(
			"missing point-count mismatch limitation: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderDoesNotMutateInput(t *testing.T) {
	item := trajectory.FlightTrajectory{
		PointCount: 2,
		Points: []trajectory.TrackPoint4D{
			{
				BarometricAltitudeM:      1000,
				BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
				VelocityMPS:              100,
				VerticalRateMPS:          -5,
				HeadingDegrees:           350,
			},
			{
				BarometricAltitudeM:      2000,
				BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
				VelocityMPS:              200,
				VerticalRateMPS:          10,
				HeadingDegrees:           10,
			},
		},
	}
	original := item
	original.Points = append(
		[]trajectory.TrackPoint4D(nil),
		item.Points...,
	)

	if _, err := New().Build(
		context.Background(),
		item,
	); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !reflect.DeepEqual(item, original) {
		t.Fatalf(
			"input mutated\ninput=%#v\noriginal=%#v",
			item,
			original,
		)
	}
}

func TestBuilderPreservesCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := New().Build(
		ctx,
		trajectory.FlightTrajectory{},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Build() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestCloneFeaturesDoesNotShareLimitations(t *testing.T) {
	features := flightfeatures.OperationalFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Limitations: []flightfeatures.FeatureLimitation{
				{Code: "original"},
			},
		},
	}

	cloned := cloneFeatures(features)
	cloned.Evidence.Limitations[0].Code = "changed"

	if features.Evidence.Limitations[0].Code !=
		"original" {
		t.Fatal("cloneFeatures() shared limitations")
	}
}

func TestOperationalBuilderContractConstantsRemainStable(
	t *testing.T,
) {
	if Version != "operational-feature-builder-v1" {
		t.Fatalf("Version = %q", Version)
	}
	if OperationalFeatureFieldCount != 11 {
		t.Fatalf(
			"OperationalFeatureFieldCount = %d",
			OperationalFeatureFieldCount,
		)
	}
}

func hasLimitation(
	limitations []flightfeatures.FeatureLimitation,
	code string,
) bool {
	for _, limitation := range limitations {
		if limitation.Code == code {
			return true
		}
	}

	return false
}
