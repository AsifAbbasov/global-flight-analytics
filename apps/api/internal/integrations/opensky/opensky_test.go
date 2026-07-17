package opensky

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/sourceconstraints"
)

func TestParseStateVectorPreservesSourceCategoryAndNullableFields(t *testing.T) {
	raw := json.RawMessage(`[
		"abc123", " AHY101 ", "Azerbaijan", 1760000000, 1760000005,
		49.8671, 40.4093, 9753.6, false, 231.5, 92.0, 2.5,
		null, 9906.0, "7700", true, 2, 6
	]`)

	state, err := ParseStateVector(raw)
	if err != nil {
		t.Fatalf("parse state vector: %v", err)
	}
	if state.ICAO24 != "abc123" {
		t.Fatalf("ICAO24 = %q", state.ICAO24)
	}
	if state.Callsign == nil || *state.Callsign != "AHY101" {
		t.Fatalf("callsign = %#v", state.Callsign)
	}
	if state.PositionSource != PositionSourceMLAT {
		t.Fatalf("position source = %d", state.PositionSource)
	}
	if state.Category != AircraftCategoryHeavy {
		t.Fatalf("category = %d", state.Category)
	}
	if state.TimePosition == nil || state.TimePosition.Unix() != 1760000000 {
		t.Fatalf("time position = %#v", state.TimePosition)
	}
	if state.SensorSerials != nil {
		t.Fatalf("sensors = %#v, want nil", state.SensorSerials)
	}
}

func TestBoundingBoxBuildsRegionalQueryAndCreditCost(t *testing.T) {
	box := BoundingBox{
		MinimumLatitude:  38.0,
		MaximumLatitude:  42.0,
		MinimumLongitude: 44.0,
		MaximumLongitude: 51.0,
	}
	cost, err := box.EstimatedStateCreditCost()
	if err != nil {
		t.Fatalf("credit cost: %v", err)
	}
	if cost != 2 {
		t.Fatalf("credit cost = %d, want 2", cost)
	}
}

func TestAnonymousClientUsesRegionalBoundingBoxWithoutAuthorization(t *testing.T) {
	var requestURL string
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestURL = request.URL.String()
		authorization = request.Header.Get("Authorization")
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("X-Rate-Limit-Remaining", "399")
		_, _ = writer.Write([]byte(`{"time":1760000005,"states":[]}`))
	}))
	defer server.Close()

	config := DefaultConfig()
	config.BaseURL = server.URL
	config.HTTPClient = server.Client()
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	box := BoundingBox{38, 42, 44, 51}
	result, err := client.GetStates(context.Background(), StatesRequest{BoundingBox: &box})
	if err != nil {
		t.Fatalf("get states: %v", err)
	}
	for _, fragment := range []string{"lamin=38", "lamax=42", "lomin=44", "lomax=51"} {
		if !strings.Contains(requestURL, fragment) {
			t.Fatalf("request URL %q missing %q", requestURL, fragment)
		}
	}
	if authorization != "" {
		t.Fatalf("anonymous request sent authorization %q", authorization)
	}
	if result.RateLimit.Remaining == nil || *result.RateLimit.Remaining != 399 {
		t.Fatalf("rate limit = %#v", result.RateLimit.Remaining)
	}
}

func TestOAuth2ClientReusesToken(t *testing.T) {
	var tokenCalls int
	var stateCalls int
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/token":
			tokenCalls++
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"access_token":"token-1","expires_in":1800,"token_type":"Bearer"}`))
		case "/states/all":
			stateCalls++
			if request.Header.Get("Authorization") != "Bearer token-1" {
				t.Fatalf("authorization = %q", request.Header.Get("Authorization"))
			}
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"time":1760000005,"states":[]}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := DefaultConfig()
	config.BaseURL = server.URL
	config.TokenURL = server.URL + "/token"
	config.ClientID = "client"
	config.ClientSecret = "secret"
	config.HTTPClient = server.Client()
	config.PollingInterval = 5 * time.Second

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client.lastStatesRequest = time.Now().UTC().Add(-time.Hour)
	if _, err := client.GetStates(context.Background(), StatesRequest{}); err != nil {
		t.Fatalf("first get states: %v", err)
	}
	client.lastStatesRequest = time.Now().UTC().Add(-time.Hour)
	if _, err := client.GetStates(context.Background(), StatesRequest{}); err != nil {
		t.Fatalf("second get states: %v", err)
	}
	if tokenCalls != 1 {
		t.Fatalf("token calls = %d, want 1", tokenCalls)
	}
	if stateCalls != 2 {
		t.Fatalf("state calls = %d, want 2", stateCalls)
	}
}

func TestAnonymousHistoricalStateRequestIsRejected(t *testing.T) {
	config := DefaultConfig()
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, err = client.GetStates(context.Background(), StatesRequest{
		Time: time.Now().UTC().Add(-time.Minute),
	})
	if !errors.Is(err, ErrAnonymousHistoricalTime) {
		t.Fatalf("error = %v, want %v", err, ErrAnonymousHistoricalTime)
	}
}

func TestOpenSkyCapabilityGuardBlocksUnsupportedClaims(t *testing.T) {
	blocked := []sourceconstraints.Capability{
		sourceconstraints.CapabilityGlobalContinuousTracking,
		sourceconstraints.CapabilityOceanicContinuousTracking,
		sourceconstraints.CapabilityOwnReceiverObservation,
		sourceconstraints.CapabilityOfficialSchedule,
		sourceconstraints.CapabilityPilotIntent,
		sourceconstraints.CapabilityATCInstruction,
		sourceconstraints.CapabilityCertifiedSeparation,
		sourceconstraints.CapabilityCommercialFleetData,
	}
	for _, capability := range blocked {
		decision, err := EvaluateCapability(capability)
		if err != nil {
			t.Fatalf("evaluate %s: %v", capability, err)
		}
		if decision.Level != sourceconstraints.DecisionLevelBlocked {
			t.Fatalf("%s level = %s, want blocked", capability, decision.Level)
		}
	}
}

func TestTrackAndAirportDisclosuresRemainExplicit(t *testing.T) {
	if len(ExperimentalTrackDisclosure()) < 3 {
		t.Fatal("expected complete experimental track disclosure")
	}
	if len(EstimatedAirportDisclosure(FlightData{})) < 3 {
		t.Fatal("expected missing-airport disclosure")
	}
}

func TestStateResponsePropagatesRoundedSnapshotTime(t *testing.T) {
	response := StateResponse{
		Time: 1760000015,
		States: []json.RawMessage{json.RawMessage(`[
			"abc123", "AHY101", "Azerbaijan", 1760000000, 1760000005,
			49.8671, 40.4093, 9753.6, false, 231.5, 92.0, 2.5,
			null, 9906.0, "7700", false, 0, 6
		]`)},
	}

	states, err := response.ParseStates()
	if err != nil {
		t.Fatalf("parse states: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("states = %d, want 1", len(states))
	}
	if states[0].SnapshotTime.Unix() != response.Time {
		t.Fatalf("snapshot time = %d, want %d", states[0].SnapshotTime.Unix(), response.Time)
	}
}

func TestStateVectorValidityAllowsPositionAtProviderBoundary(t *testing.T) {
	snapshot := time.Date(2026, time.July, 18, 0, 0, 15, 0, time.UTC)
	positionTime := snapshot.Add(-MaximumProviderFieldAge)
	latitude := 40.4093
	longitude := 49.8671

	validity, err := EvaluateStateVectorValidity(StateVector{
		SnapshotTime: snapshot,
		TimePosition: &positionTime,
		LastContact:  snapshot.Add(-5 * time.Second),
		Latitude:     &latitude,
		Longitude:    &longitude,
	})
	if err != nil {
		t.Fatalf("evaluate validity: %v", err)
	}
	if validity.PositionValidity != PositionValidityProviderValid ||
		!validity.PositionUsable {
		t.Fatalf("validity = %#v", validity)
	}
	if validity.PositionAgeSeconds == nil ||
		*validity.PositionAgeSeconds != MaximumProviderFieldAge.Seconds() {
		t.Fatalf("position age = %#v", validity.PositionAgeSeconds)
	}
}

func TestStateVectorValidityBlocksPositionBeyondProviderWindow(t *testing.T) {
	snapshot := time.Date(2026, time.July, 18, 0, 0, 16, 0, time.UTC)
	positionTime := snapshot.Add(-MaximumProviderFieldAge - time.Second)
	latitude := 40.4093
	longitude := 49.8671

	validity, err := EvaluateStateVectorValidity(StateVector{
		SnapshotTime: snapshot,
		TimePosition: &positionTime,
		LastContact:  snapshot.Add(-5 * time.Second),
		Latitude:     &latitude,
		Longitude:    &longitude,
	})
	if err != nil {
		t.Fatalf("evaluate validity: %v", err)
	}
	if validity.PositionValidity != PositionValidityStale ||
		validity.PositionUsable {
		t.Fatalf("validity = %#v", validity)
	}
}

func TestStateVectorValidityDoesNotInventMissingPosition(t *testing.T) {
	snapshot := time.Date(2026, time.July, 18, 0, 0, 0, 0, time.UTC)
	validity, err := EvaluateStateVectorValidity(StateVector{
		SnapshotTime: snapshot,
		LastContact:  snapshot,
	})
	if err != nil {
		t.Fatalf("evaluate validity: %v", err)
	}
	if validity.PositionValidity != PositionValidityUnavailable ||
		validity.PositionUsable {
		t.Fatalf("validity = %#v", validity)
	}
}

func TestOfficialUsagePolicyPreservesAttributionAndDeploymentRisk(t *testing.T) {
	policy := OfficialUsagePolicy()
	if !policy.ResearchAndNonCommercialUseOnly ||
		!policy.AttributionRequired ||
		!policy.CommercialFlightDataUnavailable ||
		!policy.CloudAccessNotGuaranteed {
		t.Fatalf("policy = %#v", policy)
	}
	if strings.TrimSpace(policy.RequiredCitation) == "" ||
		strings.TrimSpace(policy.ProviderURL) == "" {
		t.Fatalf("policy attribution is incomplete: %#v", policy)
	}
	if len(policy.Limitations) < 3 {
		t.Fatalf("limitations = %v", policy.Limitations)
	}
}
