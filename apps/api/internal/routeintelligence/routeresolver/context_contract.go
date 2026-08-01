package routeresolver

import "errors"

var ErrRouteResolutionContextRequired = errors.New(
	"route resolution context is required",
)
