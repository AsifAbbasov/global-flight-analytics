package endpointevidence

import "errors"

var (
	ErrInvalidMinimumSelectionScore = errors.New(
		"route endpoint minimum selection score must be finite and between zero and one",
	)
	ErrInvalidMinimumCandidateScoreGap = errors.New(
		"route endpoint minimum candidate score gap must be finite and between zero and one",
	)
	ErrInvalidCandidateResult = errors.New(
		"route endpoint airport candidate result is invalid",
	)
	ErrObservedAtRequired = errors.New(
		"route endpoint observed-at time is required",
	)
	ErrObservedAtNotUTC = errors.New(
		"route endpoint observed-at time must use UTC",
	)
	ErrInvalidTrajectoryQuality = errors.New(
		"route endpoint trajectory quality must be finite and between zero and one",
	)
	ErrInvalidSegmentStatus = errors.New(
		"route endpoint segment status must be observed, interpolated, or estimated",
	)
	ErrInvalidSegmentPointCount = errors.New(
		"route endpoint segment point count must be positive",
	)
	ErrInvalidCoverageGapCount = errors.New(
		"route endpoint coverage gap count must not be negative",
	)
)
