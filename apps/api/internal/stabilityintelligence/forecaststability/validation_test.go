package forecaststability

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func TestValidationRejectsTamperedVersionRecord(t *testing.T) {
	record := mustVersion(t, testProjection(), nil, "policy-v1", "build-v1", 0)
	record.OutputFingerprint = fingerprintOf("tampered")
	if err := ValidateVersionRecord(record, DefaultVersionPolicy()); err == nil {
		t.Fatal("tampered record was accepted")
	}
}

func TestValidationRejectsTamperedStabilityFingerprint(t *testing.T) {
	record := mustVersion(t, testProjection(), nil, "policy-v1", "build-v1", 0)
	result, err := EvaluateDecisionStability(StabilityRequest{
		Baseline:    record,
		Candidate:   record,
		EvaluatedAt: record.CreatedAt.Add(time.Second),
	}, DefaultStabilityPolicy())
	if err != nil {
		t.Fatal(err)
	}
	result.Provenance.InputFingerprint = fingerprintOf("tampered")
	if err := ValidateStabilityResult(result, DefaultStabilityPolicy()); err == nil {
		t.Fatal("tampered stability result was accepted")
	}
}

func mustVersion(
	t *testing.T,
	projection projectioncontract.Result,
	previous *ForecastVersionRecord,
	policyVersion string,
	implementationVersion string,
	offset time.Duration,
) ForecastVersionRecord {
	t.Helper()
	result, err := RegisterVersion(ForecastVersionRequest{
		Projection:            projection,
		PolicyVersion:         policyVersion,
		ImplementationVersion: implementationVersion,
		Previous:              previous,
		RegisteredAt:          projection.GeneratedAt.Add(time.Second + offset),
	}, DefaultVersionPolicy())
	if err != nil {
		t.Fatalf("register version: %v", err)
	}
	return result.Record
}

func testProjection() projectioncontract.Result {
	asOf := time.Date(2035, 1, 15, 12, 0, 0, 0, time.UTC)
	vertical := 200.0
	altitude := 9000.0
	points := make([]projectioncontract.ProjectionPoint, 0, 4)
	for index := 0; index < 4; index++ {
		points = append(points, projectioncontract.ProjectionPoint{
			Sequence:     index,
			ForecastTime: asOf.Add(time.Duration(index+1) * 30 * time.Second),
			Position: projectioncontract.Position{
				Latitude:  40.40,
				Longitude: 49.80 + float64(index)*0.02,
				AltitudeM: floatPointer(altitude),
			},
			Uncertainty: projectioncontract.Uncertainty{
				HorizontalRadiusM: 1000 + float64(index)*100,
				VerticalRadiusM:   floatPointer(vertical),
			},
			Confidence: projectioncontract.Confidence{
				Score:   0.80 - float64(index)*0.02,
				Level:   projectioncontract.ConfidenceLevelMedium,
				Reasons: []projectioncontract.ConfidenceReason{{Code: "bounded_horizon", Message: "Bounded horizon confidence.", Contribution: 0.8}},
			},
		})
	}
	return projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        projectioncontract.ResultStatusComplete,
		TrajectoryID:  "trajectory-stage-12-001",
		FlightID:      "flight-stage-12-001",
		AircraftID:    "aircraft-stage-12-001",
		ICAO24:        "A1B2C3",
		Callsign:      "GFA1201",
		Method:        projectioncontract.Method{Name: "short_horizon_kinematic_baseline", Version: "v1", DecisionClass: projectioncontract.DecisionClassPhysicsDerived},
		Horizon:       projectioncontract.Horizon{AsOfTime: asOf, EndTime: asOf.Add(2 * time.Minute), Step: 30 * time.Second},
		Points:        points,
		Arrival: &projectioncontract.ArrivalEstimate{
			AirportICAOCode: "UBBB",
			EarliestTime:    asOf.Add(9 * time.Minute),
			EstimatedTime:   asOf.Add(10 * time.Minute),
			LatestTime:      asOf.Add(11 * time.Minute),
			Confidence:      projectioncontract.Confidence{Score: 0.70, Level: projectioncontract.ConfidenceLevelMedium, Reasons: []projectioncontract.ConfidenceReason{{Code: "route_context", Message: "Route context supports arrival.", Contribution: 0.7}}},
			Limitations:     []projectioncontract.Limitation{{Code: "estimated_arrival", Message: "Arrival is estimated.", Scope: "arrival"}},
		},
		Confidence: projectioncontract.Confidence{Score: 0.78, Level: projectioncontract.ConfidenceLevelMedium, Reasons: []projectioncontract.ConfidenceReason{{Code: "projection_method", Message: "Projection method confidence.", Contribution: 0.78}}},
		Limitations: []projectioncontract.Limitation{
			{Code: "research_only", Message: "Research only.", Scope: "operational_use"},
			{Code: "short_horizon", Message: "Short horizon only.", Scope: "horizon"},
		},
		Explanations: []projectioncontract.Explanation{
			{Code: "method", Message: "Kinematic baseline."},
			{Code: "uncertainty", Message: "Uncertainty grows over time."},
		},
		ScopeGuard: projectioncontract.ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: fingerprintOf("projection-input"),
			Inputs: []projectioncontract.InputReference{
				{Name: "current_trajectory", Classification: projectioncontract.InputClassificationObserved, ObservedAt: asOf.Add(-time.Second)},
				{Name: "route_context", Classification: projectioncontract.InputClassificationDerived, ObservedAt: asOf.Add(-time.Minute)},
			},
			LatestInputObservedAt: asOf.Add(-time.Second),
		},
		GeneratedAt: asOf.Add(time.Second),
	}
}

func floatPointer(value float64) *float64 { return &value }

func fingerprintOf(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}
