package temporalbuilder

import "errors"

var (
	ErrContextRequired = errors.New(
		"temporal feature builder context is required",
	)
	ErrTrajectoryStartTimeRequired = errors.New(
		"temporal feature trajectory start time is required",
	)
	ErrTrajectoryEndTimeRequired = errors.New(
		"temporal feature trajectory end time is required",
	)
	ErrInvalidTrajectoryWindow = errors.New(
		"temporal feature trajectory end time is before start time",
	)
)
