package forecaststability

import "errors"

var (
	ErrInvalidVersionPolicy      = errors.New("forecast versioning policy is invalid")
	ErrInvalidStabilityPolicy    = errors.New("decision stability policy is invalid")
	ErrInvalidVersionRequest     = errors.New("forecast version registration request is invalid")
	ErrInvalidStabilityRequest   = errors.New("decision stability evaluation request is invalid")
	ErrInvalidVersionRecord      = errors.New("forecast version record is invalid")
	ErrInvalidRegistrationResult = errors.New("forecast version registration result is invalid")
	ErrInvalidStabilityResult    = errors.New("decision stability result is invalid")
)
