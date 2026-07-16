package confidencepropagation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

func Propagate(request Request, policy Policy) (Result, error) {
	if err := policy.Validate(); err != nil {
		return Result{}, err
	}

	nodes, evaluatedAt, err := normalizeRequest(request, policy)
	if err != nil {
		return Result{}, err
	}

	results := make(map[string]NodeResult, len(nodes))
	visiting := make(map[string]bool, len(nodes))
	visited := make(map[string]bool, len(nodes))
	order := make([]string, 0, len(nodes))

	var evaluate func(string) (NodeResult, error)
	evaluate = func(nodeID string) (NodeResult, error) {
		if visited[nodeID] {
			return results[nodeID], nil
		}
		if visiting[nodeID] {
			return NodeResult{}, fmt.Errorf(
				"%w: %s",
				ErrDependencyCycle,
				nodeID,
			)
		}

		visiting[nodeID] = true
		node := nodes[nodeID]
		dependencies := append(
			[]Dependency(nil),
			node.Dependencies...,
		)
		sort.Slice(
			dependencies,
			func(left int, right int) bool {
				return dependencies[left].NodeID <
					dependencies[right].NodeID
			},
		)

		weightedDependencyScore := 0.0
		dependencyWeight := 0.0
		weakestRequiredScore := 1.0
		limitingDependencyID := ""
		requiredSeen := false
		reasons := make([]Reason, 0, 3)
		limitations := make([]Limitation, 0, 2)

		for _, dependency := range dependencies {
			child, childErr := evaluate(dependency.NodeID)
			if childErr != nil {
				return NodeResult{}, childErr
			}
			weightedDependencyScore += child.Score * dependency.Weight
			dependencyWeight += dependency.Weight
			if dependency.Required &&
				(!requiredSeen || child.Score < weakestRequiredScore) {
				requiredSeen = true
				weakestRequiredScore = child.Score
				limitingDependencyID = dependency.NodeID
			}
		}

		dependencyScore := node.LocalScore
		if dependencyWeight > 0 {
			dependencyScore = weightedDependencyScore / dependencyWeight
		}

		score := node.LocalScore
		if len(dependencies) > 0 {
			score = policy.LocalWeight*node.LocalScore +
				policy.DependencyWeight*dependencyScore
		}

		if requiredSeen {
			capValue := math.Min(
				1,
				weakestRequiredScore+
					policy.RequiredDependencyAllowance,
			)
			if score > capValue {
				score = capValue
				limitations = append(
					limitations,
					Limitation{
						Code:    "weakest_required_dependency_cap",
						Message: "A required dependency capped propagated confidence.",
						Scope:   limitingDependencyID,
					},
				)
				reasons = append(
					reasons,
					Reason{
						Code:    "required_dependency_limit",
						Message: "The weakest required dependency limited confidence.",
						Impact:  capValue,
					},
				)
			}
		}

		switch node.Classification {
		case ClassificationEstimated:
			if score > policy.EstimatedConfidenceCap {
				score = policy.EstimatedConfidenceCap
			}
			limitations = append(
				limitations,
				Limitation{
					Code:    "estimated_evidence_cap",
					Message: "Estimated evidence limits propagated confidence.",
					Scope:   nodeID,
				},
			)
		case ClassificationUnknown:
			if score > policy.UnknownConfidenceCap {
				score = policy.UnknownConfidenceCap
			}
			limitations = append(
				limitations,
				Limitation{
					Code:    "unknown_evidence_cap",
					Message: "Unknown evidence limits propagated confidence.",
					Scope:   nodeID,
				},
			)
		}

		reasons = append(
			reasons,
			Reason{
				Code:    "local_confidence",
				Message: "Node-local confidence contribution.",
				Impact:  node.LocalScore,
			},
			Reason{
				Code:    "dependency_confidence",
				Message: "Weighted dependency confidence contribution.",
				Impact:  dependencyScore,
			},
		)

		item := NodeResult{
			NodeID:               nodeID,
			Score:                score,
			Level:                levelFor(score, policy),
			DependencyScore:      dependencyScore,
			WeakestRequiredScore: weakestRequiredScore,
			LimitingDependencyID: limitingDependencyID,
			Classification:       node.Classification,
			Reasons:              reasons,
			Limitations:          limitations,
		}

		visiting[nodeID] = false
		visited[nodeID] = true
		results[nodeID] = item
		order = append(order, nodeID)
		return item, nil
	}

	target, err := evaluate(request.TargetNodeID)
	if err != nil {
		return Result{}, err
	}
	for _, nodeID := range sortedNodeIDs(nodes) {
		if visited[nodeID] {
			continue
		}
		if _, evaluationErr := evaluate(nodeID); evaluationErr != nil {
			return Result{}, evaluationErr
		}
	}

	nodeResults := make([]NodeResult, 0, len(order))
	limited := false
	for _, nodeID := range order {
		item := results[nodeID]
		nodeResults = append(nodeResults, item)
		if item.Classification == ClassificationEstimated ||
			item.Classification == ClassificationUnknown ||
			len(item.Limitations) > 0 {
			limited = true
		}
	}

	status := ResultStatusComplete
	if limited {
		status = ResultStatusLimited
	}

	result := Result{
		SchemaVersion: SchemaVersionV1,
		Status:        status,
		TargetNodeID:  target.NodeID,
		Score:         target.Score,
		Level:         target.Level,
		Nodes:         nodeResults,
		Limitations: []Limitation{
			{
				Code:    "confidence_is_not_probability",
				Message: "Propagated confidence is a bounded project score, not a calibrated probability.",
				Scope:   "interpretation",
			},
			{
				Code:    "dependency_model_is_project_derived",
				Message: "Propagation weights and caps require historical replay calibration.",
				Scope:   policy.Version,
			},
		},
		ScopeGuard: ScopeGuardResearchOnly,
		Provenance: Provenance{
			TargetNodeID:  target.NodeID,
			PolicyVersion: policy.Version,
		},
		EvaluatedAt: evaluatedAt,
	}

	for _, nodeID := range sortedNodeIDs(nodes) {
		result.Provenance.NodeFingerprints = append(
			result.Provenance.NodeFingerprints,
			nodeFingerprint(nodes[nodeID]),
		)
	}
	result.Provenance.InputFingerprint = resultFingerprint(result)

	if err := ValidateResult(result, policy); err != nil {
		return Result{}, err
	}
	return result.Clone(), nil
}

func normalizeRequest(
	request Request,
	policy Policy,
) (map[string]Node, time.Time, error) {
	if strings.TrimSpace(request.TargetNodeID) == "" ||
		len(request.Nodes) == 0 ||
		len(request.Nodes) > policy.MaximumNodeCount ||
		request.EvaluatedAt.IsZero() {
		return nil, time.Time{}, fmt.Errorf(
			"%w: identity",
			ErrInvalidRequest,
		)
	}

	nodes := make(map[string]Node, len(request.Nodes))
	for _, node := range request.Nodes {
		node.ID = strings.TrimSpace(node.ID)
		node.Label = strings.TrimSpace(node.Label)
		if node.ID == "" ||
			node.Label == "" ||
			!unitInterval(node.LocalScore) ||
			!knownKind(node.Kind) ||
			!knownClassification(node.Classification) {
			return nil, time.Time{}, fmt.Errorf(
				"%w: node",
				ErrInvalidRequest,
			)
		}
		if _, exists := nodes[node.ID]; exists {
			return nil, time.Time{}, fmt.Errorf(
				"%w: duplicate node",
				ErrInvalidRequest,
			)
		}
		dependencyIDs := make(map[string]struct{}, len(node.Dependencies))
		for _, dependency := range node.Dependencies {
			dependencyID := strings.TrimSpace(dependency.NodeID)
			if dependencyID == "" ||
				!unitInterval(dependency.Weight) ||
				dependency.Weight == 0 {
				return nil, time.Time{}, fmt.Errorf(
					"%w: dependency",
					ErrInvalidRequest,
				)
			}
			if _, exists := dependencyIDs[dependencyID]; exists {
				return nil, time.Time{}, fmt.Errorf(
					"%w: duplicate dependency",
					ErrInvalidRequest,
				)
			}
			dependencyIDs[dependencyID] = struct{}{}
		}
		nodes[node.ID] = node
	}

	if _, exists := nodes[request.TargetNodeID]; !exists {
		return nil, time.Time{}, fmt.Errorf(
			"%w: target missing",
			ErrInvalidRequest,
		)
	}
	for _, node := range nodes {
		for _, dependency := range node.Dependencies {
			if _, exists := nodes[dependency.NodeID]; !exists {
				return nil, time.Time{}, fmt.Errorf(
					"%w: dependency missing",
					ErrInvalidRequest,
				)
			}
		}
	}

	return nodes, request.EvaluatedAt.UTC(), nil
}

func levelFor(score float64, policy Policy) string {
	if score >= policy.HighThreshold {
		return "high"
	}
	if score >= policy.MediumThreshold {
		return "medium"
	}
	if score > 0 {
		return "low"
	}
	return "none"
}

func knownKind(kind NodeKind) bool {
	return kind == NodeKindEvidence ||
		kind == NodeKindDecision ||
		kind == NodeKindOutput
}

func knownClassification(value Classification) bool {
	switch value {
	case ClassificationObserved,
		ClassificationOpenlySourced,
		ClassificationDerived,
		ClassificationEstimated,
		ClassificationUnknown:
		return true
	default:
		return false
	}
}

func sortedNodeIDs(nodes map[string]Node) []string {
	identifiers := make([]string, 0, len(nodes))
	for identifier := range nodes {
		identifiers = append(identifiers, identifier)
	}
	sort.Strings(identifiers)
	return identifiers
}

func nodeFingerprint(node Node) string {
	dependencies := append([]Dependency(nil), node.Dependencies...)
	sort.Slice(
		dependencies,
		func(left int, right int) bool {
			if dependencies[left].NodeID == dependencies[right].NodeID {
				return dependencies[left].Weight < dependencies[right].Weight
			}
			return dependencies[left].NodeID < dependencies[right].NodeID
		},
	)
	payload := struct {
		ID                string
		Label             string
		Kind              NodeKind
		Classification    Classification
		LocalScore        float64
		Dependencies      []Dependency
		SourceFingerprint string
	}{
		ID:                node.ID,
		Label:             node.Label,
		Kind:              node.Kind,
		Classification:    node.Classification,
		LocalScore:        node.LocalScore,
		Dependencies:      dependencies,
		SourceFingerprint: node.SourceFingerprint,
	}
	encoded, _ := json.Marshal(payload)
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func resultFingerprint(result Result) string {
	fingerprints := append(
		[]string(nil),
		result.Provenance.NodeFingerprints...,
	)
	sort.Strings(fingerprints)
	payload := struct {
		SchemaVersion string
		Status        ResultStatus
		TargetNodeID  string
		Score         float64
		Level         string
		Fingerprints  []string
		PolicyVersion string
		EvaluatedAt   time.Time
	}{
		SchemaVersion: result.SchemaVersion,
		Status:        result.Status,
		TargetNodeID:  result.TargetNodeID,
		Score:         result.Score,
		Level:         result.Level,
		Fingerprints:  fingerprints,
		PolicyVersion: result.Provenance.PolicyVersion,
		EvaluatedAt:   result.EvaluatedAt.UTC(),
	}
	encoded, _ := json.Marshal(payload)
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}
