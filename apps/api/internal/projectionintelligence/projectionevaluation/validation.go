package projectionevaluation

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func (policy EvaluationPolicy) Validate() error {
	if policy.Version != EvaluationPolicyVersion || !fingerprintPattern.MatchString(policy.InputFingerprint) {
		return fmt.Errorf("projection evaluation policy metadata is invalid")
	}
	config := Config{
		MaximumInterpolationGap:     policy.MaximumInterpolationGap,
		MaximumTruthGroundSpeedMPS:  policy.MaximumTruthGroundSpeedMPS,
		MaximumTruthVerticalRateMPS: policy.MaximumTruthVerticalRateMPS,
		MinimumEvaluatedPointCount:  policy.MinimumEvaluatedPointCount,
		MaximumHorizontalErrorM:     policy.MaximumHorizontalErrorM,
		MaximumAltitudeErrorM:       policy.MaximumAltitudeErrorM,
		LeadTimeBucketSize:          policy.LeadTimeBucketSize,
	}
	if err := config.Validate(); err != nil {
		return err
	}
	if evaluationPolicyFingerprint(policy) != policy.InputFingerprint {
		return fmt.Errorf("projection evaluation policy fingerprint is inconsistent")
	}
	return nil
}

func (result Result) Validate() error {
	if result.Version != Version || !result.Status.IsKnown() {
		return fmt.Errorf("projection evaluation version or status is invalid")
	}
	if strings.TrimSpace(result.TrajectoryID) == "" ||
		strings.TrimSpace(result.ProjectionMethod.Name) == "" ||
		strings.TrimSpace(result.ProjectionMethod.Version) == "" ||
		!result.ProjectionMethod.DecisionClass.IsKnown() {
		return fmt.Errorf("projection evaluation identity is invalid")
	}
	if result.ProjectionAsOfTime.IsZero() || result.ProjectionHorizonEndTime.IsZero() ||
		!result.ProjectionAsOfTime.Before(result.ProjectionHorizonEndTime) || result.ForecastStep <= 0 ||
		result.ProjectionGeneratedAt.IsZero() || result.EvaluatedAt.IsZero() ||
		result.ProjectionGeneratedAt.Before(result.ProjectionAsOfTime) ||
		result.EvaluatedAt.Before(result.ProjectionGeneratedAt) {
		return fmt.Errorf("projection evaluation timestamps are invalid")
	}
	if err := result.Policy.Validate(); err != nil {
		return fmt.Errorf("projection evaluation policy is invalid: %v", err)
	}
	fingerprints := []string{
		result.ProjectionInputFingerprint,
		result.ProjectionSnapshotFingerprint,
		result.TruthSnapshotFingerprint,
		result.EvaluationInputFingerprint,
		result.TruthEvidence.InputFingerprint,
	}
	for _, fingerprint := range fingerprints {
		if !fingerprintPattern.MatchString(fingerprint) {
			return fmt.Errorf("projection evaluation fingerprints are invalid")
		}
	}
	if result.TruthSnapshotFingerprint != result.TruthEvidence.InputFingerprint ||
		result.TruthEvidence.KnowledgeCutoffMode != TruthKnowledgeCutoffMode ||
		result.TruthEvidence.CanonicalPointCount < 0 ||
		result.TruthEvidence.ExcludedAfterObservationCutoffCount < 0 ||
		result.TruthEvidence.ExcludedAfterAvailabilityCutoffCount < 0 ||
		result.TruthEvidence.RejectedImplausibleInterpolationCount < 0 ||
		!sort.StringsAreSorted(result.TruthEvidence.SourceNames) {
		return fmt.Errorf("projection truth evidence summary is invalid")
	}

	for index, point := range result.Points {
		if err := validatePointEvaluation(result, point, index); err != nil {
			return err
		}
		if index > 0 && !result.Points[index-1].ForecastTime.Before(point.ForecastTime) {
			return fmt.Errorf("projection evaluation points are not time ordered")
		}
	}

	recomputedPosition := buildPositionMetrics(result.Position.ForecastPointCount, result.Points)
	if !samePositionMetrics(result.Position, recomputedPosition) {
		return fmt.Errorf("projection position metrics are inconsistent with evaluated points")
	}
	recomputedEndpoint := buildEndpointMetrics(result.ProjectionHorizonEndTime, result.Points)
	if !sameEndpointMetrics(result.Endpoint, recomputedEndpoint) {
		return fmt.Errorf("projection endpoint metrics are inconsistent with evaluated points")
	}
	recomputedConfidence := buildConfidenceMetrics(result.Points)
	if !sameConfidenceMetrics(result.Confidence, recomputedConfidence) {
		return fmt.Errorf("projection confidence metrics are inconsistent with evaluated points")
	}
	recomputedLeadTimes := buildLeadTimeMetrics(result.Points, result.Policy.LeadTimeBucketSize)
	if !sameLeadTimeMetrics(result.LeadTimes, recomputedLeadTimes) {
		return fmt.Errorf("projection lead-time metrics are inconsistent with evaluated points")
	}
	if err := validateArrivalMetrics(result.Arrival); err != nil {
		return err
	}

	switch result.Status {
	case StatusUnavailable:
		if len(result.Limitations) == 0 || result.Position.EvaluatedPointCount >= result.Policy.MinimumEvaluatedPointCount {
			return fmt.Errorf("unavailable evaluation status is inconsistent")
		}
	case StatusPartial:
		if len(result.Points) < result.Policy.MinimumEvaluatedPointCount {
			return fmt.Errorf("partial evaluation requires minimum usable truth")
		}
	case StatusComplete:
		if result.Position.ForecastPointCount == 0 || result.Position.EvaluatedPointCount != result.Position.ForecastPointCount {
			return fmt.Errorf("complete evaluation requires all forecast points")
		}
		if !((!result.Arrival.ActualTruthAvailable && !result.Arrival.PredictionAvailable) || result.Arrival.Available) {
			return fmt.Errorf("complete evaluation requires complete arrival comparability")
		}
	}
	for _, limitation := range result.Limitations {
		if strings.TrimSpace(limitation.Code) == "" || strings.TrimSpace(limitation.Message) == "" {
			return fmt.Errorf("projection evaluation limitation is invalid")
		}
	}
	return nil
}

func validatePointEvaluation(result Result, point PointEvaluation, index int) error {
	if point.Sequence < 0 || !point.ActualSource.IsKnown() || point.ForecastTime.IsZero() ||
		point.ActualTime.IsZero() || !point.ActualTime.Equal(point.ForecastTime) ||
		point.LeadTime <= 0 || point.LeadTime != point.ForecastTime.Sub(result.ProjectionAsOfTime) ||
		!validLatitude(point.ForecastLatitude) || !validLongitude(point.ForecastLongitude) ||
		!validLatitude(point.ActualLatitude) || !validLongitude(point.ActualLongitude) ||
		!positiveFinite(point.ForecastHorizontalUncertaintyM) ||
		!validConfidence(point.ForecastConfidence) {
		return fmt.Errorf("projection evaluation point is invalid at index %d", index)
	}
	expectedError := greatCircleDistanceM(point.ForecastLatitude, point.ForecastLongitude, point.ActualLatitude, point.ActualLongitude)
	expectedRatio := expectedError / result.Policy.MaximumHorizontalErrorM
	expectedWithin := expectedError <= point.ForecastHorizontalUncertaintyM
	expectedAccuracy := clampUnit(1 - expectedRatio)
	expectedGap := math.Abs(point.ForecastConfidence.Score - expectedAccuracy)
	if !almostEqual(point.HorizontalErrorM, expectedError) ||
		!almostEqual(point.HorizontalErrorRatio, expectedRatio) ||
		point.WithinHorizontalUncertainty != expectedWithin ||
		!almostEqual(point.NormalizedHorizontalAccuracy, expectedAccuracy) ||
		!almostEqual(point.AbsoluteConfidenceGap, expectedGap) {
		return fmt.Errorf("projection horizontal evaluation is inconsistent at index %d", index)
	}
	altitudeEvaluated := point.ForecastAltitudeM != nil && point.ActualAltitudeM != nil
	if !altitudeEvaluated {
		if point.AltitudeAbsoluteErrorM != nil || point.AltitudeErrorRatio != nil || point.WithinVerticalUncertainty != nil {
			return fmt.Errorf("projection altitude evaluation is inconsistent at index %d", index)
		}
		return nil
	}
	expectedAltitudeError := math.Abs(*point.ForecastAltitudeM - *point.ActualAltitudeM)
	if point.AltitudeAbsoluteErrorM == nil || point.AltitudeErrorRatio == nil ||
		!almostEqual(*point.AltitudeAbsoluteErrorM, expectedAltitudeError) ||
		!almostEqual(*point.AltitudeErrorRatio, expectedAltitudeError/result.Policy.MaximumAltitudeErrorM) {
		return fmt.Errorf("projection altitude evaluation is inconsistent at index %d", index)
	}
	if point.ForecastVerticalUncertaintyM == nil {
		if point.WithinVerticalUncertainty != nil {
			return fmt.Errorf("projection vertical coverage is inconsistent at index %d", index)
		}
	} else if point.WithinVerticalUncertainty == nil ||
		*point.WithinVerticalUncertainty != (expectedAltitudeError <= *point.ForecastVerticalUncertaintyM) {
		return fmt.Errorf("projection vertical coverage is inconsistent at index %d", index)
	}
	return nil
}

func validateArrivalMetrics(metrics ArrivalMetrics) error {
	if metrics.ActualTruthAvailable {
		if !airportICAOPattern.MatchString(metrics.ActualAirportICAOCode) || metrics.ActualBoundaryTime.IsZero() {
			return fmt.Errorf("actual arrival metrics are invalid")
		}
	} else if metrics.ActualAirportICAOCode != "" || !metrics.ActualBoundaryTime.IsZero() {
		return fmt.Errorf("actual arrival absence is inconsistent")
	}
	if metrics.PredictionAvailable {
		if !airportICAOPattern.MatchString(metrics.PredictedAirportICAOCode) || metrics.EarliestTime.IsZero() ||
			metrics.EstimatedTime.IsZero() || metrics.LatestTime.IsZero() ||
			metrics.EarliestTime.After(metrics.EstimatedTime) || metrics.EstimatedTime.After(metrics.LatestTime) {
			return fmt.Errorf("predicted arrival metrics are invalid")
		}
	} else if metrics.PredictedAirportICAOCode != "" || !metrics.EarliestTime.IsZero() ||
		!metrics.EstimatedTime.IsZero() || !metrics.LatestTime.IsZero() {
		return fmt.Errorf("predicted arrival absence is inconsistent")
	}
	expectedMatched := metrics.ActualTruthAvailable && metrics.PredictionAvailable &&
		metrics.ActualAirportICAOCode == metrics.PredictedAirportICAOCode
	if metrics.AirportMatched != expectedMatched || metrics.Available != expectedMatched {
		return fmt.Errorf("arrival match state is inconsistent")
	}
	if !metrics.Available {
		if metrics.EstimatedAbsoluteErrorSeconds != 0 || metrics.SignedErrorSeconds != 0 ||
			metrics.IntervalWidthSeconds != 0 || metrics.IntervalCoveredActual {
			return fmt.Errorf("unavailable arrival contains accuracy metrics")
		}
		return nil
	}
	expectedSigned := metrics.EstimatedTime.Sub(metrics.ActualBoundaryTime).Seconds()
	expectedWidth := metrics.LatestTime.Sub(metrics.EarliestTime).Seconds()
	expectedCovered := !metrics.ActualBoundaryTime.Before(metrics.EarliestTime) && !metrics.ActualBoundaryTime.After(metrics.LatestTime)
	if !almostEqual(metrics.SignedErrorSeconds, expectedSigned) ||
		!almostEqual(metrics.EstimatedAbsoluteErrorSeconds, math.Abs(expectedSigned)) ||
		!almostEqual(metrics.IntervalWidthSeconds, expectedWidth) || metrics.IntervalCoveredActual != expectedCovered {
		return fmt.Errorf("arrival derived metrics are inconsistent")
	}
	return nil
}

func (result AggregateResult) Validate() error {
	if result.Version != AggregateVersion || !result.Status.IsKnown() || result.EvaluationCount < 0 ||
		result.MethodCount != len(result.Methods) || !fingerprintPattern.MatchString(result.InputFingerprint) ||
		result.GeneratedAt.IsZero() {
		return fmt.Errorf("projection evaluation aggregate metadata is invalid")
	}
	if result.Status == StatusUnavailable && (len(result.Methods) != 0 || len(result.Limitations) == 0) {
		return fmt.Errorf("unavailable aggregate requires no methods and limitations")
	}
	if result.Status != StatusUnavailable && len(result.Methods) == 0 {
		return fmt.Errorf("available aggregate requires method summaries")
	}
	totalEvaluations := 0
	for index, method := range result.Methods {
		if !validMethodSummary(method) {
			return fmt.Errorf("projection evaluation method summary is invalid at index %d", index)
		}
		totalEvaluations += method.EvaluationCount
		if index > 0 && methodSummaryKey(result.Methods[index-1]) >= methodSummaryKey(method) {
			return fmt.Errorf("projection evaluation method summaries are not deterministically ordered")
		}
	}
	if totalEvaluations != result.EvaluationCount {
		return fmt.Errorf("aggregate evaluation count does not match method summaries")
	}
	for _, limitation := range result.Limitations {
		if strings.TrimSpace(limitation.Code) == "" || strings.TrimSpace(limitation.Message) == "" {
			return fmt.Errorf("projection evaluation aggregate limitation is invalid")
		}
	}
	return nil
}

func validMethodSummary(method MethodSummary) bool {
	if strings.TrimSpace(method.MethodName) == "" || strings.TrimSpace(method.MethodVersion) == "" ||
		!method.DecisionClass.IsKnown() || method.ProjectionHorizonDuration <= 0 || method.ForecastStep <= 0 ||
		method.EvaluationPolicyVersion != EvaluationPolicyVersion ||
		!fingerprintPattern.MatchString(method.EvaluationPolicyFingerprint) || method.LeadTimeBucketSize <= 0 ||
		method.EvaluationCount < 1 || method.CompleteEvaluationCount < 0 || method.PartialEvaluationCount < 0 ||
		method.UnavailableEvaluationCount < 0 || method.AccuracyEligibleEvaluationCount < 0 ||
		method.CompleteEvaluationCount+method.PartialEvaluationCount+method.UnavailableEvaluationCount != method.EvaluationCount ||
		method.AccuracyEligibleEvaluationCount != method.CompleteEvaluationCount+method.PartialEvaluationCount ||
		method.ForecastPointCount < 0 || method.EvaluatedPointCount < 0 || method.EvaluatedPointCount > method.ForecastPointCount ||
		method.AltitudeEvaluatedPointCount < 0 || method.AltitudeEvaluatedPointCount > method.EvaluatedPointCount ||
		method.EndpointEvaluationCount < 0 || method.EndpointEvaluationCount > method.AccuracyEligibleEvaluationCount ||
		method.EndpointAltitudeEvaluationCount < 0 || method.EndpointAltitudeEvaluationCount > method.EndpointEvaluationCount ||
		method.ConfidenceEvaluationPointCount != method.EvaluatedPointCount ||
		method.ActualArrivalTruthCount < 0 || method.ActualArrivalTruthCount > method.EvaluationCount ||
		method.ArrivalPredictionCount < 0 || method.ArrivalPredictionCount > method.EvaluationCount ||
		method.MatchedArrivalCount < 0 || method.MatchedArrivalCount > method.EvaluationCount ||
		method.ArrivalEvaluationCount < 0 || method.ArrivalEvaluationCount > method.MatchedArrivalCount ||
		method.ActualArrivalTruthCount != method.arrivalBothPlusMissing() ||
		method.ArrivalPredictionCount != method.arrivalBothPlusUnexpected() {
		return false
	}
	expectedPointCoverage := ratioOrZero(method.EvaluatedPointCount, method.ForecastPointCount)
	expectedArrivalRecall := ratioOrZero(method.arrivalBothCount(), method.ActualArrivalTruthCount)
	expectedAirportAccuracy := ratioOrZero(method.MatchedArrivalCount, method.arrivalBothCount())
	if !almostEqual(method.PointCoverageRatio, expectedPointCoverage) ||
		!almostEqual(method.ArrivalPredictionRecall, expectedArrivalRecall) ||
		!almostEqual(method.ArrivalAirportAccuracy, expectedAirportAccuracy) {
		return false
	}
	ratios := []float64{
		method.PointCoverageRatio, method.HorizontalUncertaintyCoverageRatio, method.VerticalUncertaintyCoverageRatio,
		method.MeanForecastConfidence, method.MeanNormalizedHorizontalAccuracy, method.MeanAbsoluteConfidenceGap,
		method.ArrivalPredictionRecall, method.ArrivalAirportAccuracy, method.ArrivalIntervalCoverageRatio,
	}
	for _, value := range ratios {
		if !unitInterval(value) {
			return false
		}
	}
	values := []float64{
		method.MeanHorizontalErrorM, method.MedianHorizontalErrorM, method.P95HorizontalErrorM, method.HorizontalRMSEM,
		method.TrajectoryMacroMeanHorizontalErrorM, method.TrajectoryMacroMeanHorizontalRMSEM,
		method.MeanAltitudeAbsoluteErrorM, method.AltitudeRMSEM,
		method.MeanEndpointHorizontalErrorM, method.EndpointHorizontalRMSEM, method.MeanEndpointAltitudeErrorM,
		method.ConfidenceCalibrationRMSE, method.MeanArrivalAbsoluteErrorSeconds,
	}
	for _, value := range values {
		if !nonNegativeFinite(value) {
			return false
		}
	}
	if !validLeadTimeMetrics(method.LeadTimes) {
		return false
	}
	return true
}

func ratioOrZero(numerator int, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func (method MethodSummary) arrivalBothCount() int {
	return method.MatchedArrivalCount + method.ArrivalAirportMismatchCount
}

func (method MethodSummary) arrivalBothPlusMissing() int {
	return method.arrivalBothCount() + method.MissingArrivalPredictionCount
}

func (method MethodSummary) arrivalBothPlusUnexpected() int {
	return method.arrivalBothCount() + method.UnexpectedArrivalPredictionCount
}

func validLeadTimeMetrics(items []LeadTimeMetrics) bool {
	for index, item := range items {
		if item.BucketStart < 0 || item.BucketEnd <= item.BucketStart || item.EvaluatedPointCount < 1 ||
			!nonNegativeFinite(item.MeanHorizontalErrorM) || !nonNegativeFinite(item.MedianHorizontalErrorM) ||
			!nonNegativeFinite(item.P95HorizontalErrorM) || !nonNegativeFinite(item.HorizontalRMSEM) ||
			!unitInterval(item.HorizontalUncertaintyCoverageRatio) || !unitInterval(item.MeanForecastConfidence) ||
			!unitInterval(item.MeanNormalizedHorizontalAccuracy) || !unitInterval(item.MeanAbsoluteConfidenceGap) {
			return false
		}
		if index > 0 && items[index-1].BucketStart >= item.BucketStart {
			return false
		}
	}
	return true
}

func samePositionMetrics(left, right PositionMetrics) bool {
	return left.ForecastPointCount == right.ForecastPointCount && left.EvaluatedPointCount == right.EvaluatedPointCount &&
		left.MissingActualPointCount == right.MissingActualPointCount && left.AltitudeEvaluatedPointCount == right.AltitudeEvaluatedPointCount &&
		allAlmostEqual(
			left.CoverageRatio, right.CoverageRatio,
			left.MeanHorizontalErrorM, right.MeanHorizontalErrorM,
			left.MedianHorizontalErrorM, right.MedianHorizontalErrorM,
			left.P95HorizontalErrorM, right.P95HorizontalErrorM,
			left.MaximumHorizontalErrorM, right.MaximumHorizontalErrorM,
			left.HorizontalRMSEM, right.HorizontalRMSEM,
			left.MeanHorizontalErrorRatio, right.MeanHorizontalErrorRatio,
			left.HorizontalUncertaintyCoverageRatio, right.HorizontalUncertaintyCoverageRatio,
			left.MeanAltitudeAbsoluteErrorM, right.MeanAltitudeAbsoluteErrorM,
			left.MedianAltitudeAbsoluteErrorM, right.MedianAltitudeAbsoluteErrorM,
			left.P95AltitudeAbsoluteErrorM, right.P95AltitudeAbsoluteErrorM,
			left.MaximumAltitudeAbsoluteErrorM, right.MaximumAltitudeAbsoluteErrorM,
			left.AltitudeRMSEM, right.AltitudeRMSEM,
			left.MeanAltitudeErrorRatio, right.MeanAltitudeErrorRatio,
			left.VerticalUncertaintyCoverageRatio, right.VerticalUncertaintyCoverageRatio,
		)
}

func sameEndpointMetrics(left, right EndpointMetrics) bool {
	return left.Available == right.Available && left.ForecastTime.Equal(right.ForecastTime) &&
		left.WithinHorizontalUncertainty == right.WithinHorizontalUncertainty &&
		left.AltitudeAvailable == right.AltitudeAvailable &&
		left.WithinVerticalUncertainty == right.WithinVerticalUncertainty &&
		allAlmostEqual(left.HorizontalErrorM, right.HorizontalErrorM, left.HorizontalErrorRatio, right.HorizontalErrorRatio,
			left.AltitudeAbsoluteErrorM, right.AltitudeAbsoluteErrorM, left.AltitudeErrorRatio, right.AltitudeErrorRatio)
}

func sameConfidenceMetrics(left, right ConfidenceMetrics) bool {
	return left.EvaluatedPointCount == right.EvaluatedPointCount && allAlmostEqual(
		left.MeanForecastConfidence, right.MeanForecastConfidence,
		left.MeanNormalizedHorizontalAccuracy, right.MeanNormalizedHorizontalAccuracy,
		left.MeanAbsoluteConfidenceGap, right.MeanAbsoluteConfidenceGap,
		left.ConfidenceCalibrationRMSE, right.ConfidenceCalibrationRMSE,
	)
}

func sameLeadTimeMetrics(left, right []LeadTimeMetrics) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].BucketStart != right[index].BucketStart || left[index].BucketEnd != right[index].BucketEnd ||
			left[index].EvaluatedPointCount != right[index].EvaluatedPointCount ||
			!allAlmostEqual(
				left[index].MeanHorizontalErrorM, right[index].MeanHorizontalErrorM,
				left[index].MedianHorizontalErrorM, right[index].MedianHorizontalErrorM,
				left[index].P95HorizontalErrorM, right[index].P95HorizontalErrorM,
				left[index].HorizontalRMSEM, right[index].HorizontalRMSEM,
				left[index].HorizontalUncertaintyCoverageRatio, right[index].HorizontalUncertaintyCoverageRatio,
				left[index].MeanForecastConfidence, right[index].MeanForecastConfidence,
				left[index].MeanNormalizedHorizontalAccuracy, right[index].MeanNormalizedHorizontalAccuracy,
				left[index].MeanAbsoluteConfidenceGap, right[index].MeanAbsoluteConfidenceGap,
			) {
			return false
		}
	}
	return true
}

func validConfidence(value projectioncontract.Confidence) bool {
	if !value.Level.IsKnown() || !unitInterval(value.Score) {
		return false
	}
	for _, reason := range value.Reasons {
		if strings.TrimSpace(reason.Code) == "" || strings.TrimSpace(reason.Message) == "" || !finite(reason.Contribution) {
			return false
		}
	}
	return true
}

func almostEqual(left, right float64) bool {
	if !finite(left) || !finite(right) {
		return false
	}
	scale := math.Max(1, math.Max(math.Abs(left), math.Abs(right)))
	return math.Abs(left-right) <= 1e-9*scale
}

func allAlmostEqual(values ...float64) bool {
	if len(values)%2 != 0 {
		return false
	}
	for index := 0; index < len(values); index += 2 {
		if !almostEqual(values[index], values[index+1]) {
			return false
		}
	}
	return true
}

func normalizeNotices(items []Notice) []Notice {
	seen := make(map[string]Notice, len(items))
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		message := strings.TrimSpace(item.Message)
		if code == "" || message == "" {
			continue
		}
		seen[code+"\x00"+message] = Notice{Code: code, Message: message}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]Notice, 0, len(keys))
	for _, key := range keys {
		result = append(result, seen[key])
	}
	return result
}

func cloneConfidence(value projectioncontract.Confidence) projectioncontract.Confidence {
	cloned := value
	cloned.Reasons = append([]projectioncontract.ConfidenceReason(nil), value.Reasons...)
	return cloned
}

func cloneFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
