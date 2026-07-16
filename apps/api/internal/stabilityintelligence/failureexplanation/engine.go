package failureexplanation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func Explain(request Request, policy Policy) (Result, error) {
	if err := policy.Validate(); err != nil {
		return Result{}, err
	}
	normalized, err := normalizeRequest(request, policy)
	if err != nil {
		return Result{}, err
	}

	failures := make([]Failure, 0, len(normalized.Signals))
	metrics := Metrics{SignalCount: len(normalized.Signals)}
	signalFingerprints := make([]string, 0, len(normalized.Signals))
	confidenceTotal := 0.0

	for _, signal := range normalized.Signals {
		priority := severityWeight(signal.Severity, policy)
		if signal.Classification == CauseClassificationUnknownCause {
			priority += policy.UnknownCausePriorityBoost
			metrics.UnknownCauseCount++
		}
		if signal.BlocksUse {
			priority += policy.BlockingPriorityBoost
			metrics.BlockingCount++
		}
		if priority > 1 {
			priority = 1
		}
		switch signal.Severity {
		case SeverityInformation:
			metrics.InformationCount++
		case SeverityWarning:
			metrics.WarningCount++
		}
		failure := Failure{
			Code:                 signal.Code,
			Category:             signal.Category,
			Severity:             signal.Severity,
			Classification:       signal.Classification,
			Summary:              signal.Summary,
			Detail:               signal.Detail,
			Source:               signal.Source,
			BlocksUse:            signal.BlocksUse,
			PriorityScore:        priority,
			EvidenceFingerprints: append([]string(nil), signal.EvidenceFingerprints...),
		}
		failures = append(failures, failure)
		signalFingerprints = append(signalFingerprints, fingerprint(signal))
		confidenceTotal += signalConfidence(signal)
	}

	sort.Slice(failures, func(i, j int) bool {
		if failures[i].PriorityScore != failures[j].PriorityScore {
			return failures[i].PriorityScore > failures[j].PriorityScore
		}
		if failures[i].Category != failures[j].Category {
			return failures[i].Category < failures[j].Category
		}
		if failures[i].Code != failures[j].Code {
			return failures[i].Code < failures[j].Code
		}
		return failures[i].Source < failures[j].Source
	})
	for index := range failures {
		failures[index].Rank = index + 1
	}
	metrics.FailureCount = len(failures)
	sort.Strings(signalFingerprints)

	confidenceScore := 1.0
	if len(normalized.Signals) > 0 {
		confidenceScore = confidenceTotal / float64(len(normalized.Signals))
	}
	confidence := Confidence{
		Score: confidenceScore,
		Level: confidenceLevel(confidenceScore),
		Reasons: []Reason{
			{Code: "classified_failure_signals", Message: "Confidence reflects the classification and evidence attached to the normalized failure signals.", Impact: confidenceScore},
		},
	}
	if metrics.UnknownCauseCount > 0 {
		confidence.Reasons = append(confidence.Reasons, Reason{Code: "unknown_cause_preserved", Message: "One or more causes remain unknown and are not replaced with inferred intent.", Impact: -0.20})
	}

	status := ResultStatusComplete
	if metrics.UnknownCauseCount > 0 || confidenceScore < policy.CompleteConfidenceMinimum {
		status = ResultStatusLimited
	}
	primaryCode := ""
	if len(failures) > 0 {
		primaryCode = failures[0].Code
	}
	result := Result{
		SchemaVersion: SchemaVersionV1,
		Status:        status,
		SubjectID:     normalized.SubjectID,
		SubjectType:   normalized.SubjectType,
		PrimaryCode:   primaryCode,
		Failures:      failures,
		Metrics:       metrics,
		Confidence:    confidence,
		Limitations: []Limitation{
			{Code: "explanation_not_causation", Message: "The engine explains observed or derived failure conditions and does not prove operational causation.", Scope: "causal_claim"},
			{Code: "source_limitations_preserved", Message: "Source limitations remain attached and are not converted into stronger conclusions.", Scope: "evidence"},
		},
		Explanations: []Explanation{
			{Code: "priority_order", Message: "Failures are ranked by explicit severity, blocking effect, and unknown-cause preservation."},
			{Code: "deterministic_output", Message: "Equivalent normalized inputs produce the same ordered explanation fingerprint."},
		},
		ScopeGuard:  ScopeGuardResearchOnly,
		Provenance:  Provenance{SignalFingerprints: signalFingerprints, PolicyVersion: policy.Version},
		EvaluatedAt: normalized.EvaluatedAt,
	}
	result.Provenance.InputFingerprint = resultFingerprint(result)
	if err := ValidateResult(result, policy); err != nil {
		return Result{}, err
	}
	return result.Clone(), nil
}

func normalizeRequest(request Request, policy Policy) (Request, error) {
	normalized := Request{
		SubjectID:   strings.TrimSpace(request.SubjectID),
		SubjectType: strings.TrimSpace(request.SubjectType),
		EvaluatedAt: request.EvaluatedAt.UTC(),
	}
	if normalized.SubjectID == "" || normalized.SubjectType == "" || normalized.EvaluatedAt.IsZero() {
		return Request{}, fmt.Errorf("invalid failure explanation request identity")
	}
	if len(request.Signals) > policy.MaximumSignalCount {
		return Request{}, fmt.Errorf("failure signal capacity exceeded")
	}
	seen := map[string]struct{}{}
	for _, input := range request.Signals {
		signal := input
		signal.Code = strings.TrimSpace(signal.Code)
		signal.Summary = strings.TrimSpace(signal.Summary)
		signal.Detail = strings.TrimSpace(signal.Detail)
		signal.Source = strings.TrimSpace(signal.Source)
		sort.Strings(signal.EvidenceFingerprints)
		signal.EvidenceFingerprints = uniqueStrings(signal.EvidenceFingerprints)
		if signal.Code == "" || signal.Summary == "" || signal.Source == "" ||
			!signal.Category.IsKnown() || !signal.Severity.IsKnown() || !signal.Classification.IsKnown() {
			return Request{}, fmt.Errorf("invalid failure signal %q", signal.Code)
		}
		for _, value := range signal.EvidenceFingerprints {
			if !strings.HasPrefix(value, "sha256:") {
				return Request{}, fmt.Errorf("invalid evidence fingerprint for %q", signal.Code)
			}
		}
		key := signal.Source + "\x00" + signal.Code
		if _, exists := seen[key]; exists {
			return Request{}, fmt.Errorf("duplicate failure signal %q", key)
		}
		seen[key] = struct{}{}
		normalized.Signals = append(normalized.Signals, signal)
	}
	sort.Slice(normalized.Signals, func(i, j int) bool {
		if normalized.Signals[i].Source != normalized.Signals[j].Source {
			return normalized.Signals[i].Source < normalized.Signals[j].Source
		}
		return normalized.Signals[i].Code < normalized.Signals[j].Code
	})
	return normalized, nil
}

func severityWeight(severity Severity, policy Policy) float64 {
	switch severity {
	case SeverityBlocking:
		return policy.SeverityWeightBlocking
	case SeverityWarning:
		return policy.SeverityWeightWarning
	default:
		return policy.SeverityWeightInformation
	}
}

func signalConfidence(signal Signal) float64 {
	score := 0.75
	switch signal.Classification {
	case CauseClassificationObservedCondition:
		score = 0.90
	case CauseClassificationDerivedCondition:
		score = 0.75
	case CauseClassificationUnknownCause:
		score = 0.30
	}
	if len(signal.EvidenceFingerprints) == 0 {
		score *= 0.80
	}
	return score
}

func confidenceLevel(score float64) string {
	switch {
	case score >= 0.80:
		return "high"
	case score >= 0.60:
		return "medium"
	case score > 0:
		return "low"
	default:
		return "none"
	}
}

func uniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}

func fingerprint(value any) string {
	encoded, _ := json.Marshal(value)
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func resultFingerprint(result Result) string {
	copy := result.Clone()
	copy.Provenance.InputFingerprint = ""
	return fingerprint(copy)
}
