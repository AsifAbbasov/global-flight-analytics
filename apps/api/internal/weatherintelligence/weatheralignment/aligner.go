package weatheralignment

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
)

var (
	ErrInvalidPolicy          = errors.New("weather alignment policy is invalid")
	ErrTrajectoryInvalid      = errors.New("weather alignment trajectory is invalid")
	ErrWeatherContractInvalid = errors.New("weather alignment weather contract is invalid")
	ErrTrustResultInvalid     = errors.New("weather alignment trust result is invalid")
	ErrResultInvalid          = errors.New("weather alignment result is invalid")
)

type Request struct {
	Trajectory  trajectory.FlightTrajectory
	Weather     weathercontract.Result
	Trust       weathertrust.Result
	Policy      Policy
	GeneratedAt time.Time
}

func Align(request Request) (Result, error) {
	if err := request.Policy.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrInvalidPolicy, err)
	}
	if err := validateTrajectory(request.Trajectory, request.Weather.AsOfTime); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrTrajectoryInvalid, err)
	}
	weatherReport := weathercontract.Validate(request.Weather)
	if weatherReport.Status != weathercontract.ValidationStatusValid {
		return Result{}, fmt.Errorf("%w: issues=%v", ErrWeatherContractInvalid, weatherReport.Issues)
	}
	if err := request.Trust.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrTrustResultInvalid, err)
	}
	if !request.Trust.AsOfTime.Equal(request.Weather.AsOfTime) {
		return Result{}, fmt.Errorf("%w: trust and weather as-of times differ", ErrTrustResultInvalid)
	}
	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() || generatedAt.Before(request.Weather.AsOfTime) {
		return Result{}, fmt.Errorf("%w: generated-at time is invalid", ErrTrajectoryInvalid)
	}

	result := Result{
		Version:       Version,
		TrajectoryID:  strings.TrimSpace(request.Trajectory.ID),
		AsOfTime:      request.Weather.AsOfTime.UTC(),
		TrustDecision: request.Trust.Decision,
		TrustScore:    request.Trust.Score,
		PointCount:    len(request.Trajectory.Points),
		Matches:       make([]Match, 0, len(request.Trajectory.Points)),
		Explanations: []Notice{
			{
				Code:    "four_dimensional_weather_alignment",
				Message: "Weather samples were compared with trajectory points by horizontal position, altitude, and time.",
			},
			{
				Code:    "weather_context_only",
				Message: "Alignment expresses contextual proximity and does not prove pilot intent, controller intent, rerouting reason, or maneuver cause.",
			},
		},
		InputFingerprint: inputFingerprint(request.Trajectory, request.Weather, request.Trust, request.Policy),
		GeneratedAt:      generatedAt,
	}

	if request.Trust.Decision == weathertrust.DecisionBlocked || !request.Trust.Usable {
		result.Status = StatusUnavailable
		for sequence, point := range request.Trajectory.Points {
			result.Matches = append(result.Matches, unmatched(
				sequence,
				point,
				request.Policy,
				"weather_trust_blocked",
				"Weather Trust Gate blocked analytical use of the weather evidence.",
			))
		}
		result.UnmatchedCount = result.PointCount
		result.Limitations = []Notice{{
			Code:    "weather_trust_blocked",
			Message: "Weather alignment is unavailable because the Weather Trust Gate blocked the evidence.",
		}}
		return validateAndClone(result)
	}

	for sequence, point := range request.Trajectory.Points {
		match := alignPoint(sequence, point, request.Weather.Samples, request.Trust, request.Policy)
		result.Matches = append(result.Matches, match)
		if match.Status == MatchStatusAligned {
			result.AlignedCount++
		} else {
			result.UnmatchedCount++
		}
	}

	if result.PointCount > 0 {
		result.CoverageRatio = float64(result.AlignedCount) / float64(result.PointCount)
	}
	switch {
	case result.AlignedCount == 0:
		result.Status = StatusUnavailable
		result.Limitations = append(result.Limitations, Notice{
			Code:    "weather_alignment_no_matches",
			Message: "No trajectory point satisfied all four-dimensional weather alignment boundaries.",
		})
	case result.AlignedCount < result.PointCount:
		result.Status = StatusLimited
		result.Limitations = append(result.Limitations, Notice{
			Code:    "weather_alignment_partial_coverage",
			Message: "Only part of the trajectory could be aligned with trusted weather evidence.",
		})
	default:
		result.Status = StatusComplete
	}

	if request.Trust.Decision == weathertrust.DecisionLimited {
		result.Limitations = append(result.Limitations, Notice{
			Code:    "weather_trust_limited",
			Message: "The Weather Trust Gate permits only limited use of the weather evidence.",
		})
		if result.Status == StatusComplete {
			result.Status = StatusLimited
			result.Limitations = append(result.Limitations, Notice{
				Code:    "full_geometric_coverage_with_limited_trust",
				Message: "All trajectory points were geometrically matched, but the upstream weather trust decision remains limited.",
			})
		}
	}

	for _, match := range result.Matches {
		result.Limitations = append(result.Limitations, match.Limitations...)
	}
	result.Limitations = normalizeNotices(result.Limitations)
	return validateAndClone(result)
}

func validateTrajectory(flightTrajectory trajectory.FlightTrajectory, asOfTime time.Time) error {
	if strings.TrimSpace(flightTrajectory.ID) == "" || asOfTime.IsZero() {
		return fmt.Errorf("trajectory identifier and as-of time are required")
	}
	for sequence, point := range flightTrajectory.Points {
		if !finite(point.Latitude) || point.Latitude < -90 || point.Latitude > 90 ||
			!finite(point.Longitude) || point.Longitude < -180 || point.Longitude > 180 ||
			point.ObservedAt.IsZero() || point.ObservedAt.After(asOfTime) {
			return fmt.Errorf("trajectory point %d is invalid or leaks future evidence", sequence)
		}
	}
	return nil
}

func alignPoint(
	sequence int,
	point trajectory.TrackPoint4D,
	samples []weathercontract.Sample,
	trust weathertrust.Result,
	policy Policy,
) Match {
	altitudeMeters, altitudeBasis := resolvePointAltitude(point)
	if altitudeBasis == AltitudeBasisUnavailable {
		return unmatched(
			sequence,
			point,
			policy,
			"trajectory_altitude_unavailable",
			"Trajectory point altitude is unavailable, so four-dimensional alignment cannot be completed.",
		)
	}

	bestScore := -1.0
	var bestMatch Match
	rejectionNotices := make([]Notice, 0)
	for _, sample := range samples {
		candidate, accepted, notice := evaluateCandidate(
			sequence,
			point,
			altitudeMeters,
			altitudeBasis,
			sample,
			trust,
			policy,
		)
		if !accepted {
			rejectionNotices = append(rejectionNotices, notice)
			continue
		}
		if candidate.Score > bestScore ||
			(candidate.Score == bestScore && earlierSample(candidate, bestMatch)) {
			bestScore = candidate.Score
			bestMatch = candidate
		}
	}

	if bestScore < 0 {
		match := unmatched(
			sequence,
			point,
			policy,
			"weather_sample_not_within_alignment_boundary",
			"No trusted weather sample satisfies the point's horizontal, temporal, vertical, and usage-scope boundaries.",
		)
		match.AltitudeBasis = altitudeBasis
		match.AltitudeMeters = cloneFloat64(&altitudeMeters)
		match.Limitations = normalizeNotices(append(match.Limitations, rejectionNotices...))
		return match
	}
	return bestMatch
}

func evaluateCandidate(
	sequence int,
	point trajectory.TrackPoint4D,
	pointAltitudeMeters float64,
	altitudeBasis AltitudeBasis,
	sample weathercontract.Sample,
	trust weathertrust.Result,
	policy Policy,
) (Match, bool, Notice) {
	if !scopeAllowsPoint(point, trust.AllowedScopes) {
		return Match{}, false, Notice{
			Code:    "weather_usage_scope_not_allowed",
			Message: "Weather Trust Gate does not allow this weather evidence for the trajectory point's ground or airborne context.",
		}
	}

	horizontalDistance := horizontalDistanceKilometers(
		point.Latitude,
		point.Longitude,
		sample.Position.Latitude,
		sample.Position.Longitude,
	)
	if horizontalDistance > policy.MaximumHorizontalDistanceKilometers {
		return Match{}, false, Notice{
			Code:    "weather_horizontal_boundary_exceeded",
			Message: "Weather sample is outside the maximum horizontal alignment distance.",
		}
	}

	temporalDistance := absoluteDuration(point.ObservedAt, sample.ValidAt)
	if temporalDistance > policy.MaximumTemporalDistance {
		return Match{}, false, Notice{
			Code:    "weather_temporal_boundary_exceeded",
			Message: "Weather sample is outside the maximum temporal alignment distance.",
		}
	}

	weatherAltitude, verticalAccepted, verticalNotice := weatherAltitudeForPoint(point, sample)
	if !verticalAccepted {
		return Match{}, false, verticalNotice
	}
	verticalDistance := math.Abs(pointAltitudeMeters - weatherAltitude)
	if verticalDistance > policy.MaximumVerticalDistanceMeters {
		return Match{}, false, Notice{
			Code:    "weather_vertical_boundary_exceeded",
			Message: "Weather sample is outside the maximum vertical alignment distance.",
		}
	}

	horizontalScore := 1 - horizontalDistance/policy.MaximumHorizontalDistanceKilometers
	temporalScore := 1 - float64(temporalDistance)/float64(policy.MaximumTemporalDistance)
	verticalScore := 1 - verticalDistance/policy.MaximumVerticalDistanceMeters
	components := policy.components(horizontalScore, temporalScore, verticalScore)
	score := weightedScore(components)
	if score < policy.MinimumMatchScore {
		return Match{}, false, Notice{
			Code:    "weather_alignment_score_below_minimum",
			Message: "Weather sample satisfies individual distance boundaries but its combined alignment score is below the minimum.",
		}
	}

	sampleSequence := sample.Sequence
	weatherValidAt := sample.ValidAt.UTC()
	horizontalDistanceCopy := horizontalDistance
	temporalDistanceCopy := temporalDistance
	verticalDistanceCopy := verticalDistance
	altitudeCopy := pointAltitudeMeters
	return Match{
		TrajectoryPointSequence:      sequence,
		TrajectoryPointID:            strings.TrimSpace(point.ID),
		TrajectoryObservedAt:         point.ObservedAt.UTC(),
		WeatherSampleSequence:        &sampleSequence,
		WeatherValidAt:               &weatherValidAt,
		Status:                       MatchStatusAligned,
		AltitudeBasis:                altitudeBasis,
		AltitudeMeters:               &altitudeCopy,
		HorizontalDistanceKilometers: &horizontalDistanceCopy,
		TemporalDistance:             &temporalDistanceCopy,
		VerticalDistanceMeters:       &verticalDistanceCopy,
		Score:                        score,
		Components:                   components,
	}, true, Notice{}
}

func scopeAllowsPoint(point trajectory.TrackPoint4D, scopes []weathertrust.UsageScope) bool {
	if point.OnGround {
		return hasScope(scopes, weathertrust.UsageScopeSurfaceContext) ||
			hasScope(scopes, weathertrust.UsageScopeTrajectoryContext)
	}
	return hasScope(scopes, weathertrust.UsageScopeTrajectoryContext)
}

func weatherAltitudeForPoint(
	point trajectory.TrackPoint4D,
	sample weathercontract.Sample,
) (float64, bool, Notice) {
	switch sample.Position.VerticalReference {
	case weathercontract.VerticalReferenceSurface:
		if !point.OnGround {
			return 0, false, Notice{
				Code:    "surface_weather_not_airborne_weather",
				Message: "Surface weather cannot be aligned to an airborne trajectory point.",
			}
		}
		return 0, true, Notice{}
	case weathercontract.VerticalReferenceMeanSeaLevel,
		weathercontract.VerticalReferencePressureLevel:
		if sample.Position.AltitudeMeters == nil {
			return 0, false, Notice{
				Code:    "weather_altitude_unavailable",
				Message: "Weather sample lacks altitude required for airborne alignment.",
			}
		}
		return *sample.Position.AltitudeMeters, true, Notice{}
	default:
		return 0, false, Notice{
			Code:    "weather_vertical_reference_unusable",
			Message: "Weather sample vertical reference cannot support four-dimensional alignment.",
		}
	}
}

func resolvePointAltitude(point trajectory.TrackPoint4D) (float64, AltitudeBasis) {
	if point.OnGround {
		return 0, AltitudeBasisGround
	}
	if flightstate.ResolveAltitudeStatus(point.GeometricAltitudeM, point.GeometricAltitudeStatus) ==
		flightstate.AltitudeStatusObserved && finite(point.GeometricAltitudeM) {
		return point.GeometricAltitudeM, AltitudeBasisGeometric
	}
	if flightstate.ResolveAltitudeStatus(point.BarometricAltitudeM, point.BarometricAltitudeStatus) ==
		flightstate.AltitudeStatusObserved && finite(point.BarometricAltitudeM) {
		return point.BarometricAltitudeM, AltitudeBasisBarometric
	}
	return 0, AltitudeBasisUnavailable
}

func unmatched(
	sequence int,
	point trajectory.TrackPoint4D,
	policy Policy,
	code string,
	message string,
) Match {
	return Match{
		TrajectoryPointSequence: sequence,
		TrajectoryPointID:       strings.TrimSpace(point.ID),
		TrajectoryObservedAt:    point.ObservedAt.UTC(),
		Status:                  MatchStatusUnmatched,
		AltitudeBasis:           AltitudeBasisUnavailable,
		Score:                   0,
		Components:              policy.components(0, 0, 0),
		Limitations:             []Notice{{Code: code, Message: message}},
	}
}

func earlierSample(left, right Match) bool {
	if left.WeatherSampleSequence == nil {
		return false
	}
	if right.WeatherSampleSequence == nil {
		return true
	}
	return *left.WeatherSampleSequence < *right.WeatherSampleSequence
}

func hasScope(scopes []weathertrust.UsageScope, target weathertrust.UsageScope) bool {
	for _, scope := range scopes {
		if scope == target {
			return true
		}
	}
	return false
}

func validateAndClone(result Result) (Result, error) {
	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrResultInvalid, err)
	}
	return result.Clone(), nil
}
