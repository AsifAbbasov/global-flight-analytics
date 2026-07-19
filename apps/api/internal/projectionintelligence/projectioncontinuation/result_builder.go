package projectioncontinuation

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

func (
	baseline *Baseline,
) buildContinuationResult(
	preparation continuationPreparation,
	plan projectionhorizon.Plan,
	pointResult continuationPointResult,
	generatedAt time.Time,
) projectioncontract.Result {
	status := projectioncontract.
		ResultStatusComplete
	limitations :=
		historicalContinuationLimitations(
			preparation.selection,
			preparation.pattern,
		)

	if continuationResultLimited(
		plan,
		preparation.selection,
		preparation.pattern,
		pointResult.altitudeComplete,
	) {
		status = projectioncontract.
			ResultStatusLimited
	}
	if plan.Truncated {
		limitations = append(
			limitations,
			projectioncontract.Limitation{
				Code:    "projection_horizon_truncated",
				Message: "Requested duration exceeded the configured maximum and was truncated.",
				Scope:   "horizon",
			},
		)
	}
	if !pointResult.altitudeComplete {
		limitations = append(
			limitations,
			projectioncontract.Limitation{
				Code:    "historical_continuation_altitude_partial",
				Message: "At least one forecast point lacked sufficient historical altitude support, so only horizontal position was published for that point.",
				Scope:   "position",
			},
		)
	}

	return projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        status,

		TrajectoryID: preparation.current.ID,
		FlightID:     preparation.current.FlightID,
		AircraftID:   preparation.current.AircraftID,
		ICAO24:       preparation.current.ICAO24,
		Callsign:     preparation.current.Callsign,

		Method: projectioncontract.Method{
			Name:    MethodName,
			Version: Version,
			DecisionClass: projectioncontract.
				DecisionClassExperimental,
		},
		Horizon: plan.ContractHorizon(),
		Points:  pointResult.points,

		Confidence: minimumPointConfidence(
			pointResult.points,
		),
		Limitations: normalizeLimitations(
			limitations,
		),
		Explanations: []projectioncontract.Explanation{
			{
				Code:    "translated_historical_continuations",
				Message: "Observed movements after each selected historical anchor were translated onto the current trajectory endpoint.",
			},
			{
				Code:    "similarity_weighted_consensus",
				Message: "Forecast coordinates combine usable historical continuations using normalized similarity weights.",
			},
			{
				Code:    "neighbor_disagreement_uncertainty",
				Message: "Published uncertainty includes configured growth and weighted disagreement between historical continuation samples.",
			},
		},
		ScopeGuard: projectioncontract.
			ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: continuationFingerprint(
				preparation.current,
				preparation.selection,
				preparation.pattern,
				plan,
				baseline.config,
			),
			Inputs: continuationInputs(
				preparation.currentEndpoint,
				preparation.selection,
			),
			LatestInputObservedAt: preparation.
				currentEndpoint.
				ObservedAt.UTC(),
		},
		GeneratedAt: generatedAt,
	}
}

func continuationResultLimited(
	plan projectionhorizon.Plan,
	selection projectionneighbors.Result,
	pattern projectionpatternconfidence.Result,
	altitudeComplete bool,
) bool {
	return plan.Truncated ||
		selection.Status !=
			projectionneighbors.StatusComplete ||
		pattern.Status !=
			projectionpatternconfidence.
				StatusComplete ||
		!altitudeComplete
}
