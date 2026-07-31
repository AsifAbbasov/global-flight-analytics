package projectionarrival

import (
	"math"
	"time"
)

func durationCeilSeconds(
	seconds float64,
) (time.Duration, bool) {
	if !nonNegativeFinite(seconds) {
		return 0, false
	}

	nanoseconds := seconds * float64(time.Second)
	if !nonNegativeFinite(nanoseconds) ||
		nanoseconds >= float64(math.MaxInt64) {
		return 0, false
	}

	return time.Duration(math.Ceil(nanoseconds)), true
}

func durationCeilFraction(
	duration time.Duration,
	fraction float64,
) (time.Duration, bool) {
	if duration < 0 ||
		!nonNegativeFinite(fraction) ||
		fraction > 1 {
		return 0, false
	}

	return durationCeilSeconds(
		fraction * duration.Seconds(),
	)
}
