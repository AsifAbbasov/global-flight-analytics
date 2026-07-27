package historicalcontract

import (
	"math"
	"testing"
)

func TestProductionMetricCatalogMatchesMaterializableContract(t *testing.T) {
	if len(SupportedMetricNames()) != 16 {
		t.Fatalf("supported metrics = %d, want 16", len(SupportedMetricNames()))
	}
	reserved := []MetricName{
		MetricNamePeakActivity,
		MetricNameAverageActivity,
		MetricNameDataFreshness,
		MetricNameCoverageScore,
	}
	for _, name := range reserved {
		if _, exists := MetricSpecFor(name); exists {
			t.Fatalf("reserved metric %q is advertised as materializable", name)
		}
	}
}

func TestValidateRejectsMetricCatalogMismatch(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Result)
		code   string
	}{
		{name: "unit", mutate: func(result *Result) { result.Metric.Unit = "items" }, code: "metric_unit_mismatch"},
		{name: "aggregation", mutate: func(result *Result) { result.Metric.Aggregation = AggregationMaximum }, code: "metric_aggregation_mismatch"},
		{name: "scope", mutate: func(result *Result) { result.Scope = Scope{Type: ScopeTypeAirport, AirportICAOCode: "UBBB"} }, code: "metric_scope_mismatch"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validCompleteResult()
			test.mutate(&result)
			report := Validate(result)
			if report.Status != ValidationStatusInvalid || !reportHasCode(report, test.code) {
				t.Fatalf("unexpected report: %#v", report)
			}
		})
	}
}

func TestValidateRejectsFractionalCountValue(t *testing.T) {
	result := validCompleteResult()
	result.Points[0].Value = 1.5
	result.Summary = Summarize(result.Points)
	result.Comparison.CurrentValue = result.Summary.Total
	result.Comparison.AbsoluteChange = result.Comparison.CurrentValue - result.Comparison.PreviousValue
	percentage := result.Comparison.AbsoluteChange / result.Comparison.PreviousValue * 100
	result.Comparison.PercentageChange = &percentage
	result.Comparison.Direction = TrendDirectionForChange(result.Comparison.AbsoluteChange)
	report := Validate(result)
	if !reportHasCode(report, "bucket_count_value_not_integral") {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestValidateAcceptsZeroEventCompleteBucket(t *testing.T) {
	result := validCompleteResult()
	result.Points[0].Value = 0
	result.Points[0].SampleCount = 0
	result.Points[0].Confidence = Confidence{
		Score:       1,
		Level:       ConfidenceLevelHigh,
		SampleCount: 0,
		Reasons:     []ConfidenceReason{{Code: "complete_coverage", Message: "Source coverage was complete with no matching events.", Contribution: 1}},
	}
	result.Summary = Summarize(result.Points)
	result.Confidence.SampleCount = totalSampleCount(result.Points)
	result.Comparison.CurrentValue = result.Summary.Total
	result.Comparison.AbsoluteChange = result.Comparison.CurrentValue - result.Comparison.PreviousValue
	percentage := result.Comparison.AbsoluteChange / result.Comparison.PreviousValue * 100
	result.Comparison.PercentageChange = &percentage
	result.Comparison.Direction = TrendDirectionForChange(result.Comparison.AbsoluteChange)
	if report := Validate(result); report.Status != ValidationStatusValid {
		t.Fatalf("zero-event complete bucket must remain valid: %#v", report)
	}
}

func TestValidateRejectsUnavailableHighConfidence(t *testing.T) {
	result := validPartialResult()
	result.Points[0] = Point{
		StartTime:  result.Points[0].StartTime,
		EndTime:    result.Points[0].EndTime,
		Status:     BucketStatusUnavailable,
		Confidence: Confidence{Score: 1, Level: ConfidenceLevelHigh},
	}
	result.Summary = Summarize(result.Points)
	result.Confidence = Confidence{}
	report := Validate(result)
	if !reportHasCode(report, "unavailable_bucket_confidence_invalid") {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestValidateRejectsPartialSeriesWithoutRepresentedBucket(t *testing.T) {
	result := validPartialResult()
	result.Points[0] = Point{
		StartTime:  result.Points[0].StartTime,
		EndTime:    result.Points[0].EndTime,
		Status:     BucketStatusUnavailable,
		Confidence: Confidence{},
	}
	result.Summary = Summary{}
	result.Confidence = Confidence{}
	report := Validate(result)
	if !reportHasCode(report, "partial_series_without_represented_bucket") {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestValidateRequiresPartialAndUnavailableExplanations(t *testing.T) {
	partial := validPartialResult()
	partial.Limitations = nil
	partial.Points[0].Limitations = nil
	partialReport := Validate(partial)
	if !reportHasCode(partialReport, "partial_series_without_limitation") ||
		!reportHasCode(partialReport, "partial_bucket_without_limitation") {
		t.Fatalf("unexpected partial report: %#v", partialReport)
	}
	unavailable := validUnavailableResult()
	unavailable.Limitations = nil
	unavailableReport := Validate(unavailable)
	if !reportHasCode(unavailableReport, "unavailable_series_without_limitation") {
		t.Fatalf("unexpected unavailable report: %#v", unavailableReport)
	}
}

func TestValidateBindsComparisonToCurrentSummary(t *testing.T) {
	result := validCompleteResult()
	result.Comparison.CurrentValue = 999
	result.Comparison.PreviousValue = 500
	result.Comparison.AbsoluteChange = 499
	percentage := 99.8
	result.Comparison.PercentageChange = &percentage
	result.Comparison.Direction = TrendDirectionUp
	report := Validate(result)
	if !reportHasCode(report, "comparison_current_summary_mismatch") {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestValidateBindsConfidenceReasonsToScore(t *testing.T) {
	result := validCompleteResult()
	result.Confidence.Reasons[0].Contribution = 0.1
	report := Validate(result)
	if !reportHasCode(report, "confidence_reason_contribution_mismatch") {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestPublicNumericHelpersRejectNonFiniteInputs(t *testing.T) {
	if got := ConfidenceLevelForScore(math.Inf(1)); got != ConfidenceLevelNone {
		t.Fatalf("infinite confidence level = %q", got)
	}
	if got := ConfidenceLevelForScore(math.NaN()); got != ConfidenceLevelNone {
		t.Fatalf("NaN confidence level = %q", got)
	}
	if got := TrendDirectionForChange(math.NaN()); got != TrendDirectionUnavailable {
		t.Fatalf("NaN trend direction = %q", got)
	}
}

func TestSchemaCoversHistoricalResultSemanticFields(t *testing.T) {
	required := []string{
		"schema_version", "status", "points.confidence", "points.limitations",
		"comparison.previous_window", "comparison.current_value", "comparison.previous_value",
		"confidence.reasons", "limitations", "provenance.source_names",
	}
	for _, name := range required {
		if _, exists := DefinitionByName(name); !exists {
			t.Fatalf("schema definition %q is missing", name)
		}
	}
	if len(CurrentSchema().Definitions) != 46 {
		t.Fatalf("schema definitions = %d, want 46", len(CurrentSchema().Definitions))
	}
}
