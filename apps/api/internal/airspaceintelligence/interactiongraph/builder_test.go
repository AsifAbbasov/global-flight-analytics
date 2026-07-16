package interactiongraph

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestBuildCreatesDeterministicCompleteGraph(t *testing.T) {
	t.Parallel()

	request := completeRequest()
	result, err := Build(request)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if result.Status != ResultStatusComplete {
		t.Fatalf("Status = %q, want %q", result.Status, ResultStatusComplete)
	}
	if len(result.Nodes) != 3 || len(result.Edges) != 2 {
		t.Fatalf(
			"graph size = %d nodes/%d edges, want 3/2",
			len(result.Nodes),
			len(result.Edges),
		)
	}
	if result.Nodes[0].ID != "node-a" ||
		result.Nodes[1].ID != "node-b" ||
		result.Nodes[2].ID != "node-c" {
		t.Fatalf("nodes are not sorted: %#v", result.Nodes)
	}
	if result.Edges[0].ID != "node-a--node-b" ||
		result.Edges[1].ID != "node-b--node-c" {
		t.Fatalf("edges are not canonical: %#v", result.Edges)
	}
	if result.Nodes[0].Degree != 1 ||
		result.Nodes[1].Degree != 2 ||
		result.Nodes[2].Degree != 1 {
		t.Fatalf("unexpected degrees: %#v", result.Nodes)
	}
	wantMetrics := GraphMetrics{
		NodeCount:               3,
		EdgeCount:               2,
		IsolatedNodeCount:       0,
		ConnectedComponentCount: 1,
		Density:                 2.0 / 3.0,
	}
	if result.Metrics != wantMetrics {
		t.Fatalf("Metrics = %#v, want %#v", result.Metrics, wantMetrics)
	}
	if math.Abs(result.Confidence.Score-0.80) > 1e-12 {
		t.Fatalf("Confidence.Score = %f, want 0.80", result.Confidence.Score)
	}
	if result.Confidence.Level != ConfidenceLevelHigh {
		t.Fatalf(
			"Confidence.Level = %q, want %q",
			result.Confidence.Level,
			ConfidenceLevelHigh,
		)
	}
	if result.ScopeGuard != ScopeGuardResearchOnly {
		t.Fatalf("ScopeGuard = %q", result.ScopeGuard)
	}
	if result.Provenance.InputFingerprint == "" {
		t.Fatal("InputFingerprint is empty")
	}
	if report := Validate(result); report.Status != ValidationStatusValid {
		t.Fatalf("Validate() = %#v", report)
	}

	reordered := completeRequest()
	reordered.Nodes[0], reordered.Nodes[2] =
		reordered.Nodes[2], reordered.Nodes[0]
	reordered.Edges[0], reordered.Edges[1] =
		reordered.Edges[1], reordered.Edges[0]
	reordered.Edges[0].SourceNodeID,
		reordered.Edges[0].TargetNodeID =
		reordered.Edges[0].TargetNodeID,
		reordered.Edges[0].SourceNodeID

	reorderedResult, err := Build(reordered)
	if err != nil {
		t.Fatalf("Build(reordered) error = %v", err)
	}
	if reorderedResult.Provenance.InputFingerprint !=
		result.Provenance.InputFingerprint {
		t.Fatalf(
			"fingerprints differ: %q != %q",
			reorderedResult.Provenance.InputFingerprint,
			result.Provenance.InputFingerprint,
		)
	}
}

func TestBuildNormalizesIdentifiers(t *testing.T) {
	t.Parallel()

	request := completeRequest()
	request.RegionCode = " az-caucasus "
	request.Nodes = request.Nodes[:2]
	request.Nodes[0].ID = ""
	request.Nodes[0].TrajectoryID = " trajectory-c "
	request.Nodes[0].ICAO24 = "c3d4e5"
	request.Nodes[0].Callsign = " gfa300 "
	request.Edges = []EdgeInput{
		fixtureEdge(
			"node-a",
			"trajectory:trajectory-c",
			0.80,
			request.GeneratedAt,
		),
	}

	result, err := Build(request)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.RegionCode != "AZ-CAUCASUS" {
		t.Fatalf("RegionCode = %q", result.RegionCode)
	}
	var normalized Node
	for _, node := range result.Nodes {
		if node.ID == "trajectory:trajectory-c" {
			normalized = node
			break
		}
	}
	if normalized.ID == "" {
		t.Fatalf("derived node was not found: %#v", result.Nodes)
	}
	if normalized.ICAO24 != "C3D4E5" || normalized.Callsign != "GFA300" {
		t.Fatalf("normalized node = %#v", normalized)
	}
}

func TestBuildUnavailableGraph(t *testing.T) {
	t.Parallel()

	asOfTime := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	result, err := Build(Request{
		RegionCode:  "az",
		AsOfTime:    asOfTime,
		GeneratedAt: asOfTime.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status != ResultStatusUnavailable {
		t.Fatalf("Status = %q, want unavailable", result.Status)
	}
	if result.Metrics != (GraphMetrics{}) {
		t.Fatalf("Metrics = %#v, want zero", result.Metrics)
	}
	if result.Confidence.Level != ConfidenceLevelNone {
		t.Fatalf("Confidence.Level = %q, want none", result.Confidence.Level)
	}
	if len(result.Provenance.SourceNames) != 0 {
		t.Fatalf("SourceNames = %#v, want empty", result.Provenance.SourceNames)
	}
	if report := Validate(result); report.Status != ValidationStatusValid {
		t.Fatalf("Validate() = %#v", report)
	}
}

func TestBuildLimitedGraphWithIsolatedNode(t *testing.T) {
	t.Parallel()

	request := completeRequest()
	request.Nodes = request.Nodes[:1]
	request.Edges = nil

	result, err := Build(request)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status != ResultStatusLimited {
		t.Fatalf("Status = %q, want limited", result.Status)
	}
	if result.Metrics.IsolatedNodeCount != 1 ||
		result.Metrics.ConnectedComponentCount != 1 {
		t.Fatalf("Metrics = %#v", result.Metrics)
	}
	if !hasLimitation(result.Limitations, "interaction_edges_unavailable") ||
		!hasLimitation(result.Limitations, "isolated_airborne_nodes_present") {
		t.Fatalf("Limitations = %#v", result.Limitations)
	}
}

func TestBuildCalculatesConnectedComponents(t *testing.T) {
	t.Parallel()

	request := completeRequest()
	request.Nodes = append(request.Nodes, fixtureNode(
		"node-d",
		"D4E5F6",
		0.60,
		request.AsOfTime.Add(-4*time.Second),
	))
	request.Edges = []EdgeInput{
		fixtureEdge("node-a", "node-b", 0.80, request.GeneratedAt),
		fixtureEdge("node-c", "node-d", 0.70, request.GeneratedAt),
	}

	result, err := Build(request)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Metrics.ConnectedComponentCount != 2 {
		t.Fatalf(
			"ConnectedComponentCount = %d, want 2",
			result.Metrics.ConnectedComponentCount,
		)
	}
	if result.Metrics.IsolatedNodeCount != 0 {
		t.Fatalf("IsolatedNodeCount = %d, want 0", result.Metrics.IsolatedNodeCount)
	}
	if math.Abs(result.Metrics.Density-(1.0/3.0)) > 1e-12 {
		t.Fatalf("Density = %f, want 1/3", result.Metrics.Density)
	}
}

func TestBuildRejectsInvalidEvidence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Request)
	}{
		{
			name: "ground node",
			mutate: func(request *Request) {
				request.Nodes[0].OnGround = true
			},
		},
		{
			name: "future node",
			mutate: func(request *Request) {
				request.Nodes[0].ObservedAt = request.AsOfTime.Add(time.Second)
			},
		},
		{
			name: "reserved edge delimiter in node identifier",
			mutate: func(request *Request) {
				request.Nodes[0].ID = "node--a"
			},
		},
		{
			name: "duplicate node",
			mutate: func(request *Request) {
				request.Nodes[1].ID = request.Nodes[0].ID
			},
		},
		{
			name: "unknown edge node",
			mutate: func(request *Request) {
				request.Edges[0].TargetNodeID = "missing"
			},
		},
		{
			name: "self edge",
			mutate: func(request *Request) {
				request.Edges[0].TargetNodeID = request.Edges[0].SourceNodeID
			},
		},
		{
			name: "duplicate reverse edge",
			mutate: func(request *Request) {
				request.Edges = append(request.Edges, EdgeInput{
					SourceNodeID:                 "node-b",
					TargetNodeID:                 "node-a",
					Kind:                         InteractionKindNearby,
					HorizontalDistanceKilometers: 8,
					EvaluatedAt:                  request.GeneratedAt,
					SourceName:                   "fixture-scanner",
					ConfidenceScore:              0.70,
				})
			},
		},
		{
			name: "negative vertical separation",
			mutate: func(request *Request) {
				value := -1.0
				request.Edges[0].VerticalSeparationMeters = &value
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := completeRequest()
			test.mutate(&request)
			_, err := Build(request)
			if !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("Build() error = %v, want ErrInvalidRequest", err)
			}
		})
	}
}

func TestResultCloneIsIndependent(t *testing.T) {
	t.Parallel()

	result, err := Build(completeRequest())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	cloned := result.Clone()

	*cloned.Nodes[0].AltitudeMeters = 1
	*cloned.Edges[0].VerticalSeparationMeters = 1
	cloned.Edges[0].Limitations = append(
		cloned.Edges[0].Limitations,
		Limitation{Code: "changed", Message: "changed", Scope: "test"},
	)
	cloned.Confidence.Reasons[0].Code = "changed"
	cloned.Limitations[0].Code = "changed"
	cloned.Explanations[0].Code = "changed"
	cloned.Provenance.SourceNames[0] = "changed"

	if *result.Nodes[0].AltitudeMeters == 1 ||
		*result.Edges[0].VerticalSeparationMeters == 1 ||
		len(result.Edges[0].Limitations) != 0 ||
		result.Confidence.Reasons[0].Code == "changed" ||
		result.Limitations[0].Code == "changed" ||
		result.Explanations[0].Code == "changed" ||
		result.Provenance.SourceNames[0] == "changed" {
		t.Fatal("Clone() shares mutable state with the original")
	}
}

func completeRequest() Request {
	asOfTime := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	generatedAt := asOfTime.Add(2 * time.Second)
	return Request{
		RegionCode:  "az-caucasus",
		AsOfTime:    asOfTime,
		GeneratedAt: generatedAt,
		Nodes: []NodeInput{
			fixtureNode("node-c", "C3D4E5", 0.70, asOfTime.Add(-3*time.Second)),
			fixtureNode("node-a", "A1B2C3", 0.90, asOfTime.Add(-1*time.Second)),
			fixtureNode("node-b", "B2C3D4", 0.80, asOfTime.Add(-2*time.Second)),
		},
		Edges: []EdgeInput{
			fixtureEdge("node-c", "node-b", 0.75, generatedAt),
			fixtureEdge("node-b", "node-a", 0.85, generatedAt),
		},
	}
}

func fixtureNode(
	id string,
	icao24 string,
	qualityScore float64,
	observedAt time.Time,
) NodeInput {
	altitude := 10_000.0
	return NodeInput{
		ID:                          id,
		TrajectoryID:                "trajectory-" + id,
		FlightID:                    "flight-" + id,
		AircraftID:                  "aircraft-" + id,
		ICAO24:                      icao24,
		Callsign:                    "gfa" + id,
		Latitude:                    40.0,
		Longitude:                   49.0,
		AltitudeMeters:              &altitude,
		AltitudeReference:           AltitudeReferenceBarometric,
		VelocityMetersPerSecond:     220,
		HeadingDegrees:              90,
		VerticalRateMetersPerSecond: 0,
		ObservedAt:                  observedAt,
		SourceName:                  "fixture-aircraft-state",
		QualityScore:                qualityScore,
	}
}

func fixtureEdge(
	sourceNodeID string,
	targetNodeID string,
	confidenceScore float64,
	evaluatedAt time.Time,
) EdgeInput {
	verticalSeparation := 300.0
	return EdgeInput{
		SourceNodeID:                 sourceNodeID,
		TargetNodeID:                 targetNodeID,
		Kind:                         InteractionKindNearby,
		HorizontalDistanceKilometers: 8,
		VerticalSeparationMeters:     &verticalSeparation,
		EvaluatedAt:                  evaluatedAt,
		SourceName:                   "fixture-scanner",
		ConfidenceScore:              confidenceScore,
	}
}

func hasLimitation(limitations []Limitation, code string) bool {
	for _, limitation := range limitations {
		if limitation.Code == code {
			return true
		}
	}
	return false
}
