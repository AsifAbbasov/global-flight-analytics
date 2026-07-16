package interactiongraph

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func Build(request Request) (Result, error) {
	if strings.TrimSpace(request.RegionCode) == "" ||
		request.AsOfTime.IsZero() ||
		request.GeneratedAt.IsZero() ||
		request.GeneratedAt.Before(request.AsOfTime) {
		return Result{}, fmt.Errorf(
			"%w: region, as-of time, and generated-at time are required",
			ErrInvalidRequest,
		)
	}

	result := Result{
		SchemaVersion: SchemaVersionV1,
		RegionCode: strings.ToUpper(
			strings.TrimSpace(request.RegionCode),
		),
		AsOfTime:    request.AsOfTime.UTC(),
		GeneratedAt: request.GeneratedAt.UTC(),
		ScopeGuard:  ScopeGuardResearchOnly,
	}

	nodes, nodeByID, err := buildNodes(
		request.Nodes,
		result.AsOfTime,
	)
	if err != nil {
		return Result{}, err
	}
	result.Nodes = nodes

	edges, err := buildEdges(
		request.Edges,
		nodeByID,
		result.GeneratedAt,
	)
	if err != nil {
		return Result{}, err
	}
	result.Edges = edges

	applyDegrees(result.Nodes, result.Edges)
	result.Metrics = calculateMetrics(result.Nodes, result.Edges)
	result.Status = statusForCounts(
		result.Metrics.NodeCount,
		result.Metrics.EdgeCount,
	)
	result.Confidence = buildConfidence(
		result.Nodes,
		result.Edges,
	)
	result.Limitations = buildLimitations(
		result.Nodes,
		result.Edges,
		result.Metrics,
	)
	result.Explanations = buildExplanations(result.Metrics)
	result.Provenance = buildProvenance(
		result.Nodes,
		result.Edges,
	)
	result.Provenance.InputFingerprint = inputFingerprint(result)

	report := Validate(result)
	if report.Status != ValidationStatusValid {
		return Result{}, fmt.Errorf(
			"%w: issues=%v",
			ErrInvalidGraph,
			report.Issues,
		)
	}
	return result.Clone(), nil
}

func buildNodes(
	inputs []NodeInput,
	asOfTime time.Time,
) ([]Node, map[string]Node, error) {
	nodes := make([]Node, 0, len(inputs))
	nodeByID := make(map[string]Node, len(inputs))
	for index, input := range inputs {
		if input.OnGround {
			return nil, nil, fmt.Errorf(
				"%w: nodes[%d] is on the ground; the airborne graph accepts airborne evidence only",
				ErrInvalidRequest,
				index,
			)
		}
		node := normalizeNode(input)
		if _, exists := nodeByID[node.ID]; exists {
			return nil, nil, fmt.Errorf(
				"%w: duplicate node identifier %q",
				ErrInvalidRequest,
				node.ID,
			)
		}
		issues := validateNode(
			nil,
			fmt.Sprintf("nodes[%d]", index),
			node,
			asOfTime,
		)
		if len(issues) > 0 {
			return nil, nil, fmt.Errorf(
				"%w: issues=%v",
				ErrInvalidRequest,
				issues,
			)
		}
		nodes = append(nodes, node)
		nodeByID[node.ID] = node
	}
	sort.Slice(nodes, func(left int, right int) bool {
		return nodes[left].ID < nodes[right].ID
	})
	return nodes, nodeByID, nil
}

func buildEdges(
	inputs []EdgeInput,
	nodeByID map[string]Node,
	generatedAt time.Time,
) ([]Edge, error) {
	edges := make([]Edge, 0, len(inputs))
	edgeIDs := make(map[string]struct{}, len(inputs))
	for index, input := range inputs {
		edge := normalizeEdge(input, nodeByID, generatedAt)
		if _, exists := edgeIDs[edge.ID]; exists {
			return nil, fmt.Errorf(
				"%w: duplicate undirected edge %q",
				ErrInvalidRequest,
				edge.ID,
			)
		}
		issues := validateEdge(
			nil,
			fmt.Sprintf("edges[%d]", index),
			edge,
			nodeByID,
			generatedAt,
		)
		if len(issues) > 0 {
			return nil, fmt.Errorf(
				"%w: issues=%v",
				ErrInvalidRequest,
				issues,
			)
		}
		edges = append(edges, edge)
		edgeIDs[edge.ID] = struct{}{}
	}
	sort.Slice(edges, func(left int, right int) bool {
		return edges[left].ID < edges[right].ID
	})
	return edges, nil
}

func normalizeNode(input NodeInput) Node {
	icao24 := strings.ToUpper(strings.TrimSpace(input.ICAO24))
	nodeID := strings.TrimSpace(input.ID)
	if nodeID == "" {
		trajectoryID := strings.TrimSpace(input.TrajectoryID)
		if trajectoryID != "" {
			nodeID = "trajectory:" + trajectoryID
		} else if icao24 != "" {
			nodeID = "icao24:" + icao24
		}
	}
	altitudeReference := input.AltitudeReference
	if altitudeReference == "" {
		altitudeReference = AltitudeReferenceUnknown
	}
	return Node{
		ID:           nodeID,
		TrajectoryID: strings.TrimSpace(input.TrajectoryID),
		FlightID:     strings.TrimSpace(input.FlightID),
		AircraftID:   strings.TrimSpace(input.AircraftID),
		ICAO24:       icao24,
		Callsign: strings.ToUpper(
			strings.TrimSpace(input.Callsign),
		),
		Latitude:          input.Latitude,
		Longitude:         input.Longitude,
		AltitudeMeters:    cloneFloat64(input.AltitudeMeters),
		AltitudeReference: altitudeReference,
		VelocityMetersPerSecond: input.
			VelocityMetersPerSecond,
		HeadingDegrees: input.HeadingDegrees,
		VerticalRateMetersPerSecond: input.
			VerticalRateMetersPerSecond,
		ObservedAt:   input.ObservedAt.UTC(),
		SourceName:   strings.TrimSpace(input.SourceName),
		QualityScore: input.QualityScore,
	}
}

func normalizeEdge(
	input EdgeInput,
	nodeByID map[string]Node,
	generatedAt time.Time,
) Edge {
	sourceNodeID := strings.TrimSpace(input.SourceNodeID)
	targetNodeID := strings.TrimSpace(input.TargetNodeID)
	if sourceNodeID > targetNodeID {
		sourceNodeID, targetNodeID = targetNodeID, sourceNodeID
	}
	kind := input.Kind
	if kind == "" {
		kind = InteractionKindUnknown
	}
	evaluatedAt := input.EvaluatedAt.UTC()
	if evaluatedAt.IsZero() {
		evaluatedAt = generatedAt
	}
	observationDifference := time.Duration(0)
	sourceNode, sourceExists := nodeByID[sourceNodeID]
	targetNode, targetExists := nodeByID[targetNodeID]
	if sourceExists && targetExists {
		observationDifference = absoluteDuration(
			sourceNode.ObservedAt.Sub(targetNode.ObservedAt),
		)
	}
	return Edge{
		ID:           canonicalEdgeID(sourceNodeID, targetNodeID),
		SourceNodeID: sourceNodeID,
		TargetNodeID: targetNodeID,
		Kind:         kind,
		HorizontalDistanceKilometers: input.
			HorizontalDistanceKilometers,
		VerticalSeparationMeters: cloneFloat64(
			input.VerticalSeparationMeters,
		),
		ObservationTimeDifference: observationDifference,
		EvaluatedAt:               evaluatedAt,
		SourceName: strings.TrimSpace(
			input.SourceName,
		),
		ConfidenceScore: input.ConfidenceScore,
		Limitations: append(
			[]Limitation(nil),
			input.Limitations...,
		),
	}
}

func canonicalEdgeID(sourceNodeID string, targetNodeID string) string {
	return sourceNodeID + "--" + targetNodeID
}

func applyDegrees(nodes []Node, edges []Edge) {
	degreeByNodeID := make(map[string]int, len(nodes))
	for _, edge := range edges {
		degreeByNodeID[edge.SourceNodeID]++
		degreeByNodeID[edge.TargetNodeID]++
	}
	for index := range nodes {
		nodes[index].Degree = degreeByNodeID[nodes[index].ID]
	}
}

func buildConfidence(nodes []Node, edges []Edge) Confidence {
	if len(nodes) == 0 {
		return Confidence{
			Score: 0,
			Level: ConfidenceLevelNone,
			Reasons: []ConfidenceReason{
				{
					Code:         "airborne_evidence_unavailable",
					Message:      "No airborne node evidence was available for the interaction graph.",
					Contribution: 0,
				},
			},
		}
	}

	nodeTotal := 0.0
	for _, node := range nodes {
		nodeTotal += node.QualityScore
	}
	nodeAverage := nodeTotal / float64(len(nodes))
	reasons := []ConfidenceReason{
		{
			Code:         "mean_node_quality",
			Message:      "The graph confidence includes the mean quality of airborne node evidence.",
			Contribution: nodeAverage,
		},
	}

	total := nodeTotal
	count := len(nodes)
	if len(edges) > 0 {
		edgeTotal := 0.0
		for _, edge := range edges {
			edgeTotal += edge.ConfidenceScore
		}
		edgeAverage := edgeTotal / float64(len(edges))
		reasons = append(reasons, ConfidenceReason{
			Code:         "mean_edge_confidence",
			Message:      "The graph confidence includes the mean confidence of prepared pairwise interaction evidence.",
			Contribution: edgeAverage,
		})
		total += edgeTotal
		count += len(edges)
	} else {
		reasons = append(reasons, ConfidenceReason{
			Code:         "interaction_edges_unavailable",
			Message:      "No pairwise interaction edge evidence was available.",
			Contribution: 0,
		})
	}

	score := total / float64(count)
	return Confidence{
		Score:   score,
		Level:   confidenceLevelForScore(score),
		Reasons: reasons,
	}
}

func buildLimitations(
	nodes []Node,
	edges []Edge,
	metrics GraphMetrics,
) []Limitation {
	limitations := []Limitation{
		{
			Code:    "research_only_not_operational_separation",
			Message: "The interaction graph is research context and must not be used as operational separation or collision-avoidance logic.",
			Scope:   "operational_use",
		},
	}
	if len(nodes) == 0 {
		limitations = append(limitations, Limitation{
			Code:    "airborne_nodes_unavailable",
			Message: "No airborne node evidence was available for this region and as-of time.",
			Scope:   "graph_coverage",
		})
	}
	if len(edges) == 0 {
		limitations = append(limitations, Limitation{
			Code:    "interaction_edges_unavailable",
			Message: "The graph contains no prepared pairwise interaction edges; this is not proof of safe separation.",
			Scope:   "interaction_coverage",
		})
	}
	if metrics.IsolatedNodeCount > 0 {
		limitations = append(limitations, Limitation{
			Code:    "isolated_airborne_nodes_present",
			Message: "One or more airborne nodes have no prepared interaction edge in the current graph.",
			Scope:   "graph_topology",
		})
	}
	missingVerticalSeparation := false
	for _, edge := range edges {
		if edge.VerticalSeparationMeters == nil {
			missingVerticalSeparation = true
			break
		}
	}
	if missingVerticalSeparation {
		limitations = append(limitations, Limitation{
			Code:    "vertical_separation_incomplete",
			Message: "At least one interaction edge lacks vertical separation evidence.",
			Scope:   "vertical_evidence",
		})
	}
	return limitations
}

func buildExplanations(metrics GraphMetrics) []Explanation {
	explanations := []Explanation{
		{
			Code:    "undirected_airborne_context_graph",
			Message: "Each node represents prepared airborne evidence and each edge represents a symmetric contextual relationship between two nodes.",
		},
		{
			Code:    "interaction_is_not_risk_classification",
			Message: "An interaction edge records contextual evidence and does not by itself classify operational separation risk.",
		},
	}
	if metrics.EdgeCount == 0 {
		explanations = append(explanations, Explanation{
			Code:    "edge_absence_is_not_safety_evidence",
			Message: "The absence of interaction edges may reflect filtering or missing evidence and must not be interpreted as a safety conclusion.",
		})
	}
	return explanations
}

func buildProvenance(nodes []Node, edges []Edge) Provenance {
	sourceSet := make(map[string]struct{})
	latestObservedAt := time.Time{}
	for _, node := range nodes {
		sourceSet[node.SourceName] = struct{}{}
		if node.ObservedAt.After(latestObservedAt) {
			latestObservedAt = node.ObservedAt
		}
	}
	for _, edge := range edges {
		sourceSet[edge.SourceName] = struct{}{}
	}
	sourceNames := make([]string, 0, len(sourceSet))
	for sourceName := range sourceSet {
		sourceNames = append(sourceNames, sourceName)
	}
	sort.Strings(sourceNames)
	return Provenance{
		SourceNames:      sourceNames,
		LatestObservedAt: latestObservedAt,
	}
}
