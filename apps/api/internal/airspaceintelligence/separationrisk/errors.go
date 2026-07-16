package separationrisk

import "errors"

var (
	ErrInvalidPolicy  = errors.New("separation risk policy is invalid")
	ErrInvalidRequest = errors.New("separation risk request is invalid")
	ErrInvalidResult  = errors.New("separation risk result is invalid")
)
