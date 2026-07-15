package projectionread

import "errors"

var (
	ErrDataSourceRequired = errors.New(
		"Projection Intelligence data source is required",
	)
	ErrComposerRequired = errors.New(
		"Projection Intelligence production composer is required",
	)
	ErrServiceUnavailable = errors.New(
		"Projection Intelligence read service is unavailable",
	)
	ErrInvalidRequest = errors.New(
		"Projection Intelligence read request is invalid",
	)
	ErrTrajectoryNotFound = errors.New(
		"Projection Intelligence trajectory was not found",
	)
	ErrRouteNotFound = errors.New(
		"Projection Intelligence route result was not found",
	)
	ErrRouteHistoryNotFound = errors.New(
		"Projection Intelligence route history was not found",
	)
	ErrTrajectoryPointLimitExceeded = errors.New(
		"Projection Intelligence trajectory point limit was exceeded",
	)
)
