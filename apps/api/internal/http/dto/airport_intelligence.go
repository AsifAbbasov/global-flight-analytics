package dto

import (
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/airportproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/passport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/ranking"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/trends"
)

type AirportIntelligenceWindow struct {
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	AsOfTime      time.Time `json:"as_of_time"`
	CompletedDays int       `json:"completed_days"`
}
type AirportIntelligenceLimitation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type AirportPassportIdentity struct {
	ICAOCode string `json:"icao_code"`
	IATACode string `json:"iata_code"`
	Name     string `json:"name"`
}
type AirportPassportLocation struct {
	City            string                  `json:"city"`
	Country         string                  `json:"country"`
	Latitude        float64                 `json:"latitude"`
	Longitude       float64                 `json:"longitude"`
	ElevationM      *float64                `json:"elevation_m"`
	ElevationStatus airport.ElevationStatus `json:"elevation_status"`
	Timezone        string                  `json:"timezone"`
}
type AirportPassportOperations struct {
	Arrivals       int `json:"arrivals"`
	Departures     int `json:"departures"`
	Activity       int `json:"activity"`
	ActiveAircraft int `json:"active_aircraft"`
}
type AirportPassportDataQuality struct {
	FreshnessScore float64   `json:"freshness_score"`
	CoverageScore  float64   `json:"coverage_score"`
	ObservedAt     time.Time `json:"observed_at"`
}
type AirportPassportResponse struct {
	Identity    AirportPassportIdentity    `json:"identity"`
	Location    AirportPassportLocation    `json:"location"`
	Operations  AirportPassportOperations  `json:"operations"`
	DataQuality AirportPassportDataQuality `json:"data_quality"`
	Description string                     `json:"description"`
	GeneratedAt time.Time                  `json:"generated_at"`
}
type AirportStatisticsResponse struct {
	ICAOCode            string    `json:"icao_code"`
	WindowStart         time.Time `json:"window_start"`
	WindowEnd           time.Time `json:"window_end"`
	Arrivals            int       `json:"arrivals"`
	Departures          int       `json:"departures"`
	TotalMovements      int       `json:"total_movements"`
	ArrivalShare        float64   `json:"arrival_share"`
	DepartureShare      float64   `json:"departure_share"`
	MovementsPerHour    float64   `json:"movements_per_hour"`
	ActiveAircraft      int       `json:"active_aircraft"`
	ActiveRoutes        int       `json:"active_routes"`
	ObservedSamples     int       `json:"observed_samples"`
	ExpectedSamples     int       `json:"expected_samples"`
	CoverageScore       float64   `json:"coverage_score"`
	FreshnessScore      float64   `json:"freshness_score"`
	LatestObservationAt time.Time `json:"latest_observation_at"`
	GeneratedAt         time.Time `json:"generated_at"`
}
type AirportRankingSummaryResponse struct {
	Position              int     `json:"position"`
	TotalAirports         int     `json:"total_airports"`
	ActivityScore         float64 `json:"activity_score"`
	DataConfidence        float64 `json:"data_confidence"`
	MovementsComponent    float64 `json:"movements_component"`
	RoutesComponent       float64 `json:"routes_component"`
	ObservationsComponent float64 `json:"observations_component"`
	IntensityComponent    float64 `json:"intensity_component"`
}
type AirportIntelligenceOverviewResponse struct {
	Version     string                          `json:"version"`
	Window      AirportIntelligenceWindow       `json:"window"`
	Passport    AirportPassportResponse         `json:"passport"`
	Statistics  AirportStatisticsResponse       `json:"statistics"`
	Ranking     AirportRankingSummaryResponse   `json:"ranking"`
	Limitations []AirportIntelligenceLimitation `json:"limitations"`
	GeneratedAt time.Time                       `json:"generated_at"`
}
type AirportIntelligenceHistoryResponse struct {
	Version     string                          `json:"version"`
	Window      AirportIntelligenceWindow       `json:"window"`
	ICAOCode    string                          `json:"icao_code"`
	Entries     []AirportStatisticsResponse     `json:"entries"`
	Limitations []AirportIntelligenceLimitation `json:"limitations"`
	GeneratedAt time.Time                       `json:"generated_at"`
}
type AirportTrendPointResponse struct {
	WindowStart      time.Time `json:"window_start"`
	WindowEnd        time.Time `json:"window_end"`
	TotalMovements   int       `json:"total_movements"`
	MovementsPerHour float64   `json:"movements_per_hour"`
	ActiveRoutes     int       `json:"active_routes"`
	CoverageScore    float64   `json:"coverage_score"`
	FreshnessScore   float64   `json:"freshness_score"`
}
type AirportIntelligenceTrendsResponse struct {
	Version                            string                          `json:"version"`
	Window                             AirportIntelligenceWindow       `json:"window"`
	ICAOCode                           string                          `json:"icao_code"`
	ComparedWindows                    int                             `json:"compared_windows"`
	WindowDurationSeconds              int64                           `json:"window_duration_seconds"`
	Direction                          string                          `json:"direction"`
	Baseline                           AirportTrendPointResponse       `json:"baseline"`
	Current                            AirportTrendPointResponse       `json:"current"`
	Peak                               AirportTrendPointResponse       `json:"peak"`
	TotalMovementsChange               int                             `json:"total_movements_change"`
	MovementsPerHourChange             float64                         `json:"movements_per_hour_change"`
	MovementsPerHourChangePercent      float64                         `json:"movements_per_hour_change_percent"`
	MovementsPerHourChangePercentKnown bool                            `json:"movements_per_hour_change_percent_known"`
	ActiveRoutesChange                 int                             `json:"active_routes_change"`
	CoverageScoreChange                float64                         `json:"coverage_score_change"`
	FreshnessScoreChange               float64                         `json:"freshness_score_change"`
	GapCount                           int                             `json:"gap_count"`
	GapDurationSeconds                 int64                           `json:"gap_duration_seconds"`
	ObservedDurationSeconds            int64                           `json:"observed_duration_seconds"`
	ContinuityScore                    float64                         `json:"continuity_score"`
	Limitations                        []AirportIntelligenceLimitation `json:"limitations"`
	GeneratedAt                        time.Time                       `json:"generated_at"`
}
type AirportRankedItemResponse struct {
	Position              int     `json:"position"`
	ICAOCode              string  `json:"icao_code"`
	IATACode              string  `json:"iata_code"`
	Name                  string  `json:"name"`
	City                  string  `json:"city"`
	Country               string  `json:"country"`
	ActivityScore         float64 `json:"activity_score"`
	DataConfidence        float64 `json:"data_confidence"`
	MovementsComponent    float64 `json:"movements_component"`
	RoutesComponent       float64 `json:"routes_component"`
	ObservationsComponent float64 `json:"observations_component"`
	IntensityComponent    float64 `json:"intensity_component"`
	CoverageScore         float64 `json:"coverage_score"`
	FreshnessScore        float64 `json:"freshness_score"`
	TotalMovements        int     `json:"total_movements"`
	ActiveRoutes          int     `json:"active_routes"`
	ObservedSamples       int     `json:"observed_samples"`
	ExpectedSamples       int     `json:"expected_samples"`
	MovementsPerHour      float64 `json:"movements_per_hour"`
	ActiveAircraft        int     `json:"active_aircraft"`
}
type AirportRankingWeightsResponse struct {
	Movements    float64 `json:"movements"`
	Routes       float64 `json:"routes"`
	Observations float64 `json:"observations"`
	Intensity    float64 `json:"intensity"`
	Coverage     float64 `json:"coverage"`
	Freshness    float64 `json:"freshness"`
}
type AirportIntelligenceRankingResponse struct {
	Version     string                          `json:"version"`
	Window      AirportIntelligenceWindow       `json:"window"`
	Weights     AirportRankingWeightsResponse   `json:"weights"`
	Airports    []AirportRankedItemResponse     `json:"airports"`
	Limitations []AirportIntelligenceLimitation `json:"limitations"`
	GeneratedAt time.Time                       `json:"generated_at"`
}

func ToAirportIntelligenceOverviewResponse(result airportproduction.OverviewResult) AirportIntelligenceOverviewResponse {
	value := result.Overview
	return AirportIntelligenceOverviewResponse{Version: result.Version, Window: toAirportIntelligenceWindow(result.Window), Passport: toAirportPassportResponse(value.Passport), Statistics: toAirportStatisticsResponse(value.Statistics), Ranking: AirportRankingSummaryResponse{Position: value.Ranking.Position, TotalAirports: value.Ranking.TotalAirports, ActivityScore: value.Ranking.ActivityScore, DataConfidence: value.Ranking.DataConfidence, MovementsComponent: value.Ranking.MovementsComponent, RoutesComponent: value.Ranking.RoutesComponent, ObservationsComponent: value.Ranking.ObservationsComponent, IntensityComponent: value.Ranking.IntensityComponent}, Limitations: toAirportIntelligenceLimitations(result.Limitations), GeneratedAt: result.GeneratedAt}
}
func ToAirportIntelligenceHistoryResponse(result airportproduction.HistoryResult) AirportIntelligenceHistoryResponse {
	entries := make([]AirportStatisticsResponse, 0, len(result.History.Entries))
	for _, entry := range result.History.Entries {
		entries = append(entries, toAirportStatisticsResponse(entry))
	}
	return AirportIntelligenceHistoryResponse{Version: result.Version, Window: toAirportIntelligenceWindow(result.Window), ICAOCode: result.History.ICAOCode, Entries: entries, Limitations: toAirportIntelligenceLimitations(result.Limitations), GeneratedAt: result.GeneratedAt}
}
func ToAirportIntelligenceTrendsResponse(result airportproduction.TrendsResult) AirportIntelligenceTrendsResponse {
	value := result.Trends
	return AirportIntelligenceTrendsResponse{Version: result.Version, Window: toAirportIntelligenceWindow(result.Window), ICAOCode: value.ICAOCode, ComparedWindows: value.ComparedWindows, WindowDurationSeconds: int64(value.WindowDuration / time.Second), Direction: string(value.Direction), Baseline: toAirportTrendPointResponse(value.Baseline), Current: toAirportTrendPointResponse(value.Current), Peak: toAirportTrendPointResponse(value.Peak), TotalMovementsChange: value.TotalMovementsChange, MovementsPerHourChange: value.MovementsPerHourChange, MovementsPerHourChangePercent: value.MovementsPerHourChangePercent, MovementsPerHourChangePercentKnown: value.MovementsPerHourChangePercentKnown, ActiveRoutesChange: value.ActiveRoutesChange, CoverageScoreChange: value.CoverageScoreChange, FreshnessScoreChange: value.FreshnessScoreChange, GapCount: value.GapCount, GapDurationSeconds: int64(value.GapDuration / time.Second), ObservedDurationSeconds: int64(value.ObservedDuration / time.Second), ContinuityScore: value.ContinuityScore, Limitations: toAirportIntelligenceLimitations(result.Limitations), GeneratedAt: result.GeneratedAt}
}
func ToAirportIntelligenceRankingResponse(result airportproduction.RankingResult, limit int) AirportIntelligenceRankingResponse {
	ranked := result.Ranking.Airports
	if limit > 0 && len(ranked) > limit {
		ranked = ranked[:limit]
	}
	items := make([]AirportRankedItemResponse, 0, len(ranked))
	for _, item := range ranked {
		airportValue := result.Airports[strings.ToUpper(strings.TrimSpace(item.ICAOCode))]
		items = append(items, toAirportRankedItemResponse(item, airportValue.IATACode, airportValue.Name, airportValue.City, airportValue.Country))
	}
	return AirportIntelligenceRankingResponse{Version: result.Version, Window: toAirportIntelligenceWindow(result.Window), Weights: AirportRankingWeightsResponse{Movements: result.Ranking.ActivityWeights.Movements, Routes: result.Ranking.ActivityWeights.Routes, Observations: result.Ranking.ActivityWeights.Observations, Intensity: result.Ranking.ActivityWeights.Intensity, Coverage: result.Ranking.ConfidenceWeights.Coverage, Freshness: result.Ranking.ConfidenceWeights.Freshness}, Airports: items, Limitations: toAirportIntelligenceLimitations(result.Limitations), GeneratedAt: result.GeneratedAt}
}
func toAirportIntelligenceWindow(value airportproduction.Window) AirportIntelligenceWindow {
	return AirportIntelligenceWindow{StartTime: value.StartTime, EndTime: value.EndTime, AsOfTime: value.AsOfTime, CompletedDays: value.CompletedDays}
}
func toAirportIntelligenceLimitations(values []airportproduction.Limitation) []AirportIntelligenceLimitation {
	result := make([]AirportIntelligenceLimitation, 0, len(values))
	for _, value := range values {
		result = append(result, AirportIntelligenceLimitation{Code: value.Code, Message: value.Message})
	}
	return result
}
func toAirportPassportResponse(value passport.Passport) AirportPassportResponse {
	elevationM, elevationStatus := ToAirportElevation(
		value.Location.ElevationM,
		value.Location.ElevationAvailable,
	)
	return AirportPassportResponse{Identity: AirportPassportIdentity{ICAOCode: value.Identity.ICAOCode, IATACode: value.Identity.IATACode, Name: value.Identity.Name}, Location: AirportPassportLocation{City: value.Location.City, Country: value.Location.Country, Latitude: value.Location.Latitude, Longitude: value.Location.Longitude, ElevationM: elevationM, ElevationStatus: elevationStatus, Timezone: value.Location.Timezone}, Operations: AirportPassportOperations{Arrivals: value.Operations.Arrivals, Departures: value.Operations.Departures, Activity: value.Operations.Activity, ActiveAircraft: value.Operations.ActiveAircraft}, DataQuality: AirportPassportDataQuality{FreshnessScore: value.DataQuality.FreshnessScore, CoverageScore: value.DataQuality.CoverageScore, ObservedAt: value.DataQuality.ObservedAt}, Description: value.Description, GeneratedAt: value.GeneratedAt}
}
func toAirportStatisticsResponse(value statistics.Statistics) AirportStatisticsResponse {
	return AirportStatisticsResponse{ICAOCode: value.ICAOCode, WindowStart: value.WindowStart, WindowEnd: value.WindowEnd, Arrivals: value.Arrivals, Departures: value.Departures, TotalMovements: value.TotalMovements, ArrivalShare: value.ArrivalShare, DepartureShare: value.DepartureShare, MovementsPerHour: value.MovementsPerHour, ActiveAircraft: value.ActiveAircraft, ActiveRoutes: value.ActiveRoutes, ObservedSamples: value.ObservedSamples, ExpectedSamples: value.ExpectedSamples, CoverageScore: value.CoverageScore, FreshnessScore: value.FreshnessScore, LatestObservationAt: value.LatestObservationAt, GeneratedAt: value.GeneratedAt}
}
func toAirportTrendPointResponse(value trends.Point) AirportTrendPointResponse {
	return AirportTrendPointResponse{WindowStart: value.WindowStart, WindowEnd: value.WindowEnd, TotalMovements: value.TotalMovements, MovementsPerHour: value.MovementsPerHour, ActiveRoutes: value.ActiveRoutes, CoverageScore: value.CoverageScore, FreshnessScore: value.FreshnessScore}
}
func toAirportRankedItemResponse(value ranking.AirportRank, iataCode, name, city, country string) AirportRankedItemResponse {
	return AirportRankedItemResponse{Position: value.Position, ICAOCode: value.ICAOCode, IATACode: iataCode, Name: name, City: city, Country: country, ActivityScore: value.ActivityScore, DataConfidence: value.DataConfidence, MovementsComponent: value.MovementsComponent, RoutesComponent: value.RoutesComponent, ObservationsComponent: value.ObservationsComponent, IntensityComponent: value.IntensityComponent, CoverageScore: value.CoverageScore, FreshnessScore: value.FreshnessScore, TotalMovements: value.TotalMovements, ActiveRoutes: value.ActiveRoutes, ObservedSamples: value.ObservedSamples, ExpectedSamples: value.ExpectedSamples, MovementsPerHour: value.MovementsPerHour, ActiveAircraft: value.ActiveAircraft}
}
