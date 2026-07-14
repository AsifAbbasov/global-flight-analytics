package extractorcomposition

import (
	"errors"
	"fmt"
)

var ErrAircraftLookupRequired = errors.New(
	"production feature extractor aircraft lookup is required",
)

type ComponentError struct {
	Component string
	Err       error
}

func (err *ComponentError) Error() string {
	if err == nil {
		return "construct production feature extractor component"
	}

	return fmt.Sprintf(
		"construct production feature extractor component %q: %v",
		err.Component,
		err.Err,
	)
}

func (err *ComponentError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Err
}
