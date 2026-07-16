package confidencepropagation

import "time"

const (
	Version                = "confidence-propagation-v1"
	SchemaVersionV1        = "confidence-propagation-v1"
	ScopeGuardResearchOnly = "research_only_not_for_operational_decision_use"
)

type ResultStatus string

const (
	ResultStatusLimited  ResultStatus = "limited"
	ResultStatusComplete ResultStatus = "complete"
)

type NodeKind string

const (
	NodeKindEvidence NodeKind = "evidence"
	NodeKindDecision NodeKind = "decision"
	NodeKindOutput   NodeKind = "output"
)

type Classification string

const (
	ClassificationObserved      Classification = "observed"
	ClassificationOpenlySourced Classification = "openly_sourced"
	ClassificationDerived       Classification = "derived"
	ClassificationEstimated     Classification = "estimated"
	ClassificationUnknown       Classification = "unknown"
)

type Dependency struct {
	NodeID   string
	Weight   float64
	Required bool
}

type Node struct {
	ID                string
	Label             string
	Kind              NodeKind
	Classification    Classification
	LocalScore        float64
	Dependencies      []Dependency
	SourceFingerprint string
}

type Request struct {
	TargetNodeID string
	Nodes        []Node
	EvaluatedAt  time.Time
}

type Reason struct {
	Code    string
	Message string
	Impact  float64
}

type Limitation struct {
	Code    string
	Message string
	Scope   string
}

type NodeResult struct {
	NodeID               string
	Score                float64
	Level                string
	DependencyScore      float64
	WeakestRequiredScore float64
	LimitingDependencyID string
	Classification       Classification
	Reasons              []Reason
	Limitations          []Limitation
}

type Provenance struct {
	InputFingerprint string
	TargetNodeID     string
	NodeFingerprints []string
	PolicyVersion    string
}

type Result struct {
	SchemaVersion string
	Status        ResultStatus
	TargetNodeID  string
	Score         float64
	Level         string
	Nodes         []NodeResult
	Limitations   []Limitation
	ScopeGuard    string
	Provenance    Provenance
	EvaluatedAt   time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Nodes = make([]NodeResult, 0, len(result.Nodes))
	for _, item := range result.Nodes {
		copied := item
		copied.Reasons = append([]Reason(nil), item.Reasons...)
		copied.Limitations = append([]Limitation(nil), item.Limitations...)
		cloned.Nodes = append(cloned.Nodes, copied)
	}
	cloned.Limitations = append([]Limitation(nil), result.Limitations...)
	cloned.Provenance.NodeFingerprints = append(
		[]string(nil),
		result.Provenance.NodeFingerprints...,
	)
	return cloned
}
