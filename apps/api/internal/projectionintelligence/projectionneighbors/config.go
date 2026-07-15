package projectionneighbors

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
)

var (
	ErrSimilarityEngineRequired = errors.New(
		"historical similarity engine is required",
	)
	ErrSimilarityPolicyKeyRequired = errors.New(
		"historical similarity policy key is required",
	)
	ErrMinimumCurrentPointCountInvalid = errors.New(
		"minimum current point count must be at least two",
	)
	ErrMaximumCandidateCountInvalid = errors.New(
		"maximum candidate count must be greater than zero",
	)
	ErrSelectionLimitInvalid = errors.New(
		"selection limit must be between one and one hundred",
	)
	ErrMinimumSimilarityScoreInvalid = errors.New(
		"minimum similarity score must be finite and between zero and one",
	)
	ErrMaximumAnchorDistanceInvalid = errors.New(
		"maximum anchor distance must be finite and greater than zero",
	)
	ErrMaximumCandidateAgeInvalid = errors.New(
		"maximum candidate age must be non-negative",
	)
)

type SimilarityEngine interface {
	Compare(
		trajectory.FlightTrajectory,
		trajectory.FlightTrajectory,
	) (historicalsimilarity.Result, error)
}

type Config struct {
	SimilarityEngine    SimilarityEngine
	SimilarityPolicyKey string

	MinimumCurrentPointCount int
	MaximumCandidateCount    int
	SelectionLimit           int

	MinimumSimilarityScore  float64
	MaximumAnchorDistanceKM float64
	MaximumCandidateAge     time.Duration
}

func (config Config) Validate() error {
	if config.SimilarityEngine == nil {
		return ErrSimilarityEngineRequired
	}
	if strings.TrimSpace(
		config.SimilarityPolicyKey,
	) == "" {
		return ErrSimilarityPolicyKeyRequired
	}
	if config.MinimumCurrentPointCount < 2 {
		return fmt.Errorf(
			"%w: %d",
			ErrMinimumCurrentPointCountInvalid,
			config.MinimumCurrentPointCount,
		)
	}
	if config.MaximumCandidateCount < 1 {
		return fmt.Errorf(
			"%w: %d",
			ErrMaximumCandidateCountInvalid,
			config.MaximumCandidateCount,
		)
	}
	if config.SelectionLimit < 1 ||
		config.SelectionLimit >
			historicalsimilarity.MaximumRankLimit {
		return fmt.Errorf(
			"%w: %d",
			ErrSelectionLimitInvalid,
			config.SelectionLimit,
		)
	}
	if !unitInterval(
		config.MinimumSimilarityScore,
	) {
		return fmt.Errorf(
			"%w: %f",
			ErrMinimumSimilarityScoreInvalid,
			config.MinimumSimilarityScore,
		)
	}
	if !finite(
		config.MaximumAnchorDistanceKM,
	) ||
		config.MaximumAnchorDistanceKM <= 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrMaximumAnchorDistanceInvalid,
			config.MaximumAnchorDistanceKM,
		)
	}
	if config.MaximumCandidateAge < 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumCandidateAgeInvalid,
			config.MaximumCandidateAge,
		)
	}

	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}

func unitInterval(value float64) bool {
	return finite(value) &&
		value >= 0 &&
		value <= 1
}
