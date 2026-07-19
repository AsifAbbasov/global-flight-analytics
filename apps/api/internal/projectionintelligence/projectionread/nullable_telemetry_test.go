package projectionread

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestTrajectoryPointQueriesDoNotFabricateNullableTelemetry(
	t *testing.T,
) {
	for name, query := range map[string]string{
		"flight":   trajectoryPointsByFlightSQL,
		"aircraft": trajectoryPointsByAircraftSQL,
	} {
		upper := strings.ToUpper(query)

		for _, forbidden := range []string{
			"COALESCE(LATITUDE",
			"COALESCE(LONGITUDE",
			"COALESCE(VELOCITY_MPS",
			"COALESCE(HEADING_DEGREES",
			"COALESCE(VERTICAL_RATE_MPS",
			"COALESCE(ON_GROUND",
		} {
			if strings.Contains(upper, forbidden) {
				t.Fatalf(
					"%s query fabricates nullable telemetry with %q",
					name,
					forbidden,
				)
			}
		}

		for _, required := range []string{
			"LATITUDE IS NOT NULL",
			"LONGITUDE IS NOT NULL",
			"VELOCITY_MPS IS NOT NULL",
			"HEADING_DEGREES IS NOT NULL",
			"VERTICAL_RATE_MPS IS NOT NULL",
			"ON_GROUND IS NOT NULL",
		} {
			if !strings.Contains(upper, required) {
				t.Fatalf(
					"%s query is missing completeness boundary %q",
					name,
					required,
				)
			}
		}
	}
}

func TestScanTrackPointRejectsIncompleteRequiredTelemetry(
	t *testing.T,
) {
	base := projectionReadPointRow(
		"state-a",
		projectionReadTestAsOfTime(),
		40.40,
		49.80,
	)

	testCases := []struct {
		name  string
		index int
		value any
	}{
		{
			name:  "latitude",
			index: 5,
			value: pgtype.Float8{},
		},
		{
			name:  "longitude",
			index: 6,
			value: pgtype.Float8{},
		},
		{
			name:  "velocity",
			index: 11,
			value: pgtype.Float8{},
		},
		{
			name:  "heading",
			index: 12,
			value: pgtype.Float8{},
		},
		{
			name:  "vertical rate",
			index: 13,
			value: pgtype.Float8{},
		},
		{
			name:  "on ground",
			index: 14,
			value: pgtype.Bool{},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				values := append(
					[]any(nil),
					base...,
				)
				values[testCase.index] =
					testCase.value

				point, usable, err :=
					scanTrackPoint(
						scriptedRow{
							values: values,
						},
					)
				if err != nil {
					t.Fatalf(
						"scanTrackPoint() error = %v",
						err,
					)
				}
				if usable {
					t.Fatalf(
						"incomplete %s telemetry was accepted: %#v",
						testCase.name,
						point,
					)
				}
			},
		)
	}
}

func TestScanTrackPointPreservesLegitimateZeroValues(
	t *testing.T,
) {
	values := projectionReadPointRow(
		"state-zero",
		projectionReadTestAsOfTime(),
		0,
		0,
	)
	values[11] = pgtype.Float8{
		Float64: 0,
		Valid:   true,
	}
	values[12] = pgtype.Float8{
		Float64: 0,
		Valid:   true,
	}
	values[13] = pgtype.Float8{
		Float64: 0,
		Valid:   true,
	}
	values[14] = pgtype.Bool{
		Bool:  false,
		Valid: true,
	}

	point, usable, err := scanTrackPoint(
		scriptedRow{
			values: values,
		},
	)
	if err != nil {
		t.Fatalf(
			"scanTrackPoint() error = %v",
			err,
		)
	}
	if !usable {
		t.Fatal(
			"present zero telemetry was treated as missing",
		)
	}
	if point.Latitude != 0 ||
		point.Longitude != 0 ||
		point.VelocityMPS != 0 ||
		point.HeadingDegrees != 0 ||
		point.VerticalRateMPS != 0 ||
		point.OnGround {
		t.Fatalf(
			"zero telemetry changed during scan: %#v",
			point,
		)
	}
}

func TestHydrateTrajectoryOmitsIncompleteTelemetryRows(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	item := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	item.Points = nil
	item.PointCount = 2

	incomplete := projectionReadPointRow(
		"state-incomplete",
		item.StartTime,
		40.40,
		49.80,
	)
	incomplete[5] = pgtype.Float8{}

	completeTime := item.StartTime.Add(
		time.Second,
	)
	complete := projectionReadPointRow(
		"state-complete",
		completeTime,
		40.41,
		49.81,
	)

	client := &scriptedClient{
		rowsQueue: []*scriptedRows{
			{
				values: [][]any{
					incomplete,
					complete,
				},
			},
		},
	}
	repository := &trajectoryRepositoryStub{
		items: map[string]trajectory.FlightTrajectory{
			item.ID: item,
		},
		errs: map[string]error{},
	}
	source := newProjectionReadTestSource(
		t,
		client,
		repository,
	)

	result, err := source.LoadCurrentTrajectory(
		context.Background(),
		item.ID,
		asOfTime,
	)
	if err != nil {
		t.Fatalf(
			"LoadCurrentTrajectory() error = %v",
			err,
		)
	}
	if len(result.Points) != 1 ||
		result.PointCount != 1 ||
		result.Points[0].ID !=
			"state-complete" ||
		!result.StartTime.Equal(
			completeTime,
		) {
		t.Fatalf(
			"incomplete telemetry was not omitted safely: %#v",
			result,
		)
	}
}
