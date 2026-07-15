package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

type HistoricalIntelligenceTimeWindow struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	AsOfTime  time.Time `json:"as_of_time"`
}

type HistoricalIntelligenceScope struct {
	Type                string `json:"type"`
	RegionCode          string `json:"region_code,omitempty"`
	AirportICAOCode     string `json:"airport_icao_code,omitempty"`
	OriginICAOCode      string `json:"origin_icao_code,omitempty"`
	DestinationICAOCode string `json:"destination_icao_code,omitempty"`
}

type HistoricalIntelligenceMetric struct {
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Aggregation string `json:"aggregation"`
}

type HistoricalIntelligenceConfidenceReason struct {
	Code         string  `json:"code"`
	Message      string  `json:"message"`
	Contribution float64 `json:"contribution"`
}

type HistoricalIntelligenceConfidence struct {
	Score       float64                                  `json:"score"`
	Level       string                                   `json:"level"`
	SampleCount int                                      `json:"sample_count"`
	Reasons     []HistoricalIntelligenceConfidenceReason `json:"reasons"`
}

type HistoricalIntelligenceLimitation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Scope   string `json:"scope"`
}

type HistoricalIntelligencePoint struct {
	StartTime     time.Time                          `json:"start_time"`
	EndTime       time.Time                          `json:"end_time"`
	Status        string                             `json:"status"`
	Value         float64                            `json:"value"`
	SampleCount   int                                `json:"sample_count"`
	CoverageRatio float64                            `json:"coverage_ratio"`
	Confidence    HistoricalIntelligenceConfidence   `json:"confidence"`
	Limitations   []HistoricalIntelligenceLimitation `json:"limitations"`
}

type HistoricalIntelligenceSummary struct {
	PointCount int     `json:"point_count"`
	Total      float64 `json:"total"`
	Minimum    float64 `json:"minimum"`
	Maximum    float64 `json:"maximum"`
	Average    float64 `json:"average"`
	Median     float64 `json:"median"`
}

type HistoricalIntelligencePeriodComparison struct {
	PreviousWindow   HistoricalIntelligenceTimeWindow `json:"previous_window"`
	CurrentValue     float64                          `json:"current_value"`
	PreviousValue    float64                          `json:"previous_value"`
	AbsoluteChange   float64                          `json:"absolute_change"`
	PercentageChange *float64                         `json:"percentage_change,omitempty"`
	Direction        string                           `json:"direction"`
}

type HistoricalIntelligenceProvenance struct {
	BuilderVersion        string    `json:"builder_version"`
	InputFingerprint      string    `json:"input_fingerprint"`
	SourceNames           []string  `json:"source_names"`
	LatestSourceUpdatedAt time.Time `json:"latest_source_updated_at"`
}

type HistoricalIntelligenceResult struct {
	SchemaVersion string                                  `json:"schema_version"`
	Status        string                                  `json:"status"`
	Metric        HistoricalIntelligenceMetric            `json:"metric"`
	Scope         HistoricalIntelligenceScope             `json:"scope"`
	Window        HistoricalIntelligenceTimeWindow        `json:"window"`
	Granularity   string                                  `json:"granularity"`
	Points        []HistoricalIntelligencePoint           `json:"points"`
	Summary       HistoricalIntelligenceSummary           `json:"summary"`
	Comparison    *HistoricalIntelligencePeriodComparison `json:"comparison,omitempty"`
	Confidence    HistoricalIntelligenceConfidence        `json:"confidence"`
	Limitations   []HistoricalIntelligenceLimitation      `json:"limitations"`
	Provenance    HistoricalIntelligenceProvenance        `json:"provenance"`
	GeneratedAt   time.Time                               `json:"generated_at"`
}

type HistoricalIntelligenceAggregateRecord struct {
	ID               string                       `json:"id"`
	InputFingerprint string                       `json:"input_fingerprint"`
	StoredAt         time.Time                    `json:"stored_at"`
	Result           HistoricalIntelligenceResult `json:"result"`
}

type HistoricalIntelligenceAggregateHistory struct {
	Items               []HistoricalIntelligenceAggregateRecord `json:"items"`
	HasMore             bool                                    `json:"has_more"`
	NextBeforeWindowEnd *time.Time                              `json:"next_before_window_end,omitempty"`
}

func ToHistoricalIntelligenceAggregateRecord(
	record historicalaggregate.Record,
) HistoricalIntelligenceAggregateRecord {
	return HistoricalIntelligenceAggregateRecord{
		ID:               record.ID,
		InputFingerprint: record.InputFingerprint,
		StoredAt:         record.StoredAt.UTC(),
		Result: toHistoricalIntelligenceResult(
			record.Result,
		),
	}
}

func ToHistoricalIntelligenceAggregateHistory(
	page historicalaggregate.Page,
) HistoricalIntelligenceAggregateHistory {
	items := make(
		[]HistoricalIntelligenceAggregateRecord,
		0,
		len(page.Records),
	)
	for _, record := range page.Records {
		items = append(
			items,
			ToHistoricalIntelligenceAggregateRecord(
				record,
			),
		)
	}

	var nextBeforeWindowEnd *time.Time
	if page.HasMore && len(page.Records) > 0 {
		value := page.Records[len(page.Records)-1].
			Key.Window.EndTime.UTC()
		nextBeforeWindowEnd = &value
	}

	return HistoricalIntelligenceAggregateHistory{
		Items:               items,
		HasMore:             page.HasMore,
		NextBeforeWindowEnd: nextBeforeWindowEnd,
	}
}

func toHistoricalIntelligenceResult(
	result historicalcontract.Result,
) HistoricalIntelligenceResult {
	points := make(
		[]HistoricalIntelligencePoint,
		0,
		len(result.Points),
	)
	for _, point := range result.Points {
		points = append(
			points,
			HistoricalIntelligencePoint{
				StartTime: point.StartTime.UTC(),
				EndTime:   point.EndTime.UTC(),
				Status:    string(point.Status),
				Value:     point.Value,
				SampleCount: point.
					SampleCount,
				CoverageRatio: point.
					CoverageRatio,
				Confidence: toHistoricalIntelligenceConfidence(
					point.Confidence,
				),
				Limitations: toHistoricalIntelligenceLimitations(
					point.Limitations,
				),
			},
		)
	}

	return HistoricalIntelligenceResult{
		SchemaVersion: string(
			result.SchemaVersion,
		),
		Status: string(result.Status),
		Metric: HistoricalIntelligenceMetric{
			Name: string(result.Metric.Name),
			Unit: result.Metric.Unit,
			Aggregation: string(
				result.Metric.Aggregation,
			),
		},
		Scope: toHistoricalIntelligenceScope(
			result.Scope,
		),
		Window: toHistoricalIntelligenceTimeWindow(
			result.Window,
		),
		Granularity: string(
			result.Granularity,
		),
		Points: points,
		Summary: HistoricalIntelligenceSummary{
			PointCount: result.Summary.PointCount,
			Total:      result.Summary.Total,
			Minimum:    result.Summary.Minimum,
			Maximum:    result.Summary.Maximum,
			Average:    result.Summary.Average,
			Median:     result.Summary.Median,
		},
		Comparison: toHistoricalIntelligenceComparison(
			result.Comparison,
		),
		Confidence: toHistoricalIntelligenceConfidence(
			result.Confidence,
		),
		Limitations: toHistoricalIntelligenceLimitations(
			result.Limitations,
		),
		Provenance: HistoricalIntelligenceProvenance{
			BuilderVersion: result.
				Provenance.BuilderVersion,
			InputFingerprint: result.
				Provenance.InputFingerprint,
			SourceNames: append(
				[]string(nil),
				result.Provenance.
					SourceNames...,
			),
			LatestSourceUpdatedAt: result.
				Provenance.
				LatestSourceUpdatedAt.UTC(),
		},
		GeneratedAt: result.GeneratedAt.UTC(),
	}
}

func toHistoricalIntelligenceScope(
	scope historicalcontract.Scope,
) HistoricalIntelligenceScope {
	return HistoricalIntelligenceScope{
		Type:            string(scope.Type),
		RegionCode:      scope.RegionCode,
		AirportICAOCode: scope.AirportICAOCode,
		OriginICAOCode:  scope.OriginICAOCode,
		DestinationICAOCode: scope.
			DestinationICAOCode,
	}
}

func toHistoricalIntelligenceTimeWindow(
	window historicalcontract.TimeWindow,
) HistoricalIntelligenceTimeWindow {
	return HistoricalIntelligenceTimeWindow{
		StartTime: window.StartTime.UTC(),
		EndTime:   window.EndTime.UTC(),
		AsOfTime:  window.AsOfTime.UTC(),
	}
}

func toHistoricalIntelligenceComparison(
	comparison *historicalcontract.PeriodComparison,
) *HistoricalIntelligencePeriodComparison {
	if comparison == nil {
		return nil
	}

	var percentageChange *float64
	if comparison.PercentageChange != nil {
		value := *comparison.PercentageChange
		percentageChange = &value
	}

	return &HistoricalIntelligencePeriodComparison{
		PreviousWindow: toHistoricalIntelligenceTimeWindow(
			comparison.PreviousWindow,
		),
		CurrentValue:     comparison.CurrentValue,
		PreviousValue:    comparison.PreviousValue,
		AbsoluteChange:   comparison.AbsoluteChange,
		PercentageChange: percentageChange,
		Direction: string(
			comparison.Direction,
		),
	}
}

func toHistoricalIntelligenceConfidence(
	confidence historicalcontract.Confidence,
) HistoricalIntelligenceConfidence {
	reasons := make(
		[]HistoricalIntelligenceConfidenceReason,
		0,
		len(confidence.Reasons),
	)
	for _, reason := range confidence.Reasons {
		reasons = append(
			reasons,
			HistoricalIntelligenceConfidenceReason{
				Code:         reason.Code,
				Message:      reason.Message,
				Contribution: reason.Contribution,
			},
		)
	}

	return HistoricalIntelligenceConfidence{
		Score:       confidence.Score,
		Level:       string(confidence.Level),
		SampleCount: confidence.SampleCount,
		Reasons:     reasons,
	}
}

func toHistoricalIntelligenceLimitations(
	limitations []historicalcontract.Limitation,
) []HistoricalIntelligenceLimitation {
	result := make(
		[]HistoricalIntelligenceLimitation,
		0,
		len(limitations),
	)
	for _, limitation := range limitations {
		result = append(
			result,
			HistoricalIntelligenceLimitation{
				Code:    limitation.Code,
				Message: limitation.Message,
				Scope:   limitation.Scope,
			},
		)
	}

	return result
}
