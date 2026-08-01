package airportresolver

import "errors"

var ErrAirportResolutionContextRequired = errors.New(
	"airport resolution context is required",
)
