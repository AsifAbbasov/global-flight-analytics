package airportresolver

import "errors"

var (
	ErrNoUsableAirports = errors.New(
		"airport candidate catalog contains no usable airports",
	)
	ErrCatalogRequired = errors.New(
		"airport candidate resolver catalog is required",
	)
	ErrInvalidMaximumDistance = errors.New(
		"airport candidate maximum distance must be finite and greater than zero",
	)
	ErrInvalidMaximumCandidates = errors.New(
		"airport candidate maximum candidates must be between one and one hundred",
	)
	ErrInvalidEndpointRole = errors.New(
		"airport candidate endpoint role must be origin or destination",
	)
	ErrInvalidPoint = errors.New(
		"airport candidate query point contains invalid coordinates",
	)
)
