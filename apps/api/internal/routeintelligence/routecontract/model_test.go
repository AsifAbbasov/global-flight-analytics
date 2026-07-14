package routecontract

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestResultCloneDoesNotShareMutableState(
	t *testing.T,
) {
	original := validCompleteResult()
	cloned := original.Clone()

	cloned.Origin.Airport.Name = "Changed"
	cloned.Origin.Evidence[0].Attributes[0].Value =
		"changed"
	cloned.Origin.Confidence.Reasons[0].Message =
		"changed"
	cloned.Origin.Limitations[0].Message =
		"changed"
	cloned.Destination.Evidence[0].Summary =
		"changed"
	cloned.Confidence.Reasons[0].Code =
		"changed"
	cloned.Limitations[0].Code =
		"changed"
	cloned.Provenance.SourceNames[0] =
		"changed"

	if original.Origin.Airport.Name == "Changed" ||
		original.Origin.Evidence[0].Attributes[0].Value ==
			"changed" ||
		original.Origin.Confidence.Reasons[0].Message ==
			"changed" ||
		original.Origin.Limitations[0].Message ==
			"changed" ||
		original.Destination.Evidence[0].Summary ==
			"changed" ||
		original.Confidence.Reasons[0].Code ==
			"changed" ||
		original.Limitations[0].Code ==
			"changed" ||
		original.Provenance.SourceNames[0] ==
			"changed" {
		t.Fatal(
			"Result.Clone() shared mutable state",
		)
	}
}

func TestResultClonePreservesNilEndpoints(
	t *testing.T,
) {
	original := validUnavailableResult()
	cloned := original.Clone()

	if cloned.Origin != nil ||
		cloned.Destination != nil {
		t.Fatalf(
			"Clone() endpoints = %#v %#v",
			cloned.Origin,
			cloned.Destination,
		)
	}
}

func TestConfidenceLevelForScore(t *testing.T) {
	tests := []struct {
		name  string
		score float64
		want  ConfidenceLevel
	}{
		{
			name:  "none",
			score: 0,
			want:  ConfidenceLevelNone,
		},
		{
			name:  "low",
			score: 0.01,
			want:  ConfidenceLevelLow,
		},
		{
			name:  "medium boundary",
			score: 0.6,
			want:  ConfidenceLevelMedium,
		},
		{
			name:  "high boundary",
			score: 0.8,
			want:  ConfidenceLevelHigh,
		},
		{
			name:  "high",
			score: 1,
			want:  ConfidenceLevelHigh,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ConfidenceLevelForScore(
				test.score,
			); got != test.want {
				t.Fatalf(
					"ConfidenceLevelForScore(%v) = %q, want %q",
					test.score,
					got,
					test.want,
				)
			}
		})
	}
}

func TestVersionConstantsRemainStable(
	t *testing.T,
) {
	if Version != "route-intelligence-contract-v1" {
		t.Fatalf("Version = %q", Version)
	}
	if SchemaVersionV1 !=
		"route-intelligence-v1" {
		t.Fatalf(
			"SchemaVersionV1 = %q",
			SchemaVersionV1,
		)
	}
	if ValidationVersion !=
		"route-intelligence-contract-validation-v1" {
		t.Fatalf(
			"ValidationVersion = %q",
			ValidationVersion,
		)
	}
}

func validCompleteResult() Result {
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		15,
		0,
		0,
		123456789,
		time.UTC,
	)
	originEvidence := validEvidence(
		EvidenceTypeTrajectoryEndpointProximity,
		asOfTime.Add(-50*time.Minute),
		"origin",
	)
	destinationEvidence := validEvidence(
		EvidenceTypeTrajectoryEndpointProximity,
		asOfTime.Add(-time.Minute),
		"destination",
	)

	return Result{
		SchemaVersion: SchemaVersionV1,
		Status:        RouteStatusComplete,
		TrajectoryID:  "8a3d6e20-2c68-4b35-a512-7d91e6a90c31",
		IdentityKey: "flight-identity-" +
			strings.Repeat("a", 64),
		FlightID:   "flight-one",
		AircraftID: "aircraft-one",
		ICAO24:     "ABC123",
		Callsign:   "J2001",
		Window: RouteWindow{
			StartTime: asOfTime.Add(-time.Hour),
			EndTime:   asOfTime.Add(-time.Minute),
			AsOfTime:  asOfTime,
		},
		Origin: &EndpointInference{
			Role: EndpointRoleOrigin,
			Airport: AirportReference{
				ICAOCode:   "UBBB",
				IATACode:   "GYD",
				Name:       "Heydar Aliyev International Airport",
				City:       "Baku",
				Country:    "Azerbaijan",
				Latitude:   40.4675,
				Longitude:  50.0467,
				ElevationM: 3,
				Timezone:   "Asia/Baku",
			},
			DistanceKM: 2.5,
			Confidence: Confidence{
				Score:         0.9,
				Level:         ConfidenceLevelHigh,
				EvidenceCount: 1,
				Reasons: []ConfidenceReason{
					{
						Code:         "origin_endpoint_proximity",
						Message:      "Trajectory start is close to the airport.",
						Contribution: 0.9,
					},
				},
			},
			Evidence: []Evidence{
				originEvidence,
			},
			Limitations: []Limitation{
				{
					Code:    "origin_probable_only",
					Message: "Origin is inferred and is not filed flight-plan data.",
					Scope:   "origin",
				},
			},
		},
		Destination: &EndpointInference{
			Role: EndpointRoleDestination,
			Airport: AirportReference{
				ICAOCode:   "UGTB",
				IATACode:   "TBS",
				Name:       "Tbilisi International Airport",
				City:       "Tbilisi",
				Country:    "Georgia",
				Latitude:   41.6692,
				Longitude:  44.9547,
				ElevationM: 495,
				Timezone:   "Asia/Tbilisi",
			},
			DistanceKM: 3.1,
			Confidence: Confidence{
				Score:         0.85,
				Level:         ConfidenceLevelHigh,
				EvidenceCount: 1,
				Reasons: []ConfidenceReason{
					{
						Code:         "destination_endpoint_proximity",
						Message:      "Trajectory end is close to the airport.",
						Contribution: 0.85,
					},
				},
			},
			Evidence: []Evidence{
				destinationEvidence,
			},
			Limitations: []Limitation{
				{
					Code:    "destination_probable_only",
					Message: "Destination is inferred and may not be the planned destination.",
					Scope:   "destination",
				},
			},
		},
		Summary: RouteSummary{
			GreatCircleDistanceKM: 448.5,
			SameAirport:           false,
		},
		Confidence: Confidence{
			Score:         0.85,
			Level:         ConfidenceLevelHigh,
			EvidenceCount: 2,
			Reasons: []ConfidenceReason{
				{
					Code:         "both_endpoints_supported",
					Message:      "Both endpoints have supporting trajectory evidence.",
					Contribution: 0.85,
				},
			},
		},
		Limitations: []Limitation{
			{
				Code:    "probable_route_only",
				Message: "Route endpoints are inferred and are not filed flight-plan data.",
				Scope:   "route",
			},
		},
		Provenance: Provenance{
			ResolverVersion: "route-resolver-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("b", 64),
			TrajectoryUpdatedAt: asOfTime.Add(-time.Minute),
			SourceNames: []string{
				"ourairports",
				"trajectory",
			},
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}
}

func validUnavailableResult() Result {
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		15,
		0,
		0,
		0,
		time.UTC,
	)

	return Result{
		SchemaVersion: SchemaVersionV1,
		Status:        RouteStatusUnavailable,
		TrajectoryID:  "8a3d6e20-2c68-4b35-a512-7d91e6a90c31",
		ICAO24:        "ABC123",
		Window: RouteWindow{
			StartTime: asOfTime.Add(-time.Hour),
			EndTime:   asOfTime.Add(-time.Minute),
			AsOfTime:  asOfTime,
		},
		Confidence: Confidence{
			Score:         0,
			Level:         ConfidenceLevelNone,
			EvidenceCount: 0,
			Reasons: []ConfidenceReason{
				{
					Code:         "no_endpoint_evidence",
					Message:      "No route endpoint evidence is available.",
					Contribution: 0,
				},
			},
		},
		Limitations: []Limitation{
			{
				Code:    "route_unavailable",
				Message: "Route inference is unavailable.",
				Scope:   "route",
			},
		},
		Provenance: Provenance{
			ResolverVersion: "route-resolver-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("b", 64),
			TrajectoryUpdatedAt: asOfTime.Add(-time.Minute),
			SourceNames: []string{
				"trajectory",
			},
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}
}

func validEvidence(
	evidenceType EvidenceType,
	observedAt time.Time,
	role string,
) Evidence {
	return Evidence{
		Type:          evidenceType,
		SourceName:    "trajectory",
		SourceVersion: "trajectory-v1",
		Score:         0.9,
		Weight:        1,
		ObservedAt:    observedAt.UTC(),
		Summary:       "Endpoint proximity evidence.",
		Attributes: []EvidenceAttribute{
			{
				Key:   "distance_km",
				Value: "2.5",
			},
			{
				Key:   "role",
				Value: role,
			},
		},
	}
}

func TestValidFixturesAreIndependent(
	t *testing.T,
) {
	first := validCompleteResult()
	second := validCompleteResult()

	first.Provenance.SourceNames[0] = "changed"
	if reflect.DeepEqual(first, second) {
		t.Fatal("fixtures unexpectedly shared state")
	}
}
