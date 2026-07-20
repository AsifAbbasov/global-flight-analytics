package featurestore

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestValidateTimestampMirrorAcceptsExpectedPostgresPrecisionLoss(
	t *testing.T,
) {
	exact := time.Date(
		2026,
		time.July,
		20,
		15,
		30,
		0,
		123456789,
		time.UTC,
	)

	tests := []time.Time{
		exact,
		exact.Round(time.Microsecond),
		exact.Add(999 * time.Nanosecond),
		exact.Add(-999 * time.Nanosecond),
		exact.In(time.FixedZone("AZT", 4*60*60)),
	}

	for _, mirror := range tests {
		if err := validateTimestampMirror(
			"as_of_time",
			mirror,
			exact,
		); err != nil {
			t.Fatalf(
				"validateTimestampMirror(%s) error = %v",
				mirror,
				err,
			)
		}
	}
}

func TestValidateTimestampMirrorRejectsOneMicrosecondOrMore(
	t *testing.T,
) {
	exact := time.Date(
		2026,
		time.July,
		20,
		15,
		30,
		0,
		123456789,
		time.UTC,
	)

	for _, mirror := range []time.Time{
		exact.Add(time.Microsecond),
		exact.Add(-time.Microsecond),
		exact.Add(time.Second),
		time.Time{},
	} {
		err := validateTimestampMirror(
			"stored_at",
			mirror,
			exact,
		)
		if !errors.Is(err, ErrCorruptSnapshot) {
			t.Fatalf(
				"validateTimestampMirror(%s) error = %v, want ErrCorruptSnapshot",
				mirror,
				err,
			)
		}

		var corruptErr *CorruptSnapshotError
		if !errors.As(err, &corruptErr) ||
			corruptErr.Field != "stored_at" {
			t.Fatalf(
				"corrupt error = %#v, want stored_at field",
				corruptErr,
			)
		}
	}
}

func TestScanRecordAcceptsConsistentTimestampMirrors(
	t *testing.T,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		20,
		16,
		0,
		0,
		123456789,
		time.UTC,
	)
	storedAt := time.Date(
		2026,
		time.July,
		20,
		16,
		5,
		0,
		987654321,
		time.UTC,
	)
	features := validPostgresFeatures(
		testTrajectoryID,
		asOfTime,
		"d",
	)
	record := expectedRecord(features, storedAt)

	loaded, err := scanRecord(
		timestampMirrorRow(
			t,
			record,
			asOfTime.Round(time.Microsecond),
			storedAt.Round(time.Microsecond),
		),
	)
	if err != nil {
		t.Fatalf("scanRecord() error = %v", err)
	}
	if !loaded.Key.AsOfTime.Equal(asOfTime) {
		t.Fatalf(
			"loaded as-of time = %s, want %s",
			loaded.Key.AsOfTime,
			asOfTime,
		)
	}
	if !loaded.StoredAt.Equal(storedAt) {
		t.Fatalf(
			"loaded stored time = %s, want %s",
			loaded.StoredAt,
			storedAt,
		)
	}
}

func TestScanRecordRejectsAsOfTimestampMirrorDrift(
	t *testing.T,
) {
	record := timestampConsistencyRecord(t)

	_, err := scanRecord(
		timestampMirrorRow(
			t,
			record,
			record.Key.AsOfTime.Add(2*time.Microsecond),
			record.StoredAt,
		),
	)
	assertCorruptTimestampField(t, err, "as_of_time")
}

func TestScanRecordRejectsStoredTimestampMirrorDrift(
	t *testing.T,
) {
	record := timestampConsistencyRecord(t)

	_, err := scanRecord(
		timestampMirrorRow(
			t,
			record,
			record.Key.AsOfTime,
			record.StoredAt.Add(2*time.Microsecond),
		),
	)
	assertCorruptTimestampField(t, err, "stored_at")
}

func timestampConsistencyRecord(t *testing.T) Record {
	t.Helper()

	asOfTime := time.Date(
		2026,
		time.July,
		20,
		17,
		0,
		0,
		123456789,
		time.UTC,
	)
	storedAt := time.Date(
		2026,
		time.July,
		20,
		17,
		5,
		0,
		987654321,
		time.UTC,
	)

	return expectedRecord(
		validPostgresFeatures(
			testTrajectoryID,
			asOfTime,
			"e",
		),
		storedAt,
	)
}

func timestampMirrorRow(
	t *testing.T,
	record Record,
	asOfTimeMirror time.Time,
	storedAtMirror time.Time,
) rowScanner {
	t.Helper()

	payload, err := json.Marshal(record.Features)
	if err != nil {
		t.Fatalf("marshal features: %v", err)
	}

	return valueRow{
		scan: func(destinations ...any) error {
			assignDatabaseRow(
				t,
				destinations,
				record.ID,
				record.Key.TrajectoryID,
				string(record.Key.SchemaVersion),
				asOfTimeMirror,
				record.Key.AsOfTime.UnixNano(),
				record.InputFingerprint,
				string(record.Features.Quality.Status),
				payload,
				storedAtMirror,
				record.StoredAt.UnixNano(),
			)

			return nil
		},
	}
}

func assertCorruptTimestampField(
	t *testing.T,
	err error,
	field string,
) {
	t.Helper()

	if !errors.Is(err, ErrCorruptSnapshot) {
		t.Fatalf(
			"error = %v, want ErrCorruptSnapshot",
			err,
		)
	}

	var corruptErr *CorruptSnapshotError
	if !errors.As(err, &corruptErr) ||
		corruptErr.Field != field {
		t.Fatalf(
			"corrupt error = %#v, want field %q",
			corruptErr,
			field,
		)
	}
}
