package weathercontext

import "errors"

var (
	ErrSnapshotContextRequired = errors.New(
		"Weather Context snapshot context is required",
	)
	ErrProductionContextRequired = errors.New(
		"Weather Context production context is required",
	)
)
