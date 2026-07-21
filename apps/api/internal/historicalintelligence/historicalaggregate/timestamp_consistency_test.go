package historicalaggregate

import (
	"errors"
	"testing"
	"time"
)

func TestScanRecordRejectsHistoricalTimestampMirrorDrift(t *testing.T) {
	t.Parallel()

	result := aggregateFixture(
		t,
		"a",
		aggregateTestTime().Add(-time.Hour),
		aggregateTestTime(),
	)

	tests := []struct {
		name  string
		index int
		field string
	}{
		{name: "window start", index: 3, field: "window_start"},
		{name: "window end", index: 5, field: "window_end"},
		{name: "as of time", index: 7, field: "as_of_time"},
		{name: "stored at", index: 9, field: "stored_at"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			row := aggregateRow(t, result)
			mirror, ok := row[test.index].(time.Time)
			if !ok {
				t.Fatalf("fixture index %d is %T", test.index, row[test.index])
			}
			row[test.index] = mirror.Add(time.Microsecond)

			_, err := scanRecord(fakeScanner{values: row})
			if !errors.Is(err, ErrCorruptResult) {
				t.Fatalf("error = %v", err)
			}
			var corrupt *CorruptResultError
			if !errors.As(err, &corrupt) || corrupt.Field != test.field {
				t.Fatalf("corrupt error = %#v, want field %s", corrupt, test.field)
			}
		})
	}
}

func TestValidateHistoricalTimestampMirrorAllowsPostgresPrecisionLoss(
	t *testing.T,
) {
	t.Parallel()

	exact := time.Date(2026, 7, 21, 1, 2, 3, 123456789, time.UTC)
	mirror := exact.Add(-211 * time.Nanosecond)
	if err := validateTimestampMirror("stored_at", mirror, exact); err != nil {
		t.Fatalf("validate mirror: %v", err)
	}
}
