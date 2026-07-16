package unknownintervention

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func Evaluate(request Request, policy Policy) (Result, error) {
	if err := policy.Validate(); err != nil {
		return Result{}, err
	}
	normalized, err := normalize(request, policy)
	if err != nil {
		return Result{}, err
	}

	metrics := Metrics{EvidenceCount: len(normalized.Evidence), EvidenceCompleteness: normalized.EvidenceCompleteness, WeakestRequiredScore: 1}
	weighted, totalWeight := 0.0, 0.0
	fingerprints := make([]string, 0, len(normalized.Evidence))
	for _, evidence := range normalized.Evidence {
		weight := 0.5
		if evidence.Required {
			weight = 1
			metrics.RequiredEvidenceCount++
			if evidence.Score < metrics.WeakestRequiredScore {
				metrics.WeakestRequiredScore = evidence.Score
			}
		}
		weighted += evidence.Score * weight
		totalWeight += weight
		if evidence.Class == EvidenceUnknown {
			metrics.UnknownEvidenceCount++
		}
		if evidence.Class == EvidenceEstimated {
			metrics.EstimatedEvidenceCount++
		}
		fingerprints = append(fingerprints, evidence.Fingerprint)
	}
	if metrics.RequiredEvidenceCount == 0 {
		metrics.WeakestRequiredScore = 1
	}
	if totalWeight > 0 {
		metrics.WeightedEvidenceScore = weighted / totalWeight
	}
	score := metrics.WeightedEvidenceScore
	if metrics.RequiredEvidenceCount > 0 && metrics.WeakestRequiredScore < score {
		score = metrics.WeakestRequiredScore
	}
	if metrics.EstimatedEvidenceCount > 0 && score > policy.EstimatedEvidenceConfidenceCap {
		score = policy.EstimatedEvidenceConfidenceCap
	}
	if metrics.UnknownEvidenceCount > 0 && score > policy.UnknownEvidenceConfidenceCap {
		score = policy.UnknownEvidenceConfidenceCap
	}
	if score > normalized.EvidenceCompleteness {
		score = normalized.EvidenceCompleteness
	}

	decision := DecisionAllowedContextOnly
	status := ResultStatusComplete
	reasons := []Reason{{Code: "contextual_evidence_score", Message: "The score summarizes evidence suitability for contextual association only.", Impact: score}}
	switch normalized.ClaimKind {
	case ClaimKindIntentAttribution:
		decision, status = DecisionWithheld, ResultStatusLimited
		reasons = append(reasons, Reason{Code: "pilot_intent_unavailable", Message: "Open surveillance data does not reveal pilot intent.", Impact: -1})
	case ClaimKindOperationalInstruction:
		decision, status = DecisionWithheld, ResultStatusLimited
		reasons = append(reasons, Reason{Code: "atc_instruction_unavailable", Message: "The project has no operational instruction feed.", Impact: -1})
	case ClaimKindCausalAttribution:
		decision, status = DecisionWithheld, ResultStatusLimited
		reasons = append(reasons, Reason{Code: "exact_cause_not_proven", Message: "Temporal or spatial association does not prove exact cause.", Impact: -1})
	default:
		if metrics.UnknownEvidenceCount > 0 || metrics.WeightedEvidenceScore < policy.LimitedConfidenceMinimum || normalized.EvidenceCompleteness < policy.LimitedCompletenessMinimum || metrics.WeakestRequiredScore < policy.RequiredEvidenceMinimum {
			decision, status = DecisionWithheld, ResultStatusLimited
		} else if metrics.EstimatedEvidenceCount > 0 || score < policy.AllowedConfidenceMinimum || normalized.EvidenceCompleteness < policy.AllowedCompletenessMinimum {
			decision, status = DecisionLimitedContext, ResultStatusLimited
		}
	}
	sort.Strings(fingerprints)
	result := Result{
		SchemaVersion: SchemaVersionV1, Status: status, SubjectID: normalized.SubjectID, ClaimKind: normalized.ClaimKind, Decision: decision, ConfidenceScore: score, Metrics: metrics, Reasons: reasons,
		Limitations: []Limitation{
			{Code: "association_not_causation", Message: "Contextual association must not be represented as proof of cause.", Scope: "causal_claim"},
			{Code: "intent_not_observed", Message: "Pilot intent and air traffic control intent are not observed by the available open-data inputs.", Scope: "intent"},
		},
		Explanations: []Explanation{{Code: "guard_decision", Message: "The guard authorizes only the claim strength supported by classified evidence."}},
		ScopeGuard:   ScopeGuardResearchOnly, Provenance: Provenance{EvidenceFingerprints: fingerprints, PolicyVersion: policy.Version}, EvaluatedAt: normalized.EvaluatedAt,
	}
	result.Provenance.InputFingerprint = resultFingerprint(result)
	if err := ValidateResult(result, policy); err != nil {
		return Result{}, err
	}
	return result.Clone(), nil
}

func normalize(request Request, policy Policy) (Request, error) {
	normalized := Request{SubjectID: strings.TrimSpace(request.SubjectID), ClaimKind: request.ClaimKind, ClaimText: strings.TrimSpace(request.ClaimText), EvidenceCompleteness: request.EvidenceCompleteness, EvaluatedAt: request.EvaluatedAt.UTC()}
	if normalized.SubjectID == "" || normalized.ClaimText == "" || !normalized.ClaimKind.IsKnown() || !unit(normalized.EvidenceCompleteness) || normalized.EvaluatedAt.IsZero() || len(request.Evidence) > policy.MaximumEvidenceCount {
		return Request{}, fmt.Errorf("invalid unknown intervention request")
	}
	seen := map[string]struct{}{}
	for _, input := range request.Evidence {
		evidence := input
		evidence.ID, evidence.Label, evidence.Source, evidence.Fingerprint, evidence.Limitation = strings.TrimSpace(evidence.ID), strings.TrimSpace(evidence.Label), strings.TrimSpace(evidence.Source), strings.TrimSpace(evidence.Fingerprint), strings.TrimSpace(evidence.Limitation)
		if evidence.ID == "" || evidence.Label == "" || !evidence.Class.IsKnown() || !unit(evidence.Score) || !strings.HasPrefix(evidence.Fingerprint, "sha256:") {
			return Request{}, fmt.Errorf("invalid evidence %q", evidence.ID)
		}
		if (evidence.Class == EvidenceUnknown || evidence.Class == EvidenceEstimated) && evidence.Limitation == "" {
			return Request{}, fmt.Errorf("estimated or unknown evidence requires limitation")
		}
		if _, exists := seen[evidence.ID]; exists {
			return Request{}, fmt.Errorf("duplicate evidence %q", evidence.ID)
		}
		seen[evidence.ID] = struct{}{}
		normalized.Evidence = append(normalized.Evidence, evidence)
	}
	sort.Slice(normalized.Evidence, func(i, j int) bool { return normalized.Evidence[i].ID < normalized.Evidence[j].ID })
	return normalized, nil
}

func resultFingerprint(result Result) string {
	copy := result.Clone()
	copy.Provenance.InputFingerprint = ""
	encoded, _ := json.Marshal(copy)
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}
