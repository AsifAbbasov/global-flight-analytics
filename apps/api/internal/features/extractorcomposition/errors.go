package extractorcomposition

import (
	"errors"
	"fmt"
)

var (
	ErrAircraftLookupRequired = errors.New(
		"production feature extractor aircraft lookup is required when enrichment is enabled",
	)
	ErrAircraftEnrichmentModeRequired = errors.New(
		"explicit aircraft enrichment mode is required; start with a default config constructor",
	)
	ErrAircraftEnrichmentConfigurationAmbiguous = errors.New(
		"aircraft lookup must be absent when enrichment is disabled",
	)
	ErrAircraftNotFoundPolicyVersionRequired = errors.New(
		"aircraft not-found policy version is required",
	)
	ErrGeographicCellPrecisionRequired = errors.New(
		"explicit geographic cell precision is required; start with DefaultConfig",
	)
	ErrAircraftPositiveCacheDurationRequired = errors.New(
		"explicit positive aircraft cache duration is required; start with DefaultConfig",
	)
	ErrAircraftNegativeCacheDurationRequired = errors.New(
		"explicit negative aircraft cache duration is required; start with DefaultConfig",
	)
)

type ComponentError struct {
	Component Component
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
