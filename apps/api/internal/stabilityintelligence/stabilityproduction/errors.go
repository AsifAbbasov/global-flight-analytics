package stabilityproduction

import "errors"

var (
	ErrProjectionReaderRequired = errors.New("Stability Intelligence projection reader is required")
	ErrServiceUnavailable       = errors.New("Stability Intelligence service is unavailable")
	ErrInvalidRequest           = errors.New("Stability Intelligence request is invalid")
	ErrTrajectoryNotFound       = errors.New("Stability Intelligence trajectory was not found")
	ErrProjectionLoadFailed     = errors.New("Stability Intelligence projection load failed")
)
