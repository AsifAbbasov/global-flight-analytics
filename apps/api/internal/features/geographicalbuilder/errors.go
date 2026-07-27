package geographicalbuilder

import "errors"

var (
	ErrContextRequired = errors.New(
		"geographical feature context is required",
	)
	ErrTrajectoryStartTimeRequired = errors.New(
		"geographical feature trajectory start time is required when a temporal window is partially specified",
	)
	ErrTrajectoryEndTimeRequired = errors.New(
		"geographical feature trajectory end time is required when a temporal window is partially specified",
	)
	ErrInvalidTrajectoryWindow = errors.New(
		"geographical feature trajectory end time is before start time",
	)
	ErrInvalidGeographicCellPrecision = errors.New(
		"geographic cell precision zero selects the default; effective precision must be between one and six",
	)
)
