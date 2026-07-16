package failureexplanation

import "time"

const (
	Version                = "failure-explanation-engine-v1"
	SchemaVersionV1        = "failure-explanation-engine-v1"
	ScopeGuardResearchOnly = "research_only_not_for_operational_failure_or_causal_decision_use"
)

type ResultStatus string

const (
	ResultStatusLimited  ResultStatus = "limited"
	ResultStatusComplete ResultStatus = "complete"
)

type Severity string

const (
	SeverityInformation Severity = "information"
	SeverityWarning     Severity = "warning"
	SeverityBlocking    Severity = "blocking"
)

func (severity Severity) IsKnown() bool {
	switch severity {
	case SeverityInformation, SeverityWarning, SeverityBlocking:
		return true
	default:
		return false
	}
}

type Category string

const (
	CategoryDataQuality Category = "data_quality"
	CategoryEvidence    Category = "evidence"
	CategoryStability   Category = "stability"
	CategoryConfidence  Category = "confidence"
	CategoryPolicy      Category = "policy"
	CategoryScope       Category = "scope"
	CategoryUnknown     Category = "unknown"
	CategorySystem      Category = "system"
)

func (category Category) IsKnown() bool {
	switch category {
	case CategoryDataQuality, CategoryEvidence, CategoryStability,
		CategoryConfidence, CategoryPolicy, CategoryScope,
		CategoryUnknown, CategorySystem:
		return true
	default:
		return false
	}
}

type CauseClassification string

const (
	CauseClassificationObservedCondition CauseClassification = "observed_condition"
	CauseClassificationDerivedCondition  CauseClassification = "derived_condition"
	CauseClassificationUnknownCause      CauseClassification = "unknown_cause"
)

func (classification CauseClassification) IsKnown() bool {
	switch classification {
	case CauseClassificationObservedCondition,
		CauseClassificationDerivedCondition,
		CauseClassificationUnknownCause:
		return true
	default:
		return false
	}
}

type Signal struct {
	Code                 string
	Category             Category
	Severity             Severity
	Classification       CauseClassification
	Summary              string
	Detail               string
	Source               string
	BlocksUse            bool
	EvidenceFingerprints []string
}

type Request struct {
	SubjectID   string
	SubjectType string
	Signals     []Signal
	EvaluatedAt time.Time
}

type Failure struct {
	Rank                 int
	Code                 string
	Category             Category
	Severity             Severity
	Classification       CauseClassification
	Summary              string
	Detail               string
	Source               string
	BlocksUse            bool
	PriorityScore        float64
	EvidenceFingerprints []string
}

type Reason struct {
	Code    string
	Message string
	Impact  float64
}

type Confidence struct {
	Score   float64
	Level   string
	Reasons []Reason
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
	SignalCount       int
	FailureCount      int
	BlockingCount     int
	WarningCount      int
	InformationCount  int
	UnknownCauseCount int
}

type Provenance struct {
	InputFingerprint   string
	SignalFingerprints []string
	PolicyVersion      string
}

type Result struct {
	SchemaVersion string
	Status        ResultStatus
	SubjectID     string
	SubjectType   string
	PrimaryCode   string
	Failures      []Failure
	Metrics       Metrics
	Confidence    Confidence
	Limitations   []Limitation
	Explanations  []Explanation
	ScopeGuard    string
	Provenance    Provenance
	EvaluatedAt   time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Failures = make([]Failure, 0, len(result.Failures))
	for _, item := range result.Failures {
		copied := item
		copied.EvidenceFingerprints = append([]string(nil), item.EvidenceFingerprints...)
		cloned.Failures = append(cloned.Failures, copied)
	}
	cloned.Confidence.Reasons = append([]Reason(nil), result.Confidence.Reasons...)
	cloned.Limitations = append([]Limitation(nil), result.Limitations...)
	cloned.Explanations = append([]Explanation(nil), result.Explanations...)
	cloned.Provenance.SignalFingerprints = append([]string(nil), result.Provenance.SignalFingerprints...)
	return cloned
}
