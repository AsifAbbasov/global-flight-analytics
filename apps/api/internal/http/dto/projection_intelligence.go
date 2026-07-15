package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionfreshness"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
)

type ProjectionIntelligenceNotice struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ProjectionIntelligenceMethod struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	DecisionClass string `json:"decision_class"`
}

type ProjectionIntelligenceHorizon struct {
	AsOfTime        time.Time `json:"as_of_time"`
	EndTime         time.Time `json:"end_time"`
	StepSeconds     int64     `json:"step_seconds"`
	DurationSeconds int64     `json:"duration_seconds"`
}

type ProjectionIntelligencePosition struct {
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	AltitudeM *float64 `json:"altitude_m,omitempty"`
}

type ProjectionIntelligenceUncertainty struct {
	HorizontalRadiusM float64  `json:"horizontal_radius_m"`
	VerticalRadiusM   *float64 `json:"vertical_radius_m,omitempty"`
}

type ProjectionIntelligenceConfidenceReason struct {
	Code         string  `json:"code"`
	Message      string  `json:"message"`
	Contribution float64 `json:"contribution"`
}

type ProjectionIntelligenceConfidence struct {
	Score   float64                                  `json:"score"`
	Level   string                                   `json:"level"`
	Reasons []ProjectionIntelligenceConfidenceReason `json:"reasons"`
}

type ProjectionIntelligenceLimitation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Scope   string `json:"scope"`
}

type ProjectionIntelligenceExplanation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ProjectionIntelligencePoint struct {
	Sequence     int                               `json:"sequence"`
	ForecastTime time.Time                         `json:"forecast_time"`
	Position     ProjectionIntelligencePosition    `json:"position"`
	Uncertainty  ProjectionIntelligenceUncertainty `json:"uncertainty"`
	Confidence   ProjectionIntelligenceConfidence  `json:"confidence"`
}

type ProjectionIntelligenceArrivalEstimate struct {
	AirportICAOCode string `json:"airport_icao_code"`

	EarliestTime  time.Time `json:"earliest_time"`
	EstimatedTime time.Time `json:"estimated_time"`
	LatestTime    time.Time `json:"latest_time"`

	Confidence  ProjectionIntelligenceConfidence   `json:"confidence"`
	Limitations []ProjectionIntelligenceLimitation `json:"limitations"`
}

type ProjectionIntelligenceInputReference struct {
	Name           string    `json:"name"`
	Classification string    `json:"classification"`
	SourceName     string    `json:"source_name"`
	ObservedAt     time.Time `json:"observed_at"`
	RetrievedAt    time.Time `json:"retrieved_at"`
	Limitation     string    `json:"limitation,omitempty"`
}

type ProjectionIntelligenceProvenance struct {
	InputFingerprint      string                                 `json:"input_fingerprint"`
	Inputs                []ProjectionIntelligenceInputReference `json:"inputs"`
	LatestInputObservedAt time.Time                              `json:"latest_input_observed_at"`
}

type ProjectionIntelligenceProjection struct {
	SchemaVersion string `json:"schema_version"`
	Status        string `json:"status"`

	TrajectoryID string `json:"trajectory_id"`
	FlightID     string `json:"flight_id,omitempty"`
	AircraftID   string `json:"aircraft_id,omitempty"`
	ICAO24       string `json:"icao24,omitempty"`
	Callsign     string `json:"callsign,omitempty"`

	Method  ProjectionIntelligenceMethod           `json:"method"`
	Horizon ProjectionIntelligenceHorizon          `json:"horizon"`
	Points  []ProjectionIntelligencePoint          `json:"points"`
	Arrival *ProjectionIntelligenceArrivalEstimate `json:"arrival,omitempty"`

	Confidence   ProjectionIntelligenceConfidence    `json:"confidence"`
	Limitations  []ProjectionIntelligenceLimitation  `json:"limitations"`
	Explanations []ProjectionIntelligenceExplanation `json:"explanations"`
	ScopeGuard   string                              `json:"scope_guard"`
	Provenance   ProjectionIntelligenceProvenance    `json:"provenance"`
	GeneratedAt  time.Time                           `json:"generated_at"`
}

type ProjectionIntelligenceNeighbor struct {
	TrajectoryID string `json:"trajectory_id"`

	SimilarityScore float64 `json:"similarity_score"`
	SimilarityLevel string  `json:"similarity_level"`

	AnchorPointIndex int       `json:"anchor_point_index"`
	AnchorObservedAt time.Time `json:"anchor_observed_at"`
	AnchorDistanceKM float64   `json:"anchor_distance_km"`

	CandidateStartTime  time.Time `json:"candidate_start_time"`
	CandidateEndTime    time.Time `json:"candidate_end_time"`
	CandidateAgeSeconds int64     `json:"candidate_age_seconds"`

	PrefixPointCount       int       `json:"prefix_point_count"`
	ContinuationPointCount int       `json:"continuation_point_count"`
	ContinuationEndTime    time.Time `json:"continuation_end_time"`
}

type ProjectionIntelligenceNeighborSelection struct {
	Version string `json:"version"`
	Status  string `json:"status"`

	CurrentTrajectoryID         string    `json:"current_trajectory_id"`
	AsOfTime                    time.Time `json:"as_of_time"`
	RequiredContinuationSeconds int64     `json:"required_continuation_seconds"`

	InputCandidateCount     int `json:"input_candidate_count"`
	CheckedCandidateCount   int `json:"checked_candidate_count"`
	QualifiedCandidateCount int `json:"qualified_candidate_count"`
	RejectedCandidateCount  int `json:"rejected_candidate_count"`

	SelectionLimit int  `json:"selection_limit"`
	Truncated      bool `json:"truncated"`

	Neighbors   []ProjectionIntelligenceNeighbor `json:"neighbors"`
	Limitations []ProjectionIntelligenceNotice   `json:"limitations"`

	InputFingerprint string `json:"input_fingerprint"`
}

type ProjectionIntelligenceScoreComponent struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Weight float64 `json:"weight"`
}

type ProjectionIntelligencePatternConfidence struct {
	Version string `json:"version"`
	Status  string `json:"status"`
	Usable  bool   `json:"usable"`

	NeighborCount       int `json:"neighbor_count"`
	TargetNeighborCount int `json:"target_neighbor_count"`

	MeanSimilarityScore     float64 `json:"mean_similarity_score"`
	MeanCandidateAgeSeconds float64 `json:"mean_candidate_age_seconds"`
	MeanAnchorDistanceKM    float64 `json:"mean_anchor_distance_km"`

	Score float64 `json:"score"`
	Level string  `json:"level"`

	Components            []ProjectionIntelligenceScoreComponent `json:"components"`
	SelectedTrajectoryIDs []string                               `json:"selected_trajectory_ids"`
	Limitations           []ProjectionIntelligenceNotice         `json:"limitations"`

	InputFingerprint string `json:"input_fingerprint"`
}

type ProjectionIntelligenceFreshness struct {
	Version  string `json:"version"`
	Decision string `json:"decision"`
	Usable   bool   `json:"usable"`

	AsOfTime time.Time `json:"as_of_time"`

	NeighborCount       int `json:"neighbor_count"`
	RecentNeighborCount int `json:"recent_neighbor_count"`

	NewestNeighborAgeSeconds int64 `json:"newest_neighbor_age_seconds"`
	MeanNeighborAgeSeconds   int64 `json:"mean_neighbor_age_seconds"`
	OldestNeighborAgeSeconds int64 `json:"oldest_neighbor_age_seconds"`

	Score      float64                                `json:"score"`
	Components []ProjectionIntelligenceScoreComponent `json:"components"`

	SelectedTrajectoryIDs []string                       `json:"selected_trajectory_ids"`
	Limitations           []ProjectionIntelligenceNotice `json:"limitations"`
	InputFingerprint      string                         `json:"input_fingerprint"`
}

type ProjectionIntelligenceRouteFrequency struct {
	Version  string `json:"version"`
	Decision string `json:"decision"`
	Usable   bool   `json:"usable"`

	RouteKey string    `json:"route_key"`
	AsOfTime time.Time `json:"as_of_time"`

	ObservationCount            int     `json:"observation_count"`
	DistinctFlightCount         int     `json:"distinct_flight_count"`
	DistinctDayCount            int     `json:"distinct_day_count"`
	RecentObservationCount      int     `json:"recent_observation_count"`
	LatestObservationAgeSeconds int64   `json:"latest_observation_age_seconds"`
	RouteConfidenceScore        float64 `json:"route_confidence_score"`

	Score       float64                                `json:"score"`
	Components  []ProjectionIntelligenceScoreComponent `json:"components"`
	Limitations []ProjectionIntelligenceNotice         `json:"limitations"`

	HistoryInputFingerprint string `json:"history_input_fingerprint"`
	InputFingerprint        string `json:"input_fingerprint"`
}

type ProjectionIntelligenceEvidence struct {
	NeighborSelection *ProjectionIntelligenceNeighborSelection `json:"neighbor_selection,omitempty"`
	PatternConfidence *ProjectionIntelligencePatternConfidence `json:"pattern_confidence,omitempty"`
	Freshness         *ProjectionIntelligenceFreshness         `json:"freshness,omitempty"`
	RouteFrequency    *ProjectionIntelligenceRouteFrequency    `json:"route_frequency,omitempty"`
}

type ProjectionIntelligenceResponse struct {
	Version string `json:"version"`

	Strategy       string `json:"strategy"`
	FallbackReason string `json:"fallback_reason,omitempty"`
	ArrivalStatus  string `json:"arrival_status"`

	Projection ProjectionIntelligenceProjection `json:"projection"`
	Evidence   ProjectionIntelligenceEvidence   `json:"evidence"`
	Notices    []ProjectionIntelligenceNotice   `json:"notices"`

	InputFingerprint string    `json:"input_fingerprint"`
	GeneratedAt      time.Time `json:"generated_at"`
}

func ToProjectionIntelligenceResponse(
	result projectionproduction.Result,
) ProjectionIntelligenceResponse {
	return ProjectionIntelligenceResponse{
		Version:        result.Version,
		Strategy:       string(result.Strategy),
		FallbackReason: result.FallbackReason,
		ArrivalStatus:  string(result.ArrivalStatus),
		Projection:     toProjectionIntelligenceProjection(result.Projection),
		Evidence: ProjectionIntelligenceEvidence{
			NeighborSelection: toProjectionIntelligenceNeighborSelection(
				result.NeighborSelection,
			),
			PatternConfidence: toProjectionIntelligencePatternConfidence(
				result.PatternConfidence,
			),
			Freshness: toProjectionIntelligenceFreshness(
				result.Freshness,
			),
			RouteFrequency: toProjectionIntelligenceRouteFrequency(
				result.RouteFrequency,
			),
		},
		Notices:          toProjectionIntelligenceProductionNotices(result.Notices),
		InputFingerprint: result.InputFingerprint,
		GeneratedAt:      result.GeneratedAt.UTC(),
	}
}

func toProjectionIntelligenceProjection(
	result projectioncontract.Result,
) ProjectionIntelligenceProjection {
	points := make(
		[]ProjectionIntelligencePoint,
		0,
		len(result.Points),
	)
	for _, point := range result.Points {
		points = append(
			points,
			ProjectionIntelligencePoint{
				Sequence:     point.Sequence,
				ForecastTime: point.ForecastTime.UTC(),
				Position: ProjectionIntelligencePosition{
					Latitude:  point.Position.Latitude,
					Longitude: point.Position.Longitude,
					AltitudeM: cloneProjectionFloat(point.Position.AltitudeM),
				},
				Uncertainty: ProjectionIntelligenceUncertainty{
					HorizontalRadiusM: point.Uncertainty.HorizontalRadiusM,
					VerticalRadiusM: cloneProjectionFloat(
						point.Uncertainty.VerticalRadiusM,
					),
				},
				Confidence: toProjectionIntelligenceConfidence(
					point.Confidence,
				),
			},
		)
	}

	inputs := make(
		[]ProjectionIntelligenceInputReference,
		0,
		len(result.Provenance.Inputs),
	)
	for _, input := range result.Provenance.Inputs {
		inputs = append(
			inputs,
			ProjectionIntelligenceInputReference{
				Name:           input.Name,
				Classification: string(input.Classification),
				SourceName:     input.SourceName,
				ObservedAt:     input.ObservedAt.UTC(),
				RetrievedAt:    input.RetrievedAt.UTC(),
				Limitation:     input.Limitation,
			},
		)
	}

	return ProjectionIntelligenceProjection{
		SchemaVersion: string(result.SchemaVersion),
		Status:        string(result.Status),
		TrajectoryID:  result.TrajectoryID,
		FlightID:      result.FlightID,
		AircraftID:    result.AircraftID,
		ICAO24:        result.ICAO24,
		Callsign:      result.Callsign,
		Method: ProjectionIntelligenceMethod{
			Name:          result.Method.Name,
			Version:       result.Method.Version,
			DecisionClass: string(result.Method.DecisionClass),
		},
		Horizon: ProjectionIntelligenceHorizon{
			AsOfTime:        result.Horizon.AsOfTime.UTC(),
			EndTime:         result.Horizon.EndTime.UTC(),
			StepSeconds:     int64(result.Horizon.Step / time.Second),
			DurationSeconds: int64(result.Horizon.Duration() / time.Second),
		},
		Points:  points,
		Arrival: toProjectionIntelligenceArrival(result.Arrival),
		Confidence: toProjectionIntelligenceConfidence(
			result.Confidence,
		),
		Limitations: toProjectionIntelligenceLimitations(
			result.Limitations,
		),
		Explanations: toProjectionIntelligenceExplanations(
			result.Explanations,
		),
		ScopeGuard: string(result.ScopeGuard),
		Provenance: ProjectionIntelligenceProvenance{
			InputFingerprint:      result.Provenance.InputFingerprint,
			Inputs:                inputs,
			LatestInputObservedAt: result.Provenance.LatestInputObservedAt.UTC(),
		},
		GeneratedAt: result.GeneratedAt.UTC(),
	}
}

func toProjectionIntelligenceArrival(
	arrival *projectioncontract.ArrivalEstimate,
) *ProjectionIntelligenceArrivalEstimate {
	if arrival == nil {
		return nil
	}

	return &ProjectionIntelligenceArrivalEstimate{
		AirportICAOCode: arrival.AirportICAOCode,
		EarliestTime:    arrival.EarliestTime.UTC(),
		EstimatedTime:   arrival.EstimatedTime.UTC(),
		LatestTime:      arrival.LatestTime.UTC(),
		Confidence: toProjectionIntelligenceConfidence(
			arrival.Confidence,
		),
		Limitations: toProjectionIntelligenceLimitations(
			arrival.Limitations,
		),
	}
}

func toProjectionIntelligenceConfidence(
	confidence projectioncontract.Confidence,
) ProjectionIntelligenceConfidence {
	reasons := make(
		[]ProjectionIntelligenceConfidenceReason,
		0,
		len(confidence.Reasons),
	)
	for _, reason := range confidence.Reasons {
		reasons = append(
			reasons,
			ProjectionIntelligenceConfidenceReason{
				Code:         reason.Code,
				Message:      reason.Message,
				Contribution: reason.Contribution,
			},
		)
	}

	return ProjectionIntelligenceConfidence{
		Score:   confidence.Score,
		Level:   string(confidence.Level),
		Reasons: reasons,
	}
}

func toProjectionIntelligenceLimitations(
	items []projectioncontract.Limitation,
) []ProjectionIntelligenceLimitation {
	result := make(
		[]ProjectionIntelligenceLimitation,
		0,
		len(items),
	)
	for _, item := range items {
		result = append(
			result,
			ProjectionIntelligenceLimitation{
				Code:    item.Code,
				Message: item.Message,
				Scope:   item.Scope,
			},
		)
	}

	return result
}

func toProjectionIntelligenceExplanations(
	items []projectioncontract.Explanation,
) []ProjectionIntelligenceExplanation {
	result := make(
		[]ProjectionIntelligenceExplanation,
		0,
		len(items),
	)
	for _, item := range items {
		result = append(
			result,
			ProjectionIntelligenceExplanation{
				Code:    item.Code,
				Message: item.Message,
			},
		)
	}

	return result
}

func toProjectionIntelligenceNeighborSelection(
	selection *projectionneighbors.Result,
) *ProjectionIntelligenceNeighborSelection {
	if selection == nil {
		return nil
	}

	neighbors := make(
		[]ProjectionIntelligenceNeighbor,
		0,
		len(selection.Neighbors),
	)
	for _, neighbor := range selection.Neighbors {
		neighbors = append(
			neighbors,
			ProjectionIntelligenceNeighbor{
				TrajectoryID:       neighbor.TrajectoryID,
				SimilarityScore:    neighbor.SimilarityScore,
				SimilarityLevel:    string(neighbor.SimilarityLevel),
				AnchorPointIndex:   neighbor.AnchorPointIndex,
				AnchorObservedAt:   neighbor.AnchorObservedAt.UTC(),
				AnchorDistanceKM:   neighbor.AnchorDistanceKM,
				CandidateStartTime: neighbor.CandidateStartTime.UTC(),
				CandidateEndTime:   neighbor.CandidateEndTime.UTC(),
				CandidateAgeSeconds: int64(
					neighbor.CandidateAge / time.Second,
				),
				PrefixPointCount:       neighbor.PrefixPointCount,
				ContinuationPointCount: neighbor.ContinuationPointCount,
				ContinuationEndTime:    neighbor.ContinuationEndTime.UTC(),
			},
		)
	}

	return &ProjectionIntelligenceNeighborSelection{
		Version:             selection.Version,
		Status:              string(selection.Status),
		CurrentTrajectoryID: selection.CurrentTrajectoryID,
		AsOfTime:            selection.AsOfTime.UTC(),
		RequiredContinuationSeconds: int64(
			selection.RequiredContinuationDuration / time.Second,
		),
		InputCandidateCount:     selection.InputCandidateCount,
		CheckedCandidateCount:   selection.CheckedCandidateCount,
		QualifiedCandidateCount: selection.QualifiedCandidateCount,
		RejectedCandidateCount:  selection.RejectedCandidateCount,
		SelectionLimit:          selection.SelectionLimit,
		Truncated:               selection.Truncated,
		Neighbors:               neighbors,
		Limitations: toProjectionIntelligenceNeighborNotices(
			selection.Limitations,
		),
		InputFingerprint: selection.InputFingerprint,
	}
}

func toProjectionIntelligencePatternConfidence(
	pattern *projectionpatternconfidence.Result,
) *ProjectionIntelligencePatternConfidence {
	if pattern == nil {
		return nil
	}

	return &ProjectionIntelligencePatternConfidence{
		Version:                 pattern.Version,
		Status:                  string(pattern.Status),
		Usable:                  pattern.Usable,
		NeighborCount:           pattern.NeighborCount,
		TargetNeighborCount:     pattern.TargetNeighborCount,
		MeanSimilarityScore:     pattern.MeanSimilarityScore,
		MeanCandidateAgeSeconds: pattern.MeanCandidateAgeSeconds,
		MeanAnchorDistanceKM:    pattern.MeanAnchorDistanceKM,
		Score:                   pattern.Score,
		Level:                   string(pattern.Level),
		Components: toProjectionIntelligencePatternComponents(
			pattern.Components,
		),
		SelectedTrajectoryIDs: append(
			[]string(nil),
			pattern.SelectedTrajectoryIDs...,
		),
		Limitations: toProjectionIntelligencePatternNotices(
			pattern.Limitations,
		),
		InputFingerprint: pattern.InputFingerprint,
	}
}

func toProjectionIntelligenceFreshness(
	freshness *projectionfreshness.Result,
) *ProjectionIntelligenceFreshness {
	if freshness == nil {
		return nil
	}

	return &ProjectionIntelligenceFreshness{
		Version:             freshness.Version,
		Decision:            string(freshness.Decision),
		Usable:              freshness.Usable,
		AsOfTime:            freshness.AsOfTime.UTC(),
		NeighborCount:       freshness.NeighborCount,
		RecentNeighborCount: freshness.RecentNeighborCount,
		NewestNeighborAgeSeconds: int64(
			freshness.NewestNeighborAge / time.Second,
		),
		MeanNeighborAgeSeconds: int64(
			freshness.MeanNeighborAge / time.Second,
		),
		OldestNeighborAgeSeconds: int64(
			freshness.OldestNeighborAge / time.Second,
		),
		Score: freshness.Score,
		Components: toProjectionIntelligenceFreshnessComponents(
			freshness.Components,
		),
		SelectedTrajectoryIDs: append(
			[]string(nil),
			freshness.SelectedTrajectoryIDs...,
		),
		Limitations: toProjectionIntelligenceFreshnessNotices(
			freshness.Limitations,
		),
		InputFingerprint: freshness.InputFingerprint,
	}
}

func toProjectionIntelligenceRouteFrequency(
	frequency *projectionroutefrequency.Result,
) *ProjectionIntelligenceRouteFrequency {
	if frequency == nil {
		return nil
	}

	return &ProjectionIntelligenceRouteFrequency{
		Version:                frequency.Version,
		Decision:               string(frequency.Decision),
		Usable:                 frequency.Usable,
		RouteKey:               frequency.RouteKey,
		AsOfTime:               frequency.AsOfTime.UTC(),
		ObservationCount:       frequency.ObservationCount,
		DistinctFlightCount:    frequency.DistinctFlightCount,
		DistinctDayCount:       frequency.DistinctDayCount,
		RecentObservationCount: frequency.RecentObservationCount,
		LatestObservationAgeSeconds: int64(
			frequency.LatestObservationAge / time.Second,
		),
		RouteConfidenceScore: frequency.RouteConfidenceScore,
		Score:                frequency.Score,
		Components: toProjectionIntelligenceRouteFrequencyComponents(
			frequency.Components,
		),
		Limitations: toProjectionIntelligenceRouteFrequencyNotices(
			frequency.Limitations,
		),
		HistoryInputFingerprint: frequency.HistoryInputFingerprint,
		InputFingerprint:        frequency.InputFingerprint,
	}
}

func toProjectionIntelligencePatternComponents(
	items []projectionpatternconfidence.Component,
) []ProjectionIntelligenceScoreComponent {
	result := make(
		[]ProjectionIntelligenceScoreComponent,
		0,
		len(items),
	)
	for _, item := range items {
		result = append(
			result,
			ProjectionIntelligenceScoreComponent{
				Name:   string(item.Name),
				Score:  item.Score,
				Weight: item.Weight,
			},
		)
	}

	return result
}

func toProjectionIntelligenceFreshnessComponents(
	items []projectionfreshness.Component,
) []ProjectionIntelligenceScoreComponent {
	result := make(
		[]ProjectionIntelligenceScoreComponent,
		0,
		len(items),
	)
	for _, item := range items {
		result = append(
			result,
			ProjectionIntelligenceScoreComponent{
				Name:   string(item.Name),
				Score:  item.Score,
				Weight: item.Weight,
			},
		)
	}

	return result
}

func toProjectionIntelligenceRouteFrequencyComponents(
	items []projectionroutefrequency.Component,
) []ProjectionIntelligenceScoreComponent {
	result := make(
		[]ProjectionIntelligenceScoreComponent,
		0,
		len(items),
	)
	for _, item := range items {
		result = append(
			result,
			ProjectionIntelligenceScoreComponent{
				Name:   string(item.Name),
				Score:  item.Score,
				Weight: item.Weight,
			},
		)
	}

	return result
}

func toProjectionIntelligenceNeighborNotices(
	items []projectionneighbors.Notice,
) []ProjectionIntelligenceNotice {
	result := make(
		[]ProjectionIntelligenceNotice,
		0,
		len(items),
	)
	for _, item := range items {
		result = append(
			result,
			ProjectionIntelligenceNotice{
				Code:    item.Code,
				Message: item.Message,
			},
		)
	}

	return result
}

func toProjectionIntelligencePatternNotices(
	items []projectionpatternconfidence.Notice,
) []ProjectionIntelligenceNotice {
	result := make(
		[]ProjectionIntelligenceNotice,
		0,
		len(items),
	)
	for _, item := range items {
		result = append(
			result,
			ProjectionIntelligenceNotice{
				Code:    item.Code,
				Message: item.Message,
			},
		)
	}

	return result
}

func toProjectionIntelligenceFreshnessNotices(
	items []projectionfreshness.Notice,
) []ProjectionIntelligenceNotice {
	result := make(
		[]ProjectionIntelligenceNotice,
		0,
		len(items),
	)
	for _, item := range items {
		result = append(
			result,
			ProjectionIntelligenceNotice{
				Code:    item.Code,
				Message: item.Message,
			},
		)
	}

	return result
}

func toProjectionIntelligenceRouteFrequencyNotices(
	items []projectionroutefrequency.Notice,
) []ProjectionIntelligenceNotice {
	result := make(
		[]ProjectionIntelligenceNotice,
		0,
		len(items),
	)
	for _, item := range items {
		result = append(
			result,
			ProjectionIntelligenceNotice{
				Code:    item.Code,
				Message: item.Message,
			},
		)
	}

	return result
}

func toProjectionIntelligenceProductionNotices(
	items []projectionproduction.Notice,
) []ProjectionIntelligenceNotice {
	result := make(
		[]ProjectionIntelligenceNotice,
		0,
		len(items),
	)
	for _, item := range items {
		result = append(
			result,
			ProjectionIntelligenceNotice{
				Code:    item.Code,
				Message: item.Message,
			},
		)
	}

	return result
}

func cloneProjectionFloat(
	value *float64,
) *float64 {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
