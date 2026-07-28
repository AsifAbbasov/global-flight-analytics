package historicalcomparison

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalseries"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func TestAttachBuildsAtomicComparisonProvenanceAndQuality(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		comparisonFixture{
			startTime:       asOfTime.Add(-4 * time.Hour),
			endTime:         asOfTime.Add(-2 * time.Hour),
			asOfTime:        asOfTime,
			values:          []float64{1, 2},
			coverages:       []float64{1, 1},
			builderVersion:  "previous-builder",
			fingerprintRune: "b",
			sourceNames: []string{
				"shared-source",
				"previous-source",
			},
			latestSourceUpdatedAt: asOfTime.
				Add(-30 * time.Minute),
		},
	)
	current := comparisonSeries(
		t,
		comparisonFixture{
			startTime:       asOfTime.Add(-2 * time.Hour),
			endTime:         asOfTime,
			asOfTime:        asOfTime,
			values:          []float64{3, 3},
			coverages:       []float64{1, 1},
			builderVersion:  "current-builder",
			fingerprintRune: "a",
			sourceNames: []string{
				"current-source",
				"shared-source",
			},
			latestSourceUpdatedAt: asOfTime.
				Add(-10 * time.Minute),
		},
	)

	result, err := Attach(current, previous)
	if err != nil {
		t.Fatalf("attach comparison: %v", err)
	}
	if result.Comparison == nil {
		t.Fatal("expected period comparison")
	}
	if result.Comparison.CurrentValue != 6 ||
		result.Comparison.PreviousValue != 3 ||
		result.Comparison.AbsoluteChange != 3 {
		t.Fatalf(
			"unexpected comparison values: %#v",
			result.Comparison,
		)
	}
	if result.Comparison.PercentageChange == nil ||
		*result.Comparison.PercentageChange != 100 {
		t.Fatalf(
			"expected 100 percent increase, got %#v",
			result.Comparison.PercentageChange,
		)
	}
	if result.Comparison.Direction !=
		historicalcontract.TrendDirectionUp {
		t.Fatalf(
			"expected upward trend, got %s",
			result.Comparison.Direction,
		)
	}

	if !strings.Contains(
		result.Provenance.BuilderVersion,
		Version,
	) ||
		!strings.Contains(
			result.Provenance.BuilderVersion,
			"current-builder",
		) ||
		!strings.Contains(
			result.Provenance.BuilderVersion,
			"previous-builder",
		) {
		t.Fatalf(
			"comparison builder provenance is incomplete: %q",
			result.Provenance.BuilderVersion,
		)
	}
	if result.Provenance.InputFingerprint ==
		current.Provenance.InputFingerprint ||
		result.Provenance.InputFingerprint ==
			previous.Provenance.InputFingerprint {
		t.Fatal(
			"comparison fingerprint must bind both periods",
		)
	}
	expectedSources := []string{
		"current-source",
		"previous-source",
		"shared-source",
	}
	if strings.Join(
		result.Provenance.SourceNames,
		",",
	) != strings.Join(expectedSources, ",") {
		t.Fatalf(
			"unexpected merged sources: %#v",
			result.Provenance.SourceNames,
		)
	}
	if !result.Provenance.LatestSourceUpdatedAt.Equal(
		asOfTime.Add(-10 * time.Minute),
	) {
		t.Fatalf(
			"unexpected latest source time: %s",
			result.Provenance.LatestSourceUpdatedAt,
		)
	}
	if result.Confidence.Score != 1 ||
		result.Confidence.SampleCount !=
			previous.Confidence.SampleCount {
		t.Fatalf(
			"unexpected comparison confidence: %#v",
			result.Confidence,
		)
	}
}

func TestAttachRejectsUnequalCoverageProfiles(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-4 * time.Hour),
			endTime:   asOfTime.Add(-2 * time.Hour),
			asOfTime:  asOfTime,
			values:    []float64{1, 2},
			coverages: []float64{0.2, 0.2},
		},
	)
	current := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-2 * time.Hour),
			endTime:   asOfTime,
			asOfTime:  asOfTime,
			values:    []float64{3, 3},
			coverages: []float64{1, 1},
		},
	)

	_, err := Attach(current, previous)
	if !errors.Is(err, ErrCoverageMismatch) {
		t.Fatalf(
			"expected coverage mismatch, got %v",
			err,
		)
	}
}

func TestAttachCarriesMatchedPartialCoverageQuality(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-4 * time.Hour),
			endTime:   asOfTime.Add(-2 * time.Hour),
			asOfTime:  asOfTime,
			values:    []float64{1, 2},
			coverages: []float64{0.5, 0.5},
			limitations: []historicalcontract.Limitation{
				{
					Code:    "previous_dataset_constraint",
					Message: "Previous evidence is constrained.",
					Scope:   "dataset",
				},
			},
		},
	)
	current := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-2 * time.Hour),
			endTime:   asOfTime,
			asOfTime:  asOfTime,
			values:    []float64{3, 3},
			coverages: []float64{0.5, 0.5},
		},
	)

	result, err := Attach(current, previous)
	if err != nil {
		t.Fatalf(
			"attach matched partial comparison: %v",
			err,
		)
	}
	if result.Status !=
		historicalcontract.SeriesStatusPartial {
		t.Fatalf(
			"comparison status=%s want partial",
			result.Status,
		)
	}
	if result.Confidence.Score != 0.5 ||
		result.Confidence.Level !=
			historicalcontract.ConfidenceLevelForScore(
				0.5,
			) {
		t.Fatalf(
			"unexpected partial comparison confidence: %#v",
			result.Confidence,
		)
	}
	assertLimitationCode(
		t,
		result.Limitations,
		"historical_comparison_period_quality",
	)
	assertLimitationCode(
		t,
		result.Limitations,
		"historical_comparison_previous_period_limitations",
	)
	assertLimitationCode(
		t,
		result.Limitations,
		"historical_comparison_matched_partial_coverage",
	)
}

func TestValidateCompatibilityClassifiesEveryMismatch(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-4 * time.Hour),
			endTime:   asOfTime.Add(-2 * time.Hour),
			asOfTime:  asOfTime,
			values:    []float64{1, 2},
			coverages: []float64{1, 1},
		},
	)
	current := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-2 * time.Hour),
			endTime:   asOfTime,
			asOfTime:  asOfTime,
			values:    []float64{3, 3},
			coverages: []float64{1, 1},
		},
	)

	tests := []struct {
		name string
		kind error
		edit func(
			*historicalcontract.Result,
			*historicalcontract.Result,
		)
	}{
		{
			name: "schema",
			kind: ErrSchemaMismatch,
			edit: func(
				current *historicalcontract.Result,
				_ *historicalcontract.Result,
			) {
				current.SchemaVersion = "other-schema"
			},
		},
		{
			name: "metric",
			kind: ErrMetricMismatch,
			edit: func(
				current *historicalcontract.Result,
				_ *historicalcontract.Result,
			) {
				current.Metric.Unit = "other-unit"
			},
		},
		{
			name: "scope",
			kind: ErrScopeMismatch,
			edit: func(
				current *historicalcontract.Result,
				_ *historicalcontract.Result,
			) {
				current.Scope = historicalcontract.Scope{
					Type:       historicalcontract.ScopeTypeRegion,
					RegionCode: "az",
				}
			},
		},
		{
			name: "granularity",
			kind: ErrGranularityMismatch,
			edit: func(
				current *historicalcontract.Result,
				_ *historicalcontract.Result,
			) {
				current.Granularity =
					historicalcontract.GranularityDay
			},
		},
		{
			name: "as_of",
			kind: ErrAsOfTimeMismatch,
			edit: func(
				_ *historicalcontract.Result,
				previous *historicalcontract.Result,
			) {
				previous.Window.AsOfTime =
					previous.Window.AsOfTime.
						Add(time.Second)
			},
		},
		{
			name: "duration",
			kind: ErrWindowDurationMismatch,
			edit: func(
				current *historicalcontract.Result,
				_ *historicalcontract.Result,
			) {
				current.Window.StartTime =
					current.Window.StartTime.
						Add(time.Minute)
			},
		},
		{
			name: "adjacency",
			kind: ErrWindowNotAdjacent,
			edit: func(
				_ *historicalcontract.Result,
				previous *historicalcontract.Result,
			) {
				previous.Window.EndTime =
					previous.Window.EndTime.
						Add(-time.Minute)
				previous.Window.StartTime =
					previous.Window.StartTime.
						Add(-time.Minute)
			},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				left := current.Clone()
				right := previous.Clone()
				test.edit(&left, &right)

				err := validateCompatibility(
					left,
					right,
				)
				if !errors.Is(err, test.kind) {
					t.Fatalf(
						"expected %v, got %v",
						test.kind,
						err,
					)
				}
			},
		)
	}
}

func TestSelectPeriodValuesSupportsEveryAggregation(
	t *testing.T,
) {
	summary := historicalcontract.Summary{
		Total:   10,
		Minimum: 2,
		Maximum: 8,
		Average: 5,
		Median:  4,
	}
	expected := map[historicalcontract.Aggregation]float64{
		historicalcontract.AggregationCount:   10,
		historicalcontract.AggregationSum:     10,
		historicalcontract.AggregationMinimum: 2,
		historicalcontract.AggregationMaximum: 8,
		historicalcontract.AggregationAverage: 5,
		historicalcontract.AggregationMedian:  4,
		historicalcontract.AggregationRatio:   5,
	}

	for aggregation, want := range expected {
		t.Run(
			string(aggregation),
			func(t *testing.T) {
				current := historicalcontract.Result{
					Metric: historicalcontract.Metric{
						Aggregation: aggregation,
					},
					Summary: summary,
				}
				previous := current.Clone()
				previous.Summary.Total = 9

				values, err := selectPeriodValues(
					current,
					previous,
				)
				if err != nil {
					t.Fatalf(
						"select values: %v",
						err,
					)
				}
				if values.current != want {
					t.Fatalf(
						"current value=%f want=%f",
						values.current,
						want,
					)
				}
			},
		)
	}
}

func TestAttachClassifiesDecreaseAndFlatDirection(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	tests := []struct {
		name      string
		previous  []float64
		current   []float64
		direction historicalcontract.TrendDirection
	}{
		{
			name:      "decrease",
			previous:  []float64{3, 3},
			current:   []float64{1, 2},
			direction: historicalcontract.TrendDirectionDown,
		},
		{
			name:      "flat",
			previous:  []float64{1, 2},
			current:   []float64{1, 2},
			direction: historicalcontract.TrendDirectionFlat,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				previous := comparisonSeries(
					t,
					comparisonFixture{
						startTime: asOfTime.Add(
							-4 * time.Hour,
						),
						endTime: asOfTime.Add(
							-2 * time.Hour,
						),
						asOfTime:  asOfTime,
						values:    test.previous,
						coverages: []float64{1, 1},
					},
				)
				current := comparisonSeries(
					t,
					comparisonFixture{
						startTime: asOfTime.Add(
							-2 * time.Hour,
						),
						endTime:   asOfTime,
						asOfTime:  asOfTime,
						values:    test.current,
						coverages: []float64{1, 1},
					},
				)

				result, err := Attach(
					current,
					previous,
				)
				if err != nil {
					t.Fatalf(
						"attach: %v",
						err,
					)
				}
				if result.Comparison.Direction !=
					test.direction {
					t.Fatalf(
						"direction=%s want=%s",
						result.Comparison.Direction,
						test.direction,
					)
				}
			},
		)
	}
}

func TestAttachKeepsZeroBasePercentageUndefined(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-4 * time.Hour),
			endTime:   asOfTime.Add(-2 * time.Hour),
			asOfTime:  asOfTime,
			values:    []float64{0, 0},
			coverages: []float64{1, 1},
		},
	)
	current := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-2 * time.Hour),
			endTime:   asOfTime,
			asOfTime:  asOfTime,
			values:    []float64{1, 0},
			coverages: []float64{1, 1},
		},
	)

	result, err := Attach(current, previous)
	if err != nil {
		t.Fatalf(
			"attach zero-base comparison: %v",
			err,
		)
	}
	if result.Comparison.PercentageChange != nil {
		t.Fatalf(
			"expected undefined percentage, got %f",
			*result.Comparison.PercentageChange,
		)
	}
}

func TestAttachRejectsPercentageOverflowWithArithmeticError(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	metric := historicalcontract.Metric{
		Name:        historicalcontract.MetricNameRouteConfidence,
		Unit:        "ratio",
		Aggregation: historicalcontract.AggregationAverage,
	}
	previous := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-2 * time.Hour),
			endTime:   asOfTime.Add(-time.Hour),
			asOfTime:  asOfTime,
			metric:    metric,
			values: []float64{
				math.SmallestNonzeroFloat64,
			},
			coverages: []float64{1},
		},
	)
	current := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-time.Hour),
			endTime:   asOfTime,
			asOfTime:  asOfTime,
			metric:    metric,
			values:    []float64{1},
			coverages: []float64{1},
		},
	)

	_, err := Attach(current, previous)
	if !errors.Is(
		err,
		ErrComparisonArithmeticInvalid,
	) {
		t.Fatalf(
			"expected arithmetic error, got %v",
			err,
		)
	}
	if errors.Is(err, ErrCurrentResultInvalid) {
		t.Fatal(
			"comparison arithmetic must not be classified as invalid current input",
		)
	}
}

func TestComparisonFingerprintChangesWithPreviousEvidence(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		comparisonFixture{
			startTime:       asOfTime.Add(-4 * time.Hour),
			endTime:         asOfTime.Add(-2 * time.Hour),
			asOfTime:        asOfTime,
			values:          []float64{1, 2},
			coverages:       []float64{1, 1},
			fingerprintRune: "b",
		},
	)
	current := comparisonSeries(
		t,
		comparisonFixture{
			startTime:       asOfTime.Add(-2 * time.Hour),
			endTime:         asOfTime,
			asOfTime:        asOfTime,
			values:          []float64{3, 3},
			coverages:       []float64{1, 1},
			fingerprintRune: "a",
		},
	)

	first, err := Attach(current, previous)
	if err != nil {
		t.Fatalf("first attach: %v", err)
	}
	changedPrevious := previous.Clone()
	changedPrevious.Provenance.InputFingerprint =
		"sha256:" + strings.Repeat("c", 64)

	second, err := Attach(current, changedPrevious)
	if err != nil {
		t.Fatalf("second attach: %v", err)
	}
	if first.Provenance.InputFingerprint ==
		second.Provenance.InputFingerprint {
		t.Fatal(
			"comparison fingerprint must change with previous evidence",
		)
	}
}

func TestAttachRejectsNestedComparison(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-4 * time.Hour),
			endTime:   asOfTime.Add(-2 * time.Hour),
			asOfTime:  asOfTime,
			values:    []float64{1, 2},
			coverages: []float64{1, 1},
		},
	)
	current := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-2 * time.Hour),
			endTime:   asOfTime,
			asOfTime:  asOfTime,
			values:    []float64{3, 3},
			coverages: []float64{1, 1},
		},
	)
	percentage := 100.0
	current.Comparison = &historicalcontract.PeriodComparison{
		PreviousWindow:   previous.Window,
		CurrentValue:     6,
		PreviousValue:    3,
		AbsoluteChange:   3,
		PercentageChange: &percentage,
		Direction:        historicalcontract.TrendDirectionUp,
	}

	_, err := Attach(current, previous)
	if !errors.Is(
		err,
		ErrNestedComparisonUnsupported,
	) {
		t.Fatalf(
			"expected nested comparison error, got %v",
			err,
		)
	}
}

func TestAttachDoesNotShareComparisonState(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-4 * time.Hour),
			endTime:   asOfTime.Add(-2 * time.Hour),
			asOfTime:  asOfTime,
			values:    []float64{1, 2},
			coverages: []float64{1, 1},
		},
	)
	current := comparisonSeries(
		t,
		comparisonFixture{
			startTime: asOfTime.Add(-2 * time.Hour),
			endTime:   asOfTime,
			asOfTime:  asOfTime,
			values:    []float64{3, 3},
			coverages: []float64{1, 1},
		},
	)

	result, err := Attach(current, previous)
	if err != nil {
		t.Fatalf("attach comparison: %v", err)
	}

	clone := result.Clone()
	*clone.Comparison.PercentageChange = 999
	if *result.Comparison.PercentageChange == 999 {
		t.Fatal(
			"comparison percentage pointer must be cloned",
		)
	}
}

type comparisonFixture struct {
	startTime time.Time
	endTime   time.Time
	asOfTime  time.Time

	metric historicalcontract.Metric
	scope  historicalcontract.Scope

	values    []float64
	coverages []float64

	builderVersion        string
	fingerprintRune       string
	sourceNames           []string
	latestSourceUpdatedAt time.Time
	limitations           []historicalcontract.Limitation
}

func comparisonSeries(
	t *testing.T,
	fixture comparisonFixture,
) historicalcontract.Result {
	t.Helper()

	metric := fixture.metric
	if metric.Name == "" {
		metric = historicalcontract.Metric{
			Name:        historicalcontract.MetricNameFlightCount,
			Unit:        "flights",
			Aggregation: historicalcontract.AggregationCount,
		}
	}
	scope := fixture.scope
	if scope.Type == "" {
		scope = historicalcontract.Scope{
			Type: historicalcontract.ScopeTypeGlobal,
		}
	}
	builderVersion := fixture.builderVersion
	if builderVersion == "" {
		builderVersion = "comparison-test-builder"
	}
	fingerprintRune := fixture.fingerprintRune
	if fingerprintRune == "" {
		fingerprintRune = "a"
	}
	sourceNames := fixture.sourceNames
	if len(sourceNames) == 0 {
		sourceNames = []string{"comparison-test-source"}
	}
	latestSourceUpdatedAt :=
		fixture.latestSourceUpdatedAt
	if latestSourceUpdatedAt.IsZero() {
		latestSourceUpdatedAt = fixture.endTime
	}

	window := historicalcontract.TimeWindow{
		StartTime: fixture.startTime,
		EndTime:   fixture.endTime,
		AsOfTime:  fixture.asOfTime,
	}
	buckets := make(
		[]historicalwindow.Bucket,
		0,
		len(fixture.values),
	)
	bucketValues := make(
		[]historicalseries.BucketValue,
		0,
		len(fixture.values),
	)

	for index, value := range fixture.values {
		coverage := fixture.coverages[index]
		bucket := historicalwindow.Bucket{
			Key: "bucket-" +
				string(rune('a'+index)),
			Sequence: index,
			StartTime: fixture.startTime.Add(
				time.Duration(index) * time.Hour,
			),
			EndTime: fixture.startTime.Add(
				time.Duration(index+1) * time.Hour,
			),
		}
		buckets = append(buckets, bucket)

		state := historicalseries.CoverageStateComplete
		loadedCount := int64(10)
		matchedCount := int64(10)
		switch {
		case coverage == 0:
			state = historicalseries.
				CoverageStateUnavailable
			loadedCount = 0
			matchedCount = 0
		case coverage < 1:
			state = historicalseries.
				CoverageStatePartial
			loadedCount = int64(
				math.Round(coverage * 10),
			)
			if loadedCount < 1 {
				loadedCount = 1
			}
		}

		bucketValues = append(
			bucketValues,
			historicalseries.BucketValue{
				Bucket:      bucket,
				Value:       value,
				SampleCount: int(loadedCount),
				Coverage: historicalseries.CoverageEvidence{
					State:        state,
					LoadedCount:  loadedCount,
					MatchedCount: matchedCount,
					Ratio:        coverage,
				},
			},
		)
	}

	result, err := historicalseries.Build(
		historicalseries.BuildRequest{
			Metric: metric,
			Scope:  scope,
			Plan: historicalwindow.Plan{
				Version: historicalwindow.Version,
				Fingerprint: "comparison-plan-" +
					fixture.startTime.Format(
						time.RFC3339,
					),
				RequestedStartTime: fixture.startTime,
				RequestedEndTime:   fixture.endTime,
				AsOfTime:           fixture.asOfTime,
				Granularity: historicalcontract.
					GranularityHour,
				EffectiveWindow:    &window,
				Buckets:            buckets,
				MaximumBucketCount: 100,
			},
			Values:         bucketValues,
			BuilderVersion: builderVersion,
			InputFingerprint: "sha256:" +
				strings.Repeat(
					fingerprintRune,
					64,
				),
			SourceNames:           sourceNames,
			LatestSourceUpdatedAt: latestSourceUpdatedAt,
			GeneratedAt:           fixture.asOfTime,
			Limitations:           fixture.limitations,
		},
	)
	if err != nil {
		t.Fatalf(
			"build comparison fixture: %v",
			err,
		)
	}
	return result
}

func assertLimitationCode(
	t *testing.T,
	limitations []historicalcontract.Limitation,
	code string,
) {
	t.Helper()
	for _, limitation := range limitations {
		if limitation.Code == code {
			return
		}
	}
	t.Fatalf(
		"limitation %q is missing: %#v",
		code,
		limitations,
	)
}

func comparisonTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		28,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}
