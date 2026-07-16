package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheruncertainty"
)

const WeatherContextResponseVersion = "weather-context-api-v1"

type WeatherContextNotice struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type WeatherContextScoreComponent struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Weight float64 `json:"weight"`
}

type WeatherContextConfidenceReason struct {
	Code         string  `json:"code"`
	Message      string  `json:"message"`
	Contribution float64 `json:"contribution"`
}

type WeatherContextConfidence struct {
	Score   float64                          `json:"score"`
	Level   string                           `json:"level"`
	Reasons []WeatherContextConfidenceReason `json:"reasons"`
}

type WeatherContextLimitation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Scope   string `json:"scope,omitempty"`
}

type WeatherContextPosition struct {
	Latitude          float64  `json:"latitude"`
	Longitude         float64  `json:"longitude"`
	AltitudeMeters    *float64 `json:"altitude_meters,omitempty"`
	VerticalReference string   `json:"vertical_reference"`
}

type WeatherContextSource struct {
	Provider                       string   `json:"provider"`
	Dataset                        string   `json:"dataset"`
	EvidenceKind                   string   `json:"evidence_kind"`
	HorizontalResolutionKilometers *float64 `json:"horizontal_resolution_kilometers,omitempty"`
	TemporalResolutionSeconds      int64    `json:"temporal_resolution_seconds"`
}

type WeatherContextFeatureVector struct {
	TemperatureCelsius       *float64 `json:"temperature_celsius,omitempty"`
	RelativeHumidityPercent  *float64 `json:"relative_humidity_percent,omitempty"`
	PrecipitationMillimeters *float64 `json:"precipitation_millimeters,omitempty"`
	RainMillimeters          *float64 `json:"rain_millimeters,omitempty"`
	CloudCoverPercent        *float64 `json:"cloud_cover_percent,omitempty"`
	SurfacePressureHPA       *float64 `json:"surface_pressure_hpa,omitempty"`
	WindSpeedMetersPerSecond *float64 `json:"wind_speed_meters_per_second,omitempty"`
	WindDirectionDegrees     *float64 `json:"wind_direction_degrees,omitempty"`
	WindGustsMetersPerSecond *float64 `json:"wind_gusts_meters_per_second,omitempty"`
	ConditionCode            *int     `json:"condition_code,omitempty"`
	ConditionCodeScheme      string   `json:"condition_code_scheme,omitempty"`
	PresentCount             int      `json:"present_count"`
}

type WeatherContextSample struct {
	Sequence    int                         `json:"sequence"`
	Position    WeatherContextPosition      `json:"position"`
	Source      WeatherContextSource        `json:"source"`
	Features    WeatherContextFeatureVector `json:"features"`
	ValidAt     time.Time                   `json:"valid_at"`
	AvailableAt time.Time                   `json:"available_at"`
	RetrievedAt time.Time                   `json:"retrieved_at"`
}

type WeatherFeatureContractResponse struct {
	SchemaVersion string    `json:"schema_version"`
	Status        string    `json:"status"`
	TrajectoryID  string    `json:"trajectory_id"`
	AsOfTime      time.Time `json:"as_of_time"`

	Samples []WeatherContextSample `json:"samples"`

	Confidence   WeatherContextConfidence   `json:"confidence"`
	Limitations  []WeatherContextLimitation `json:"limitations"`
	Explanations []WeatherContextNotice     `json:"explanations"`
	ScopeGuard   string                     `json:"scope_guard"`

	InputFingerprint  string    `json:"input_fingerprint"`
	SourceNames       []string  `json:"source_names"`
	LatestAvailableAt time.Time `json:"latest_available_at"`
	GeneratedAt       time.Time `json:"generated_at"`
}

type WeatherTrustResponse struct {
	Version  string    `json:"version"`
	Decision string    `json:"decision"`
	Usable   bool      `json:"usable"`
	AsOfTime time.Time `json:"as_of_time"`

	Score      float64                        `json:"score"`
	Components []WeatherContextScoreComponent `json:"components"`

	AllowedScopes []string               `json:"allowed_scopes"`
	Limitations   []WeatherContextNotice `json:"limitations"`
	Explanations  []WeatherContextNotice `json:"explanations"`

	InputFingerprint string `json:"input_fingerprint"`
}

type WeatherAlignmentComponent struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Weight float64 `json:"weight"`
}

type WeatherAlignmentMatch struct {
	TrajectoryPointSequence int       `json:"trajectory_point_sequence"`
	TrajectoryPointID       string    `json:"trajectory_point_id,omitempty"`
	TrajectoryObservedAt    time.Time `json:"trajectory_observed_at"`

	WeatherSampleSequence *int       `json:"weather_sample_sequence,omitempty"`
	WeatherValidAt        *time.Time `json:"weather_valid_at,omitempty"`

	Status string `json:"status"`

	AltitudeBasis  string   `json:"altitude_basis"`
	AltitudeMeters *float64 `json:"altitude_meters,omitempty"`

	HorizontalDistanceKilometers *float64                    `json:"horizontal_distance_kilometers,omitempty"`
	TemporalDistanceSeconds      *float64                    `json:"temporal_distance_seconds,omitempty"`
	VerticalDistanceMeters       *float64                    `json:"vertical_distance_meters,omitempty"`
	Score                        float64                     `json:"score"`
	Components                   []WeatherAlignmentComponent `json:"components"`
	Limitations                  []WeatherContextNotice      `json:"limitations"`
}

type WeatherAlignmentResponse struct {
	Version      string    `json:"version"`
	Status       string    `json:"status"`
	TrajectoryID string    `json:"trajectory_id"`
	AsOfTime     time.Time `json:"as_of_time"`

	TrustDecision string  `json:"trust_decision"`
	TrustScore    float64 `json:"trust_score"`

	PointCount     int     `json:"point_count"`
	AlignedCount   int     `json:"aligned_count"`
	UnmatchedCount int     `json:"unmatched_count"`
	CoverageRatio  float64 `json:"coverage_ratio"`

	Matches []WeatherAlignmentMatch `json:"matches"`

	Limitations  []WeatherContextNotice `json:"limitations"`
	Explanations []WeatherContextNotice `json:"explanations"`

	InputFingerprint string    `json:"input_fingerprint"`
	GeneratedAt      time.Time `json:"generated_at"`
}

type WeatherMetricSummary struct {
	PresentCount  int      `json:"present_count"`
	CoverageRatio float64  `json:"coverage_ratio"`
	Minimum       *float64 `json:"minimum,omitempty"`
	Maximum       *float64 `json:"maximum,omitempty"`
	Mean          *float64 `json:"mean,omitempty"`
}

type WeatherDirectionSummary struct {
	PresentCount         int      `json:"present_count"`
	CoverageRatio        float64  `json:"coverage_ratio"`
	MeanDirectionDegrees *float64 `json:"mean_direction_degrees,omitempty"`
	Concentration        *float64 `json:"concentration,omitempty"`
}

type WeatherConditionFrequency struct {
	Scheme string  `json:"scheme"`
	Code   int     `json:"code"`
	Count  int     `json:"count"`
	Share  float64 `json:"share"`
}

type WeatherEncounterPoint struct {
	TrajectoryPointSequence int       `json:"trajectory_point_sequence"`
	TrajectoryPointID       string    `json:"trajectory_point_id,omitempty"`
	TrajectoryObservedAt    time.Time `json:"trajectory_observed_at"`
	WeatherSampleSequence   int       `json:"weather_sample_sequence"`
	WeatherValidAt          time.Time `json:"weather_valid_at"`
	AlignmentScore          float64   `json:"alignment_score"`
	FeatureCount            int       `json:"feature_count"`
}

type WeatherEncounterResponse struct {
	Version      string    `json:"version"`
	Status       string    `json:"status"`
	TrajectoryID string    `json:"trajectory_id"`
	AsOfTime     time.Time `json:"as_of_time"`

	AlignmentStatus        string  `json:"alignment_status"`
	AlignmentCoverageRatio float64 `json:"alignment_coverage_ratio"`

	PointCount           int     `json:"point_count"`
	EncounterPointCount  int     `json:"encounter_point_count"`
	UnprofiledPointCount int     `json:"unprofiled_point_count"`
	ProfileCoverageRatio float64 `json:"profile_coverage_ratio"`

	EncounterStartedAt *time.Time `json:"encounter_started_at,omitempty"`
	EncounterEndedAt   *time.Time `json:"encounter_ended_at,omitempty"`

	TemperatureCelsius       WeatherMetricSummary    `json:"temperature_celsius"`
	RelativeHumidityPercent  WeatherMetricSummary    `json:"relative_humidity_percent"`
	PrecipitationMillimeters WeatherMetricSummary    `json:"precipitation_millimeters"`
	RainMillimeters          WeatherMetricSummary    `json:"rain_millimeters"`
	CloudCoverPercent        WeatherMetricSummary    `json:"cloud_cover_percent"`
	SurfacePressureHPA       WeatherMetricSummary    `json:"surface_pressure_hpa"`
	WindSpeedMetersPerSecond WeatherMetricSummary    `json:"wind_speed_meters_per_second"`
	WindDirectionDegrees     WeatherDirectionSummary `json:"wind_direction_degrees"`
	WindGustsMetersPerSecond WeatherMetricSummary    `json:"wind_gusts_meters_per_second"`

	Conditions        []WeatherConditionFrequency `json:"conditions"`
	DominantCondition *WeatherConditionFrequency  `json:"dominant_condition,omitempty"`
	Points            []WeatherEncounterPoint     `json:"points"`

	Limitations  []WeatherContextNotice `json:"limitations"`
	Explanations []WeatherContextNotice `json:"explanations"`

	InputFingerprint string    `json:"input_fingerprint"`
	GeneratedAt      time.Time `json:"generated_at"`
}

type WeatherUncertaintyPointAdjustment struct {
	Sequence        int       `json:"sequence"`
	ForecastTime    time.Time `json:"forecast_time"`
	HorizonProgress float64   `json:"horizon_progress"`
	Multiplier      float64   `json:"multiplier"`

	OriginalHorizontalRadiusM float64 `json:"original_horizontal_radius_m"`
	AdjustedHorizontalRadiusM float64 `json:"adjusted_horizontal_radius_m"`

	OriginalVerticalRadiusM *float64 `json:"original_vertical_radius_m,omitempty"`
	AdjustedVerticalRadiusM *float64 `json:"adjusted_vertical_radius_m,omitempty"`

	OriginalConfidenceScore float64 `json:"original_confidence_score"`
	AdjustedConfidenceScore float64 `json:"adjusted_confidence_score"`
}

type WeatherUncertaintyArrivalAdjustment struct {
	Multiplier float64 `json:"multiplier"`

	OriginalEarliestTime  time.Time `json:"original_earliest_time"`
	OriginalEstimatedTime time.Time `json:"original_estimated_time"`
	OriginalLatestTime    time.Time `json:"original_latest_time"`

	AdjustedEarliestTime  time.Time `json:"adjusted_earliest_time"`
	AdjustedEstimatedTime time.Time `json:"adjusted_estimated_time"`
	AdjustedLatestTime    time.Time `json:"adjusted_latest_time"`

	OriginalConfidenceScore float64 `json:"original_confidence_score"`
	AdjustedConfidenceScore float64 `json:"adjusted_confidence_score"`
}

type WeatherUncertaintyResponse struct {
	Version string `json:"version"`
	Status  string `json:"status"`

	TrajectoryID string    `json:"trajectory_id"`
	AsOfTime     time.Time `json:"as_of_time"`

	SeverityScore     float64                        `json:"severity_score"`
	WeatherMultiplier float64                        `json:"weather_multiplier"`
	Components        []WeatherContextScoreComponent `json:"components"`

	PointAdjustments  []WeatherUncertaintyPointAdjustment  `json:"point_adjustments"`
	ArrivalAdjustment *WeatherUncertaintyArrivalAdjustment `json:"arrival_adjustment,omitempty"`

	AdjustedProjection ProjectionIntelligenceProjection `json:"adjusted_projection"`

	Limitations  []WeatherContextNotice `json:"limitations"`
	Explanations []WeatherContextNotice `json:"explanations"`

	InputFingerprint string    `json:"input_fingerprint"`
	GeneratedAt      time.Time `json:"generated_at"`
}

type WeatherContextResponse struct {
	Version string `json:"version"`

	TrajectoryID string    `json:"trajectory_id"`
	AsOfTime     time.Time `json:"as_of_time"`

	Weather     WeatherFeatureContractResponse `json:"weather"`
	Trust       WeatherTrustResponse           `json:"trust"`
	Alignment   WeatherAlignmentResponse       `json:"alignment"`
	Encounter   WeatherEncounterResponse       `json:"encounter"`
	Uncertainty WeatherUncertaintyResponse     `json:"uncertainty"`

	InputFingerprint string    `json:"input_fingerprint"`
	GeneratedAt      time.Time `json:"generated_at"`
}

func ToWeatherContextResponse(
	weather weathercontract.Result,
	trust weathertrust.Result,
	alignment weatheralignment.Result,
	encounter weatherencounter.Result,
	uncertainty weatheruncertainty.Result,
	inputFingerprint string,
	generatedAt time.Time,
) WeatherContextResponse {
	return WeatherContextResponse{
		Version:          WeatherContextResponseVersion,
		TrajectoryID:     weather.TrajectoryID,
		AsOfTime:         weather.AsOfTime.UTC(),
		Weather:          toWeatherFeatureContractResponse(weather),
		Trust:            toWeatherTrustResponse(trust),
		Alignment:        toWeatherAlignmentResponse(alignment),
		Encounter:        toWeatherEncounterResponse(encounter),
		Uncertainty:      toWeatherUncertaintyResponse(uncertainty),
		InputFingerprint: inputFingerprint,
		GeneratedAt:      generatedAt.UTC(),
	}
}

func toWeatherFeatureContractResponse(
	result weathercontract.Result,
) WeatherFeatureContractResponse {
	samples := make([]WeatherContextSample, 0, len(result.Samples))
	for _, sample := range result.Samples {
		samples = append(samples, WeatherContextSample{
			Sequence: sample.Sequence,
			Position: WeatherContextPosition{
				Latitude:          sample.Position.Latitude,
				Longitude:         sample.Position.Longitude,
				AltitudeMeters:    cloneWeatherContextFloat(sample.Position.AltitudeMeters),
				VerticalReference: string(sample.Position.VerticalReference),
			},
			Source: WeatherContextSource{
				Provider:                       sample.Source.Provider,
				Dataset:                        sample.Source.Dataset,
				EvidenceKind:                   string(sample.Source.EvidenceKind),
				HorizontalResolutionKilometers: cloneWeatherContextFloat(sample.Source.HorizontalResolutionKilometers),
				TemporalResolutionSeconds:      int64(sample.Source.TemporalResolution / time.Second),
			},
			Features: WeatherContextFeatureVector{
				TemperatureCelsius:       cloneWeatherContextFloat(sample.Features.TemperatureCelsius),
				RelativeHumidityPercent:  cloneWeatherContextFloat(sample.Features.RelativeHumidityPercent),
				PrecipitationMillimeters: cloneWeatherContextFloat(sample.Features.PrecipitationMillimeters),
				RainMillimeters:          cloneWeatherContextFloat(sample.Features.RainMillimeters),
				CloudCoverPercent:        cloneWeatherContextFloat(sample.Features.CloudCoverPercent),
				SurfacePressureHPA:       cloneWeatherContextFloat(sample.Features.SurfacePressureHPA),
				WindSpeedMetersPerSecond: cloneWeatherContextFloat(sample.Features.WindSpeedMetersPerSecond),
				WindDirectionDegrees:     cloneWeatherContextFloat(sample.Features.WindDirectionDegrees),
				WindGustsMetersPerSecond: cloneWeatherContextFloat(sample.Features.WindGustsMetersPerSecond),
				ConditionCode:            cloneWeatherContextInt(sample.Features.ConditionCode),
				ConditionCodeScheme:      sample.Features.ConditionCodeScheme,
				PresentCount:             sample.Features.PresentCount(),
			},
			ValidAt:     sample.ValidAt.UTC(),
			AvailableAt: sample.AvailableAt.UTC(),
			RetrievedAt: sample.RetrievedAt.UTC(),
		})
	}

	reasons := make([]WeatherContextConfidenceReason, 0, len(result.Confidence.Reasons))
	for _, reason := range result.Confidence.Reasons {
		reasons = append(reasons, WeatherContextConfidenceReason{
			Code:         reason.Code,
			Message:      reason.Message,
			Contribution: reason.Contribution,
		})
	}

	return WeatherFeatureContractResponse{
		SchemaVersion: string(result.SchemaVersion),
		Status:        string(result.Status),
		TrajectoryID:  result.TrajectoryID,
		AsOfTime:      result.AsOfTime.UTC(),
		Samples:       samples,
		Confidence: WeatherContextConfidence{
			Score:   result.Confidence.Score,
			Level:   string(result.Confidence.Level),
			Reasons: reasons,
		},
		Limitations:       toWeatherContractLimitations(result.Limitations),
		Explanations:      toWeatherContractExplanations(result.Explanations),
		ScopeGuard:        string(result.ScopeGuard),
		InputFingerprint:  result.Provenance.InputFingerprint,
		SourceNames:       append([]string(nil), result.Provenance.SourceNames...),
		LatestAvailableAt: result.Provenance.LatestAvailableAt.UTC(),
		GeneratedAt:       result.GeneratedAt.UTC(),
	}
}

func toWeatherTrustResponse(result weathertrust.Result) WeatherTrustResponse {
	components := make([]WeatherContextScoreComponent, 0, len(result.Components))
	for _, component := range result.Components {
		components = append(components, WeatherContextScoreComponent{
			Name:   string(component.Name),
			Score:  component.Score,
			Weight: component.Weight,
		})
	}
	scopes := make([]string, 0, len(result.AllowedScopes))
	for _, scope := range result.AllowedScopes {
		scopes = append(scopes, string(scope))
	}

	return WeatherTrustResponse{
		Version:          result.Version,
		Decision:         string(result.Decision),
		Usable:           result.Usable,
		AsOfTime:         result.AsOfTime.UTC(),
		Score:            result.Score,
		Components:       components,
		AllowedScopes:    scopes,
		Limitations:      toWeatherTrustNotices(result.Limitations),
		Explanations:     toWeatherTrustNotices(result.Explanations),
		InputFingerprint: result.InputFingerprint,
	}
}

func toWeatherAlignmentResponse(
	result weatheralignment.Result,
) WeatherAlignmentResponse {
	matches := make([]WeatherAlignmentMatch, 0, len(result.Matches))
	for _, match := range result.Matches {
		components := make([]WeatherAlignmentComponent, 0, len(match.Components))
		for _, component := range match.Components {
			components = append(components, WeatherAlignmentComponent{
				Name:   string(component.Name),
				Score:  component.Score,
				Weight: component.Weight,
			})
		}
		matches = append(matches, WeatherAlignmentMatch{
			TrajectoryPointSequence:      match.TrajectoryPointSequence,
			TrajectoryPointID:            match.TrajectoryPointID,
			TrajectoryObservedAt:         match.TrajectoryObservedAt.UTC(),
			WeatherSampleSequence:        cloneWeatherContextInt(match.WeatherSampleSequence),
			WeatherValidAt:               cloneWeatherContextTime(match.WeatherValidAt),
			Status:                       string(match.Status),
			AltitudeBasis:                string(match.AltitudeBasis),
			AltitudeMeters:               cloneWeatherContextFloat(match.AltitudeMeters),
			HorizontalDistanceKilometers: cloneWeatherContextFloat(match.HorizontalDistanceKilometers),
			TemporalDistanceSeconds:      weatherContextDurationSeconds(match.TemporalDistance),
			VerticalDistanceMeters:       cloneWeatherContextFloat(match.VerticalDistanceMeters),
			Score:                        match.Score,
			Components:                   components,
			Limitations:                  toWeatherAlignmentNotices(match.Limitations),
		})
	}

	return WeatherAlignmentResponse{
		Version:          result.Version,
		Status:           string(result.Status),
		TrajectoryID:     result.TrajectoryID,
		AsOfTime:         result.AsOfTime.UTC(),
		TrustDecision:    string(result.TrustDecision),
		TrustScore:       result.TrustScore,
		PointCount:       result.PointCount,
		AlignedCount:     result.AlignedCount,
		UnmatchedCount:   result.UnmatchedCount,
		CoverageRatio:    result.CoverageRatio,
		Matches:          matches,
		Limitations:      toWeatherAlignmentNotices(result.Limitations),
		Explanations:     toWeatherAlignmentNotices(result.Explanations),
		InputFingerprint: result.InputFingerprint,
		GeneratedAt:      result.GeneratedAt.UTC(),
	}
}

func toWeatherEncounterResponse(
	result weatherencounter.Result,
) WeatherEncounterResponse {
	conditions := make([]WeatherConditionFrequency, 0, len(result.Conditions))
	for _, condition := range result.Conditions {
		conditions = append(conditions, WeatherConditionFrequency{
			Scheme: condition.Scheme,
			Code:   condition.Code,
			Count:  condition.Count,
			Share:  condition.Share,
		})
	}

	var dominant *WeatherConditionFrequency
	if result.DominantCondition != nil {
		dominant = &WeatherConditionFrequency{
			Scheme: result.DominantCondition.Scheme,
			Code:   result.DominantCondition.Code,
			Count:  result.DominantCondition.Count,
			Share:  result.DominantCondition.Share,
		}
	}

	points := make([]WeatherEncounterPoint, 0, len(result.Points))
	for _, point := range result.Points {
		points = append(points, WeatherEncounterPoint{
			TrajectoryPointSequence: point.TrajectoryPointSequence,
			TrajectoryPointID:       point.TrajectoryPointID,
			TrajectoryObservedAt:    point.TrajectoryObservedAt.UTC(),
			WeatherSampleSequence:   point.WeatherSampleSequence,
			WeatherValidAt:          point.WeatherValidAt.UTC(),
			AlignmentScore:          point.AlignmentScore,
			FeatureCount:            point.FeatureCount,
		})
	}

	return WeatherEncounterResponse{
		Version:                  result.Version,
		Status:                   string(result.Status),
		TrajectoryID:             result.TrajectoryID,
		AsOfTime:                 result.AsOfTime.UTC(),
		AlignmentStatus:          string(result.AlignmentStatus),
		AlignmentCoverageRatio:   result.AlignmentCoverageRatio,
		PointCount:               result.PointCount,
		EncounterPointCount:      result.EncounterPointCount,
		UnprofiledPointCount:     result.UnprofiledPointCount,
		ProfileCoverageRatio:     result.ProfileCoverageRatio,
		EncounterStartedAt:       cloneWeatherContextTime(result.EncounterStartedAt),
		EncounterEndedAt:         cloneWeatherContextTime(result.EncounterEndedAt),
		TemperatureCelsius:       toWeatherMetricSummary(result.TemperatureCelsius),
		RelativeHumidityPercent:  toWeatherMetricSummary(result.RelativeHumidityPercent),
		PrecipitationMillimeters: toWeatherMetricSummary(result.PrecipitationMillimeters),
		RainMillimeters:          toWeatherMetricSummary(result.RainMillimeters),
		CloudCoverPercent:        toWeatherMetricSummary(result.CloudCoverPercent),
		SurfacePressureHPA:       toWeatherMetricSummary(result.SurfacePressureHPA),
		WindSpeedMetersPerSecond: toWeatherMetricSummary(result.WindSpeedMetersPerSecond),
		WindDirectionDegrees: WeatherDirectionSummary{
			PresentCount:         result.WindDirectionDegrees.PresentCount,
			CoverageRatio:        result.WindDirectionDegrees.CoverageRatio,
			MeanDirectionDegrees: cloneWeatherContextFloat(result.WindDirectionDegrees.MeanDirectionDegrees),
			Concentration:        cloneWeatherContextFloat(result.WindDirectionDegrees.Concentration),
		},
		WindGustsMetersPerSecond: toWeatherMetricSummary(result.WindGustsMetersPerSecond),
		Conditions:               conditions,
		DominantCondition:        dominant,
		Points:                   points,
		Limitations:              toWeatherEncounterNotices(result.Limitations),
		Explanations:             toWeatherEncounterNotices(result.Explanations),
		InputFingerprint:         result.InputFingerprint,
		GeneratedAt:              result.GeneratedAt.UTC(),
	}
}

func toWeatherUncertaintyResponse(
	result weatheruncertainty.Result,
) WeatherUncertaintyResponse {
	components := make([]WeatherContextScoreComponent, 0, len(result.Components))
	for _, component := range result.Components {
		components = append(components, WeatherContextScoreComponent{
			Name:   string(component.Name),
			Score:  component.Score,
			Weight: component.Weight,
		})
	}

	adjustments := make([]WeatherUncertaintyPointAdjustment, 0, len(result.PointAdjustments))
	for _, adjustment := range result.PointAdjustments {
		adjustments = append(adjustments, WeatherUncertaintyPointAdjustment{
			Sequence:                  adjustment.Sequence,
			ForecastTime:              adjustment.ForecastTime.UTC(),
			HorizonProgress:           adjustment.HorizonProgress,
			Multiplier:                adjustment.Multiplier,
			OriginalHorizontalRadiusM: adjustment.OriginalHorizontalRadiusM,
			AdjustedHorizontalRadiusM: adjustment.AdjustedHorizontalRadiusM,
			OriginalVerticalRadiusM:   cloneWeatherContextFloat(adjustment.OriginalVerticalRadiusM),
			AdjustedVerticalRadiusM:   cloneWeatherContextFloat(adjustment.AdjustedVerticalRadiusM),
			OriginalConfidenceScore:   adjustment.OriginalConfidenceScore,
			AdjustedConfidenceScore:   adjustment.AdjustedConfidenceScore,
		})
	}

	var arrival *WeatherUncertaintyArrivalAdjustment
	if result.ArrivalAdjustment != nil {
		arrival = &WeatherUncertaintyArrivalAdjustment{
			Multiplier:              result.ArrivalAdjustment.Multiplier,
			OriginalEarliestTime:    result.ArrivalAdjustment.OriginalEarliestTime.UTC(),
			OriginalEstimatedTime:   result.ArrivalAdjustment.OriginalEstimatedTime.UTC(),
			OriginalLatestTime:      result.ArrivalAdjustment.OriginalLatestTime.UTC(),
			AdjustedEarliestTime:    result.ArrivalAdjustment.AdjustedEarliestTime.UTC(),
			AdjustedEstimatedTime:   result.ArrivalAdjustment.AdjustedEstimatedTime.UTC(),
			AdjustedLatestTime:      result.ArrivalAdjustment.AdjustedLatestTime.UTC(),
			OriginalConfidenceScore: result.ArrivalAdjustment.OriginalConfidenceScore,
			AdjustedConfidenceScore: result.ArrivalAdjustment.AdjustedConfidenceScore,
		}
	}

	return WeatherUncertaintyResponse{
		Version:            result.Version,
		Status:             string(result.Status),
		TrajectoryID:       result.TrajectoryID,
		AsOfTime:           result.AsOfTime.UTC(),
		SeverityScore:      result.SeverityScore,
		WeatherMultiplier:  result.WeatherMultiplier,
		Components:         components,
		PointAdjustments:   adjustments,
		ArrivalAdjustment:  arrival,
		AdjustedProjection: toProjectionIntelligenceProjection(result.AdjustedProjection),
		Limitations:        toWeatherUncertaintyNotices(result.Limitations),
		Explanations:       toWeatherUncertaintyNotices(result.Explanations),
		InputFingerprint:   result.InputFingerprint,
		GeneratedAt:        result.GeneratedAt.UTC(),
	}
}

func toWeatherMetricSummary(
	summary weatherencounter.MetricSummary,
) WeatherMetricSummary {
	return WeatherMetricSummary{
		PresentCount:  summary.PresentCount,
		CoverageRatio: summary.CoverageRatio,
		Minimum:       cloneWeatherContextFloat(summary.Minimum),
		Maximum:       cloneWeatherContextFloat(summary.Maximum),
		Mean:          cloneWeatherContextFloat(summary.Mean),
	}
}

func toWeatherContractLimitations(
	items []weathercontract.Limitation,
) []WeatherContextLimitation {
	result := make([]WeatherContextLimitation, 0, len(items))
	for _, item := range items {
		result = append(result, WeatherContextLimitation{
			Code:    item.Code,
			Message: item.Message,
			Scope:   item.Scope,
		})
	}
	return result
}

func toWeatherContractExplanations(
	items []weathercontract.Explanation,
) []WeatherContextNotice {
	result := make([]WeatherContextNotice, 0, len(items))
	for _, item := range items {
		result = append(result, WeatherContextNotice{
			Code:    item.Code,
			Message: item.Message,
		})
	}
	return result
}

func toWeatherTrustNotices(
	items []weathertrust.Notice,
) []WeatherContextNotice {
	result := make([]WeatherContextNotice, 0, len(items))
	for _, item := range items {
		result = append(result, WeatherContextNotice{
			Code:    item.Code,
			Message: item.Message,
		})
	}
	return result
}

func toWeatherAlignmentNotices(
	items []weatheralignment.Notice,
) []WeatherContextNotice {
	result := make([]WeatherContextNotice, 0, len(items))
	for _, item := range items {
		result = append(result, WeatherContextNotice{
			Code:    item.Code,
			Message: item.Message,
		})
	}
	return result
}

func toWeatherEncounterNotices(
	items []weatherencounter.Notice,
) []WeatherContextNotice {
	result := make([]WeatherContextNotice, 0, len(items))
	for _, item := range items {
		result = append(result, WeatherContextNotice{
			Code:    item.Code,
			Message: item.Message,
		})
	}
	return result
}

func toWeatherUncertaintyNotices(
	items []weatheruncertainty.Notice,
) []WeatherContextNotice {
	result := make([]WeatherContextNotice, 0, len(items))
	for _, item := range items {
		result = append(result, WeatherContextNotice{
			Code:    item.Code,
			Message: item.Message,
		})
	}
	return result
}

func cloneWeatherContextFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneWeatherContextInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneWeatherContextTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}

func weatherContextDurationSeconds(value *time.Duration) *float64 {
	if value == nil {
		return nil
	}
	seconds := value.Seconds()
	return &seconds
}
