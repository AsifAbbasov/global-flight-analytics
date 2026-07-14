package flightphase

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type Phase string

const (
	PhaseUnknown Phase = "unknown"
	PhaseGround  Phase = "ground"
	PhaseTakeoff Phase = "takeoff"
	PhaseClimb   Phase = "climb"
	PhaseCruise  Phase = "cruise"
	PhaseDescent Phase = "descent"
	PhaseLanding Phase = "landing"
)

func (phase Phase) IsKnown() bool {
	switch phase {
	case PhaseUnknown,
		PhaseGround,
		PhaseTakeoff,
		PhaseClimb,
		PhaseCruise,
		PhaseDescent,
		PhaseLanding:
		return true
	default:
		return false
	}
}

type ReasonCode string

const (
	ReasonOnGroundFlag              ReasonCode = "on_ground_flag"
	ReasonLowAltitudeAndSpeed       ReasonCode = "low_altitude_and_speed"
	ReasonPositiveVerticalRate      ReasonCode = "positive_vertical_rate"
	ReasonNegativeVerticalRate      ReasonCode = "negative_vertical_rate"
	ReasonLowAltitudeDeparture      ReasonCode = "low_altitude_departure"
	ReasonLowAltitudeArrival        ReasonCode = "low_altitude_arrival"
	ReasonHighAltitudeStableLevel   ReasonCode = "high_altitude_stable_level"
	ReasonAltitudeUnavailable       ReasonCode = "altitude_unavailable"
	ReasonVerticalRateUnavailable   ReasonCode = "vertical_rate_unavailable"
	ReasonPhaseThresholdsUnresolved ReasonCode = "phase_thresholds_unresolved"
)

type Notice struct {
	Code    string
	Message string
}

type PointEvidence struct {
	PointID               string
	ObservedAt            time.Time
	Phase                 Phase
	Confidence            float64
	AltitudeM             float64
	AltitudeAvailable     bool
	VelocityMPS           float64
	VelocityAvailable     bool
	VerticalRateMPS       float64
	VerticalRateAvailable bool
	OnGround              bool
	Reasons               []ReasonCode
}

type Segment struct {
	Phase      Phase
	StartTime  time.Time
	EndTime    time.Time
	PointCount int
	Confidence float64
	Reasons    []ReasonCode
}

type Result struct {
	AlgorithmVersion     string
	CurrentPhase         Phase
	InputPointCount      int
	ClassifiedPointCount int
	ExcludedPointCount   int
	Points               []PointEvidence
	Segments             []Segment
	Limitations          []Notice
}

func (result Result) Clone() Result {
	clone := result
	clone.Points = make([]PointEvidence, len(result.Points))
	for index, point := range result.Points {
		clone.Points[index] = point
		clone.Points[index].Reasons = append(
			[]ReasonCode(nil),
			point.Reasons...,
		)
	}

	clone.Segments = make([]Segment, len(result.Segments))
	for index, segment := range result.Segments {
		clone.Segments[index] = segment
		clone.Segments[index].Reasons = append(
			[]ReasonCode(nil),
			segment.Reasons...,
		)
	}

	clone.Limitations = append(
		[]Notice(nil),
		result.Limitations...,
	)

	return clone
}

var (
	ErrAlgorithmVersionInvalid = errors.New(
		"flight phase algorithm version is invalid",
	)
	ErrCurrentPhaseInvalid = errors.New(
		"current flight phase is invalid",
	)
	ErrPointCountsInvalid = errors.New(
		"flight phase point counts are inconsistent",
	)
	ErrPointPhaseInvalid = errors.New(
		"point flight phase is invalid",
	)
	ErrPointConfidenceInvalid = errors.New(
		"point flight phase confidence must be finite and between zero and one",
	)
	ErrPointTimeMissing = errors.New(
		"classified flight phase point time is required",
	)
	ErrPointOrderInvalid = errors.New(
		"classified flight phase points must be ordered by observation time",
	)
	ErrPointReasonRequired = errors.New(
		"classified flight phase point requires at least one reason",
	)
	ErrSegmentInvalid = errors.New(
		"flight phase segment is invalid",
	)
	ErrSegmentCoverageInvalid = errors.New(
		"flight phase segments do not cover classified points",
	)
	ErrNoticeInvalid = errors.New(
		"flight phase notice requires code and message",
	)
)

func (result Result) Validate() error {
	if result.AlgorithmVersion != AlgorithmVersion {
		return fmt.Errorf(
			"%w: %q",
			ErrAlgorithmVersionInvalid,
			result.AlgorithmVersion,
		)
	}
	if !result.CurrentPhase.IsKnown() {
		return fmt.Errorf(
			"%w: %q",
			ErrCurrentPhaseInvalid,
			result.CurrentPhase,
		)
	}
	if result.InputPointCount < 0 ||
		result.ClassifiedPointCount < 0 ||
		result.ExcludedPointCount < 0 ||
		result.ClassifiedPointCount != len(result.Points) ||
		result.InputPointCount !=
			result.ClassifiedPointCount+
				result.ExcludedPointCount {
		return ErrPointCountsInvalid
	}

	for index, point := range result.Points {
		if !point.Phase.IsKnown() {
			return fmt.Errorf(
				"%w: index=%d phase=%q",
				ErrPointPhaseInvalid,
				index,
				point.Phase,
			)
		}
		if math.IsNaN(point.Confidence) ||
			math.IsInf(point.Confidence, 0) ||
			point.Confidence < 0 ||
			point.Confidence > 1 {
			return fmt.Errorf(
				"%w: index=%d confidence=%f",
				ErrPointConfidenceInvalid,
				index,
				point.Confidence,
			)
		}
		if point.ObservedAt.IsZero() {
			return fmt.Errorf(
				"%w: index=%d",
				ErrPointTimeMissing,
				index,
			)
		}
		if index > 0 &&
			point.ObservedAt.Before(
				result.Points[index-1].ObservedAt,
			) {
			return fmt.Errorf(
				"%w: index=%d",
				ErrPointOrderInvalid,
				index,
			)
		}
		if len(point.Reasons) == 0 {
			return fmt.Errorf(
				"%w: index=%d",
				ErrPointReasonRequired,
				index,
			)
		}
	}

	coveredPointCount := 0
	for index, segment := range result.Segments {
		if !segment.Phase.IsKnown() ||
			segment.PointCount <= 0 ||
			segment.StartTime.IsZero() ||
			segment.EndTime.IsZero() ||
			segment.EndTime.Before(segment.StartTime) ||
			math.IsNaN(segment.Confidence) ||
			math.IsInf(segment.Confidence, 0) ||
			segment.Confidence < 0 ||
			segment.Confidence > 1 ||
			len(segment.Reasons) == 0 {
			return fmt.Errorf(
				"%w: index=%d",
				ErrSegmentInvalid,
				index,
			)
		}

		coveredPointCount += segment.PointCount
	}
	if coveredPointCount != result.ClassifiedPointCount {
		return fmt.Errorf(
			"%w: covered=%d classified=%d",
			ErrSegmentCoverageInvalid,
			coveredPointCount,
			result.ClassifiedPointCount,
		)
	}

	for index, limitation := range result.Limitations {
		if strings.TrimSpace(limitation.Code) == "" ||
			strings.TrimSpace(limitation.Message) == "" {
			return fmt.Errorf(
				"%w: index=%d",
				ErrNoticeInvalid,
				index,
			)
		}
	}

	if len(result.Points) == 0 {
		if result.CurrentPhase != PhaseUnknown ||
			len(result.Segments) != 0 {
			return ErrSegmentCoverageInvalid
		}
		return nil
	}

	if result.CurrentPhase !=
		result.Points[len(result.Points)-1].Phase {
		return ErrCurrentPhaseInvalid
	}

	return nil
}
