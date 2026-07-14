package historicalcontract

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestValidateAcceptsCompleteHistoricalSeries(
	t *testing.T,
) {
	result := validCompleteResult()
	report := Validate(result)

	if report.Version != ValidationVersion ||
		report.Status !=
			ValidationStatusValid ||
		report.ErrorCount != 0 ||
		report.WarningCount != 0 ||
		len(report.Issues) != 0 {
		t.Fatalf(
			"unexpected validation report: %#v",
			report,
		)
	}
}

func TestValidateAcceptsUnavailableSeries(
	t *testing.T,
) {
	result := validUnavailableResult()
	report := Validate(result)

	if report.Status !=
		ValidationStatusValid ||
		report.ErrorCount != 0 {
		t.Fatalf(
			"unexpected validation report: %#v",
			report,
		)
	}
}

func TestValidateAcceptsPartialSeries(
	t *testing.T,
) {
	result := validPartialResult()
	report := Validate(result)

	if report.Status !=
		ValidationStatusValid ||
		report.ErrorCount != 0 ||
		report.WarningCount != 0 {
		t.Fatalf(
			"unexpected validation report: %#v",
			report,
		)
	}
}

func TestValidateAcceptsHourlyWeeklyAndCustomBuckets(
	t *testing.T,
) {
	tests := []struct {
		name   string
		result Result
	}{
		{
			name: "hour",
			result: validSingleBucketResult(
				GranularityHour,
				time.Date(
					2026,
					time.July,
					1,
					12,
					0,
					0,
					0,
					time.UTC,
				),
				time.Hour,
			),
		},
		{
			name: "week",
			result: validSingleBucketResult(
				GranularityWeek,
				time.Date(
					2026,
					time.June,
					29,
					0,
					0,
					0,
					0,
					time.UTC,
				),
				7*24*time.Hour,
			),
		},
		{
			name: "custom",
			result: validSingleBucketResult(
				GranularityCustom,
				time.Date(
					2026,
					time.July,
					1,
					12,
					15,
					0,
					0,
					time.UTC,
				),
				90*time.Minute,
			),
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				report := Validate(
					test.result,
				)
				if report.Status !=
					ValidationStatusValid {
					t.Fatalf(
						"unexpected report: %#v",
						report,
					)
				}
			},
		)
	}
}

func TestValidateRejectsContractViolations(
	t *testing.T,
) {
	tests := []struct {
		name     string
		mutate   func(*Result)
		wantCode string
	}{
		{
			name: "schema version",
			mutate: func(result *Result) {
				result.SchemaVersion = "future"
			},
			wantCode: "unsupported_schema_version",
		},
		{
			name: "metric",
			mutate: func(result *Result) {
				result.Metric.Name = "unknown"
			},
			wantCode: "metric_name_unsupported",
		},
		{
			name: "scope",
			mutate: func(result *Result) {
				result.Scope.RegionCode =
					"azerbaijan"
			},
			wantCode: "scope_field_not_applicable",
		},
		{
			name: "window future",
			mutate: func(result *Result) {
				result.Window.AsOfTime =
					result.Window.EndTime.Add(
						-time.Minute,
					)
			},
			wantCode: "window_exceeds_as_of_time",
		},
		{
			name: "bucket alignment",
			mutate: func(result *Result) {
				result.Points[0].StartTime =
					result.Points[0].
						StartTime.Add(
						time.Minute,
					)
			},
			wantCode: "day_bucket_misaligned",
		},
		{
			name: "bucket overlap",
			mutate: func(result *Result) {
				result.Points[1].StartTime =
					result.Points[0].
						EndTime.Add(
						-time.Hour,
					)
			},
			wantCode: "bucket_overlap",
		},
		{
			name: "complete gap",
			mutate: func(result *Result) {
				result.Points =
					result.Points[1:]
				result.Summary = Summarize(
					result.Points,
				)
				result.Confidence.SampleCount =
					totalSampleCount(
						result.Points,
					)
			},
			wantCode: "complete_series_coverage_invalid",
		},
		{
			name: "summary",
			mutate: func(result *Result) {
				result.Summary.Total = 999
			},
			wantCode: "summary_value_mismatch",
		},
		{
			name: "comparison arithmetic",
			mutate: func(result *Result) {
				result.Comparison.
					AbsoluteChange = 999
			},
			wantCode: "comparison_absolute_change_mismatch",
		},
		{
			name: "source future",
			mutate: func(result *Result) {
				result.Provenance.
					LatestSourceUpdatedAt =
					result.Window.AsOfTime.Add(
						time.Second,
					)
			},
			wantCode: "source_updated_after_as_of_time",
		},
		{
			name: "source order",
			mutate: func(result *Result) {
				result.Provenance.SourceNames =
					[]string{
						"route_results",
						"flight_trajectories",
					}
			},
			wantCode: "source_names_not_sorted",
		},
		{
			name: "confidence sample count",
			mutate: func(result *Result) {
				result.Points[0].
					Confidence.SampleCount++
			},
			wantCode: "confidence_sample_count_mismatch",
		},
		{
			name: "future bucket",
			mutate: func(result *Result) {
				result.Window.AsOfTime =
					result.Window.EndTime
				result.Points[2].EndTime =
					result.Window.EndTime.Add(
						time.Hour,
					)
			},
			wantCode: "bucket_future_evidence",
		},
		{
			name: "invalid value",
			mutate: func(result *Result) {
				result.Points[0].Value =
					math.NaN()
			},
			wantCode: "bucket_value_invalid",
		},
		{
			name: "duplicate limitation",
			mutate: func(result *Result) {
				result.Limitations = append(
					result.Limitations,
					result.Limitations[0],
				)
			},
			wantCode: "limitation_duplicate",
		},
		{
			name: "generated before as of",
			mutate: func(result *Result) {
				result.GeneratedAt =
					result.Window.AsOfTime.Add(
						-time.Second,
					)
			},
			wantCode: "generated_before_as_of_time",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				result :=
					validCompleteResult()
				test.mutate(&result)

				report := Validate(result)
				if report.Status !=
					ValidationStatusInvalid {
					t.Fatalf(
						"report status = %q, issues=%#v",
						report.Status,
						report.Issues,
					)
				}
				if !reportHasCode(
					report,
					test.wantCode,
				) {
					t.Fatalf(
						"report does not contain %q: %#v",
						test.wantCode,
						report.Issues,
					)
				}
			},
		)
	}
}

func TestValidateRejectsInvalidScopeVariants(
	t *testing.T,
) {
	tests := []struct {
		name  string
		scope Scope
		code  string
	}{
		{
			name: "region uppercase",
			scope: Scope{
				Type:       ScopeTypeRegion,
				RegionCode: "AZERBAIJAN",
			},
			code: "region_code_invalid",
		},
		{
			name: "airport missing",
			scope: Scope{
				Type: ScopeTypeAirport,
			},
			code: "airport_icao_invalid",
		},
		{
			name: "route destination missing",
			scope: Scope{
				Type:           ScopeTypeRoute,
				OriginICAOCode: "UBBB",
			},
			code: "airport_icao_invalid",
		},
		{
			name: "unknown",
			scope: Scope{
				Type: "unknown",
			},
			code: "scope_type_invalid",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				result :=
					validCompleteResult()
				result.Scope = test.scope

				report := Validate(result)
				if !reportHasCode(
					report,
					test.code,
				) {
					t.Fatalf(
						"unexpected report: %#v",
						report,
					)
				}
			},
		)
	}
}

func TestValidateRejectsUnavailableBucketWithData(
	t *testing.T,
) {
	result := validPartialResult()
	result.Points[0].Status =
		BucketStatusUnavailable
	result.Points[0].Value = 10
	result.Points[0].SampleCount = 5
	result.Points[0].CoverageRatio = 0.5
	result.Points[0].Confidence =
		validConfidence(0.5, 5)
	result.Summary = Summary{}

	report := Validate(result)
	if !reportHasCode(
		report,
		"unavailable_bucket_has_data",
	) {
		t.Fatalf(
			"unexpected report: %#v",
			report,
		)
	}
}

func TestValidateRequiresUndefinedPercentageForZeroBaseline(
	t *testing.T,
) {
	result := validCompleteResult()
	percentage := 100.0
	result.Comparison.PreviousValue = 0
	result.Comparison.CurrentValue = 60
	result.Comparison.AbsoluteChange = 60
	result.Comparison.PercentageChange =
		&percentage
	result.Comparison.Direction =
		TrendDirectionUp

	report := Validate(result)
	if !reportHasCode(
		report,
		"comparison_percentage_change_undefined",
	) {
		t.Fatalf(
			"unexpected report: %#v",
			report,
		)
	}
}

func TestValidationIssuesAreDeterministicallySorted(
	t *testing.T,
) {
	result := validCompleteResult()
	result.SchemaVersion = "future"
	result.Metric.Unit = " items "
	result.Provenance.InputFingerprint =
		"bad"

	first := Validate(result)
	second := Validate(result)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf(
			"reports differ: %#v %#v",
			first,
			second,
		)
	}

	for index := 1; index < len(first.Issues); index++ {
		previous := first.Issues[index-1]
		current := first.Issues[index]
		if current.Field < previous.Field {
			t.Fatalf(
				"issues are not sorted: %#v",
				first.Issues,
			)
		}
	}
}

func TestValidationReportCloneDoesNotShareIssues(
	t *testing.T,
) {
	result := validCompleteResult()
	result.SchemaVersion = "future"

	report := Validate(result)
	cloned := report.Clone()
	cloned.Issues[0].Code = "changed"

	if report.Issues[0].Code == "changed" {
		t.Fatal(
			"ValidationReport.Clone() shared issues",
		)
	}
}

func validCompleteResult() Result {
	startTime := time.Date(
		2026,
		time.July,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(
		3 * 24 * time.Hour,
	)
	asOfTime := endTime.Add(time.Hour)
	percentageChange := 50.0

	points := []Point{
		validPoint(
			startTime,
			10,
			10,
		),
		validPoint(
			startTime.Add(24*time.Hour),
			20,
			20,
		),
		validPoint(
			startTime.Add(48*time.Hour),
			30,
			30,
		),
	}

	return Result{
		SchemaVersion: SchemaVersionV1,
		Status:        SeriesStatusComplete,
		Metric: Metric{
			Name:        MetricNameFlightCount,
			Unit:        "flights",
			Aggregation: AggregationCount,
		},
		Scope: Scope{
			Type: ScopeTypeGlobal,
		},
		Window: TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
		Granularity: GranularityDay,
		Points:      points,
		Summary:     Summarize(points),
		Comparison: &PeriodComparison{
			PreviousWindow: TimeWindow{
				StartTime: startTime.Add(
					-3 * 24 * time.Hour,
				),
				EndTime:  startTime,
				AsOfTime: asOfTime,
			},
			CurrentValue:     60,
			PreviousValue:    40,
			AbsoluteChange:   20,
			PercentageChange: &percentageChange,
			Direction:        TrendDirectionUp,
		},
		Confidence: Confidence{
			Score:       0.95,
			Level:       ConfidenceLevelHigh,
			SampleCount: 60,
			Reasons: []ConfidenceReason{
				{
					Code:         "complete_source_coverage",
					Message:      "Historical source coverage is complete.",
					Contribution: 0.95,
				},
			},
		},
		Limitations: []Limitation{
			{
				Code:    "historical_observation_only",
				Message: "Historical result represents observed open-data coverage.",
				Scope:   "series",
			},
		},
		Provenance: Provenance{
			BuilderVersion: "historical-builder-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("a", 64),
			SourceNames: []string{
				"flight_trajectories",
				"route_results",
			},
			LatestSourceUpdatedAt: endTime,
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}
}

func validUnavailableResult() Result {
	startTime := time.Date(
		2026,
		time.July,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(
		24 * time.Hour,
	)
	asOfTime := endTime.Add(time.Hour)

	return Result{
		SchemaVersion: SchemaVersionV1,
		Status:        SeriesStatusUnavailable,
		Metric: Metric{
			Name:        MetricNameActiveAircraft,
			Unit:        "aircraft",
			Aggregation: AggregationMaximum,
		},
		Scope: Scope{
			Type:       ScopeTypeRegion,
			RegionCode: "azerbaijan",
		},
		Window: TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
		Granularity: GranularityHour,
		Summary:     Summary{},
		Confidence: Confidence{
			Score:       0,
			Level:       ConfidenceLevelNone,
			SampleCount: 0,
		},
		Limitations: []Limitation{
			{
				Code:    "historical_data_unavailable",
				Message: "No persisted observations cover the requested range.",
				Scope:   "series",
			},
		},
		Provenance: Provenance{
			BuilderVersion: "historical-builder-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("b", 64),
			SourceNames: []string{
				"flight_states",
			},
			LatestSourceUpdatedAt: startTime,
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}
}

func validPartialResult() Result {
	startTime := time.Date(
		2026,
		time.July,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(
		2 * 24 * time.Hour,
	)
	asOfTime := endTime.Add(time.Hour)

	points := []Point{
		{
			StartTime:     startTime,
			EndTime:       startTime.Add(24 * time.Hour),
			Status:        BucketStatusPartial,
			Value:         5,
			SampleCount:   5,
			CoverageRatio: 0.5,
			Confidence:    validConfidence(0.5, 5),
			Limitations: []Limitation{
				{
					Code:    "partial_source_coverage",
					Message: "Only half of the expected source interval is available.",
					Scope:   "bucket",
				},
			},
		},
	}

	return Result{
		SchemaVersion: SchemaVersionV1,
		Status:        SeriesStatusPartial,
		Metric: Metric{
			Name:        MetricNameTrajectoryCount,
			Unit:        "trajectories",
			Aggregation: AggregationCount,
		},
		Scope: Scope{
			Type:            ScopeTypeAirport,
			AirportICAOCode: "UBBB",
		},
		Window: TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
		Granularity: GranularityDay,
		Points:      points,
		Summary:     Summarize(points),
		Confidence:  validConfidence(0.5, 5),
		Limitations: []Limitation{
			{
				Code:    "partial_historical_window",
				Message: "The requested historical range is only partially covered.",
				Scope:   "series",
			},
		},
		Provenance: Provenance{
			BuilderVersion: "historical-builder-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("c", 64),
			SourceNames: []string{
				"flight_trajectories",
			},
			LatestSourceUpdatedAt: startTime.Add(12 * time.Hour),
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}
}

func validSingleBucketResult(
	granularity Granularity,
	startTime time.Time,
	duration time.Duration,
) Result {
	endTime := startTime.Add(duration)
	asOfTime := endTime.Add(time.Hour)
	point := Point{
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        BucketStatusComplete,
		Value:         1,
		SampleCount:   1,
		CoverageRatio: 1,
		Confidence:    validConfidence(1, 1),
	}

	return Result{
		SchemaVersion: SchemaVersionV1,
		Status:        SeriesStatusComplete,
		Metric: Metric{
			Name:        MetricNameObservationCount,
			Unit:        "observations",
			Aggregation: AggregationCount,
		},
		Scope: Scope{
			Type:                ScopeTypeRoute,
			OriginICAOCode:      "UBBB",
			DestinationICAOCode: "UGTB",
		},
		Window: TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
		Granularity: granularity,
		Points: []Point{
			point,
		},
		Summary: Summarize(
			[]Point{point},
		),
		Confidence: validConfidence(1, 1),
		Provenance: Provenance{
			BuilderVersion: "historical-builder-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("d", 64),
			SourceNames: []string{
				"flight_states",
			},
			LatestSourceUpdatedAt: endTime,
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}
}

func validPoint(
	startTime time.Time,
	value float64,
	sampleCount int,
) Point {
	return Point{
		StartTime:     startTime,
		EndTime:       startTime.Add(24 * time.Hour),
		Status:        BucketStatusComplete,
		Value:         value,
		SampleCount:   sampleCount,
		CoverageRatio: 1,
		Confidence:    validConfidence(0.95, sampleCount),
	}
}

func validConfidence(
	score float64,
	sampleCount int,
) Confidence {
	return Confidence{
		Score:       score,
		Level:       ConfidenceLevelForScore(score),
		SampleCount: sampleCount,
		Reasons: []ConfidenceReason{
			{
				Code:         "observed_samples",
				Message:      "Confidence is derived from persisted source samples.",
				Contribution: score,
			},
		},
	}
}

func reportHasCode(
	report ValidationReport,
	code string,
) bool {
	for _, issue := range report.Issues {
		if issue.Code == code {
			return true
		}
	}

	return false
}
