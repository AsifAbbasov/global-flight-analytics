package routestore

import (
	"errors"
	"testing"
	"time"
)

func TestScanRecordRejectsRouteTimestampMirrorDrift(t *testing.T) {
	t.Parallel()

	result := validRouteResult()
	storedAt := result.GeneratedAt.Add(time.Second)

	tests := []struct {
		name  string
		index int
		field string
	}{
		{name: "as of time", index: 3, field: "as_of_time"},
		{name: "stored at", index: 10, field: "stored_at"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			row := rowForResult(t, result, storedAt)
			mirror, ok := row.values[test.index].(time.Time)
			if !ok {
				t.Fatalf("fixture index %d is %T", test.index, row.values[test.index])
			}
			row.values[test.index] = mirror.Add(time.Microsecond)

			_, err := scanRecord(row)
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

func TestValidateRouteTimestampMirrorAllowsPostgresPrecisionLoss(t *testing.T) {
	t.Parallel()

	exact := time.Date(2026, 7, 21, 1, 2, 3, 123456789, time.UTC)
	mirror := exact.Add(211 * time.Nanosecond)
	if err := validateTimestampMirror("as_of_time", mirror, exact); err != nil {
		t.Fatalf("validate mirror: %v", err)
	}
}
