// Package formulabenchmark converts validated offline projection evaluation
// aggregates into reproducible benchmark decisions. It never changes
// production formula configuration automatically.
package formulabenchmark

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/researchdataset"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionevaluation"
)

const (
	Version = "formula-benchmark-report-v1"

	FingerprintVersion = "formula-benchmark-input-fingerprint-v1"

	DefaultPolicyVersion = "projection-formula-release-gate-v1"

	MaximumClaim = "bounded_offline_benchmark_evidence_only"
)

type Status string

const (
	StatusInsufficientEvidence Status = "insufficient_evidence"
	StatusBenchmarkFailed      Status = "benchmark_failed"
	StatusBenchmarkPassed      Status = "benchmark_passed"
)

func (status Status) IsKnown() bool {
	switch status {
	case StatusInsufficientEvidence,
		StatusBenchmarkFailed,
		StatusBenchmarkPassed:
		return true
	default:
		return false
	}
}

type CheckStatus string

const (
	CheckStatusPassed CheckStatus = "passed"
	CheckStatusFailed CheckStatus = "failed"
)

func (status CheckStatus) IsKnown() bool {
	switch status {
	case CheckStatusPassed, CheckStatusFailed:
		return true
	default:
		return false
	}
}

type Policy struct {
	Version string `json:"version"`

	MinimumEvaluationCount int `json:"minimum_evaluation_count"`

	MinimumCompleteEvaluationRatio float64 `json:"minimum_complete_evaluation_ratio"`
	MinimumPointCoverageRatio      float64 `json:"minimum_point_coverage_ratio"`
	MinimumAltitudeEvaluationRatio float64 `json:"minimum_altitude_evaluation_ratio"`
	MinimumArrivalEvaluationRatio  float64 `json:"minimum_arrival_evaluation_ratio"`

	MaximumMeanHorizontalErrorM float64 `json:"maximum_mean_horizontal_error_m"`
	MaximumP95HorizontalErrorM  float64 `json:"maximum_p95_horizontal_error_m"`

	MinimumHorizontalUncertaintyCoverageRatio float64 `json:"minimum_horizontal_uncertainty_coverage_ratio"`
	MaximumHorizontalUncertaintyCoverageRatio float64 `json:"maximum_horizontal_uncertainty_coverage_ratio"`

	MaximumMeanAltitudeAbsoluteErrorM       float64 `json:"maximum_mean_altitude_absolute_error_m"`
	MinimumVerticalUncertaintyCoverageRatio float64 `json:"minimum_vertical_uncertainty_coverage_ratio"`
	MaximumVerticalUncertaintyCoverageRatio float64 `json:"maximum_vertical_uncertainty_coverage_ratio"`

	MaximumMeanArrivalAbsoluteErrorSeconds float64 `json:"maximum_mean_arrival_absolute_error_seconds"`
	MinimumArrivalIntervalCoverageRatio    float64 `json:"minimum_arrival_interval_coverage_ratio"`
}

func (policy Policy) Validate() error {
	if strings.TrimSpace(policy.Version) == "" ||
		policy.MinimumEvaluationCount < 1 ||
		!unitInterval(policy.MinimumCompleteEvaluationRatio) ||
		!unitInterval(policy.MinimumPointCoverageRatio) ||
		!unitInterval(policy.MinimumAltitudeEvaluationRatio) ||
		!unitInterval(policy.MinimumArrivalEvaluationRatio) ||
		!positiveFinite(policy.MaximumMeanHorizontalErrorM) ||
		!positiveFinite(policy.MaximumP95HorizontalErrorM) ||
		policy.MaximumP95HorizontalErrorM <
			policy.MaximumMeanHorizontalErrorM ||
		!coverageRange(
			policy.MinimumHorizontalUncertaintyCoverageRatio,
			policy.MaximumHorizontalUncertaintyCoverageRatio,
		) ||
		!positiveFinite(
			policy.MaximumMeanAltitudeAbsoluteErrorM,
		) ||
		!coverageRange(
			policy.MinimumVerticalUncertaintyCoverageRatio,
			policy.MaximumVerticalUncertaintyCoverageRatio,
		) ||
		!positiveFinite(
			policy.MaximumMeanArrivalAbsoluteErrorSeconds,
		) ||
		!unitInterval(
			policy.MinimumArrivalIntervalCoverageRatio,
		) {
		return fmt.Errorf("formula benchmark policy is invalid")
	}

	return nil
}

type Request struct {
	PlanID string `json:"plan_id"`

	Manifest researchdataset.Manifest `json:"manifest"`

	ProjectionAggregate projectionevaluation.AggregateResult `json:"projection_aggregate"`

	Policy Policy `json:"policy"`

	GeneratedAt time.Time `json:"generated_at"`
}

type MetricCheck struct {
	Code string `json:"code"`

	Status CheckStatus `json:"status"`

	ActualValue float64 `json:"actual_value"`

	MinimumValue *float64 `json:"minimum_value,omitempty"`
	MaximumValue *float64 `json:"maximum_value,omitempty"`

	EvidenceGate bool `json:"evidence_gate"`

	Message string `json:"message"`
}

type MethodReport struct {
	MethodName    string `json:"method_name"`
	MethodVersion string `json:"method_version"`
	DecisionClass string `json:"decision_class"`

	Status Status `json:"status"`

	EvaluationCount int `json:"evaluation_count"`

	CompleteEvaluationRatio float64 `json:"complete_evaluation_ratio"`
	PointCoverageRatio      float64 `json:"point_coverage_ratio"`
	AltitudeEvaluationRatio float64 `json:"altitude_evaluation_ratio"`
	ArrivalEvaluationRatio  float64 `json:"arrival_evaluation_ratio"`

	MeanHorizontalErrorM float64 `json:"mean_horizontal_error_m"`
	P95HorizontalErrorM  float64 `json:"p95_horizontal_error_m"`

	HorizontalUncertaintyCoverageRatio float64 `json:"horizontal_uncertainty_coverage_ratio"`

	MeanAltitudeAbsoluteErrorM       float64 `json:"mean_altitude_absolute_error_m"`
	VerticalUncertaintyCoverageRatio float64 `json:"vertical_uncertainty_coverage_ratio"`

	MeanArrivalAbsoluteErrorSeconds float64 `json:"mean_arrival_absolute_error_seconds"`
	ArrivalIntervalCoverageRatio    float64 `json:"arrival_interval_coverage_ratio"`

	Checks []MetricCheck `json:"checks"`
}

type Limitation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Report struct {
	Version string `json:"version"`

	Status Status `json:"status"`

	PlanID    string             `json:"plan_id"`
	DatasetID researchdataset.ID `json:"dataset_id"`

	PolicyVersion string `json:"policy_version"`

	ProjectionAggregateFingerprint string `json:"projection_aggregate_fingerprint"`
	InputFingerprint               string `json:"input_fingerprint"`

	EvaluationCount int `json:"evaluation_count"`
	MethodCount     int `json:"method_count"`

	Methods []MethodReport `json:"methods"`

	CalibrationAllowed             bool `json:"calibration_allowed"`
	AutomaticFormulaChangesAllowed bool `json:"automatic_formula_changes_allowed"`
	ManualReviewRequired           bool `json:"manual_review_required"`

	MaximumClaim string `json:"maximum_claim"`

	Limitations []Limitation `json:"limitations"`

	GeneratedAt time.Time `json:"generated_at"`
}

func (report Report) Validate() error {
	if report.Version != Version ||
		!report.Status.IsKnown() ||
		strings.TrimSpace(report.PlanID) == "" ||
		report.DatasetID == "" ||
		strings.TrimSpace(report.PolicyVersion) == "" ||
		!fingerprintValid(
			report.ProjectionAggregateFingerprint,
		) ||
		!fingerprintValid(report.InputFingerprint) ||
		report.EvaluationCount < 0 ||
		report.MethodCount != len(report.Methods) ||
		report.CalibrationAllowed ||
		report.AutomaticFormulaChangesAllowed ||
		!report.ManualReviewRequired ||
		report.MaximumClaim != MaximumClaim ||
		report.GeneratedAt.IsZero() {
		return fmt.Errorf(
			"formula benchmark report metadata is invalid",
		)
	}

	for index, method := range report.Methods {
		if strings.TrimSpace(method.MethodName) == "" ||
			strings.TrimSpace(method.MethodVersion) == "" ||
			strings.TrimSpace(method.DecisionClass) == "" ||
			!method.Status.IsKnown() ||
			method.EvaluationCount < 1 ||
			len(method.Checks) == 0 {
			return fmt.Errorf(
				"formula benchmark method report is invalid at index %d",
				index,
			)
		}
		if index > 0 {
			previous := report.Methods[index-1]
			previousKey := previous.MethodName +
				"\x00" +
				previous.MethodVersion
			currentKey := method.MethodName +
				"\x00" +
				method.MethodVersion
			if previousKey >= currentKey {
				return fmt.Errorf(
					"formula benchmark methods are not deterministically ordered",
				)
			}
		}
		for _, check := range method.Checks {
			if strings.TrimSpace(check.Code) == "" ||
				!check.Status.IsKnown() ||
				strings.TrimSpace(check.Message) == "" ||
				!finite(check.ActualValue) {
				return fmt.Errorf(
					"formula benchmark check is invalid",
				)
			}
		}
	}

	if len(report.Limitations) == 0 {
		return fmt.Errorf(
			"formula benchmark report requires limitations",
		)
	}
	for _, limitation := range report.Limitations {
		if strings.TrimSpace(limitation.Code) == "" ||
			strings.TrimSpace(limitation.Message) == "" {
			return fmt.Errorf(
				"formula benchmark limitation is invalid",
			)
		}
	}

	return nil
}

func normalizeMethodReports(items []MethodReport) []MethodReport {
	result := append([]MethodReport(nil), items...)
	sort.Slice(
		result,
		func(left, right int) bool {
			leftKey := result[left].MethodName +
				"\x00" +
				result[left].MethodVersion
			rightKey := result[right].MethodName +
				"\x00" +
				result[right].MethodVersion
			return leftKey < rightKey
		},
	)
	return result
}

func ratio(numerator int, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func unitInterval(value float64) bool {
	return finite(value) && value >= 0 && value <= 1
}

func coverageRange(minimum float64, maximum float64) bool {
	return unitInterval(minimum) &&
		unitInterval(maximum) &&
		minimum <= maximum
}

func positiveFinite(value float64) bool {
	return finite(value) && value > 0
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}
