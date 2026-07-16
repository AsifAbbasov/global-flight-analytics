package unknownintervention

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func TestEvaluateAllowsStrongContextOnly(t *testing.T) {
	result, err := Evaluate(baseRequest(ClaimKindContextualAssociation, []Evidence{{ID: "track", Label: "Observed track", Class: EvidenceObserved, Score: .9, Required: true, Fingerprint: hash("track")}, {ID: "weather", Label: "Open weather", Class: EvidenceOpenlySourced, Score: .85, Fingerprint: hash("weather")}}), DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != DecisionAllowedContextOnly || result.Status != ResultStatusComplete {
		t.Fatalf("result=%#v", result)
	}
}

func TestEvaluateWithholdsIntentAndExactCause(t *testing.T) {
	for _, kind := range []ClaimKind{ClaimKindIntentAttribution, ClaimKindOperationalInstruction, ClaimKindCausalAttribution} {
		result, err := Evaluate(baseRequest(kind, []Evidence{{ID: "track", Label: "Observed track", Class: EvidenceObserved, Score: .95, Required: true, Fingerprint: hash("track")}}), DefaultPolicy())
		if err != nil {
			t.Fatal(err)
		}
		if result.Decision != DecisionWithheld {
			t.Fatalf("kind=%s result=%#v", kind, result)
		}
	}
}

func TestEvaluateUnknownRequiredEvidenceWithholds(t *testing.T) {
	request := baseRequest(ClaimKindContextualAssociation, []Evidence{{ID: "cause", Label: "Unknown cause", Class: EvidenceUnknown, Score: .3, Required: true, Fingerprint: hash("unknown"), Limitation: "Cause is not observed."}})
	result, err := Evaluate(request, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != DecisionWithheld || result.Metrics.UnknownEvidenceCount != 1 {
		t.Fatalf("result=%#v", result)
	}
}

func TestFingerprintDeterministicAcrossEvidenceOrder(t *testing.T) {
	evidence := []Evidence{{ID: "a", Label: "A", Class: EvidenceObserved, Score: .9, Required: true, Fingerprint: hash("a")}, {ID: "b", Label: "B", Class: EvidenceDerived, Score: .8, Fingerprint: hash("b")}}
	request := baseRequest(ClaimKindContextualAssociation, evidence)
	left, err := Evaluate(request, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	request.Evidence[0], request.Evidence[1] = request.Evidence[1], request.Evidence[0]
	right, err := Evaluate(request, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if left.Provenance.InputFingerprint != right.Provenance.InputFingerprint {
		t.Fatal("fingerprint changed")
	}
}

func baseRequest(kind ClaimKind, evidence []Evidence) Request {
	return Request{SubjectID: "forecast-1", ClaimKind: kind, ClaimText: "The observed change is associated with available context.", Evidence: evidence, EvidenceCompleteness: .9, EvaluatedAt: time.Now().UTC()}
}
func hash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}
