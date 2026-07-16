package airspaceregionanalytics

import "errors"

var (
	ErrInvalidPolicy  = errors.New("airspace region analytics policy is invalid")
	ErrInvalidRequest = errors.New("airspace region analytics request is invalid")
	ErrInvalidResult  = errors.New("airspace region analytics result is invalid")
)
