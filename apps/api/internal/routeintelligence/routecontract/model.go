package routecontract

import "time"

const Version = "route-intelligence-contract-v1"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "route-intelligence-v1"

type RouteStatus string

const (
	RouteStatusUnavailable RouteStatus = "unavailable"
	RouteStatusPartial     RouteStatus = "partial"
	RouteStatusComplete    RouteStatus = "complete"
)

type ConfidenceLevel string

const (
	ConfidenceLevelNone   ConfidenceLevel = "none"
	ConfidenceLevelLow    ConfidenceLevel = "low"
	ConfidenceLevelMedium ConfidenceLevel = "medium"
	ConfidenceLevelHigh   ConfidenceLevel = "high"
)

type EndpointRole string

const (
	EndpointRoleOrigin      EndpointRole = "origin"
	EndpointRoleDestination EndpointRole = "destination"
)

type EvidenceType string

const (
	EvidenceTypeTrajectoryEndpointProximity EvidenceType = "trajectory_endpoint_proximity"
	EvidenceTypeGroundCycle                 EvidenceType = "ground_cycle"
	EvidenceTypeCallsignRouteToken          EvidenceType = "callsign_route_token"
	EvidenceTypeSourceFlightIdentity        EvidenceType = "source_flight_identity"
	EvidenceTypeAirportActivity             EvidenceType = "airport_activity"
	EvidenceTypeExternalReference           EvidenceType = "external_reference"
)

type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

type ValidationSeverity string

const (
	ValidationSeverityError   ValidationSeverity = "error"
	ValidationSeverityWarning ValidationSeverity = "warning"
)

type RouteWindow struct {
	StartTime time.Time
	EndTime   time.Time
	AsOfTime  time.Time
}

type AirportReference struct {
	ICAOCode   string
	IATACode   string
	Name       string
	City       string
	Country    string
	Latitude   float64
	Longitude  float64
	ElevationM float64
	Timezone   string
}

type EvidenceAttribute struct {
	Key   string
	Value string
}

type Evidence struct {
	Type          EvidenceType
	SourceName    string
	SourceVersion string
	Score         float64
	Weight        float64
	ObservedAt    time.Time
	Summary       string
	Attributes    []EvidenceAttribute
}

type ConfidenceReason struct {
	Code         string
	Message      string
	Contribution float64
}

type Confidence struct {
	Score         float64
	Level         ConfidenceLevel
	EvidenceCount int
	Reasons       []ConfidenceReason
}

type Limitation struct {
	Code    string
	Message string
	Scope   string
}

type EndpointInference struct {
	Role        EndpointRole
	Airport     AirportReference
	DistanceKM  float64
	Confidence  Confidence
	Evidence    []Evidence
	Limitations []Limitation
}

type RouteSummary struct {
	GreatCircleDistanceKM float64
	SameAirport           bool
}

type Provenance struct {
	ResolverVersion     string
	InputFingerprint    string
	TrajectoryUpdatedAt time.Time
	SourceNames         []string
}

type Result struct {
	SchemaVersion SchemaVersion
	Status        RouteStatus

	TrajectoryID string
	IdentityKey  string
	FlightID     string
	AircraftID   string
	ICAO24       string
	Callsign     string

	Window      RouteWindow
	Origin      *EndpointInference
	Destination *EndpointInference
	Summary     RouteSummary
	Confidence  Confidence
	Limitations []Limitation
	Provenance  Provenance
	GeneratedAt time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Origin = cloneEndpoint(result.Origin)
	cloned.Destination = cloneEndpoint(result.Destination)
	cloned.Confidence = cloneConfidence(result.Confidence)
	cloned.Limitations = cloneLimitations(result.Limitations)
	cloned.Provenance.SourceNames = append(
		[]string(nil),
		result.Provenance.SourceNames...,
	)

	return cloned
}

func cloneEndpoint(
	endpoint *EndpointInference,
) *EndpointInference {
	if endpoint == nil {
		return nil
	}

	cloned := *endpoint
	cloned.Confidence = cloneConfidence(endpoint.Confidence)
	cloned.Evidence = cloneEvidence(endpoint.Evidence)
	cloned.Limitations = cloneLimitations(
		endpoint.Limitations,
	)

	return &cloned
}

func cloneConfidence(
	confidence Confidence,
) Confidence {
	cloned := confidence
	cloned.Reasons = append(
		[]ConfidenceReason(nil),
		confidence.Reasons...,
	)

	return cloned
}

func cloneEvidence(
	items []Evidence,
) []Evidence {
	cloned := make(
		[]Evidence,
		0,
		len(items),
	)
	for _, item := range items {
		copied := item
		copied.Attributes = append(
			[]EvidenceAttribute(nil),
			item.Attributes...,
		)
		cloned = append(cloned, copied)
	}

	return cloned
}

func cloneLimitations(
	items []Limitation,
) []Limitation {
	return append(
		[]Limitation(nil),
		items...,
	)
}

func ConfidenceLevelForScore(
	score float64,
) ConfidenceLevel {
	switch {
	case score >= 0.8:
		return ConfidenceLevelHigh
	case score >= 0.6:
		return ConfidenceLevelMedium
	case score > 0:
		return ConfidenceLevelLow
	default:
		return ConfidenceLevelNone
	}
}
