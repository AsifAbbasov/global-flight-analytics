package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceregionanalytics"
)

type AirspaceRegionAnalyticsResponse struct {
	Version       string    `json:"version"`
	SchemaVersion string    `json:"schema_version"`
	Status        string    `json:"status"`
	RegionCode    string    `json:"region_code"`
	WindowStart   time.Time `json:"window_start"`
	WindowEnd     time.Time `json:"window_end"`

	Occupancy        AirspaceOccupancyResponse          `json:"occupancy"`
	SectorComplexity []AirspaceSectorComplexityResponse `json:"sector_complexity"`
	Metrics          AirspaceRegionMetricsResponse      `json:"metrics"`
	Confidence       AirspaceConfidenceResponse         `json:"confidence"`
	Limitations      []AirspaceLimitationResponse       `json:"limitations"`
	Explanations     []AirspaceExplanationResponse      `json:"explanations"`
	ScopeGuard       string                             `json:"scope_guard"`
	Provenance       AirspaceProvenanceResponse         `json:"provenance"`
	GeneratedAt      time.Time                          `json:"generated_at"`
}

type AirspaceOccupancyResponse struct {
	BucketDurationSeconds int64                             `json:"bucket_duration_seconds"`
	LatitudeCellDegrees   float64                           `json:"latitude_cell_degrees"`
	LongitudeCellDegrees  float64                           `json:"longitude_cell_degrees"`
	AltitudeBandMeters    float64                           `json:"altitude_band_meters"`
	Buckets               []AirspaceOccupancyBucketResponse `json:"buckets"`
	Metrics               AirspaceOccupancyMetricsResponse  `json:"metrics"`
}

type AirspaceOccupancyBucketResponse struct {
	ID        string                                 `json:"id"`
	StartTime time.Time                              `json:"start_time"`
	EndTime   time.Time                              `json:"end_time"`
	Cells     []AirspaceOccupancyCellResponse        `json:"cells"`
	Metrics   AirspaceOccupancyBucketMetricsResponse `json:"metrics"`
}

type AirspaceOccupancyCellResponse struct {
	ID                string    `json:"id"`
	BucketID          string    `json:"bucket_id"`
	BucketStart       time.Time `json:"bucket_start"`
	BucketEnd         time.Time `json:"bucket_end"`
	LatitudeIndex     int       `json:"latitude_index"`
	LongitudeIndex    int       `json:"longitude_index"`
	AltitudeBandIndex int       `json:"altitude_band_index"`
	AltitudeKnown     bool      `json:"altitude_known"`
	AircraftNodeIDs   []string  `json:"aircraft_node_ids"`
	AircraftCount     int       `json:"aircraft_count"`
	MeanQualityScore  float64   `json:"mean_quality_score"`
}

type AirspaceOccupancyBucketMetricsResponse struct {
	AircraftCount        int     `json:"aircraft_count"`
	OccupiedCellCount    int     `json:"occupied_cell_count"`
	UnknownAltitudeCount int     `json:"unknown_altitude_count"`
	MeanQualityScore     float64 `json:"mean_quality_score"`
}

type AirspaceOccupancyMetricsResponse struct {
	BucketCount              int     `json:"bucket_count"`
	ExpectedBucketCount      int     `json:"expected_bucket_count"`
	OccupiedCellCount        int     `json:"occupied_cell_count"`
	AircraftObservationCount int     `json:"aircraft_observation_count"`
	UniqueAircraftCount      int     `json:"unique_aircraft_count"`
	UnknownAltitudeCount     int     `json:"unknown_altitude_count"`
	PeakAircraftPerBucket    int     `json:"peak_aircraft_per_bucket"`
	PeakOccupiedCells        int     `json:"peak_occupied_cells"`
	MeanAircraftPerBucket    float64 `json:"mean_aircraft_per_bucket"`
	TemporalCoverage         float64 `json:"temporal_coverage"`
}

type AirspaceSectorComplexityResponse struct {
	ID                     string                           `json:"id"`
	BucketID               string                           `json:"bucket_id"`
	BucketStart            time.Time                        `json:"bucket_start"`
	BucketEnd              time.Time                        `json:"bucket_end"`
	LatitudeIndex          int                              `json:"latitude_index"`
	LongitudeIndex         int                              `json:"longitude_index"`
	AircraftNodeIDs        []string                         `json:"aircraft_node_ids"`
	AircraftCount          int                              `json:"aircraft_count"`
	AltitudeBandCount      int                              `json:"altitude_band_count"`
	UnknownAltitudeCount   int                              `json:"unknown_altitude_count"`
	CandidatePairCount     int                              `json:"candidate_pair_count"`
	ConvergingPairCount    int                              `json:"converging_pair_count"`
	ContextualRiskCount    int                              `json:"contextual_risk_count"`
	ElevatedRiskCount      int                              `json:"elevated_risk_count"`
	HighRiskCount          int                              `json:"high_risk_count"`
	IndeterminateRiskCount int                              `json:"indeterminate_risk_count"`
	HeadingDispersion      float64                          `json:"heading_dispersion"`
	SpeedVariability       float64                          `json:"speed_variability"`
	Score                  float64                          `json:"score"`
	Level                  string                           `json:"level"`
	Components             []AirspaceScoreComponentResponse `json:"components"`
	Confidence             AirspaceConfidenceResponse       `json:"confidence"`
	Limitations            []AirspaceLimitationResponse     `json:"limitations"`
	Explanations           []AirspaceExplanationResponse    `json:"explanations"`
}

type AirspaceRegionMetricsResponse struct {
	SnapshotCount             int     `json:"snapshot_count"`
	BucketCount               int     `json:"bucket_count"`
	UniqueAircraftCount       int     `json:"unique_aircraft_count"`
	AircraftObservationCount  int     `json:"aircraft_observation_count"`
	OccupiedCellCount         int     `json:"occupied_cell_count"`
	SectorReportCount         int     `json:"sector_report_count"`
	CurrentAircraftCount      int     `json:"current_aircraft_count"`
	PeakAircraftPerBucket     int     `json:"peak_aircraft_per_bucket"`
	MeanAircraftPerBucket     float64 `json:"mean_aircraft_per_bucket"`
	MeanComplexityScore       float64 `json:"mean_complexity_score"`
	PeakComplexityScore       float64 `json:"peak_complexity_score"`
	AirspacePressureIndex     float64 `json:"airspace_pressure_index"`
	PeakAirspacePressureIndex float64 `json:"peak_airspace_pressure_index"`
	ModerateSectorCount       int     `json:"moderate_sector_count"`
	HighSectorCount           int     `json:"high_sector_count"`
	SevereSectorCount         int     `json:"severe_sector_count"`
	ContextualRiskCount       int     `json:"contextual_risk_count"`
	ElevatedRiskCount         int     `json:"elevated_risk_count"`
	HighRiskCount             int     `json:"high_risk_count"`
	IndeterminateRiskCount    int     `json:"indeterminate_risk_count"`
	UnknownAltitudeCount      int     `json:"unknown_altitude_count"`
	TemporalCoverage          float64 `json:"temporal_coverage"`
	OccupancyTrend            string  `json:"occupancy_trend"`
	HighestComplexityLevel    string  `json:"highest_complexity_level"`
}

type AirspaceScoreComponentResponse struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Weight float64 `json:"weight"`
}

type AirspaceConfidenceReasonResponse struct {
	Code         string  `json:"code"`
	Message      string  `json:"message"`
	Contribution float64 `json:"contribution"`
}

type AirspaceConfidenceResponse struct {
	Score      float64                            `json:"score"`
	Level      string                             `json:"level"`
	Components []AirspaceScoreComponentResponse   `json:"components"`
	Reasons    []AirspaceConfidenceReasonResponse `json:"reasons"`
}

type AirspaceLimitationResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Scope   string `json:"scope"`
}

type AirspaceExplanationResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AirspaceProvenanceResponse struct {
	InputFingerprint  string    `json:"input_fingerprint"`
	SceneFingerprints []string  `json:"scene_fingerprints"`
	ScanFingerprints  []string  `json:"scan_fingerprints"`
	RiskFingerprints  []string  `json:"risk_fingerprints"`
	SourceNames       []string  `json:"source_names"`
	LatestObservedAt  time.Time `json:"latest_observed_at"`
}

func ToAirspaceRegionAnalyticsResponse(
	result airspaceregionanalytics.Result,
) AirspaceRegionAnalyticsResponse {
	response := AirspaceRegionAnalyticsResponse{
		Version:       airspaceproduction.Version,
		SchemaVersion: string(result.SchemaVersion),
		Status:        string(result.Status),
		RegionCode:    result.RegionCode,
		WindowStart:   result.WindowStart.UTC(),
		WindowEnd:     result.WindowEnd.UTC(),
		Occupancy:     toAirspaceOccupancyResponse(result.Occupancy),
		SectorComplexity: make(
			[]AirspaceSectorComplexityResponse,
			0,
			len(result.SectorComplexity),
		),
		Metrics:      toAirspaceRegionMetricsResponse(result.Metrics),
		Confidence:   toAirspaceConfidenceResponse(result.Confidence),
		Limitations:  toAirspaceLimitations(result.Limitations),
		Explanations: toAirspaceExplanations(result.Explanations),
		ScopeGuard:   string(result.ScopeGuard),
		Provenance: AirspaceProvenanceResponse{
			InputFingerprint:  result.Provenance.InputFingerprint,
			SceneFingerprints: append([]string(nil), result.Provenance.SceneFingerprints...),
			ScanFingerprints:  append([]string(nil), result.Provenance.ScanFingerprints...),
			RiskFingerprints:  append([]string(nil), result.Provenance.RiskFingerprints...),
			SourceNames:       append([]string(nil), result.Provenance.SourceNames...),
			LatestObservedAt:  result.Provenance.LatestObservedAt.UTC(),
		},
		GeneratedAt: result.GeneratedAt.UTC(),
	}
	for _, report := range result.SectorComplexity {
		response.SectorComplexity = append(
			response.SectorComplexity,
			toAirspaceSectorComplexityResponse(report),
		)
	}
	return response
}

func toAirspaceOccupancyResponse(
	index airspaceregionanalytics.TemporalOccupancyIndex,
) AirspaceOccupancyResponse {
	response := AirspaceOccupancyResponse{
		BucketDurationSeconds: int64(index.BucketDuration / time.Second),
		LatitudeCellDegrees:   index.LatitudeCellDegrees,
		LongitudeCellDegrees:  index.LongitudeCellDegrees,
		AltitudeBandMeters:    index.AltitudeBandMeters,
		Buckets: make(
			[]AirspaceOccupancyBucketResponse,
			0,
			len(index.Buckets),
		),
		Metrics: AirspaceOccupancyMetricsResponse{
			BucketCount:              index.Metrics.BucketCount,
			ExpectedBucketCount:      index.Metrics.ExpectedBucketCount,
			OccupiedCellCount:        index.Metrics.OccupiedCellCount,
			AircraftObservationCount: index.Metrics.AircraftObservationCount,
			UniqueAircraftCount:      index.Metrics.UniqueAircraftCount,
			UnknownAltitudeCount:     index.Metrics.UnknownAltitudeCount,
			PeakAircraftPerBucket:    index.Metrics.PeakAircraftPerBucket,
			PeakOccupiedCells:        index.Metrics.PeakOccupiedCells,
			MeanAircraftPerBucket:    index.Metrics.MeanAircraftPerBucket,
			TemporalCoverage:         index.Metrics.TemporalCoverage,
		},
	}
	for _, bucket := range index.Buckets {
		bucketResponse := AirspaceOccupancyBucketResponse{
			ID:        bucket.ID,
			StartTime: bucket.StartTime.UTC(),
			EndTime:   bucket.EndTime.UTC(),
			Cells: make(
				[]AirspaceOccupancyCellResponse,
				0,
				len(bucket.Cells),
			),
			Metrics: AirspaceOccupancyBucketMetricsResponse{
				AircraftCount:        bucket.Metrics.AircraftCount,
				OccupiedCellCount:    bucket.Metrics.OccupiedCellCount,
				UnknownAltitudeCount: bucket.Metrics.UnknownAltitudeCount,
				MeanQualityScore:     bucket.Metrics.MeanQualityScore,
			},
		}
		for _, cell := range bucket.Cells {
			bucketResponse.Cells = append(
				bucketResponse.Cells,
				AirspaceOccupancyCellResponse{
					ID:                cell.ID,
					BucketID:          cell.BucketID,
					BucketStart:       cell.BucketStart.UTC(),
					BucketEnd:         cell.BucketEnd.UTC(),
					LatitudeIndex:     cell.LatitudeIndex,
					LongitudeIndex:    cell.LongitudeIndex,
					AltitudeBandIndex: cell.AltitudeBandIndex,
					AltitudeKnown:     cell.AltitudeKnown,
					AircraftNodeIDs:   append([]string(nil), cell.AircraftNodeIDs...),
					AircraftCount:     cell.AircraftCount,
					MeanQualityScore:  cell.MeanQualityScore,
				},
			)
		}
		response.Buckets = append(response.Buckets, bucketResponse)
	}
	return response
}

func toAirspaceSectorComplexityResponse(
	report airspaceregionanalytics.SectorComplexityReport,
) AirspaceSectorComplexityResponse {
	return AirspaceSectorComplexityResponse{
		ID:                     report.ID,
		BucketID:               report.BucketID,
		BucketStart:            report.BucketStart.UTC(),
		BucketEnd:              report.BucketEnd.UTC(),
		LatitudeIndex:          report.LatitudeIndex,
		LongitudeIndex:         report.LongitudeIndex,
		AircraftNodeIDs:        append([]string(nil), report.AircraftNodeIDs...),
		AircraftCount:          report.AircraftCount,
		AltitudeBandCount:      report.AltitudeBandCount,
		UnknownAltitudeCount:   report.UnknownAltitudeCount,
		CandidatePairCount:     report.CandidatePairCount,
		ConvergingPairCount:    report.ConvergingPairCount,
		ContextualRiskCount:    report.ContextualRiskCount,
		ElevatedRiskCount:      report.ElevatedRiskCount,
		HighRiskCount:          report.HighRiskCount,
		IndeterminateRiskCount: report.IndeterminateRiskCount,
		HeadingDispersion:      report.HeadingDispersion,
		SpeedVariability:       report.SpeedVariability,
		Score:                  report.Score,
		Level:                  string(report.Level),
		Components:             toAirspaceScoreComponents(report.Components),
		Confidence:             toAirspaceConfidenceResponse(report.Confidence),
		Limitations:            toAirspaceLimitations(report.Limitations),
		Explanations:           toAirspaceExplanations(report.Explanations),
	}
}

func toAirspaceRegionMetricsResponse(
	metrics airspaceregionanalytics.RegionMetrics,
) AirspaceRegionMetricsResponse {
	return AirspaceRegionMetricsResponse{
		SnapshotCount:             metrics.SnapshotCount,
		BucketCount:               metrics.BucketCount,
		UniqueAircraftCount:       metrics.UniqueAircraftCount,
		AircraftObservationCount:  metrics.AircraftObservationCount,
		OccupiedCellCount:         metrics.OccupiedCellCount,
		SectorReportCount:         metrics.SectorReportCount,
		CurrentAircraftCount:      metrics.CurrentAircraftCount,
		PeakAircraftPerBucket:     metrics.PeakAircraftPerBucket,
		MeanAircraftPerBucket:     metrics.MeanAircraftPerBucket,
		MeanComplexityScore:       metrics.MeanComplexityScore,
		PeakComplexityScore:       metrics.PeakComplexityScore,
		AirspacePressureIndex:     metrics.AirspacePressureIndex,
		PeakAirspacePressureIndex: metrics.PeakAirspacePressureIndex,
		ModerateSectorCount:       metrics.ModerateSectorCount,
		HighSectorCount:           metrics.HighSectorCount,
		SevereSectorCount:         metrics.SevereSectorCount,
		ContextualRiskCount:       metrics.ContextualRiskCount,
		ElevatedRiskCount:         metrics.ElevatedRiskCount,
		HighRiskCount:             metrics.HighRiskCount,
		IndeterminateRiskCount:    metrics.IndeterminateRiskCount,
		UnknownAltitudeCount:      metrics.UnknownAltitudeCount,
		TemporalCoverage:          metrics.TemporalCoverage,
		OccupancyTrend:            string(metrics.OccupancyTrend),
		HighestComplexityLevel:    string(metrics.HighestComplexityLevel),
	}
}

func toAirspaceConfidenceResponse(
	confidence airspaceregionanalytics.Confidence,
) AirspaceConfidenceResponse {
	return AirspaceConfidenceResponse{
		Score:      confidence.Score,
		Level:      string(confidence.Level),
		Components: toAirspaceScoreComponents(confidence.Components),
		Reasons: func() []AirspaceConfidenceReasonResponse {
			result := make(
				[]AirspaceConfidenceReasonResponse,
				0,
				len(confidence.Reasons),
			)
			for _, reason := range confidence.Reasons {
				result = append(result, AirspaceConfidenceReasonResponse{
					Code:         reason.Code,
					Message:      reason.Message,
					Contribution: reason.Contribution,
				})
			}
			return result
		}(),
	}
}

func toAirspaceScoreComponents(
	components []airspaceregionanalytics.ScoreComponent,
) []AirspaceScoreComponentResponse {
	result := make(
		[]AirspaceScoreComponentResponse,
		0,
		len(components),
	)
	for _, component := range components {
		result = append(result, AirspaceScoreComponentResponse{
			Name:   component.Name,
			Score:  component.Score,
			Weight: component.Weight,
		})
	}
	return result
}

func toAirspaceLimitations(
	limitations []airspaceregionanalytics.Limitation,
) []AirspaceLimitationResponse {
	result := make(
		[]AirspaceLimitationResponse,
		0,
		len(limitations),
	)
	for _, limitation := range limitations {
		result = append(result, AirspaceLimitationResponse{
			Code:    limitation.Code,
			Message: limitation.Message,
			Scope:   limitation.Scope,
		})
	}
	return result
}

func toAirspaceExplanations(
	explanations []airspaceregionanalytics.Explanation,
) []AirspaceExplanationResponse {
	result := make(
		[]AirspaceExplanationResponse,
		0,
		len(explanations),
	)
	for _, explanation := range explanations {
		result = append(result, AirspaceExplanationResponse{
			Code:    explanation.Code,
			Message: explanation.Message,
		})
	}
	return result
}
