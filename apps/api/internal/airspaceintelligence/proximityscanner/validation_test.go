package proximityscanner

import "testing"

func TestValidateAcceptsScanResult(t *testing.T) {
	result, err := Scan(scanRequest(t), DefaultPolicy())
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if report := Validate(result, DefaultPolicy()); report.Status != ValidationStatusValid {
		t.Fatalf("Validate() = %+v", report)
	}
}

func TestValidateRejectsTamperedMetrics(t *testing.T) {
	result, err := Scan(scanRequest(t), DefaultPolicy())
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	result.Metrics.CandidatePairCount++
	if report := Validate(result, DefaultPolicy()); report.Status != ValidationStatusInvalid {
		t.Fatalf("Validate() = %+v, want invalid", report)
	}
}

func TestValidateRejectsGraphCandidateMismatch(t *testing.T) {
	result, err := Scan(scanRequest(t), DefaultPolicy())
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	result.Graph.Edges = nil
	result.Graph.Metrics.EdgeCount = 0
	if report := Validate(result, DefaultPolicy()); report.Status != ValidationStatusInvalid {
		t.Fatalf("Validate() = %+v, want invalid", report)
	}
}
