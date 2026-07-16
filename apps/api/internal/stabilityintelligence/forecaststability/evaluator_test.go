package forecaststability

import (
	"testing"
	"time"
)

func TestEvaluateDecisionStabilityUnchanged(t *testing.T) {
	record := mustVersion(t, testProjection(), nil, "policy-v1", "build-v1", 0)
	result, err := EvaluateDecisionStability(StabilityRequest{
		Baseline:    record,
		Candidate:   record,
		EvaluatedAt: record.CreatedAt.Add(time.Second),
	}, DefaultStabilityPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.Level != StabilityLevelUnchanged || result.Score != 1 {
		t.Fatalf("result = %#v", result)
	}
}

func TestEvaluateDecisionStabilityStableForSmallShift(t *testing.T) {
	baseline := mustVersion(t, testProjection(), nil, "policy-v1", "build-v1", 0)
	candidateProjection := testProjection()
	for index := range candidateProjection.Points {
		candidateProjection.Points[index].Position.Longitude += 0.005
	}
	candidateProjection.GeneratedAt = candidateProjection.GeneratedAt.Add(time.Minute)
	candidateProjection.Provenance.InputFingerprint = fingerprintOf("stable-input")
	candidate := mustVersion(t, candidateProjection, &baseline, "policy-v1", "build-v1", time.Minute)
	result, err := EvaluateDecisionStability(StabilityRequest{
		Baseline:    baseline,
		Candidate:   candidate,
		EvaluatedAt: candidate.CreatedAt.Add(time.Second),
	}, DefaultStabilityPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.Level != StabilityLevelStable {
		t.Fatalf("level = %q metrics=%#v", result.Level, result.Metrics)
	}
}

func TestEvaluateDecisionStabilityMaterialForMethodChange(t *testing.T) {
	baseline := mustVersion(t, testProjection(), nil, "policy-v1", "build-v1", 0)
	candidateProjection := testProjection()
	candidateProjection.Method.Name = "historical_neighbor_continuation"
	candidateProjection.Method.Version = "v2"
	candidateProjection.GeneratedAt = candidateProjection.GeneratedAt.Add(time.Minute)
	candidateProjection.Provenance.InputFingerprint = fingerprintOf("method-input")
	candidate := mustVersion(t, candidateProjection, &baseline, "policy-v1", "build-v2", time.Minute)
	result, err := EvaluateDecisionStability(StabilityRequest{
		Baseline:    baseline,
		Candidate:   candidate,
		EvaluatedAt: candidate.CreatedAt.Add(time.Second),
	}, DefaultStabilityPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.Level != StabilityLevelMaterialChange || !result.Metrics.MethodChanged {
		t.Fatalf("result = %#v", result)
	}
}

func TestEvaluateDecisionStabilityIndeterminateWhenAlignmentIsLow(t *testing.T) {
	baseline := mustVersion(t, testProjection(), nil, "policy-v1", "build-v1", 0)
	candidateProjection := testProjection()
	for index := range candidateProjection.Points {
		candidateProjection.Points[index].ForecastTime = candidateProjection.Points[index].ForecastTime.Add(15 * time.Second)
	}
	candidateProjection.Horizon.AsOfTime = candidateProjection.Horizon.AsOfTime.Add(15 * time.Second)
	candidateProjection.Horizon.EndTime = candidateProjection.Horizon.EndTime.Add(15 * time.Second)
	candidateProjection.GeneratedAt = candidateProjection.GeneratedAt.Add(time.Minute)
	candidateProjection.Provenance.InputFingerprint = fingerprintOf("unaligned-input")
	candidate := mustVersion(t, candidateProjection, &baseline, "policy-v1", "build-v1", time.Minute)
	result, err := EvaluateDecisionStability(StabilityRequest{
		Baseline:    baseline,
		Candidate:   candidate,
		EvaluatedAt: candidate.CreatedAt.Add(time.Second),
	}, DefaultStabilityPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.Level != StabilityLevelIndeterminate || result.Status != ResultStatusLimited {
		t.Fatalf("result = %#v", result)
	}
}

func TestStabilityFingerprintDeterministic(t *testing.T) {
	baseline := mustVersion(t, testProjection(), nil, "policy-v1", "build-v1", 0)
	candidateProjection := testProjection()
	candidateProjection.Points[0].Position.Longitude += 0.02
	candidateProjection.GeneratedAt = candidateProjection.GeneratedAt.Add(time.Minute)
	candidateProjection.Provenance.InputFingerprint = fingerprintOf("deterministic-input")
	candidate := mustVersion(t, candidateProjection, &baseline, "policy-v1", "build-v1", time.Minute)
	request := StabilityRequest{Baseline: baseline, Candidate: candidate, EvaluatedAt: candidate.CreatedAt.Add(time.Second)}
	left, err := EvaluateDecisionStability(request, DefaultStabilityPolicy())
	if err != nil {
		t.Fatal(err)
	}
	right, err := EvaluateDecisionStability(request, DefaultStabilityPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if left.Provenance.InputFingerprint != right.Provenance.InputFingerprint {
		t.Fatal("stability fingerprint is not deterministic")
	}
}
