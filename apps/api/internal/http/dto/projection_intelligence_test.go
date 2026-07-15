package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionfreshness"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
)

func TestToProjectionIntelligenceResponseUsesStableJSONContract(
	t *testing.T,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		16,
		12,
		0,
		0,
		0,
		time.FixedZone("UTC plus four", 4*60*60),
	)
	result := projectionHTTPDTOFixture(asOfTime)

	converted := ToProjectionIntelligenceResponse(
		result,
	)

	if converted.Version !=
		projectionproduction.Version ||
		converted.Strategy !=
			"historical_neighbor_continuation" ||
		converted.ArrivalStatus != "attached" ||
		converted.Projection.Horizon.StepSeconds != 60 ||
		converted.Projection.Horizon.DurationSeconds != 180 ||
		len(converted.Projection.Points) != 1 ||
		converted.Projection.Arrival == nil ||
		converted.Evidence.NeighborSelection == nil ||
		converted.Evidence.PatternConfidence == nil ||
		converted.Evidence.Freshness == nil ||
		converted.Evidence.RouteFrequency == nil {
		t.Fatalf(
			"unexpected converted Projection Intelligence response: %#v",
			converted,
		)
	}

	if converted.GeneratedAt.Location() != time.UTC ||
		converted.Projection.GeneratedAt.Location() != time.UTC ||
		converted.Projection.Horizon.AsOfTime.Location() != time.UTC ||
		converted.Projection.Points[0].
			ForecastTime.Location() != time.UTC {
		t.Fatal(
			"Projection Intelligence timestamps were not normalized to UTC",
		)
	}

	payload, err := json.Marshal(converted)
	if err != nil {
		t.Fatalf(
			"marshal Projection Intelligence response: %v",
			err,
		)
	}

	text := string(payload)
	for _, fragment := range []string{
		`"strategy":"historical_neighbor_continuation"`,
		`"arrival_status":"attached"`,
		`"horizontal_radius_m"`,
		`"input_fingerprint"`,
		`"neighbor_selection"`,
		`"pattern_confidence"`,
		`"freshness"`,
		`"route_frequency"`,
		`"latest_observation_age_seconds"`,
		`"scope_guard":"research_only_not_for_operational_use"`,
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf(
				"JSON contract is missing %s: %s",
				fragment,
				text,
			)
		}
	}
}

func TestToProjectionIntelligenceResponseOmitsUnavailableEvidence(
	t *testing.T,
) {
	result := projectionproduction.Result{
		Version: projectionproduction.Version,
		Strategy: projectionproduction.
			StrategyKinematic,
		FallbackReason: "route_contract_invalid",
		ArrivalStatus: projectionproduction.
			ArrivalStatusSkipped,
		Projection: projectioncontract.Result{
			SchemaVersion: projectioncontract.
				SchemaVersionV1,
			Status: projectioncontract.
				ResultStatusUnavailable,
			TrajectoryID: "trajectory-001",
			Method: projectioncontract.Method{
				Name:    "short_horizon_kinematic_baseline",
				Version: "test-v1",
				DecisionClass: projectioncontract.
					DecisionClassPhysicsDerived,
			},
			Horizon: projectioncontract.Horizon{
				AsOfTime: time.Date(
					2026,
					time.July,
					16,
					12,
					0,
					0,
					0,
					time.UTC,
				),
				EndTime: time.Date(
					2026,
					time.July,
					16,
					12,
					1,
					0,
					0,
					time.UTC,
				),
				Step: time.Minute,
			},
			Points: []projectioncontract.ProjectionPoint{},
			Confidence: projectioncontract.Confidence{
				Score: 0,
				Level: projectioncontract.
					ConfidenceLevelNone,
				Reasons: []projectioncontract.
					ConfidenceReason{},
			},
			Limitations:  []projectioncontract.Limitation{},
			Explanations: []projectioncontract.Explanation{},
			ScopeGuard: projectioncontract.
				ScopeGuardResearchOnly,
			Provenance: projectioncontract.Provenance{
				Inputs: []projectioncontract.InputReference{},
			},
			GeneratedAt: time.Date(
				2026,
				time.July,
				16,
				12,
				0,
				1,
				0,
				time.UTC,
			),
		},
		Notices: []projectionproduction.Notice{
			{
				Code:    "route_contract_invalid",
				Message: "Route contract was invalid.",
			},
		},
		InputFingerprint: "sha256:" +
			strings.Repeat("f", 64),
		GeneratedAt: time.Date(
			2026,
			time.July,
			16,
			12,
			0,
			1,
			0,
			time.UTC,
		),
	}

	converted := ToProjectionIntelligenceResponse(
		result,
	)
	payload, err := json.Marshal(converted)
	if err != nil {
		t.Fatalf(
			"marshal Projection Intelligence response: %v",
			err,
		)
	}

	text := string(payload)
	for _, fragment := range []string{
		`"neighbor_selection"`,
		`"pattern_confidence"`,
		`"freshness"`,
		`"route_frequency"`,
	} {
		if strings.Contains(text, fragment) {
			t.Fatalf(
				"unavailable evidence field %s must be omitted: %s",
				fragment,
				text,
			)
		}
	}
}

func projectionHTTPDTOFixture(
	asOfTime time.Time,
) projectionproduction.Result {
	utcAsOf := asOfTime.UTC()
	altitudeM := 1000.0
	verticalRadiusM := 120.0
	arrival := &projectioncontract.ArrivalEstimate{
		AirportICAOCode: "LTBA",
		EarliestTime: utcAsOf.Add(
			2 * time.Minute,
		),
		EstimatedTime: utcAsOf.Add(
			3 * time.Minute,
		),
		LatestTime: utcAsOf.Add(
			4 * time.Minute,
		),
		Confidence: projectioncontract.Confidence{
			Score: 0.8,
			Level: projectioncontract.
				ConfidenceLevelHigh,
			Reasons: []projectioncontract.ConfidenceReason{
				{
					Code:         "arrival",
					Message:      "Arrival confidence.",
					Contribution: 0.8,
				},
			},
		},
		Limitations: []projectioncontract.Limitation{
			{
				Code:    "airport_radius",
				Message: "Arrival means airport-radius entry.",
				Scope:   "arrival",
			},
		},
	}

	selection := &projectionneighbors.Result{
		Version:                      projectionneighbors.Version,
		Status:                       projectionneighbors.StatusComplete,
		CurrentTrajectoryID:          "trajectory-001",
		AsOfTime:                     utcAsOf,
		RequiredContinuationDuration: 3 * time.Minute,
		InputCandidateCount:          1,
		CheckedCandidateCount:        1,
		QualifiedCandidateCount:      1,
		RejectedCandidateCount:       0,
		SelectionLimit:               1,
		Neighbors: []projectionneighbors.Neighbor{
			{
				TrajectoryID:           "historical-001",
				SimilarityScore:        0.9,
				SimilarityLevel:        historicalsimilarity.LevelHigh,
				AnchorPointIndex:       2,
				AnchorObservedAt:       utcAsOf.Add(-24 * time.Hour),
				AnchorDistanceKM:       1.5,
				CandidateStartTime:     utcAsOf.Add(-25 * time.Hour),
				CandidateEndTime:       utcAsOf.Add(-24 * time.Hour),
				CandidateAge:           24 * time.Hour,
				PrefixPointCount:       3,
				ContinuationPointCount: 3,
				ContinuationEndTime:    utcAsOf.Add(-24*time.Hour + 3*time.Minute),
			},
		},
		Limitations: []projectionneighbors.Notice{},
		InputFingerprint: "sha256:" +
			strings.Repeat("a", 64),
	}

	pattern := &projectionpatternconfidence.Result{
		Version: projectionpatternconfidence.Version,
		Status: projectionpatternconfidence.
			StatusComplete,
		Usable:                  true,
		NeighborCount:           1,
		TargetNeighborCount:     1,
		MeanSimilarityScore:     0.9,
		MeanCandidateAgeSeconds: 86400,
		MeanAnchorDistanceKM:    1.5,
		Score:                   0.85,
		Level: projectioncontract.
			ConfidenceLevelHigh,
		Components: []projectionpatternconfidence.Component{
			{
				Name: projectionpatternconfidence.
					ComponentSimilarity,
				Score:  0.9,
				Weight: 1,
			},
		},
		SelectedTrajectoryIDs: []string{
			"historical-001",
		},
		Limitations: []projectionpatternconfidence.Notice{},
		InputFingerprint: "sha256:" +
			strings.Repeat("b", 64),
	}

	freshness := &projectionfreshness.Result{
		Version:             projectionfreshness.Version,
		Decision:            projectionfreshness.DecisionAllowed,
		Usable:              true,
		AsOfTime:            utcAsOf,
		NeighborCount:       1,
		RecentNeighborCount: 1,
		NewestNeighborAge:   24 * time.Hour,
		MeanNeighborAge:     24 * time.Hour,
		OldestNeighborAge:   24 * time.Hour,
		Score:               0.9,
		Components: []projectionfreshness.Component{
			{
				Name: projectionfreshness.
					ComponentNewestAge,
				Score:  0.9,
				Weight: 1,
			},
		},
		SelectedTrajectoryIDs: []string{
			"historical-001",
		},
		Limitations: []projectionfreshness.Notice{},
		InputFingerprint: "sha256:" +
			strings.Repeat("c", 64),
	}

	frequency := &projectionroutefrequency.Result{
		Version:                projectionroutefrequency.Version,
		Decision:               projectionroutefrequency.DecisionAllowed,
		Usable:                 true,
		RouteKey:               "UBBB>LTBA",
		AsOfTime:               utcAsOf,
		ObservationCount:       10,
		DistinctFlightCount:    8,
		DistinctDayCount:       6,
		RecentObservationCount: 4,
		LatestObservationAge:   24 * time.Hour,
		RouteConfidenceScore:   0.9,
		Score:                  0.85,
		Components: []projectionroutefrequency.Component{
			{
				Name: projectionroutefrequency.
					ComponentObservationCount,
				Score:  1,
				Weight: 1,
			},
		},
		Limitations: []projectionroutefrequency.Notice{},
		HistoryInputFingerprint: "sha256:" +
			strings.Repeat("d", 64),
		InputFingerprint: "sha256:" +
			strings.Repeat("e", 64),
	}

	generatedAt := utcAsOf.Add(time.Second)

	return projectionproduction.Result{
		Version: projectionproduction.Version,
		Strategy: projectionproduction.
			StrategyHistoricalNeighbor,
		ArrivalStatus: projectionproduction.
			ArrivalStatusAttached,
		Projection: projectioncontract.Result{
			SchemaVersion: projectioncontract.
				SchemaVersionV1,
			Status:       projectioncontract.ResultStatusComplete,
			TrajectoryID: "trajectory-001",
			FlightID:     "flight-001",
			AircraftID:   "aircraft-001",
			ICAO24:       "4A1234",
			Callsign:     "AHY123",
			Method: projectioncontract.Method{
				Name:    "local_historical_neighbor_continuation",
				Version: "test-v1",
				DecisionClass: projectioncontract.
					DecisionClassExperimental,
			},
			Horizon: projectioncontract.Horizon{
				AsOfTime: utcAsOf,
				EndTime:  utcAsOf.Add(3 * time.Minute),
				Step:     time.Minute,
			},
			Points: []projectioncontract.ProjectionPoint{
				{
					Sequence:     0,
					ForecastTime: utcAsOf.Add(time.Minute),
					Position: projectioncontract.Position{
						Latitude:  40.4,
						Longitude: 49.9,
						AltitudeM: &altitudeM,
					},
					Uncertainty: projectioncontract.Uncertainty{
						HorizontalRadiusM: 800,
						VerticalRadiusM:   &verticalRadiusM,
					},
					Confidence: projectioncontract.Confidence{
						Score: 0.8,
						Level: projectioncontract.
							ConfidenceLevelHigh,
						Reasons: []projectioncontract.ConfidenceReason{
							{
								Code:         "point",
								Message:      "Point confidence.",
								Contribution: 0.8,
							},
						},
					},
				},
			},
			Arrival: arrival,
			Confidence: projectioncontract.Confidence{
				Score: 0.8,
				Level: projectioncontract.
					ConfidenceLevelHigh,
				Reasons: []projectioncontract.ConfidenceReason{
					{
						Code:         "result",
						Message:      "Result confidence.",
						Contribution: 0.8,
					},
				},
			},
			Limitations: []projectioncontract.Limitation{},
			Explanations: []projectioncontract.Explanation{
				{
					Code:    "historical_neighbor",
					Message: "Historical continuation selected.",
				},
			},
			ScopeGuard: projectioncontract.
				ScopeGuardResearchOnly,
			Provenance: projectioncontract.Provenance{
				InputFingerprint: "sha256:" +
					strings.Repeat("1", 64),
				Inputs: []projectioncontract.InputReference{
					{
						Name: "current_trajectory",
						Classification: projectioncontract.
							InputClassificationObserved,
						SourceName:  "test",
						ObservedAt:  utcAsOf,
						RetrievedAt: generatedAt,
					},
				},
				LatestInputObservedAt: utcAsOf,
			},
			GeneratedAt: generatedAt,
		},
		NeighborSelection: selection,
		PatternConfidence: pattern,
		Freshness:         freshness,
		RouteFrequency:    frequency,
		Notices:           []projectionproduction.Notice{},
		InputFingerprint: "sha256:" +
			strings.Repeat("2", 64),
		GeneratedAt: generatedAt,
	}
}
