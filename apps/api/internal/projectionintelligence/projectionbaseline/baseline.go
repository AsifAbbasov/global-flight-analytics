package projectionbaseline

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

const (
	Version    = "short-horizon-kinematic-baseline-v1"
	MethodName = "short_horizon_kinematic_baseline"
)

var (
	ErrTrajectoryIDRequired = errors.New(
		"projection trajectory id is required",
	)
	ErrGeneratedAtInvalid = errors.New(
		"projection generated-at time must not be before the as-of time",
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
			ErrHorizonPlannerRequired
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

	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(
			plan.AsOfTime,
		) {
		return projectioncontract.Result{},
			ErrGeneratedAtInvalid
	}

	snapshot, futurePointCount :=
		trajectorySnapshotAt(
			request.Trajectory,
			plan.AsOfTime,
		)

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
	if limitation, valid :=
		validateLatestKinematics(
			latestPoint,
			baseline.config.AllowOnGround,
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

	altitudeM, altitudeAvailable :=
		usableAltitude(latestPoint)

	points, err := baseline.projectPoints(
		snapshot,
		latestPoint,
		altitudeM,
		altitudeAvailable,
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
	if !altitudeAvailable {
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
	if futurePointCount > 0 {
		limitations = append(
			limitations,
			projectioncontract.Limitation{
				Code: "future_observations_excluded",
				Message: fmt.Sprintf(
					"%d trajectory points after the as-of time were excluded from projection inputs.",
					futurePointCount,
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
		altitudeAvailable,
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
				Message: "When altitude is available, altitude is propagated using the latest observed vertical rate.",
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
	altitudeM float64,
	altitudeAvailable bool,
	plan projectionhorizon.Plan,
) ([]projectioncontract.ProjectionPoint, error) {
	result := make(
		[]projectioncontract.ProjectionPoint,
		0,
		len(plan.ForecastTimes),
	)

	horizonDurationSeconds :=
		plan.EffectiveDuration.Seconds()
	if !positiveFinite(
		horizonDurationSeconds,
	) {
		return nil,
			ErrProjectionComputationInvalid
	}

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

		if altitudeAvailable {
			projectedAltitude :=
				altitudeM +
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

		progress := forecastTime.Sub(
			plan.AsOfTime,
		).Seconds() /
			horizonDurationSeconds
		score := item.QualityScore *
			(1 -
				baseline.config.
					MaximumConfidenceLoss*
					progress)
		score = clampUnit(score)

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
							Code:         "trajectory_quality_and_horizon_decay",
							Message:      "Point confidence starts from trajectory quality and decreases with forecast horizon according to the configured maximum loss.",
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

func trajectorySnapshotAt(
	item trajectory.FlightTrajectory,
	asOfTime time.Time,
) (trajectory.FlightTrajectory, int) {
	asOfTime = asOfTime.UTC()
	points := make(
		[]trajectory.TrackPoint4D,
		0,
		len(item.Points),
	)
	futurePointCount := 0

	for _, point := range item.Points {
		if point.ObservedAt.IsZero() {
			continue
		}
		if point.ObservedAt.UTC().After(
			asOfTime,
		) {
			futurePointCount++
			continue
		}
		points = append(
			points,
			point,
		)
	}

	sort.SliceStable(
		points,
		func(left int, right int) bool {
			leftTime := points[left].
				ObservedAt.UTC()
			rightTime := points[right].
				ObservedAt.UTC()
			if leftTime.Equal(rightTime) {
				return points[left].ID <
					points[right].ID
			}

			return leftTime.Before(
				rightTime,
			)
		},
	)

	gaps := make(
		[]trajectory.CoverageGap,
		0,
		len(item.CoverageGaps),
	)
	for _, gap := range item.CoverageGaps {
		if gap.StartTime.IsZero() ||
			gap.StartTime.UTC().After(
				asOfTime,
			) {
			continue
		}
		gaps = append(
			gaps,
			gap,
		)
	}

	snapshot := item
	snapshot.Points = points
	snapshot.PointCount = len(points)
	snapshot.CoverageGaps = gaps
	snapshot.CoverageGapCount = len(gaps)

	if len(points) == 0 {
		snapshot.StartTime = time.Time{}
		snapshot.EndTime = time.Time{}
		snapshot.DurationSeconds = 0
		return snapshot, futurePointCount
	}

	snapshot.StartTime =
		points[0].ObservedAt.UTC()
	snapshot.EndTime =
		points[len(points)-1].
			ObservedAt.UTC()
	snapshot.DurationSeconds = int64(
		snapshot.EndTime.Sub(
			snapshot.StartTime,
		).Seconds(),
	)
	if snapshot.UpdatedAt.After(asOfTime) {
		snapshot.UpdatedAt = asOfTime
	}

	return snapshot, futurePointCount
}

func validateLatestKinematics(
	point trajectory.TrackPoint4D,
	allowOnGround bool,
) (projectioncontract.Limitation, bool) {
	switch {
	case !finiteLatitude(point.Latitude) ||
		!finiteLongitude(point.Longitude):
		return projectioncontract.Limitation{
			Code:    "projection_position_invalid",
			Message: "Latest trajectory position is invalid.",
			Scope:   "input",
		}, false

	case !nonNegativeFinite(
		point.VelocityMPS,
	):
		return projectioncontract.Limitation{
			Code:    "projection_velocity_invalid",
			Message: "Latest trajectory velocity is invalid.",
			Scope:   "input",
		}, false

	case !finite(point.HeadingDegrees):
		return projectioncontract.Limitation{
			Code:    "projection_heading_invalid",
			Message: "Latest trajectory heading is invalid.",
			Scope:   "input",
		}, false

	case !finite(point.VerticalRateMPS):
		return projectioncontract.Limitation{
			Code:    "projection_vertical_rate_invalid",
			Message: "Latest trajectory vertical rate is invalid.",
			Scope:   "input",
		}, false

	case point.OnGround && !allowOnGround:
		return projectioncontract.Limitation{
			Code:    "projection_on_ground_not_allowed",
			Message: "Configured projection policy does not allow an on-ground baseline.",
			Scope:   "input",
		}, false

	default:
		return projectioncontract.Limitation{}, true
	}
}

func usableAltitude(
	point trajectory.TrackPoint4D,
) (float64, bool) {
	geometricStatus :=
		flightstate.ResolveAltitudeStatus(
			point.GeometricAltitudeM,
			point.GeometricAltitudeStatus,
		)
	if usableAltitudeStatus(
		geometricStatus,
	) &&
		finite(point.GeometricAltitudeM) {
		return point.GeometricAltitudeM,
			true
	}

	barometricStatus :=
		flightstate.ResolveAltitudeStatus(
			point.BarometricAltitudeM,
			point.BarometricAltitudeStatus,
		)
	if usableAltitudeStatus(
		barometricStatus,
	) &&
		finite(point.BarometricAltitudeM) {
		return point.BarometricAltitudeM,
			true
	}

	return 0, false
}

func usableAltitudeStatus(
	status flightstate.AltitudeStatus,
) bool {
	return status ==
		flightstate.AltitudeStatusObserved ||
		status ==
			flightstate.AltitudeStatusGround
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
	altitudeAvailable bool,
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
		},
		{
			Name: "trajectory_quality",
			Classification: projectioncontract.
				InputClassificationDerived,
			SourceName: "trajectory_quality",
			ObservedAt: observedAt,
		},
	}

	if altitudeAvailable {
		result = append(
			result,
			projectioncontract.InputReference{
				Name: "altitude",
				Classification: projectioncontract.
					InputClassificationObserved,
				SourceName: sourceName,
				ObservedAt: observedAt,
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

func clampUnit(
	value float64,
) float64 {
	if math.IsNaN(value) ||
		math.IsInf(value, 0) ||
		value <= 0 {
		return 0
	}
	if value >= 1 {
		return 1
	}

	return value
}

func float64Pointer(
	value float64,
) *float64 {
	return &value
}
