package endpointevidence

import (
	"math"
	"regexp"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/airportresolver"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func validateInput(input Input) error {
	result := input.Candidates
	if result.Version != airportresolver.ResolverVersion ||
		result.CatalogVersion != airportresolver.CatalogVersion ||
		(result.Role != routecontract.EndpointRoleOrigin &&
			result.Role != routecontract.EndpointRoleDestination) ||
		!fingerprintPattern.MatchString(
			result.CatalogFingerprint,
		) ||
		!fingerprintPattern.MatchString(
			result.InputFingerprint,
		) ||
		!validPoint(result.Point.Latitude, result.Point.Longitude) ||
		result.MaximumDistanceKM <= 0 ||
		math.IsNaN(result.MaximumDistanceKM) ||
		math.IsInf(result.MaximumDistanceKM, 0) ||
		result.MaximumCandidates < 1 ||
		result.MaximumCandidates >
			airportresolver.MaximumCandidateLimit ||
		result.CatalogAirportCount < 1 ||
		result.EligibleCandidateCount < 0 ||
		result.FilteredByRadiusCount < 0 ||
		result.TruncatedCandidateCount < 0 ||
		result.CatalogAirportCount !=
			result.EligibleCandidateCount+
				result.FilteredByRadiusCount ||
		result.EligibleCandidateCount !=
			len(result.Candidates)+
				result.TruncatedCandidateCount ||
		len(result.Candidates) >
			result.MaximumCandidates {
		return ErrInvalidCandidateResult
	}

	previousDistance := -1.0
	for index, candidate := range result.Candidates {
		if candidate.Rank != index+1 ||
			candidate.Airport.ICAOCode == "" ||
			candidate.DistanceKM < 0 ||
			math.IsNaN(candidate.DistanceKM) ||
			math.IsInf(candidate.DistanceKM, 0) ||
			!finiteRatio(candidate.ProximityScore) ||
			candidate.DistanceKM >
				result.MaximumDistanceKM ||
			candidate.DistanceKM < previousDistance {
			return ErrInvalidCandidateResult
		}
		previousDistance = candidate.DistanceKM
	}

	if input.ObservedAt.IsZero() {
		return ErrObservedAtRequired
	}
	if input.ObservedAt.Location() != time.UTC {
		return ErrObservedAtNotUTC
	}
	if !finiteRatio(input.TrajectoryQuality) {
		return ErrInvalidTrajectoryQuality
	}
	switch input.SegmentStatus {
	case trajectory.SegmentStatusObserved,
		trajectory.SegmentStatusInterpolated,
		trajectory.SegmentStatusEstimated:
	default:
		return ErrInvalidSegmentStatus
	}
	if input.SegmentPointCount < 1 {
		return ErrInvalidSegmentPointCount
	}
	if input.CoverageGapCount < 0 {
		return ErrInvalidCoverageGapCount
	}

	return nil
}

func validPoint(
	latitude float64,
	longitude float64,
) bool {
	return !math.IsNaN(latitude) &&
		!math.IsInf(latitude, 0) &&
		latitude >= -90 &&
		latitude <= 90 &&
		!math.IsNaN(longitude) &&
		!math.IsInf(longitude, 0) &&
		longitude >= -180 &&
		longitude <= 180
}
