package metricquery

import "errors"

var ErrQueryContextRequired = errors.New(
	"metric query context is required",
)
