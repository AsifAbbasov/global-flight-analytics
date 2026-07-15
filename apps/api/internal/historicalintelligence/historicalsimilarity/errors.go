package historicalsimilarity

import "errors"

var (
	ErrMinimumPointCountInvalid = errors.New(
		"historical similarity minimum point count must be at least two",
	)
	ErrSampleCountInvalid = errors.New(
		"historical similarity sample count must be at least two",
	)
	ErrDistanceThresholdInvalid = errors.New(
		"historical similarity distance threshold must be finite and positive",
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
	ErrRankLimitInvalid = errors.New(
		"historical similarity rank limit must be between one and one hundred",
	)
	ErrResultInvalid = errors.New(
		"historical similarity result is invalid",
	)
)
