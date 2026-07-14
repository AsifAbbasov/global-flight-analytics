package datasetprofiler

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

const Version = "flight-feature-dataset-profiler-v1"

type Config struct {
	TargetSchemaVersion flightfeatures.SchemaVersion
	Now                 func() time.Time
}

type Request struct {
	Features []flightfeatures.FlightFeatures
}

type Profile struct {
	Version             string
	SchemaVersion       flightfeatures.SchemaVersion
	GeneratedAt         time.Time
	TotalRecordCount    int
	AcceptedRecordCount int
	RejectedRecordCount int

	DuplicateSnapshotCount   int
	ConflictingSnapshotCount int

	Cardinality CardinalityProfile
	Validation  ValidationProfile
	Time        TimeProfile
	Quality     QualityProfile
	Groups      []GroupProfile
	Sources     []FrequencyProfile
	Limitations []LimitationProfile
	Rejections  []FrequencyProfile
}

type CardinalityProfile struct {
	UniqueTrajectoryCount int
	UniqueIdentityCount   int
	UniqueFlightCount     int
	UniqueAircraftCount   int
	UniqueICAO24Count     int
	UniqueCallsignCount   int
}

type ValidationProfile struct {
	ValidCount       int
	LimitedCount     int
	InvalidCount     int
	UnvalidatedCount int
	UnknownCount     int
}

type TimeProfile struct {
	EarliestWindowStart time.Time
	LatestWindowEnd     time.Time
	EarliestAsOfTime    time.Time
	LatestAsOfTime      time.Time
	EarliestExtractedAt time.Time
	LatestExtractedAt   time.Time
}

type QualityProfile struct {
	CompletenessScore NumericProfile
	InputQualityScore NumericProfile
	SupportingPoints  NumericProfile
}

type NumericProfile struct {
	Count        int
	InvalidCount int
	Minimum      float64
	Maximum      float64
	Mean         float64
	Median       float64
	Percentile95 float64
}

type GroupProfile struct {
	Group                     flightfeatures.FeatureGroup
	SchemaFieldCount          int
	RecordCount               int
	AvailableCount            int
	PartialCount              int
	UnavailableCount          int
	UnknownStatusCount        int
	AvailableRatio            float64
	PartialRatio              float64
	UnavailableRatio          float64
	MeanFieldCompleteness     float64
	MeanSupportingPointCount  float64
	LimitationOccurrenceCount int
}

type FrequencyProfile struct {
	Value               string
	OccurrenceCount     int
	AffectedRecordCount int
}

type LimitationProfile struct {
	Code                string
	OccurrenceCount     int
	AffectedRecordCount int
}

func (profile Profile) Clone() Profile {
	cloned := profile
	cloned.Groups = append(
		[]GroupProfile(nil),
		profile.Groups...,
	)
	cloned.Sources = append(
		[]FrequencyProfile(nil),
		profile.Sources...,
	)
	cloned.Limitations = append(
		[]LimitationProfile(nil),
		profile.Limitations...,
	)
	cloned.Rejections = append(
		[]FrequencyProfile(nil),
		profile.Rejections...,
	)

	return cloned
}
