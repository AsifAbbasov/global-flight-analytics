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
		pointResult.confidenceComplete,
		pointResult.plausibilityFiltered,
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
	if pointResult.plausibilityFiltered {
		limitations = append(
			limitations,
			projectioncontract.Limitation{
				Code:    "historical_continuation_plausibility_filtered",
				Message: "At least one historical continuation sample was excluded because its interpolation gap, horizontal speed, or vertical speed exceeded the configured plausibility policy.",
				Scope:   "support",
			},
		)
	}
	if !pointResult.confidenceComplete {
		limitations = append(
			limitations,
			projectioncontract.Limitation{
				Code:    "historical_continuation_confidence_none",
				Message: "At least one forecast point has zero confidence, so the projection cannot be classified as complete.",
				Scope:   "confidence",
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
				Code:    "interpolation_plausibility_policy",
				Message: "Historical continuation segments are accepted only when their time gap and implied horizontal and vertical rates satisfy the configured plausibility policy.",
			},
			{
				Code:    "neighbor_disagreement_uncertainty_and_confidence",
				Message: "Configured uncertainty and weighted neighbor disagreement are conservatively added; the resulting agreement factor also reduces confidence.",
			},
			{
				Code:    "effective_weighted_support",
				Message: "Usable historical support is measured with effective sample size so concentrated similarity weights cannot overstate confidence.",
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
				preparation.candidateByID,
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
	confidenceComplete bool,
	plausibilityFiltered bool,
) bool {
	return plan.Truncated ||
		selection.Status !=
			projectionneighbors.StatusComplete ||
		pattern.Status !=
			projectionpatternconfidence.
				StatusComplete ||
		!altitudeComplete ||
		!confidenceComplete ||
		plausibilityFiltered
}
