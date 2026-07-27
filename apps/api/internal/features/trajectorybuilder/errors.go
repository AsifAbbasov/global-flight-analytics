package trajectorybuilder

import "errors"

var (
	ErrContextRequired             = errors.New("trajectory feature builder context is required")
	ErrTrajectoryStartTimeRequired = errors.New("trajectory start time is required when end time is present")
	ErrTrajectoryEndTimeRequired   = errors.New("trajectory end time is required when start time is present")
	ErrInvalidTrajectoryWindow     = errors.New("trajectory end time precedes start time")
)
