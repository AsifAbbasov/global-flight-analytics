package endpointevidence

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/airportresolver"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

const (
	Version = "route-endpoint-evidence-v1"

	DefaultMinimumSelectionScore    = 0.60
	DefaultMinimumCandidateScoreGap = 0.05
)

type SelectionStatus string

const (
	SelectionStatusUnavailable  SelectionStatus = "unavailable"
	SelectionStatusInsufficient SelectionStatus = "insufficient"
	SelectionStatusAmbiguous    SelectionStatus = "ambiguous"
	SelectionStatusSelected     SelectionStatus = "selected"
)

type Config struct {
	MinimumSelectionScore    float64
	MinimumCandidateScoreGap float64
}

type Input struct {
	Candidates        airportresolver.Result
	ObservedAt        time.Time
	TrajectoryQuality float64
	SegmentStatus     trajectory.SegmentStatus
	SegmentPointCount int
	CoverageGapCount  int
}

type Result struct {
	Version                string
	Status                 SelectionStatus
	Role                   routecontract.EndpointRole
	Endpoint               *routecontract.EndpointInference
	CandidateCount         int
	SelectedCandidateRank  int
	SelectedCandidateScore float64
	RunnerUpCandidateScore float64
	CandidateScoreGap      float64
	InputFingerprint       string
	Limitations            []routecontract.Limitation
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Endpoint = cloneEndpoint(result.Endpoint)
	cloned.Limitations = append(
		[]routecontract.Limitation(nil),
		result.Limitations...,
	)

	return cloned
}

func cloneEndpoint(
	endpoint *routecontract.EndpointInference,
) *routecontract.EndpointInference {
	if endpoint == nil {
		return nil
	}

	cloned := *endpoint
	cloned.Confidence.Reasons = append(
		[]routecontract.ConfidenceReason(nil),
		endpoint.Confidence.Reasons...,
	)
	cloned.Evidence = make(
		[]routecontract.Evidence,
		0,
		len(endpoint.Evidence),
	)
	for _, item := range endpoint.Evidence {
		copied := item
		copied.Attributes = append(
			[]routecontract.EvidenceAttribute(nil),
			item.Attributes...,
		)
		cloned.Evidence = append(cloned.Evidence, copied)
	}
	cloned.Limitations = append(
		[]routecontract.Limitation(nil),
		endpoint.Limitations...,
	)

	return &cloned
}
