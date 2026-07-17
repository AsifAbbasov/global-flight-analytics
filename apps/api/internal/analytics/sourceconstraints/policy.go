package sourceconstraints

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrSourceIDRequired        = errors.New("source id is required")
	ErrSourceClassUnknown      = errors.New("source class is unknown")
	ErrCapabilityUnknown       = errors.New("capability is unknown")
	ErrConstraintSetInvalid    = errors.New("project constraints do not match the fixed project boundary")
	ErrAttributionTextRequired = errors.New("source attribution text is required")
)

func Evaluate(request Request) (Decision, error) {
	if err := validateRequest(request); err != nil {
		return Decision{}, err
	}

	if blocked, reason := blockedByResourceBoundary(
		request.Constraints,
		request.Source,
	); blocked {
		return blockedDecision(request, reason), nil
	}

	switch request.Capability {
	case CapabilityRegionalLiveObservation:
		if !request.Source.SupportsRegionalObservation {
			return blockedDecision(
				request,
				"The source does not expose bounded regional observations.",
			), nil
		}
		return applySourceObligations(
			request.Source,
			Decision{
				ContractVersion:      ContractVersion,
				SourceID:             request.Source.ID,
				Capability:           request.Capability,
				Level:                DecisionLevelAllowed,
				MaximumClaimStrength: ClaimStrengthObserved,
				Reasons: []string{
					"The source exposes free externally collected regional observations.",
					"Observed means reported by the external source, not observed by project-owned sensors.",
				},
				RequiredLabels: []string{
					"external community observation",
					"coverage may be incomplete",
					"research use only",
				},
				ScopeGuards: commonScopeGuards(),
			},
		), nil

	case CapabilityHistoricalFlightObservation:
		if !request.Source.SupportsHistoricalFlights {
			return blockedDecision(
				request,
				"The source does not expose bounded historical flight observations.",
			), nil
		}
		return limitedDecision(
			request,
			ClaimStrengthDerived,
			"Historical flight records are provider-processed observations with bounded retention and availability.",
			"provider-derived historical record",
		), nil

	case CapabilityEstimatedAirportContext:
		if !request.Source.SupportsEstimatedAirports {
			return blockedDecision(
				request,
				"The source does not expose estimated airport context.",
			), nil
		}
		return limitedDecision(
			request,
			ClaimStrengthEstimated,
			"Departure and arrival airports are estimates and are not official airport operations data.",
			"estimated airport context",
		), nil

	case CapabilityExperimentalTrackContext:
		if !request.Source.SupportsExperimentalTracks {
			return blockedDecision(
				request,
				"The source does not expose experimental track context.",
			), nil
		}
		return limitedDecision(
			request,
			ClaimStrengthDerived,
			"The track endpoint is experimental and cannot replace the project Track Builder.",
			"experimental provider track",
		), nil

	case CapabilityGlobalContinuousTracking:
		if !request.Source.ContinuousCoverage {
			return blockedDecision(
				request,
				"Continuous global coverage is unavailable from the configured free external source.",
			), nil
		}
		return blockedDecision(
			request,
			"The project cannot claim continuous global tracking without owned collection, satellite, or commercial coverage.",
		), nil

	case CapabilityOceanicContinuousTracking:
		if !request.Source.OceanicCoverage {
			return blockedDecision(
				request,
				"Continuous oceanic coverage is unavailable from the configured source.",
			), nil
		}
		return blockedDecision(
			request,
			"The project has no satellite or commercial oceanic surveillance access.",
		), nil

	case CapabilityOwnReceiverObservation:
		return blockedDecision(
			request,
			"The project has no receiver network or first-party sensor infrastructure.",
		), nil

	case CapabilityOfficialSchedule:
		return blockedDecision(
			request,
			"Official schedules require an authoritative operational or licensed commercial source.",
		), nil

	case CapabilityOfficialDelayCause:
		return blockedDecision(
			request,
			"Official delay causes cannot be inferred from open surveillance observations.",
		), nil

	case CapabilityPilotIntent:
		return blockedDecision(
			request,
			"Pilot intent is not observable from the available open data.",
		), nil

	case CapabilityATCInstruction:
		return blockedDecision(
			request,
			"Air traffic control instructions are not available in the configured sources.",
		), nil

	case CapabilityCertifiedSeparation:
		return blockedDecision(
			request,
			"Free external observations cannot support certified separation monitoring or safety-critical alerts.",
		), nil

	case CapabilityOperationalWeather:
		return blockedDecision(
			request,
			"Public weather context is not certified operational aviation weather.",
		), nil

	case CapabilityCommercialFleetData:
		return blockedDecision(
			request,
			"The project has no licensed commercial aviation data access.",
		), nil

	default:
		return Decision{}, fmt.Errorf(
			"%w: %q",
			ErrCapabilityUnknown,
			request.Capability,
		)
	}
}

func validateRequest(request Request) error {
	if strings.TrimSpace(request.Source.ID) == "" {
		return ErrSourceIDRequired
	}
	if !isKnownSourceClass(request.Source.Class) {
		return fmt.Errorf("%w: %q", ErrSourceClassUnknown, request.Source.Class)
	}
	if request.Constraints != FixedProjectConstraints() {
		return ErrConstraintSetInvalid
	}
	if request.Source.AttributionRequired &&
		strings.TrimSpace(request.Source.AttributionText) == "" {
		return ErrAttributionTextRequired
	}
	if !isKnownCapability(request.Capability) {
		return fmt.Errorf("%w: %q", ErrCapabilityUnknown, request.Capability)
	}
	return nil
}

func blockedByResourceBoundary(
	constraints ProjectConstraints,
	source SourceProfile,
) (bool, string) {
	if constraints.FreeSourcesOnly && !source.FreeAccess {
		return true, "The source is not available under the free-source-only project boundary."
	}
	if source.RequiresOwnInfrastructure &&
		!constraints.HasOwnCollectionInfrastructure {
		return true, "The source requires first-party collection infrastructure that the project does not own."
	}
	if source.SatelliteDerived && !constraints.HasSatelliteAccess {
		return true, "The source requires satellite access that the project does not have."
	}
	if source.Commercial &&
		!constraints.HasCommercialAviationDataAccess {
		return true, "The source requires licensed commercial aviation data access."
	}
	if source.OfficialOperational && constraints.ResearchOnly {
		return true, "The source is official operational data and is outside the research-only project boundary."
	}
	return false, ""
}

func limitedDecision(
	request Request,
	strength ClaimStrength,
	reason string,
	label string,
) Decision {
	return applySourceObligations(request.Source, Decision{
		ContractVersion:      ContractVersion,
		SourceID:             request.Source.ID,
		Capability:           request.Capability,
		Level:                DecisionLevelLimited,
		MaximumClaimStrength: strength,
		Reasons: []string{
			reason,
			"The result must preserve provider provenance and uncertainty.",
		},
		RequiredLabels: []string{
			label,
			"not official operational data",
			"research use only",
		},
		ScopeGuards: commonScopeGuards(),
	})
}

func applySourceObligations(
	source SourceProfile,
	decision Decision,
) Decision {
	if source.AttributionRequired {
		decision.Reasons = appendUnique(
			decision.Reasons,
			"Published outputs must preserve the provider attribution requirement.",
		)
		decision.RequiredLabels = appendUnique(
			decision.RequiredLabels,
			"OpenSky Network attribution required",
		)
		decision.ScopeGuards = appendUnique(
			decision.ScopeGuards,
			"Publish the required source citation on public web pages, articles, and presentations.",
		)
	}
	if source.NonCommercialUseOnly {
		decision.RequiredLabels = appendUnique(
			decision.RequiredLabels,
			"non-commercial research use only",
		)
		decision.ScopeGuards = appendUnique(
			decision.ScopeGuards,
			"Do not convert the free research feed into an unsupported commercial real-time service.",
		)
	}
	if source.CloudHostingAvailabilityNotGuaranteed {
		decision.ScopeGuards = appendUnique(
			decision.ScopeGuards,
			"Treat provider access from large cloud-hosting IP ranges as non-guaranteed and retain fallback behavior.",
		)
	}
	return decision
}

func appendUnique(values []string, candidate string) []string {
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
}

func blockedDecision(request Request, reason string) Decision {
	return Decision{
		ContractVersion:      ContractVersion,
		SourceID:             request.Source.ID,
		Capability:           request.Capability,
		Level:                DecisionLevelBlocked,
		MaximumClaimStrength: ClaimStrengthBlocked,
		Reasons:              []string{reason},
		RequiredLabels: []string{
			"capability unavailable",
			"research use only",
		},
		ScopeGuards: commonScopeGuards(),
	}
}

func commonScopeGuards() []string {
	return []string{
		"Do not claim first-party sensor observation.",
		"Do not claim satellite coverage.",
		"Do not claim commercial or official operational authority.",
		"Do not convert estimated or derived evidence into observed fact.",
		"Do not publish safety-critical or directive aviation guidance.",
	}
}

func isKnownSourceClass(value SourceClass) bool {
	switch value {
	case SourceClassPublicCommunitySurveillance,
		SourceClassPublicStaticDataset,
		SourceClassPublicWeatherService,
		SourceClassFirstPartyReceiverNetwork,
		SourceClassSatelliteSurveillance,
		SourceClassCommercialAviation,
		SourceClassOfficialOperational:
		return true
	default:
		return false
	}
}

func isKnownCapability(value Capability) bool {
	switch value {
	case CapabilityRegionalLiveObservation,
		CapabilityHistoricalFlightObservation,
		CapabilityEstimatedAirportContext,
		CapabilityExperimentalTrackContext,
		CapabilityGlobalContinuousTracking,
		CapabilityOceanicContinuousTracking,
		CapabilityOwnReceiverObservation,
		CapabilityOfficialSchedule,
		CapabilityOfficialDelayCause,
		CapabilityPilotIntent,
		CapabilityATCInstruction,
		CapabilityCertifiedSeparation,
		CapabilityOperationalWeather,
		CapabilityCommercialFleetData:
		return true
	default:
		return false
	}
}
