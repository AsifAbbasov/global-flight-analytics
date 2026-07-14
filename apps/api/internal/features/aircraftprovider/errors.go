package aircraftprovider

import (
	"errors"
	"fmt"
)

var (
	ErrLookupRequired = errors.New(
		"aircraft feature lookup is required",
	)
	ErrInvalidPositiveCacheTTL = errors.New(
		"aircraft feature positive cache ttl must be greater than zero",
	)
	ErrInvalidNegativeCacheTTL = errors.New(
		"aircraft feature negative cache ttl must be greater than zero",
	)
	ErrInvalidICAO24 = errors.New(
		"aircraft feature reference has an invalid icao24",
	)
	ErrAircraftIdentityMismatch = errors.New(
		"aircraft feature lookup returned a different icao24",
	)
)

type LookupError struct {
	ICAO24 string
	Err    error
}

func (err *LookupError) Error() string {
	if err == nil {
		return "aircraft feature lookup failed"
	}

	return fmt.Sprintf(
		"lookup aircraft features for %s: %v",
		err.ICAO24,
		err.Err,
	)
}

func (err *LookupError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Err
}
