package routestore

import "time"

const postgresTimestampMirrorTolerance = time.Microsecond

func validateTimestampMirror(
	field string,
	mirror time.Time,
	exact time.Time,
) error {
	if mirror.IsZero() || exact.IsZero() {
		return &CorruptResultError{Field: field}
	}

	delta := mirror.UTC().Sub(exact.UTC())
	if delta <= -postgresTimestampMirrorTolerance ||
		delta >= postgresTimestampMirrorTolerance {
		return &CorruptResultError{Field: field}
	}

	return nil
}
