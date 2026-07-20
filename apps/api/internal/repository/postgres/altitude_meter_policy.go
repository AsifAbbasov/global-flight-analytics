package postgres

import (
	"errors"
	"fmt"
	"math"
)

const (
	minimumPostgresIntegerAltitudeMeters = -1 << 31
	maximumPostgresIntegerAltitudeMeters = 1<<31 - 1
)

var (
	ErrAltitudeMetersNotFinite = errors.New(
		"altitude meters must be finite",
	)
	ErrAltitudeMetersOutsidePostgresIntegerRange = errors.New(
		"altitude meters outside PostgreSQL integer range",
	)
)

// altitudeMetersToPostgresInteger applies the only supported conversion from
// provider/domain altitude precision to the whole-meter PostgreSQL integer
// representation. math.Round rounds to the nearest integer and rounds exact
// half values away from zero.
func altitudeMetersToPostgresInteger(
	value float64,
) (int32, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf(
			"%w: %v",
			ErrAltitudeMetersNotFinite,
			value,
		)
	}

	rounded := math.Round(value)
	if rounded < float64(minimumPostgresIntegerAltitudeMeters) ||
		rounded > float64(maximumPostgresIntegerAltitudeMeters) {
		return 0, fmt.Errorf(
			"%w: source=%v rounded=%v minimum=%d maximum=%d",
			ErrAltitudeMetersOutsidePostgresIntegerRange,
			value,
			rounded,
			minimumPostgresIntegerAltitudeMeters,
			maximumPostgresIntegerAltitudeMeters,
		)
	}

	return int32(rounded), nil
}
