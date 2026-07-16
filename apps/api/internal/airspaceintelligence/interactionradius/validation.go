package interactionradius

import (
	"fmt"
	"strings"
)

type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

type ValidationIssue struct {
	Path    string
	Message string
}

type ValidationReport struct {
	Status ValidationStatus
	Issues []ValidationIssue
}

func Validate(decision Decision) ValidationReport {
	issues := make([]ValidationIssue, 0)
	add := func(path string, message string) {
		issues = append(issues, ValidationIssue{Path: path, Message: message})
	}

	if decision.SchemaVersion != SchemaVersionV1 {
		add("schema_version", "schema version is invalid")
	}
	if !decision.Status.IsKnown() {
		add("status", "decision status is invalid")
	}
	if strings.TrimSpace(decision.RegionCode) == "" {
		add("region_code", "region code is required")
	}
	if strings.TrimSpace(decision.NodeID) == "" {
		add("node_id", "node identifier is required")
	}
	if !decision.MotionClass.IsKnown() {
		add("motion_class", "motion class is invalid")
	}
	if decision.AsOfTime.IsZero() || decision.GeneratedAt.IsZero() ||
		decision.GeneratedAt.Before(decision.AsOfTime) {
		add("times", "as-of and generated-at times are invalid")
	}
	if decision.MaximumObservationAge <= 0 ||
		decision.MaximumPairTimeDifference <= 0 ||
		decision.LookaheadDuration <= 0 {
		add("policy_boundaries", "published policy boundaries must be positive")
	}
	if decision.Status == DecisionStatusBlocked {
		if decision.HorizontalRadiusKilometers != 0 ||
			decision.VerticalRadiusMeters != 0 {
			add("radii", "blocked decisions must publish zero radii")
		}
	} else if !positiveFinite(decision.HorizontalRadiusKilometers) ||
		!positiveFinite(decision.VerticalRadiusMeters) {
		add("radii", "usable decisions must publish positive finite radii")
	}
	if len(decision.Components) != 4 {
		add("components", "exactly four confidence components are required")
	}
	weightTotal := 0.0
	for index, component := range decision.Components {
		if strings.TrimSpace(component.Name) == "" ||
			!unitInterval(component.Score) ||
			!unitInterval(component.Weight) {
			add(fmt.Sprintf("components[%d]", index), "component is invalid")
		}
		weightTotal += component.Weight
	}
	if absolute(weightTotal-1) > 1e-9 {
		add("components", "component weights must sum to one")
	}
	if !unitInterval(decision.Confidence.Score) ||
		!decision.Confidence.Level.IsKnown() ||
		len(decision.Confidence.Reasons) == 0 {
		add("confidence", "confidence is invalid")
	}
	if len(decision.Limitations) == 0 || len(decision.Explanations) == 0 {
		add("explainability", "limitations and explanations are required")
	}
	if decision.ScopeGuard != ScopeGuardResearchOnly {
		add("scope_guard", "research-only scope guard is required")
	}
	if len(decision.Provenance.SourceNames) == 0 ||
		strings.TrimSpace(decision.Provenance.SourceNames[0]) == "" ||
		decision.Provenance.ObservedAt.IsZero() ||
		decision.Provenance.ObservedAt.After(decision.AsOfTime) {
		add("provenance", "provenance is invalid")
	}
	if !validFingerprint(decision.Provenance.InputFingerprint) {
		add("provenance.input_fingerprint", fingerprintIssue("provenance.input_fingerprint"))
	}
	if decision.VerticalFilteringPermitted &&
		decision.Status == DecisionStatusLimited &&
		decision.VerticalRadiusMeters <= 0 {
		add("vertical_filtering", "limited vertical filtering requires a positive radius")
	}

	status := ValidationStatusValid
	if len(issues) > 0 {
		status = ValidationStatusInvalid
	}
	return ValidationReport{Status: status, Issues: issues}
}

func absolute(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
