package historicalroute

import "errors"

var (
	ErrSnapshotVersionInvalid = errors.New(
		"historical route intelligence requires the current historical read snapshot version",
	)
	ErrMetricUnsupported = errors.New(
		"historical route intelligence metric is unsupported",
	)
	ErrRouteScopeIncomplete = errors.New(
		"historical route scope requires both origin and destination ICAO codes or neither",
	)
	ErrOriginICAOInvalid = errors.New(
		"historical route origin ICAO code must contain four uppercase alphanumeric characters",
	)
	ErrDestinationICAOInvalid = errors.New(
		"historical route destination ICAO code must contain four uppercase alphanumeric characters",
	)
)
