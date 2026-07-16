package failureexplanation

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func TestExplainRanksBlockingUnknownCauseFirst(t *testing.T) {
	result, err := Explain(Request{
		SubjectID: "forecast-1", SubjectType: "projection", EvaluatedAt: time.Now().UTC(),
		Signals: []Signal{
			{Code: "low_confidence", Category: CategoryConfidence, Severity: SeverityWarning, Classification: CauseClassificationDerivedCondition, Summary: "Confidence is low.", Source: "confidence", EvidenceFingerprints: []string{hash("a")}},
			{Code: "unknown_change", Category: CategoryUnknown, Severity: SeverityBlocking, Classification: CauseClassificationUnknownCause, Summary: "Change cause is unknown.", Source: "guard", BlocksUse: true},
		},
	}, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.PrimaryCode != "unknown_change" || result.Status != ResultStatusLimited || result.Metrics.UnknownCauseCount != 1 || result.Metrics.BlockingCount != 1 {
		t.Fatalf("result = %#v", result)
	}
}

func TestExplainDeterministicAcrossInputOrder(t *testing.T) {
	signals := []Signal{
		{Code: "a", Category: CategoryEvidence, Severity: SeverityWarning, Classification: CauseClassificationDerivedCondition, Summary: "A.", Source: "source", EvidenceFingerprints: []string{hash("a")}},
		{Code: "b", Category: CategoryPolicy, Severity: SeverityInformation, Classification: CauseClassificationObservedCondition, Summary: "B.", Source: "source", EvidenceFingerprints: []string{hash("b")}},
	}
	request := Request{SubjectID: "subject", SubjectType: "test", Signals: signals, EvaluatedAt: time.Now().UTC()}
	left, err := Explain(request, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	request.Signals[0], request.Signals[1] = request.Signals[1], request.Signals[0]
	right, err := Explain(request, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if left.Provenance.InputFingerprint != right.Provenance.InputFingerprint {
		t.Fatal("fingerprint changed under input reorder")
	}
}

func TestExplainRejectsDuplicateSourceCode(t *testing.T) {
	signal := Signal{Code: "same", Category: CategorySystem, Severity: SeverityWarning, Classification: CauseClassificationDerivedCondition, Summary: "Same.", Source: "source"}
	_, err := Explain(Request{SubjectID: "subject", SubjectType: "test", Signals: []Signal{signal, signal}, EvaluatedAt: time.Now().UTC()}, DefaultPolicy())
	if err == nil {
		t.Fatal("duplicate signal accepted")
	}
}

func hash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}
