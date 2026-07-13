package registry

import "errors"

var (
	ErrMetricAlreadyRegistered = errors.New("metric already registered")
	ErrMetricNotFound          = errors.New("metric not found")
)
