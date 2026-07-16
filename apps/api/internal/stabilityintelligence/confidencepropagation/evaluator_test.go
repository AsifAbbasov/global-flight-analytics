package confidencepropagation

import (
	"testing"
	"time"
)

func TestPropagateCapsByWeakestRequiredDependency(t *testing.T) {
	request := Request{
		TargetNodeID: "output",
		EvaluatedAt: time.Date(
			2035,
			time.January,
			15,
			12,
			0,
			0,
			0,
			time.UTC,
		),
		Nodes: []Node{
			{
				ID:             "observed",
				Label:          "Observed trajectory",
				Kind:           NodeKindEvidence,
				Classification: ClassificationObserved,
				LocalScore:     0.90,
			},
			{
				ID:             "estimated",
				Label:          "Estimated route",
				Kind:           NodeKindEvidence,
				Classification: ClassificationEstimated,
				LocalScore:     0.45,
			},
			{
				ID:             "decision",
				Label:          "Projection decision",
				Kind:           NodeKindDecision,
				Classification: ClassificationDerived,
				LocalScore:     0.90,
				Dependencies: []Dependency{
					{
						NodeID:   "observed",
						Weight:   0.50,
						Required: true,
					},
					{
						NodeID:   "estimated",
						Weight:   0.50,
						Required: true,
					},
				},
			},
			{
				ID:             "output",
				Label:          "Forecast output",
				Kind:           NodeKindOutput,
				Classification: ClassificationDerived,
				LocalScore:     0.95,
				Dependencies: []Dependency{
					{
						NodeID:   "decision",
						Weight:   1,
						Required: true,
					},
				},
			},
		},
	}

	result, err := Propagate(request, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.Score > 0.65 {
		t.Fatalf(
			"score = %.6f, expected weakest-link cap",
			result.Score,
		)
	}
	if result.Status != ResultStatusLimited {
		t.Fatalf("status = %q", result.Status)
	}
	if err := ValidateResult(result, DefaultPolicy()); err != nil {
		t.Fatal(err)
	}
}

func TestPropagateRejectsCycle(t *testing.T) {
	request := Request{
		TargetNodeID: "a",
		EvaluatedAt:  time.Now().UTC(),
		Nodes: []Node{
			{
				ID:             "a",
				Label:          "A",
				Kind:           NodeKindDecision,
				Classification: ClassificationDerived,
				LocalScore:     0.8,
				Dependencies: []Dependency{
					{
						NodeID:   "b",
						Weight:   1,
						Required: true,
					},
				},
			},
			{
				ID:             "b",
				Label:          "B",
				Kind:           NodeKindDecision,
				Classification: ClassificationDerived,
				LocalScore:     0.8,
				Dependencies: []Dependency{
					{
						NodeID:   "a",
						Weight:   1,
						Required: true,
					},
				},
			},
		},
	}

	if _, err := Propagate(request, DefaultPolicy()); err == nil {
		t.Fatal("cycle was accepted")
	}
}

func TestPropagationFingerprintIgnoresNodeOrder(t *testing.T) {
	nodes := []Node{
		{
			ID:             "source",
			Label:          "Source",
			Kind:           NodeKindEvidence,
			Classification: ClassificationObserved,
			LocalScore:     0.8,
		},
		{
			ID:             "output",
			Label:          "Output",
			Kind:           NodeKindOutput,
			Classification: ClassificationDerived,
			LocalScore:     0.8,
			Dependencies: []Dependency{
				{
					NodeID:   "source",
					Weight:   1,
					Required: true,
				},
			},
		},
	}
	evaluatedAt := time.Date(
		2035,
		time.January,
		15,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	left, err := Propagate(
		Request{
			TargetNodeID: "output",
			Nodes:        nodes,
			EvaluatedAt:  evaluatedAt,
		},
		DefaultPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}

	nodes[0], nodes[1] = nodes[1], nodes[0]
	right, err := Propagate(
		Request{
			TargetNodeID: "output",
			Nodes:        nodes,
			EvaluatedAt:  evaluatedAt,
		},
		DefaultPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if left.Provenance.InputFingerprint !=
		right.Provenance.InputFingerprint {
		t.Fatal("node order changed fingerprint")
	}
}

func TestValidateResultRejectsTamperedFingerprint(t *testing.T) {
	result, err := Propagate(
		Request{
			TargetNodeID: "output",
			EvaluatedAt:  time.Now().UTC(),
			Nodes: []Node{
				{
					ID:             "output",
					Label:          "Output",
					Kind:           NodeKindOutput,
					Classification: ClassificationObserved,
					LocalScore:     0.8,
				},
			},
		},
		DefaultPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}
	result.Provenance.InputFingerprint = "sha256:tampered"
	if err := ValidateResult(result, DefaultPolicy()); err == nil {
		t.Fatal("tampered result was accepted")
	}
}
