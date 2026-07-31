package projectionproduction

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontinuation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

func validateProjectionPostconditions(
	result projectioncontract.Result,
	request Request,
	plan projectionhorizon.Plan,
	strategy Strategy,
) error {
	report := projectioncontract.Validate(result)
	if report.Status != projectioncontract.ValidationStatusValid {
		return fmt.Errorf("%w: %#v", ErrProjectionContractInvalid, report.Issues)
	}
	if result.TrajectoryID != request.CurrentTrajectory.ID ||
		result.FlightID != request.CurrentTrajectory.FlightID ||
		result.AircraftID != request.CurrentTrajectory.AircraftID ||
		result.ICAO24 != request.CurrentTrajectory.ICAO24 ||
		result.Callsign != request.CurrentTrajectory.Callsign {
		return fmt.Errorf("%w: projection identity does not match current trajectory", ErrProjectionContractInvalid)
	}
	if !projectionHorizonMatchesPlan(result, plan) ||
		!result.GeneratedAt.Equal(request.GeneratedAt.UTC()) {
		return fmt.Errorf("%w: projection time boundary does not match authorized plan", ErrProjectionContractInvalid)
	}

	switch strategy {
	case StrategyHistoricalNeighbor:
		if result.Method.Name != projectioncontinuation.MethodName ||
			result.Method.Version != projectioncontinuation.Version ||
			result.Method.DecisionClass != projectioncontract.DecisionClassExperimental ||
			result.Status == projectioncontract.ResultStatusUnavailable {
			return fmt.Errorf("%w: historical projector postconditions failed", ErrHistoricalProjectionFailed)
		}
	case StrategyKinematic:
		if result.Method.Name != projectionbaseline.MethodName ||
			result.Method.Version != projectionbaseline.Version ||
			result.Method.DecisionClass != projectioncontract.DecisionClassPhysicsDerived {
			return fmt.Errorf("%w: kinematic projector postconditions failed", ErrKinematicProjectionFailed)
		}
	default:
		return fmt.Errorf("%w: unknown production strategy", ErrProjectionContractInvalid)
	}
	return nil
}

func projectionHorizonMatchesPlan(
	result projectioncontract.Result,
	plan projectionhorizon.Plan,
) bool {
	expected := plan.ContractHorizon()
	return result.Horizon.AsOfTime.Equal(expected.AsOfTime) &&
		result.Horizon.EndTime.Equal(expected.EndTime) &&
		result.Horizon.Step == expected.Step
}
