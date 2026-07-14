package dataqualitycontract

import "time"

type Provenance struct {
	SourceName       string    `json:"source_name"`
	SourceRecordTime time.Time `json:"source_record_time"`
	ReceivedAt       time.Time `json:"received_at"`
	IngestionRunID   string    `json:"ingestion_run_id"`
	Transformation   string    `json:"transformation"`
	AlgorithmVersion string    `json:"algorithm_version"`
	InputFingerprint string    `json:"input_fingerprint"`
}

type FreshnessStatus string

const (
	FreshnessStatusFresh   FreshnessStatus = "fresh"
	FreshnessStatusAging   FreshnessStatus = "aging"
	FreshnessStatusStale   FreshnessStatus = "stale"
	FreshnessStatusUnknown FreshnessStatus = "unknown"
)

type FreshnessInput struct {
	ObservedAt       time.Time
	EvaluatedAt      time.Time
	ExpectedInterval time.Duration
	StaleAfter       time.Duration
}

type Freshness struct {
	Score                   float64         `json:"score"`
	Status                  FreshnessStatus `json:"status"`
	AgeSeconds              float64         `json:"age_seconds"`
	ExpectedIntervalSeconds float64         `json:"expected_interval_seconds"`
	StaleAfterSeconds       float64         `json:"stale_after_seconds"`
	ObservedAt              time.Time       `json:"observed_at"`
	EvaluatedAt             time.Time       `json:"evaluated_at"`
	Explanation             string          `json:"explanation"`
}

type SamplingDensityInput struct {
	WindowStart      time.Time
	WindowEnd        time.Time
	ExpectedInterval time.Duration
	ObservationTimes []time.Time
}

type SamplingDensity struct {
	Score                float64       `json:"score"`
	ObservedSampleCount  int           `json:"observed_sample_count"`
	ExpectedSampleCount  int           `json:"expected_sample_count"`
	CoveredIntervalCount int           `json:"covered_interval_count"`
	TotalIntervalCount   int           `json:"total_interval_count"`
	DuplicateSampleCount int           `json:"duplicate_sample_count"`
	WindowStart          time.Time     `json:"window_start"`
	WindowEnd            time.Time     `json:"window_end"`
	ExpectedInterval     time.Duration `json:"expected_interval"`
	Explanation          string        `json:"explanation"`
}

type Permission struct {
	Allowed bool     `json:"allowed"`
	Reasons []string `json:"reasons"`
}

type AnalyticsPermissions struct {
	RouteInference       Permission `json:"route_inference"`
	PhaseDetection       Permission `json:"phase_detection"`
	HistoricalAnalytics  Permission `json:"historical_analytics"`
	HistoricalSimilarity Permission `json:"historical_similarity"`
	Projection           Permission `json:"projection"`
}

type Notice struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Report struct {
	ContractVersion string               `json:"contract_version"`
	Provenance      Provenance           `json:"provenance"`
	Freshness       Freshness            `json:"freshness"`
	SamplingDensity SamplingDensity      `json:"sampling_density"`
	Permissions     AnalyticsPermissions `json:"analytics_permissions"`
	MissingFields   []string             `json:"missing_fields"`
	Warnings        []Notice             `json:"warnings"`
	Limitations     []Notice             `json:"limitations"`
	EvaluatedAt     time.Time            `json:"evaluated_at"`
}
