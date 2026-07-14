package historicalairport

import "errors"

var (
	ErrSnapshotVersionInvalid = errors.New(
		"historical airport intelligence requires the current historical read snapshot version",
	)
	ErrMetricUnsupported = errors.New(
		"historical airport intelligence metric is unsupported",
	)
	ErrAirportICAOInvalid = errors.New(
		"historical airport intelligence requires a four-character uppercase alphanumeric ICAO code",
	)
)
