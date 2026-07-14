package flightphase

import (
	"fmt"
	"math"
	"sort"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const (
	LimitationCodeNoTrajectoryPoints = "flight_phase_points_unavailable"
	LimitationCodeZeroTimeExcluded   = "flight_phase_zero_time_points_excluded"
	LimitationCodeInputReordered     = "flight_phase_points_reordered"
	LimitationCodeDuplicateTime      = "flight_phase_duplicate_timestamps"
	LimitationCodeGeometricFallback  = "flight_phase_geometric_altitude_fallback"
	LimitationCodeUnknownPoints      = "flight_phase_unknown_points"
)

type Detector struct {
	config Config
}

func New(config Config) (*Detector, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate basic flight phase config: %w",
			err,
		)
	}

	return &Detector{
		config: config,
	}, nil
}

func NewDefault() *Detector {
	detector, err := New(DefaultConfig())
	if err != nil {
		panic(
			fmt.Sprintf(
				"default basic flight phase config is invalid: %v",
				err,
			),
		)
	}

	return detector
}

type orderedPoint struct {
	point         trajectory.TrackPoint4D
	originalIndex int
}

func (detector *Detector) Detect(
	item trajectory.FlightTrajectory,
) (Result, error) {
	result := Result{
		AlgorithmVersion: AlgorithmVersion,
		CurrentPhase:     PhaseUnknown,
		InputPointCount:  len(item.Points),
	}

	if len(item.Points) == 0 {
		result.Limitations = []Notice{
			{
				Code:    LimitationCodeNoTrajectoryPoints,
				Message: "No trajectory points were available for basic flight-phase detection.",
			},
		}
		if err := result.Validate(); err != nil {
			return Result{}, err
		}
		return result, nil
	}

	ordered := make(
		[]orderedPoint,
		0,
		len(item.Points),
	)
	zeroTimeCount := 0
	for index, point := range item.Points {
		if point.ObservedAt.IsZero() {
			zeroTimeCount++
			continue
		}

		point.ObservedAt = point.ObservedAt.UTC()
		ordered = append(
			ordered,
			orderedPoint{
				point:         point,
				originalIndex: index,
			},
		)
	}

	reordered := false
	sort.SliceStable(
		ordered,
		func(left int, right int) bool {
			leftTime := ordered[left].point.ObservedAt
			rightTime := ordered[right].point.ObservedAt
			if leftTime.Equal(rightTime) {
				return ordered[left].originalIndex <
					ordered[right].originalIndex
			}
			return leftTime.Before(rightTime)
		},
	)
	for index, item := range ordered {
		if item.originalIndex != index {
			reordered = true
			break
		}
	}

	duplicateTimeCount := 0
	for index := 1; index < len(ordered); index++ {
		if ordered[index].point.ObservedAt.Equal(
			ordered[index-1].point.ObservedAt,
		) {
			duplicateTimeCount++
		}
	}

	result.ExcludedPointCount = zeroTimeCount
	result.ClassifiedPointCount = len(ordered)
	result.Points = make(
		[]PointEvidence,
		0,
		len(ordered),
	)

	geometricFallbackCount := 0
	unknownPointCount := 0
	for _, value := range ordered {
		evidence, geometricFallback :=
			detector.classifyPoint(value.point)
		if geometricFallback {
			geometricFallbackCount++
		}
		if evidence.Phase == PhaseUnknown {
			unknownPointCount++
		}
		result.Points = append(
			result.Points,
			evidence,
		)
	}

	result.Segments = buildSegments(result.Points)
	if len(result.Points) > 0 {
		result.CurrentPhase =
			result.Points[len(result.Points)-1].Phase
	}

	if zeroTimeCount > 0 {
		result.Limitations = append(
			result.Limitations,
			Notice{
				Code: LimitationCodeZeroTimeExcluded,
				Message: fmt.Sprintf(
					"%d trajectory points without observation time were excluded.",
					zeroTimeCount,
				),
			},
		)
	}
	if reordered {
		result.Limitations = append(
			result.Limitations,
			Notice{
				Code:    LimitationCodeInputReordered,
				Message: "Trajectory points were reordered by observation time before phase detection.",
			},
		)
	}
	if duplicateTimeCount > 0 {
		result.Limitations = append(
			result.Limitations,
			Notice{
				Code: LimitationCodeDuplicateTime,
				Message: fmt.Sprintf(
					"%d duplicate observation timestamps were retained in stable input order.",
					duplicateTimeCount,
				),
			},
		)
	}
	if geometricFallbackCount > 0 {
		result.Limitations = append(
			result.Limitations,
			Notice{
				Code: LimitationCodeGeometricFallback,
				Message: fmt.Sprintf(
					"Geometric altitude was used for %d classified points where barometric altitude was unavailable.",
					geometricFallbackCount,
				),
			},
		)
	}
	if unknownPointCount > 0 {
		result.Limitations = append(
			result.Limitations,
			Notice{
				Code: LimitationCodeUnknownPoints,
				Message: fmt.Sprintf(
					"%d classified points did not satisfy a supported phase rule and remain unknown.",
					unknownPointCount,
				),
			},
		)
	}

	if len(ordered) == 0 {
		result.Limitations = append(
			result.Limitations,
			Notice{
				Code:    LimitationCodeNoTrajectoryPoints,
				Message: "No trajectory points with usable observation time were available for basic flight-phase detection.",
			},
		)
	}

	if err := result.Validate(); err != nil {
		return Result{},
			fmt.Errorf(
				"validate basic flight phase result: %w",
				err,
			)
	}

	return result.Clone(), nil
}

func (detector *Detector) classifyPoint(
	point trajectory.TrackPoint4D,
) (PointEvidence, bool) {
	altitude, altitudeAvailable, geometricFallback :=
		resolvePointAltitude(point)
	velocityAvailable := finite(point.VelocityMPS) &&
		point.VelocityMPS >= 0
	verticalRateAvailable := finite(point.VerticalRateMPS)

	evidence := PointEvidence{
		PointID:               point.ID,
		ObservedAt:            point.ObservedAt.UTC(),
		Phase:                 PhaseUnknown,
		Confidence:            0,
		AltitudeM:             altitude,
		AltitudeAvailable:     altitudeAvailable,
		VelocityMPS:           point.VelocityMPS,
		VelocityAvailable:     velocityAvailable,
		VerticalRateMPS:       point.VerticalRateMPS,
		VerticalRateAvailable: verticalRateAvailable,
		OnGround:              point.OnGround,
	}

	if point.OnGround {
		evidence.Phase = PhaseGround
		evidence.Confidence = 1
		evidence.Reasons = []ReasonCode{
			ReasonOnGroundFlag,
		}
		return evidence, geometricFallback
	}

	if altitudeAvailable &&
		velocityAvailable &&
		altitude <= detector.config.GroundMaximumAltitudeM &&
		point.VelocityMPS <=
			detector.config.GroundMaximumSpeedMPS {
		evidence.Phase = PhaseGround
		evidence.Confidence = 0.85
		evidence.Reasons = []ReasonCode{
			ReasonLowAltitudeAndSpeed,
		}
		return evidence, geometricFallback
	}

	if !altitudeAvailable {
		evidence.Reasons = append(
			evidence.Reasons,
			ReasonAltitudeUnavailable,
		)
	}
	if !verticalRateAvailable {
		evidence.Reasons = append(
			evidence.Reasons,
			ReasonVerticalRateUnavailable,
		)
		return evidence, geometricFallback
	}

	switch {
	case point.VerticalRateMPS >=
		detector.config.ClimbMinimumVerticalRateMPS:
		if altitudeAvailable &&
			altitude <=
				detector.config.TakeoffMaximumAltitudeM {
			evidence.Phase = PhaseTakeoff
			evidence.Confidence = 0.9
			evidence.Reasons = []ReasonCode{
				ReasonPositiveVerticalRate,
				ReasonLowAltitudeDeparture,
			}
		} else {
			evidence.Phase = PhaseClimb
			evidence.Confidence = 0.8
			evidence.Reasons = []ReasonCode{
				ReasonPositiveVerticalRate,
			}
		}

	case point.VerticalRateMPS <=
		detector.config.DescentMaximumVerticalRateMPS:
		if altitudeAvailable &&
			altitude <=
				detector.config.LandingMaximumAltitudeM {
			evidence.Phase = PhaseLanding
			evidence.Confidence = 0.9
			evidence.Reasons = []ReasonCode{
				ReasonNegativeVerticalRate,
				ReasonLowAltitudeArrival,
			}
		} else {
			evidence.Phase = PhaseDescent
			evidence.Confidence = 0.8
			evidence.Reasons = []ReasonCode{
				ReasonNegativeVerticalRate,
			}
		}

	case altitudeAvailable &&
		altitude >= detector.config.CruiseMinimumAltitudeM &&
		math.Abs(point.VerticalRateMPS) <=
			detector.config.
				CruiseMaximumAbsoluteVerticalRateMPS:
		evidence.Phase = PhaseCruise
		evidence.Confidence = 0.75
		evidence.Reasons = []ReasonCode{
			ReasonHighAltitudeStableLevel,
		}

	default:
		evidence.Reasons = append(
			evidence.Reasons,
			ReasonPhaseThresholdsUnresolved,
		)
	}

	return evidence, geometricFallback
}

func resolvePointAltitude(
	point trajectory.TrackPoint4D,
) (float64, bool, bool) {
	barometricValue, barometricAvailable :=
		usableAltitude(
			point.BarometricAltitudeM,
			point.BarometricAltitudeStatus,
		)
	if barometricAvailable {
		return barometricValue, true, false
	}

	geometricValue, geometricAvailable :=
		usableAltitude(
			point.GeometricAltitudeM,
			point.GeometricAltitudeStatus,
		)
	if geometricAvailable {
		return geometricValue, true, true
	}

	return 0, false, false
}

func usableAltitude(
	value float64,
	status flightstate.AltitudeStatus,
) (float64, bool) {
	switch flightstate.ResolveAltitudeStatus(
		value,
		status,
	) {
	case flightstate.AltitudeStatusObserved:
		if !finite(value) {
			return 0, false
		}
		return value, true

	case flightstate.AltitudeStatusGround:
		return 0, true

	default:
		return 0, false
	}
}

func buildSegments(
	points []PointEvidence,
) []Segment {
	if len(points) == 0 {
		return nil
	}

	segments := make([]Segment, 0)
	start := 0

	for index := 1; index <= len(points); index++ {
		if index < len(points) &&
			points[index].Phase == points[start].Phase {
			continue
		}

		segments = append(
			segments,
			buildSegment(points[start:index]),
		)
		start = index
	}

	return segments
}

func buildSegment(
	points []PointEvidence,
) Segment {
	reasonSet := make(map[ReasonCode]struct{})
	confidenceTotal := 0.0

	for _, point := range points {
		confidenceTotal += point.Confidence
		for _, reason := range point.Reasons {
			reasonSet[reason] = struct{}{}
		}
	}

	reasons := make(
		[]ReasonCode,
		0,
		len(reasonSet),
	)
	for reason := range reasonSet {
		reasons = append(reasons, reason)
	}
	sort.SliceStable(
		reasons,
		func(left int, right int) bool {
			return reasons[left] < reasons[right]
		},
	)

	return Segment{
		Phase:      points[0].Phase,
		StartTime:  points[0].ObservedAt,
		EndTime:    points[len(points)-1].ObservedAt,
		PointCount: len(points),
		Confidence: confidenceTotal /
			float64(len(points)),
		Reasons: reasons,
	}
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}
