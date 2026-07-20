package featurestore

import "time"

const postgresTimestampMirrorTolerance = time.Microsecond

// validateTimestampMirror verifies that PostgreSQL's microsecond-precision
// timestamptz mirror still represents the exact instant stored in Unix
// nanoseconds. A sub-microsecond difference is expected precision loss;
// one microsecond or more is corruption.
func validateTimestampMirror(
	field string,
	mirror time.Time,
	exact time.Time,
) error {
	if mirror.IsZero() || exact.IsZero() {
		return &CorruptSnapshotError{Field: field}
	}

	delta := mirror.UTC().Sub(exact.UTC())
	if delta <= -postgresTimestampMirrorTolerance ||
		delta >= postgresTimestampMirrorTolerance {
		return &CorruptSnapshotError{Field: field}
	}

	return nil
}
