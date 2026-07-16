package weathercontract

import "time"

const Version = "weather-feature-contract-v1"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "weather-feature-v1"

type ResultStatus string

const (
	ResultStatusUnavailable ResultStatus = "unavailable"
	ResultStatusLimited     ResultStatus = "limited"
	ResultStatusComplete    ResultStatus = "complete"
)

func (status ResultStatus) IsKnown() bool {
	switch status {
	case ResultStatusUnavailable,
		ResultStatusLimited,
		ResultStatusComplete:
		return true
	default:
		return false
	}
}

type EvidenceKind string

const (
	EvidenceKindObservation EvidenceKind = "observation"
	EvidenceKindAnalysis    EvidenceKind = "analysis"
	EvidenceKindForecast    EvidenceKind = "forecast"
)

func (kind EvidenceKind) IsKnown() bool {
	switch kind {
	case EvidenceKindObservation,
		EvidenceKindAnalysis,
		EvidenceKindForecast:
		return true
	default:
		return false
	}
}

type VerticalReference string

const (
	VerticalReferenceSurface       VerticalReference = "surface"
	VerticalReferenceMeanSeaLevel  VerticalReference = "mean_sea_level"
	VerticalReferencePressureLevel VerticalReference = "pressure_level"
	VerticalReferenceUnknown       VerticalReference = "unknown"
)

func (reference VerticalReference) IsKnown() bool {
	switch reference {
	case VerticalReferenceSurface,
		VerticalReferenceMeanSeaLevel,
		VerticalReferencePressureLevel,
		VerticalReferenceUnknown:
		return true
	default:
		return false
	}
}

type ScopeGuard string

const (
	ScopeGuardContextOnly ScopeGuard = "weather_context_only_not_proof_of_cause"
)

type ConfidenceLevel string

const (
	ConfidenceLevelNone   ConfidenceLevel = "none"
	ConfidenceLevelLow    ConfidenceLevel = "low"
	ConfidenceLevelMedium ConfidenceLevel = "medium"
	ConfidenceLevelHigh   ConfidenceLevel = "high"
)

func (level ConfidenceLevel) IsKnown() bool {
	switch level {
	case ConfidenceLevelNone,
		ConfidenceLevelLow,
		ConfidenceLevelMedium,
		ConfidenceLevelHigh:
		return true
	default:
		return false
	}
}

func LevelForScore(score float64) ConfidenceLevel {
	switch {
	case score <= 0:
		return ConfidenceLevelNone
	case score < 0.50:
		return ConfidenceLevelLow
	case score < 0.80:
		return ConfidenceLevelMedium
	default:
		return ConfidenceLevelHigh
	}
}

type Position struct {
	Latitude  float64
	Longitude float64

	AltitudeMeters    *float64
	VerticalReference VerticalReference
}

type Source struct {
	Provider     string
	Dataset      string
	EvidenceKind EvidenceKind

	HorizontalResolutionKilometers *float64
	TemporalResolution             time.Duration
}

type FeatureVector struct {
	TemperatureCelsius       *float64
	RelativeHumidityPercent  *float64
	PrecipitationMillimeters *float64
	RainMillimeters          *float64
	CloudCoverPercent        *float64
	SurfacePressureHPA       *float64
	WindSpeedMetersPerSecond *float64
	WindDirectionDegrees     *float64
	WindGustsMetersPerSecond *float64

	ConditionCode       *int
	ConditionCodeScheme string
}

func (features FeatureVector) PresentCount() int {
	values := []*float64{
		features.TemperatureCelsius,
		features.RelativeHumidityPercent,
		features.PrecipitationMillimeters,
		features.RainMillimeters,
		features.CloudCoverPercent,
		features.SurfacePressureHPA,
		features.WindSpeedMetersPerSecond,
		features.WindDirectionDegrees,
		features.WindGustsMetersPerSecond,
	}

	count := 0
	for _, value := range values {
		if value != nil {
			count++
		}
	}
	if features.ConditionCode != nil {
		count++
	}

	return count
}

type Sample struct {
	Sequence int

	Position Position
	Source   Source
	Features FeatureVector

	ValidAt     time.Time
	AvailableAt time.Time
	RetrievedAt time.Time
}

type ConfidenceReason struct {
	Code         string
	Message      string
	Contribution float64
}

type Confidence struct {
	Score   float64
	Level   ConfidenceLevel
	Reasons []ConfidenceReason
}

type Limitation struct {
	Code    string
	Message string
	Scope   string
}

type Explanation struct {
	Code    string
	Message string
}

type Provenance struct {
	InputFingerprint  string
	SourceNames       []string
	LatestAvailableAt time.Time
}

type Result struct {
	SchemaVersion SchemaVersion
	Status        ResultStatus

	TrajectoryID string
	AsOfTime     time.Time

	Samples []Sample

	Confidence   Confidence
	Limitations  []Limitation
	Explanations []Explanation
	ScopeGuard   ScopeGuard
	Provenance   Provenance
	GeneratedAt  time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Samples = make(
		[]Sample,
		0,
		len(result.Samples),
	)
	for _, sample := range result.Samples {
		cloned.Samples = append(
			cloned.Samples,
			cloneSample(sample),
		)
	}
	cloned.Confidence = cloneConfidence(
		result.Confidence,
	)
	cloned.Limitations = append(
		[]Limitation(nil),
		result.Limitations...,
	)
	cloned.Explanations = append(
		[]Explanation(nil),
		result.Explanations...,
	)
	cloned.Provenance.SourceNames = append(
		[]string(nil),
		result.Provenance.SourceNames...,
	)

	return cloned
}

func cloneSample(sample Sample) Sample {
	cloned := sample
	cloned.Position.AltitudeMeters = cloneFloat64(
		sample.Position.AltitudeMeters,
	)
	cloned.Source.HorizontalResolutionKilometers =
		cloneFloat64(
			sample.Source.
				HorizontalResolutionKilometers,
		)
	cloned.Features = cloneFeatureVector(
		sample.Features,
	)

	return cloned
}

func cloneFeatureVector(
	features FeatureVector,
) FeatureVector {
	return FeatureVector{
		TemperatureCelsius: cloneFloat64(
			features.TemperatureCelsius,
		),
		RelativeHumidityPercent: cloneFloat64(
			features.RelativeHumidityPercent,
		),
		PrecipitationMillimeters: cloneFloat64(
			features.PrecipitationMillimeters,
		),
		RainMillimeters: cloneFloat64(
			features.RainMillimeters,
		),
		CloudCoverPercent: cloneFloat64(
			features.CloudCoverPercent,
		),
		SurfacePressureHPA: cloneFloat64(
			features.SurfacePressureHPA,
		),
		WindSpeedMetersPerSecond: cloneFloat64(
			features.WindSpeedMetersPerSecond,
		),
		WindDirectionDegrees: cloneFloat64(
			features.WindDirectionDegrees,
		),
		WindGustsMetersPerSecond: cloneFloat64(
			features.WindGustsMetersPerSecond,
		),
		ConditionCode: cloneInt(
			features.ConditionCode,
		),
		ConditionCodeScheme: features.ConditionCodeScheme,
	}
}

func cloneConfidence(
	confidence Confidence,
) Confidence {
	cloned := confidence
	cloned.Reasons = append(
		[]ConfidenceReason(nil),
		confidence.Reasons...,
	)

	return cloned
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
