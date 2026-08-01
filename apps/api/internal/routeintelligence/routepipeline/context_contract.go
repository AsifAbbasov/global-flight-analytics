package routepipeline

import "errors"

var ErrRoutePipelineContextRequired = errors.New(
	"route pipeline context is required",
)
