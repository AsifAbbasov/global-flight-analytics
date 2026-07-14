package dto

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routestore"
	"testing"
	"time"
)

func TestToRouteIntelligenceRecord(t *testing.T) {
	now := time.Date(2026, time.July, 14, 18, 0, 0, 123, time.UTC)
	record := routestore.Record{
		ID: "route-record-test", InputFingerprint: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", StoredAt: now.Add(time.Second),
		Key:    routestore.ResultKey{TrajectoryID: "8a3d6e20-2c68-4b35-a512-7d91e6a90c31", SchemaVersion: routecontract.SchemaVersionV1, AsOfTime: now},
		Result: routecontract.Result{SchemaVersion: routecontract.SchemaVersionV1, Status: routecontract.RouteStatusPartial, TrajectoryID: "8a3d6e20-2c68-4b35-a512-7d91e6a90c31", ICAO24: "ABC123", Window: routecontract.RouteWindow{StartTime: now.Add(-time.Hour), EndTime: now, AsOfTime: now}, Origin: &routecontract.EndpointInference{Role: routecontract.EndpointRoleOrigin, Airport: routecontract.AirportReference{ICAOCode: "UBBB", Name: "Baku", Latitude: 40, Longitude: 50}, Confidence: routecontract.Confidence{Score: .9, Level: routecontract.ConfidenceLevelHigh}}, Confidence: routecontract.Confidence{Score: .45, Level: routecontract.ConfidenceLevelLow, EvidenceCount: 1}, Provenance: routecontract.Provenance{ResolverVersion: "route-resolver-v1", InputFingerprint: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", TrajectoryUpdatedAt: now, SourceNames: []string{"trajectory"}}, GeneratedAt: now.Add(time.Second)},
	}
	got := ToRouteIntelligenceRecord(record)
	if got.ID != record.ID || got.Result.Status != "partial" || got.Result.Origin == nil || got.Result.Origin.Airport.ICAOCode != "UBBB" {
		t.Fatalf("unexpected conversion: %#v", got)
	}
	got.Result.Provenance.SourceNames[0] = "changed"
	if record.Result.Provenance.SourceNames[0] != "trajectory" {
		t.Fatal("conversion shared mutable state")
	}
}

func TestToRouteIntelligenceHistoryCursor(t *testing.T) {
	now := time.Now().UTC()
	page := routestore.Page{HasMore: true, Records: []routestore.Record{{Key: routestore.ResultKey{AsOfTime: now}}}}
	got := ToRouteIntelligenceHistory(page)
	if !got.HasMore || got.NextBeforeAsOfTime == nil || !got.NextBeforeAsOfTime.Equal(now) {
		t.Fatalf("unexpected history: %#v", got)
	}
}
