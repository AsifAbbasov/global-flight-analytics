package trackbuilder

import (
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/gapdetector"
)

const trackBuilderQualityComparisonTolerance = 1e-9

func TestNewBuilderRejectsInvalidGapDetectorConfig(
	t *testing.T,
) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "negative max time gap",
			config: Config{
				GapDetectorConfig: gapdetector.Config{
					MaxTimeGap: -time.Second,
				},
			},
		},
		{
			name: "negative max ground speed",
			config: Config{
				GapDetectorConfig: gapdetector.Config{
					MaxGroundSpeedMPS: -1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				builder, err := NewBuilder(
					test.config,
				)

				if err == nil {
					t.Fatal(
						"expected constructor error, got nil",
					)
				}

				if builder != nil {
					t.Fatal(
						"expected nil builder for invalid configuration",
					)
				}
			},
		)
	}
}

func TestBuildManyEmptyInput(t *testing.T) {
	builder := mustNewBuilder(
		t,
		Config{},
	)

	result := builder.BuildMany(nil)

	if len(result) != 0 {
		t.Fatalf(
			"expected zero trajectories, got %d",
			len(result),
		)
	}
}

func TestBuildManyContinuousTrajectoryPreservesQualityAcrossSorting(
	t *testing.T,
) {
	builder := mustNewBuilder(
		t,
		Config{},
	)

	observedAt := time.Date(
		2026,
		7,
		2,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	inputs := []InputState{
		{
			State: makeFlightState(
				"state-3",
				"ABC123",
				"AHY101",
				40.4300,
				49.8900,
				observedAt.Add(60*time.Second),
			),
			QualityScore: 1.0,
		},
		{
			State: makeFlightState(
				"state-1",
				"ABC123",
				"AHY101",
				40.4100,
				49.8700,
				observedAt,
			),
			QualityScore: 0.2,
		},
		{
			State: makeFlightState(
				"state-2",
				"ABC123",
				"AHY101",
				40.4200,
				49.8800,
				observedAt.Add(30*time.Second),
			),
			QualityScore: 0.6,
		},
	}

	result := builder.BuildMany(inputs)

	builtTrajectory := requireTrajectory(
		t,
		result,
		"ABC123",
	)

	if builtTrajectory.ICAO24 != "ABC123" {
		t.Fatalf(
			"expected ICAO24 ABC123, got %s",
			builtTrajectory.ICAO24,
		)
	}

	if builtTrajectory.PointCount != 3 {
		t.Fatalf(
			"expected 3 points, got %d",
			builtTrajectory.PointCount,
		)
	}

	if builtTrajectory.SegmentCount != 1 {
		t.Fatalf(
			"expected 1 segment, got %d",
			builtTrajectory.SegmentCount,
		)
	}

	if builtTrajectory.CoverageGapCount != 0 {
		t.Fatalf(
			"expected 0 coverage gaps, got %d",
			builtTrajectory.CoverageGapCount,
		)
	}

	if !builtTrajectory.StartTime.Equal(observedAt) {
		t.Fatalf(
			"expected start time %s, got %s",
			observedAt,
			builtTrajectory.StartTime,
		)
	}

	if len(builtTrajectory.Points) != 3 {
		t.Fatalf(
			"expected 3 point objects, got %d",
			len(builtTrajectory.Points),
		)
	}

	if builtTrajectory.Points[0].FlightStateID != "state-1" {
		t.Fatalf(
			"expected first sorted point state-1, got %s",
			builtTrajectory.Points[0].FlightStateID,
		)
	}

	if builtTrajectory.Points[1].FlightStateID != "state-2" {
		t.Fatalf(
			"expected second sorted point state-2, got %s",
			builtTrajectory.Points[1].FlightStateID,
		)
	}

	if builtTrajectory.Points[2].FlightStateID != "state-3" {
		t.Fatalf(
			"expected third sorted point state-3, got %s",
			builtTrajectory.Points[2].FlightStateID,
		)
	}

	if len(builtTrajectory.Segments) != 1 {
		t.Fatalf(
			"expected 1 segment object, got %d",
			len(builtTrajectory.Segments),
		)
	}

	assertTrackBuilderQualityClose(
		t,
		0.6,
		builtTrajectory.Segments[0].QualityScore,
	)

	assertTrackBuilderQualityClose(
		t,
		0.6,
		builtTrajectory.QualityScore,
	)
}

func TestBuildManyTrajectoryWithCoverageGapUsesExactSegmentMembership(
	t *testing.T,
) {
	builder := mustNewBuilder(
		t,
		Config{
			GapDetectorConfig: gapdetector.Config{
				MaxTimeGap: time.Minute,
			},
		},
	)

	observedAt := time.Date(
		2026,
		7,
		2,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	inputs := []InputState{
		{
			State: makeFlightState(
				"state-3",
				"ABC123",
				"AHY101",
				40.4300,
				49.8900,
				observedAt.Add(150*time.Second),
			),
			QualityScore: 0.9,
		},
		{
			State: makeFlightState(
				"state-1",
				"ABC123",
				"AHY101",
				40.4100,
				49.8700,
				observedAt,
			),
			QualityScore: 0.2,
		},
		{
			State: makeFlightState(
				"state-2",
				"ABC123",
				"AHY101",
				40.4200,
				49.8800,
				observedAt.Add(30*time.Second),
			),
			QualityScore: 0.6,
		},
	}

	result := builder.BuildMany(inputs)

	builtTrajectory := requireTrajectory(
		t,
		result,
		"ABC123",
	)

	if builtTrajectory.PointCount != 3 {
		t.Fatalf(
			"expected 3 points, got %d",
			builtTrajectory.PointCount,
		)
	}

	if builtTrajectory.SegmentCount != 2 {
		t.Fatalf(
			"expected 2 segments, got %d",
			builtTrajectory.SegmentCount,
		)
	}

	if builtTrajectory.CoverageGapCount != 1 {
		t.Fatalf(
			"expected 1 coverage gap, got %d",
			builtTrajectory.CoverageGapCount,
		)
	}

	if len(builtTrajectory.CoverageGaps) != 1 {
		t.Fatalf(
			"expected 1 coverage gap object, got %d",
			len(builtTrajectory.CoverageGaps),
		)
	}

	if builtTrajectory.CoverageGaps[0].Reason !=
		trajectory.CoverageGapReasonTimeGap {
		t.Fatalf(
			"expected time gap reason, got %s",
			builtTrajectory.CoverageGaps[0].Reason,
		)
	}

	if len(builtTrajectory.Segments) != 2 {
		t.Fatalf(
			"expected 2 segment objects, got %d",
			len(builtTrajectory.Segments),
		)
	}

	if builtTrajectory.Segments[0].PointCount != 2 {
		t.Fatalf(
			"expected first segment to contain 2 points, got %d",
			builtTrajectory.Segments[0].PointCount,
		)
	}

	if builtTrajectory.Segments[1].PointCount != 1 {
		t.Fatalf(
			"expected second segment to contain 1 point, got %d",
			builtTrajectory.Segments[1].PointCount,
		)
	}

	assertTrackBuilderQualityClose(
		t,
		0.4,
		builtTrajectory.Segments[0].QualityScore,
	)

	assertTrackBuilderQualityClose(
		t,
		0.9,
		builtTrajectory.Segments[1].QualityScore,
	)

	expectedTrajectoryQuality := (0.4*2 + 0.9) / 3.0

	assertTrackBuilderQualityClose(
		t,
		expectedTrajectoryQuality,
		builtTrajectory.QualityScore,
	)
}

func TestBuildManyGroupsByAircraftAndPreservesQuality(
	t *testing.T,
) {
	builder := mustNewBuilder(
		t,
		Config{},
	)

	observedAt := time.Date(
		2026,
		7,
		2,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	inputs := []InputState{
		{
			State: makeFlightState(
				"state-2",
				"ABC123",
				"AHY101",
				40.4200,
				49.8800,
				observedAt.Add(30*time.Second),
			),
			QualityScore: 0.6,
		},
		{
			State: makeFlightState(
				"state-3",
				"DEF456",
				"THY202",
				41.0000,
				49.0000,
				observedAt,
			),
			QualityScore: 0.9,
		},
		{
			State: makeFlightState(
				"state-1",
				"ABC123",
				"AHY101",
				40.4100,
				49.8700,
				observedAt,
			),
			QualityScore: 0.8,
		},
	}

	result := builder.BuildMany(inputs)

	if len(result) != 2 {
		t.Fatalf(
			"expected 2 trajectories, got %d",
			len(result),
		)
	}

	abcTrajectory := requireTrajectory(
		t,
		result,
		"ABC123",
	)

	defTrajectory := requireTrajectory(
		t,
		result,
		"DEF456",
	)

	if abcTrajectory.PointCount != 2 {
		t.Fatalf(
			"expected ABC123 to have 2 points, got %d",
			abcTrajectory.PointCount,
		)
	}

	if defTrajectory.PointCount != 1 {
		t.Fatalf(
			"expected DEF456 to have 1 point, got %d",
			defTrajectory.PointCount,
		)
	}

	if len(abcTrajectory.Segments) != 1 {
		t.Fatalf(
			"expected ABC123 to have 1 segment, got %d",
			len(abcTrajectory.Segments),
		)
	}

	if len(defTrajectory.Segments) != 1 {
		t.Fatalf(
			"expected DEF456 to have 1 segment, got %d",
			len(defTrajectory.Segments),
		)
	}

	for _, point := range abcTrajectory.Points {
		if point.ICAO24 != "ABC123" {
			t.Fatalf(
				"expected ABC123 trajectory to contain only ABC123 points, got %s",
				point.ICAO24,
			)
		}
	}

	for _, point := range defTrajectory.Points {
		if point.ICAO24 != "DEF456" {
			t.Fatalf(
				"expected DEF456 trajectory to contain only DEF456 points, got %s",
				point.ICAO24,
			)
		}
	}

	assertTrackBuilderQualityClose(
		t,
		0.7,
		abcTrajectory.Segments[0].QualityScore,
	)

	assertTrackBuilderQualityClose(
		t,
		0.7,
		abcTrajectory.QualityScore,
	)

	assertTrackBuilderQualityClose(
		t,
		0.9,
		defTrajectory.Segments[0].QualityScore,
	)

	assertTrackBuilderQualityClose(
		t,
		0.9,
		defTrajectory.QualityScore,
	)
}

func mustNewBuilder(
	t *testing.T,
	config Config,
) *Builder {
	t.Helper()

	builder, err := NewBuilder(
		config,
	)
	if err != nil {
		t.Fatalf(
			"create track builder: %v",
			err,
		)
	}

	return builder
}

func requireTrajectory(
	t *testing.T,
	trajectories map[string]trajectory.FlightTrajectory,
	icao24 string,
) trajectory.FlightTrajectory {
	t.Helper()

	builtTrajectory, exists := trajectories[icao24]

	if !exists {
		t.Fatalf(
			"expected trajectory for %s",
			icao24,
		)
	}

	return builtTrajectory
}

func assertTrackBuilderQualityClose(
	t *testing.T,
	expected float64,
	actual float64,
) {
	t.Helper()

	difference := math.Abs(
		expected - actual,
	)

	if difference > trackBuilderQualityComparisonTolerance {
		t.Fatalf(
			"expected quality %.12f, got %.12f",
			expected,
			actual,
		)
	}
}

func makeFlightState(
	id string,
	icao24 string,
	callsign string,
	latitude float64,
	longitude float64,
	observedAt time.Time,
) flightstate.FlightState {
	return flightstate.FlightState{
		ID:                  id,
		FlightID:            "flight-" + icao24,
		AircraftID:          "aircraft-" + icao24,
		ICAO24:              icao24,
		Callsign:            callsign,
		Latitude:            latitude,
		Longitude:           longitude,
		BarometricAltitudeM: 10000,
		GeometricAltitudeM:  10000,
		VelocityMPS:         230,
		HeadingDegrees:      90,
		VerticalRateMPS:     0,
		OnGround:            false,
		OriginCountry:       "Azerbaijan",
		ObservedAt:          observedAt,
		SourceName:          "test",
	}
}
