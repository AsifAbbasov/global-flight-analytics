package projectionbaseline

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

const (
	Version    = "short-horizon-kinematic-baseline-v2"
	MethodName = "short_horizon_kinematic_baseline"
)

var (
	ErrBaselineUnavailable = errors.New(
		"projection baseline is unavailable",
	)
	ErrTrajectoryIDRequired = errors.New(
		"projection trajectory id is required",
	)
	ErrGeneratedAtInvalid = errors.New(
		"projection generated-at time must not be before the as-of time",
	)
	ErrHorizonPlanInvalid = errors.New(
		"projection horizon planner returned an invalid plan",
	)
	ErrProjectionContractInvalid = errors.New(
		"generated projection contract is invalid",
	)
	ErrProjectionComputationInvalid = errors.New(
		"projection computation produced an invalid value",
	)
)

type Baseline struct {
	config Config
}

func New(
	config Config,
) (*Baseline, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate short-horizon projection baseline config: %w",
			err,
		)
	}

	return &Baseline{
		config: config,
	}, nil
}

type Request struct {
	Trajectory trajectory.FlightTrajectory

	AsOfTime          time.Time
	RequestedDuration time.Duration
	GeneratedAt       time.Time
}

func (
	baseline *Baseline,
) Project(
	request Request,
) (projectioncontract.Result, error) {
	if baseline == nil {
		return projectioncontract.Result{},
			ErrBaselineUnavailable
	}
	if strings.TrimSpace(
		request.Trajectory.ID,
	) == "" {
		return projectioncontract.Result{},
			ErrTrajectoryIDRequired
	}

	plan, err := baseline.config.
		HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime: request.AsOfTime,
			RequestedDuration: request.
				RequestedDuration,
		},
	)
	if err != nil {
		return projectioncontract.Result{},
			fmt.Errorf(
				"build projection horizon: %w",
				err,
			)
	}
	if err := plan.Validate(); err != nil {
		return projectioncontract.Result{},
			fmt.Errorf(
				"%w: %v",
				ErrHorizonPlanInvalid,
				err,
			)
	}

	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(
			plan.AsOfTime,
		) {
		return projectioncontract.Result{},
			ErrGeneratedAtInvalid
	}

	cutoff := buildCutoffSnapshot(
		request.Trajectory,
		plan.AsOfTime,
	)
	snapshot := cutoff.Trajectory
	if !cutoff.QualityEvidenceAvailable {
		return baseline.validatedUnavailable(
			snapshot,
			plan,
			generatedAt,
			[]projectioncontract.Limitation{
				{
					Code:    "projection_cutoff_quality_unavailable",
					Message: "Cutoff-safe trajectory quality evidence does not cover the latest included observation.",
					Scope:   "provenance",
				},
			},
		)
	}

	evaluation := baseline.config.
		EligibilityEvaluator.Evaluate(
		snapshot,
		plan.AsOfTime,
	)
	decision, exists := evaluation.Decision(
		trajectoryeligibility.
			CapabilityProjection,
	)
	if !exists {
		return baseline.validatedUnavailable(
			snapshot,
			plan,
			generatedAt,
			[]projectioncontract.Limitation{
				{
					Code:    "projection_eligibility_decision_missing",
					Message: "Projection eligibility did not return a projection decision.",
					Scope:   "result",
				},
			},
		)
	}
	if !decision.Allowed {
		return baseline.validatedUnavailable(
			snapshot,
			plan,
			generatedAt,
			eligibilityLimitations(
				decision.Reasons,
			),
		)
	}

	if len(snapshot.Points) == 0 {
		return baseline.validatedUnavailable(
			snapshot,
			plan,
			generatedAt,
			[]projectioncontract.Limitation{
				{
					Code:    "projection_point_unavailable",
					Message: "No trajectory point was available at or before the as-of time.",
					Scope:   "input",
				},
			},
		)
	}

	latestPoint := snapshot.Points[len(snapshot.Points)-1]
	kinematics := kinematicPolicy{
		AllowOnGround: baseline.config.AllowOnGround,
	}
	if limitation, valid := kinematics.validate(
		latestPoint,
	); !valid {
		return baseline.validatedUnavailable(
			snapshot,
			plan,
			generatedAt,
			[]projectioncontract.Limitation{
				limitation,
			},
		)
	}

	altitude := selectAltitude(latestPoint)

	points, err := baseline.projectPoints(
		snapshot,
		latestPoint,
		altitude,
		plan,
	)
	if err != nil {
		return projectioncontract.Result{},
			err
	}

	status := projectioncontract.
		ResultStatusComplete
	limitations := baselineLimitations()
	if plan.Truncated {
		status = projectioncontract.
			ResultStatusLimited
		limitations = append(
			limitations,
			projectioncontract.Limitation{
				Code:    "projection_horizon_truncated",
				Message: "Requested duration exceeded the configured maximum and was truncated.",
				Scope:   "horizon",
			},
		)
	}
	if !altitude.Available {
		status = projectioncontract.
			ResultStatusLimited
		limitations = append(
			limitations,
			projectioncontract.Limitation{
				Code:    "projection_altitude_unavailable",
				Message: "Horizontal projection is available, but altitude could not be projected.",
				Scope:   "position",
			},
		)
	}
	if cutoff.excludedEvidenceCount() > 0 {
		limitations = append(
			limitations,
			projectioncontract.Limitation{
				Code: "future_observations_excluded",
				Message: fmt.Sprintf(
					"Cutoff isolation excluded %d points, %d segments, and %d coverage gaps that were not fully available at the as-of time.",
					cutoff.ExcludedPointCount,
					cutoff.ExcludedSegmentCount,
					cutoff.ExcludedGapCount,
				),
				Scope: "provenance",
			},
		)
	}

	resultConfidence := minimumPointConfidence(
		points,
	)
	fingerprint := inputFingerprint(
		snapshot,
		latestPoint,
		plan,
		baseline.config,
	)

	inputs := projectionInputs(
		snapshot,
		latestPoint,
		altitude,
	)

	result := projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        status,

		TrajectoryID: snapshot.ID,
		FlightID:     snapshot.FlightID,
		AircraftID:   snapshot.AircraftID,
		ICAO24:       snapshot.ICAO24,
		Callsign:     snapshot.Callsign,

		Method: projectioncontract.Method{
			Name:    MethodName,
			Version: Version,
			DecisionClass: projectioncontract.
				DecisionClassPhysicsDerived,
		},
		Horizon: plan.ContractHorizon(),
		Points:  points,

		Confidence: resultConfidence,
		Limitations: append(
			[]projectioncontract.Limitation(nil),
			limitations...,
		),
		Explanations: []projectioncontract.Explanation{
			{
				Code:    "constant_ground_track_propagation",
				Message: "Each forecast point propagates the latest observed ground speed and heading over a spherical direct-geodesic step.",
			},
			{
				Code:    "linear_vertical_rate_propagation",
				Message: "When altitude is available, the selected geometric or barometric altitude reference is propagated using the latest observed vertical rate.",
			},
			{
				Code:    "explicit_uncertainty_growth",
				Message: "Horizontal and vertical uncertainty grow from caller-provided baseline values and rates.",
			},
		},
		ScopeGuard: projectioncontract.
			ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: fingerprint,
			Inputs:           inputs,
			LatestInputObservedAt: latestPoint.
				ObservedAt.UTC(),
		},
		GeneratedAt: generatedAt,
	}

	return validateResult(result)
}

func (
	baseline *Baseline,
) projectPoints(
	item trajectory.FlightTrajectory,
	latestPoint trajectory.TrackPoint4D,
	altitude altitudeSelection,
	plan projectionhorizon.Plan,
) ([]projectioncontract.ProjectionPoint, error) {
	result := make(
		[]projectioncontract.ProjectionPoint,
		0,
		len(plan.ForecastTimes),
	)

	for index, forecastTime := range plan.ForecastTimes {
		motionSeconds := forecastTime.Sub(
			latestPoint.ObservedAt.UTC(),
		).Seconds()
		if motionSeconds < 0 ||
			!finite(motionSeconds) {
			return nil,
				ErrProjectionComputationInvalid
		}

		distanceM := latestPoint.
			VelocityMPS * motionSeconds
		latitude, longitude, valid :=
			destinationPoint(
				latestPoint.Latitude,
				latestPoint.Longitude,
				latestPoint.HeadingDegrees,
				distanceM,
			)
		if !valid {
			return nil,
				ErrProjectionComputationInvalid
		}

		position := projectioncontract.Position{
			Latitude:  latitude,
			Longitude: longitude,
		}
		uncertainty :=
			projectioncontract.Uncertainty{
				HorizontalRadiusM: baseline.config.
					InitialHorizontalUncertaintyM +
					baseline.config.
						HorizontalUncertaintyGrowthMPS*
						motionSeconds,
			}

		if altitude.Available {
			projectedAltitude :=
				altitude.ValueM +
					latestPoint.
						VerticalRateMPS*
						motionSeconds
			verticalUncertainty :=
				baseline.config.
					InitialVerticalUncertaintyM +
					baseline.config.
						VerticalUncertaintyGrowthMPS*
						motionSeconds

			if !finite(projectedAltitude) ||
				!positiveFinite(
					verticalUncertainty,
				) {
				return nil,
					ErrProjectionComputationInvalid
			}

			position.AltitudeM =
				float64Pointer(
					projectedAltitude,
				)
			uncertainty.VerticalRadiusM =
				float64Pointer(
					verticalUncertainty,
				)
		}

		if !positiveFinite(
			uncertainty.HorizontalRadiusM,
		) {
			return nil,
				ErrProjectionComputationInvalid
		}

		score, confidenceValid := calculatePointConfidence(
			pointConfidenceInput{
				TrajectoryQuality:     item.QualityScore,
				LatestObservedAt:      latestPoint.ObservedAt,
				AsOfTime:              plan.AsOfTime,
				ForecastTime:          forecastTime,
				MaximumObservationAge: baseline.config.effectiveMaximumObservationAge(),
				EffectiveHorizon:      plan.EffectiveDuration,
				MaximumHorizonLoss:    baseline.config.MaximumConfidenceLoss,
			},
		)
		if !confidenceValid {
			return nil, ErrProjectionComputationInvalid
		}

		result = append(
			result,
			projectioncontract.ProjectionPoint{
				Sequence:     index,
				ForecastTime: forecastTime.UTC(),
				Position:     position,
				Uncertainty:  uncertainty,
				Confidence: projectioncontract.Confidence{
					Score: score,
					Level: baseline.
						confidenceLevel(score),
					Reasons: []projectioncontract.ConfidenceReason{
						{
							Code:         "trajectory_quality_observation_age_and_horizon_decay",
							Message:      "Point confidence combines cutoff-safe trajectory quality, latest-observation age, and forecast-horizon decay.",
							Contribution: score,
						},
					},
				},
			},
		)
	}

	return result, nil
}

func (
	baseline *Baseline,
) confidenceLevel(
	score float64,
) projectioncontract.ConfidenceLevel {
	switch {
	case score >= baseline.config.
		HighConfidenceMinimum:
		return projectioncontract.
			ConfidenceLevelHigh
	case score >= baseline.config.
		MediumConfidenceMinimum:
		return projectioncontract.
			ConfidenceLevelMedium
	case score > 0:
		return projectioncontract.
			ConfidenceLevelLow
	default:
		return projectioncontract.
			ConfidenceLevelNone
	}
}

func (
	baseline *Baseline,
) validatedUnavailable(
	item trajectory.FlightTrajectory,
	plan projectionhorizon.Plan,
	generatedAt time.Time,
	limitations []projectioncontract.Limitation,
) (projectioncontract.Result, error) {
	if len(limitations) == 0 {
		limitations = []projectioncontract.Limitation{
			{
				Code:    "projection_unavailable",
				Message: "Projection is unavailable for the supplied trajectory.",
				Scope:   "result",
			},
		}
	}

	result := projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status: projectioncontract.
			ResultStatusUnavailable,

		TrajectoryID: item.ID,
		FlightID:     item.FlightID,
		AircraftID:   item.AircraftID,
		ICAO24:       item.ICAO24,
		Callsign:     item.Callsign,

		Method: projectioncontract.Method{
			Name:    MethodName,
			Version: Version,
			DecisionClass: projectioncontract.
				DecisionClassPhysicsDerived,
		},
		Horizon: plan.ContractHorizon(),
		Confidence: projectioncontract.Confidence{
			Score: 0,
			Level: projectioncontract.
				ConfidenceLevelNone,
		},
		Limitations: append(
			[]projectioncontract.Limitation(nil),
			limitations...,
		),
		ScopeGuard: projectioncontract.
			ScopeGuardResearchOnly,
		Provenance: unavailableProvenance(
			item,
			plan,
			baseline.config,
		),
		GeneratedAt: generatedAt,
	}

	return validateResult(result)
}

func validateResult(
	result projectioncontract.Result,
) (projectioncontract.Result, error) {
	report := projectioncontract.Validate(
		result,
	)
	if report.Status !=
		projectioncontract.
			ValidationStatusValid {
		return projectioncontract.Result{},
			fmt.Errorf(
				"%w: %#v",
				ErrProjectionContractInvalid,
				report.Issues,
			)
	}

	return result.Clone(), nil
}

func eligibilityLimitations(
	reasons []trajectoryeligibility.ReasonCode,
) []projectioncontract.Limitation {
	if len(reasons) == 0 {
		return []projectioncontract.Limitation{
			{
				Code:    "projection_eligibility_denied",
				Message: "Projection eligibility denied the trajectory without a reason code.",
				Scope:   "eligibility",
			},
		}
	}

	result := make(
		[]projectioncontract.Limitation,
		0,
		len(reasons),
	)
	for _, reason := range reasons {
		result = append(
			result,
			projectioncontract.Limitation{
				Code: "projection_eligibility_" +
					string(reason),
				Message: "Projection eligibility denied the trajectory because " +
					strings.ReplaceAll(
						string(reason),
						"_",
						" ",
					) +
					".",
				Scope: "eligibility",
			},
		)
	}

	return result
}

func baselineLimitations() []projectioncontract.Limitation {
	return []projectioncontract.Limitation{
		{
			Code:    "constant_ground_track_assumption",
			Message: "Ground speed and heading are held constant across the short projection horizon.",
			Scope:   "method",
		},
		{
			Code:    "no_wind_adjustment",
			Message: "Wind and weather are not applied by this baseline.",
			Scope:   "method",
		},
		{
			Code:    "no_operational_intent",
			Message: "Official flight plan, Air Traffic Control intent, pilot intent, and future maneuvers are unavailable.",
			Scope:   "method",
		},
		{
			Code:    "research_only",
			Message: "Projection is a research estimate and must not be used for operational aviation decisions.",
			Scope:   "result",
		},
	}
}

func projectionInputs(
	item trajectory.FlightTrajectory,
	point trajectory.TrackPoint4D,
	altitude altitudeSelection,
) []projectioncontract.InputReference {
	sourceName := strings.TrimSpace(
		point.SourceName,
	)
	observedAt := point.ObservedAt.UTC()

	result := []projectioncontract.InputReference{
		{
			Name: "latest_position",
			Classification: projectioncontract.
				InputClassificationObserved,
			SourceName: sourceName,
			ObservedAt: observedAt,
		},
		{
			Name: "ground_speed",
			Classification: projectioncontract.
				InputClassificationObserved,
			SourceName: sourceName,
			ObservedAt: observedAt,
		},
		{
			Name: "ground_track",
			Classification: projectioncontract.
				InputClassificationObserved,
			SourceName: sourceName,
			ObservedAt: observedAt,
		},
		{
			Name: "vertical_rate",
			Classification: projectioncontract.
				InputClassificationObserved,
			SourceName: sourceName,
			ObservedAt: observedAt,
			Limitation: "Source telemetry does not identify whether vertical rate uses the selected geometric or barometric altitude reference.",
		},
		{
			Name: "trajectory_quality",
			Classification: projectioncontract.
				InputClassificationDerived,
			SourceName: "trajectory_quality",
			ObservedAt: observedAt,
		},
	}

	if altitude.Available {
		result = append(
			result,
			projectioncontract.InputReference{
				Name: "altitude",
				Classification: projectioncontract.
					InputClassificationObserved,
				SourceName: sourceName,
				ObservedAt: observedAt,
				Limitation: altitude.provenanceLimitation(),
			},
		)
	}

	return result
}

func minimumPointConfidence(
	points []projectioncontract.ProjectionPoint,
) projectioncontract.Confidence {
	if len(points) == 0 {
		return projectioncontract.Confidence{
			Score: 0,
			Level: projectioncontract.
				ConfidenceLevelNone,
		}
	}

	minimum := points[0].Confidence
	for _, point := range points[1:] {
		if point.Confidence.Score <
			minimum.Score {
			minimum = point.Confidence
		}
	}

	minimum.Reasons = []projectioncontract.ConfidenceReason{
		{
			Code:         "minimum_point_confidence",
			Message:      "Result confidence equals the lowest confidence across projected points.",
			Contribution: minimum.Score,
		},
	}

	return minimum
}

func float64Pointer(
	value float64,
) *float64 {
	return &value
}
