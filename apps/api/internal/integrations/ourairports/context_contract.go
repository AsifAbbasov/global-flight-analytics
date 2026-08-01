package ourairports

import "errors"

var ErrLoadContextRequired = errors.New(
	"OurAirports load context is required",
)
