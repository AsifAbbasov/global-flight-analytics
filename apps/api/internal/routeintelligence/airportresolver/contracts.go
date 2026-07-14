package airportresolver

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

const (
	CatalogVersion  = "airport-candidate-catalog-v1"
	ResolverVersion = "airport-candidate-resolver-v1"

	DefaultMaximumDistanceKM = 120.0
	DefaultMaximumCandidates = 5
	MaximumCandidateLimit    = 100
)

type ExclusionReason string

const (
	ExclusionReasonInvalidICAOCode    ExclusionReason = "invalid_icao_code"
	ExclusionReasonInvalidIATACode    ExclusionReason = "invalid_iata_code"
	ExclusionReasonMissingName        ExclusionReason = "missing_name"
	ExclusionReasonInvalidCoordinates ExclusionReason = "invalid_coordinates"
	ExclusionReasonInvalidElevation   ExclusionReason = "invalid_elevation"
	ExclusionReasonDuplicateICAOCode  ExclusionReason = "duplicate_icao_code"
)

type ExclusionSummary struct {
	Reason ExclusionReason
	Count  int
}

type CatalogBuildReport struct {
	Version       string
	InputCount    int
	AcceptedCount int
	ExcludedCount int
	Exclusions    []ExclusionSummary
	Fingerprint   string
}

func (report CatalogBuildReport) Clone() CatalogBuildReport {
	cloned := report
	cloned.Exclusions = append(
		[]ExclusionSummary(nil),
		report.Exclusions...,
	)

	return cloned
}

type Point struct {
	Latitude  float64
	Longitude float64
}

type Query struct {
	Role  routecontract.EndpointRole
	Point Point
}

type Candidate struct {
	Rank           int
	Airport        routecontract.AirportReference
	DistanceKM     float64
	ProximityScore float64
}

type Result struct {
	Version                 string
	Role                    routecontract.EndpointRole
	Point                   Point
	MaximumDistanceKM       float64
	MaximumCandidates       int
	CatalogVersion          string
	CatalogFingerprint      string
	CatalogAirportCount     int
	EligibleCandidateCount  int
	FilteredByRadiusCount   int
	TruncatedCandidateCount int
	Candidates              []Candidate
	InputFingerprint        string
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Candidates = append(
		[]Candidate(nil),
		result.Candidates...,
	)

	return cloned
}

type Config struct {
	Catalog           *Catalog
	MaximumDistanceKM float64
	MaximumCandidates int
}
