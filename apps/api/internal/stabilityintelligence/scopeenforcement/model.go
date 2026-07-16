package scopeenforcement

import "time"

const (
	Version                = "scope-guard-enforcement-v1"
	SchemaVersionV1        = "scope-guard-enforcement-v1"
	ScopeGuardResearchOnly = "research_only_scope_enforcement_not_operational_authorization"
)

type ResultStatus string

const (
	ResultStatusLimited  ResultStatus = "limited"
	ResultStatusComplete ResultStatus = "complete"
)

type ClaimScope string

const (
	ScopeResearchAnalysis      ClaimScope = "research_analysis"
	ScopeResearchVisualization ClaimScope = "research_visualization"
	ScopeOperationalDecision   ClaimScope = "operational_decision"
	ScopeAirTrafficControl     ClaimScope = "air_traffic_control"
	ScopeFlightPlanning        ClaimScope = "flight_planning"
	ScopeSafetyCritical        ClaimScope = "safety_critical"
)

func (scope ClaimScope) IsKnown() bool {
	switch scope {
	case ScopeResearchAnalysis, ScopeResearchVisualization, ScopeOperationalDecision, ScopeAirTrafficControl, ScopeFlightPlanning, ScopeSafetyCritical:
		return true
	default:
		return false
	}
}

type ClaimStrength string

const (
	StrengthDescriptive ClaimStrength = "descriptive"
	StrengthAnalytical  ClaimStrength = "analytical"
	StrengthCausal      ClaimStrength = "causal"
	StrengthDirective   ClaimStrength = "directive"
	StrengthCertain     ClaimStrength = "certain"
)

func (strength ClaimStrength) IsKnown() bool {
	switch strength {
	case StrengthDescriptive, StrengthAnalytical, StrengthCausal, StrengthDirective, StrengthCertain:
		return true
	default:
		return false
	}
}

type Decision string

const (
	DecisionAllowed Decision = "allowed"
	DecisionLimited Decision = "limited"
	DecisionBlocked Decision = "blocked"
)

type Claim struct {
	Code        string
	Text        string
	Capability  string
	Scope       ClaimScope
	Strength    ClaimStrength
	SourceGuard string
}
type Request struct {
	SubjectID      string
	DeclaredGuards []string
	Claims         []Claim
	EvaluatedAt    time.Time
}
type Violation struct {
	Code      string
	ClaimCode string
	Message   string
	Blocking  bool
}
type ClaimResult struct {
	Claim      Claim
	Decision   Decision
	Violations []Violation
}
type Metrics struct {
	ClaimCount   int
	AllowedCount int
	LimitedCount int
	BlockedCount int
	GuardCount   int
}
type Limitation struct {
	Code    string
	Message string
	Scope   string
}
type Provenance struct {
	InputFingerprint  string
	ClaimFingerprints []string
	DeclaredGuards    []string
	PolicyVersion     string
}
type Result struct {
	SchemaVersion string
	Status        ResultStatus
	SubjectID     string
	Decision      Decision
	Claims        []ClaimResult
	Violations    []Violation
	Metrics       Metrics
	Limitations   []Limitation
	ScopeGuard    string
	Provenance    Provenance
	EvaluatedAt   time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Claims = make([]ClaimResult, 0, len(result.Claims))
	for _, item := range result.Claims {
		copied := item
		copied.Violations = append([]Violation(nil), item.Violations...)
		cloned.Claims = append(cloned.Claims, copied)
	}
	cloned.Violations = append([]Violation(nil), result.Violations...)
	cloned.Limitations = append([]Limitation(nil), result.Limitations...)
	cloned.Provenance.ClaimFingerprints = append([]string(nil), result.Provenance.ClaimFingerprints...)
	cloned.Provenance.DeclaredGuards = append([]string(nil), result.Provenance.DeclaredGuards...)
	return cloned
}
