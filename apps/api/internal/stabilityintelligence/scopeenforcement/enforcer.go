package scopeenforcement

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func Enforce(request Request, policy Policy) (Result, error) {
	if err := policy.Validate(); err != nil {
		return Result{}, err
	}
	normalized, err := normalize(request, policy)
	if err != nil {
		return Result{}, err
	}
	guardSet := map[string]bool{}
	for _, guard := range normalized.DeclaredGuards {
		guardSet[guard] = true
	}
	claimResults := make([]ClaimResult, 0, len(normalized.Claims))
	violations := []Violation{}
	metrics := Metrics{ClaimCount: len(normalized.Claims), GuardCount: len(normalized.DeclaredGuards)}
	fingerprints := []string{}
	overall := DecisionAllowed
	for _, claim := range normalized.Claims {
		item := ClaimResult{Claim: claim, Decision: DecisionAllowed}
		if !guardSet[claim.SourceGuard] || !policy.KnownGuards[claim.SourceGuard] {
			item.Violations = append(item.Violations, Violation{Code: "missing_or_unknown_source_guard", ClaimCode: claim.Code, Message: "The claim is not protected by a declared known research-only guard.", Blocking: true})
		}
		if claim.Scope == ScopeOperationalDecision || claim.Scope == ScopeAirTrafficControl || claim.Scope == ScopeFlightPlanning || claim.Scope == ScopeSafetyCritical {
			item.Violations = append(item.Violations, Violation{Code: "operational_scope_forbidden", ClaimCode: claim.Code, Message: "The requested scope is outside the research-only project boundary.", Blocking: true})
		}
		if claim.Strength == StrengthDirective || claim.Strength == StrengthCertain {
			item.Violations = append(item.Violations, Violation{Code: "directive_or_certainty_forbidden", ClaimCode: claim.Code, Message: "Directive or certainty language is forbidden for inferential open-data analytics.", Blocking: true})
		} else if claim.Strength == StrengthCausal {
			item.Violations = append(item.Violations, Violation{Code: "causal_claim_limited", ClaimCode: claim.Code, Message: "A causal statement must be reduced to contextual association unless independently proven.", Blocking: false})
		}
		for _, violation := range item.Violations {
			if violation.Blocking {
				item.Decision = DecisionBlocked
				break
			}
			item.Decision = DecisionLimited
		}
		switch item.Decision {
		case DecisionAllowed:
			metrics.AllowedCount++
		case DecisionLimited:
			metrics.LimitedCount++
			if overall == DecisionAllowed {
				overall = DecisionLimited
			}
		case DecisionBlocked:
			metrics.BlockedCount++
			overall = DecisionBlocked
		}
		violations = append(violations, item.Violations...)
		claimResults = append(claimResults, item)
		fingerprints = append(fingerprints, fingerprint(claim))
	}
	sort.Strings(fingerprints)
	status := ResultStatusComplete
	if overall != DecisionAllowed {
		status = ResultStatusLimited
	}
	result := Result{SchemaVersion: SchemaVersionV1, Status: status, SubjectID: normalized.SubjectID, Decision: overall, Claims: claimResults, Violations: violations, Metrics: metrics,
		Limitations: []Limitation{{Code: "enforcement_not_authorization", Message: "Passing this guard preserves research scope and does not authorize operational use.", Scope: "publication"}},
		ScopeGuard:  ScopeGuardResearchOnly, Provenance: Provenance{ClaimFingerprints: fingerprints, DeclaredGuards: append([]string(nil), normalized.DeclaredGuards...), PolicyVersion: policy.Version}, EvaluatedAt: normalized.EvaluatedAt}
	result.Provenance.InputFingerprint = resultFingerprint(result)
	if err := ValidateResult(result, policy); err != nil {
		return Result{}, err
	}
	return result.Clone(), nil
}

func normalize(request Request, policy Policy) (Request, error) {
	normalized := Request{SubjectID: strings.TrimSpace(request.SubjectID), EvaluatedAt: request.EvaluatedAt.UTC()}
	if normalized.SubjectID == "" || normalized.EvaluatedAt.IsZero() || len(request.Claims) > policy.MaximumClaimCount {
		return Request{}, fmt.Errorf("invalid scope enforcement request")
	}
	guardSeen := map[string]struct{}{}
	for _, value := range request.DeclaredGuards {
		guard := strings.TrimSpace(value)
		if guard == "" {
			return Request{}, fmt.Errorf("empty declared guard")
		}
		if _, exists := guardSeen[guard]; exists {
			return Request{}, fmt.Errorf("duplicate declared guard")
		}
		guardSeen[guard] = struct{}{}
		normalized.DeclaredGuards = append(normalized.DeclaredGuards, guard)
	}
	sort.Strings(normalized.DeclaredGuards)
	claimSeen := map[string]struct{}{}
	for _, input := range request.Claims {
		claim := input
		claim.Code, claim.Text, claim.Capability, claim.SourceGuard = strings.TrimSpace(claim.Code), strings.TrimSpace(claim.Text), strings.TrimSpace(claim.Capability), strings.TrimSpace(claim.SourceGuard)
		if claim.Code == "" || claim.Text == "" || claim.Capability == "" || claim.SourceGuard == "" || !claim.Scope.IsKnown() || !claim.Strength.IsKnown() {
			return Request{}, fmt.Errorf("invalid claim %q", claim.Code)
		}
		if _, exists := claimSeen[claim.Code]; exists {
			return Request{}, fmt.Errorf("duplicate claim %q", claim.Code)
		}
		claimSeen[claim.Code] = struct{}{}
		normalized.Claims = append(normalized.Claims, claim)
	}
	sort.Slice(normalized.Claims, func(i, j int) bool { return normalized.Claims[i].Code < normalized.Claims[j].Code })
	return normalized, nil
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
