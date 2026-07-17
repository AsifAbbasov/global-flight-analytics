package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/sourceconstraints"
)

func main() {
	capabilities := []sourceconstraints.Capability{
		sourceconstraints.CapabilityRegionalLiveObservation,
		sourceconstraints.CapabilityHistoricalFlightObservation,
		sourceconstraints.CapabilityEstimatedAirportContext,
		sourceconstraints.CapabilityExperimentalTrackContext,
		sourceconstraints.CapabilityGlobalContinuousTracking,
		sourceconstraints.CapabilityOceanicContinuousTracking,
		sourceconstraints.CapabilityOwnReceiverObservation,
		sourceconstraints.CapabilityOfficialSchedule,
		sourceconstraints.CapabilityOfficialDelayCause,
		sourceconstraints.CapabilityPilotIntent,
		sourceconstraints.CapabilityATCInstruction,
		sourceconstraints.CapabilityCertifiedSeparation,
		sourceconstraints.CapabilityOperationalWeather,
		sourceconstraints.CapabilityCommercialFleetData,
	}

	decisions := make([]sourceconstraints.Decision, 0, len(capabilities))
	for _, capability := range capabilities {
		decision, err := sourceconstraints.Evaluate(sourceconstraints.Request{
			Constraints: sourceconstraints.FixedProjectConstraints(),
			Source:      sourceconstraints.OpenSkyProfile(),
			Capability:  capability,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "evaluate %s: %v\n", capability, err)
			os.Exit(1)
		}
		decisions = append(decisions, decision)
	}

	payload, err := json.MarshalIndent(decisions, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode source constraint decisions: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(payload))
}
