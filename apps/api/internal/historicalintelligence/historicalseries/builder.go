package historicalseries

import (
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func Build(
	request BuildRequest,
) (historicalcontract.Result, error) {
	window, available, err := resolveWindow(
		request.Plan,
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	if len(request.Values) != len(request.Plan.Buckets) {
		return historicalcontract.Result{},
			ErrBucketValueCountInvalid
	}

	for index, value := range request.Values {
		planned := request.Plan.Buckets[index]
		if value.Bucket.Sequence != planned.Sequence ||
			value.Bucket.Key != planned.Key ||
			!value.Bucket.StartTime.Equal(planned.StartTime) ||
			!value.Bucket.EndTime.Equal(planned.EndTime) {
			return historicalcontract.Result{},
				ErrBucketValueOrderInvalid
		}
	}

	if math.IsNaN(request.DataCoverageRatio) ||
		math.IsInf(request.DataCoverageRatio, 0) ||
		request.DataCoverageRatio < 0 ||
		request.DataCoverageRatio > 1 {
		return historicalcontract.Result{},
			ErrCoverageRatioInvalid
	}

	builderVersion := strings.TrimSpace(
		request.BuilderVersion,
	)
	if builderVersion == "" {
		return historicalcontract.Result{},
			ErrBuilderVersionRequired
	}
	if !fingerprintPattern.MatchString(
		request.InputFingerprint,
	) {
		return historicalcontract.Result{},
			ErrFingerprintInvalid
	}

	sourceNames := normalizeSourceNames(
		request.SourceNames,
	)
	if len(sourceNames) == 0 {
		return historicalcontract.Result{},
			ErrSourceNamesRequired
	}

	latestSourceUpdatedAt :=
		request.LatestSourceUpdatedAt
	if latestSourceUpdatedAt.IsZero() {
		latestSourceUpdatedAt = window.EndTime
	}
	latestSourceUpdatedAt =
		latestSourceUpdatedAt.UTC()
	if latestSourceUpdatedAt.After(
		window.AsOfTime,
	) {
		return historicalcontract.Result{},
			ErrLatestSourceTimeInvalid
	}

	generatedAt := request.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = window.AsOfTime
	}
	generatedAt = generatedAt.UTC()
	if generatedAt.Before(window.AsOfTime) {
		return historicalcontract.Result{},
			ErrGeneratedAtInvalid
	}

	limitations := normalizeLimitations(
		append(
			append(
				[]historicalcontract.Limitation(nil),
				request.Limitations...,
			),
			planLimitations(request.Plan)...,
		),
	)

	result := historicalcontract.Result{
		SchemaVersion: historicalcontract.SchemaVersionV1,
		Status:        historicalcontract.SeriesStatusUnavailable,
		Metric:        request.Metric,
		Scope:         request.Scope,
		Window:        window,
		Granularity:   request.Plan.Granularity,
		Points:        []historicalcontract.Point{},
		Limitations:   limitations,
		Provenance: historicalcontract.Provenance{
			BuilderVersion:        builderVersion,
			InputFingerprint:      request.InputFingerprint,
			SourceNames:           sourceNames,
			LatestSourceUpdatedAt: latestSourceUpdatedAt,
		},
		GeneratedAt: generatedAt,
	}

	if !available {
		result.Limitations = normalizeLimitations(
			append(
				result.Limitations,
				historicalcontract.Limitation{
					Code:    "historical_window_unavailable",
					Message: "No complete historical bucket is available for the requested window.",
					Scope:   "series",
				},
			),
		)
		result.Confidence = confidence(
			0,
			0,
			"historical_window_unavailable",
			"No historical bucket could be represented.",
		)
		result.Summary = historicalcontract.Summarize(
			result.Points,
		)

		return validateResult(result)
	}

	result.Points = make(
		[]historicalcontract.Point,
		0,
		len(request.Values),
	)

	for _, value := range request.Values {
		point, pointErr := buildPoint(
			value,
			request.DataCoverageRatio,
		)
		if pointErr != nil {
			return historicalcontract.Result{},
				pointErr
		}
		result.Points = append(
			result.Points,
			point,
		)
	}

	switch {
	case request.DataCoverageRatio == 1:
		result.Status =
			historicalcontract.SeriesStatusComplete
	case len(result.Points) > 0:
		result.Status =
			historicalcontract.SeriesStatusPartial
		result.Limitations = normalizeLimitations(
			append(
				result.Limitations,
				historicalcontract.Limitation{
					Code:    "historical_data_partial_coverage",
					Message: "Historical source coverage is a conservative lower bound because one or more bounded reads were incomplete.",
					Scope:   "series",
				},
			),
		)
	default:
		result.Status =
			historicalcontract.SeriesStatusUnavailable
	}

	result.Summary = historicalcontract.Summarize(
		result.Points,
	)

	totalSamples := 0
	for _, point := range result.Points {
		totalSamples += point.SampleCount
	}
	result.Confidence = confidence(
		request.DataCoverageRatio,
		totalSamples,
		"historical_data_coverage",
		"Confidence reflects the conservative represented share of the bounded historical read.",
	)

	return validateResult(result)
}

func resolveWindow(
	plan historicalwindow.Plan,
) (historicalcontract.TimeWindow, bool, error) {
	if plan.Version != historicalwindow.Version {
		return historicalcontract.TimeWindow{},
			false,
			ErrPlanVersionInvalid
	}

	if plan.EffectiveWindow != nil &&
		len(plan.Buckets) > 0 {
		window := *plan.EffectiveWindow
		window.StartTime = window.StartTime.UTC()
		window.EndTime = window.EndTime.UTC()
		window.AsOfTime = window.AsOfTime.UTC()

		if !window.StartTime.Before(window.EndTime) ||
			window.EndTime.After(window.AsOfTime) {
			return historicalcontract.TimeWindow{},
				false,
				ErrPlanWindowInvalid
		}

		return window, true, nil
	}

	startTime := plan.RequestedStartTime.UTC()
	endTime := plan.RequestedEndTime.UTC()
	asOfTime := plan.AsOfTime.UTC()
	if endTime.After(asOfTime) {
		endTime = asOfTime
	}
	if !startTime.Before(endTime) {
		return historicalcontract.TimeWindow{},
			false,
			ErrPlanWindowInvalid
	}

	return historicalcontract.TimeWindow{
		StartTime: startTime,
		EndTime:   endTime,
		AsOfTime:  asOfTime,
	}, false, nil
}

func buildPoint(
	value BucketValue,
	coverageRatio float64,
) (historicalcontract.Point, error) {
	status := historicalcontract.BucketStatusComplete
	limitations := []historicalcontract.Limitation{}

	switch {
	case coverageRatio == 0:
		if value.Value != 0 ||
			value.SampleCount != 0 {
			return historicalcontract.Point{},
				ErrUnavailableBucketHasData
		}
		status = historicalcontract.
			BucketStatusUnavailable

	case coverageRatio < 1:
		status = historicalcontract.
			BucketStatusPartial
		limitations = []historicalcontract.Limitation{
			{
				Code:    "historical_bucket_partial_coverage",
				Message: "Bucket value is based on a bounded source read with conservative partial coverage.",
				Scope:   "bucket",
			},
		}
	}

	return historicalcontract.Point{
		StartTime:     value.Bucket.StartTime.UTC(),
		EndTime:       value.Bucket.EndTime.UTC(),
		Status:        status,
		Value:         value.Value,
		SampleCount:   value.SampleCount,
		CoverageRatio: coverageRatio,
		Confidence: confidence(
			coverageRatio,
			value.SampleCount,
			"historical_bucket_coverage",
			"Bucket confidence reflects represented source coverage.",
		),
		Limitations: limitations,
	}, nil
}

func confidence(
	score float64,
	sampleCount int,
	code string,
	message string,
) historicalcontract.Confidence {
	return historicalcontract.Confidence{
		Score: score,
		Level: historicalcontract.
			ConfidenceLevelForScore(score),
		SampleCount: sampleCount,
		Reasons: []historicalcontract.ConfidenceReason{
			{
				Code:         code,
				Message:      message,
				Contribution: score,
			},
		},
	}
}

func planLimitations(
	plan historicalwindow.Plan,
) []historicalcontract.Limitation {
	result := make(
		[]historicalcontract.Limitation,
		0,
		len(plan.Exclusions)+1,
	)

	for _, exclusion := range plan.Exclusions {
		result = append(
			result,
			historicalcontract.Limitation{
				Code: "historical_window_" +
					string(exclusion.Reason),
				Message: "The requested historical window contains an excluded interval that is not represented by a complete analytical bucket.",
				Scope:   "window",
			},
		)
	}

	if plan.TruncatedByAsOfTime {
		result = append(
			result,
			historicalcontract.Limitation{
				Code:    "historical_window_truncated_by_as_of_time",
				Message: "The historical window was truncated at the analytical as-of time to prevent future evidence.",
				Scope:   "window",
			},
		)
	}

	return result
}

func normalizeSourceNames(
	values []string,
) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))

	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	sort.Strings(result)
	return result
}

func normalizeLimitations(
	values []historicalcontract.Limitation,
) []historicalcontract.Limitation {
	seen := make(map[string]struct{})
	result := make(
		[]historicalcontract.Limitation,
		0,
		len(values),
	)

	for _, value := range values {
		value.Code = strings.TrimSpace(value.Code)
		value.Message = strings.TrimSpace(
			value.Message,
		)
		value.Scope = strings.TrimSpace(value.Scope)
		if value.Code == "" ||
			value.Message == "" ||
			value.Scope == "" {
			continue
		}

		key := value.Scope + "\x00" + value.Code
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}

	sort.SliceStable(
		result,
		func(left int, right int) bool {
			if result[left].Scope != result[right].Scope {
				return result[left].Scope <
					result[right].Scope
			}
			return result[left].Code <
				result[right].Code
		},
	)

	return result
}

func validateResult(
	result historicalcontract.Result,
) (historicalcontract.Result, error) {
	report := historicalcontract.Validate(result)
	if report.Status !=
		historicalcontract.ValidationStatusValid {
		return historicalcontract.Result{},
			&ContractValidationError{
				Report: report.Clone(),
			}
	}

	return result.Clone(), nil
}
