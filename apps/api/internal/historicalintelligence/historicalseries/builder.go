package historicalseries

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

type validatedBuildRequest struct {
	window                historicalcontract.TimeWindow
	windowAvailable       bool
	builderVersion        string
	sourceNames           []string
	latestSourceUpdatedAt time.Time
	generatedAt           time.Time
	limitations           []historicalcontract.Limitation
}

func Build(
	request BuildRequest,
) (historicalcontract.Result, error) {
	canonicalPlan, err := historicalwindow.CanonicalizePlan(
		request.Plan,
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}
	request.Plan = canonicalPlan

	validated, err := validateBuildRequest(&request)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	result := newResult(request, validated)
	if !validated.windowAvailable {
		return buildUnavailableWindowResult(result)
	}

	points, err := buildSeriesPoints(request.Values)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	coverageScore := temporalCoverageScore(points)
	status, representedPointCount := deriveSeriesStatus(points)
	result.Status = status
	result.Points = points

	switch status {
	case historicalcontract.SeriesStatusComplete:

	case historicalcontract.SeriesStatusPartial:
		result.Limitations, err = appendGeneratedLimitation(
			result.Limitations,
			historicalcontract.Limitation{
				Code:    "historical_data_partial_coverage",
				Message: "Historical bucket evidence is incomplete; each represented bucket carries its own conservative coverage lower bound.",
				Scope:   "series",
			},
		)
		if err != nil {
			return historicalcontract.Result{}, err
		}

	case historicalcontract.SeriesStatusUnavailable:
		result.Points = []historicalcontract.Point{}
		result.Limitations, err = appendGeneratedLimitation(
			result.Limitations,
			historicalcontract.Limitation{
				Code:    "historical_data_unavailable",
				Message: "No historical bucket contains represented source evidence.",
				Scope:   "series",
			},
		)
		if err != nil {
			return historicalcontract.Result{}, err
		}
	}

	result.Summary = historicalcontract.Summarize(
		result.Points,
	)

	totalSamples, err := checkedSampleCount(
		result.Points,
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}
	if representedPointCount == 0 {
		coverageScore = 0
	}
	result.Confidence = confidence(
		coverageScore,
		totalSamples,
		"historical_bucket_coverage",
		"Confidence reflects the mean temporal coverage across planned historical buckets.",
	)

	return validateResult(result)
}

func validateBuildRequest(
	request *BuildRequest,
) (validatedBuildRequest, error) {
	window, available, err := resolveWindow(request.Plan)
	if err != nil {
		return validatedBuildRequest{}, err
	}

	if err := validateBucketValues(request); err != nil {
		return validatedBuildRequest{}, err
	}

	builderVersion := strings.TrimSpace(
		request.BuilderVersion,
	)
	if builderVersion == "" {
		return validatedBuildRequest{},
			ErrBuilderVersionRequired
	}
	if !fingerprintPattern.MatchString(
		request.InputFingerprint,
	) {
		return validatedBuildRequest{},
			ErrFingerprintInvalid
	}

	sourceNames, err := normalizeSourceNames(
		request.SourceNames,
	)
	if err != nil {
		return validatedBuildRequest{}, err
	}

	latestSourceUpdatedAt :=
		request.LatestSourceUpdatedAt
	if latestSourceUpdatedAt.IsZero() {
		return validatedBuildRequest{},
			ErrLatestSourceTimeRequired
	}
	latestSourceUpdatedAt =
		latestSourceUpdatedAt.UTC()
	if latestSourceUpdatedAt.After(
		window.AsOfTime,
	) {
		return validatedBuildRequest{},
			ErrLatestSourceTimeInvalid
	}

	generatedAt := request.GeneratedAt
	if generatedAt.IsZero() {
		return validatedBuildRequest{},
			ErrGeneratedAtRequired
	}
	generatedAt = generatedAt.UTC()
	if generatedAt.Before(window.AsOfTime) {
		return validatedBuildRequest{},
			ErrGeneratedAtInvalid
	}

	limitations, err := normalizeLimitations(
		append(
			append(
				[]historicalcontract.Limitation(nil),
				request.Limitations...,
			),
			planLimitations(request.Plan)...,
		),
	)
	if err != nil {
		return validatedBuildRequest{}, err
	}

	return validatedBuildRequest{
		window:                window,
		windowAvailable:       available,
		builderVersion:        builderVersion,
		sourceNames:           sourceNames,
		latestSourceUpdatedAt: latestSourceUpdatedAt,
		generatedAt:           generatedAt,
		limitations:           limitations,
	}, nil
}

func validateBucketValues(
	request *BuildRequest,
) error {
	if len(request.Values) !=
		len(request.Plan.Buckets) {
		return ErrBucketValueCountInvalid
	}

	for index, value := range request.Values {
		planned := request.Plan.Buckets[index]
		if !value.Bucket.StartTime.Equal(
			planned.StartTime,
		) ||
			!value.Bucket.EndTime.Equal(
				planned.EndTime,
			) {
			return ErrBucketValueOrderInvalid
		}
		request.Values[index].Bucket = planned

		if err := validateCoverageEvidence(
			request.Values[index],
		); err != nil {
			return err
		}
	}

	return nil
}

func newResult(
	request BuildRequest,
	validated validatedBuildRequest,
) historicalcontract.Result {
	return historicalcontract.Result{
		SchemaVersion: historicalcontract.SchemaVersionV1,
		Status:        historicalcontract.SeriesStatusUnavailable,
		Metric:        request.Metric,
		Scope:         request.Scope,
		Window:        validated.window,
		Granularity:   request.Plan.Granularity,
		Points:        []historicalcontract.Point{},
		Limitations:   validated.limitations,
		Provenance: historicalcontract.Provenance{
			BuilderVersion:        validated.builderVersion,
			InputFingerprint:      request.InputFingerprint,
			SourceNames:           validated.sourceNames,
			LatestSourceUpdatedAt: validated.latestSourceUpdatedAt,
		},
		GeneratedAt: validated.generatedAt,
	}
}

func buildUnavailableWindowResult(
	result historicalcontract.Result,
) (historicalcontract.Result, error) {
	var err error
	result.Limitations, err = appendGeneratedLimitation(
		result.Limitations,
		historicalcontract.Limitation{
			Code:    "historical_window_unavailable",
			Message: "No complete historical bucket is available for the requested window.",
			Scope:   "series",
		},
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}
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

func buildSeriesPoints(
	values []BucketValue,
) ([]historicalcontract.Point, error) {
	points := make(
		[]historicalcontract.Point,
		0,
		len(values),
	)
	for _, value := range values {
		point, err := buildPoint(value)
		if err != nil {
			return nil, err
		}
		points = append(points, point)
	}
	return points, nil
}

func deriveSeriesStatus(
	points []historicalcontract.Point,
) (historicalcontract.SeriesStatus, int) {
	if len(points) == 0 {
		return historicalcontract.
				SeriesStatusUnavailable,
			0
	}

	represented := 0
	complete := 0
	for _, point := range points {
		switch point.Status {
		case historicalcontract.BucketStatusComplete:
			represented++
			complete++
		case historicalcontract.BucketStatusPartial:
			represented++
		}
	}

	switch {
	case represented == 0:
		return historicalcontract.
				SeriesStatusUnavailable,
			0
	case complete == len(points):
		return historicalcontract.
				SeriesStatusComplete,
			represented
	default:
		return historicalcontract.
				SeriesStatusPartial,
			represented
	}
}

func temporalCoverageScore(
	points []historicalcontract.Point,
) float64 {
	if len(points) == 0 {
		return 0
	}

	total := 0.0
	compensation := 0.0
	for _, point := range points {
		corrected :=
			point.CoverageRatio - compensation
		next := total + corrected
		compensation = (next - total) - corrected
		total = next
	}

	score := total / float64(len(points))
	if math.IsNaN(score) ||
		math.IsInf(score, 0) ||
		score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func checkedSampleCount(
	points []historicalcontract.Point,
) (int, error) {
	total := 0
	maximum := int(^uint(0) >> 1)
	for _, point := range points {
		if point.SampleCount < 0 {
			return 0, ErrBucketSampleCountInvalid
		}
		if point.SampleCount >
			maximum-total {
			return 0, ErrSampleCountOverflow
		}
		total += point.SampleCount
	}
	return total, nil
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
) (historicalcontract.Point, error) {
	if err := validateCoverageEvidence(
		value,
	); err != nil {
		return historicalcontract.Point{}, err
	}

	status := historicalcontract.BucketStatus(
		value.Coverage.State,
	)
	limitations := []historicalcontract.Limitation{}

	if value.Coverage.State ==
		CoverageStatePartial {
		limitations = []historicalcontract.Limitation{
			{
				Code:    "historical_bucket_partial_coverage",
				Message: "Bucket evidence is incomplete; coverage is a conservative lower bound derived from explicit dataset evidence.",
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
		CoverageRatio: value.Coverage.Ratio,
		Confidence: confidence(
			value.Coverage.Ratio,
			value.SampleCount,
			"historical_bucket_coverage",
			"Bucket confidence reflects explicit bucket coverage evidence.",
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

	for index, exclusion := range plan.Exclusions {
		result = append(
			result,
			historicalcontract.Limitation{
				Code: fmt.Sprintf(
					"historical_window_%s_%d",
					exclusion.Reason,
					index,
				),
				Message: fmt.Sprintf(
					"The requested historical window excludes interval [%s, %s) because %s.",
					exclusion.StartTime.UTC().
						Format(time.RFC3339Nano),
					exclusion.EndTime.UTC().
						Format(time.RFC3339Nano),
					exclusion.Reason,
				),
				Scope: "window",
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
) ([]string, error) {
	if len(values) == 0 {
		return nil, ErrSourceNamesRequired
	}

	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))

	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" ||
			normalized != value {
			return nil, ErrSourceNameInvalid
		}
		if _, exists := seen[normalized]; exists {
			return nil, ErrSourceNameDuplicate
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	sort.Strings(result)
	return result, nil
}

func normalizeLimitations(
	values []historicalcontract.Limitation,
) ([]historicalcontract.Limitation, error) {
	seen := make(map[string]struct{})
	result := make(
		[]historicalcontract.Limitation,
		0,
		len(values),
	)

	for _, value := range values {
		normalized := historicalcontract.Limitation{
			Code:    strings.TrimSpace(value.Code),
			Message: strings.TrimSpace(value.Message),
			Scope:   strings.TrimSpace(value.Scope),
		}
		if normalized.Code == "" ||
			normalized.Message == "" ||
			normalized.Scope == "" ||
			normalized.Code != value.Code ||
			normalized.Scope != value.Scope {
			return nil, ErrLimitationInvalid
		}

		key := normalized.Scope + "\x00" +
			normalized.Code
		if _, exists := seen[key]; exists {
			return nil, ErrLimitationDuplicate
		}
		seen[key] = struct{}{}
		result = append(result, normalized)
	}

	sort.SliceStable(
		result,
		func(left int, right int) bool {
			if result[left].Scope !=
				result[right].Scope {
				return result[left].Scope <
					result[right].Scope
			}
			return result[left].Code <
				result[right].Code
		},
	)

	return result, nil
}

func appendGeneratedLimitation(
	values []historicalcontract.Limitation,
	value historicalcontract.Limitation,
) ([]historicalcontract.Limitation, error) {
	return normalizeLimitations(
		append(
			append(
				[]historicalcontract.Limitation(nil),
				values...,
			),
			value,
		),
	)
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
