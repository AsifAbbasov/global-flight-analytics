package historicalsimilarity

import "errors"

var (
	ErrEngineRequired = errors.New(
		"historical similarity engine is required",
	)
	ErrMinimumPointCountInvalid = errors.New(
		"historical similarity minimum point count must be between two and the maximum input point count",
	)
	ErrSampleCountInvalid = errors.New(
		"historical similarity sample count must be between two and the configured hard maximum",
	)
	ErrDistanceScaleInvalid = errors.New(
		"historical similarity distance score scales must be finite and positive",
	)
	ErrWeightInvalid = errors.New(
		"historical similarity weights must be finite, non-negative, and sum to one",
	)
	ErrReferenceNotComparable = errors.New(
		"historical similarity reference trajectory is not comparable",
	)
	ErrCandidateNotComparable = errors.New(
		"historical similarity candidate trajectory is not comparable",
	)
	ErrSameTrajectory = errors.New(
		"historical similarity requires two different trajectories",
	)
	ErrTrajectoryPointLimitExceeded = errors.New(
		"historical similarity trajectory point count exceeds the hard maximum",
	)
	ErrTrajectoryQualityInvalid = errors.New(
		"historical similarity trajectory quality evidence is invalid",
	)
	ErrTrajectorySegmentInvalid = errors.New(
		"historical similarity trajectory segment evidence is invalid",
	)
	ErrTrajectoryGapInvalid = errors.New(
		"historical similarity trajectory coverage-gap evidence is invalid",
	)
	ErrResultInvalid = errors.New(
		"historical similarity result is invalid",
	)
)
