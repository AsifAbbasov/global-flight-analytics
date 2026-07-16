package interactionradius

import (
	"errors"
	"testing"
	"time"
)

func TestEvaluateAllowedHighSpeedDecision(t *testing.T) {
	policy := DefaultPolicy()
	request := validRequest()
	request.VelocityMetersPerSecond = 250
	request.QualityScore = 0.90

	decision, err := Evaluate(request, policy)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if decision.Status != DecisionStatusAllowed {
		t.Fatalf("Status = %q, want %q", decision.Status, DecisionStatusAllowed)
	}
	if decision.MotionClass != MotionClassHighSpeed {
		t.Fatalf("MotionClass = %q, want %q", decision.MotionClass, MotionClassHighSpeed)
	}
	if decision.HorizontalRadiusKilometers <= policy.BaseHorizontalRadiusKilometers ||
		decision.HorizontalRadiusKilometers > policy.MaximumHorizontalRadiusKilometers {
		t.Fatalf("HorizontalRadiusKilometers = %v", decision.HorizontalRadiusKilometers)
	}
	if !decision.VerticalFilteringPermitted {
		t.Fatal("VerticalFilteringPermitted = false, want true")
	}
	if Validate(decision).Status != ValidationStatusValid {
		t.Fatalf("decision validation failed: %#v", Validate(decision).Issues)
	}
}

func TestEvaluateUnknownAltitudeLimitsAndWidensVerticalSearch(t *testing.T) {
	policy := DefaultPolicy()
	request := validRequest()
	request.AltitudeMeters = nil
	request.AltitudeReference = AltitudeReferenceUnknown

	decision, err := Evaluate(request, policy)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if decision.Status != DecisionStatusLimited {
		t.Fatalf("Status = %q, want %q", decision.Status, DecisionStatusLimited)
	}
	if decision.VerticalFilteringPermitted {
		t.Fatal("VerticalFilteringPermitted = true, want false")
	}
	if decision.VerticalRadiusMeters != policy.MaximumVerticalRadiusMeters {
		t.Fatalf("VerticalRadiusMeters = %v, want %v", decision.VerticalRadiusMeters, policy.MaximumVerticalRadiusMeters)
	}
}

func TestEvaluateStaleObservationBlocksSearch(t *testing.T) {
	policy := DefaultPolicy()
	request := validRequest()
	request.ObservedAt = request.AsOfTime.Add(-policy.MaximumObservationAge - time.Second)

	decision, err := Evaluate(request, policy)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if decision.Status != DecisionStatusBlocked {
		t.Fatalf("Status = %q, want %q", decision.Status, DecisionStatusBlocked)
	}
	if decision.HorizontalRadiusKilometers != 0 || decision.VerticalRadiusMeters != 0 {
		t.Fatalf("blocked radii = %v/%v, want zero", decision.HorizontalRadiusKilometers, decision.VerticalRadiusMeters)
	}
}

func TestEvaluateLowQualityBlocksSearch(t *testing.T) {
	policy := DefaultPolicy()
	request := validRequest()
	request.QualityScore = policy.MinimumUsableQuality - 0.01

	decision, err := Evaluate(request, policy)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if decision.Status != DecisionStatusBlocked {
		t.Fatalf("Status = %q, want %q", decision.Status, DecisionStatusBlocked)
	}
}

func TestEvaluateRejectsFutureAndGroundEvidence(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Request)
	}{
		{
			name: "future observation",
			mutate: func(request *Request) {
				request.ObservedAt = request.AsOfTime.Add(time.Second)
			},
		},
		{
			name: "ground evidence",
			mutate: func(request *Request) {
				request.OnGround = true
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validRequest()
			test.mutate(&request)
			_, err := Evaluate(request, DefaultPolicy())
			if !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("error = %v, want ErrInvalidRequest", err)
			}
		})
	}
}

func TestEvaluateIsDeterministicAndCloneSafe(t *testing.T) {
	request := validRequest()
	policy := DefaultPolicy()
	first, err := Evaluate(request, policy)
	if err != nil {
		t.Fatalf("first Evaluate() error = %v", err)
	}
	second, err := Evaluate(request, policy)
	if err != nil {
		t.Fatalf("second Evaluate() error = %v", err)
	}
	if first.Provenance.InputFingerprint != second.Provenance.InputFingerprint {
		t.Fatalf("fingerprints differ: %q != %q", first.Provenance.InputFingerprint, second.Provenance.InputFingerprint)
	}
	clone := first.Clone()
	clone.Limitations[0].Code = "changed"
	clone.Confidence.Reasons[0].Code = "changed"
	if first.Limitations[0].Code == "changed" || first.Confidence.Reasons[0].Code == "changed" {
		t.Fatal("Clone() shared mutable slices")
	}
}

func validRequest() Request {
	asOfTime := time.Date(2026, 7, 16, 15, 0, 0, 0, time.UTC)
	altitude := 10_000.0
	return Request{
		RegionCode:                  "caucasus",
		NodeID:                      "trajectory:current-1",
		ICAO24:                      "abc123",
		Callsign:                    "gfa11",
		VelocityMetersPerSecond:     210,
		VerticalRateMetersPerSecond: 1.5,
		AltitudeMeters:              &altitude,
		AltitudeReference:           AltitudeReferenceBarometric,
		ObservedAt:                  asOfTime.Add(-20 * time.Second),
		AsOfTime:                    asOfTime,
		GeneratedAt:                 asOfTime.Add(time.Second),
		SourceName:                  "airplanes-live",
		QualityScore:                0.85,
	}
}
