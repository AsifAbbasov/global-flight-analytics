package formulabenchmark

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/researchbenchmark"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/researchdataset"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionevaluation"
)

var (
	ErrRequestInvalid = errors.New(
		"formula benchmark request is invalid",
	)
	ErrPlanKindInvalid = errors.New(
		"formula benchmark plan kind is invalid",
	)
	ErrDatasetMismatch = errors.New(
		"formula benchmark dataset does not match the plan",
	)
	ErrAggregateUnavailable = errors.New(
		"projection evaluation aggregate is unavailable",
	)
)

func DefaultPolicy() Policy {
	return Policy{
		Version: DefaultPolicyVersion,

		MinimumEvaluationCount: 30,

		MinimumCompleteEvaluationRatio: 0.80,
		MinimumPointCoverageRatio:      0.80,
		MinimumAltitudeEvaluationRatio: 0.25,
		MinimumArrivalEvaluationRatio:  0.25,

		MaximumMeanHorizontalErrorM: 20_000,
		MaximumP95HorizontalErrorM:  60_000,

		MinimumHorizontalUncertaintyCoverageRatio: 0.80,
		MaximumHorizontalUncertaintyCoverageRatio: 1.00,

		MaximumMeanAltitudeAbsoluteErrorM:       1_500,
		MinimumVerticalUncertaintyCoverageRatio: 0.75,
		MaximumVerticalUncertaintyCoverageRatio: 1.00,

		MaximumMeanArrivalAbsoluteErrorSeconds: 900,
		MinimumArrivalIntervalCoverageRatio:    0.75,
	}
}

func Evaluate(request Request) (Report, error) {
	plan, err := researchbenchmark.PlanByID(request.PlanID)
	if err != nil {
		return Report{}, fmt.Errorf(
			"%w: %v",
			ErrRequestInvalid,
			err,
		)
	}
	if plan.Kind !=
		researchbenchmark.KindProjectionFormulaEvaluation {
		return Report{}, fmt.Errorf(
			"%w: plan=%s kind=%s",
			ErrPlanKindInvalid,
			plan.ID,
			plan.Kind,
		)
	}
	if err := researchbenchmark.Validate(plan); err != nil {
		return Report{}, fmt.Errorf(
			"%w: validate plan: %v",
			ErrRequestInvalid,
			err,
		)
	}

	manifestDecision, err :=
		researchdataset.ValidateManifest(request.Manifest)
	if err != nil {
		return Report{}, fmt.Errorf(
			"%w: validate manifest: %v",
			ErrRequestInvalid,
			err,
		)
	}
	if !manifestDecision.Allowed {
		return Report{}, fmt.Errorf(
			"%w: manifest was not allowed",
			ErrRequestInvalid,
		)
	}
	if request.Manifest.DatasetID != plan.DatasetID {
		return Report{}, fmt.Errorf(
			"%w: plan=%s manifest=%s",
			ErrDatasetMismatch,
			plan.DatasetID,
			request.Manifest.DatasetID,
		)
	}

	if err := request.Policy.Validate(); err != nil {
		return Report{}, fmt.Errorf(
			"%w: %v",
			ErrRequestInvalid,
			err,
		)
	}
	if err := request.ProjectionAggregate.Validate(); err != nil {
		return Report{}, fmt.Errorf(
			"%w: validate projection aggregate: %v",
			ErrRequestInvalid,
			err,
		)
	}
	if request.ProjectionAggregate.Status ==
		projectionevaluation.StatusUnavailable {
		return Report{}, ErrAggregateUnavailable
	}
	if request.GeneratedAt.IsZero() ||
		request.GeneratedAt.Before(
			request.ProjectionAggregate.GeneratedAt,
		) {
		return Report{}, fmt.Errorf(
			"%w: generated time is invalid",
			ErrRequestInvalid,
		)
	}
	if int64(request.ProjectionAggregate.EvaluationCount) >
		plan.MaximumRecords ||
		int64(request.ProjectionAggregate.EvaluationCount) >
			request.Manifest.MaximumRecords {
		return Report{}, fmt.Errorf(
			"%w: evaluation count exceeds the bounded manifest or plan",
			ErrRequestInvalid,
		)
	}

	methodReports := make(
		[]MethodReport,
		0,
		len(request.ProjectionAggregate.Methods),
	)
	overallStatus := StatusBenchmarkPassed

	for _, method := range request.ProjectionAggregate.Methods {
		methodReport := evaluateMethod(method, request.Policy)
		methodReports = append(methodReports, methodReport)

		switch methodReport.Status {
		case StatusBenchmarkFailed:
			overallStatus = StatusBenchmarkFailed
		case StatusInsufficientEvidence:
			if overallStatus != StatusBenchmarkFailed {
				overallStatus =
					StatusInsufficientEvidence
			}
		}
	}

	methodReports = normalizeMethodReports(methodReports)

	report := Report{
		Version: Version,

		Status: overallStatus,

		PlanID:    plan.ID,
		DatasetID: plan.DatasetID,

		PolicyVersion: request.Policy.Version,

		ProjectionAggregateFingerprint: request.ProjectionAggregate.InputFingerprint,

		EvaluationCount: request.ProjectionAggregate.EvaluationCount,
		MethodCount:     len(methodReports),
		Methods:         methodReports,

		CalibrationAllowed:             false,
		AutomaticFormulaChangesAllowed: false,
		ManualReviewRequired:           true,

		MaximumClaim: MaximumClaim,

		Limitations: []Limitation{
			{
				Code:    "offline_dataset_scope_only",
				Message: "The report applies only to the bounded offline dataset manifest supplied to this run.",
			},
			{
				Code:    "thresholds_are_release_gates_not_scientific_constants",
				Message: "Policy thresholds are project release gates and are not universal scientific calibration constants.",
			},
			{
				Code:    "automatic_calibration_prohibited",
				Message: "The benchmark cannot change production formulas, weights, or uncertainty intervals automatically.",
			},
			{
				Code:    "manual_review_required",
				Message: "A passing report permits manual engineering review only and does not authorize a production formula change.",
			},
		},

		GeneratedAt: request.GeneratedAt.UTC(),
	}
	report.InputFingerprint = fingerprintRequest(request)

	if err := report.Validate(); err != nil {
		return Report{}, fmt.Errorf(
			"%w: validate report: %v",
			ErrRequestInvalid,
			err,
		)
	}

	return report, nil
}

func evaluateMethod(
	method projectionevaluation.MethodSummary,
	policy Policy,
) MethodReport {
	completeRatio := ratio(
		method.CompleteEvaluationCount,
		method.EvaluationCount,
	)
	altitudeRatio := ratio(
		method.AltitudeEvaluatedPointCount,
		method.EvaluatedPointCount,
	)
	arrivalRatio := ratio(
		method.ArrivalEvaluationCount,
		method.EvaluationCount,
	)

	checks := []MetricCheck{
		minimumCheck(
			"minimum_evaluation_count",
			float64(method.EvaluationCount),
			float64(policy.MinimumEvaluationCount),
			true,
			"Evaluation count must satisfy the minimum evidence gate.",
		),
		minimumCheck(
			"minimum_complete_evaluation_ratio",
			completeRatio,
			policy.MinimumCompleteEvaluationRatio,
			true,
			"Complete evaluation coverage must satisfy the minimum evidence gate.",
		),
		minimumCheck(
			"minimum_point_coverage_ratio",
			method.PointCoverageRatio,
			policy.MinimumPointCoverageRatio,
			true,
			"Projection point truth coverage must satisfy the minimum evidence gate.",
		),
		minimumCheck(
			"minimum_altitude_evaluation_ratio",
			altitudeRatio,
			policy.MinimumAltitudeEvaluationRatio,
			true,
			"Altitude evidence coverage must satisfy the minimum evidence gate.",
		),
		minimumCheck(
			"minimum_arrival_evaluation_ratio",
			arrivalRatio,
			policy.MinimumArrivalEvaluationRatio,
			true,
			"Arrival evidence coverage must satisfy the minimum evidence gate.",
		),
		maximumCheck(
			"maximum_mean_horizontal_error_m",
			method.MeanHorizontalErrorM,
			policy.MaximumMeanHorizontalErrorM,
			false,
			"Mean horizontal error must remain within the release gate.",
		),
		maximumCheck(
			"maximum_p95_horizontal_error_m",
			method.P95HorizontalErrorM,
			policy.MaximumP95HorizontalErrorM,
			false,
			"P95 horizontal error must remain within the release gate.",
		),
		rangeCheck(
			"horizontal_uncertainty_coverage_ratio",
			method.HorizontalUncertaintyCoverageRatio,
			policy.MinimumHorizontalUncertaintyCoverageRatio,
			policy.MaximumHorizontalUncertaintyCoverageRatio,
			false,
			"Horizontal uncertainty coverage must remain inside the release gate range.",
		),
		maximumCheck(
			"maximum_mean_altitude_absolute_error_m",
			method.MeanAltitudeAbsoluteErrorM,
			policy.MaximumMeanAltitudeAbsoluteErrorM,
			false,
			"Mean altitude error must remain within the release gate.",
		),
		rangeCheck(
			"vertical_uncertainty_coverage_ratio",
			method.VerticalUncertaintyCoverageRatio,
			policy.MinimumVerticalUncertaintyCoverageRatio,
			policy.MaximumVerticalUncertaintyCoverageRatio,
			false,
			"Vertical uncertainty coverage must remain inside the release gate range.",
		),
		maximumCheck(
			"maximum_mean_arrival_absolute_error_seconds",
			method.MeanArrivalAbsoluteErrorSeconds,
			policy.MaximumMeanArrivalAbsoluteErrorSeconds,
			false,
			"Mean arrival-time error must remain within the release gate.",
		),
		minimumCheck(
			"minimum_arrival_interval_coverage_ratio",
			method.ArrivalIntervalCoverageRatio,
			policy.MinimumArrivalIntervalCoverageRatio,
			false,
			"Arrival interval coverage must satisfy the release gate.",
		),
	}

	status := StatusBenchmarkPassed
	for _, check := range checks {
		if check.Status == CheckStatusPassed {
			continue
		}
		if check.EvidenceGate {
			if status != StatusBenchmarkFailed {
				status = StatusInsufficientEvidence
			}
			continue
		}
		status = StatusBenchmarkFailed
	}

	return MethodReport{
		MethodName:    method.MethodName,
		MethodVersion: method.MethodVersion,
		DecisionClass: string(method.DecisionClass),

		Status: status,

		EvaluationCount: method.EvaluationCount,

		CompleteEvaluationRatio: completeRatio,
		PointCoverageRatio:      method.PointCoverageRatio,
		AltitudeEvaluationRatio: altitudeRatio,
		ArrivalEvaluationRatio:  arrivalRatio,

		MeanHorizontalErrorM: method.MeanHorizontalErrorM,
		P95HorizontalErrorM:  method.P95HorizontalErrorM,

		HorizontalUncertaintyCoverageRatio: method.HorizontalUncertaintyCoverageRatio,

		MeanAltitudeAbsoluteErrorM:       method.MeanAltitudeAbsoluteErrorM,
		VerticalUncertaintyCoverageRatio: method.VerticalUncertaintyCoverageRatio,

		MeanArrivalAbsoluteErrorSeconds: method.MeanArrivalAbsoluteErrorSeconds,
		ArrivalIntervalCoverageRatio:    method.ArrivalIntervalCoverageRatio,

		Checks: checks,
	}
}

func minimumCheck(
	code string,
	actual float64,
	minimum float64,
	evidenceGate bool,
	message string,
) MetricCheck {
	status := CheckStatusFailed
	if actual >= minimum {
		status = CheckStatusPassed
	}
	return MetricCheck{
		Code: code,

		Status: status,

		ActualValue:  actual,
		MinimumValue: floatPointer(minimum),

		EvidenceGate: evidenceGate,

		Message: message,
	}
}

func maximumCheck(
	code string,
	actual float64,
	maximum float64,
	evidenceGate bool,
	message string,
) MetricCheck {
	status := CheckStatusFailed
	if actual <= maximum {
		status = CheckStatusPassed
	}
	return MetricCheck{
		Code: code,

		Status: status,

		ActualValue:  actual,
		MaximumValue: floatPointer(maximum),

		EvidenceGate: evidenceGate,

		Message: message,
	}
}

func rangeCheck(
	code string,
	actual float64,
	minimum float64,
	maximum float64,
	evidenceGate bool,
	message string,
) MetricCheck {
	status := CheckStatusFailed
	if actual >= minimum && actual <= maximum {
		status = CheckStatusPassed
	}
	return MetricCheck{
		Code: code,

		Status: status,

		ActualValue:  actual,
		MinimumValue: floatPointer(minimum),
		MaximumValue: floatPointer(maximum),

		EvidenceGate: evidenceGate,

		Message: message,
	}
}

func fingerprintRequest(request Request) string {
	canonical := struct {
		Version string `json:"version"`

		PlanID string `json:"plan_id"`

		Manifest researchdataset.Manifest `json:"manifest"`

		ProjectionAggregate projectionevaluation.AggregateResult `json:"projection_aggregate"`

		Policy Policy `json:"policy"`

		GeneratedAt string `json:"generated_at"`
	}{
		Version: FingerprintVersion,

		PlanID: strings.TrimSpace(request.PlanID),

		Manifest: request.Manifest,

		ProjectionAggregate: request.ProjectionAggregate,

		Policy: request.Policy,

		GeneratedAt: request.GeneratedAt.
			UTC().
			Format(
				"2006-01-02T15:04:05.000000000Z",
			),
	}

	encoded, err := json.Marshal(canonical)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func fingerprintValid(value string) bool {
	normalized := strings.TrimSpace(value)
	if !strings.HasPrefix(normalized, "sha256:") {
		return false
	}
	decoded, err := hex.DecodeString(
		strings.TrimPrefix(normalized, "sha256:"),
	)
	return err == nil && len(decoded) == sha256.Size
}

func floatPointer(value float64) *float64 {
	return &value
}
