package failureexplanation

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/confidencepropagation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/forecastanalysis"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/forecaststability"
)

func SignalsFromDecisionStability(result forecaststability.StabilityResult) []Signal {
	severity := SeverityInformation
	blocks := false
	switch result.Level {
	case forecaststability.StabilityLevelChanged:
		severity = SeverityWarning
	case forecaststability.StabilityLevelMaterialChange,
		forecaststability.StabilityLevelIndeterminate:
		severity = SeverityBlocking
		blocks = true
	}
	classification := CauseClassificationDerivedCondition
	if result.Level == forecaststability.StabilityLevelIndeterminate {
		classification = CauseClassificationUnknownCause
	}
	return []Signal{{
		Code:                 "decision_stability_" + string(result.Level),
		Category:             CategoryStability,
		Severity:             severity,
		Classification:       classification,
		Summary:              fmt.Sprintf("Decision stability level is %s.", result.Level),
		Detail:               "The comparison describes forecast decision change and does not identify operational cause.",
		Source:               forecaststability.Version,
		BlocksUse:            blocks,
		EvidenceFingerprints: []string{result.Provenance.InputFingerprint},
	}}
}

func SignalsFromForecastAnalysis(result forecastanalysis.Result) []Signal {
	severity := SeverityInformation
	blocks := false
	if result.Health == forecastanalysis.HealthWatch {
		severity = SeverityWarning
	}
	if result.Health == forecastanalysis.HealthUnstable || result.Health == forecastanalysis.HealthInsufficient {
		severity = SeverityBlocking
		blocks = true
	}
	classification := CauseClassificationDerivedCondition
	if result.Health == forecastanalysis.HealthInsufficient {
		classification = CauseClassificationUnknownCause
	}
	return []Signal{{
		Code:                 "forecast_history_" + string(result.Health),
		Category:             CategoryStability,
		Severity:             severity,
		Classification:       classification,
		Summary:              fmt.Sprintf("Forecast history health is %s.", result.Health),
		Detail:               fmt.Sprintf("Observed forecast-version trend is %s.", result.Trend),
		Source:               forecastanalysis.Version,
		BlocksUse:            blocks,
		EvidenceFingerprints: []string{result.Provenance.InputFingerprint},
	}}
}

func SignalsFromConfidencePropagation(result confidencepropagation.Result) []Signal {
	severity := SeverityInformation
	blocks := false
	if result.Level == "low" || result.Status == confidencepropagation.ResultStatusLimited {
		severity = SeverityWarning
	}
	if result.Score == 0 {
		severity = SeverityBlocking
		blocks = true
	}
	return []Signal{{
		Code:                 "propagated_confidence_" + result.Level,
		Category:             CategoryConfidence,
		Severity:             severity,
		Classification:       CauseClassificationDerivedCondition,
		Summary:              fmt.Sprintf("Propagated confidence level is %s.", result.Level),
		Detail:               "The final score is limited by required dependency evidence.",
		Source:               confidencepropagation.Version,
		BlocksUse:            blocks,
		EvidenceFingerprints: []string{result.Provenance.InputFingerprint},
	}}
}
