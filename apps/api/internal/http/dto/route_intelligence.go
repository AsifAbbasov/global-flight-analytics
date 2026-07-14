package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routestore"
)

type RouteIntelligenceEvidenceAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RouteIntelligenceEvidence struct {
	Type          string                               `json:"type"`
	SourceName    string                               `json:"source_name"`
	SourceVersion string                               `json:"source_version"`
	Score         float64                              `json:"score"`
	Weight        float64                              `json:"weight"`
	ObservedAt    time.Time                            `json:"observed_at"`
	Summary       string                               `json:"summary"`
	Attributes    []RouteIntelligenceEvidenceAttribute `json:"attributes"`
}

type RouteIntelligenceConfidenceReason struct {
	Code         string  `json:"code"`
	Message      string  `json:"message"`
	Contribution float64 `json:"contribution"`
}

type RouteIntelligenceConfidence struct {
	Score         float64                             `json:"score"`
	Level         string                              `json:"level"`
	EvidenceCount int                                 `json:"evidence_count"`
	Reasons       []RouteIntelligenceConfidenceReason `json:"reasons"`
}

type RouteIntelligenceLimitation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Scope   string `json:"scope"`
}

type RouteIntelligenceAirport struct {
	ICAOCode   string  `json:"icao_code"`
	IATACode   string  `json:"iata_code"`
	Name       string  `json:"name"`
	City       string  `json:"city"`
	Country    string  `json:"country"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	ElevationM float64 `json:"elevation_m"`
	Timezone   string  `json:"timezone"`
}

type RouteIntelligenceEndpoint struct {
	Role        string                        `json:"role"`
	Airport     RouteIntelligenceAirport      `json:"airport"`
	DistanceKM  float64                       `json:"distance_km"`
	Confidence  RouteIntelligenceConfidence   `json:"confidence"`
	Evidence    []RouteIntelligenceEvidence   `json:"evidence"`
	Limitations []RouteIntelligenceLimitation `json:"limitations"`
}

type RouteIntelligenceWindow struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	AsOfTime  time.Time `json:"as_of_time"`
}

type RouteIntelligenceSummary struct {
	GreatCircleDistanceKM float64 `json:"great_circle_distance_km"`
	SameAirport           bool    `json:"same_airport"`
}

type RouteIntelligenceProvenance struct {
	ResolverVersion     string    `json:"resolver_version"`
	InputFingerprint    string    `json:"input_fingerprint"`
	TrajectoryUpdatedAt time.Time `json:"trajectory_updated_at"`
	SourceNames         []string  `json:"source_names"`
}

type RouteIntelligenceResult struct {
	SchemaVersion string                        `json:"schema_version"`
	Status        string                        `json:"status"`
	TrajectoryID  string                        `json:"trajectory_id"`
	IdentityKey   string                        `json:"identity_key"`
	FlightID      string                        `json:"flight_id"`
	AircraftID    string                        `json:"aircraft_id"`
	ICAO24        string                        `json:"icao24"`
	Callsign      string                        `json:"callsign"`
	Window        RouteIntelligenceWindow       `json:"window"`
	Origin        *RouteIntelligenceEndpoint    `json:"origin,omitempty"`
	Destination   *RouteIntelligenceEndpoint    `json:"destination,omitempty"`
	Summary       RouteIntelligenceSummary      `json:"summary"`
	Confidence    RouteIntelligenceConfidence   `json:"confidence"`
	Limitations   []RouteIntelligenceLimitation `json:"limitations"`
	Provenance    RouteIntelligenceProvenance   `json:"provenance"`
	GeneratedAt   time.Time                     `json:"generated_at"`
}

type RouteIntelligenceRecord struct {
	ID               string                  `json:"id"`
	InputFingerprint string                  `json:"input_fingerprint"`
	StoredAt         time.Time               `json:"stored_at"`
	Result           RouteIntelligenceResult `json:"result"`
}

type RouteIntelligenceHistory struct {
	Items              []RouteIntelligenceRecord `json:"items"`
	HasMore            bool                      `json:"has_more"`
	NextBeforeAsOfTime *time.Time                `json:"next_before_as_of_time,omitempty"`
}

func ToRouteIntelligenceRecord(record routestore.Record) RouteIntelligenceRecord {
	return RouteIntelligenceRecord{
		ID:               record.ID,
		InputFingerprint: record.InputFingerprint,
		StoredAt:         record.StoredAt,
		Result:           toRouteIntelligenceResult(record.Result),
	}
}

func ToRouteIntelligenceHistory(page routestore.Page) RouteIntelligenceHistory {
	items := make([]RouteIntelligenceRecord, 0, len(page.Records))
	for _, record := range page.Records {
		items = append(items, ToRouteIntelligenceRecord(record))
	}
	var nextBeforeAsOfTime *time.Time
	if page.HasMore && len(page.Records) > 0 {
		value := page.Records[len(page.Records)-1].Key.AsOfTime.UTC()
		nextBeforeAsOfTime = &value
	}
	return RouteIntelligenceHistory{Items: items, HasMore: page.HasMore, NextBeforeAsOfTime: nextBeforeAsOfTime}
}

func toRouteIntelligenceResult(item routecontract.Result) RouteIntelligenceResult {
	return RouteIntelligenceResult{
		SchemaVersion: string(item.SchemaVersion), Status: string(item.Status),
		TrajectoryID: item.TrajectoryID, IdentityKey: item.IdentityKey,
		FlightID: item.FlightID, AircraftID: item.AircraftID,
		ICAO24: item.ICAO24, Callsign: item.Callsign,
		Window: RouteIntelligenceWindow{StartTime: item.Window.StartTime, EndTime: item.Window.EndTime, AsOfTime: item.Window.AsOfTime},
		Origin: toRouteIntelligenceEndpoint(item.Origin), Destination: toRouteIntelligenceEndpoint(item.Destination),
		Summary:     RouteIntelligenceSummary{GreatCircleDistanceKM: item.Summary.GreatCircleDistanceKM, SameAirport: item.Summary.SameAirport},
		Confidence:  toRouteIntelligenceConfidence(item.Confidence),
		Limitations: toRouteIntelligenceLimitations(item.Limitations),
		Provenance: RouteIntelligenceProvenance{
			ResolverVersion:     item.Provenance.ResolverVersion,
			InputFingerprint:    item.Provenance.InputFingerprint,
			TrajectoryUpdatedAt: item.Provenance.TrajectoryUpdatedAt,
			SourceNames:         append([]string(nil), item.Provenance.SourceNames...),
		},
		GeneratedAt: item.GeneratedAt,
	}
}

func toRouteIntelligenceEndpoint(item *routecontract.EndpointInference) *RouteIntelligenceEndpoint {
	if item == nil {
		return nil
	}
	return &RouteIntelligenceEndpoint{
		Role: string(item.Role),
		Airport: RouteIntelligenceAirport{
			ICAOCode: item.Airport.ICAOCode, IATACode: item.Airport.IATACode,
			Name: item.Airport.Name, City: item.Airport.City, Country: item.Airport.Country,
			Latitude: item.Airport.Latitude, Longitude: item.Airport.Longitude,
			ElevationM: item.Airport.ElevationM, Timezone: item.Airport.Timezone,
		},
		DistanceKM:  item.DistanceKM,
		Confidence:  toRouteIntelligenceConfidence(item.Confidence),
		Evidence:    toRouteIntelligenceEvidence(item.Evidence),
		Limitations: toRouteIntelligenceLimitations(item.Limitations),
	}
}

func toRouteIntelligenceConfidence(item routecontract.Confidence) RouteIntelligenceConfidence {
	reasons := make([]RouteIntelligenceConfidenceReason, 0, len(item.Reasons))
	for _, reason := range item.Reasons {
		reasons = append(reasons, RouteIntelligenceConfidenceReason{Code: reason.Code, Message: reason.Message, Contribution: reason.Contribution})
	}
	return RouteIntelligenceConfidence{Score: item.Score, Level: string(item.Level), EvidenceCount: item.EvidenceCount, Reasons: reasons}
}

func toRouteIntelligenceEvidence(items []routecontract.Evidence) []RouteIntelligenceEvidence {
	result := make([]RouteIntelligenceEvidence, 0, len(items))
	for _, item := range items {
		attributes := make([]RouteIntelligenceEvidenceAttribute, 0, len(item.Attributes))
		for _, attribute := range item.Attributes {
			attributes = append(attributes, RouteIntelligenceEvidenceAttribute{Key: attribute.Key, Value: attribute.Value})
		}
		result = append(result, RouteIntelligenceEvidence{
			Type: string(item.Type), SourceName: item.SourceName, SourceVersion: item.SourceVersion,
			Score: item.Score, Weight: item.Weight, ObservedAt: item.ObservedAt,
			Summary: item.Summary, Attributes: attributes,
		})
	}
	return result
}

func toRouteIntelligenceLimitations(items []routecontract.Limitation) []RouteIntelligenceLimitation {
	result := make([]RouteIntelligenceLimitation, 0, len(items))
	for _, item := range items {
		result = append(result, RouteIntelligenceLimitation{Code: item.Code, Message: item.Message, Scope: item.Scope})
	}
	return result
}
