package historicalwindow

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestFloorAndCeilBoundary(
	t *testing.T,
) {
	value := time.Date(
		2026,
		time.July,
		1,
		12,
		34,
		56,
		789,
		time.FixedZone("test", 4*60*60),
	)

	tests := []struct {
		name        string
		granularity historicalcontract.Granularity
		wantFloor   time.Time
		wantCeil    time.Time
	}{
		{
			name: "hour",
			granularity: historicalcontract.
				GranularityHour,
			wantFloor: time.Date(
				2026,
				time.July,
				1,
				8,
				0,
				0,
				0,
				time.UTC,
			),
			wantCeil: time.Date(
				2026,
				time.July,
				1,
				9,
				0,
				0,
				0,
				time.UTC,
			),
		},
		{
			name: "day",
			granularity: historicalcontract.
				GranularityDay,
			wantFloor: time.Date(
				2026,
				time.July,
				1,
				0,
				0,
				0,
				0,
				time.UTC,
			),
			wantCeil: time.Date(
				2026,
				time.July,
				2,
				0,
				0,
				0,
				0,
				time.UTC,
			),
		},
		{
			name: "week",
			granularity: historicalcontract.
				GranularityWeek,
			wantFloor: time.Date(
				2026,
				time.June,
				29,
				0,
				0,
				0,
				0,
				time.UTC,
			),
			wantCeil: time.Date(
				2026,
				time.July,
				6,
				0,
				0,
				0,
				0,
				time.UTC,
			),
		},
		{
			name: "custom",
			granularity: historicalcontract.
				GranularityCustom,
			wantFloor: value.UTC(),
			wantCeil:  value.UTC(),
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				floor, err := FloorBoundary(
					value,
					test.granularity,
				)
				if err != nil {
					t.Fatalf(
						"FloorBoundary() error = %v",
						err,
					)
				}
				ceil, err := CeilBoundary(
					value,
					test.granularity,
				)
				if err != nil {
					t.Fatalf(
						"CeilBoundary() error = %v",
						err,
					)
				}

				if !floor.Equal(
					test.wantFloor,
				) ||
					floor.Location() !=
						time.UTC {
					t.Fatalf(
						"floor = %s, want %s",
						floor,
						test.wantFloor,
					)
				}
				if !ceil.Equal(
					test.wantCeil,
				) ||
					ceil.Location() !=
						time.UTC {
					t.Fatalf(
						"ceil = %s, want %s",
						ceil,
						test.wantCeil,
					)
				}
			},
		)
	}
}

func TestCeilBoundaryReturnsAlignedBoundary(
	t *testing.T,
) {
	value := time.Date(
		2026,
		time.July,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	got, err := CeilBoundary(
		value,
		historicalcontract.GranularityHour,
	)
	if err != nil {
		t.Fatalf(
			"CeilBoundary() error = %v",
			err,
		)
	}
	if !got.Equal(value) {
		t.Fatalf(
			"ceil = %s, want %s",
			got,
			value,
		)
	}
}

func TestNextBoundary(
	t *testing.T,
) {
	monday := time.Date(
		2026,
		time.June,
		29,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	tests := []struct {
		granularity historicalcontract.Granularity
		want        time.Time
	}{
		{
			granularity: historicalcontract.
				GranularityHour,
			want: monday.Add(time.Hour),
		},
		{
			granularity: historicalcontract.
				GranularityDay,
			want: monday.AddDate(0, 0, 1),
		},
		{
			granularity: historicalcontract.
				GranularityWeek,
			want: monday.AddDate(0, 0, 7),
		},
	}

	for _, test := range tests {
		got, err := NextBoundary(
			monday,
			test.granularity,
		)
		if err != nil {
			t.Fatalf(
				"NextBoundary() error = %v",
				err,
			)
		}
		if !got.Equal(test.want) {
			t.Fatalf(
				"next = %s, want %s",
				got,
				test.want,
			)
		}
	}
}

func TestBoundaryFunctionsRejectUnsupportedGranularity(
	t *testing.T,
) {
	value := time.Now().UTC()

	for _, call := range []func() error{
		func() error {
			_, err := FloorBoundary(
				value,
				"minute",
			)
			return err
		},
		func() error {
			_, err := CeilBoundary(
				value,
				"minute",
			)
			return err
		},
		func() error {
			_, err := NextBoundary(
				value,
				historicalcontract.
					GranularityCustom,
			)
			return err
		},
	} {
		if err := call(); !errors.Is(
			err,
			ErrUnsupportedGranularity,
		) {
			t.Fatalf(
				"error = %v, want %v",
				err,
				ErrUnsupportedGranularity,
			)
		}
	}
}
