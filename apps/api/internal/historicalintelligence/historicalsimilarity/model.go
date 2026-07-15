package historicalsimilarity

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

const (
	Version            = "historical-trajectory-similarity-v1"
	FingerprintVersion = "historical-trajectory-similarity-fingerprint-v1"

	DefaultMinimumPointCount = 4
	DefaultSampleCount       = 16
	MaximumRankLimit         = 100
)

type Level string

const (
	LevelNone   Level = "none"
	LevelLow    Level = "low"
	LevelMedium Level = "medium"
	LevelHigh   Level = "high"
)

type ComponentName string

const (
	ComponentGeometry   ComponentName = "geometry"
	ComponentEndpoints  ComponentName = "endpoints"
	ComponentPathLength ComponentName = "path_length"
	ComponentDuration   ComponentName = "duration"
)

type Component struct {
	Name          ComponentName
	Score         float64
	Weight        float64
	ObservedValue float64
	Unit          string
}

type Notice struct {
	Code    string
	Message string
}

type Result struct {
	Version string

	ReferenceTrajectoryID string
	CandidateTrajectoryID string

	Score float64
	Level Level

	ReferencePointCount int
	CandidatePointCount int
	SampleCount         int

	MeanDistanceKM           float64
	MaximumDistanceKM        float64
	StartEndpointDistanceKM  float64
	EndEndpointDistanceKM    float64
	ReferencePathLengthKM    float64
	CandidatePathLengthKM    float64
	ReferenceDurationSeconds float64
	CandidateDurationSeconds float64

	Components  []Component
	Reasons     []string
	Limitations []Notice

	InputFingerprint string
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Components = append(
		[]Component(nil),
		result.Components...,
	)
	cloned.Reasons = append(
		[]string(nil),
		result.Reasons...,
	)
	cloned.Limitations = append(
		[]Notice(nil),
		result.Limitations...,
	)

	return cloned
}

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func (result Result) Validate() error {
	if result.Version != Version {
		return fmt.Errorf(
			"%w: version=%q",
			ErrResultInvalid,
			result.Version,
		)
	}
	if strings.TrimSpace(
		result.ReferenceTrajectoryID,
	) == "" ||
		strings.TrimSpace(
			result.CandidateTrajectoryID,
		) == "" ||
		result.ReferenceTrajectoryID ==
			result.CandidateTrajectoryID {
		return fmt.Errorf(
			"%w: trajectory identifiers",
			ErrResultInvalid,
		)
	}
	if !ratio(result.Score) ||
		result.Level != LevelForScore(
			result.Score,
		) {
		return fmt.Errorf(
			"%w: score or level",
			ErrResultInvalid,
		)
	}
	if result.ReferencePointCount < 2 ||
		result.CandidatePointCount < 2 ||
		result.SampleCount < 2 {
		return fmt.Errorf(
			"%w: point counts",
			ErrResultInvalid,
		)
	}

	for _, value := range []float64{
		result.MeanDistanceKM,
		result.MaximumDistanceKM,
		result.StartEndpointDistanceKM,
		result.EndEndpointDistanceKM,
		result.ReferencePathLengthKM,
		result.CandidatePathLengthKM,
		result.ReferenceDurationSeconds,
		result.CandidateDurationSeconds,
	} {
		if !finite(value) || value < 0 {
			return fmt.Errorf(
				"%w: non-negative measurement",
				ErrResultInvalid,
			)
		}
	}

	if len(result.Components) != 4 {
		return fmt.Errorf(
			"%w: component count",
			ErrResultInvalid,
		)
	}
	weightTotal := 0.0
	seen := make(map[ComponentName]struct{})
	for _, component := range result.Components {
		if _, exists := seen[component.Name]; exists {
			return fmt.Errorf(
				"%w: duplicate component",
				ErrResultInvalid,
			)
		}
		seen[component.Name] = struct{}{}

		if !ratio(component.Score) ||
			!finite(component.Weight) ||
			component.Weight < 0 ||
			!finite(component.ObservedValue) ||
			component.ObservedValue < 0 ||
			strings.TrimSpace(component.Unit) == "" {
			return fmt.Errorf(
				"%w: component",
				ErrResultInvalid,
			)
		}
		weightTotal += component.Weight
	}
	if math.Abs(weightTotal-1) > 1e-9 {
		return fmt.Errorf(
			"%w: component weights",
			ErrResultInvalid,
		)
	}
	if len(result.Reasons) == 0 {
		return fmt.Errorf(
			"%w: reasons",
			ErrResultInvalid,
		)
	}
	for _, reason := range result.Reasons {
		if strings.TrimSpace(reason) == "" {
			return fmt.Errorf(
				"%w: reason",
				ErrResultInvalid,
			)
		}
	}
	for _, limitation := range result.Limitations {
		if strings.TrimSpace(limitation.Code) == "" ||
			strings.TrimSpace(limitation.Message) == "" {
			return fmt.Errorf(
				"%w: limitation",
				ErrResultInvalid,
			)
		}
	}
	if !fingerprintPattern.MatchString(
		result.InputFingerprint,
	) {
		return fmt.Errorf(
			"%w: input fingerprint",
			ErrResultInvalid,
		)
	}

	return nil
}

func LevelForScore(score float64) Level {
	switch {
	case score >= 0.8:
		return LevelHigh
	case score >= 0.6:
		return LevelMedium
	case score > 0:
		return LevelLow
	default:
		return LevelNone
	}
}

func ratio(value float64) bool {
	return finite(value) &&
		value >= 0 &&
		value <= 1
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}
