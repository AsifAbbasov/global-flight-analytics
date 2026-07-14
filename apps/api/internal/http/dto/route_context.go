package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/routecontext"
)

type RouteContextNotice struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type RouteContextConfidence struct {
	Score   float64              `json:"score"`
	Level   string               `json:"level"`
	Reasons []RouteContextNotice `json:"reasons"`
}

type RouteContextAirport struct {
	ICAOCode    string  `json:"icao_code"`
	IATACode    string  `json:"iata_code"`
	Name        string  `json:"name"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ElevationM  float64 `json:"elevation_m"`
	Timezone    string  `json:"timezone"`
	Description string  `json:"description"`
}

type RouteContextAirportCandidate struct {
	Airport    RouteContextAirport    `json:"airport"`
	DistanceKM float64                `json:"distance_km"`
	Confidence RouteContextConfidence `json:"confidence"`
}

type AircraftRouteContext struct {
	ICAO24       string                        `json:"icao24"`
	TrajectoryID string                        `json:"trajectory_id"`
	Origin       *RouteContextAirportCandidate `json:"origin,omitempty"`
	Destination  *RouteContextAirportCandidate `json:"destination,omitempty"`
	Confidence   RouteContextConfidence        `json:"confidence"`
	Limitations  []RouteContextNotice          `json:"limitations"`
	GeneratedAt  time.Time                     `json:"generated_at"`
}

func ToAircraftRouteContext(
	item routecontext.Context,
) AircraftRouteContext {
	return AircraftRouteContext{
		ICAO24:       item.ICAO24,
		TrajectoryID: item.TrajectoryID,
		Origin:       toRouteContextAirportCandidate(item.Origin),
		Destination:  toRouteContextAirportCandidate(item.Destination),
		Confidence:   toRouteContextConfidence(item.Confidence),
		Limitations:  toRouteContextNotices(item.Limitations),
		GeneratedAt:  item.GeneratedAt,
	}
}

func toRouteContextAirportCandidate(
	item *routecontext.AirportCandidate,
) *RouteContextAirportCandidate {
	if item == nil {
		return nil
	}

	return &RouteContextAirportCandidate{
		Airport:    toRouteContextAirport(item.Airport),
		DistanceKM: item.DistanceKM,
		Confidence: toRouteContextConfidence(item.Confidence),
	}
}

func toRouteContextAirport(
	item airport.Airport,
) RouteContextAirport {
	return RouteContextAirport{
		ICAOCode:    item.ICAOCode,
		IATACode:    item.IATACode,
		Name:        item.Name,
		City:        item.City,
		Country:     item.Country,
		Latitude:    item.Latitude,
		Longitude:   item.Longitude,
		ElevationM:  item.ElevationM,
		Timezone:    item.Timezone,
		Description: item.Description,
	}
}

func toRouteContextConfidence(
	item routecontext.Confidence,
) RouteContextConfidence {
	return RouteContextConfidence{
		Score:   item.Score,
		Level:   string(item.Level),
		Reasons: toRouteContextNotices(item.Reasons),
	}
}

func toRouteContextNotices(
	items []routecontext.Notice,
) []RouteContextNotice {
	result := make(
		[]RouteContextNotice,
		0,
		len(items),
	)

	for _, item := range items {
		result = append(
			result,
			RouteContextNotice{
				Code:    item.Code,
				Message: item.Message,
			},
		)
	}

	return result
}
