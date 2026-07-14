package flightfeatures

import "time"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "flight-features-v1"

type AvailabilityStatus string

const (
	AvailabilityStatusAvailable   AvailabilityStatus = "available"
	AvailabilityStatusPartial     AvailabilityStatus = "partial"
	AvailabilityStatusUnavailable AvailabilityStatus = "unavailable"
)

type ValidationStatus string

const (
	ValidationStatusUnvalidated ValidationStatus = "unvalidated"
	ValidationStatusValid       ValidationStatus = "valid"
	ValidationStatusLimited     ValidationStatus = "limited"
	ValidationStatusInvalid     ValidationStatus = "invalid"
)

type FeatureWindow struct {
	StartTime time.Time
	EndTime   time.Time
	AsOfTime  time.Time
}

type GroupEvidence struct {
	Status               AvailabilityStatus
	AvailableFieldCount  int
	TotalFieldCount      int
	SupportingPointCount int
	Limitations          []FeatureLimitation
}

type TemporalFeatures struct {
	Evidence            GroupEvidence
	DurationSeconds     int64
	StartHourUTC        int
	EndHourUTC          int
	StartWeekday        time.Weekday
	EndWeekday          time.Weekday
	StartMinuteOfDayUTC int
	EndMinuteOfDayUTC   int
	CrossesUTCMidnight  bool
}

type GeographicalFeatures struct {
	Evidence                  GroupEvidence
	StartLatitude             float64
	StartLongitude            float64
	EndLatitude               float64
	EndLongitude              float64
	MinimumLatitude           float64
	MaximumLatitude           float64
	MinimumLongitude          float64
	MaximumLongitude          float64
	LatitudeSpanDegrees       float64
	LongitudeSpanDegrees      float64
	GreatCircleDistanceKM     float64
	ObservedPathDistanceKM    float64
	MaximumDisplacementKM     float64
	CrossesAntimeridian       bool
	UniqueGeographicCellCount int
	GeographicCellPrecision   int
}

type OperationalFeatures struct {
	Evidence                       GroupEvidence
	MinimumAltitudeM               float64
	MaximumAltitudeM               float64
	MeanAltitudeM                  float64
	AltitudeRangeM                 float64
	MeanVelocityMPS                float64
	MaximumVelocityMPS             float64
	MeanAbsoluteVerticalRateMPS    float64
	MaximumAbsoluteVerticalRateMPS float64
	HeadingChangeDegrees           float64
	GroundObservationShare         float64
	AirborneObservationShare       float64
}

type TrajectoryFeatures struct {
	Evidence                    GroupEvidence
	PointCount                  int
	SegmentCount                int
	CoverageGapCount            int
	TrajectoryQualityScore      float64
	ObservedSegmentCount        int
	InterpolatedSegmentCount    int
	EstimatedSegmentCount       int
	InvalidSegmentCount         int
	ObservedSegmentShare        float64
	InterpolatedSegmentShare    float64
	EstimatedSegmentShare       float64
	InvalidSegmentShare         float64
	MeanSamplingIntervalSeconds float64
	MaximumSamplingGapSeconds   float64
	CoverageRatio               float64
	PathEfficiencyRatio         float64
}

type AircraftFeatures struct {
	Evidence     GroupEvidence
	Registration string
	Manufacturer string
	Model        string
	AircraftType string
	Airline      string
	Country      string
}

type FeatureLimitation struct {
	Code    string
	Message string
}

type FeatureQuality struct {
	Status               ValidationStatus
	CompletenessScore    float64
	InputQualityScore    float64
	SupportingPointCount int
	Limitations          []FeatureLimitation
}

type FeatureProvenance struct {
	ExtractorVersion    string
	InputFingerprint    string
	TrajectoryUpdatedAt time.Time
	SourceNames         []string
}

type FlightFeatures struct {
	SchemaVersion SchemaVersion
	TrajectoryID  string
	IdentityKey   string
	FlightID      string
	AircraftID    string
	ICAO24        string
	Callsign      string
	Window        FeatureWindow
	ExtractedAt   time.Time

	Temporal     TemporalFeatures
	Geographical GeographicalFeatures
	Operational  OperationalFeatures
	Trajectory   TrajectoryFeatures
	Aircraft     AircraftFeatures

	Quality    FeatureQuality
	Provenance FeatureProvenance
}

func (features FlightFeatures) Clone() FlightFeatures {
	cloned := features

	cloned.Temporal.Evidence = cloneGroupEvidence(
		features.Temporal.Evidence,
	)
	cloned.Geographical.Evidence = cloneGroupEvidence(
		features.Geographical.Evidence,
	)
	cloned.Operational.Evidence = cloneGroupEvidence(
		features.Operational.Evidence,
	)
	cloned.Trajectory.Evidence = cloneGroupEvidence(
		features.Trajectory.Evidence,
	)
	cloned.Aircraft.Evidence = cloneGroupEvidence(
		features.Aircraft.Evidence,
	)
	cloned.Quality.Limitations = cloneLimitations(
		features.Quality.Limitations,
	)
	cloned.Provenance.SourceNames = append(
		[]string(nil),
		features.Provenance.SourceNames...,
	)

	return cloned
}

func cloneGroupEvidence(
	evidence GroupEvidence,
) GroupEvidence {
	cloned := evidence
	cloned.Limitations = cloneLimitations(
		evidence.Limitations,
	)

	return cloned
}

func cloneLimitations(
	items []FeatureLimitation,
) []FeatureLimitation {
	return append(
		[]FeatureLimitation(nil),
		items...,
	)
}
