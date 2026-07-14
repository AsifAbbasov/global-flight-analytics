package airportresolver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"sort"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type Resolver struct {
	catalog           *Catalog
	maximumDistanceKM float64
	maximumCandidates int
}

func New(
	config Config,
) (*Resolver, error) {
	if config.Catalog == nil {
		return nil, ErrCatalogRequired
	}

	maximumDistanceKM := config.MaximumDistanceKM
	if maximumDistanceKM == 0 {
		maximumDistanceKM =
			DefaultMaximumDistanceKM
	}
	if maximumDistanceKM <= 0 ||
		math.IsNaN(maximumDistanceKM) ||
		math.IsInf(maximumDistanceKM, 0) {
		return nil, ErrInvalidMaximumDistance
	}

	maximumCandidates := config.MaximumCandidates
	if maximumCandidates == 0 {
		maximumCandidates =
			DefaultMaximumCandidates
	}
	if maximumCandidates < 1 ||
		maximumCandidates >
			MaximumCandidateLimit {
		return nil, ErrInvalidMaximumCandidates
	}

	return &Resolver{
		catalog:           config.Catalog,
		maximumDistanceKM: maximumDistanceKM,
		maximumCandidates: maximumCandidates,
	}, nil
}

func (resolver *Resolver) Resolve(
	ctx context.Context,
	query Query,
) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if query.Role != routecontract.EndpointRoleOrigin &&
		query.Role !=
			routecontract.EndpointRoleDestination {
		return Result{}, ErrInvalidEndpointRole
	}
	if !validLatitude(query.Point.Latitude) ||
		!validLongitude(query.Point.Longitude) {
		return Result{}, ErrInvalidPoint
	}

	normalizedPoint := Point{
		Latitude: normalizeSignedZero(
			query.Point.Latitude,
		),
		Longitude: normalizeSignedZero(
			query.Point.Longitude,
		),
	}

	eligible := make(
		[]Candidate,
		0,
		resolver.catalog.Size(),
	)
	filteredByRadiusCount := 0

	for index, airport := range resolver.catalog.airports {
		if index%256 == 0 {
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
		}

		distanceKM := haversineDistanceKM(
			normalizedPoint,
			Point{
				Latitude:  airport.Latitude,
				Longitude: airport.Longitude,
			},
		)
		if distanceKM >
			resolver.maximumDistanceKM {
			filteredByRadiusCount++
			continue
		}

		eligible = append(
			eligible,
			Candidate{
				Airport:    airport,
				DistanceKM: distanceKM,
				ProximityScore: clamp01(
					1 -
						distanceKM/
							resolver.
								maximumDistanceKM,
				),
			},
		)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	sort.SliceStable(
		eligible,
		func(left int, right int) bool {
			leftCandidate := eligible[left]
			rightCandidate := eligible[right]

			if leftCandidate.DistanceKM !=
				rightCandidate.DistanceKM {
				return leftCandidate.DistanceKM <
					rightCandidate.DistanceKM
			}
			if leftCandidate.Airport.ICAOCode !=
				rightCandidate.Airport.ICAOCode {
				return leftCandidate.Airport.ICAOCode <
					rightCandidate.Airport.ICAOCode
			}
			if leftCandidate.Airport.IATACode !=
				rightCandidate.Airport.IATACode {
				return leftCandidate.Airport.IATACode <
					rightCandidate.Airport.IATACode
			}

			return leftCandidate.Airport.Name <
				rightCandidate.Airport.Name
		},
	)

	eligibleCandidateCount := len(eligible)
	truncatedCandidateCount := 0
	if len(eligible) > resolver.maximumCandidates {
		truncatedCandidateCount =
			len(eligible) -
				resolver.maximumCandidates
		eligible = eligible[:resolver.maximumCandidates]
	}

	for index := range eligible {
		eligible[index].Rank = index + 1
	}

	inputFingerprint, err := resolverFingerprint(
		resolver.catalog.fingerprint,
		query.Role,
		normalizedPoint,
		resolver.maximumDistanceKM,
		resolver.maximumCandidates,
	)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Version:                 ResolverVersion,
		Role:                    query.Role,
		Point:                   normalizedPoint,
		MaximumDistanceKM:       resolver.maximumDistanceKM,
		MaximumCandidates:       resolver.maximumCandidates,
		CatalogVersion:          resolver.catalog.Version(),
		CatalogFingerprint:      resolver.catalog.fingerprint,
		CatalogAirportCount:     resolver.catalog.Size(),
		EligibleCandidateCount:  eligibleCandidateCount,
		FilteredByRadiusCount:   filteredByRadiusCount,
		TruncatedCandidateCount: truncatedCandidateCount,
		Candidates: append(
			[]Candidate(nil),
			eligible...,
		),
		InputFingerprint: inputFingerprint,
	}.Clone(), nil
}

func resolverFingerprint(
	catalogFingerprint string,
	role routecontract.EndpointRole,
	point Point,
	maximumDistanceKM float64,
	maximumCandidates int,
) (string, error) {
	input := struct {
		CatalogFingerprint string                     `json:"catalog_fingerprint"`
		Role               routecontract.EndpointRole `json:"role"`
		Point              Point                      `json:"point"`
		MaximumDistanceKM  float64                    `json:"maximum_distance_km"`
		MaximumCandidates  int                        `json:"maximum_candidates"`
	}{
		CatalogFingerprint: catalogFingerprint,
		Role:               role,
		Point:              point,
		MaximumDistanceKM:  maximumDistanceKM,
		MaximumCandidates:  maximumCandidates,
	}

	encoded, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(encoded)

	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
