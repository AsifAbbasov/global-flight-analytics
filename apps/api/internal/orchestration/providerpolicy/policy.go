package providerpolicy

import (
	"errors"
	"fmt"
)

type Provider string

const (
	ProviderAirplanesLive Provider = "airplanes.live"
	ProviderOpenMeteo     Provider = "open_meteo"
	ProviderOurAirports   Provider = "ourairports"
	ProviderOpenSky       Provider = "opensky"
)

type Provenance string

const (
	ProvenanceSourceBacked     Provenance = "SOURCE-BACKED"
	ProvenanceProviderDirected Provenance = "PROVIDER-DIRECTED"
)

type BudgetMode string

const (
	BudgetModeFixedWindow       BudgetMode = "fixed-window"
	BudgetModeProviderReported  BudgetMode = "provider-reported"
	BudgetModePublicationDriven BudgetMode = "publication-driven"
)

type Window string

const (
	WindowSecond Window = "second"
	WindowMinute Window = "minute"
	WindowHour   Window = "hour"
	WindowDay    Window = "day"
	WindowMonth  Window = "month"
)

type RequestLimit struct {
	MaxRequests int
	Window      Window
	Provenance  Provenance
	Reference   string
}

type ProviderReportedBudget struct {
	RemainingHeader         string
	RetryAfterSecondsHeader string
	Provenance              Provenance
	Reference               string
}

type PublicationPolicy struct {
	Cadence    string
	Provenance Provenance
	Reference  string
}

type Policy struct {
	Provider               Provider
	BudgetMode             BudgetMode
	RequestLimits          []RequestLimit
	ProviderReportedBudget *ProviderReportedBudget
	PublicationPolicy      *PublicationPolicy
}

var ErrUnknownProvider = errors.New(
	"unknown provider access policy",
)

func Get(
	provider Provider,
) (Policy, error) {
	switch provider {
	case ProviderAirplanesLive:
		return airplanesLivePolicy(), nil
	case ProviderOpenMeteo:
		return openMeteoPolicy(), nil
	case ProviderOurAirports:
		return ourAirportsPolicy(), nil
	case ProviderOpenSky:
		return openSkyPolicy(), nil
	default:
		return Policy{}, fmt.Errorf(
			"%w: %s",
			ErrUnknownProvider,
			provider,
		)
	}
}

func All() []Policy {
	return []Policy{
		airplanesLivePolicy(),
		openMeteoPolicy(),
		ourAirportsPolicy(),
		openSkyPolicy(),
	}
}

func Validate(
	policy Policy,
) error {
	if policy.Provider == "" {
		return errors.New(
			"provider is required",
		)
	}

	switch policy.BudgetMode {
	case BudgetModeFixedWindow:
		return validateFixedWindowPolicy(policy)
	case BudgetModeProviderReported:
		return validateProviderReportedPolicy(policy)
	case BudgetModePublicationDriven:
		return validatePublicationDrivenPolicy(policy)
	default:
		return fmt.Errorf(
			"unsupported budget mode: %s",
			policy.BudgetMode,
		)
	}
}

func validateFixedWindowPolicy(
	policy Policy,
) error {
	if len(policy.RequestLimits) == 0 {
		return errors.New(
			"fixed-window policy requires request limits",
		)
	}

	for index, limit := range policy.RequestLimits {
		if limit.MaxRequests <= 0 {
			return fmt.Errorf(
				"request limit %d must be greater than zero",
				index,
			)
		}

		if !isSupportedWindow(limit.Window) {
			return fmt.Errorf(
				"request limit %d has unsupported window: %s",
				index,
				limit.Window,
			)
		}

		if limit.Provenance == "" {
			return fmt.Errorf(
				"request limit %d provenance is required",
				index,
			)
		}

		if limit.Reference == "" {
			return fmt.Errorf(
				"request limit %d reference is required",
				index,
			)
		}
	}

	return nil
}

func validateProviderReportedPolicy(
	policy Policy,
) error {
	budget := policy.ProviderReportedBudget
	if budget == nil {
		return errors.New(
			"provider-reported policy requires provider budget metadata",
		)
	}

	if budget.RemainingHeader == "" {
		return errors.New(
			"provider remaining budget header is required",
		)
	}

	if budget.RetryAfterSecondsHeader == "" {
		return errors.New(
			"provider retry-after header is required",
		)
	}

	if budget.Provenance == "" {
		return errors.New(
			"provider-reported budget provenance is required",
		)
	}

	if budget.Reference == "" {
		return errors.New(
			"provider-reported budget reference is required",
		)
	}

	return nil
}

func validatePublicationDrivenPolicy(
	policy Policy,
) error {
	publication := policy.PublicationPolicy
	if publication == nil {
		return errors.New(
			"publication-driven policy requires publication metadata",
		)
	}

	if publication.Cadence == "" {
		return errors.New(
			"publication cadence is required",
		)
	}

	if publication.Provenance == "" {
		return errors.New(
			"publication provenance is required",
		)
	}

	if publication.Reference == "" {
		return errors.New(
			"publication reference is required",
		)
	}

	return nil
}

func isSupportedWindow(
	window Window,
) bool {
	switch window {
	case WindowSecond,
		WindowMinute,
		WindowHour,
		WindowDay,
		WindowMonth:
		return true
	default:
		return false
	}
}

func airplanesLivePolicy() Policy {
	const reference = "https://airplanes.live/api-guide/"

	return Policy{
		Provider:   ProviderAirplanesLive,
		BudgetMode: BudgetModeFixedWindow,
		RequestLimits: []RequestLimit{
			{
				MaxRequests: 1,
				Window:      WindowSecond,
				Provenance:  ProvenanceSourceBacked,
				Reference:   reference,
			},
		},
	}
}

func openMeteoPolicy() Policy {
	const reference = "https://open-meteo.com/en/terms"

	return Policy{
		Provider:   ProviderOpenMeteo,
		BudgetMode: BudgetModeFixedWindow,
		RequestLimits: []RequestLimit{
			{
				MaxRequests: 600,
				Window:      WindowMinute,
				Provenance:  ProvenanceSourceBacked,
				Reference:   reference,
			},
			{
				MaxRequests: 5000,
				Window:      WindowHour,
				Provenance:  ProvenanceSourceBacked,
				Reference:   reference,
			},
			{
				MaxRequests: 10000,
				Window:      WindowDay,
				Provenance:  ProvenanceSourceBacked,
				Reference:   reference,
			},
			{
				MaxRequests: 300000,
				Window:      WindowMonth,
				Provenance:  ProvenanceSourceBacked,
				Reference:   reference,
			},
		},
	}
}

func ourAirportsPolicy() Policy {
	const reference = "https://ourairports.com/data/"

	return Policy{
		Provider:   ProviderOurAirports,
		BudgetMode: BudgetModePublicationDriven,
		PublicationPolicy: &PublicationPolicy{
			Cadence:    "nightly",
			Provenance: ProvenanceSourceBacked,
			Reference:  reference,
		},
	}
}

func openSkyPolicy() Policy {
	const reference = "https://openskynetwork.github.io/opensky-api/rest.html"

	return Policy{
		Provider:   ProviderOpenSky,
		BudgetMode: BudgetModeProviderReported,
		ProviderReportedBudget: &ProviderReportedBudget{
			RemainingHeader:         "X-Rate-Limit-Remaining",
			RetryAfterSecondsHeader: "X-Rate-Limit-Retry-After-Seconds",
			Provenance:              ProvenanceProviderDirected,
			Reference:               reference,
		},
	}
}
