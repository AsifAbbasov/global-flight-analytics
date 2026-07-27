package flightfeatures

import "time"

// TemporalDurationRoundingPolicy names the deterministic conversion from a
// nanosecond-resolution observation window to the integer duration exposed by
// the version-one feature schema.
type TemporalDurationRoundingPolicy string

const (
	// TemporalDurationRoundingPolicyTruncateFractionalSeconds discards any
	// fractional second by using Go duration division, which truncates toward
	// zero. Valid feature windows are non-negative, so this is equivalent to
	// flooring the duration to complete seconds.
	TemporalDurationRoundingPolicyTruncateFractionalSeconds TemporalDurationRoundingPolicy = "truncate_fractional_seconds_toward_zero"

	CurrentTemporalDurationRoundingPolicy = TemporalDurationRoundingPolicyTruncateFractionalSeconds
)

// TemporalDurationSeconds converts one observation window to the schema's
// whole-second duration using CurrentTemporalDurationRoundingPolicy.
func TemporalDurationSeconds(startTime time.Time, endTime time.Time) int64 {
	return int64(endTime.Sub(startTime) / time.Second)
}
