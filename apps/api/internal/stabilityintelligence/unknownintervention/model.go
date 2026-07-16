package unknownintervention

import "time"

const (
	Version                = "unknown-intervention-guard-v1"
	SchemaVersionV1        = "unknown-intervention-guard-v1"
	ScopeGuardResearchOnly = "research_only_no_pilot_intent_atc_instruction_or_exact_cause_claim"
)

type ResultStatus string

const (
	ResultStatusLimited  ResultStatus = "limited"
	ResultStatusComplete ResultStatus = "complete"
)

type ClaimKind string

const (
	ClaimKindContextualAssociation  ClaimKind = "contextual_association"
	ClaimKindCausalAttribution      ClaimKind = "causal_attribution"
	ClaimKindIntentAttribution      ClaimKind = "intent_attribution"
	ClaimKindOperationalInstruction ClaimKind = "operational_instruction"
)

func (kind ClaimKind) IsKnown() bool {
	switch kind {
	case ClaimKindContextualAssociation, ClaimKindCausalAttribution, ClaimKindIntentAttribution, ClaimKindOperationalInstruction:
		return true
	default:
		return false
	}
}

type EvidenceClass string

const (
	EvidenceObserved      EvidenceClass = "observed"
	EvidenceOpenlySourced EvidenceClass = "openly_sourced"
	EvidenceDerived       EvidenceClass = "derived"
	EvidenceEstimated     EvidenceClass = "estimated"
	EvidenceUnknown       EvidenceClass = "unknown"
)

func (class EvidenceClass) IsKnown() bool {
	switch class {
	case EvidenceObserved, EvidenceOpenlySourced, EvidenceDerived, EvidenceEstimated, EvidenceUnknown:
		return true
	default:
		return false
	}
}

type Decision string

const (
	DecisionAllowedContextOnly Decision = "allowed_context_only"
	DecisionLimitedContext     Decision = "limited_context"
	DecisionWithheld           Decision = "withheld"
)

type Evidence struct {
	ID          string
	Label       string
	Class       EvidenceClass
	Score       float64
	Required    bool
	Source      string
	Fingerprint string
	Limitation  string
}

type Request struct {
	SubjectID            string
	ClaimKind            ClaimKind
	ClaimText            string
	Evidence             []Evidence
	EvidenceCompleteness float64
	EvaluatedAt          time.Time
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
type Explanation struct {
	Code    string
	Message string
}
type Metrics struct {
	EvidenceCount          int
	RequiredEvidenceCount  int
	UnknownEvidenceCount   int
	EstimatedEvidenceCount int
	WeightedEvidenceScore  float64
	WeakestRequiredScore   float64
	EvidenceCompleteness   float64
}
type Provenance struct {
	InputFingerprint     string
	EvidenceFingerprints []string
	PolicyVersion        string
}
type Result struct {
	SchemaVersion   string
	Status          ResultStatus
	SubjectID       string
	ClaimKind       ClaimKind
	Decision        Decision
	ConfidenceScore float64
	Metrics         Metrics
	Reasons         []Reason
	Limitations     []Limitation
	Explanations    []Explanation
	ScopeGuard      string
	Provenance      Provenance
	EvaluatedAt     time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Reasons = append([]Reason(nil), result.Reasons...)
	cloned.Limitations = append([]Limitation(nil), result.Limitations...)
	cloned.Explanations = append([]Explanation(nil), result.Explanations...)
	cloned.Provenance.EvidenceFingerprints = append([]string(nil), result.Provenance.EvidenceFingerprints...)
	return cloned
}
