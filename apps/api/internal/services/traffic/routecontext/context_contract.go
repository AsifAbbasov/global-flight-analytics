package routecontext

import "errors"

var ErrRouteContextRequired = errors.New(
	"route context is required",
)
