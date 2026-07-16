package interactiongraph

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

type ValidationIssue struct {
	Field   string
	Message string
}

type ValidationReport struct {
	Status ValidationStatus
	Issues []ValidationIssue
}

var icao24Pattern = regexp.MustCompile(`^[A-F0-9]{6}$`)

func Validate(result Result) ValidationReport {
	issues := make([]ValidationIssue, 0)

	if result.SchemaVersion != SchemaVersionV1 {
		issues = appendIssue(
			issues,
			"schema_version",
			"schema version is invalid",
		)
	}
	if !result.Status.IsKnown() {
		issues = appendIssue(
			issues,
			"status",
			"result status is invalid",
		)
	}
	if strings.TrimSpace(result.RegionCode) == "" {
		issues = appendIssue(
			issues,
			"region_code",
			"region code is required",
		)
	}
	if result.AsOfTime.IsZero() {
		issues = appendIssue(
			issues,
			"as_of_time",
			"as-of time is required",
		)
	}
	if result.GeneratedAt.IsZero() ||
		result.GeneratedAt.Before(result.AsOfTime) {
		issues = appendIssue(
			issues,
			"generated_at",
			"generated-at time must not be before as-of time",
		)
	}
	if result.ScopeGuard != ScopeGuardResearchOnly {
		issues = appendIssue(
			issues,
			"scope_guard",
			"research-only separation scope guard is required",
		)
	}

	nodeByID := make(map[string]Node, len(result.Nodes))
	for index, node := range result.Nodes {
		field := fmt.Sprintf("nodes[%d]", index)
		issues = validateNode(
			issues,
			field,
			node,
			result.AsOfTime,
		)
		if _, exists := nodeByID[node.ID]; exists {
			issues = appendIssue(
				issues,
				field+".id",
				"node identifier is duplicated",
			)
		}
		nodeByID[node.ID] = node
	}

	edgeByID := make(map[string]struct{}, len(result.Edges))
	degreeByNodeID := make(map[string]int, len(result.Nodes))
	for index, edge := range result.Edges {
		field := fmt.Sprintf("edges[%d]", index)
		issues = validateEdge(
			issues,
			field,
			edge,
			nodeByID,
			result.GeneratedAt,
		)
		if _, exists := edgeByID[edge.ID]; exists {
			issues = appendIssue(
				issues,
				field+".id",
				"edge identifier is duplicated",
			)
		}
		edgeByID[edge.ID] = struct{}{}
		degreeByNodeID[edge.SourceNodeID]++
		degreeByNodeID[edge.TargetNodeID]++
	}

	for index, node := range result.Nodes {
		if node.Degree != degreeByNodeID[node.ID] {
			issues = appendIssue(
				issues,
				fmt.Sprintf("nodes[%d].degree", index),
				"node degree does not match graph edges",
			)
		}
	}

	expectedMetrics := calculateMetrics(result.Nodes, result.Edges)
	if result.Metrics != expectedMetrics {
		issues = appendIssue(
			issues,
			"metrics",
			"graph metrics do not match nodes and edges",
		)
	}
	if result.Status != statusForCounts(
		expectedMetrics.NodeCount,
		expectedMetrics.EdgeCount,
	) {
		issues = appendIssue(
			issues,
			"status",
			"result status does not match graph structure",
		)
	}

	if !unitInterval(result.Confidence.Score) {
		issues = appendIssue(
			issues,
			"confidence.score",
			"confidence score must be between zero and one",
		)
	}
	if !result.Confidence.Level.IsKnown() ||
		result.Confidence.Level != confidenceLevelForScore(
			result.Confidence.Score,
		) {
		issues = appendIssue(
			issues,
			"confidence.level",
			"confidence level does not match confidence score",
		)
	}
	if len(result.Confidence.Reasons) == 0 {
		issues = appendIssue(
			issues,
			"confidence.reasons",
			"at least one confidence reason is required",
		)
	}
	for index, reason := range result.Confidence.Reasons {
		if strings.TrimSpace(reason.Code) == "" ||
			strings.TrimSpace(reason.Message) == "" ||
			!finite(reason.Contribution) {
			issues = appendIssue(
				issues,
				fmt.Sprintf("confidence.reasons[%d]", index),
				"confidence reason is invalid",
			)
		}
	}

	if len(result.Limitations) == 0 {
		issues = appendIssue(
			issues,
			"limitations",
			"at least one limitation is required",
		)
	}
	if len(result.Explanations) == 0 {
		issues = appendIssue(
			issues,
			"explanations",
			"at least one explanation is required",
		)
	}
	issues = validateStatements(issues, result)

	if len(result.Nodes) > 0 &&
		len(result.Provenance.SourceNames) == 0 {
		issues = appendIssue(
			issues,
			"provenance.source_names",
			"at least one source name is required for graph evidence",
		)
	}
	if !sort.StringsAreSorted(result.Provenance.SourceNames) ||
		hasDuplicateStrings(result.Provenance.SourceNames) {
		issues = appendIssue(
			issues,
			"provenance.source_names",
			"source names must be unique and sorted",
		)
	}
	if len(result.Nodes) == 0 &&
		!result.Provenance.LatestObservedAt.IsZero() {
		issues = appendIssue(
			issues,
			"provenance.latest_observed_at",
			"latest observed time must be empty without nodes",
		)
	}
	if !result.Provenance.LatestObservedAt.IsZero() &&
		result.Provenance.LatestObservedAt.After(result.AsOfTime) {
		issues = appendIssue(
			issues,
			"provenance.latest_observed_at",
			"latest observed time crosses the as-of boundary",
		)
	}
	if result.Provenance.InputFingerprint != inputFingerprint(result) {
		issues = appendIssue(
			issues,
			"provenance.input_fingerprint",
			"input fingerprint does not match graph evidence",
		)
	}

	if len(issues) > 0 {
		return ValidationReport{
			Status: ValidationStatusInvalid,
			Issues: issues,
		}
	}
	return ValidationReport{Status: ValidationStatusValid}
}

func validateNode(
	issues []ValidationIssue,
	field string,
	node Node,
	asOfTime time.Time,
) []ValidationIssue {
	if strings.TrimSpace(node.ID) == "" {
		issues = appendIssue(
			issues,
			field+".id",
			"node identifier is required",
		)
	}
	if strings.Contains(node.ID, "--") {
		issues = appendIssue(
			issues,
			field+".id",
			"node identifier contains the reserved edge delimiter",
		)
	}
	if !icao24Pattern.MatchString(node.ICAO24) {
		issues = appendIssue(
			issues,
			field+".icao24",
			"ICAO 24-bit address must be six uppercase hexadecimal characters",
		)
	}
	if !finite(node.Latitude) ||
		node.Latitude < -90 ||
		node.Latitude > 90 {
		issues = appendIssue(
			issues,
			field+".latitude",
			"latitude is invalid",
		)
	}
	if !finite(node.Longitude) ||
		node.Longitude < -180 ||
		node.Longitude > 180 {
		issues = appendIssue(
			issues,
			field+".longitude",
			"longitude is invalid",
		)
	}
	if !node.AltitudeReference.IsKnown() {
		issues = appendIssue(
			issues,
			field+".altitude_reference",
			"altitude reference is invalid",
		)
	}
	if node.AltitudeReference != AltitudeReferenceUnknown &&
		node.AltitudeMeters == nil {
		issues = appendIssue(
			issues,
			field+".altitude_meters",
			"known altitude reference requires altitude",
		)
	}
	if node.AltitudeMeters != nil &&
		!finite(*node.AltitudeMeters) {
		issues = appendIssue(
			issues,
			field+".altitude_meters",
			"altitude must be finite",
		)
	}
	if !finite(node.VelocityMetersPerSecond) ||
		node.VelocityMetersPerSecond < 0 {
		issues = appendIssue(
			issues,
			field+".velocity_meters_per_second",
			"velocity must be finite and non-negative",
		)
	}
	if !finite(node.HeadingDegrees) ||
		node.HeadingDegrees < 0 ||
		node.HeadingDegrees >= 360 {
		issues = appendIssue(
			issues,
			field+".heading_degrees",
			"heading must be in the range [0, 360)",
		)
	}
	if !finite(node.VerticalRateMetersPerSecond) {
		issues = appendIssue(
			issues,
			field+".vertical_rate_meters_per_second",
			"vertical rate must be finite",
		)
	}
	if node.ObservedAt.IsZero() || node.ObservedAt.After(asOfTime) {
		issues = appendIssue(
			issues,
			field+".observed_at",
			"node observation must exist and not cross the as-of boundary",
		)
	}
	if strings.TrimSpace(node.SourceName) == "" {
		issues = appendIssue(
			issues,
			field+".source_name",
			"source name is required",
		)
	}
	if !unitInterval(node.QualityScore) {
		issues = appendIssue(
			issues,
			field+".quality_score",
			"quality score must be between zero and one",
		)
	}
	if node.Degree < 0 {
		issues = appendIssue(
			issues,
			field+".degree",
			"node degree must be non-negative",
		)
	}
	return issues
}

func validateEdge(
	issues []ValidationIssue,
	field string,
	edge Edge,
	nodeByID map[string]Node,
	generatedAt time.Time,
) []ValidationIssue {
	if strings.TrimSpace(edge.SourceNodeID) == "" ||
		strings.TrimSpace(edge.TargetNodeID) == "" {
		issues = appendIssue(
			issues,
			field+".nodes",
			"both edge node identifiers are required",
		)
	}
	if edge.SourceNodeID >= edge.TargetNodeID {
		issues = appendIssue(
			issues,
			field+".nodes",
			"undirected edge node identifiers must be canonical and distinct",
		)
	}
	if edge.ID != canonicalEdgeID(
		edge.SourceNodeID,
		edge.TargetNodeID,
	) {
		issues = appendIssue(
			issues,
			field+".id",
			"edge identifier does not match canonical node order",
		)
	}
	sourceNode, sourceExists := nodeByID[edge.SourceNodeID]
	targetNode, targetExists := nodeByID[edge.TargetNodeID]
	if !sourceExists || !targetExists {
		issues = appendIssue(
			issues,
			field+".nodes",
			"edge references an unknown node",
		)
	}
	if !edge.Kind.IsKnown() {
		issues = appendIssue(
			issues,
			field+".kind",
			"interaction kind is invalid",
		)
	}
	if !finite(edge.HorizontalDistanceKilometers) ||
		edge.HorizontalDistanceKilometers < 0 {
		issues = appendIssue(
			issues,
			field+".horizontal_distance_kilometers",
			"horizontal distance must be finite and non-negative",
		)
	}
	if edge.VerticalSeparationMeters != nil &&
		(!finite(*edge.VerticalSeparationMeters) ||
			*edge.VerticalSeparationMeters < 0) {
		issues = appendIssue(
			issues,
			field+".vertical_separation_meters",
			"vertical separation must be finite and non-negative",
		)
	}
	if edge.ObservationTimeDifference < 0 {
		issues = appendIssue(
			issues,
			field+".observation_time_difference",
			"observation time difference must be non-negative",
		)
	}
	if sourceExists && targetExists {
		expectedDifference := absoluteDuration(
			sourceNode.ObservedAt.Sub(targetNode.ObservedAt),
		)
		if edge.ObservationTimeDifference != expectedDifference {
			issues = appendIssue(
				issues,
				field+".observation_time_difference",
				"observation time difference does not match edge nodes",
			)
		}
		latestObservation := sourceNode.ObservedAt
		if targetNode.ObservedAt.After(latestObservation) {
			latestObservation = targetNode.ObservedAt
		}
		if edge.EvaluatedAt.Before(latestObservation) {
			issues = appendIssue(
				issues,
				field+".evaluated_at",
				"edge evaluation predates its latest node observation",
			)
		}
	}
	if edge.EvaluatedAt.IsZero() ||
		edge.EvaluatedAt.After(generatedAt) {
		issues = appendIssue(
			issues,
			field+".evaluated_at",
			"edge evaluation must exist and not exceed generated-at time",
		)
	}
	if strings.TrimSpace(edge.SourceName) == "" {
		issues = appendIssue(
			issues,
			field+".source_name",
			"edge source name is required",
		)
	}
	if !unitInterval(edge.ConfidenceScore) {
		issues = appendIssue(
			issues,
			field+".confidence_score",
			"edge confidence score must be between zero and one",
		)
	}
	for index, limitation := range edge.Limitations {
		if strings.TrimSpace(limitation.Code) == "" ||
			strings.TrimSpace(limitation.Message) == "" ||
			strings.TrimSpace(limitation.Scope) == "" {
			issues = appendIssue(
				issues,
				fmt.Sprintf("%s.limitations[%d]", field, index),
				"edge limitation is invalid",
			)
		}
	}
	return issues
}

func validateStatements(
	issues []ValidationIssue,
	result Result,
) []ValidationIssue {
	for index, limitation := range result.Limitations {
		if strings.TrimSpace(limitation.Code) == "" ||
			strings.TrimSpace(limitation.Message) == "" ||
			strings.TrimSpace(limitation.Scope) == "" {
			issues = appendIssue(
				issues,
				fmt.Sprintf("limitations[%d]", index),
				"limitation is invalid",
			)
		}
	}
	for index, explanation := range result.Explanations {
		if strings.TrimSpace(explanation.Code) == "" ||
			strings.TrimSpace(explanation.Message) == "" {
			issues = appendIssue(
				issues,
				fmt.Sprintf("explanations[%d]", index),
				"explanation is invalid",
			)
		}
	}
	return issues
}

func calculateMetrics(
	nodes []Node,
	edges []Edge,
) GraphMetrics {
	metrics := GraphMetrics{
		NodeCount: len(nodes),
		EdgeCount: len(edges),
	}
	if len(nodes) == 0 {
		return metrics
	}

	adjacency := make(map[string][]string, len(nodes))
	for _, node := range nodes {
		adjacency[node.ID] = nil
	}
	for _, edge := range edges {
		adjacency[edge.SourceNodeID] = append(
			adjacency[edge.SourceNodeID],
			edge.TargetNodeID,
		)
		adjacency[edge.TargetNodeID] = append(
			adjacency[edge.TargetNodeID],
			edge.SourceNodeID,
		)
	}
	for _, node := range nodes {
		if len(adjacency[node.ID]) == 0 {
			metrics.IsolatedNodeCount++
		}
	}

	visited := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		if visited[node.ID] {
			continue
		}
		metrics.ConnectedComponentCount++
		queue := []string{node.ID}
		visited[node.ID] = true
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			for _, neighbor := range adjacency[current] {
				if visited[neighbor] {
					continue
				}
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	if len(nodes) >= 2 {
		metrics.Density = float64(2*len(edges)) /
			float64(len(nodes)*(len(nodes)-1))
	}
	return metrics
}

func appendIssue(
	issues []ValidationIssue,
	field string,
	message string,
) []ValidationIssue {
	return append(issues, ValidationIssue{
		Field:   field,
		Message: message,
	})
}

func unitInterval(value float64) bool {
	return finite(value) && value >= 0 && value <= 1
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func hasDuplicateStrings(values []string) bool {
	for index := 1; index < len(values); index++ {
		if values[index] == values[index-1] {
			return true
		}
	}
	return false
}

func absoluteDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}
