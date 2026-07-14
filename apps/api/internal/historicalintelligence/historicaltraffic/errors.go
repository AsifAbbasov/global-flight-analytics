package historicaltraffic

import "errors"

var (
	ErrSnapshotVersionInvalid = errors.New(
		"historical traffic requires the current historical read snapshot version",
	)
	ErrMetricUnsupported = errors.New(
		"historical traffic metric is unsupported",
	)
)
