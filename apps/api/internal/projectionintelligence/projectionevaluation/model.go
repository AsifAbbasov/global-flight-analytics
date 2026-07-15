package projectionevaluation

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

const (
	Version                     = "projection-replay-evaluation-v1"
	FingerprintVersion          = "projection-replay-evaluation-fingerprint-v1"
	AggregateVersion            = "projection-replay-aggregate-v1"
	AggregateFingerprintVersion = "projection-replay-aggregate-fingerprint-v1"
)

type Status string

const (
	StatusUnavailable Status = "unavailable"
	StatusPartial     Status = "partial"
	StatusComplete    Status = "complete"
)

func (status Status) IsKnown() bool {
	switch status {
	case StatusUnavailable,
		StatusPartial,
		StatusComplete:
		return true
	default:
		return false
	}
}

type ActualPointSource string

const (
	ActualPointSourceObserved     ActualPointSource = "observed"
	ActualPointSourceInterpolated ActualPointSource = "interpolated"
)

func (source ActualPointSource) IsKnown() bool {
	switch source {
	case ActualPointSourceObserved,
		ActualPointSourceInterpolated:
		return true
	default:
		return false
	}
}

type Notice struct {
	Code    string
	Message string
}

type ActualArrival struct {
	AirportICAOCode string
	BoundaryTime    time.Time
	SourceName      string
	ObservedAt      time.Time
}

type PointEvaluation struct {
	Sequence     int
	ForecastTime time.Time

	ActualSource ActualPointSource
	ActualTime   time.Time

	ForecastLatitude  float64
	ForecastLongitude float64
	ActualLatitude    float64
	ActualLongitude   float64

	HorizontalErrorM            float64
	HorizontalErrorRatio        float64
	WithinHorizontalUncertainty bool

	ForecastAltitudeM         *float64
	ActualAltitudeM           *float64
	AltitudeAbsoluteErrorM    *float64
	AltitudeErrorRatio        *float64
	WithinVerticalUncertainty *bool

	ForecastConfidence projectioncontract.Confidence
}

func (item PointEvaluation) Clone() PointEvaluation {
	cloned := item
	cloned.ForecastAltitudeM =
		cloneFloat(item.ForecastAltitudeM)
	cloned.ActualAltitudeM =
		cloneFloat(item.ActualAltitudeM)
	cloned.AltitudeAbsoluteErrorM =
		cloneFloat(
			item.AltitudeAbsoluteErrorM,
		)
	cloned.AltitudeErrorRatio =
		cloneFloat(item.AltitudeErrorRatio)
	cloned.WithinVerticalUncertainty =
		cloneBool(
			item.WithinVerticalUncertainty,
		)
	cloned.ForecastConfidence =
		cloneConfidence(item.ForecastConfidence)

	return cloned
}

type PositionMetrics struct {
	ForecastPointCount      int
	EvaluatedPointCount     int
	MissingActualPointCount int
	CoverageRatio           float64

	MeanHorizontalErrorM    float64
	MedianHorizontalErrorM  float64
	P95HorizontalErrorM     float64
	MaximumHorizontalErrorM float64
	HorizontalRMSEM         float64

	MeanHorizontalErrorRatio           float64
	HorizontalUncertaintyCoverageRatio float64

	AltitudeEvaluatedPointCount   int
	MeanAltitudeAbsoluteErrorM    float64
	MedianAltitudeAbsoluteErrorM  float64
	P95AltitudeAbsoluteErrorM     float64
	MaximumAltitudeAbsoluteErrorM float64
	AltitudeRMSEM                 float64

	MeanAltitudeErrorRatio           float64
	VerticalUncertaintyCoverageRatio float64
}

type ArrivalMetrics struct {
	Available bool

	AirportICAOCode    string
	ActualBoundaryTime time.Time

	EarliestTime  time.Time
	EstimatedTime time.Time
	LatestTime    time.Time

	EstimatedAbsoluteErrorSeconds float64
	SignedErrorSeconds            float64
	IntervalWidthSeconds          float64
	IntervalCoveredActual         bool
}

func (metrics ArrivalMetrics) Clone() ArrivalMetrics {
	return metrics
}

type Result struct {
	Version string
	Status  Status

	TrajectoryID          string
	ProjectionMethod      projectioncontract.Method
	ProjectionAsOfTime    time.Time
	ProjectionGeneratedAt time.Time
	EvaluatedAt           time.Time

	ProjectionInputFingerprint string
	EvaluationInputFingerprint string

	Points   []PointEvaluation
	Position PositionMetrics
	Arrival  ArrivalMetrics

	Limitations []Notice
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Points = make(
		[]PointEvaluation,
		len(result.Points),
	)
	for index, point := range result.Points {
		cloned.Points[index] =
			point.Clone()
	}
	cloned.Arrival = result.Arrival.Clone()
	cloned.Limitations = append(
		[]Notice(nil),
		result.Limitations...,
	)

	return cloned
}

type MethodSummary struct {
	MethodName    string
	MethodVersion string
	DecisionClass projectioncontract.DecisionClass

	EvaluationCount            int
	CompleteEvaluationCount    int
	PartialEvaluationCount     int
	UnavailableEvaluationCount int

	ForecastPointCount  int
	EvaluatedPointCount int
	PointCoverageRatio  float64

	MeanHorizontalErrorM               float64
	MedianHorizontalErrorM             float64
	P95HorizontalErrorM                float64
	HorizontalRMSEM                    float64
	HorizontalUncertaintyCoverageRatio float64

	AltitudeEvaluatedPointCount      int
	MeanAltitudeAbsoluteErrorM       float64
	AltitudeRMSEM                    float64
	VerticalUncertaintyCoverageRatio float64

	ArrivalEvaluationCount          int
	MeanArrivalAbsoluteErrorSeconds float64
	ArrivalIntervalCoverageRatio    float64
}

type AggregateResult struct {
	Version string
	Status  Status

	EvaluationCount int
	MethodCount     int

	Methods     []MethodSummary
	Limitations []Notice

	InputFingerprint string
	GeneratedAt      time.Time
}

func (result AggregateResult) Clone() AggregateResult {
	cloned := result
	cloned.Methods = append(
		[]MethodSummary(nil),
		result.Methods...,
	)
	cloned.Limitations = append(
		[]Notice(nil),
		result.Limitations...,
	)

	return cloned
}

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func (result Result) Validate() error {
	if result.Version != Version {
		return fmt.Errorf(
			"projection evaluation version is invalid: %q",
			result.Version,
		)
	}
	if !result.Status.IsKnown() {
		return fmt.Errorf(
			"projection evaluation status is invalid: %q",
			result.Status,
		)
	}
	if strings.TrimSpace(result.TrajectoryID) == "" ||
		strings.TrimSpace(
			result.ProjectionMethod.Name,
		) == "" ||
		strings.TrimSpace(
			result.ProjectionMethod.Version,
		) == "" ||
		!result.ProjectionMethod.
			DecisionClass.IsKnown() {
		return fmt.Errorf(
			"projection evaluation identity is invalid",
		)
	}
	if result.ProjectionAsOfTime.IsZero() ||
		result.ProjectionGeneratedAt.IsZero() ||
		result.EvaluatedAt.IsZero() ||
		result.ProjectionGeneratedAt.Before(
			result.ProjectionAsOfTime,
		) ||
		result.EvaluatedAt.Before(
			result.ProjectionGeneratedAt,
		) {
		return fmt.Errorf(
			"projection evaluation timestamps are invalid",
		)
	}
	if !fingerprintPattern.MatchString(
		result.ProjectionInputFingerprint,
	) ||
		!fingerprintPattern.MatchString(
			result.EvaluationInputFingerprint,
		) {
		return fmt.Errorf(
			"projection evaluation fingerprints are invalid",
		)
	}

	if !validPositionMetrics(
		result.Position,
		len(result.Points),
	) {
		return fmt.Errorf(
			"projection evaluation counts or coverage are invalid",
		)
	}

	switch result.Status {
	case StatusUnavailable:
		if len(result.Limitations) == 0 {
			return fmt.Errorf(
				"unavailable evaluation requires at least one limitation",
			)
		}
	case StatusComplete:
		if result.Position.ForecastPointCount == 0 ||
			result.Position.EvaluatedPointCount !=
				result.Position.ForecastPointCount {
			return fmt.Errorf(
				"complete evaluation requires all forecast points",
			)
		}
	case StatusPartial:
		if len(result.Points) == 0 {
			return fmt.Errorf(
				"partial evaluation requires at least one evaluated point",
			)
		}
	}

	for index, point := range result.Points {
		if point.Sequence < 0 ||
			!point.ActualSource.IsKnown() ||
			point.ForecastTime.IsZero() ||
			point.ActualTime.IsZero() ||
			!point.ActualTime.Equal(
				point.ForecastTime,
			) ||
			!validLatitude(
				point.ForecastLatitude,
			) ||
			!validLongitude(
				point.ForecastLongitude,
			) ||
			!validLatitude(
				point.ActualLatitude,
			) ||
			!validLongitude(
				point.ActualLongitude,
			) ||
			!nonNegativeFinite(
				point.HorizontalErrorM,
			) ||
			!nonNegativeFinite(
				point.HorizontalErrorRatio,
			) ||
			!point.ForecastConfidence.
				Level.IsKnown() ||
			!unitInterval(
				point.ForecastConfidence.Score,
			) {
			return fmt.Errorf(
				"projection evaluation point is invalid at index %d",
				index,
			)
		}
		if index > 0 {
			previous :=
				result.Points[index-1]
			if !previous.ForecastTime.Before(
				point.ForecastTime,
			) {
				return fmt.Errorf(
					"projection evaluation points are not time ordered",
				)
			}
		}
		if point.AltitudeAbsoluteErrorM != nil &&
			(!nonNegativeFinite(
				*point.AltitudeAbsoluteErrorM,
			) ||
				point.ForecastAltitudeM == nil ||
				point.ActualAltitudeM == nil) {
			return fmt.Errorf(
				"projection altitude evaluation is invalid",
			)
		}
	}

	for _, limitation := range result.Limitations {
		if strings.TrimSpace(
			limitation.Code,
		) == "" ||
			strings.TrimSpace(
				limitation.Message,
			) == "" {
			return fmt.Errorf(
				"projection evaluation limitation is invalid",
			)
		}
	}

	if result.Arrival.Available {
		if len(result.Arrival.AirportICAOCode) != 4 ||
			result.Arrival.ActualBoundaryTime.IsZero() ||
			result.Arrival.EarliestTime.IsZero() ||
			result.Arrival.EstimatedTime.IsZero() ||
			result.Arrival.LatestTime.IsZero() ||
			result.Arrival.EarliestTime.After(
				result.Arrival.EstimatedTime,
			) ||
			result.Arrival.EstimatedTime.After(
				result.Arrival.LatestTime,
			) ||
			!nonNegativeFinite(
				result.Arrival.
					EstimatedAbsoluteErrorSeconds,
			) ||
			!finite(
				result.Arrival.
					SignedErrorSeconds,
			) ||
			!nonNegativeFinite(
				result.Arrival.
					IntervalWidthSeconds,
			) {
			return fmt.Errorf(
				"projection arrival evaluation is invalid",
			)
		}
	}

	return nil
}

func (result AggregateResult) Validate() error {
	if result.Version != AggregateVersion ||
		!result.Status.IsKnown() ||
		result.EvaluationCount < 0 ||
		result.MethodCount !=
			len(result.Methods) ||
		!fingerprintPattern.MatchString(
			result.InputFingerprint,
		) ||
		result.GeneratedAt.IsZero() {
		return fmt.Errorf(
			"projection evaluation aggregate metadata is invalid",
		)
	}

	if result.Status == StatusUnavailable &&
		(len(result.Methods) != 0 ||
			len(result.Limitations) == 0) {
		return fmt.Errorf(
			"unavailable aggregate requires no methods and limitations",
		)
	}
	if result.Status != StatusUnavailable &&
		len(result.Methods) == 0 {
		return fmt.Errorf(
			"available aggregate requires method summaries",
		)
	}

	totalEvaluations := 0
	for index, method := range result.Methods {
		if !validMethodSummary(method) {
			return fmt.Errorf(
				"projection evaluation method summary is invalid at index %d",
				index,
			)
		}
		totalEvaluations +=
			method.EvaluationCount
		if index > 0 {
			previous :=
				result.Methods[index-1]
			previousKey :=
				previous.MethodName +
					"\x00" +
					previous.MethodVersion
			currentKey :=
				method.MethodName +
					"\x00" +
					method.MethodVersion
			if previousKey >= currentKey {
				return fmt.Errorf(
					"projection evaluation method summaries are not deterministically ordered",
				)
			}
		}
	}

	if totalEvaluations !=
		result.EvaluationCount {
		return fmt.Errorf(
			"aggregate evaluation count does not match method summaries",
		)
	}

	for _, limitation := range result.Limitations {
		if strings.TrimSpace(
			limitation.Code,
		) == "" ||
			strings.TrimSpace(
				limitation.Message,
			) == "" {
			return fmt.Errorf(
				"projection evaluation aggregate limitation is invalid",
			)
		}
	}

	return nil
}

func validPositionMetrics(
	metrics PositionMetrics,
	pointCount int,
) bool {
	if metrics.ForecastPointCount < 0 ||
		metrics.EvaluatedPointCount < 0 ||
		metrics.MissingActualPointCount < 0 ||
		metrics.EvaluatedPointCount !=
			pointCount ||
		metrics.EvaluatedPointCount+
			metrics.MissingActualPointCount !=
			metrics.ForecastPointCount ||
		metrics.AltitudeEvaluatedPointCount < 0 ||
		metrics.AltitudeEvaluatedPointCount >
			metrics.EvaluatedPointCount {
		return false
	}

	ratios := []float64{
		metrics.CoverageRatio,
		metrics.
			HorizontalUncertaintyCoverageRatio,
		metrics.
			VerticalUncertaintyCoverageRatio,
	}
	for _, value := range ratios {
		if !unitInterval(value) {
			return false
		}
	}

	values := []float64{
		metrics.MeanHorizontalErrorM,
		metrics.MedianHorizontalErrorM,
		metrics.P95HorizontalErrorM,
		metrics.MaximumHorizontalErrorM,
		metrics.HorizontalRMSEM,
		metrics.MeanHorizontalErrorRatio,
		metrics.MeanAltitudeAbsoluteErrorM,
		metrics.MedianAltitudeAbsoluteErrorM,
		metrics.P95AltitudeAbsoluteErrorM,
		metrics.MaximumAltitudeAbsoluteErrorM,
		metrics.AltitudeRMSEM,
		metrics.MeanAltitudeErrorRatio,
	}
	for _, value := range values {
		if !nonNegativeFinite(value) {
			return false
		}
	}

	return true
}

func validMethodSummary(
	method MethodSummary,
) bool {
	if strings.TrimSpace(
		method.MethodName,
	) == "" ||
		strings.TrimSpace(
			method.MethodVersion,
		) == "" ||
		!method.DecisionClass.IsKnown() ||
		method.EvaluationCount < 1 ||
		method.CompleteEvaluationCount < 0 ||
		method.PartialEvaluationCount < 0 ||
		method.UnavailableEvaluationCount < 0 ||
		method.CompleteEvaluationCount+
			method.PartialEvaluationCount+
			method.UnavailableEvaluationCount !=
			method.EvaluationCount ||
		method.ForecastPointCount < 0 ||
		method.EvaluatedPointCount < 0 ||
		method.EvaluatedPointCount >
			method.ForecastPointCount ||
		method.AltitudeEvaluatedPointCount < 0 ||
		method.AltitudeEvaluatedPointCount >
			method.EvaluatedPointCount ||
		method.ArrivalEvaluationCount < 0 ||
		method.ArrivalEvaluationCount >
			method.EvaluationCount {
		return false
	}

	ratios := []float64{
		method.PointCoverageRatio,
		method.
			HorizontalUncertaintyCoverageRatio,
		method.
			VerticalUncertaintyCoverageRatio,
		method.
			ArrivalIntervalCoverageRatio,
	}
	for _, value := range ratios {
		if !unitInterval(value) {
			return false
		}
	}

	values := []float64{
		method.MeanHorizontalErrorM,
		method.MedianHorizontalErrorM,
		method.P95HorizontalErrorM,
		method.HorizontalRMSEM,
		method.MeanAltitudeAbsoluteErrorM,
		method.AltitudeRMSEM,
		method.MeanArrivalAbsoluteErrorSeconds,
	}
	for _, value := range values {
		if !nonNegativeFinite(value) {
			return false
		}
	}

	return true
}

func normalizeNotices(
	items []Notice,
) []Notice {
	seen := make(
		map[string]Notice,
		len(items),
	)
	for _, item := range items {
		code := strings.TrimSpace(
			item.Code,
		)
		message := strings.TrimSpace(
			item.Message,
		)
		if code == "" || message == "" {
			continue
		}
		key := code + "\x00" +
			message
		seen[key] = Notice{
			Code:    code,
			Message: message,
		}
	}

	keys := make(
		[]string,
		0,
		len(seen),
	)
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(
		[]Notice,
		0,
		len(keys),
	)
	for _, key := range keys {
		result = append(
			result,
			seen[key],
		)
	}

	return result
}

func cloneConfidence(
	value projectioncontract.Confidence,
) projectioncontract.Confidence {
	cloned := value
	cloned.Reasons = append(
		[]projectioncontract.ConfidenceReason(nil),
		value.Reasons...,
	)

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
