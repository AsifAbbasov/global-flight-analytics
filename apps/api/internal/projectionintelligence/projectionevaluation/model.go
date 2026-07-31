package projectionevaluation

import (
	"regexp"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

const (
	Version                     = "projection-replay-evaluation-v2"
	FingerprintVersion          = "projection-replay-evaluation-fingerprint-v2"
	ProjectionSnapshotVersion   = "projection-replay-projection-snapshot-v2"
	TruthSnapshotVersion        = "projection-replay-truth-snapshot-v2"
	AggregateVersion            = "projection-replay-aggregate-v2"
	AggregateFingerprintVersion = "projection-replay-aggregate-fingerprint-v2"
	TruthKnowledgeCutoffMode    = "point_availability_evidence"
)

type Status string

const (
	StatusUnavailable Status = "unavailable"
	StatusPartial     Status = "partial"
	StatusComplete    Status = "complete"
)

func (status Status) IsKnown() bool {
	switch status {
	case StatusUnavailable, StatusPartial, StatusComplete:
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
	case ActualPointSourceObserved, ActualPointSourceInterpolated:
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
	AvailableAt     time.Time
}

type TruthAvailability struct {
	PointID     string
	SourceName  string
	AvailableAt time.Time
}

type EvaluationPolicy struct {
	Version string

	MaximumInterpolationGap     time.Duration
	MaximumTruthGroundSpeedMPS  float64
	MaximumTruthVerticalRateMPS float64
	MinimumEvaluatedPointCount  int
	MaximumHorizontalErrorM     float64
	MaximumAltitudeErrorM       float64
	LeadTimeBucketSize          time.Duration

	InputFingerprint string
}

type TruthEvidenceSummary struct {
	KnowledgeCutoffMode string
	CanonicalPointCount int

	ExcludedAfterObservationCutoffCount   int
	ExcludedAfterAvailabilityCutoffCount  int
	RejectedImplausibleInterpolationCount int

	LatestObservedAt  time.Time
	LatestAvailableAt time.Time
	SourceNames       []string

	InputFingerprint string
}

func (summary TruthEvidenceSummary) Clone() TruthEvidenceSummary {
	cloned := summary
	cloned.SourceNames = append([]string(nil), summary.SourceNames...)
	return cloned
}

type PointEvaluation struct {
	Sequence     int
	ForecastTime time.Time
	LeadTime     time.Duration

	ActualSource ActualPointSource
	ActualTime   time.Time

	ForecastLatitude  float64
	ForecastLongitude float64
	ActualLatitude    float64
	ActualLongitude   float64

	ForecastHorizontalUncertaintyM float64
	HorizontalErrorM               float64
	HorizontalErrorRatio           float64
	WithinHorizontalUncertainty    bool

	ForecastAltitudeM            *float64
	ActualAltitudeM              *float64
	ForecastVerticalUncertaintyM *float64
	AltitudeAbsoluteErrorM       *float64
	AltitudeErrorRatio           *float64
	WithinVerticalUncertainty    *bool

	ForecastConfidence           projectioncontract.Confidence
	NormalizedHorizontalAccuracy float64
	AbsoluteConfidenceGap        float64
}

func (item PointEvaluation) Clone() PointEvaluation {
	cloned := item
	cloned.ForecastAltitudeM = cloneFloat(item.ForecastAltitudeM)
	cloned.ActualAltitudeM = cloneFloat(item.ActualAltitudeM)
	cloned.ForecastVerticalUncertaintyM = cloneFloat(item.ForecastVerticalUncertaintyM)
	cloned.AltitudeAbsoluteErrorM = cloneFloat(item.AltitudeAbsoluteErrorM)
	cloned.AltitudeErrorRatio = cloneFloat(item.AltitudeErrorRatio)
	cloned.WithinVerticalUncertainty = cloneBool(item.WithinVerticalUncertainty)
	cloned.ForecastConfidence = cloneConfidence(item.ForecastConfidence)
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

type EndpointMetrics struct {
	Available    bool
	ForecastTime time.Time

	HorizontalErrorM            float64
	HorizontalErrorRatio        float64
	WithinHorizontalUncertainty bool

	AltitudeAvailable         bool
	AltitudeAbsoluteErrorM    float64
	AltitudeErrorRatio        float64
	WithinVerticalUncertainty bool
}

type ConfidenceMetrics struct {
	EvaluatedPointCount int

	MeanForecastConfidence           float64
	MeanNormalizedHorizontalAccuracy float64
	MeanAbsoluteConfidenceGap        float64
	ConfidenceCalibrationRMSE        float64
}

type LeadTimeMetrics struct {
	BucketStart time.Duration
	BucketEnd   time.Duration

	EvaluatedPointCount int

	MeanHorizontalErrorM               float64
	MedianHorizontalErrorM             float64
	P95HorizontalErrorM                float64
	HorizontalRMSEM                    float64
	HorizontalUncertaintyCoverageRatio float64

	MeanForecastConfidence           float64
	MeanNormalizedHorizontalAccuracy float64
	MeanAbsoluteConfidenceGap        float64
}

type ArrivalMetrics struct {
	ActualTruthAvailable bool
	PredictionAvailable  bool
	AirportMatched       bool
	Available            bool

	PredictedAirportICAOCode string
	ActualAirportICAOCode    string
	ActualBoundaryTime       time.Time

	EarliestTime  time.Time
	EstimatedTime time.Time
	LatestTime    time.Time

	EstimatedAbsoluteErrorSeconds float64
	SignedErrorSeconds            float64
	IntervalWidthSeconds          float64
	IntervalCoveredActual         bool
}

func (metrics ArrivalMetrics) Clone() ArrivalMetrics { return metrics }

type Result struct {
	Version string
	Status  Status

	TrajectoryID             string
	ProjectionMethod         projectioncontract.Method
	ProjectionAsOfTime       time.Time
	ProjectionHorizonEndTime time.Time
	ForecastStep             time.Duration
	ProjectionGeneratedAt    time.Time
	EvaluatedAt              time.Time

	Policy        EvaluationPolicy
	TruthEvidence TruthEvidenceSummary

	ProjectionInputFingerprint    string
	ProjectionSnapshotFingerprint string
	TruthSnapshotFingerprint      string
	EvaluationInputFingerprint    string

	Points     []PointEvaluation
	Position   PositionMetrics
	Endpoint   EndpointMetrics
	Confidence ConfidenceMetrics
	LeadTimes  []LeadTimeMetrics
	Arrival    ArrivalMetrics

	Limitations []Notice
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Points = make([]PointEvaluation, len(result.Points))
	for index, point := range result.Points {
		cloned.Points[index] = point.Clone()
	}
	cloned.TruthEvidence = result.TruthEvidence.Clone()
	cloned.LeadTimes = append([]LeadTimeMetrics(nil), result.LeadTimes...)
	cloned.Arrival = result.Arrival.Clone()
	cloned.Limitations = append([]Notice(nil), result.Limitations...)
	return cloned
}

type MethodSummary struct {
	MethodName    string
	MethodVersion string
	DecisionClass projectioncontract.DecisionClass

	ProjectionHorizonDuration   time.Duration
	ForecastStep                time.Duration
	EvaluationPolicyVersion     string
	EvaluationPolicyFingerprint string
	LeadTimeBucketSize          time.Duration

	EvaluationCount                 int
	CompleteEvaluationCount         int
	PartialEvaluationCount          int
	UnavailableEvaluationCount      int
	AccuracyEligibleEvaluationCount int

	ForecastPointCount  int
	EvaluatedPointCount int
	PointCoverageRatio  float64

	MeanHorizontalErrorM               float64
	MedianHorizontalErrorM             float64
	P95HorizontalErrorM                float64
	HorizontalRMSEM                    float64
	HorizontalUncertaintyCoverageRatio float64

	TrajectoryMacroMeanHorizontalErrorM float64
	TrajectoryMacroMeanHorizontalRMSEM  float64

	AltitudeEvaluatedPointCount      int
	MeanAltitudeAbsoluteErrorM       float64
	AltitudeRMSEM                    float64
	VerticalUncertaintyCoverageRatio float64

	EndpointEvaluationCount         int
	MeanEndpointHorizontalErrorM    float64
	EndpointHorizontalRMSEM         float64
	EndpointAltitudeEvaluationCount int
	MeanEndpointAltitudeErrorM      float64

	ConfidenceEvaluationPointCount   int
	MeanForecastConfidence           float64
	MeanNormalizedHorizontalAccuracy float64
	MeanAbsoluteConfidenceGap        float64
	ConfidenceCalibrationRMSE        float64

	LeadTimes []LeadTimeMetrics

	ActualArrivalTruthCount          int
	ArrivalPredictionCount           int
	MatchedArrivalCount              int
	MissingArrivalPredictionCount    int
	UnexpectedArrivalPredictionCount int
	ArrivalAirportMismatchCount      int
	ArrivalPredictionRecall          float64
	ArrivalAirportAccuracy           float64

	ArrivalEvaluationCount          int
	MeanArrivalAbsoluteErrorSeconds float64
	ArrivalIntervalCoverageRatio    float64
}

func (summary MethodSummary) Clone() MethodSummary {
	cloned := summary
	cloned.LeadTimes = append([]LeadTimeMetrics(nil), summary.LeadTimes...)
	return cloned
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
	cloned.Methods = make([]MethodSummary, len(result.Methods))
	for index, method := range result.Methods {
		cloned.Methods[index] = method.Clone()
	}
	cloned.Limitations = append([]Notice(nil), result.Limitations...)
	return cloned
}

var fingerprintPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var airportICAOPattern = regexp.MustCompile(`^[A-Z0-9]{4}$`)
