package passport

import "errors"

var (
	ErrInvalidIdentity             = errors.New("invalid airport identity")
	ErrInvalidCoordinates          = errors.New("invalid airport coordinates")
	ErrInvalidOperations           = errors.New("invalid airport operations")
	ErrInvalidDataQuality          = errors.New("invalid airport data quality")
	ErrInvalidTime                 = errors.New("invalid airport passport time")
	ErrInvalidServiceConfiguration = errors.New("invalid airport passport service configuration")
)
