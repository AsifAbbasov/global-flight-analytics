package temporalbuilder

import "errors"

var (
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
