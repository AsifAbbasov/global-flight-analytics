package weathertrust

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	Version            = "weather-trust-gate-v1"
	FingerprintVersion = "weather-trust-gate-fingerprint-v1"
)

type Decision string

const (
	DecisionBlocked Decision = "blocked"
	DecisionLimited Decision = "limited"
	DecisionAllowed Decision = "allowed"
)

func (decision Decision) IsKnown() bool {
	switch decision {
	case DecisionBlocked, DecisionLimited, DecisionAllowed:
		return true
	default:
		return false
	}
}

type UsageScope string

const (
	UsageScopeSurfaceContext        UsageScope = "surface_context"
	UsageScopeTrajectoryContext     UsageScope = "trajectory_context"
	UsageScopeProjectionUncertainty UsageScope = "projection_uncertainty"
)

func (scope UsageScope) IsKnown() bool {
	switch scope {
	case UsageScopeSurfaceContext,
		UsageScopeTrajectoryContext,
		UsageScopeProjectionUncertainty:
		return true
	default:
		return false
	}
}

type ComponentName string

const (
	ComponentContractConfidence    ComponentName = "contract_confidence"
	ComponentTemporalFreshness     ComponentName = "temporal_freshness"
	ComponentFeatureCompleteness   ComponentName = "feature_completeness"
	ComponentVerticalApplicability ComponentName = "vertical_applicability"
)

type Component struct {
	Name   ComponentName
	Score  float64
	Weight float64
}

type Notice struct {
	Code    string
	Message string
}

type Result struct {
	Version  string
	Decision Decision
	Usable   bool

	AsOfTime time.Time

	Score      float64
	Components []Component

	AllowedScopes []UsageScope
	Limitations   []Notice
	Explanations  []Notice

	InputFingerprint string
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Components = append([]Component(nil), result.Components...)
	cloned.AllowedScopes = append([]UsageScope(nil), result.AllowedScopes...)
	cloned.Limitations = append([]Notice(nil), result.Limitations...)
	cloned.Explanations = append([]Notice(nil), result.Explanations...)
	return cloned
}

var fingerprintPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func (result Result) Validate() error {
	if result.Version != Version || !result.Decision.IsKnown() {
		return fmt.Errorf("weather trust version or decision is invalid")
	}
	if result.AsOfTime.IsZero() {
		return fmt.Errorf("weather trust as-of time is required")
	}
	if !unitInterval(result.Score) {
		return fmt.Errorf("weather trust score must be between zero and one")
	}
	if len(result.Components) != 4 {
		return fmt.Errorf("weather trust result requires four score components")
	}

	expectedComponents := map[ComponentName]struct{}{
		ComponentContractConfidence:    {},
		ComponentTemporalFreshness:     {},
		ComponentFeatureCompleteness:   {},
		ComponentVerticalApplicability: {},
	}
	seenComponents := make(map[ComponentName]struct{}, len(result.Components))
	weightTotal := 0.0
	for _, component := range result.Components {
		if _, exists := expectedComponents[component.Name]; !exists {
			return fmt.Errorf("weather trust component name is invalid")
		}
		if _, exists := seenComponents[component.Name]; exists {
			return fmt.Errorf("weather trust component is duplicated")
		}
		seenComponents[component.Name] = struct{}{}
		if !unitInterval(component.Score) ||
			!finite(component.Weight) || component.Weight < 0 {
			return fmt.Errorf("weather trust component value is invalid")
		}
		weightTotal += component.Weight
	}
	if absolute(weightTotal-1) > 1e-9 {
		return fmt.Errorf("weather trust component weights must sum to one")
	}

	if !sort.SliceIsSorted(result.AllowedScopes, func(left int, right int) bool {
		return result.AllowedScopes[left] < result.AllowedScopes[right]
	}) {
		return fmt.Errorf("weather trust allowed scopes must be sorted")
	}
	seenScopes := make(map[UsageScope]struct{})
	for _, scope := range result.AllowedScopes {
		if !scope.IsKnown() {
			return fmt.Errorf("weather trust usage scope is invalid")
		}
		if _, exists := seenScopes[scope]; exists {
			return fmt.Errorf("weather trust usage scope is duplicated")
		}
		seenScopes[scope] = struct{}{}
	}

	for _, collection := range [][]Notice{result.Limitations, result.Explanations} {
		for _, notice := range collection {
			if strings.TrimSpace(notice.Code) == "" ||
				strings.TrimSpace(notice.Message) == "" {
				return fmt.Errorf("weather trust notice is invalid")
			}
		}
	}

	if !fingerprintPattern.MatchString(result.InputFingerprint) {
		return fmt.Errorf("weather trust input fingerprint is invalid")
	}

	switch result.Decision {
	case DecisionBlocked:
		if result.Usable || len(result.AllowedScopes) != 0 || len(result.Limitations) == 0 {
			return fmt.Errorf("blocked weather trust result must be unusable, scope-free, and limited")
		}
	case DecisionLimited:
		if !result.Usable || len(result.AllowedScopes) == 0 || len(result.Limitations) == 0 {
			return fmt.Errorf("limited weather trust result must remain usable with scopes and limitations")
		}
	case DecisionAllowed:
		if !result.Usable || len(result.AllowedScopes) == 0 {
			return fmt.Errorf("allowed weather trust result must be usable with at least one scope")
		}
	}
	if len(result.Explanations) == 0 {
		return fmt.Errorf("weather trust result requires an explanation")
	}
	return nil
}

func normalizeNotices(notices []Notice) []Notice {
	unique := make(map[string]Notice)
	for _, notice := range notices {
		code := strings.TrimSpace(notice.Code)
		message := strings.TrimSpace(notice.Message)
		if code == "" || message == "" {
			continue
		}
		key := code + "\x00" + message
		unique[key] = Notice{Code: code, Message: message}
	}
	keys := make([]string, 0, len(unique))
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]Notice, 0, len(keys))
	for _, key := range keys {
		result = append(result, unique[key])
	}
	return result
}

func normalizeScopes(scopes []UsageScope) []UsageScope {
	unique := make(map[UsageScope]struct{})
	for _, scope := range scopes {
		if scope.IsKnown() {
			unique[scope] = struct{}{}
		}
	}
	result := make([]UsageScope, 0, len(unique))
	for scope := range unique {
		result = append(result, scope)
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left] < result[right]
	})
	return result
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func unitInterval(value float64) bool {
	return finite(value) && value >= 0 && value <= 1
}

func clampUnit(value float64) float64 {
	switch {
	case !finite(value):
		return 0
	case value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}

func absolute(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
