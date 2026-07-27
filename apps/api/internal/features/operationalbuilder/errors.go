package operationalbuilder

import "errors"

var (
	ErrContextRequired             = errors.New("operational feature builder context is required")
	ErrTrajectoryStartTimeRequired = errors.New("operational trajectory start time is required")
	ErrTrajectoryEndTimeRequired   = errors.New("operational trajectory end time is required")
	ErrInvalidTrajectoryWindow     = errors.New("operational trajectory end time precedes start time")
)
